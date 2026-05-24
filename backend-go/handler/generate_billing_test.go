package handler

import (
	"os"
	"path/filepath"
	"testing"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBillingHandlerTest(t *testing.T) {
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
		t.Fatalf("AutoMigrate BillingRecord: %v", err)
	}
	// Need Image table for saveGeneratedImages to work
	if err := database.DB.AutoMigrate(&database.Image{}); err != nil {
		t.Fatalf("AutoMigrate Image: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := database.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
}

func TestSaveGeneratedImagesWithAttribution_ReturnsCorrectPairings(t *testing.T) {
	setupBillingHandlerTest(t)

	images := []service.GeneratedImage{
		{Base64: "data:image/png;base64,iVBORw0KGgo=", EndpointBaseURL: "https://ep1.example/v1", UnitCostX10000: 1000, RevisedPrompt: "prompt1"},
		{Base64: "data:image/png;base64,iVBORw0KGgo=", EndpointBaseURL: "https://ep1.example/v1", UnitCostX10000: 1000, RevisedPrompt: "prompt2"},
		{Base64: "data:image/png;base64,iVBORw0KGgo=", EndpointBaseURL: "https://ep2.example/v1", UnitCostX10000: 2000, RevisedPrompt: "prompt3"},
	}

	saved := saveGeneratedImagesWithAttribution("user-test", images)

	if len(saved) != 3 {
		t.Fatalf("expected 3 saved images, got %d", len(saved))
	}

	for i, s := range saved {
		if s.OutputImageID == "" {
			t.Errorf("image %d: OutputImageID is empty", i)
		}
		if s.Generated.EndpointBaseURL != images[i].EndpointBaseURL {
			t.Errorf("image %d: EndpointBaseURL = %q, want %q", i, s.Generated.EndpointBaseURL, images[i].EndpointBaseURL)
		}
		if s.Generated.UnitCostX10000 != images[i].UnitCostX10000 {
			t.Errorf("image %d: UnitCostX10000 = %d, want %d", i, s.Generated.UnitCostX10000, images[i].UnitCostX10000)
		}
		if s.Generated.RevisedPrompt != images[i].RevisedPrompt {
			t.Errorf("image %d: RevisedPrompt = %q, want %q", i, s.Generated.RevisedPrompt, images[i].RevisedPrompt)
		}
	}
}

func TestSaveGeneratedImagesWithAttribution_PartialFailureDoesNotShiftPairings(t *testing.T) {
	setupBillingHandlerTest(t)

	// Second image has invalid base64 — should fail to save
	images := []service.GeneratedImage{
		{Base64: "data:image/png;base64,iVBORw0KGgo=", EndpointBaseURL: "https://ep1.example/v1", UnitCostX10000: 1000, RevisedPrompt: "good"},
		{Base64: "not-a-valid-data-url", EndpointBaseURL: "https://ep1.example/v1", UnitCostX10000: 1000, RevisedPrompt: "bad"},
		{Base64: "data:image/png;base64,iVBORw0KGgo=", EndpointBaseURL: "https://ep2.example/v1", UnitCostX10000: 2000, RevisedPrompt: "good2"},
	}

	saved := saveGeneratedImagesWithAttribution("user-test", images)

	if len(saved) != 2 {
		t.Fatalf("expected 2 saved images (middle one failed), got %d", len(saved))
	}

	// First saved entry should be image[0], not image[1]
	if saved[0].Generated.RevisedPrompt != "good" {
		t.Errorf("saved[0]: RevisedPrompt = %q, want 'good' (not shifted to 'bad')", saved[0].Generated.RevisedPrompt)
	}
	if saved[0].Generated.EndpointBaseURL != "https://ep1.example/v1" {
		t.Errorf("saved[0]: EndpointBaseURL = %q, want https://ep1.example/v1", saved[0].Generated.EndpointBaseURL)
	}

	// Second saved entry should be image[2], not image[1]
	if saved[1].Generated.RevisedPrompt != "good2" {
		t.Errorf("saved[1]: RevisedPrompt = %q, want 'good2' (not shifted to 'good')", saved[1].Generated.RevisedPrompt)
	}
	if saved[1].Generated.EndpointBaseURL != "https://ep2.example/v1" {
		t.Errorf("saved[1]: EndpointBaseURL = %q, want https://ep2.example/v1", saved[1].Generated.EndpointBaseURL)
	}
}

func TestRecordBillingForSavedImages_RowCountEqualsSaveCount(t *testing.T) {
	setupBillingHandlerTest(t)

	saved := []savedGeneratedImage{
		{OutputImageID: "img-1", Generated: service.GeneratedImage{EndpointBaseURL: "https://ep1.example/v1", UnitCostX10000: 1000}},
		{OutputImageID: "img-2", Generated: service.GeneratedImage{EndpointBaseURL: "https://ep1.example/v1", UnitCostX10000: 1000}},
	}

	// Set a known sale price
	config.App.SalePriceX10000 = 200000 // 20.00 yuan

	billingInput := buildBillingInput("task-abc", "user-xyz", "test-user-label", saved)
	err := service.RecordBillingForSuccessfulImages(billingInput)
	if err != nil {
		t.Fatalf("RecordBillingForSuccessfulImages: %v", err)
	}

	// Verify billing count matches saved count, not result count or requested n
	var count int64
	if err := database.DB.Model(&database.BillingRecord{}).Count(&count).Error; err != nil {
		t.Fatalf("query billing count: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 billing rows (matches SaveDataURLImage success count), got %d", count)
	}

	// Verify snapshots are used
	var rows []database.BillingRecord
	database.DB.Order("output_image_id asc").Find(&rows)
	for _, r := range rows {
		if r.TaskID != "task-abc" {
			t.Errorf("TaskID = %q, want task-abc", r.TaskID)
		}
		if r.UserID != "user-xyz" {
			t.Errorf("UserID = %q, want user-xyz", r.UserID)
		}
		if r.UserLabelSnapshot != "test-user-label" {
			t.Errorf("UserLabelSnapshot = %q, want test-user-label", r.UserLabelSnapshot)
		}
		if r.UnitSaleX10000 != 200000 {
			t.Errorf("UnitSaleX10000 = %d, want 200000 (snapshot from config)", r.UnitSaleX10000)
		}
		if r.SuccessImageCount != 1 {
			t.Errorf("SuccessImageCount = %d, want 1", r.SuccessImageCount)
		}
	}
}

func TestBuildPerImageMetadata_NoShiftOnPartialSave(t *testing.T) {
	// savedIDs and generatedImages are paired by the caller explicitly.
	// The test verifies that when we have a saved-success slice, the
	// metadata maps are built correctly without index-shifting.
	saved := []savedGeneratedImage{
		{OutputImageID: "id-a", Generated: service.GeneratedImage{ActualParams: map[string]interface{}{"size": "1024x1024"}, RevisedPrompt: "prompt-a"}},
		{OutputImageID: "id-c", Generated: service.GeneratedImage{ActualParams: map[string]interface{}{"size": "512x512"}, RevisedPrompt: "prompt-c"}},
	}

	apBi, rpBi := buildPerImageMetadataFromSaved(saved)

	if len(apBi) != 2 {
		t.Fatalf("expected 2 metadata entries, got %d", len(apBi))
	}
	if apBi["id-a"]["size"] != "1024x1024" {
		t.Errorf("id-a size = %v, want 1024x1024", apBi["id-a"]["size"])
	}
	if apBi["id-c"]["size"] != "512x512" {
		t.Errorf("id-c size = %v, want 512x512", apBi["id-c"]["size"])
	}
	if rpBi["id-a"] != "prompt-a" {
		t.Errorf("id-a revised = %q, want prompt-a", rpBi["id-a"])
	}
	if rpBi["id-c"] != "prompt-c" {
		t.Errorf("id-c revised = %q, want prompt-c", rpBi["id-c"])
	}
}

func TestBuildBillingInput_UsesConfigSalePriceSnapshot(t *testing.T) {
	config.App.SalePriceX10000 = 999900 // 99.99 yuan

	saved := []savedGeneratedImage{
		{OutputImageID: "img-x", Generated: service.GeneratedImage{EndpointBaseURL: "https://ep.example/v1", UnitCostX10000: 50000}},
	}

	billingInput := buildBillingInput("task-sale", "user-sale", "label-sale", saved)

	if billingInput.UnitSaleX10000 != 999900 {
		t.Errorf("UnitSaleX10000 = %d, want 999900 (snapshot from config.GetSalePriceX10000)", billingInput.UnitSaleX10000)
	}
	if billingInput.TaskID != "task-sale" {
		t.Errorf("TaskID = %q, want task-sale", billingInput.TaskID)
	}
	if billingInput.UserID != "user-sale" {
		t.Errorf("UserID = %q, want user-sale", billingInput.UserID)
	}
	if billingInput.UserLabelSnapshot != "label-sale" {
		t.Errorf("UserLabelSnapshot = %q, want label-sale", billingInput.UserLabelSnapshot)
	}
	if len(billingInput.Images) != 1 {
		t.Fatalf("expected 1 billing image input, got %d", len(billingInput.Images))
	}
	if billingInput.Images[0].EndpointBaseURLSnapshot != "https://ep.example/v1" {
		t.Errorf("EndpointBaseURLSnapshot = %q, want https://ep.example/v1", billingInput.Images[0].EndpointBaseURLSnapshot)
	}
}
