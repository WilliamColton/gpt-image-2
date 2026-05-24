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

func setupBillingTest(t *testing.T) {
	t.Helper()

	tmp := t.TempDir()
	config.App = &config.Config{
		DataDir:     filepath.Join(tmp, "data"),
		UploadDir:   filepath.Join(tmp, "upload"),
		JWTSecret:   "test-secret",
		AdminApikey: "test-admin",
	}
	if err := os.MkdirAll(config.App.DataDir, 0755); err != nil {
		t.Fatalf("create data dir: %v", err)
	}
	if err := os.MkdirAll(config.App.UploadDir, 0755); err != nil {
		t.Fatalf("create upload dir: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(filepath.Join(config.App.DataDir, "test.sqlite")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	database.DB = db
	if err := database.DB.AutoMigrate(&database.BillingRecord{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := database.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
}

func TestRecordBillingForSuccessfulImages_ThreeImages(t *testing.T) {
	setupBillingTest(t)

	input := BillingBatchInput{
		TaskID:            "task-abc",
		UserID:            "user-xyz",
		UserLabelSnapshot: "test-user-label",
		UnitSaleX10000:    200000, // 20.00 yuan
		Images: []BillingImageInput{
			{OutputImageID: "img-1", EndpointBaseURLSnapshot: "https://api1.example.com", UnitCostX10000: 123400},
			{OutputImageID: "img-2", EndpointBaseURLSnapshot: "https://api1.example.com", UnitCostX10000: 123400},
			{OutputImageID: "img-3", EndpointBaseURLSnapshot: "https://api2.example.com", UnitCostX10000: 50000},
		},
		CreatedAt: 1716451200000,
	}

	err := RecordBillingForSuccessfulImages(input)
	if err != nil {
		t.Fatalf("RecordBillingForSuccessfulImages: %v", err)
	}

	// Verify exactly 3 rows created
	var rows []database.BillingRecord
	if err := database.DB.Order("output_image_id asc").Find(&rows).Error; err != nil {
		t.Fatalf("query billing_records: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 billing rows, got %d", len(rows))
	}

	// Verify IDs are non-empty and unique
	ids := make(map[string]bool)
	for _, r := range rows {
		if r.ID == "" {
			t.Fatalf("billing row has empty ID")
		}
		if ids[r.ID] {
			t.Fatalf("duplicate billing row ID: %s", r.ID)
		}
		ids[r.ID] = true
	}

	// Verify each row has correct snapshots and computations
	for i, r := range rows {
		if r.TaskID != "task-abc" {
			t.Errorf("row %d: TaskID = %q, want task-abc", i, r.TaskID)
		}
		if r.UserID != "user-xyz" {
			t.Errorf("row %d: UserID = %q, want user-xyz", i, r.UserID)
		}
		if r.UserLabelSnapshot != "test-user-label" {
			t.Errorf("row %d: UserLabelSnapshot = %q, want test-user-label", i, r.UserLabelSnapshot)
		}
		if r.SuccessImageCount != 1 {
			t.Errorf("row %d: SuccessImageCount = %d, want 1", i, r.SuccessImageCount)
		}
		if r.UnitSaleX10000 != 200000 {
			t.Errorf("row %d: UnitSaleX10000 = %d, want 200000", i, r.UnitSaleX10000)
		}
		if r.CreatedAt != 1716451200000 {
			t.Errorf("row %d: CreatedAt = %d, want 1716451200000", i, r.CreatedAt)
		}
		// Cost should match the per-image unit cost
		if r.CostX10000 != r.UnitCostX10000 {
			t.Errorf("row %d: CostX10000 = %d, want %d", i, r.CostX10000, r.UnitCostX10000)
		}
		// Revenue should match the unit sale price (one image each)
		if r.RevenueX10000 != r.UnitSaleX10000 {
			t.Errorf("row %d: RevenueX10000 = %d, want %d", i, r.RevenueX10000, r.UnitSaleX10000)
		}
		// Profit = Revenue - Cost
		expectedProfit := r.UnitSaleX10000 - r.UnitCostX10000
		if r.ProfitX10000 != expectedProfit {
			t.Errorf("row %d: ProfitX10000 = %d, want %d", i, r.ProfitX10000, expectedProfit)
		}
	}
}

func TestRecordBillingForSuccessfulImages_EmptySlice(t *testing.T) {
	setupBillingTest(t)

	input := BillingBatchInput{
		TaskID:            "task-abc",
		UserID:            "user-xyz",
		UserLabelSnapshot: "test-user-label",
		UnitSaleX10000:    200000,
		Images:            []BillingImageInput{},
	}

	err := RecordBillingForSuccessfulImages(input)
	if err != nil {
		t.Fatalf("empty images should not error: %v", err)
	}

	var count int64
	database.DB.Model(&database.BillingRecord{}).Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 rows for empty images, got %d", count)
	}
}

func TestRecordBillingForSuccessfulImages_PreservesAfterTaskUserDeletion(t *testing.T) {
	setupBillingTest(t)

	// Also need User and Task tables for this test
	if err := database.DB.AutoMigrate(&database.User{}, &database.Task{}); err != nil {
		t.Fatalf("AutoMigrate User/Task: %v", err)
	}

	// Create user and task
	database.DB.Create(&database.User{ID: "u-del", Label: "u", Role: "user", Status: "active", CreatedAt: 1000})
	database.DB.Create(&database.Task{ID: "t-del", UserID: "u-del", Prompt: "p", ParamsJSON: "{}", InputImageIDsJSON: "[]", OutputImageIDsJSON: "[]", Status: "completed", CreatedAt: 1000})

	// Record billing
	input := BillingBatchInput{
		TaskID:            "t-del",
		UserID:            "u-del",
		UserLabelSnapshot: "u",
		UnitSaleX10000:    100000,
		Images: []BillingImageInput{
			{OutputImageID: "img-del-1", EndpointBaseURLSnapshot: "https://api.example.com", UnitCostX10000: 50000},
		},
	}

	err := RecordBillingForSuccessfulImages(input)
	if err != nil {
		t.Fatalf("RecordBillingForSuccessfulImages: %v", err)
	}

	var count int64
	database.DB.Model(&database.BillingRecord{}).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 billing row, got %d", count)
	}

	// Delete task
	database.DB.Delete(&database.Task{}, "id = ?", "t-del")
	database.DB.Model(&database.BillingRecord{}).Count(&count)
	if count != 1 {
		t.Fatalf("billing row lost after task deletion: got %d, want 1", count)
	}

	// Delete user
	database.DB.Delete(&database.User{}, "id = ?", "u-del")
	database.DB.Model(&database.BillingRecord{}).Count(&count)
	if count != 1 {
		t.Fatalf("billing row lost after user deletion: got %d, want 1", count)
	}
}

func TestRecordBillingForSuccessfulImages_CreatedAtFallback(t *testing.T) {
	setupBillingTest(t)

	input := BillingBatchInput{
		TaskID:            "task-fb",
		UserID:            "user-fb",
		UserLabelSnapshot: "fb-user",
		UnitSaleX10000:    100000,
		Images: []BillingImageInput{
			{OutputImageID: "img-fb-1", EndpointBaseURLSnapshot: "https://api.example.com", UnitCostX10000: 50000},
		},
		// CreatedAt = 0, should fall back to time.Now().UnixMilli()
	}

	err := RecordBillingForSuccessfulImages(input)
	if err != nil {
		t.Fatalf("RecordBillingForSuccessfulImages: %v", err)
	}

	var row database.BillingRecord
	if err := database.DB.First(&row).Error; err != nil {
		t.Fatalf("read billing row: %v", err)
	}
	if row.CreatedAt == 0 {
		t.Fatal("CreatedAt should be set to time.Now().UnixMilli() when input is 0")
	}
}
