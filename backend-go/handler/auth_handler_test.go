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

func setupAuthHandlerTest(t *testing.T) *gin.Engine {
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
	}
	if err := os.MkdirAll(config.App.DataDir, 0755); err != nil {
		t.Fatalf("create temp data dir: %v", err)
	}
	if err := os.MkdirAll(config.App.UploadDir, 0755); err != nil {
		t.Fatalf("create temp upload dir: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(filepath.Join(config.App.DataDir, "test.sqlite")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	database.DB = db
	if err := database.DB.AutoMigrate(&database.User{}, &database.RedemptionCode{}, &database.Image{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := database.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	r := gin.New()
	auth := r.Group("/api/auth")
	auth.POST("/login", AuthLogin)
	auth.GET("/me", middleware.AuthMiddleware(), AuthMe)
	auth.POST("/redeem", middleware.AuthMiddleware(), AuthRedeem)
	auth.POST("/login-password", AuthLoginPassword)
	auth.POST("/register", AuthRegister)
	auth.POST("/migrate", middleware.AuthMiddleware(), AuthMigrate)
	auth.POST("/change-password", middleware.AuthMiddleware(), AuthChangePassword)
	auth.PUT("/invite-code", middleware.AuthMiddleware(), AuthSetInviteCode)
	auth.GET("/invite-code", middleware.AuthMiddleware(), AuthGetInviteCode)

	return r
}

func authTokenForUser(t *testing.T, userID string) string {
	t.Helper()
	token, err := service.SignToken(userID, "user", config.App.JWTSecret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func createAuthTestUser(t *testing.T, userID, username, password string) {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashStr := string(hash)
	now := time.Now().UnixMilli()
	label := userID
	if len(label) > 8 {
		label = label[:8]
	}
	u := &database.User{
		ID:           userID,
		Label:        label,
		Username:     &username,
		PasswordHash: &hashStr,
		Role:         "user",
		Status:       "active",
		Quota:        50,
		UsedCount:    0,
		CreatedAt:    now,
		LastLoginAt:  &now,
	}
	if err := database.DB.Create(u).Error; err != nil {
		t.Fatalf("create test user: %v", err)
	}
}

func createAuthLegacyUser(t *testing.T, userID, label string) {
	t.Helper()
	now := time.Now().UnixMilli()
	u := &database.User{
		ID:          userID,
		Label:       label,
		Role:        "user",
		Status:      "active",
		Quota:       50,
		UsedCount:   0,
		CreatedAt:   now,
		LastLoginAt: &now,
	}
	if err := database.DB.Create(u).Error; err != nil {
		t.Fatalf("create legacy user: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Task 3 Tests: AuthLoginPassword
// ---------------------------------------------------------------------------

func TestAuthLoginPassword_Success(t *testing.T) {
	r := setupAuthHandlerTest(t)
	createAuthTestUser(t, "user-1", "testuser", "correctpw")

	body := `{"username":"testuser","password":"correctpw"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(resp.Body.Bytes(), &m); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if _, ok := m["token"]; !ok {
		t.Error("response missing token")
	}
	if _, ok := m["user"]; !ok {
		t.Error("response missing user")
	}
	if _, ok := m["needsMigration"]; !ok {
		t.Error("response missing needsMigration")
	}
}

func TestAuthLoginPassword_InvalidCredentials(t *testing.T) {
	r := setupAuthHandlerTest(t)
	createAuthTestUser(t, "user-1", "testuser", "correctpw")

	body := `{"username":"testuser","password":"wrongpw"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", resp.Code, resp.Body.String())
	}

	var m map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &m)
	if errStr, ok := m["error"].(string); !ok || errStr != "用户名或密码错误" {
		t.Errorf("error = %v, want '用户名或密码错误'", m["error"])
	}
}

func TestAuthLoginPassword_MissingFields(t *testing.T) {
	r := setupAuthHandlerTest(t)

	body := `{"username":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Task 3 Tests: AuthRegister
// ---------------------------------------------------------------------------

func TestAuthRegister_Success(t *testing.T) {
	r := setupAuthHandlerTest(t)

	body := `{"username":"newuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(resp.Body.Bytes(), &m); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if _, ok := m["token"]; !ok {
		t.Error("response missing token")
	}
	if _, ok := m["user"]; !ok {
		t.Error("response missing user")
	}
}

func TestAuthRegister_WeakPassword(t *testing.T) {
	r := setupAuthHandlerTest(t)

	body := `{"username":"newuser","password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestAuthRegister_ShortUsername(t *testing.T) {
	r := setupAuthHandlerTest(t)

	body := `{"username":"ab","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Task 3 Tests: AuthMigrate
// ---------------------------------------------------------------------------

func TestAuthMigrate_Success(t *testing.T) {
	r := setupAuthHandlerTest(t)
	createAuthLegacyUser(t, "user-legacy", "LEGACY")
	token := authTokenForUser(t, "user-legacy")

	body := `{"username":"migrateduser","password":"newpass123","confirmPassword":"newpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/migrate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var m map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &m)
	if userV, ok := m["user"]; !ok || userV == nil {
		t.Error("response missing user")
	}
}

func TestAuthMigrate_Unauthenticated(t *testing.T) {
	r := setupAuthHandlerTest(t)

	body := `{"username":"migrateduser","password":"newpass123","confirmPassword":"newpass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/migrate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestAuthMigrate_PasswordMismatch(t *testing.T) {
	r := setupAuthHandlerTest(t)
	createAuthLegacyUser(t, "user-legacy", "LEGACY")
	token := authTokenForUser(t, "user-legacy")

	body := `{"username":"migrateduser","password":"newpass123","confirmPassword":"different"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/migrate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Task 3 Tests: AuthChangePassword
// ---------------------------------------------------------------------------

func TestAuthChangePassword_Success(t *testing.T) {
	r := setupAuthHandlerTest(t)
	createAuthTestUser(t, "user-1", "testuser", "oldpassword")
	token := authTokenForUser(t, "user-1")

	body := `{"oldPassword":"oldpassword","newPassword":"newpassword","confirmPassword":"newpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestAuthChangePassword_WrongOldPassword(t *testing.T) {
	r := setupAuthHandlerTest(t)
	createAuthTestUser(t, "user-1", "testuser", "correctpw")
	token := authTokenForUser(t, "user-1")

	body := `{"oldPassword":"wrongpw","newPassword":"newpassword","confirmPassword":"newpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Task 3 Tests: AuthSetInviteCode and AuthGetInviteCode
// ---------------------------------------------------------------------------

func TestAuthSetAndGetInviteCode(t *testing.T) {
	r := setupAuthHandlerTest(t)
	createAuthTestUser(t, "user-1", "testuser", "correctpw")
	token := authTokenForUser(t, "user-1")

	// Set invite code
	body := `{"code":"MYINVITE"}`
	req := httptest.NewRequest(http.MethodPut, "/api/auth/invite-code", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("set invite code: expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	// Get invite code
	getReq := httptest.NewRequest(http.MethodGet, "/api/auth/invite-code", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getResp := httptest.NewRecorder()
	r.ServeHTTP(getResp, getReq)

	if getResp.Code != http.StatusOK {
		t.Fatalf("get invite code: expected 200, got %d body=%s", getResp.Code, getResp.Body.String())
	}

	var m map[string]interface{}
	json.Unmarshal(getResp.Body.Bytes(), &m)
	if code, ok := m["code"]; !ok || code != "MYINVITE" {
		t.Errorf("code = %v, want 'MYINVITE'", m["code"])
	}
}

// ---------------------------------------------------------------------------
// Task 3 Tests: AuthLogin (modified) returns needsMigration
// ---------------------------------------------------------------------------

func TestAuthLogin_ReturnsNeedsMigration(t *testing.T) {
	r := setupAuthHandlerTest(t)

	// Create an unused redemption code
	now := time.Now().UnixMilli()
	rc := &database.RedemptionCode{
		ID:        "rc-1",
		Code:      "TEST-CODE",
		Quota:     50,
		CreatedAt: now,
	}
	if err := database.DB.Create(rc).Error; err != nil {
		t.Fatalf("create redemption code: %v", err)
	}

	body := `{"code":"TEST-CODE"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var m map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &m)
	if nm, ok := m["needsMigration"]; !ok {
		t.Error("login response missing needsMigration field")
	} else if nm != true {
		t.Errorf("needsMigration = %v, want true for new code user", nm)
	}
}

// ---------------------------------------------------------------------------
// Task 3 Tests: AuthMe returns username and needsMigration
// ---------------------------------------------------------------------------

func TestAuthMe_ReturnsUsernameAndNeedsMigration(t *testing.T) {
	r := setupAuthHandlerTest(t)
	createAuthTestUser(t, "user-1", "testuser", "correctpw")
	token := authTokenForUser(t, "user-1")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var m map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &m)
	user, ok := m["user"].(map[string]interface{})
	if !ok {
		t.Fatal("response missing user object")
	}
	if un, ok := user["username"]; !ok || un == nil || un == "" {
		t.Error("user should have username field")
	}
	if _, ok := user["needsMigration"]; !ok {
		t.Error("user should have needsMigration field")
	}
}

func TestAuthMe_LegacyUserNeedsMigration(t *testing.T) {
	r := setupAuthHandlerTest(t)
	createAuthLegacyUser(t, "user-legacy", "LEGACY")
	token := authTokenForUser(t, "user-legacy")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var m map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &m)
	user := m["user"].(map[string]interface{})
	if nm, ok := user["needsMigration"]; !ok || nm != true {
		t.Errorf("needsMigration = %v, want true for legacy user", user["needsMigration"])
	}
}
