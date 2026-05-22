package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type ApiEndpoint struct {
	BaseURL             string `json:"baseUrl"`
	APIKey              string `json:"apiKey"`
	MaxConcurrency      int    `json:"maxConcurrency"` // 0 = 无限制
	Priority            int    `json:"priority"`
	CostPerImageX10000  int64  `json:"costPerImageX10000"`
}

// ApiEndpoints is the runtime endpoint pool, managed via admin dashboard.
var (
	endpointsMu  sync.RWMutex
	persistMu    sync.Mutex
	ApiEndpoints []ApiEndpoint
)

// GetEndpointPool returns the current endpoint pool for failover.
func GetEndpointPool() []ApiEndpoint {
	endpointsMu.RLock()
	defer endpointsMu.RUnlock()
	return cloneEndpoints(ApiEndpoints)
}

// SetEndpoints replaces the endpoint pool (called from admin API) and persists to config.json.
func SetEndpoints(eps []ApiEndpoint) {
	setEndpoints(eps, true)
}

func setEndpoints(eps []ApiEndpoint, persist bool) {
	cloned := cloneEndpoints(eps)
	sort.SliceStable(cloned, func(i, j int) bool {
		return cloned[i].Priority > cloned[j].Priority
	})
	endpointsMu.Lock()
	ApiEndpoints = cloned
	endpointsMu.Unlock()

	if persist {
		persistEndpoints(cloned)
	}
}

func cloneEndpoints(eps []ApiEndpoint) []ApiEndpoint {
	if len(eps) == 0 {
		return nil
	}
	cloned := make([]ApiEndpoint, len(eps))
	copy(cloned, eps)
	return cloned
}

type Config struct {
	RootDir          string        `json:"-"`
	DataDir          string        `json:"-"`
	UploadDir        string        `json:"-"`
	Port             int           `json:"port"`
	JWTSecret        string        `json:"jwtSecret"`
	AdminApikey      string        `json:"adminApikey"`
	Model            string        `json:"model"`
	APIMode          string        `json:"apiMode"`
	Timeout          int           `json:"timeout"`
	CodexCLI         bool          `json:"codexCli"`
	ApiEndpoints     []ApiEndpoint `json:"apiEndpoints"`
	SalePriceX10000  int64         `json:"salePriceX10000"`
}

var App *Config

func Load() error {
	rootDir := getRootDir()
	App = &Config{
		RootDir:     rootDir,
		DataDir:     filepath.Join(rootDir, "data"),
		UploadDir:   filepath.Join(rootDir, "upload"),
		Port:        3001,
		JWTSecret:   "change-me",
		AdminApikey: "change-me-admin-apikey",
		Model:       "gpt-image-2",
		APIMode:     "images",
		Timeout:     6000,
		CodexCLI:    true,
	}

	data, err := os.ReadFile(filepath.Join(rootDir, "config.json"))
	if err != nil {
		return nil
	}
	_ = json.Unmarshal(data, App)

	setEndpoints(App.ApiEndpoints, false)

	return nil
}

// getRootDir is a variable so tests can override it.
var getRootDir = func() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

// GetSalePriceX10000 returns the runtime global sale price in X10000 units.
func GetSalePriceX10000() int64 {
	return App.SalePriceX10000
}

// SetPricingConfig atomically sets endpoint costs, global sale price, and persists to config.json.
func SetPricingConfig(eps []ApiEndpoint, salePriceX10000 int64) {
	cloned := cloneEndpoints(eps)
	sort.SliceStable(cloned, func(i, j int) bool {
		return cloned[i].Priority > cloned[j].Priority
	})

	endpointsMu.Lock()
	ApiEndpoints = cloned
	App.SalePriceX10000 = salePriceX10000
	endpointsMu.Unlock()

	persistPricingConfig(cloned, salePriceX10000)
}

// persistEndpoints writes the current endpoint pool to config.json.
func persistEndpoints(eps []ApiEndpoint) {
	persistMu.Lock()
	defer persistMu.Unlock()

	configPath := filepath.Join(getRootDir(), "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// config.json might not exist; start from empty object
		data = []byte("{}")
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		raw = map[string]json.RawMessage{}
	}

	epsJSON, err := json.Marshal(eps)
	if err != nil {
		slog.Error("persist endpoints: marshal failed", "error", err)
		return
	}
	raw["apiEndpoints"] = epsJSON

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		slog.Error("persist endpoints: marshal config failed", "error", err)
		return
	}

	if err := os.WriteFile(configPath, out, 0644); err != nil {
		slog.Error("persist endpoints: write failed", "error", err)
		return
	}
	slog.Info("endpoints persisted to config.json", "count", len(eps))
}

// persistPricingConfig writes apiEndpoints and salePriceX10000 to config.json in one atomic write.
func persistPricingConfig(eps []ApiEndpoint, salePriceX10000 int64) {
	persistMu.Lock()
	defer persistMu.Unlock()

	configPath := filepath.Join(getRootDir(), "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// config.json might not exist; start from empty object
		data = []byte("{}")
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		raw = map[string]json.RawMessage{}
	}

	epsJSON, err := json.Marshal(eps)
	if err != nil {
		slog.Error("persist pricing: marshal endpoints failed", "error", err)
		return
	}
	raw["apiEndpoints"] = epsJSON

	saleJSON, err := json.Marshal(salePriceX10000)
	if err != nil {
		slog.Error("persist pricing: marshal salePriceX10000 failed", "error", err)
		return
	}
	raw["salePriceX10000"] = saleJSON

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		slog.Error("persist pricing: marshal config failed", "error", err)
		return
	}

	if err := os.WriteFile(configPath, out, 0644); err != nil {
		slog.Error("persist pricing: write failed", "error", err)
		return
	}
	slog.Info("pricing config persisted to config.json", "endpoint_count", len(eps), "salePriceX10000", salePriceX10000)
}
