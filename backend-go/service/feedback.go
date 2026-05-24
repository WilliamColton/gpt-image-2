package service

import (
	"fmt"
	"log/slog"
	"time"

	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/util"

	"gorm.io/gorm"
)

const (
	BugFeedbackCategoryBug     = "bug"
	BugFeedbackCategoryFeature = "feature"
	BugFeedbackStatusOpen      = "open"
	BugFeedbackStatusReviewing = "reviewing"
	BugFeedbackStatusResolved  = "resolved"
)

type CreateBugFeedbackInput struct {
	UserID    string
	UserLabel string
	Category  string
	Content   string
	Contact   string
}

type BugFeedback struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	UserLabel string `json:"userLabel"`
	Category  string `json:"category"`
	Content   string `json:"content"`
	Contact   string `json:"contact"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

func IsValidBugFeedbackCategory(category string) bool {
	return category == BugFeedbackCategoryBug || category == BugFeedbackCategoryFeature
}

func IsValidBugFeedbackStatus(status string) bool {
	return status == BugFeedbackStatusOpen || status == BugFeedbackStatusReviewing || status == BugFeedbackStatusResolved
}

func CreateBugFeedback(input CreateBugFeedbackInput) (*BugFeedback, error) {
	now := time.Now().UnixMilli()
	model := database.Feedback{
		ID:        util.GenerateID(),
		UserID:    input.UserID,
		UserLabel: input.UserLabel,
		Category:  input.Category,
		Content:   input.Content,
		Contact:   input.Contact,
		Status:    BugFeedbackStatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := database.DB.Create(&model).Error; err != nil {
		slog.Error("保存反馈失败", "user_id", input.UserID, "error", err)
		return nil, err
	}
	return bugFeedbackFromModel(&model), nil
}

func ListBugFeedbacks(status string) ([]BugFeedback, error) {
	var models []database.Feedback
	query := database.DB.Order("created_at DESC")
	if status != "" {
		if !IsValidBugFeedbackStatus(status) {
			return nil, fmt.Errorf("反馈状态无效")
		}
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&models).Error; err != nil {
		slog.Error("查询反馈列表失败", "status", status, "error", err)
		return nil, err
	}
	feedbacks := make([]BugFeedback, len(models))
	for i := range models {
		feedbacks[i] = *bugFeedbackFromModel(&models[i])
	}
	return feedbacks, nil
}

func UpdateBugFeedbackStatus(id string, status string) (*BugFeedback, error) {
	if !IsValidBugFeedbackStatus(status) {
		return nil, fmt.Errorf("反馈状态无效")
	}
	var model database.Feedback
	if err := database.DB.Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("反馈不存在")
		}
		slog.Error("查询反馈失败", "feedback_id", id, "error", err)
		return nil, err
	}
	now := time.Now().UnixMilli()
	if err := database.DB.Model(&model).Updates(map[string]interface{}{
		"status":     status,
		"updated_at": now,
	}).Error; err != nil {
		slog.Error("更新反馈状态失败", "feedback_id", id, "status", status, "error", err)
		return nil, err
	}
	model.Status = status
	model.UpdatedAt = now
	return bugFeedbackFromModel(&model), nil
}

func bugFeedbackFromModel(m *database.Feedback) *BugFeedback {
	return &BugFeedback{
		ID:        m.ID,
		UserID:    m.UserID,
		UserLabel: m.UserLabel,
		Category:  m.Category,
		Content:   m.Content,
		Contact:   m.Contact,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
