package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAnalyticsHandlerTest(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.json")
	initial := `{"port":3001,"jwtSecret":"test","adminApikey":"test-admin","apiEndpoints":[]}`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	origRootFn := config.GetRootDir()()
	config.SetRootDir(func() string { return tmp })
	t.Cleanup(func() { config.SetRootDir(func() string { return origRootFn }) })

	config.App = nil
	if err := config.Load(); err != nil {
		t.Fatalf("config.Load: %v", err)
	}

	// Setup test DB with billing_records table
	db, err := gorm.Open(sqlite.Open(filepath.Join(tmp, "test.sqlite")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	database.DB = db
	if err := database.DB.AutoMigrate(&database.BillingRecord{}); err != nil {
		t.Fatalf("AutoMigrate BillingRecord: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := database.DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		database.DB = nil
	})

	r := gin.New()
	adminAuth := r.Group("/api/admin")
	adminAuth.GET("/analytics/summary", AdminBillingSummary)
	adminAuth.GET("/analytics/trend", AdminBillingTrend)
	adminAuth.GET("/analytics/endpoints", AdminBillingEndpointBreakdown)
	adminAuth.GET("/analytics/users", AdminBillingUserBreakdown)

	return r
}

func adminAnalyticsToken(t *testing.T) string {
	t.Helper()
	token, err := service.SignToken("admin", "admin", config.App.JWTSecret)
	if err != nil {
		t.Fatalf("sign admin token: %v", err)
	}
	return token
}

func doAdminAnalyticsGet(t *testing.T, r *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+adminAnalyticsToken(t))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}

func TestAdminBillingSummary_DefaultRange(t *testing.T) {
	r := setupAnalyticsHandlerTest(t)

	resp := doAdminAnalyticsGet(t, r, "/api/admin/analytics/summary")

	if resp.Code != http.StatusOK {
		t.Fatalf("GET summary: expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var body struct {
		Meta struct {
			Range      string `json:"range"`
			From       int64  `json:"from"`
			To         int64  `json:"to"`
			MoneyScale int64  `json:"moneyScale"`
		} `json:"meta"`
		Summary struct {
			RevenueX10000 int64 `json:"revenueX10000"`
			CostX10000    int64 `json:"costX10000"`
			ProfitX10000  int64 `json:"profitX10000"`
			SuccessImages int   `json:"successImages"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Meta.MoneyScale != 10000 {
		t.Errorf("MoneyScale = %d, want 10000", body.Meta.MoneyScale)
	}
	// Default range should be 7d
	if body.Meta.Range != "7d" {
		t.Errorf("Range = %q, want 7d", body.Meta.Range)
	}
	if body.Meta.From == 0 || body.Meta.To == 0 {
		t.Errorf("meta.from/to should be non-zero, got from=%d to=%d", body.Meta.From, body.Meta.To)
	}
}

func TestAdminBillingSummary_ExplicitRange(t *testing.T) {
	r := setupAnalyticsHandlerTest(t)

	resp := doAdminAnalyticsGet(t, r, "/api/admin/analytics/summary?range=today")

	if resp.Code != http.StatusOK {
		t.Fatalf("GET summary?range=today: expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var body struct {
		Meta struct {
			Range      string `json:"range"`
			From       int64  `json:"from"`
			To         int64  `json:"to"`
			MoneyScale int64  `json:"moneyScale"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Meta.Range != "today" {
		t.Errorf("Range = %q, want today", body.Meta.Range)
	}
	if body.Meta.MoneyScale != 10000 {
		t.Errorf("MoneyScale = %d, want 10000", body.Meta.MoneyScale)
	}
}

func TestAdminBillingSummary_InvalidRange(t *testing.T) {
	r := setupAnalyticsHandlerTest(t)

	resp := doAdminAnalyticsGet(t, r, "/api/admin/analytics/summary?range=90d")

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("GET summary?range=90d: expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Error == "" {
		t.Error("expected non-empty error message for invalid range")
	}
}

func TestAdminBillingTrend_MoneyScaleInMeta(t *testing.T) {
	r := setupAnalyticsHandlerTest(t)

	resp := doAdminAnalyticsGet(t, r, "/api/admin/analytics/trend?range=7d")

	if resp.Code != http.StatusOK {
		t.Fatalf("GET trend: expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var body struct {
		Meta struct {
			Range      string `json:"range"`
			MoneyScale int64  `json:"moneyScale"`
		} `json:"meta"`
		Trend []struct {
			Bucket string `json:"bucket"`
		} `json:"trend"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Meta.MoneyScale != 10000 {
		t.Errorf("MoneyScale = %d, want 10000", body.Meta.MoneyScale)
	}
	if body.Meta.Range != "7d" {
		t.Errorf("Range = %q, want 7d", body.Meta.Range)
	}
}

func TestAdminBillingTrend_InvalidRange(t *testing.T) {
	r := setupAnalyticsHandlerTest(t)

	resp := doAdminAnalyticsGet(t, r, "/api/admin/analytics/trend?range=90d")

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("GET trend?range=90d: expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestAdminBillingEndpointBreakdown_MoneyScaleInMeta(t *testing.T) {
	r := setupAnalyticsHandlerTest(t)

	resp := doAdminAnalyticsGet(t, r, "/api/admin/analytics/endpoints?range=30d")

	if resp.Code != http.StatusOK {
		t.Fatalf("GET endpoints: expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var body struct {
		Meta struct {
			Range      string `json:"range"`
			MoneyScale int64  `json:"moneyScale"`
		} `json:"meta"`
		Rows []struct {
			EndpointBaseURL string `json:"endpointBaseUrl"`
			RevenueX10000   int64  `json:"revenueX10000"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Meta.MoneyScale != 10000 {
		t.Errorf("MoneyScale = %d, want 10000", body.Meta.MoneyScale)
	}
	if body.Meta.Range != "30d" {
		t.Errorf("Range = %q, want 30d", body.Meta.Range)
	}
}

func TestAdminBillingEndpointBreakdown_InvalidRange(t *testing.T) {
	r := setupAnalyticsHandlerTest(t)

	resp := doAdminAnalyticsGet(t, r, "/api/admin/analytics/endpoints?range=90d")

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("GET endpoints?range=90d: expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestAdminBillingUserBreakdown_MoneyScaleInMeta(t *testing.T) {
	r := setupAnalyticsHandlerTest(t)

	resp := doAdminAnalyticsGet(t, r, "/api/admin/analytics/users?range=all")

	if resp.Code != http.StatusOK {
		t.Fatalf("GET users: expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var body struct {
		Meta struct {
			Range      string `json:"range"`
			MoneyScale int64  `json:"moneyScale"`
		} `json:"meta"`
		Rows []struct {
			UserID    string `json:"userId"`
			UserLabel string `json:"userLabel"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Meta.MoneyScale != 10000 {
		t.Errorf("MoneyScale = %d, want 10000", body.Meta.MoneyScale)
	}
	if body.Meta.Range != "all" {
		t.Errorf("Range = %q, want all", body.Meta.Range)
	}
}

func TestAdminBillingUserBreakdown_InvalidRange(t *testing.T) {
	r := setupAnalyticsHandlerTest(t)

	resp := doAdminAnalyticsGet(t, r, "/api/admin/analytics/users?range=90d")

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("GET users?range=90d: expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestAdminBillingSummary_ContainsData(t *testing.T) {
	r := setupAnalyticsHandlerTest(t)

	// Use a timestamp slightly in the past so it's guaranteed within the 7d range
	oneHourAgo := time.Now().Add(-1 * time.Hour).UnixMilli()

	database.DB.Create(&database.BillingRecord{
		ID:                      "test-br-1",
		TaskID:                  "task-1",
		UserID:                  "user-1",
		UserLabelSnapshot:       "Alice",
		EndpointBaseURLSnapshot: "https://api.example.com",
		OutputImageID:           "img-1",
		SuccessImageCount:       2,
		UnitCostX10000:          10000,
		UnitSaleX10000:          50000,
		CostX10000:              10000,
		RevenueX10000:           50000,
		ProfitX10000:            40000,
		CreatedAt:               oneHourAgo,
	})

	resp := doAdminAnalyticsGet(t, r, "/api/admin/analytics/summary?range=7d")

	if resp.Code != http.StatusOK {
		t.Fatalf("GET summary: expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var body struct {
		Meta struct {
			Range string `json:"range"`
		} `json:"meta"`
		Summary struct {
			RevenueX10000 int64 `json:"revenueX10000"`
			SuccessImages int   `json:"successImages"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Summary.RevenueX10000 != 50000 {
		t.Errorf("RevenueX10000 = %d, want 50000", body.Summary.RevenueX10000)
	}
	if body.Summary.SuccessImages != 2 {
		t.Errorf("SuccessImages = %d, want 2", body.Summary.SuccessImages)
	}
}
