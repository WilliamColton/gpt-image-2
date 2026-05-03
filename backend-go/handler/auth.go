package handler

import (
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
		Apikey string `json:"apikey"`
	}
	_ = c.ShouldBindJSON(&body)
	token, user, err := service.LoginWithApikey(body.Apikey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func AuthMe(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	if user != nil {
		user.ImageCount = service.CountGeneratedImages(user.ID)
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}
