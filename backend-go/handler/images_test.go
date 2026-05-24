package handler

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/middleware"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupImagesHandlerTest(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
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
	if err := database.DB.AutoMigrate(&database.User{}, &database.Image{}); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}
	users := []database.User{
		{ID: "user-1", Label: "user 1", Role: "user", Status: "active", CreatedAt: time.Now().UnixMilli()},
		{ID: "user-2", Label: "user 2", Role: "user", Status: "active", CreatedAt: time.Now().UnixMilli()},
	}
	if err := database.DB.Create(&users).Error; err != nil {
		t.Fatalf("创建测试用户失败: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := database.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	r := gin.New()
	images := r.Group("/api/images", middleware.AuthMiddleware())
	images.POST("", ImagesUpload)
	images.GET("/:id", ImagesGet)
	images.DELETE("/:id", ImagesDelete)
	return r
}

func tokenForTestUser(t *testing.T, userID string) string {
	t.Helper()
	token, err := service.SignToken(userID, "user", config.App.JWTSecret)
	if err != nil {
		t.Fatalf("签发测试 token 失败: %v", err)
	}
	return token
}

func uploadTestImage(t *testing.T, r http.Handler, token string, name string, body []byte, contentType string) string {
	t.Helper()

	var form bytes.Buffer
	writer := multipart.NewWriter(&form)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, name))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("创建 multipart 文件失败: %v", err)
	}
	if _, err := part.Write(body); err != nil {
		t.Fatalf("写入 multipart 文件失败: %v", err)
	}
	if err := writer.WriteField("source", "upload"); err != nil {
		t.Fatalf("写入 source 字段失败: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("关闭 multipart writer 失败: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/images", &form)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("上传图片应成功，status=%d body=%s", resp.Code, resp.Body.String())
	}

	id := extractJSONID(resp.Body.String())
	if id == "" {
		t.Fatalf("上传响应应包含 id: %s", resp.Body.String())
	}
	return id
}

func extractJSONID(body string) string {
	marker := `"id":"`
	start := strings.Index(body, marker)
	if start < 0 {
		return ""
	}
	start += len(marker)
	end := strings.Index(body[start:], `"`)
	if end < 0 {
		return ""
	}
	return body[start : start+end]
}

func TestImagesGetReturnsImageForOwner(t *testing.T) {
	r := setupImagesHandlerTest(t)
	token := tokenForTestUser(t, "user-1")
	id := uploadTestImage(t, r, token, "image.png", []byte("png bytes"), "image/png")

	req := httptest.NewRequest(http.MethodGet, "/api/images/"+id+"?token="+token, nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("读取自己的图片应成功，status=%d body=%s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("Content-Type 应来自上传文件头，got %q", got)
	}
	if got := resp.Header().Get("Content-Length"); got != "9" {
		t.Fatalf("Content-Length 应为图片大小，got %q", got)
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin 应为 *，got %q", got)
	}
	if got := resp.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("当前图片接口未定义显式 Cache-Control，got %q", got)
	}
}

func TestImagesGetRejectsOtherUser(t *testing.T) {
	r := setupImagesHandlerTest(t)
	ownerToken := tokenForTestUser(t, "user-1")
	otherToken := tokenForTestUser(t, "user-2")
	id := uploadTestImage(t, r, ownerToken, "image.png", []byte("png bytes"), "")

	req := httptest.NewRequest(http.MethodGet, "/api/images/"+id+"?token="+otherToken, nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("其他用户读取图片应返回 404，got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestImagesDeleteRemovesImage(t *testing.T) {
	r := setupImagesHandlerTest(t)
	token := tokenForTestUser(t, "user-1")
	id := uploadTestImage(t, r, token, "image.png", []byte("png bytes"), "image/png")

	var row database.Image
	if err := database.DB.Where("id = ?", id).First(&row).Error; err != nil {
		t.Fatalf("上传后应存在图片记录: %v", err)
	}
	filePath, err := filepath.Abs(filepath.Join(config.App.UploadDir, row.FilePath))
	if err != nil {
		t.Fatalf("解析图片路径失败: %v", err)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/images/"+id+"?token="+token, nil)
	deleteResp := httptest.NewRecorder()
	r.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("删除图片应成功，status=%d body=%s", deleteResp.Code, deleteResp.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/images/"+id+"?token="+token, nil)
	getResp := httptest.NewRecorder()
	r.ServeHTTP(getResp, getReq)
	if getResp.Code != http.StatusNotFound {
		t.Fatalf("删除后读取图片应返回 404，got %d", getResp.Code)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("删除后图片文件应被移除，stat err=%v", err)
	}
}
