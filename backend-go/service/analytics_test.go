package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/util"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAnalyticsTest(t *testing.T) time.Time {
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
		database.DB = nil
	})

	return time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
}

// seedBillingRows inserts BillingRecord rows for testing analytics.
func seedBillingRows(t *testing.T, records []database.BillingRecord) {
	t.Helper()
	for i := range records {
		if records[i].ID == "" {
			records[i].ID = util.GenerateID()
		}
	}
	if err := database.DB.Create(&records).Error; err != nil {
		t.Fatalf("seed billing rows: %v", err)
	}
}

// --- ParseAnalyticsRange tests ---

func TestParseAnalyticsRange_Today(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 30, 0, 0, time.UTC)

	r, err := ParseAnalyticsRange("today", now)
	if err != nil {
		t.Fatalf("ParseAnalyticsRange: %v", err)
	}
	if r.Label != "today" {
		t.Errorf("Label = %q, want today", r.Label)
	}

	expectedFrom := time.Date(2026, 5, 23, 0, 0, 0, 0, time.UTC)
	if r.From != expectedFrom.UnixMilli() {
		t.Errorf("From = %d, want %d (start of today)", r.From, expectedFrom.UnixMilli())
	}
	if r.To != now.UnixMilli() {
		t.Errorf("To = %d, want %d", r.To, now.UnixMilli())
	}
}

func TestParseAnalyticsRange_7d(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)

	r, err := ParseAnalyticsRange("7d", now)
	if err != nil {
		t.Fatalf("ParseAnalyticsRange: %v", err)
	}
	if r.Label != "7d" {
		t.Errorf("Label = %q, want 7d", r.Label)
	}

	expectedFrom := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	if r.From != expectedFrom.UnixMilli() {
		t.Errorf("From = %d, want %d", r.From, expectedFrom.UnixMilli())
	}
	if r.To != now.UnixMilli() {
		t.Errorf("To = %d, want %d", r.To, now.UnixMilli())
	}
}

func TestParseAnalyticsRange_30d(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)

	r, err := ParseAnalyticsRange("30d", now)
	if err != nil {
		t.Fatalf("ParseAnalyticsRange: %v", err)
	}
	if r.Label != "30d" {
		t.Errorf("Label = %q, want 30d", r.Label)
	}

	expectedFrom := time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)
	if r.From != expectedFrom.UnixMilli() {
		t.Errorf("From = %d, want %d", r.From, expectedFrom.UnixMilli())
	}
	if r.To != now.UnixMilli() {
		t.Errorf("To = %d, want %d", r.To, now.UnixMilli())
	}
}

func TestParseAnalyticsRange_EmptyDefaultsTo7d(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)

	r7d, _ := ParseAnalyticsRange("7d", now)
	rEmpty, err := ParseAnalyticsRange("", now)
	if err != nil {
		t.Fatalf("ParseAnalyticsRange empty: %v", err)
	}
	if rEmpty.Label != "7d" {
		t.Errorf("empty Label = %q, want 7d", rEmpty.Label)
	}
	if rEmpty.From != r7d.From {
		t.Errorf("empty From = %d, want %d (same as 7d)", rEmpty.From, r7d.From)
	}
	if rEmpty.To != r7d.To {
		t.Errorf("empty To = %d, want %d (same as 7d)", rEmpty.To, r7d.To)
	}
}

func TestParseAnalyticsRange_All(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)

	r, err := ParseAnalyticsRange("all", now)
	if err != nil {
		t.Fatalf("ParseAnalyticsRange all: %v", err)
	}
	if r.Label != "all" {
		t.Errorf("Label = %q, want all", r.Label)
	}
	if r.From != 0 {
		t.Errorf("From = %d, want 0", r.From)
	}
	if r.To != now.UnixMilli() {
		t.Errorf("To = %d, want %d", r.To, now.UnixMilli())
	}
}

