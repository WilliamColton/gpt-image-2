package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthServiceTest(t *testing.T) string {
	t.Helper()

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
	if err := database.DB.AutoMigrate(&database.User{}, &database.RedemptionCode{}); err != nil {
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

// Helper to create a user with a bcrypt-hashed password directly in the DB.
func testLabel(userID string) string {
	if len(userID) >= 8 {
		return userID[:8]
	}
	return userID
}

func createTestUserWithPassword(t *testing.T, userID, username, password string, quota int, role string) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}
	hashStr := string(hash)
	now := time.Now().UnixMilli()
	u := &database.User{
		ID:           userID,
		Label:        testLabel(userID),
		Username:     &username,
		PasswordHash: &hashStr,
		Role:         role,
		Status:       "active",
		Quota:        quota,
		UsedCount:    0,
		CreatedAt:    now,
		LastLoginAt:  &now,
	}
	if err := database.DB.Create(u).Error; err != nil {
		t.Fatalf("create test user: %v", err)
	}
}

// Helper to create a user with password_hash IS NULL (legacy user).
func createTestUserWithCode(t *testing.T, userID, label string, quota int) {
	t.Helper()
	now := time.Now().UnixMilli()
	u := &database.User{
		ID:          userID,
		Label:       label,
		Role:        "user",
		Status:      "active",
		Quota:       quota,
		UsedCount:   0,
		CreatedAt:   now,
		LastLoginAt: &now,
	}
	if err := database.DB.Create(u).Error; err != nil {
		t.Fatalf("create test user: %v", err)
	}
}

