package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAdminHandlerTest(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()
	config.App = &config.Config{
		DataDir:             filepath.Join(tmp, "data"),
		UploadDir:           filepath.Join(tmp, "upload"),
		JWTSecret:           "test-secret",
		AdminApikey:         "test-admin",
		InviteInviterReward: 5,
		InviteInviteeReward: 3,
		InviteDefaultQuota:  10,
			InviteEnabled:       true,
	}
	if err := os.MkdirAll(config.App.DataDir, 0755); err != nil {
		t.Fatalf("create temp data dir: %v", err)
	}
	if err := os.MkdirAll(config.App.UploadDir, 0755); err != nil {
		t.Fatalf("create temp upload dir: %v", err)
	}

	// Create a config.json so SetInviteConfig can read-modify-write
	configPath := filepath.Join(tmp, "config.json")
	os.WriteFile(configPath, []byte("{}"), 0644)
	// Override getRootDir to return tmp
	config.SetRootDir(func() string { return tmp })

	db, err := gorm.Open(sqlite.Open(filepath.Join(config.App.DataDir, "test.sqlite")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	database.DB = db
	if err := database.DB.AutoMigrate(&database.User{}, &database.RedemptionCode{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := database.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		config.SetRootDir(func() string {
			dir, _ := os.Getwd()
			return dir
		})
	})

	r := gin.New()
	adminAuth := r.Group("/api/admin", middleware.AdminMiddleware())
	adminAuth.PUT("/users/:id/password", AdminResetPassword)
	adminAuth.GET("/invite-config", AdminGetInviteConfig)
	adminAuth.PUT("/invite-config", AdminUpdateInviteConfig)
	adminAuth.GET("/invites", AdminListInvites)

	return r
}

func adminTokenForTest(t *testing.T) string {
	t.Helper()
	token, err := service.SignToken("admin-1", "admin", config.App.JWTSecret)
	if err != nil {
		t.Fatalf("sign admin token: %v", err)
	}
	return token
}

func createAdminTestUser(t *testing.T, userID string) {
	t.Helper()
	now := time.Now().UnixMilli()
	u := &database.User{
		ID:          userID,
		Label:       "test-user",
		Role:        "user",
		Status:      "active",
		Quota:       50,
		UsedCount:   0,
		CreatedAt:   now,
		LastLoginAt: &now,
	}
	if err := database.DB.Create(u).Error; err != nil {
		t.Fatalf("create test user: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Task 4 Tests: AdminResetPassword
// ---------------------------------------------------------------------------

func TestAdminResetPassword_Success(t *testing.T) {
	r := setupAdminHandlerTest(t)
	createAdminTestUser(t, "user-1")
	token := adminTokenForTest(t)

	body := `{"password":"adminnewpass"}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/user-1/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	// Verify password hash was set in DB
	var dbUser database.User
	if err := database.DB.Where("id = ?", "user-1").First(&dbUser).Error; err != nil {
		t.Fatalf("find user: %v", err)
	}
	if dbUser.PasswordHash == nil {
		t.Fatal("PasswordHash should not be nil after admin reset")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*dbUser.PasswordHash), []byte("adminnewpass")); err != nil {
		t.Error("stored hash should verify against admin-set password")
	}
}

func TestAdminResetPassword_WithoutAdminAuth(t *testing.T) {
	r := setupAdminHandlerTest(t)
	createAdminTestUser(t, "user-1")

	// No auth header
	body := `{"password":"adminnewpass"}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/user-1/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized && resp.Code != http.StatusForbidden {
		t.Fatalf("expected 401/403 without admin auth, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestAdminResetPassword_ShortPassword(t *testing.T) {
	r := setupAdminHandlerTest(t)
	createAdminTestUser(t, "user-1")
	token := adminTokenForTest(t)

	body := `{"password":"short"}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/user-1/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for short password, got %d body=%s", resp.Code, resp.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Task 4 Tests: AdminGetInviteConfig / AdminUpdateInviteConfig
// ---------------------------------------------------------------------------

func TestAdminGetInviteConfig_ReturnsDefaults(t *testing.T) {
	r := setupAdminHandlerTest(t)
	token := adminTokenForTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/invite-config", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var m map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &m)
	if inviter, ok := m["inviterReward"]; !ok {
		t.Error("missing inviterReward")
	} else if v, ok := inviter.(float64); !ok || int(v) != 5 {
		t.Errorf("inviterReward = %v, want 5", inviter)
	}
	if invitee, ok := m["inviteeReward"]; !ok {
		t.Error("missing inviteeReward")
	} else if v, ok := invitee.(float64); !ok || int(v) != 3 {
		t.Errorf("inviteeReward = %v, want 3", invitee)
	}
	if quota, ok := m["defaultQuota"]; !ok {
		t.Error("missing defaultQuota")
	} else if v, ok := quota.(float64); !ok || int(v) != 10 {
		t.Errorf("defaultQuota = %v, want 10", quota)
	}
}

func TestAdminUpdateInviteConfig_Success(t *testing.T) {
	r := setupAdminHandlerTest(t)
	token := adminTokenForTest(t)

	body := `{"inviterReward":20,"inviteeReward":15,"defaultQuota":50}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/invite-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	// Verify config was updated
	if config.App.InviteInviterReward != 20 {
		t.Errorf("InviteInviterReward = %d, want 20", config.App.InviteInviterReward)
	}
	if config.App.InviteInviteeReward != 15 {
		t.Errorf("InviteInviteeReward = %d, want 15", config.App.InviteInviteeReward)
	}
	if config.App.InviteDefaultQuota != 50 {
		t.Errorf("InviteDefaultQuota = %d, want 50", config.App.InviteDefaultQuota)
	}
}

func TestAdminUpdateInviteConfig_NegativeValues(t *testing.T) {
	r := setupAdminHandlerTest(t)
	token := adminTokenForTest(t)

	body := `{"inviterReward":-1,"inviteeReward":5,"defaultQuota":10}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/invite-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative reward, got %d body=%s", resp.Code, resp.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Task 4 Tests: AdminListInvites
// ---------------------------------------------------------------------------

func TestAdminListInvites_ReturnsList(t *testing.T) {
	r := setupAdminHandlerTest(t)
	token := adminTokenForTest(t)

	// Create a user with invite code
	now := time.Now().UnixMilli()
	ic := "MYINVITE"
	un := "inviter"
	ph := "$2a$10$aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	inviter := &database.User{
		ID:              "inviter-1",
		Label:           "inviter-1",
		Username:        &un,
		PasswordHash:    &ph,
		InviteCode:      &ic,
		InviteCodeSetAt: &now,
		Role:            "user",
		Status:          "active",
		Quota:           100,
		UsedCount:       0,
		CreatedAt:       now,
		LastLoginAt:     &now,
	}
	if err := database.DB.Create(inviter).Error; err != nil {
		t.Fatalf("create inviter: %v", err)
	}

	// Create an invited user
	invitedBy := "MYINVITE"
	iU := "invited-user"
	invitee := &database.User{
		ID:              "invitee-1",
		Label:           "invitee-1",
		Username:        &iU,
		PasswordHash:    &ph,
		InvitedBy:       &invitedBy,
		Role:            "user",
		Status:          "active",
		Quota:           13,
		UsedCount:       0,
		CreatedAt:       now,
		LastLoginAt:     &now,
	}
	if err := database.DB.Create(invitee).Error; err != nil {
		t.Fatalf("create invitee: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/invites", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var m map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &m)
	invites, ok := m["invites"].([]interface{})
	if !ok {
		t.Fatalf("invites field missing or not array: %v", m)
	}
	if len(invites) != 1 {
		t.Fatalf("len(invites) = %d, want 1", len(invites))
	}
	row := invites[0].(map[string]interface{})
	if un, ok := row["username"]; !ok || un != "inviter" {
		t.Errorf("username = %v, want 'inviter'", un)
	}
	if code, ok := row["inviteCode"]; !ok || code != "MYINVITE" {
		t.Errorf("inviteCode = %v, want 'MYINVITE'", code)
	}
	if count, ok := row["usageCount"]; !ok {
		t.Error("missing usageCount")
	} else if v, ok := count.(float64); !ok || int(v) != 1 {
		t.Errorf("usageCount = %v, want 1", count)
	}
}
