package handler

import (
	"net/http"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

func ConfigPublic(c *gin.Context) {
	cfg := service.AppConfig{
		CodexCLI:      config.App.CodexCLI,
		APIMode:       config.App.APIMode,
		Model:         config.App.Model,
		Timeout:       config.App.Timeout,
		InviteEnabled: config.App.InviteEnabled,
	}
	c.JSON(http.StatusOK, cfg)
}
