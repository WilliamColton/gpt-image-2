package database

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestUserInviteColumnsAutoMigrate verifies AutoMigrate adds the new
// invite-code columns (password_hash, username, invite_code,
// invite_code_set_at, invited_by) without data loss for existing rows.
func TestUserInviteColumnsAutoMigrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	// Migrate with OLD User model first (simulate pre-migration schema).
	// We can't use the real User because it will have the new fields,
	// so we create a minimal old-style struct only for the initial migration.
	type oldUser struct {
		ID          string `gorm:"primaryKey;type:text"`
		Label       string `gorm:"type:text;not null"`
		Role        string `gorm:"type:text;not null"`
		Status      string `gorm:"type:text;not null;default:active"`
		CreatedAt   int64  `gorm:"not null"`
		LastLoginAt *int64
		Quota       int `gorm:"not null;default:0"`
		UsedCount   int `gorm:"not null;default:0"`
	}
	oldUserTableName := func(oldUser) string { return "users" }
	if err := db.Table("users").AutoMigrate(&oldUser{}); err != nil {
		t.Fatalf("AutoMigrate old schema: %v", err)
	}
	_ = oldUserTableName

	// Insert an existing user under the old schema (no new columns).
	existingUserID := "old-user-001"
	now := int64(1716451200000)
	db.Exec(`INSERT INTO users (id, label, role, status, created_at, quota, used_count) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		existingUserID, "legacy-user", "user", "active", now, 5, 2)

	// Now run AutoMigrate with the real (new) User struct.
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate new User schema: %v", err)
	}

	// Verify the existing row survived and has NULL for new columns.
	var u User
	if err := db.Where("id = ?", existingUserID).First(&u).Error; err != nil {
		t.Fatalf("existing user row disappeared after AutoMigrate: %v", err)
	}
	if u.ID != existingUserID {
		t.Errorf("expected id %q, got %q", existingUserID, u.ID)
	}
	if u.Quota != 5 || u.UsedCount != 2 {
		t.Errorf("existing data corrupted: quota=%d usedCount=%d", u.Quota, u.UsedCount)
	}
	if u.PasswordHash != nil {
		t.Error("PasswordHash should be NULL for legacy user")
	}
	if u.Username != nil {
		t.Error("Username should be NULL for legacy user")
	}
	if u.InviteCode != nil {
		t.Error("InviteCode should be NULL for legacy user")
	}
	if u.InviteCodeSetAt != nil {
		t.Error("InviteCodeSetAt should be NULL for legacy user")
	}
	if u.InvitedBy != nil {
		t.Error("InvitedBy should be NULL for legacy user")
	}

	// Insert a new user with the new columns filled.
	username := "testuser"
	passwordHash := "$2a$10$dummyhash"
	inviteCode := "COOLCODE"
	invitedBy := "OTHERCODE"
	inviteSetAt := int64(1716451200000)
	newUser := &User{
		ID:             "new-user-002",
		Label:          "newuser",
		Role:           "user",
		Status:         "active",
		CreatedAt:      now,
		Quota:          10,
		UsedCount:      0,
		PasswordHash:   &passwordHash,
		Username:       &username,
		InviteCode:     &inviteCode,
		InviteCodeSetAt: &inviteSetAt,
		InvitedBy:      &invitedBy,
	}
	if err := db.Create(newUser).Error; err != nil {
		t.Fatalf("create new user with invite columns: %v", err)
	}

	// Read it back.
	var got User
	if err := db.Where("id = ?", "new-user-002").First(&got).Error; err != nil {
		t.Fatalf("read back new user: %v", err)
	}
	if got.PasswordHash == nil || *got.PasswordHash != passwordHash {
		t.Errorf("PasswordHash mismatch: got %v, want %s", got.PasswordHash, passwordHash)
	}
	if got.Username == nil || *got.Username != username {
		t.Errorf("Username mismatch: got %v, want %s", got.Username, username)
	}
	if got.InviteCode == nil || *got.InviteCode != inviteCode {
		t.Errorf("InviteCode mismatch: got %v, want %s", got.InviteCode, inviteCode)
	}
	if got.InviteCodeSetAt == nil || *got.InviteCodeSetAt != inviteSetAt {
		t.Errorf("InviteCodeSetAt mismatch: got %v, want %d", got.InviteCodeSetAt, inviteSetAt)
	}
	if got.InvitedBy == nil || *got.InvitedBy != invitedBy {
		t.Errorf("InvitedBy mismatch: got %v, want %s", got.InvitedBy, invitedBy)
	}
}

// TestMultipleUsersNullUsername verifies that SQLite allows multiple rows
// with NULL username even when a uniqueIndex exists on the column.
func TestMultipleUsersNullUsername(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	now := int64(1716451200000)
	u1 := &User{ID: "u1", Label: "u1", Role: "user", Status: "active", CreatedAt: now, Quota: 5}
	u2 := &User{ID: "u2", Label: "u2", Role: "user", Status: "active", CreatedAt: now, Quota: 5}
	if err := db.Create(u1).Error; err != nil {
		t.Fatalf("create u1: %v", err)
	}
	if err := db.Create(u2).Error; err != nil {
		t.Fatalf("create u2 (second NULL username — must not trigger UNIQUE): %v", err)
	}

	// Both should exist with NULL username.
	var count int64
	db.Model(&User{}).Where("username IS NULL").Count(&count)
	if count != 2 {
		t.Errorf("expected 2 users with NULL username, got %d", count)
	}
}

func TestBillingRecordAutoMigrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := db.AutoMigrate(&BillingRecord{}); err != nil {
		t.Fatalf("AutoMigrate BillingRecord: %v", err)
	}

	// Verify the table exists by checking if we can query it
	var count int64
	if err := db.Model(&BillingRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("query billing_records: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows, got %d", count)
	}

	// Verify we can create a billing record
	rec := &BillingRecord{
		ID:                      "test-id-1",
		TaskID:                  "task-1",
		UserID:                  "user-1",
		UserLabelSnapshot:       "test-user",
		EndpointBaseURLSnapshot: "https://api.openai.com",
		OutputImageID:           "img-1",
		SuccessImageCount:       1,
		UnitCostX10000:          123400,
		UnitSaleX10000:          200000,
		CostX10000:              123400,
		RevenueX10000:           200000,
		ProfitX10000:            76600,
		CreatedAt:               1716451200000,
	}
	if err := db.Create(rec).Error; err != nil {
		t.Fatalf("create BillingRecord: %v", err)
	}

	// Read back and verify
	var got BillingRecord
	if err := db.First(&got, "id = ?", "test-id-1").Error; err != nil {
		t.Fatalf("read BillingRecord: %v", err)
	}
	if got.TaskID != "task-1" || got.UserID != "user-1" || got.UnitCostX10000 != 123400 {
		t.Fatalf("unexpected BillingRecord: %+v", got)
	}
}

func TestBillingRecordNoCascadeOnTaskDelete(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := db.AutoMigrate(&User{}, &Task{}, &BillingRecord{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	// Create user, task, and billing record
	db.Create(&User{ID: "u1", Label: "u", Role: "user", Status: "active", CreatedAt: 1000})
	db.Create(&Task{ID: "t1", UserID: "u1", Prompt: "p", ParamsJSON: "{}", InputImageIDsJSON: "[]", OutputImageIDsJSON: "[]", Status: "completed", CreatedAt: 1000})
	db.Create(&BillingRecord{
		ID: "b1", TaskID: "t1", UserID: "u1",
		UserLabelSnapshot: "u", EndpointBaseURLSnapshot: "ep", OutputImageID: "img",
		SuccessImageCount: 1,
		UnitCostX10000: 1000, UnitSaleX10000: 2000, CostX10000: 1000, RevenueX10000: 2000, ProfitX10000: 1000,
		CreatedAt: 1000,
	})

	// Delete the task — billing record must survive (no cascade)
	db.Delete(&Task{}, "id = ?", "t1")
	var bCount int64
	db.Model(&BillingRecord{}).Count(&bCount)
	if bCount != 1 {
		t.Fatalf("billing record was cascade-deleted with task: got %d records, want 1", bCount)
	}

	// Delete the user — billing record must still survive
	db.Delete(&User{}, "id = ?", "u1")
	bCount = 0
	db.Model(&BillingRecord{}).Count(&bCount)
	if bCount != 1 {
		t.Fatalf("billing record was cascade-deleted with user: got %d records, want 1", bCount)
	}
}
