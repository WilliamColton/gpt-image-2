package handler

import (
	"net/http"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

func ConfigPublic(c *gin.Context) {
	cfg := service.AppConfig{
		BaseURL:  config.App.Defaults.BaseURL,
		CodexCLI: config.App.Defaults.CodexCLI,
		APIMode:  config.App.Defaults.APIMode,
		Model:    config.App.Defaults.Model,
		Timeout:  config.App.Defaults.Timeout,
	}
	c.JSON(http.StatusOK, cfg)
}