func TestParseAnalyticsRange_Invalid(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)

	_, err := ParseAnalyticsRange("90d", now)
	if err == nil {
		t.Fatal("expected error for 90d")
	}

	_, err = ParseAnalyticsRange("week", now)
	if err == nil {
		t.Fatal("expected error for 'week'")
	}
}

// --- GetBillingSummary tests ---

func TestGetBillingSummary_EmptyRange(t *testing.T) {
	now := setupAnalyticsTest(t)
	r, err := ParseAnalyticsRange("7d", now)
	if err != nil {
		t.Fatalf("ParseAnalyticsRange: %v", err)
	}

	summary, meta, err := GetBillingSummary(r)
	if err != nil {
		t.Fatalf("GetBillingSummary: %v", err)
	}
	if summary.RevenueX10000 != 0 {
		t.Errorf("RevenueX10000 = %d, want 0", summary.RevenueX10000)
	}
	if summary.CostX10000 != 0 {
		t.Errorf("CostX10000 = %d, want 0", summary.CostX10000)
	}
	if summary.ProfitX10000 != 0 {
		t.Errorf("ProfitX10000 = %d, want 0", summary.ProfitX10000)
	}
	if summary.SuccessImages != 0 {
		t.Errorf("SuccessImages = %d, want 0", summary.SuccessImages)
	}
	if meta.MoneyScale != MoneyScale {
		t.Errorf("MoneyScale = %d, want %d", meta.MoneyScale, MoneyScale)
	}
}

