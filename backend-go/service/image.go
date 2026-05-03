package service

import (
	"encoding/base64"
	"fmt"
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

	idRow := &struct {
		ID        string
		FilePath  string
		Mime      string
		Size      int64
		Source    string
		CreatedAt int64
	}{}
	err := database.DB.QueryRow("SELECT id, file_path, mime, size, source, created_at FROM images WHERE user_id = ? AND sha256 = ? LIMIT 1", userID, sha256).
		Scan(&idRow.ID, &idRow.FilePath, &idRow.Mime, &idRow.Size, &idRow.Source, &idRow.CreatedAt)
	if err == nil {
		return &Image{ID: idRow.ID, Mime: idRow.Mime, Size: idRow.Size, Source: idRow.Source, CreatedAt: idRow.CreatedAt}, nil
	}

	id := util.GenerateID()
	ext := mimeExt[mime]
	if ext == "" {
		ext = "png"
	}
	dir := util.EnsureUserUploadDir(config.App.UploadDir, userID)
	absPath := filepath.Join(dir, fmt.Sprintf("%s.%s", id, ext))
	if err := os.WriteFile(absPath, buf, 0644); err != nil {
		return nil, fmt.Errorf("图片保存失败")
	}
	now := time.Now().UnixMilli()
	relPath := util.ToUploadRelativePath(config.App.UploadDir, absPath)

	_, err = database.DB.Exec(`
		INSERT INTO images (id, user_id, file_path, mime, size, sha256, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, id, userID, relPath, mime, len(buf), sha256, source, now)
	if err != nil {
		return nil, fmt.Errorf("图片保存失败")
	}
	img, err := GetImageForUser(userID, id)
	if err != nil {
		return nil, fmt.Errorf("图片保存失败")
	}
	return img, nil
}

func SaveDataURLImage(userID, dataURL, source string) (*Image, error) {
	buf, mime, err := DataURLToBuffer(dataURL)
	if err != nil {
		return nil, err
	}
	return SaveImageBuffer(userID, buf, mime, source)
}

func GetImageForUser(userID, imageID string) (*Image, error) {
	img := &Image{UserID: userID}
	err := database.DB.QueryRow("SELECT id, file_path, mime, size, sha256, source, created_at FROM images WHERE id = ? AND user_id = ?", imageID, userID).
		Scan(&img.ID, &img.FilePath, &img.Mime, &img.Size, &img.Sha256, &img.Source, &img.CreatedAt)
	if err != nil {
		return nil, err
	}
	return img, nil
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
	database.DB.Exec("DELETE FROM images WHERE id = ? AND user_id = ?", imageID, userID)
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
