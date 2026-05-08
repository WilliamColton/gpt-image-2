package middleware

import (
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态无效"})
			return
		}
		user, err := service.FindUserByID(sub)
		if err != nil || user.Status == "disabled" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态无效"})
			return
		}
		c.Set("user", &service.AuthUser{ID: user.ID, Label: user.Label, Role: user.Role})
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
		_, role, err := service.VerifyToken(token, config.App.JWTSecret)
		if err != nil || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无管理员权限"})
			return
		}
		c.Next()
	}
}
