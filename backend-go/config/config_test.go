package config

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGetEndpointPool_MultiEndpoint(t *testing.T) {
	setEndpoints([]ApiEndpoint{
		{BaseURL: "https://a.com/v1", APIKey: "sk-a"},
		{BaseURL: "https://b.com/v1", APIKey: "sk-b"},
	}, false)
	pool := GetEndpointPool()
	if len(pool) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(pool))
	}
	if pool[0].BaseURL != "https://a.com/v1" || pool[0].APIKey != "sk-a" {
		t.Errorf("endpoint 0 mismatch: %+v", pool[0])
	}
	if pool[1].BaseURL != "https://b.com/v1" || pool[1].APIKey != "sk-b" {
		t.Errorf("endpoint 1 mismatch: %+v", pool[1])
	}
}

func TestGetEndpointPool_SingleEndpoint(t *testing.T) {
	setEndpoints([]ApiEndpoint{
		{BaseURL: "https://custom.com/v1", APIKey: "sk-x"},
	}, false)
	pool := GetEndpointPool()
	if len(pool) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(pool))
	}
	if pool[0].BaseURL != "https://custom.com/v1" || pool[0].APIKey != "sk-x" {
		t.Errorf("endpoint mismatch: %+v", pool[0])
	}
}

func TestGetEndpointPool_Empty(t *testing.T) {
	setEndpoints(nil, false)
	pool := GetEndpointPool()
	if len(pool) != 0 {
		t.Fatalf("expected 0 endpoints, got %d", len(pool))
	}
}

func TestSetEndpoints(t *testing.T) {
	setEndpoints(nil, false)
	setEndpoints([]ApiEndpoint{
		{BaseURL: "https://new.com/v1", APIKey: "sk-new"},
	}, false)
	pool := GetEndpointPool()
	if len(pool) != 1 || pool[0].BaseURL != "https://new.com/v1" {
		t.Errorf("setEndpoints failed: %+v", pool)
	}
}

func TestGetEndpointPool_ReturnsCopy(t *testing.T) {
	setEndpoints([]ApiEndpoint{{BaseURL: "https://copy.com/v1", APIKey: "sk-copy"}}, false)
	pool := GetEndpointPool()
	pool[0].BaseURL = "https://mutated.com/v1"

	fresh := GetEndpointPool()
	if fresh[0].BaseURL != "https://copy.com/v1" {
		t.Errorf("GetEndpointPool should return a copy, got %+v", fresh)
	}
}

func TestGetEndpointPool_SortsByPriorityDescending(t *testing.T) {
	setEndpoints([]ApiEndpoint{
		{BaseURL: "https://default.com/v1", APIKey: "sk-default"},
		{BaseURL: "https://high.com/v1", APIKey: "sk-high", Priority: 100},
		{BaseURL: "https://low.com/v1", APIKey: "sk-low", Priority: 10},
	}, false)

	pool := GetEndpointPool()
	if len(pool) != 3 {
		t.Fatalf("expected 3 endpoints, got %d", len(pool))
	}
	if pool[0].BaseURL != "https://high.com/v1" || pool[1].BaseURL != "https://low.com/v1" || pool[2].BaseURL != "https://default.com/v1" {
		t.Fatalf("endpoints not sorted by priority: %+v", pool)
	}
}

func TestGetEndpointPool_PreservesOrderForEqualPriority(t *testing.T) {
	setEndpoints([]ApiEndpoint{
		{BaseURL: "https://a.com/v1", APIKey: "sk-a", Priority: 10},
		{BaseURL: "https://b.com/v1", APIKey: "sk-b", Priority: 10},
		{BaseURL: "https://c.com/v1", APIKey: "sk-c", Priority: 10},
	}, false)

	pool := GetEndpointPool()
	if pool[0].BaseURL != "https://a.com/v1" || pool[1].BaseURL != "https://b.com/v1" || pool[2].BaseURL != "https://c.com/v1" {
		t.Fatalf("equal priority order should be stable: %+v", pool)
	}
}

// --- Pricing tests ---

func TestMissingPricingFieldsDefaultToZero(t *testing.T) {
	raw := `{"apiEndpoints": [{"baseUrl": "https://a.com/v1", "apiKey": "sk-a"}]}`
	var cfg Config
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if cfg.SalePriceX10000 != 0 {
		t.Errorf("SalePriceX10000 should default to 0, got %d", cfg.SalePriceX10000)
	}
	if len(cfg.ApiEndpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(cfg.ApiEndpoints))
	}
	if cfg.ApiEndpoints[0].CostPerImageX10000 != 0 {
		t.Errorf("CostPerImageX10000 should default to 0, got %d", cfg.ApiEndpoints[0].CostPerImageX10000)
	}
}

func TestEndpointCostJSONRoundTrip(t *testing.T) {
	ep := ApiEndpoint{
		BaseURL:             "https://x.com/v1",
		APIKey:              "sk-x",
		MaxConcurrency:      3,
		Priority:            1,
		CostPerImageX10000:  12345,
	}
	data, err := json.Marshal(ep)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var parsed ApiEndpoint
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if parsed.CostPerImageX10000 != 12345 {
		t.Errorf("CostPerImageX10000 round-trip mismatch: got %d, want 12345", parsed.CostPerImageX10000)
	}
	if parsed.MaxConcurrency != 3 || parsed.Priority != 1 {
		t.Errorf("existing fields changed: %+v", parsed)
	}
}

