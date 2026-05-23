package handler

import (
	"log/slog"
	"net/http"

	"gpt-image-playground/backend/middleware"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AuthLogin(c *gin.Context) {
	var body struct {
		Code string `json:"code"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入兑换码"})
		return
	}
	token, user, err := service.LoginWithCode(body.Code)
	if err != nil {
		slog.Warn("登录失败", "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func AuthRedeem(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	var body struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入兑换码"})
		return
	}
	if err := service.RedeemForUser(user.ID, body.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Return updated user info
	u, err := service.FindUserByID(user.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "quota": u.Quota, "usedCount": u.UsedCount})
}

func AuthMe(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	if user != nil {
		user.ImageCount = service.CountGeneratedImages(user.ID)
		u, err := service.FindUserByID(user.ID)
		if err == nil {
			user.Quota = u.Quota
			user.UsedCount = u.UsedCount
		}
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// --- New handlers (stubs for RED phase) ---

func AuthLoginPassword(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func AuthRegister(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func AuthMigrate(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func AuthChangePassword(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func AuthSetInviteCode(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

func AuthGetInviteCode(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
