package handler

import (
	"log/slog"
	"net/http"

	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

func AnnouncementPublic(c *gin.Context) {
	announcement, err := service.GetAnnouncement()
	if err != nil {
		slog.Error("获取公告失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取公告失败"})
		return
	}
	c.JSON(http.StatusOK, announcement)
}

func AdminGetAnnouncement(c *gin.Context) {
	announcement, err := service.GetAnnouncement()
	if err != nil {
		slog.Error("获取公告失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取公告失败"})
		return
	}
	c.JSON(http.StatusOK, announcement)
}

func AdminUpdateAnnouncement(c *gin.Context) {
	var body struct {
		Content string `json:"content"`
		Enabled bool   `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	announcement, err := service.UpdateAnnouncement(body.Content, body.Enabled)
	if err != nil {
		slog.Error("更新公告失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新公告失败"})
		return
	}
	c.JSON(http.StatusOK, announcement)
}
