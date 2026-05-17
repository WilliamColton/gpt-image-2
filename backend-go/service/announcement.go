package service

import (
	"time"

	"gpt-image-playground/backend/database"

	"gorm.io/gorm/clause"
)

const announcementID = "default"

type Announcement struct {
	Content   string `json:"content"`
	Enabled   bool   `json:"enabled"`
	UpdatedAt int64  `json:"updatedAt"`
}

func GetAnnouncement() (*Announcement, error) {
	var a database.Announcement
	if err := database.DB.Where("id = ?", announcementID).First(&a).Error; err != nil {
		return nil, err
	}
	return announcementFromModel(&a), nil
}

func UpdateAnnouncement(content string, enabled bool) (*Announcement, error) {
	now := time.Now().UnixMilli()
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	model := database.Announcement{
		ID:        announcementID,
		Content:   content,
		Enabled:   enabledInt,
		UpdatedAt: now,
	}
	if err := database.DB.Clauses(clause.OnConflict{UpdateAll: true}).Create(&model).Error; err != nil {
		return nil, err
	}
	return announcementFromModel(&model), nil
}

func announcementFromModel(a *database.Announcement) *Announcement {
	return &Announcement{
		Content:   a.Content,
		Enabled:   a.Enabled == 1,
		UpdatedAt: a.UpdatedAt,
	}
}
