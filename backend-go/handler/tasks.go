package handler

import (
	"encoding/json"
	"net/http"

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

	var body map[string]interface{}
	_ = c.ShouldBindJSON(&body)

	task, err := service.GetTask(user.ID, taskID)
	if err != nil {
		b, _ := json.Marshal(body)
		var newTask service.TaskRecord
		if err := json.Unmarshal(b, &newTask); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		newTask.ID = taskID
		if err := service.UpsertTask(user.ID, &newTask); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	if v, ok := body["isFavorite"]; ok {
		if b, ok := v.(bool); ok {
			task.IsFavorite = b
		}
	}
	if v, ok := body["params"]; ok {
		task.Params = v
	}
	if v, ok := body["status"]; ok {
		if s, ok := v.(string); ok {
			task.Status = s
		}
	}
	if v, ok := body["outputImages"]; ok {
		if ids, ok := v.([]interface{}); ok {
			strs := make([]string, 0, len(ids))
			for _, id := range ids {
				if s, ok := id.(string); ok {
					strs = append(strs, s)
				}
			}
			task.OutputImages = strs
		}
	}
	if v, ok := body["error"]; ok {
		if s, ok := v.(string); ok {
			task.Error = &s
		}
	}
	if v, ok := body["finishedAt"]; ok {
		if n, ok := v.(float64); ok {
			t := int64(n)
			task.FinishedAt = &t
		}
	}
	if v, ok := body["elapsed"]; ok {
		if n, ok := v.(float64); ok {
			t := int64(n)
			task.Elapsed = &t
		}
	}
	if v, ok := body["actualParams"]; ok {
		task.ActualParams = v
	}
	if v, ok := body["actualParamsByImage"]; ok {
		task.ActualParamsByImage = v
	}
	if v, ok := body["revisedPromptByImage"]; ok {
		task.RevisedPromptByImage = v
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