func TestGetSalePriceX10000(t *testing.T) {
	// Save original App state
	prevApp := App
	t.Cleanup(func() { App = prevApp })

	App = &Config{SalePriceX10000: 50000}
	got := GetSalePriceX10000()
	if got != 50000 {
		t.Errorf("GetSalePriceX10000() = %d, want 50000", got)
	}
}

func TestSetPricingConfigSortsByPriority(t *testing.T) {
	// Save original App state
	prevApp := App
	t.Cleanup(func() { App = prevApp })

	// Use temp dir to avoid touching real config.json
	dir := t.TempDir()
	origGetRootDir := getRootDir
	t.Cleanup(func() { getRootDir = origGetRootDir })
	getRootDir = func() string { return dir }

	App = &Config{SalePriceX10000: 0}
	eps := []ApiEndpoint{
		{BaseURL: "https://low.com/v1", APIKey: "sk-low", Priority: 10, CostPerImageX10000: 1000},
		{BaseURL: "https://high.com/v1", APIKey: "sk-high", Priority: 100, CostPerImageX10000: 2000},
	}
	SetPricingConfig(eps, 55555)

	if App.SalePriceX10000 != 55555 {
		t.Errorf("SalePriceX10000 = %d, want 55555", App.SalePriceX10000)
	}

	pool := GetEndpointPool()
	if pool[0].BaseURL != "https://high.com/v1" {
		t.Errorf("expected high priority first, got %s", pool[0].BaseURL)
	}
	if pool[0].CostPerImageX10000 != 2000 {
		t.Errorf("high endpoint cost = %d, want 2000", pool[0].CostPerImageX10000)
	}
	if pool[1].CostPerImageX10000 != 1000 {
		t.Errorf("low endpoint cost = %d, want 1000", pool[1].CostPerImageX10000)
	}
}

func TestSetPricingConfigPersistsToFile(t *testing.T) {
	// Use a temp dir with a config.json
	dir := t.TempDir()
	configPath := dir + "/config.json"
	// Create initial config.json
	initial := `{"port": 3001, "jwtSecret": "test", "apiEndpoints": []}`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatalf("write initial config.json: %v", err)
	}

	// Save and override getRootDir
	origGetRootDir := getRootDir
	t.Cleanup(func() { getRootDir = origGetRootDir })
	getRootDir = func() string { return dir }

	// Save original App state
	prevApp := App
	t.Cleanup(func() { App = prevApp })

	App = &Config{SalePriceX10000: 0}
	eps := []ApiEndpoint{
		{BaseURL: "https://e.com/v1", APIKey: "sk-e", Priority: 1, CostPerImageX10000: 9999},
	}
	SetPricingConfig(eps, 77777)

	// Read back the config.json
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config.json: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal config.json: %v", err)
	}

	// Verify salePriceX10000 is in config
	saleStr, ok := raw["salePriceX10000"]
	if !ok {
		t.Fatal("config.json missing salePriceX10000")
	}
	var salePrice int64
	if err := json.Unmarshal(saleStr, &salePrice); err != nil {
		t.Fatalf("unmarshal salePriceX10000: %v", err)
	}
	if salePrice != 77777 {
		t.Errorf("salePriceX10000 in file = %d, want 77777", salePrice)
	}

	// Verify endpoints include cost
	epsRaw, ok := raw["apiEndpoints"]
	if !ok {
		t.Fatal("config.json missing apiEndpoints")
	}
	var persistedEps []ApiEndpoint
	if err := json.Unmarshal(epsRaw, &persistedEps); err != nil {
		t.Fatalf("unmarshal persisted endpoints: %v", err)
	}
	if len(persistedEps) != 1 || persistedEps[0].CostPerImageX10000 != 9999 {
		t.Errorf("persisted endpoint cost mismatch: %+v", persistedEps)
	}
}

func TestSetEndpointsPreservesCostWhenPresent(t *testing.T) {
	prevApp := App
	t.Cleanup(func() { App = prevApp })

	// Use temp dir to avoid touching real config.json
	dir := t.TempDir()
	origGetRootDir := getRootDir
	t.Cleanup(func() { getRootDir = origGetRootDir })
	getRootDir = func() string { return dir }

	App = &Config{SalePriceX10000: 0}
	// Set initial pricing that includes costs
	SetPricingConfig([]ApiEndpoint{
		{BaseURL: "https://keep.com/v1", APIKey: "sk-keep", Priority: 1, CostPerImageX10000: 4321},
	}, 10000)

	// Now call SetEndpoints — it should preserve the cost field
	SetEndpoints([]ApiEndpoint{
		{BaseURL: "https://keep.com/v1", APIKey: "sk-keep", Priority: 1, CostPerImageX10000: 4321},
	})

	pool := GetEndpointPool()
	if pool[0].CostPerImageX10000 != 4321 {
		t.Errorf("SetEndpoints should preserve CostPerImageX10000, got %d", pool[0].CostPerImageX10000)
	}
}
