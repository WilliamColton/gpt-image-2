package handler

import (
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
	auth.GET("/:id/stream", TaskStream)
	return r
}

func TestTaskStreamMissingTaskReturnsErrorStatusEvent(t *testing.T) {
	r := setupTasksHandlerTest(t)
	token, err := service.SignToken("user-1", "user", config.App.JWTSecret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

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
