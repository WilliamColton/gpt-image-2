package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"gpt-image-playground/backend/middleware"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

func ImagesUpload(c *gin.Context) {
	user := middleware.GetAuthUser(c)

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传图片"})
		return
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取图片失败"})
		return
	}

	mime := header.Header.Get("Content-Type")
	if mime == "" {
		mime = "image/png"
	}

	source := service.ParseImageSource(c.PostForm("source"))

	img, err := service.SaveImageBuffer(user.ID, buf, mime, source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        img.ID,
		"url":       fmt.Sprintf("/api/images/%s", img.ID),
		"createdAt": img.CreatedAt,
		"source":    img.Source,
	})
}

func ImagesGet(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	imageID := c.Param("id")

	img, filePath, err := service.ReadImageFileForUser(user.ID, imageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在"})
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取图片失败"})
		return
	}

	c.Data(http.StatusOK, img.Mime, data)
}

func ImagesDelete(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	imageID := c.Param("id")
	service.DeleteImageForUser(user.ID, imageID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
