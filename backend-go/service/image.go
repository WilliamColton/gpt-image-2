package service

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/util"
)

var mimeExt = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/webp": "webp",
}

var dataURLRegex = regexp.MustCompile(`^data:([^;]+);base64,(.+)$`)

func DataURLToBuffer(dataURL string) ([]byte, string, error) {
	match := dataURLRegex.FindStringSubmatch(dataURL)
	if match == nil {
		return nil, "", fmt.Errorf("图片 dataUrl 格式无效")
	}
	buf, err := base64.StdEncoding.DecodeString(match[2])
	if err != nil {
		return nil, "", fmt.Errorf("图片 dataUrl 格式无效")
	}
	return buf, match[1], nil
}

func BufferToDataURL(buf []byte, mime string) string {
	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(buf))
}

func SaveImageBuffer(userID string, buf []byte, mime, source string) (*Image, error) {
	sha256 := util.Sha256Buffer(buf)

	// Check for existing image with same hash
	var existing database.Image
	err := database.DB.Where("user_id = ? AND sha256 = ?", userID, sha256).First(&existing).Error
	if err == nil {
		return &Image{ID: existing.ID, Mime: existing.Mime, Size: existing.Size, Source: existing.Source, CreatedAt: existing.CreatedAt}, nil
	}

	id := util.GenerateID()
	ext := mimeExt[mime]
	if ext == "" {
		ext = "png"
	}
	dir := util.EnsureUserUploadDir(config.App.UploadDir, userID)
	absPath := filepath.Join(dir, fmt.Sprintf("%s.%s", id, ext))
	if err := os.WriteFile(absPath, buf, 0644); err != nil {
		slog.Error("写入图片文件失败", "user_id", userID, "path", absPath, "error", err)
		return nil, fmt.Errorf("图片保存失败")
	}
	now := time.Now().UnixMilli()
	relPath := util.ToUploadRelativePath(config.App.UploadDir, absPath)

	img := &database.Image{
		ID:        id,
		UserID:    userID,
		FilePath:  relPath,
		Mime:      mime,
		Size:      int64(len(buf)),
		Sha256:    sha256,
		Source:    source,
		CreatedAt: now,
	}
	if err := database.DB.Create(img).Error; err != nil {
		slog.Error("保存图片记录失败", "user_id", userID, "image_id", id, "error", err)
		return nil, fmt.Errorf("图片保存失败")
	}
	return &Image{ID: img.ID, Mime: img.Mime, Size: img.Size, Source: img.Source, CreatedAt: img.CreatedAt}, nil
}

func SaveDataURLImage(userID, dataURL, source string) (*Image, error) {
	buf, mime, err := DataURLToBuffer(dataURL)
	if err != nil {
		return nil, err
	}
	return SaveImageBuffer(userID, buf, mime, source)
}

func GetImageForUser(userID, imageID string) (*Image, error) {
	var img database.Image
	err := database.DB.Where("id = ? AND user_id = ?", imageID, userID).First(&img).Error
	if err != nil {
		slog.Error("查询图片失败", "user_id", userID, "image_id", imageID, "error", err)
		return nil, err
	}
	return &Image{ID: img.ID, UserID: img.UserID, FilePath: img.FilePath, Mime: img.Mime, Size: img.Size, Sha256: img.Sha256, Source: img.Source, CreatedAt: img.CreatedAt}, nil
}

func ReadImageDataURLForUser(userID, imageID string) (string, error) {
	img, err := GetImageForUser(userID, imageID)
	if err != nil {
		return "", err
	}
	absPath, err := util.ResolveUploadPath(config.App.UploadDir, img.FilePath)
	if err != nil {
		return "", err
	}
	buf, err := os.ReadFile(absPath)
	if err != nil {
		slog.Error("读取图片文件失败", "user_id", userID, "image_id", imageID, "path", absPath, "error", err)
		return "", err
	}
	return BufferToDataURL(buf, img.Mime), nil
}

func ReadImageFileForUser(userID, imageID string) (image *Image, filePath string, err error) {
	img, err := GetImageForUser(userID, imageID)
	if err != nil {
		return nil, "", err
	}
	absPath, err := util.ResolveUploadPath(config.App.UploadDir, img.FilePath)
	if err != nil {
		return nil, "", err
	}
	return img, absPath, nil
}

func DeleteImageForUser(userID, imageID string) {
	img, err := GetImageForUser(userID, imageID)
	if err != nil {
		return
	}
	if err := database.DB.Where("id = ? AND user_id = ?", imageID, userID).Delete(&database.Image{}).Error; err != nil {
		slog.Error("删除图片记录失败", "user_id", userID, "image_id", imageID, "error", err)
		return
	}
	absPath, err := util.ResolveUploadPath(config.App.UploadDir, img.FilePath)
	if err == nil {
		os.Remove(absPath)
	}
}

func ParseImageSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case "mask":
		return "mask"
	case "generated":
		return "generated"
	default:
		return "upload"
	}
}
