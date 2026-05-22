package database

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
