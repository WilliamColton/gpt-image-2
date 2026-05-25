package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ""
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			return
		}
		sub, _, err := service.VerifyToken(token, config.App.JWTSecret)
		if err != nil {
			slog.Warn("JWT 验证失败", "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态无效"})
			return
		}
		user, err := service.FindUserByID(sub)
		if err != nil {
			slog.Warn("用户不存在", "user_id", sub, "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态无效"})
			return
		}
		if user.Status == "disabled" {
			slog.Warn("用户已被禁用", "user_id", sub)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态无效"})
			return
		}
		c.Set("user", &service.AuthUser{ID: user.ID, Label: user.Label, Role: user.Role})
		c.Set("userID", user.ID)
		c.Next()
	}
}

func GetAuthUser(c *gin.Context) *service.AuthUser {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	authUser, _ := user.(*service.AuthUser)
	return authUser
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := ""
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "请先登录管理后台"})
			return
		}
		sub, role, err := service.VerifyToken(token, config.App.JWTSecret)
		if err != nil || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无管理员权限"})
			return
		}

		// Verify admin exists and is active in DB
		user, err := service.FindUserByID(sub)
		if err != nil {
			slog.Warn("管理员用户不存在", "user_id", sub)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无管理员权限"})
			return
		}
		if user.Status == "disabled" {
			slog.Warn("管理员已被禁用", "user_id", sub)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无管理员权限"})
			return
		}
		if user.Role != "admin" {
			slog.Warn("非管理员角色尝试访问管理后台", "user_id", sub, "role", user.Role)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无管理员权限"})
			return
		}

		c.Set("adminUserID", user.ID)
		c.Next()
	}
}
