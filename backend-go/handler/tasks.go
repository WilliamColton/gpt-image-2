package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"gpt-image-playground/backend/middleware"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

func TasksList(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	tasks, err := service.ListTasks(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func TasksUpdate(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	taskID := c.Param("id")

	var body struct {
		IsFavorite *bool `json:"isFavorite"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	task, err := service.GetTask(user.ID, taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if body.IsFavorite != nil {
		task.IsFavorite = *body.IsFavorite
	}

	if err := service.UpsertTask(user.ID, task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func TasksDelete(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	taskID := c.Param("id")
	service.DeleteTask(user.ID, taskID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func TasksClear(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	service.ClearTasks(user.ID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func TaskStream(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	taskID := c.Param("id")

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	sendEvent := func(task *service.TaskRecord) bool {
		data, _ := json.Marshal(task)
		_, err := fmt.Fprintf(c.Writer, "event: task-update\ndata: %s\n\n", data)
		if err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	// Send current state immediately
	task, err := service.GetTask(user.ID, taskID)
	if err != nil {
		slog.Warn("SSE: 任务不存在", "user_id", user.ID, "task_id", taskID)
		now := time.Now().UnixMilli()
		errorMsg := "任务不存在"
		errorTask := &service.TaskRecord{
			ID:            taskID,
			Prompt:        "",
			Params:        service.TaskParams{N: 1},
			InputImageIDs: []string{},
			OutputImages:  []string{},
			Status:        "error",
			Error:         &errorMsg,
			CreatedAt:     now,
			FinishedAt:    &now,
		}
		sendEvent(errorTask)
		return
	}
	if !sendEvent(task) {
		return
	}
	if task.Status == "done" || task.Status == "error" {
		return
	}

	// Poll DB until task finishes or client disconnects
	pollTicker := time.NewTicker(1 * time.Second)
	defer pollTicker.Stop()
	heartbeatTicker := time.NewTicker(15 * time.Second)
	defer heartbeatTicker.Stop()
	timeout := time.After(10 * time.Minute)

	ctx := c.Request.Context()
	lastStatus := task.Status

	for {
		select {
		case <-ctx.Done():
			return
		case <-timeout:
			slog.Warn("SSE: 超时断开", "user_id", user.ID, "task_id", taskID)
			return
		case <-heartbeatTicker.C:
			// Send heartbeat comment to keep connection alive
			if _, err := fmt.Fprintf(c.Writer, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-pollTicker.C:
			t, err := service.GetTask(user.ID, taskID)
			if err != nil {
				continue
			}
			if t.Status != lastStatus || t.Status == "done" || t.Status == "error" {
				if !sendEvent(t) {
					return
				}
				lastStatus = t.Status
				if t.Status == "done" || t.Status == "error" {
					return
				}
			}
		}
	}
}
