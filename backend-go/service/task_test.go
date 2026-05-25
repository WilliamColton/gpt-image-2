package service

import (
	"testing"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTaskServiceTest(t *testing.T) {
	t.Helper()
	config.App = &config.Config{JWTSecret: "test-secret"}
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	database.DB = db
	if err := database.DB.AutoMigrate(&database.Task{}); err != nil {
		t.Fatalf("AutoMigrate Task: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := database.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
}

func TestCountPendingImagesClampsPersistedN(t *testing.T) {
	setupTaskServiceTest(t)
	now := time.Now().UnixMilli()
	tasks := []database.Task{
		{
			ID:                 "task-huge",
			UserID:             "user-1",
			Prompt:             "prompt",
			ParamsJSON:         `{"n":999}`,
			InputImageIDsJSON:  `[]`,
			OutputImageIDsJSON: `[]`,
			Status:             "queued",
			CreatedAt:          now,
		},
		{
			ID:                 "task-zero",
			UserID:             "user-1",
			Prompt:             "prompt",
			ParamsJSON:         `{"n":0}`,
			InputImageIDsJSON:  `[]`,
			OutputImageIDsJSON: `[]`,
			Status:             "running",
			CreatedAt:          now,
		},
	}
	if err := database.DB.Create(&tasks).Error; err != nil {
		t.Fatalf("create tasks: %v", err)
	}

	if got := CountPendingImages("user-1"); got != MaxTaskN+1 {
		t.Fatalf("CountPendingImages = %d, want %d", got, MaxTaskN+1)
	}
}
