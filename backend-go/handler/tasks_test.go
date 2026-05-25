package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/middleware"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTasksHandlerTest(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()
	config.App = &config.Config{
		DataDir:     filepath.Join(tmp, "data"),
		UploadDir:   filepath.Join(tmp, "upload"),
		JWTSecret:   "test-secret",
		AdminApikey: "test-admin",
	}
	if err := os.MkdirAll(config.App.DataDir, 0755); err != nil {
		t.Fatalf("create data dir: %v", err)
	}
	if err := os.MkdirAll(config.App.UploadDir, 0755); err != nil {
		t.Fatalf("create upload dir: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(filepath.Join(config.App.DataDir, "test.sqlite")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	database.DB = db
	if err := database.DB.AutoMigrate(&database.User{}, &database.Task{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	now := time.Now().UnixMilli()
	if err := database.DB.Create(&database.User{
		ID:          "user-1",
		Label:       "user",
		Role:        "user",
		Status:      "active",
		CreatedAt:   now,
		LastLoginAt: &now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := database.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	r := gin.New()
	auth := r.Group("/api/tasks", middleware.AuthMiddleware())
	auth.PUT("/:id", TasksUpdate)
	auth.GET("/:id/stream", TaskStream)
	return r
}

func authTokenForTasksTest(t *testing.T) string {
	t.Helper()
	token, err := service.SignToken("user-1", "user", config.App.JWTSecret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func createTaskForTasksHandlerTest(t *testing.T, task *service.TaskRecord) {
	t.Helper()
	if err := service.UpsertTask("user-1", task); err != nil {
		t.Fatalf("create task: %v", err)
	}
}

func putTaskForTasksHandlerTest(t *testing.T, r *gin.Engine, taskID string, body map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/"+taskID, bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+authTokenForTasksTest(t))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}

func TestTaskUpdateOnlyChangesFavorite(t *testing.T) {
	r := setupTasksHandlerTest(t)
	finishedAt := int64(200)
	elapsed := int64(100)
	task := &service.TaskRecord{
		ID:            "task-1",
		Prompt:        "original prompt",
		Params:        service.TaskParams{Size: "1024x1024", Quality: "auto", OutputFormat: "png", N: 4},
		InputImageIDs: []string{"input-1"},
		OutputImages:  []string{},
		Status:        "running",
		IsFavorite:    false,
		CreatedAt:     100,
		FinishedAt:    &finishedAt,
		Elapsed:       &elapsed,
	}
	createTaskForTasksHandlerTest(t, task)

	resp := putTaskForTasksHandlerTest(t, r, "task-1", map[string]interface{}{
		"isFavorite":   true,
		"status":       "done",
		"params":       map[string]interface{}{"n": 1},
		"outputImages": []string{"forged-output"},
		"error":        "forged-error",
		"finishedAt":   999,
		"elapsed":      999,
	})

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	updated, err := service.GetTask("user-1", "task-1")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if !updated.IsFavorite {
		t.Fatalf("expected favorite to be updated")
	}
	if updated.Status != "running" {
		t.Fatalf("status changed to %q", updated.Status)
	}
	if len(updated.OutputImages) != 0 {
		t.Fatalf("output images changed: %#v", updated.OutputImages)
	}
	if updated.Error != nil {
		t.Fatalf("error changed: %v", *updated.Error)
	}
	params, ok := updated.Params.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected params type %#v", updated.Params)
	}
	if params["n"].(float64) != 4 {
		t.Fatalf("params changed: %#v", updated.Params)
	}
}

func TestTaskUpdateMissingTaskDoesNotCreate(t *testing.T) {
	r := setupTasksHandlerTest(t)

	resp := putTaskForTasksHandlerTest(t, r, "missing-task", map[string]interface{}{
		"isFavorite": true,
		"status":     "done",
		"prompt":     "forged",
	})

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", resp.Code, resp.Body.String())
	}
	if _, err := service.GetTask("user-1", "missing-task"); err == nil {
		t.Fatalf("missing task was created")
	}
}

func TestTaskStreamMissingTaskReturnsErrorStatusEvent(t *testing.T) {
	r := setupTasksHandlerTest(t)
	token := authTokenForTasksTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/missing-task/stream", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	body := resp.Body.String()
	for _, want := range []string{"event: task-update", `"status":"error"`, `"error":"任务不存在"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("SSE body missing %q: %s", want, body)
		}
	}
}
