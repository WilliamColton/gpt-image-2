package handler

import (
	"log/slog"
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
		slog.Warn("管理员密钥错误")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "管理员密钥错误"})
		return
	}
	// Issue JWT with role=admin. Use a fixed admin user ID.
	token, err := service.SignToken("admin", "admin", config.App.JWTSecret)
	if err != nil {
		slog.Error("管理员 JWT 签发失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func AdminListUsers(c *gin.Context) {
	users, err := service.ListAllUsers()
	if err != nil {
		slog.Error("获取用户列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func AdminUpdateQuota(c *gin.Context) {
	userID := c.Param("id")
	var body struct {
		Mode           string `json:"mode"` // "delta" (default) or "set"
		Delta          int    `json:"delta"`
		ResetUsedCount bool   `json:"resetUsedCount"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	if body.Mode == "set" {
		if err := service.SetUserQuotaAbs(userID, body.Delta); err != nil {
			slog.Error("设置配额失败", "user_id", userID, "quota", body.Delta, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "设置配额失败"})
			return
		}
	} else {
		if err := service.UpdateUserQuota(userID, body.Delta, body.ResetUsedCount); err != nil {
			slog.Error("更新配额失败", "user_id", userID, "delta", body.Delta, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新配额失败"})
			return
		}
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
		slog.Error("更新状态失败", "user_id", userID, "status", body.Status, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminDeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if err := service.DeleteUser(userID); err != nil {
		slog.Error("删除用户失败", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminCreateCode(c *gin.Context) {
	var body struct {
		Quota int `json:"quota"`
		Count int `json:"count"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if body.Count <= 0 {
		body.Count = 1
	}
	if body.Count > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "单次最多创建 100 个兑换码"})
		return
	}
	var codes []service.RedemptionCode
	for i := 0; i < body.Count; i++ {
		rc, err := service.CreateRedemptionCode(body.Quota)
		if err != nil {
			slog.Error("创建兑换码失败", "quota", body.Quota, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		codes = append(codes, *rc)
	}
	c.JSON(http.StatusOK, gin.H{"codes": codes})
}

func AdminListCodes(c *gin.Context) {
	codes, err := service.ListRedemptionCodes()
	if err != nil {
		slog.Error("获取兑换码列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取兑换码列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"codes": codes})
}
