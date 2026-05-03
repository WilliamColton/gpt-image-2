package util

import (
	"os"
	"path/filepath"
	"strings"
)

func EnsureRuntimeDirs(dataDir, uploadDir string) {
	os.MkdirAll(dataDir, 0755)
	os.MkdirAll(uploadDir, 0755)
}

func GetUserUploadDir(uploadDir, userID string) string {
	return filepath.Join(uploadDir, userID)
}

func EnsureUserUploadDir(uploadDir, userID string) string {
	dir := GetUserUploadDir(uploadDir, userID)
	os.MkdirAll(dir, 0755)
	return dir
}

func ToUploadRelativePath(uploadDir, filePath string) string {
	rel, _ := filepath.Rel(uploadDir, filePath)
	return strings.ReplaceAll(rel, "\\", "/")
}

func ResolveUploadPath(uploadDir, relativePath string) (string, error) {
	absPath, err := filepath.Abs(filepath.Join(uploadDir, relativePath))
	if err != nil {
		return "", err
	}
	absUpload, err := filepath.Abs(uploadDir)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(absPath, absUpload+string(filepath.Separator)) && absPath != absUpload {
		return "", os.ErrNotExist
	}
	return absPath, nil
}
