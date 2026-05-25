package service

import (
	"os"
	"path/filepath"
	"testing"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupImageServiceTest(t *testing.T) string {
	t.Helper()

	tmp := t.TempDir()
	config.App = &config.Config{
		DataDir:     filepath.Join(tmp, "data"),
		UploadDir:   filepath.Join(tmp, "upload"),
		JWTSecret:   "test-secret",
		AdminApikey: "test-admin",
	}
	if err := os.MkdirAll(config.App.DataDir, 0755); err != nil {
		t.Fatalf("创建临时数据目录失败: %v", err)
	}
	if err := os.MkdirAll(config.App.UploadDir, 0755); err != nil {
		t.Fatalf("创建临时上传目录失败: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(filepath.Join(config.App.DataDir, "test.sqlite")), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	database.DB = db
	if err := database.DB.AutoMigrate(&database.Image{}); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := database.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return tmp
}

func TestSaveImageBufferReusesSameUserDuplicate(t *testing.T) {
	setupImageServiceTest(t)

	buf := []byte("same image bytes")
	first, err := SaveImageBuffer("user-1", buf, "image/png", "upload")
	if err != nil {
		t.Fatalf("首次保存图片失败: %v", err)
	}
	second, err := SaveImageBuffer("user-1", buf, "image/png", "generated")
	if err != nil {
		t.Fatalf("重复保存图片失败: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("同一用户相同图片应复用 id，got %q and %q", first.ID, second.ID)
	}

	var count int64
	if err := database.DB.Model(&database.Image{}).Where("user_id = ?", "user-1").Count(&count).Error; err != nil {
		t.Fatalf("统计图片记录失败: %v", err)
	}
	if count != 1 {
		t.Fatalf("同一用户相同图片应只有一条记录，got %d", count)
	}

	var row database.Image
	if err := database.DB.Where("id = ?", first.ID).First(&row).Error; err != nil {
		t.Fatalf("查询图片记录失败: %v", err)
	}
	if row.Mime != "image/png" || row.Size != int64(len(buf)) || row.Source != "upload" || row.Sha256 == "" {
		t.Fatalf("图片元数据不符合预期: %+v", row)
	}

	absPath, err := filepath.Abs(filepath.Join(config.App.UploadDir, row.FilePath))
	if err != nil {
		t.Fatalf("解析图片路径失败: %v", err)
	}
	if _, err := os.Stat(absPath); err != nil {
		t.Fatalf("图片文件应已写入磁盘: %v", err)
	}
}

func TestSaveImageBufferRemovesFileWhenDBCreateFails(t *testing.T) {
	setupImageServiceTest(t)
	buf := []byte("orphan cleanup bytes")

	sqlDB, err := database.DB.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close sql db: %v", err)
	}

	if _, err := SaveImageBuffer("user-1", buf, "image/png", "upload"); err == nil {
		t.Fatal("DB 关闭后保存图片应失败")
	}

	userDir := filepath.Join(config.App.UploadDir, "user-1")
	entries, err := os.ReadDir(userDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("read user upload dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("DB 保存失败后不应遗留图片文件，got %d", len(entries))
	}
}

func TestSaveImageBufferDoesNotReuseAcrossUsers(t *testing.T) {
	setupImageServiceTest(t)

	buf := []byte("same image bytes")
	first, err := SaveImageBuffer("user-1", buf, "image/png", "upload")
	if err != nil {
		t.Fatalf("保存 user-1 图片失败: %v", err)
	}
	second, err := SaveImageBuffer("user-2", buf, "image/png", "upload")
	if err != nil {
		t.Fatalf("保存 user-2 图片失败: %v", err)
	}

	if first.ID == second.ID {
		t.Fatalf("不同用户相同图片不应复用同一 id: %q", first.ID)
	}

	var count int64
	if err := database.DB.Model(&database.Image{}).Count(&count).Error; err != nil {
		t.Fatalf("统计图片记录失败: %v", err)
	}
	if count != 2 {
		t.Fatalf("不同用户相同图片应有两条记录，got %d", count)
	}
}
