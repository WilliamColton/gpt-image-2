package service

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/util"

	"gorm.io/gorm"
)

const (
	changelogVersionMaxRunes = 64
	changelogTitleMaxRunes   = 100
	changelogContentMaxRunes = 20000
)

type ChangelogEntry struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Published   bool   `json:"published"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	PublishedAt *int64 `json:"publishedAt"`
}

type ChangelogEntryInput struct {
	Version   string
	Title     string
	Content   string
	Published bool
}

func ListChangelogEntries(includeDrafts bool) ([]ChangelogEntry, error) {
	var models []database.ChangelogEntry
	query := database.DB.Order("published_at DESC").Order("updated_at DESC").Order("created_at DESC")
	if !includeDrafts {
		query = query.Where("published = ?", 1)
	}
	if err := query.Find(&models).Error; err != nil {
		slog.Error("查询更新日志列表失败", "include_drafts", includeDrafts, "error", err)
		return nil, err
	}
	entries := make([]ChangelogEntry, len(models))
	for i := range models {
		entries[i] = *changelogEntryFromModel(&models[i])
	}
	return entries, nil
}

func GetLatestPublishedChangelog() (*ChangelogEntry, error) {
	var model database.ChangelogEntry
	err := database.DB.Where("published = ?", 1).Order("published_at DESC").Order("updated_at DESC").Order("created_at DESC").First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		slog.Error("查询最新更新日志失败", "error", err)
		return nil, err
	}
	return changelogEntryFromModel(&model), nil
}

func CreateChangelogEntry(input ChangelogEntryInput) (*ChangelogEntry, error) {
	version, title, content, err := normalizeChangelogInput(input)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	published := 0
	var publishedAt *int64
	if input.Published {
		published = 1
		publishedAt = &now
	}
	model := database.ChangelogEntry{
		ID:          util.GenerateID(),
		Version:     version,
		Title:       title,
		Content:     content,
		Published:   published,
		CreatedAt:   now,
		UpdatedAt:   now,
		PublishedAt: publishedAt,
	}
	if err := database.DB.Create(&model).Error; err != nil {
		slog.Error("创建更新日志失败", "version", version, "error", err)
		return nil, err
	}
	return changelogEntryFromModel(&model), nil
}

func UpdateChangelogEntry(id string, input ChangelogEntryInput) (*ChangelogEntry, error) {
	version, title, content, err := normalizeChangelogInput(input)
	if err != nil {
		return nil, err
	}
	var model database.ChangelogEntry
	if err := database.DB.Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("更新日志不存在")
		}
		slog.Error("查询更新日志失败", "changelog_id", id, "error", err)
		return nil, err
	}
	now := time.Now().UnixMilli()
	published := 0
	if input.Published {
		published = 1
		if model.Published == 0 {
			model.PublishedAt = &now
		}
	}
	model.Version = version
	model.Title = title
	model.Content = content
	model.Published = published
	model.UpdatedAt = now
	if err := database.DB.Save(&model).Error; err != nil {
		slog.Error("更新更新日志失败", "changelog_id", id, "error", err)
		return nil, err
	}
	return changelogEntryFromModel(&model), nil
}

func DeleteChangelogEntry(id string) error {
	result := database.DB.Where("id = ?", id).Delete(&database.ChangelogEntry{})
	if result.Error != nil {
		slog.Error("删除更新日志失败", "changelog_id", id, "error", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("更新日志不存在")
	}
	return nil
}

func normalizeChangelogInput(input ChangelogEntryInput) (string, string, string, error) {
	version := strings.TrimSpace(input.Version)
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if input.Published && version == "" {
		return "", "", "", fmt.Errorf("发布更新日志前请填写版本号")
	}
	if utf8.RuneCountInString(version) > changelogVersionMaxRunes {
		return "", "", "", fmt.Errorf("版本号最多 %d 字", changelogVersionMaxRunes)
	}
	if utf8.RuneCountInString(title) > changelogTitleMaxRunes {
		return "", "", "", fmt.Errorf("标题最多 %d 字", changelogTitleMaxRunes)
	}
	if utf8.RuneCountInString(content) > changelogContentMaxRunes {
		return "", "", "", fmt.Errorf("更新日志内容最多 %d 字", changelogContentMaxRunes)
	}
	return version, title, content, nil
}

func changelogEntryFromModel(m *database.ChangelogEntry) *ChangelogEntry {
	return &ChangelogEntry{
		ID:          m.ID,
		Version:     m.Version,
		Title:       m.Title,
		Content:     m.Content,
		Published:   m.Published == 1,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		PublishedAt: m.PublishedAt,
	}
}
