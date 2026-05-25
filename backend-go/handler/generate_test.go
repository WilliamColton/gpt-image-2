package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func setupGenerateHandlerTest(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()
	config.App = &config.Config{
		DataDir:     filepath.Join(tmp, "data"),
		UploadDir:   filepath.Join(tmp, "upload"),
		JWTSecret:   "test-secret",
		AdminApikey: "test-admin",
		Model:       "gpt-image-2",
	}
	config.ApiEndpoints = nil
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
		Quota:       service.MaxTaskN,
		CreatedAt:   now,
		LastLoginAt: &now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		config.ApiEndpoints = nil
		sqlDB, err := database.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	r := gin.New()
	auth := r.Group("/api", middleware.AuthMiddleware())
	auth.POST("/generate", GenerateImage)
	return r
}

func TestGenerateImageClampsNBeforeQuotaAndPersist(t *testing.T) {
	r := setupGenerateHandlerTest(t)
	token, err := service.SignToken("user-1", "user", config.App.JWTSecret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	input := map[string]interface{}{
		"taskId":        "task-n-clamp",
		"prompt":        "prompt",
		"inputImageIds": []string{},
		"codexCli":      false,
		"params": map[string]interface{}{
			"size":               "auto",
			"quality":            "auto",
			"output_format":      "png",
			"output_compression": nil,
			"moderation":         "auto",
			"n":                  999,
		},
	}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/api/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var task database.Task
	if err := database.DB.Where("id = ?", "task-n-clamp").First(&task).Error; err != nil {
		t.Fatalf("find task: %v", err)
	}
	var params service.TaskParams
	if err := json.Unmarshal([]byte(task.ParamsJSON), &params); err != nil {
		t.Fatalf("unmarshal params: %v", err)
	}
	if params.N != service.MaxTaskN {
		t.Fatalf("persisted n = %d, want %d", params.N, service.MaxTaskN)
	}
}
