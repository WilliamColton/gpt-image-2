package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

func setupPricingHandlerTest(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.json")
	initial := `{"port":3001,"jwtSecret":"test","adminApikey":"test-admin","apiEndpoints":[]}`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Save the original getRootDir function and replace with temp dir
	origRootFn := config.GetRootDir()()
	config.SetRootDir(func() string { return tmp })
	t.Cleanup(func() { config.SetRootDir(func() string { return origRootFn }) })

	// Load config
	config.App = nil
	if err := config.Load(); err != nil {
		t.Fatalf("config.Load: %v", err)
	}

	r := gin.New()
	adminAuth := r.Group("/api/admin")
	adminAuth.GET("/config/pricing", AdminGetPricingConfig)
	adminAuth.PUT("/config/pricing", AdminUpdatePricingConfig)
	adminAuth.GET("/config/endpoints", AdminGetEndpoints)
	adminAuth.PUT("/config/endpoints", AdminUpdateEndpoints)

	return r
}

func adminPricingToken(t *testing.T) string {
	t.Helper()
	token, err := service.SignToken("admin", "admin", config.App.JWTSecret)
	if err != nil {
		t.Fatalf("sign admin token: %v", err)
	}
	return token
}

func TestAdminPricing_GetReturnsMoneyScale(t *testing.T) {
	r := setupPricingHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/config/pricing", nil)
	req.Header.Set("Authorization", "Bearer "+adminPricingToken(t))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("GET pricing: expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var body struct {
		Endpoints       []config.ApiEndpoint `json:"endpoints"`
		SalePriceX10000 int64                `json:"salePriceX10000"`
		MoneyScale      int64                `json:"moneyScale"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.MoneyScale != 10000 {
		t.Errorf("MoneyScale = %d, want 10000", body.MoneyScale)
	}
}

func TestAdminPricing_PutSuccessReturnsMoneyScale(t *testing.T) {
	r := setupPricingHandlerTest(t)

	input := map[string]interface{}{
		"endpoints": []map[string]interface{}{
			{"baseUrl": "https://a.com/v1", "apiKey": "sk-a", "priority": 1, "maxConcurrency": 2, "costPerImageX10000": 12345},
		},
		"salePriceX10000": json.Number("50000"),
	}
	bodyBytes, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/config/pricing", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminPricingToken(t))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("PUT pricing: expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var result struct {
		Ok              bool                 `json:"ok"`
		Endpoints       []config.ApiEndpoint `json:"endpoints"`
		SalePriceX10000 int64                `json:"salePriceX10000"`
		MoneyScale      int64                `json:"moneyScale"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !result.Ok {
		t.Error("ok should be true")
	}
	if result.SalePriceX10000 != 50000 {
		t.Errorf("SalePriceX10000 = %d, want 50000", result.SalePriceX10000)
	}
	if result.MoneyScale != 10000 {
		t.Errorf("MoneyScale = %d, want 10000", result.MoneyScale)
	}
	if len(result.Endpoints) != 1 {
		t.Fatalf("expected 1 endpoint in response, got %d", len(result.Endpoints))
	}
	if result.Endpoints[0].CostPerImageX10000 != 12345 {
		t.Errorf("endpoint cost = %d, want 12345", result.Endpoints[0].CostPerImageX10000)
	}
}

func TestAdminPricing_PutRejectsEmptyEndpointAPIKey(t *testing.T) {
	r := setupPricingHandlerTest(t)

	input := map[string]interface{}{
		"endpoints": []map[string]interface{}{
			{"baseUrl": "https://a.com/v1", "apiKey": "   ", "priority": 1},
		},
		"salePriceX10000": json.Number("50000"),
	}
	bodyBytes, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/config/pricing", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminPricingToken(t))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("PUT pricing: expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestAdminUpdateEndpoints_RejectsEmptyAPIKey(t *testing.T) {
	r := setupPricingHandlerTest(t)

	input := map[string]interface{}{
		"endpoints": []map[string]interface{}{
			{"baseUrl": "https://a.com/v1", "apiKey": "", "priority": 1},
		},
	}
	bodyBytes, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/config/endpoints", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminPricingToken(t))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("PUT endpoints: expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestAdminPricing_PutRejectsNegativeSalePrice(t *testing.T) {
	r := setupPricingHandlerTest(t)

	input := map[string]interface{}{
		"endpoints": []map[string]interface{}{
			{"baseUrl": "https://a.com/v1", "apiKey": "sk-a", "priority": 1},
		},
		"salePriceX10000": json.Number("-1"),
	}
	bodyBytes, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/config/pricing", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminPricingToken(t))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative salePriceX10000, got %d", resp.Code)
	}
}

func TestAdminPricing_PutRejectsNegativeEndpointCost(t *testing.T) {
	r := setupPricingHandlerTest(t)

	input := map[string]interface{}{
		"endpoints": []map[string]interface{}{
			{"baseUrl": "https://a.com/v1", "apiKey": "sk-a", "priority": 1, "costPerImageX10000": json.Number("-1")},
		},
		"salePriceX10000": json.Number("10000"),
	}
	bodyBytes, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/config/pricing", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminPricingToken(t))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative CostPerImageX10000, got %d", resp.Code)
	}
}

func TestAdminUpdateEndpoints_RejectsNegativeCost(t *testing.T) {
	r := setupPricingHandlerTest(t)

	input := map[string]interface{}{
		"endpoints": []map[string]interface{}{
			{"baseUrl": "https://a.com/v1", "apiKey": "sk-a", "priority": 1, "maxConcurrency": 2, "costPerImageX10000": json.Number("-99")},
		},
	}
	bodyBytes, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/config/endpoints", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminPricingToken(t))
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative CostPerImageX10000 on endpoints route, got %d body=%s", resp.Code, resp.Body.String())
	}
}