// Helper to create a user with an invite code set.
func createTestUserWithInviteCode(t *testing.T, userID, username, inviteCode string, quota int) {
	t.Helper()
	ph := "$2a$10$aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	now := time.Now().UnixMilli()
	icSetAt := now
	u := &database.User{
		ID:              userID,
		Label:           testLabel(userID),
		Username:        &username,
		PasswordHash:    &ph,
		InviteCode:      &inviteCode,
		InviteCodeSetAt: &icSetAt,
		Role:            "user",
		Status:          "active",
		Quota:           quota,
		UsedCount:       0,
		CreatedAt:       now,
		LastLoginAt:     &now,
	}
	if err := database.DB.Create(u).Error; err != nil {
		t.Fatalf("create test user with invite code: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Task 2 Tests: LoginWithPassword and RegisterUser
// ---------------------------------------------------------------------------

func TestLoginWithPassword_Success(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithPassword(t, "user-1", "testuser", "correctpw", 50, "user")

	token, user, needsMigrate, err := LoginWithPassword("testuser", "correctpw")
	if err != nil {
		t.Fatalf("LoginWithPassword should succeed: %v", err)
	}
	if token == "" {
		t.Error("token should not be empty")
	}
	if user == nil {
		t.Fatal("user should not be nil")
	}
	if user.ID != "user-1" {
		t.Errorf("user.ID = %q, want 'user-1'", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("user.Username = %q, want 'testuser'", user.Username)
	}
	if needsMigrate {
		t.Error("needsMigrate should be false when PasswordHash is set")
	}
	if user.NeedsMigration {
		t.Error("user.NeedsMigration should be false when PasswordHash is set")
	}
}

func TestLoginWithPassword_WrongPassword(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithPassword(t, "user-1", "testuser", "correctpw", 50, "user")

	token, _, _, err := LoginWithPassword("testuser", "wrongpw")
	if err == nil {
		t.Fatal("LoginWithPassword with wrong password should return error")
	}
	if token != "" {
		t.Error("token should be empty on error")
	}
	if err.Error() != "用户名或密码错误" {
		t.Errorf("error = %q, want '用户名或密码错误'", err.Error())
	}
}

func TestLoginWithPassword_UserNotFound(t *testing.T) {
	setupAuthServiceTest(t)

	token, _, _, err := LoginWithPassword("nonexistent", "any")
	if err == nil {
		t.Fatal("LoginWithPassword for nonexistent user should return error")
	}
	if token != "" {
		t.Error("token should be empty on error")
	}
	// Must not leak user existence
	if err.Error() != "用户名或密码错误" {
		t.Errorf("error = %q, want '用户名或密码错误' (must not leak existence)", err.Error())
	}
}

func TestLoginWithPassword_NoPasswordHash(t *testing.T) {
	setupAuthServiceTest(t)
	// Create a user with a username but no password hash (legacy user with username set via migration placeholder)
	now := time.Now().UnixMilli()
	username := "nohashuser"
	u := &database.User{
		ID:          "user-nohash",
		Label:       "user-nohash",
		Username:    &username,
		Role:        "user",
		Status:      "active",
		Quota:       50,
		UsedCount:   0,
		CreatedAt:   now,
		LastLoginAt: &now,
		// PasswordHash is nil intentionally
	}
	if err := database.DB.Create(u).Error; err != nil {
		t.Fatalf("create no-hash user: %v", err)
	}

	token, _, _, err := LoginWithPassword("nohashuser", "any")
	if err == nil {
		t.Fatal("LoginWithPassword for user without password hash should return error")
	}
	if token != "" {
		t.Error("token should be empty on error")
	}
	if err.Error() != "该账号尚未设置密码，请使用兑换码登录后设置密码" {
		t.Errorf("error = %q, want hint to use redeem code", err.Error())
	}
}

func TestLoginWithPassword_DisabledUser(t *testing.T) {
	setupAuthServiceTest(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpw"), bcrypt.DefaultCost)
	hashStr := string(hash)
	now := time.Now().UnixMilli()
	username := "disabled-user"
	u := &database.User{
		ID:           "user-disabled",
		Label:        "disabled",
		Username:     &username,
		PasswordHash: &hashStr,
		Role:         "user",
		Status:       "disabled",
		Quota:        50,
		CreatedAt:    now,
	}
	if err := database.DB.Create(u).Error; err != nil {
		t.Fatalf("create disabled user: %v", err)
	}

	token, _, _, err := LoginWithPassword("disabled-user", "correctpw")
	if err == nil {
		t.Fatal("LoginWithPassword for disabled user should return error")
	}
	if token != "" {
		t.Error("token should be empty on error")
	}
	if err.Error() != "账号已被禁用" {
		t.Errorf("error = %q, want '账号已被禁用'", err.Error())
	}
}

func TestRegisterUser_SuccessWithoutInviteCode(t *testing.T) {
	setupAuthServiceTest(t)

	token, user, err := RegisterUser("newuser", "password12345678", "")
	if err != nil {
		t.Fatalf("RegisterUser should succeed: %v", err)
	}
	if token == "" {
		t.Error("token should not be empty")
	}
	if user == nil {
		t.Fatal("user should not be nil")
	}
	if user.Username != "newuser" {
		t.Errorf("user.Username = %q, want 'newuser'", user.Username)
	}
	if user.Role != "user" {
		t.Errorf("user.Role = %q, want 'user'", user.Role)
	}
	if user.Quota != 10 {
		t.Errorf("user.Quota = %d, want 10 (default quota)", user.Quota)
	}
	if user.NeedsMigration {
		t.Error("newly registered user should not need migration")
	}

	// Verify the user was created in the database with bcrypt hash
	var dbUser database.User
	if err := database.DB.Where("username = ?", "newuser").First(&dbUser).Error; err != nil {
		t.Fatalf("find created user: %v", err)
	}
	if dbUser.PasswordHash == nil {
		t.Error("created user should have a password hash")
	} else {
		if err := bcrypt.CompareHashAndPassword([]byte(*dbUser.PasswordHash), []byte("password12345678")); err != nil {
			t.Error("stored hash should verify against original password")
		}
	}
}

func TestRegisterUser_DuplicateUsername(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithPassword(t, "user-1", "dupuser", "somepass", 50, "user")

	token, _, err := RegisterUser("dupuser", "password12345678", "")
	if err == nil {
		t.Fatal("RegisterUser with duplicate username should return error")
	}
	if token != "" {
		t.Error("token should be empty on error")
	}
	if err.Error() != "用户名已被使用" {
		t.Errorf("error = %q, want '用户名已被使用'", err.Error())
	}
}

func TestRegisterUser_InvalidInviteCode(t *testing.T) {
	setupAuthServiceTest(t)

	token, _, err := RegisterUser("newuser", "password12345678", "NONEXISTENT_INVITE")
	if err == nil {
		t.Fatal("RegisterUser with invalid invite code should return error")
	}
	if token != "" {
		t.Error("token should be empty on error")
	}
	if err.Error() != "邀请码无效" {
		t.Errorf("error = %q, want '邀请码无效'", err.Error())
	}
}

func TestRegisterUser_SuccessWithValidInviteCode(t *testing.T) {
	setupAuthServiceTest(t)

	// Pre-create an inviter user who has set their invite code
	createTestUserWithInviteCode(t, "inviter-1", "inviteruser", "MYCODE", 100)

	token, user, err := RegisterUser("inviteduser", "password12345678", "  MYCODE  ") // trimmed
	if err != nil {
		t.Fatalf("RegisterUser with valid invite code should succeed: %v", err)
	}
	if token == "" {
		t.Error("token should not be empty")
	}
	if user == nil {
		t.Fatal("user should not be nil")
	}
	// Invitee reward: default 10 + invitee reward 3 = 13
	if user.Quota != 13 {
		t.Errorf("user.Quota = %d, want 13 (default 10 + inviteeReward 3)", user.Quota)
	}

	// Check InvitedBy was set
	var dbUser database.User
	if err := database.DB.Where("username = ?", "inviteduser").First(&dbUser).Error; err != nil {
		t.Fatalf("find invited user: %v", err)
	}
	if dbUser.InvitedBy == nil || *dbUser.InvitedBy != "MYCODE" {
		t.Errorf("InvitedBy = %v, want 'MYCODE'", dbUser.InvitedBy)
	}

	// Check inviter got their reward
	var inviter database.User
	if err := database.DB.Where("id = ?", "inviter-1").First(&inviter).Error; err != nil {
		t.Fatalf("find inviter: %v", err)
	}
	if inviter.Quota != 105 {
		t.Errorf("inviter.Quota = %d, want 105 (original 100 + inviterReward 5)", inviter.Quota)
	}
}

func TestRegisterUser_UsernameTooShort(t *testing.T) {
	setupAuthServiceTest(t)

	token, _, err := RegisterUser("ab", "password12345678", "")
	if err == nil {
		t.Fatal("RegisterUser with short username should return error")
	}
	if token != "" {
		t.Error("token should be empty on error")
	}
	if err.Error() != "用户名须为 3-20 个字符" {
		t.Errorf("error = %q, want '用户名须为 3-20 个字符'", err.Error())
	}
}

func TestRegisterUser_UsernameTooLong(t *testing.T) {
	setupAuthServiceTest(t)

	longName := strings.Repeat("x", 21)
	token, _, err := RegisterUser(longName, "password12345678", "")
	if err == nil {
		t.Fatal("RegisterUser with long username should return error")
	}
	if token != "" {
		t.Error("token should be empty on error")
	}
	if err.Error() != "用户名须为 3-20 个字符" {
		t.Errorf("error = %q, want '用户名须为 3-20 个字符'", err.Error())
	}
}

func TestRegisterUser_PasswordTooShort(t *testing.T) {
	setupAuthServiceTest(t)

	token, _, err := RegisterUser("validuser", "short", "")
	if err == nil {
		t.Fatal("RegisterUser with short password should return error")
	}
	if token != "" {
		t.Error("token should be empty on error")
	}
	if err.Error() != "密码至少需要 8 个字符" {
		t.Errorf("error = %q, want '密码至少需要 8 个字符'", err.Error())
	}
}

func TestRegisterUser_ChineseUsername(t *testing.T) {
	setupAuthServiceTest(t)

	// 3 Chinese characters = 3 runes, meets 3-char minimum
	token, user, err := RegisterUser("用户名", "password12345678", "")
	if err != nil {
		t.Fatalf("RegisterUser with Chinese username should succeed: %v", err)
	}
	if token == "" {
		t.Error("token should not be empty")
	}
	if user.Username != "用户名" {
		t.Errorf("user.Username = %q, want '用户名'", user.Username)
	}
}

// ---------------------------------------------------------------------------
// Task 2 Test: LoginWithCode now returns needsMigration
// ---------------------------------------------------------------------------

func TestLoginWithCode_ReturnsNeedsMigrationForOldUser(t *testing.T) {
	setupAuthServiceTest(t)

	// Create a redemption code and a user who used it (legacy, no password)
	now := time.Now().UnixMilli()
	rc := &database.RedemptionCode{
		ID:        "rc-1",
		Code:      "LEGACY-CODE",
		Quota:     50,
		CreatedAt: now,
		UsedBy:    nil,
	}
	if err := database.DB.Create(rc).Error; err != nil {
		t.Fatalf("create redemption code: %v", err)
	}

	// LoginWithCode creates a new user for unused code
	// First call: new user, password hash will be nil
	token, user, err := LoginWithCode("LEGACY-CODE")
	if err != nil {
		t.Fatalf("LoginWithCode (first, creates user): %v", err)
	}
	if token == "" || user == nil {
		t.Fatal("new user created via code login should return token + user")
	}
	if !user.NeedsMigration {
		t.Error("new user created via code login should have NeedsMigration=true (no password)")
	}
}

func TestLoginWithCode_ExistingUserNeedsMigration(t *testing.T) {
	setupAuthServiceTest(t)

	// Create a redemption code
	now := time.Now().UnixMilli()
	rc := &database.RedemptionCode{
		ID:        "rc-2",
		Code:      "ANOTHER-CODE",
		Quota:     50,
		CreatedAt: now,
		UsedBy:    nil,
	}
	if err := database.DB.Create(rc).Error; err != nil {
		t.Fatalf("create redemption code: %v", err)
	}

	// First login creates the user
	LoginWithCode("ANOTHER-CODE")

	// Second login (existing user)
	token, user, err := LoginWithCode("ANOTHER-CODE")
	if err != nil {
		t.Fatalf("LoginWithCode (second, existing user): %v", err)
	}
	if token == "" || user == nil {
		t.Fatal("existing user login should return token + user")
	}
	if !user.NeedsMigration {
		t.Error("legacy user (no password) login via code should have NeedsMigration=true")
	}
}

// ---------------------------------------------------------------------------
// Task 3 Tests: MigrateUser, ChangePassword, SetInviteCode, GetInviteCode,
//              ListInvites, AdminResetPassword
// ---------------------------------------------------------------------------

func TestMigrateUser_Success(t *testing.T) {
	setupAuthServiceTest(t)
	// Create a legacy user with no username or password
	createTestUserWithCode(t, "user-legacy", "OLDLABEL", 50)

	updated, err := MigrateUser("user-legacy", "newusername", "newpass123")
	if err != nil {
		t.Fatalf("MigrateUser should succeed: %v", err)
	}
	if updated == nil {
		t.Fatal("updated user should not be nil")
	}
	if updated.Username != "newusername" {
		t.Errorf("Username = %q, want 'newusername'", updated.Username)
	}
	if updated.NeedsMigration {
		t.Error("after migration, NeedsMigration should be false")
	}

	// Verify in database
	var dbUser database.User
	if err := database.DB.Where("id = ?", "user-legacy").First(&dbUser).Error; err != nil {
		t.Fatalf("find migrated user: %v", err)
	}
	if dbUser.Username == nil || *dbUser.Username != "newusername" {
		t.Errorf("DB Username = %v, want 'newusername'", dbUser.Username)
	}
	if dbUser.PasswordHash == nil {
		t.Error("DB PasswordHash should not be nil after migration")
	} else {
		if err := bcrypt.CompareHashAndPassword([]byte(*dbUser.PasswordHash), []byte("newpass123")); err != nil {
			t.Error("stored hash should verify against the new password")
		}
	}
}

func TestMigrateUser_DuplicateUsername(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithPassword(t, "user-1", "takenname", "somepass", 50, "user")
	createTestUserWithCode(t, "user-legacy", "OLDLABEL", 50)

	_, err := MigrateUser("user-legacy", "takenname", "newpass123")
	if err == nil {
		t.Fatal("MigrateUser with duplicate username should return error")
	}
	if err.Error() != "用户名已被使用" {
		t.Errorf("error = %q, want '用户名已被使用'", err.Error())
	}
}

func TestMigrateUser_UsernameTooShort(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithCode(t, "user-legacy", "OLDLABEL", 50)

	_, err := MigrateUser("user-legacy", "ab", "newpass123")
	if err == nil {
		t.Fatal("MigrateUser with short username should return error")
	}
	if err.Error() != "用户名须为 3-20 个字符" {
		t.Errorf("error = %q, want '用户名须为 3-20 个字符'", err.Error())
	}
}

func TestMigrateUser_PasswordTooShort(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithCode(t, "user-legacy", "OLDLABEL", 50)

	_, err := MigrateUser("user-legacy", "validuser", "short")
	if err == nil {
		t.Fatal("MigrateUser with short password should return error")
	}
	if err.Error() != "密码至少需要 8 个字符" {
		t.Errorf("error = %q, want '密码至少需要 8 个字符'", err.Error())
	}
}

func TestMigrateUser_CanUseSameUsername(t *testing.T) {
	setupAuthServiceTest(t)
	// User tries to set their own username to the same value (should succeed)
	// Actually the plan says user has NO username yet (legacy). Let's test
	// that they can pick any available username.
	createTestUserWithCode(t, "user-legacy", "OLDLABEL", 50)

	updated, err := MigrateUser("user-legacy", "myname", "password12345678")
	if err != nil {
		t.Fatalf("MigrateUser should succeed: %v", err)
	}
	if updated.Username != "myname" {
		t.Errorf("Username = %q, want 'myname'", updated.Username)
	}
}

func TestChangePassword_Success(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithPassword(t, "user-1", "testuser", "oldpassword", 50, "user")

	err := ChangePassword("user-1", "oldpassword", "newpassword")
	if err != nil {
		t.Fatalf("ChangePassword should succeed: %v", err)
	}

	// Verify the new password is stored
	var dbUser database.User
	if err := database.DB.Where("id = ?", "user-1").First(&dbUser).Error; err != nil {
		t.Fatalf("find user: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*dbUser.PasswordHash), []byte("newpassword")); err != nil {
		t.Error("stored hash should verify against new password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*dbUser.PasswordHash), []byte("oldpassword")); err == nil {
		t.Error("stored hash should NOT verify against old password")
	}
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithPassword(t, "user-1", "testuser", "correctpw", 50, "user")

	err := ChangePassword("user-1", "wrongpw", "newpassword")
	if err == nil {
		t.Fatal("ChangePassword with wrong old password should return error")
	}
	if err.Error() != "旧密码不正确" {
		t.Errorf("error = %q, want '旧密码不正确'", err.Error())
	}
}

func TestChangePassword_NoPasswordHash(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithCode(t, "user-nohash", "LEGACY", 50)

	err := ChangePassword("user-nohash", "any", "newpassword")
	if err == nil {
		t.Fatal("ChangePassword for user without password should return error")
	}
	if err.Error() != "该账号尚未设置密码" {
		t.Errorf("error = %q, want '该账号尚未设置密码'", err.Error())
	}
}

func TestChangePassword_NewPasswordTooShort(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithPassword(t, "user-1", "testuser", "correctpw", 50, "user")

	err := ChangePassword("user-1", "correctpw", "short")
	if err == nil {
		t.Fatal("ChangePassword with short new password should return error")
	}
	if err.Error() != "密码至少需要 8 个字符" {
		t.Errorf("error = %q, want '密码至少需要 8 个字符'", err.Error())
	}
}

func TestSetInviteCode_Success(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithPassword(t, "user-1", "testuser", "correctpw", 50, "user")

	err := SetInviteCode("user-1", "MYCODE")
	if err != nil {
		t.Fatalf("SetInviteCode should succeed: %v", err)
	}

	// Verify in database
	var dbUser database.User
	if err := database.DB.Where("id = ?", "user-1").First(&dbUser).Error; err != nil {
		t.Fatalf("find user: %v", err)
	}
	if dbUser.InviteCode == nil || *dbUser.InviteCode != "MYCODE" {
		t.Errorf("InviteCode = %v, want 'MYCODE'", dbUser.InviteCode)
	}
	if dbUser.InviteCodeSetAt == nil {
		t.Error("InviteCodeSetAt should not be nil")
	}
}

func TestSetInviteCode_DuplicateFails(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithInviteCode(t, "user-1", "user1", "TAKEN", 50)
	createTestUserWithPassword(t, "user-2", "user2", "correctpw", 50, "user")

	err := SetInviteCode("user-2", "TAKEN")
	if err == nil {
		t.Fatal("SetInviteCode with duplicate code should return error")
	}
	if err.Error() != "该邀请码已被使用" {
		t.Errorf("error = %q, want '该邀请码已被使用'", err.Error())
	}
}

func TestSetInviteCode_ReplaceOwnCode(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithInviteCode(t, "user-1", "user1", "OLDCODE", 50)

	// Replace own invite code
	err := SetInviteCode("user-1", "NEWCODE")
	if err != nil {
		t.Fatalf("SetInviteCode replacing own code should succeed: %v", err)
	}

	var dbUser database.User
	if err := database.DB.Where("id = ?", "user-1").First(&dbUser).Error; err != nil {
		t.Fatalf("find user: %v", err)
	}
	if dbUser.InviteCode == nil || *dbUser.InviteCode != "NEWCODE" {
		t.Errorf("InviteCode = %v, want 'NEWCODE'", dbUser.InviteCode)
	}

	// Verify that another user can now take OLDCODE
	createTestUserWithPassword(t, "user-2", "user2", "pass12345", 50, "user")
	err2 := SetInviteCode("user-2", "OLDCODE")
	if err2 != nil {
		t.Fatalf("SetInviteCode with freed code should succeed: %v", err2)
	}
}

func TestSetInviteCode_EmptyCode(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithPassword(t, "user-1", "testuser", "correctpw", 50, "user")

	err := SetInviteCode("user-1", "   ")
	if err == nil {
		t.Fatal("SetInviteCode with empty code should return error")
	}
	if err.Error() != "邀请码不能为空" {
		t.Errorf("error = %q, want '邀请码不能为空'", err.Error())
	}
}

func TestGetInviteCode_Success(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithInviteCode(t, "user-1", "user1", "MYCODE", 50)

	code, setAt, err := GetInviteCode("user-1")
	if err != nil {
		t.Fatalf("GetInviteCode should succeed: %v", err)
	}
	if code == nil || *code != "MYCODE" {
		t.Errorf("code = %v, want 'MYCODE'", code)
	}
	if setAt == nil {
		t.Error("setAt should not be nil")
	}
	fmt.Println(setAt) // use the var
}

func TestGetInviteCode_NotSet(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithPassword(t, "user-1", "testuser", "correctpw", 50, "user")

	code, _, err := GetInviteCode("user-1")
	if err != nil {
		t.Fatalf("GetInviteCode should succeed even if not set: %v", err)
	}
	if code != nil {
		t.Errorf("code = %v, want nil (not set)", code)
	}
}

func TestListInvites_Success(t *testing.T) {
	setupAuthServiceTest(t)

	// Create an inviter
	createTestUserWithInviteCode(t, "inviter-1", "inviter1", "COOLCODE", 100)

	// Create some invited users (setting InvitedBy directly in DB)
	now := time.Now().UnixMilli()
	invitedBy := "COOLCODE"
	// Directly create users who used the invite
	invitedUsers := []database.User{
		{ID: "iu-1", Label: "iu1", Username: stringPtr("invited1"), PasswordHash: stringPtr("$2a$10$hash"), InvitedBy: &invitedBy, Role: "user", Status: "active", Quota: 10, CreatedAt: now},
		{ID: "iu-2", Label: "iu2", Username: stringPtr("invited2"), PasswordHash: stringPtr("$2a$10$hash"), InvitedBy: &invitedBy, Role: "user", Status: "active", Quota: 10, CreatedAt: now},
	}
	if err := database.DB.Create(&invitedUsers).Error; err != nil {
		t.Fatalf("create invited users: %v", err)
	}

	rows, err := ListInvites()
	if err != nil {
		t.Fatalf("ListInvites should succeed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("ListInvites count = %d, want 1", len(rows))
	}
	if rows[0].Username != "inviter1" {
		t.Errorf("Username = %q, want 'inviter1'", rows[0].Username)
	}
	if rows[0].InviteCode != "COOLCODE" {
		t.Errorf("InviteCode = %q, want 'COOLCODE'", rows[0].InviteCode)
	}
	if rows[0].UsageCount != 2 {
		t.Errorf("UsageCount = %d, want 2", rows[0].UsageCount)
	}
}

func TestListInvites_Empty(t *testing.T) {
	setupAuthServiceTest(t)
	rows, err := ListInvites()
	if err != nil {
		t.Fatalf("ListInvites should succeed: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("ListInvites count = %d, want 0", len(rows))
	}
}

func TestAdminResetPassword_Success(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithCode(t, "user-legacy", "LEGACY", 50)

	err := AdminResetPassword("user-legacy", "adminpass")
	if err != nil {
		t.Fatalf("AdminResetPassword should succeed: %v", err)
	}

	// Verify password hash was set
	var dbUser database.User
	if err := database.DB.Where("id = ?", "user-legacy").First(&dbUser).Error; err != nil {
		t.Fatalf("find user: %v", err)
	}
	if dbUser.PasswordHash == nil {
		t.Error("PasswordHash should not be nil after admin reset")
	} else {
		if err := bcrypt.CompareHashAndPassword([]byte(*dbUser.PasswordHash), []byte("adminpass")); err != nil {
			t.Error("stored hash should verify against admin-set password")
		}
	}
}

func TestAdminResetPassword_Overwrites(t *testing.T) {
	setupAuthServiceTest(t)
	createTestUserWithPassword(t, "user-1", "testuser", "oldpass", 50, "user")

	err := AdminResetPassword("user-1", "newadminpass")
	if err != nil {
		t.Fatalf("AdminResetPassword should succeed: %v", err)
	}

	var dbUser database.User
	if err := database.DB.Where("id = ?", "user-1").First(&dbUser).Error; err != nil {
		t.Fatalf("find user: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*dbUser.PasswordHash), []byte("newadminpass")); err != nil {
		t.Error("stored hash should verify against new password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*dbUser.PasswordHash), []byte("oldpass")); err == nil {
		t.Error("stored hash should NOT verify against old password")
	}
}

func TestAdminResetPassword_UserNotFound(t *testing.T) {
	setupAuthServiceTest(t)

	err := AdminResetPassword("nonexistent", "adminpass")
	if err == nil {
		t.Fatal("AdminResetPassword for nonexistent user should return error")
	}
	if err.Error() != "用户不存在" {
		t.Errorf("error = %q, want '用户不存在'", err.Error())
	}
}

func TestAdminResetPassword_PasswordTooShort(t *testing.T) {
	setupAuthServiceTest(t)

	err := AdminResetPassword("user-1", "short")
	if err == nil {
		t.Fatal("AdminResetPassword with short password should return error")
	}
	if err.Error() != "密码至少需要 8 个字符" {
		t.Errorf("error = %q, want '密码至少需要 8 个字符'", err.Error())
	}
}

// ---------------------------------------------------------------------------
// Task 10 Tests: PasswordHash, UsernameValidation, PasswordValidation
// ---------------------------------------------------------------------------

func TestPasswordHash(t *testing.T) {
	setupAuthServiceTest(t)

	hash, err := hashPassword("testpassword123")
	if err != nil {
		t.Fatalf("hashPassword failed: %v", err)
	}
	if hash == "" {
		t.Error("hash should not be empty")
	}

	// Correct password should verify
	if !checkPassword(hash, "testpassword123") {
		t.Error("checkPassword should return true for correct password")
	}

	// Wrong password should not verify
	if checkPassword(hash, "wrongpassword") {
		t.Error("checkPassword should return false for wrong password")
	}
}

func TestPasswordHashNil(t *testing.T) {
	setupAuthServiceTest(t)

	hash1, err := hashPassword("samepassword")
	if err != nil {
		t.Fatalf("hashPassword failed: %v", err)
	}
	if hash1 == "" {
		t.Error("hash should not be empty (nil)")
	}

	hash2, err := hashPassword("samepassword")
	if err != nil {
		t.Fatalf("hashPassword failed: %v", err)
	}

	// Two hashes of the same password should differ due to salt
	if hash1 == hash2 {
		t.Error("two hashes of same password should differ (bcrypt salt)")
	}

	// Hash should not contain plaintext password
	if strings.Contains(hash1, "samepassword") {
		t.Error("hash should not contain plaintext password")
	}

	// Hash should start with bcrypt marker $2a$
	if !strings.HasPrefix(hash1, "$2a$") {
		t.Errorf("hash should start with $2a$ (bcrypt marker), got prefix: %s", hash1[:10])
	}
}

func TestUsernameValidation_ChineseName(t *testing.T) {
	// Valid: 3-char Chinese username
	setupAuthServiceTest(t)
	_, _, err := RegisterUser("測試員", "password12345678", "")
	if err != nil {
		t.Errorf("Chinese username '測試員' (3 runes) should be valid, got error: %v", err)
	}
}

func TestUsernameValidation_TooShort(t *testing.T) {
	// Too short: 2 characters
	setupAuthServiceTest(t)
	_, _, err := RegisterUser("ab", "password12345678", "")
	if err == nil || err.Error() != "用户名须为 3-20 个字符" {
		t.Errorf("expected '用户名须为 3-20 个字符' for 2-char username, got: %v", err)
	}
}

func TestUsernameValidation_TooLong(t *testing.T) {
	// Too long: 21 characters
	setupAuthServiceTest(t)
	_, _, err := RegisterUser(strings.Repeat("x", 21), "password12345678", "")
	if err == nil || err.Error() != "用户名须为 3-20 个字符" {
		t.Errorf("expected '用户名须为 3-20 个字符' for 21-char username, got: %v", err)
	}
}

func TestPasswordValidation_TooShort(t *testing.T) {
	// Too short: 7 characters
	setupAuthServiceTest(t)
	_, _, err := RegisterUser("validuser", "1234567", "")
	if err == nil || err.Error() != "密码至少需要 8 个字符" {
		t.Errorf("expected '密码至少需要 8 个字符' for 7-char password, got: %v", err)
	}
}

func TestPasswordValidation_ExactEight(t *testing.T) {
	// Exact 8 characters — should pass
	setupAuthServiceTest(t)
	_, _, err := RegisterUser("validuser2", "12345678", "")
	if err != nil {
		t.Errorf("8-char password should be valid, got error: %v", err)
	}
}

// helper to get string pointer
func stringPtr(s string) *string {
	return &s
}