func TestGetBillingSummary_7dRange(t *testing.T) {
	now := setupAnalyticsTest(t)

	// Seed rows at different offsets
	today6AM := now.Add(-6 * time.Hour).UnixMilli()     // within 7d
	yesterdayNoon := now.Add(-24 * time.Hour).UnixMilli() // within 7d
	tenDaysAgo := now.Add(-240 * time.Hour).UnixMilli()   // outside 7d

	seedBillingRows(t, []database.BillingRecord{
		{TaskID: "t1", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img1", SuccessImageCount: 2, UnitCostX10000: 10000, UnitSaleX10000: 50000, CostX10000: 10000, RevenueX10000: 50000, ProfitX10000: 40000, CreatedAt: today6AM},
		{TaskID: "t2", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img2", SuccessImageCount: 1, UnitCostX10000: 20000, UnitSaleX10000: 50000, CostX10000: 20000, RevenueX10000: 50000, ProfitX10000: 30000, CreatedAt: yesterdayNoon},
		{TaskID: "t3", UserID: "u2", UserLabelSnapshot: "Bob", EndpointBaseURLSnapshot: "https://api2.example.com", OutputImageID: "img3", SuccessImageCount: 3, UnitCostX10000: 50000, UnitSaleX10000: 50000, CostX10000: 50000, RevenueX10000: 50000, ProfitX10000: 0, CreatedAt: tenDaysAgo},
	})

	r, err := ParseAnalyticsRange("7d", now)
	if err != nil {
		t.Fatalf("ParseAnalyticsRange: %v", err)
	}

	summary, meta, err := GetBillingSummary(r)
	if err != nil {
		t.Fatalf("GetBillingSummary: %v", err)
	}

	// Only t1 and t2 are in range: revenue = 50000+50000 = 100000
	if summary.RevenueX10000 != 100000 {
		t.Errorf("RevenueX10000 = %d, want 100000", summary.RevenueX10000)
	}
	// cost = 10000+20000 = 30000
	if summary.CostX10000 != 30000 {
		t.Errorf("CostX10000 = %d, want 30000", summary.CostX10000)
	}
	// profit = 40000+30000 = 70000
	if summary.ProfitX10000 != 70000 {
		t.Errorf("ProfitX10000 = %d, want 70000", summary.ProfitX10000)
	}
	// successImages = 2+1 = 3
	if summary.SuccessImages != 3 {
		t.Errorf("SuccessImages = %d, want 3", summary.SuccessImages)
	}
	if meta.MoneyScale != MoneyScale {
		t.Errorf("MoneyScale = %d, want %d", meta.MoneyScale, MoneyScale)
	}
}

func TestGetBillingSummary_TodayRange(t *testing.T) {
	now := setupAnalyticsTest(t)

	today6AM := now.Add(-6 * time.Hour).UnixMilli()
	yesterdayNoon := now.Add(-24 * time.Hour).UnixMilli()

	seedBillingRows(t, []database.BillingRecord{
		{TaskID: "t1", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img1", SuccessImageCount: 1, UnitCostX10000: 10000, UnitSaleX10000: 100000, CostX10000: 10000, RevenueX10000: 100000, ProfitX10000: 90000, CreatedAt: today6AM},
		{TaskID: "t2", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img2", SuccessImageCount: 1, UnitCostX10000: 20000, UnitSaleX10000: 100000, CostX10000: 20000, RevenueX10000: 100000, ProfitX10000: 80000, CreatedAt: yesterdayNoon},
	})

	r, _ := ParseAnalyticsRange("today", now)
	summary, _, err := GetBillingSummary(r)
	if err != nil {
		t.Fatalf("GetBillingSummary: %v", err)
	}
	// Only today's row is in range
	if summary.SuccessImages != 1 {
		t.Errorf("SuccessImages = %d, want 1", summary.SuccessImages)
	}
	if summary.RevenueX10000 != 100000 {
		t.Errorf("RevenueX10000 = %d, want 100000", summary.RevenueX10000)
	}
}

func TestGetBillingSummary_AllRange(t *testing.T) {
	now := setupAnalyticsTest(t)

	fortyDaysAgo := now.Add(-960 * time.Hour).UnixMilli()
	today := now.UnixMilli()

	seedBillingRows(t, []database.BillingRecord{
		{TaskID: "t1", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img1", SuccessImageCount: 1, UnitCostX10000: 10000, UnitSaleX10000: 10000, CostX10000: 10000, RevenueX10000: 10000, ProfitX10000: 0, CreatedAt: fortyDaysAgo},
		{TaskID: "t2", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img2", SuccessImageCount: 1, UnitCostX10000: 10000, UnitSaleX10000: 10000, CostX10000: 10000, RevenueX10000: 10000, ProfitX10000: 0, CreatedAt: today},
	})

	r, _ := ParseAnalyticsRange("all", now)
	summary, _, err := GetBillingSummary(r)
	if err != nil {
		t.Fatalf("GetBillingSummary: %v", err)
	}
	// Both rows included
	if summary.SuccessImages != 2 {
		t.Errorf("SuccessImages = %d, want 2", summary.SuccessImages)
	}
	if summary.RevenueX10000 != 20000 {
		t.Errorf("RevenueX10000 = %d, want 20000", summary.RevenueX10000)
	}
}

// --- GetBillingTrend tests ---

func TestGetBillingTrend_MultipleBuckets(t *testing.T) {
	now := setupAnalyticsTest(t)

	today6AM := now.Add(-6 * time.Hour).UnixMilli()
	yesterdayNoon := now.Add(-24 * time.Hour).UnixMilli()
	threeDaysAgo := now.Add(-72 * time.Hour).UnixMilli()

	seedBillingRows(t, []database.BillingRecord{
		{TaskID: "t1", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img1", SuccessImageCount: 2, UnitCostX10000: 10000, UnitSaleX10000: 50000, CostX10000: 10000, RevenueX10000: 50000, ProfitX10000: 40000, CreatedAt: today6AM},
		{TaskID: "t2", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img2", SuccessImageCount: 1, UnitCostX10000: 10000, UnitSaleX10000: 30000, CostX10000: 10000, RevenueX10000: 30000, ProfitX10000: 20000, CreatedAt: today6AM},
		{TaskID: "t3", UserID: "u2", UserLabelSnapshot: "Bob", EndpointBaseURLSnapshot: "https://api2.example.com", OutputImageID: "img3", SuccessImageCount: 1, UnitCostX10000: 20000, UnitSaleX10000: 50000, CostX10000: 20000, RevenueX10000: 50000, ProfitX10000: 30000, CreatedAt: yesterdayNoon},
		{TaskID: "t4", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img4", SuccessImageCount: 1, UnitCostX10000: 10000, UnitSaleX10000: 40000, CostX10000: 10000, RevenueX10000: 40000, ProfitX10000: 30000, CreatedAt: threeDaysAgo},
	})

	r, _ := ParseAnalyticsRange("7d", now)
	points, meta, err := GetBillingTrend(r)
	if err != nil {
		t.Fatalf("GetBillingTrend: %v", err)
	}

	// Expect 3 buckets: 2026-05-23, 2026-05-22, 2026-05-20
	if len(points) != 3 {
		t.Fatalf("got %d buckets, want 3", len(points))
	}

	// Check ordering (ascending by bucket)
	if points[0].Bucket >= points[1].Bucket || points[1].Bucket >= points[2].Bucket {
		t.Errorf("buckets not in ascending order: %v", points)
	}

	// 2026-05-23: 2 rows (t1 + t2): images=3, revenue=80000, cost=20000, profit=60000
	if points[2].Bucket != "2026-05-23" {
		t.Errorf("last bucket = %q, want 2026-05-23", points[2].Bucket)
	}
	if points[2].SuccessImages != 3 {
		t.Errorf("2026-05-23 SuccessImages = %d, want 3", points[2].SuccessImages)
	}
	if points[2].RevenueX10000 != 80000 {
		t.Errorf("2026-05-23 RevenueX10000 = %d, want 80000", points[2].RevenueX10000)
	}

	// 2026-05-22: 1 row: images=1, revenue=50000
	if points[1].Bucket != "2026-05-22" {
		t.Errorf("middle bucket = %q, want 2026-05-22", points[1].Bucket)
	}

	// 2026-05-20: 1 row: images=1, revenue=40000
	if points[0].Bucket != "2026-05-20" {
		t.Errorf("first bucket = %q, want 2026-05-20", points[0].Bucket)
	}

	if meta.MoneyScale != MoneyScale {
		t.Errorf("MoneyScale = %d, want %d", meta.MoneyScale, MoneyScale)
	}
}

func TestGetBillingTrend_EmptyRange(t *testing.T) {
	now := setupAnalyticsTest(t)

	r, _ := ParseAnalyticsRange("7d", now)
	points, _, err := GetBillingTrend(r)
	if err != nil {
		t.Fatalf("GetBillingTrend: %v", err)
	}
	if len(points) != 0 {
		t.Errorf("expected 0 points for empty range, got %d", len(points))
	}
}

// --- GetBillingEndpointBreakdown tests ---

func TestGetBillingEndpointBreakdown(t *testing.T) {
	now := setupAnalyticsTest(t)

	today := now.UnixMilli()

	seedBillingRows(t, []database.BillingRecord{
		// api1: revenue=130000, cost=20000, profit=100000, images=3
		{TaskID: "t1", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img1", SuccessImageCount: 1, UnitCostX10000: 10000, UnitSaleX10000: 50000, CostX10000: 10000, RevenueX10000: 50000, ProfitX10000: 40000, CreatedAt: today},
		{TaskID: "t2", UserID: "u2", UserLabelSnapshot: "Bob", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img2", SuccessImageCount: 2, UnitCostX10000: 10000, UnitSaleX10000: 40000, CostX10000: 10000, RevenueX10000: 80000, ProfitX10000: 60000, CreatedAt: today},
		// api2: revenue=60000, cost=20000, profit=40000, images=1
		{TaskID: "t3", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api2.example.com", OutputImageID: "img3", SuccessImageCount: 1, UnitCostX10000: 20000, UnitSaleX10000: 60000, CostX10000: 20000, RevenueX10000: 60000, ProfitX10000: 40000, CreatedAt: today},
	})

	r, _ := ParseAnalyticsRange("7d", now)
	rows, meta, err := GetBillingEndpointBreakdown(r)
	if err != nil {
		t.Fatalf("GetBillingEndpointBreakdown: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("got %d endpoint rows, want 2", len(rows))
	}

	// Sorted by profit desc: api1 (profit=100000) should be first
	if rows[0].EndpointBaseURL != "https://api1.example.com" {
		t.Errorf("first row endpoint = %q, want api1 (highest profit)", rows[0].EndpointBaseURL)
	}
	if rows[1].EndpointBaseURL != "https://api2.example.com" {
		t.Errorf("second row endpoint = %q, want api2", rows[1].EndpointBaseURL)
	}

	// api1 aggregates
	if rows[0].RevenueX10000 != 130000 {
		t.Errorf("api1 RevenueX10000 = %d, want 130000", rows[0].RevenueX10000)
	}
	if rows[0].CostX10000 != 20000 {
		t.Errorf("api1 CostX10000 = %d, want 20000", rows[0].CostX10000)
	}
	if rows[0].ProfitX10000 != 100000 {
		t.Errorf("api1 ProfitX10000 = %d, want 100000", rows[0].ProfitX10000)
	}
	if rows[0].SuccessImages != 3 {
		t.Errorf("api1 SuccessImages = %d, want 3", rows[0].SuccessImages)
	}
	// profitRateBps = 100000 * 10000 / 130000 = 7692 (integer division)
	expectedBps := int64(100000 * 10000 / 130000)
	if rows[0].ProfitRateBps != expectedBps {
		t.Errorf("api1 ProfitRateBps = %d, want %d", rows[0].ProfitRateBps, expectedBps)
	}

	// api2 aggregates
	if rows[1].RevenueX10000 != 60000 {
		t.Errorf("api2 RevenueX10000 = %d, want 60000", rows[1].RevenueX10000)
	}
	if rows[1].ProfitX10000 != 40000 {
		t.Errorf("api2 ProfitX10000 = %d, want 40000", rows[1].ProfitX10000)
	}
	expectedBps2 := int64(40000 * 10000 / 60000)
	if rows[1].ProfitRateBps != expectedBps2 {
		t.Errorf("api2 ProfitRateBps = %d, want %d", rows[1].ProfitRateBps, expectedBps2)
	}

	if meta.MoneyScale != MoneyScale {
		t.Errorf("MoneyScale = %d, want %d", meta.MoneyScale, MoneyScale)
	}
}

func TestGetBillingEndpointBreakdown_ZeroRevenue(t *testing.T) {
	now := setupAnalyticsTest(t)
	today := now.UnixMilli()

	seedBillingRows(t, []database.BillingRecord{
		{TaskID: "t1", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://free.example.com", OutputImageID: "img1", SuccessImageCount: 1, UnitCostX10000: 0, UnitSaleX10000: 0, CostX10000: 0, RevenueX10000: 0, ProfitX10000: 0, CreatedAt: today},
	})

	r, _ := ParseAnalyticsRange("7d", now)
	rows, _, err := GetBillingEndpointBreakdown(r)
	if err != nil {
		t.Fatalf("GetBillingEndpointBreakdown: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	if rows[0].ProfitRateBps != 0 {
		t.Errorf("ProfitRateBps = %d, want 0 (zero revenue)", rows[0].ProfitRateBps)
	}
}

// --- GetBillingUserBreakdown tests ---

func TestGetBillingUserBreakdown(t *testing.T) {
	now := setupAnalyticsTest(t)
	today := now.UnixMilli()

	seedBillingRows(t, []database.BillingRecord{
		// Alice: revenue=100000, cost=20000, profit=80000, images=2
		{TaskID: "t1", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img1", SuccessImageCount: 1, UnitCostX10000: 10000, UnitSaleX10000: 40000, CostX10000: 10000, RevenueX10000: 40000, ProfitX10000: 30000, CreatedAt: today},
		{TaskID: "t2", UserID: "u1", UserLabelSnapshot: "Alice", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img2", SuccessImageCount: 1, UnitCostX10000: 10000, UnitSaleX10000: 60000, CostX10000: 10000, RevenueX10000: 60000, ProfitX10000: 50000, CreatedAt: today},
		// Bob: revenue=50000, cost=20000, profit=30000, images=1
		{TaskID: "t3", UserID: "u2", UserLabelSnapshot: "Bob", EndpointBaseURLSnapshot: "https://api2.example.com", OutputImageID: "img3", SuccessImageCount: 1, UnitCostX10000: 20000, UnitSaleX10000: 50000, CostX10000: 20000, RevenueX10000: 50000, ProfitX10000: 30000, CreatedAt: today},
	})

	r, _ := ParseAnalyticsRange("7d", now)
	rows, meta, err := GetBillingUserBreakdown(r)
	if err != nil {
		t.Fatalf("GetBillingUserBreakdown: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("got %d user rows, want 2", len(rows))
	}

	// Sorted by profit desc: Alice (80000) first
	if rows[0].UserID != "u1" {
		t.Errorf("first row UserID = %q, want u1 (Alice, highest profit)", rows[0].UserID)
	}
	if rows[0].UserLabel != "Alice" {
		t.Errorf("first row UserLabel = %q, want Alice", rows[0].UserLabel)
	}
	if rows[0].RevenueX10000 != 100000 {
		t.Errorf("Alice RevenueX10000 = %d, want 100000", rows[0].RevenueX10000)
	}
	if rows[0].CostX10000 != 20000 {
		t.Errorf("Alice CostX10000 = %d, want 20000", rows[0].CostX10000)
	}
	if rows[0].ProfitX10000 != 80000 {
		t.Errorf("Alice ProfitX10000 = %d, want 80000", rows[0].ProfitX10000)
	}
	if rows[0].SuccessImages != 2 {
		t.Errorf("Alice SuccessImages = %d, want 2", rows[0].SuccessImages)
	}
	expectedBps := int64(80000 * 10000 / 100000)
	if rows[0].ProfitRateBps != expectedBps {
		t.Errorf("Alice ProfitRateBps = %d, want %d", rows[0].ProfitRateBps, expectedBps)
	}

	// Bob second
	if rows[1].UserID != "u2" {
		t.Errorf("second row UserID = %q, want u2", rows[1].UserID)
	}
	if rows[1].UserLabel != "Bob" {
		t.Errorf("second row UserLabel = %q, want Bob", rows[1].UserLabel)
	}
	if rows[1].ProfitX10000 != 30000 {
		t.Errorf("Bob ProfitX10000 = %d, want 30000", rows[1].ProfitX10000)
	}

	if meta.MoneyScale != MoneyScale {
		t.Errorf("MoneyScale = %d, want %d", meta.MoneyScale, MoneyScale)
	}
}

func TestGetBillingUserBreakdown_ZeroRevenue(t *testing.T) {
	now := setupAnalyticsTest(t)
	today := now.UnixMilli()

	seedBillingRows(t, []database.BillingRecord{
		{TaskID: "t1", UserID: "u-free", UserLabelSnapshot: "FreeUser", EndpointBaseURLSnapshot: "https://api1.example.com", OutputImageID: "img1", SuccessImageCount: 1, UnitCostX10000: 0, UnitSaleX10000: 0, CostX10000: 0, RevenueX10000: 0, ProfitX10000: 0, CreatedAt: today},
	})

	r, _ := ParseAnalyticsRange("7d", now)
	rows, _, err := GetBillingUserBreakdown(r)
	if err != nil {
		t.Fatalf("GetBillingUserBreakdown: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	if rows[0].ProfitRateBps != 0 {
		t.Errorf("ProfitRateBps = %d, want 0 (zero revenue)", rows[0].ProfitRateBps)
	}
}
