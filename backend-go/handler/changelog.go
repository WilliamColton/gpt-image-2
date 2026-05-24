package handler

import (
	"log/slog"
	"net/http"

	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

type changelogBody struct {
	Version   string `json:"version"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Published bool   `json:"published"`
}

func ChangelogLatestPublic(c *gin.Context) {
	entry, err := service.GetLatestPublishedChangelog()
	if err != nil {
		slog.Error("获取最新更新日志失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取更新日志失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"changelog": entry})
}

func ChangelogListPublic(c *gin.Context) {
	entries, err := service.ListChangelogEntries(false)
	if err != nil {
		slog.Error("获取更新日志列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取更新日志列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"changelogs": entries})
}

func AdminListChangelogs(c *gin.Context) {
	entries, err := service.ListChangelogEntries(true)
	if err != nil {
		slog.Error("获取更新日志列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取更新日志列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"changelogs": entries})
}

func AdminCreateChangelog(c *gin.Context) {
	var body changelogBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	entry, err := service.CreateChangelogEntry(service.ChangelogEntryInput{
		Version:   body.Version,
		Title:     body.Title,
		Content:   body.Content,
		Published: body.Published,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, entry)
}

func AdminUpdateChangelog(c *gin.Context) {
	var body changelogBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	entry, err := service.UpdateChangelogEntry(c.Param("id"), service.ChangelogEntryInput{
		Version:   body.Version,
		Title:     body.Title,
		Content:   body.Content,
		Published: body.Published,
	})
	if err != nil {
		if err.Error() == "更新日志不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": "更新日志不存在"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entry)
}

func AdminDeleteChangelog(c *gin.Context) {
	if err := service.DeleteChangelogEntry(c.Param("id")); err != nil {
		if err.Error() == "更新日志不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": "更新日志不存在"})
			return
		}
		slog.Error("删除更新日志失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除更新日志失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
