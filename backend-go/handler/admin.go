package handler

import (
	"net/http"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

func AdminLogin(c *gin.Context) {
	var body struct {
		Apikey string `json:"apikey"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Apikey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入管理员密钥"})
		return
	}
	if body.Apikey != config.App.AdminApikey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "管理员密钥错误"})
		return
	}
	// Issue JWT with role=admin. Use a fixed admin user ID.
	token, err := service.SignToken("admin", "admin", config.App.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func AdminListUsers(c *gin.Context) {
	users, err := service.ListAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func AdminUpdateQuota(c *gin.Context) {
	userID := c.Param("id")
	var body struct {
		Delta          int  `json:"delta"`
		ResetUsedCount bool `json:"resetUsedCount"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if err := service.UpdateUserQuota(userID, body.Delta, body.ResetUsedCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新配额失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminToggleStatus(c *gin.Context) {
	userID := c.Param("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if body.Status != "active" && body.Status != "disabled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "状态值无效"})
		return
	}
	if err := service.SetUserStatus(userID, body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
