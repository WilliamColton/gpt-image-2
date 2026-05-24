package handler

import (
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"gpt-image-playground/backend/middleware"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

func FeedbackCreate(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	var body struct {
		Category string `json:"category"`
		Content  string `json:"content"`
		Contact  string `json:"contact"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	category := strings.TrimSpace(body.Category)
	content := strings.TrimSpace(body.Content)
	contact := strings.TrimSpace(body.Contact)
	if category == "" {
		category = service.BugFeedbackCategoryBug
	}

	if !service.IsValidBugFeedbackCategory(category) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "反馈分类无效"})
		return
	}
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请填写问题描述"})
		return
	}
	if utf8.RuneCountInString(content) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "问题描述最多 2000 字"})
		return
	}
	if utf8.RuneCountInString(contact) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "联系方式最多 200 字"})
		return
	}

	feedback, err := service.CreateBugFeedback(service.CreateBugFeedbackInput{
		UserID:    user.ID,
		UserLabel: user.Label,
		Category:  category,
		Content:   content,
		Contact:   contact,
	})
	if err != nil {
		slog.Error("提交反馈失败", "user_id", user.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交反馈失败"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"feedback": feedback})
}

func AdminListFeedbacks(c *gin.Context) {
	status := strings.TrimSpace(c.Query("status"))
	if status != "" && !service.IsValidBugFeedbackStatus(status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "反馈状态无效"})
		return
	}
	feedbacks, err := service.ListBugFeedbacks(status)
	if err != nil {
		slog.Error("获取反馈列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取反馈列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"feedbacks": feedbacks})
}

func AdminUpdateFeedbackStatus(c *gin.Context) {
	feedbackID := c.Param("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	status := strings.TrimSpace(body.Status)
	if !service.IsValidBugFeedbackStatus(status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "反馈状态无效"})
		return
	}

	feedback, err := service.UpdateBugFeedbackStatus(feedbackID, status)
	if err != nil {
		if err.Error() == "反馈不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": "反馈不存在"})
			return
		}
		slog.Error("更新反馈状态失败", "feedback_id", feedbackID, "status", status, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新反馈状态失败"})
		return
	}
	c.JSON(http.StatusOK, feedback)
}
