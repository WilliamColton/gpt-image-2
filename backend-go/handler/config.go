package handler

import (
	"net/http"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

func ConfigPublic(c *gin.Context) {
	// With multi-endpoint support, BaseURL is no longer a single defaults field.
	// Use the first endpoint's BaseURL for backward-compatible config response.
	baseURL := ""
	endpoints := config.App.GetEndpointPool()
	if len(endpoints) > 0 {
		baseURL = endpoints[0].BaseURL
	}
	cfg := service.AppConfig{
		BaseURL:  baseURL,
		CodexCLI: config.App.Defaults.CodexCLI,
		APIMode:  config.App.Defaults.APIMode,
		Model:    config.App.Defaults.Model,
		Timeout:  config.App.Defaults.Timeout,
	}
	c.JSON(http.StatusOK, cfg)
}
