package handler

import (
	"log/slog"
	"net/http"
	"strings"

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
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user, "needsMigration": user.NeedsMigration})
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
			user.Username = u.Username
			user.NeedsMigration = u.PasswordHash == nil
		}
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// --- AuthLoginPassword ---

func AuthLoginPassword(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Username == "" || body.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入用户名和密码"})
		return
	}
	token, user, needsMigration, err := service.LoginWithPassword(body.Username, body.Password)
	if err != nil {
		slog.Warn("密码登录失败", "username", body.Username, "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user, "needsMigration": needsMigration})
}

// --- AuthRegister ---

func AuthRegister(c *gin.Context) {
	var body struct {
		InviteCode string `json:"inviteCode"`
		Username   string `json:"username"`
		Password   string `json:"password"`
	}
	_ = c.ShouldBindJSON(&body)
	if len([]rune(body.Username)) < 3 || len([]rune(body.Username)) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名须为 3-20 个字符"})
		return
	}
	if len(body.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少需要 8 个字符"})
		return
	}
	token, user, err := service.RegisterUser(body.Username, body.Password, body.InviteCode)
	if err != nil {
		slog.Warn("注册失败", "username", body.Username, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

// --- AuthMigrate ---

func AuthMigrate(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	var body struct {
		Username        string `json:"username"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Password != body.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "两次输入的密码不一致"})
		return
	}
	if len([]rune(body.Username)) < 3 || len([]rune(body.Username)) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名须为 3-20 个字符"})
		return
	}
	if len(body.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少需要 8 个字符"})
		return
	}
	updatedUser, err := service.MigrateUser(user.ID, body.Username, body.Password)
	if err != nil {
		slog.Warn("迁移失败", "user_id", user.ID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": updatedUser})
}

// --- AuthChangePassword ---

func AuthChangePassword(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	var body struct {
		OldPassword     string `json:"oldPassword"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.NewPassword != body.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "两次输入的密码不一致"})
		return
	}
	if len(body.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少需要 8 个字符"})
		return
	}
	if err := service.ChangePassword(user.ID, body.OldPassword, body.NewPassword); err != nil {
		slog.Warn("修改密码失败", "user_id", user.ID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// --- AuthSetInviteCode ---

func AuthSetInviteCode(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	var body struct {
		Code string `json:"code"`
	}
	_ = c.ShouldBindJSON(&body)
	if strings.TrimSpace(body.Code) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邀请码不能为空"})
		return
	}
	if err := service.SetInviteCode(user.ID, body.Code); err != nil {
		slog.Warn("设置邀请码失败", "user_id", user.ID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// --- AuthGetInviteCode ---

func AuthGetInviteCode(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	code, setAt, err := service.GetInviteCode(user.ID)
	if err != nil {
		slog.Warn("获取邀请码失败", "user_id", user.ID, "error", err)
		c.JSON(http.StatusOK, gin.H{"code": nil, "setAt": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "setAt": setAt})
}
