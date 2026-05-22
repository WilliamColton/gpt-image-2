package database

import (
	"fmt"
	"path/filepath"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/util"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() error {
	dbPath := filepath.Join(config.App.DataDir, "app.sqlite")
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_foreign_keys=ON"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	sqlDB, _ := DB.DB()
	sqlDB.SetMaxOpenConns(1)

	if err := DB.AutoMigrate(&User{}, &RedemptionCode{}, &Image{}, &Task{}, &Announcement{}, &Feedback{}, &ChangelogEntry{}); err != nil {
		return fmt.Errorf("建表失败: %w", err)
	}

	if err := initAdmin(); err != nil {
		return err
	}
	if err := initAnnouncement(); err != nil {
		return err
	}
	return nil
}

func initAdmin() error {
	var count int64
	DB.Model(&User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return nil
	}
	admin := &User{
		ID:        util.GenerateID(),
		Label:     "admin",
		Role:      "admin",
		Status:    "active",
		CreatedAt: time.Now().UnixMilli(),
	}
	if err := DB.Create(admin).Error; err != nil {
		return fmt.Errorf("创建管理员失败: %w", err)
	}
	return nil
}

func initAnnouncement() error {
	var count int64
	DB.Model(&Announcement{}).Where("id = ?", "default").Count(&count)
	if count > 0 {
		return nil
	}
	announcement := &Announcement{
		ID:        "default",
		Content:   "",
		Enabled:   0,
		UpdatedAt: time.Now().UnixMilli(),
	}
	if err := DB.Create(announcement).Error; err != nil {
		return fmt.Errorf("创建默认公告失败: %w", err)
	}
	return nil
}
