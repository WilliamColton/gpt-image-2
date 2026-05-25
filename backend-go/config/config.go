package config

import (
	"encoding/json"
	"fmt"
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
func SetEndpoints(eps []ApiEndpoint) error {
	return setEndpoints(eps, true)
}

func setEndpoints(eps []ApiEndpoint, persist bool) error {
	cloned := cloneEndpoints(eps)
	sort.SliceStable(cloned, func(i, j int) bool {
		return cloned[i].Priority > cloned[j].Priority
	})
	endpointsMu.Lock()
	ApiEndpoints = cloned
	if App != nil {
		App.ApiEndpoints = cloned
	}
	endpointsMu.Unlock()

	if persist {
		return persistEndpoints(cloned)
	}
	return nil
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
	ApiEndpoints        []ApiEndpoint `json:"apiEndpoints"`
	SalePriceX10000     int64         `json:"salePriceX10000"`
	InviteInviterReward int           `json:"inviteInviterReward"`
	InviteInviteeReward int           `json:"inviteInviteeReward"`
	InviteDefaultQuota  int           `json:"inviteDefaultQuota"`
	InviteEnabled        bool          `json:"inviteEnabled"`
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
		CodexCLI:      true,
		InviteEnabled: true,
	}

	configPath := filepath.Join(rootDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取 config.json 失败: %w", err)
	}
	if err := json.Unmarshal(data, App); err != nil {
		return fmt.Errorf("解析 config.json 失败: %w", err)
	}

	if err := setEndpoints(App.ApiEndpoints, false); err != nil {
		return err
	}

	// Safety check: warn if default weak secrets are still in use
	if App.JWTSecret == "change-me" {
		slog.Warn("JWTSecret 仍为默认值，请立即更改为强随机字符串")
	}
	if App.AdminApikey == "change-me-admin-apikey" {
		slog.Warn("AdminApikey 仍为默认值，请立即更改为强随机字符串")
	}

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

// GetRootDir returns the current root-dir resolver (for test introspection).
func GetRootDir() func() string {
	return getRootDir
}

// SetRootDir replaces the root-dir resolver (for tests).
func SetRootDir(fn func() string) {
	getRootDir = fn
}

// GetSalePriceX10000 returns the runtime global sale price in X10000 units.
func GetSalePriceX10000() int64 {
	return App.SalePriceX10000
}

// SetPricingConfig atomically sets endpoint costs, global sale price, and persists to config.json.
func SetPricingConfig(eps []ApiEndpoint, salePriceX10000 int64) error {
	cloned := cloneEndpoints(eps)
	sort.SliceStable(cloned, func(i, j int) bool {
		return cloned[i].Priority > cloned[j].Priority
	})

	endpointsMu.Lock()
	ApiEndpoints = cloned
	if App != nil {
		App.SalePriceX10000 = salePriceX10000
		App.ApiEndpoints = cloned
	}
	endpointsMu.Unlock()

	return persistPricingConfig(cloned, salePriceX10000)
}

// persistEndpoints writes the current endpoint pool to config.json.
func persistEndpoints(eps []ApiEndpoint) error {
	persistMu.Lock()
	defer persistMu.Unlock()

	configPath := filepath.Join(getRootDir(), "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		data = []byte("{}")
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		raw = map[string]json.RawMessage{}
	}

	epsJSON, err := json.Marshal(eps)
	if err != nil {
		return fmt.Errorf("marshal endpoints: %w", err)
	}
	raw["apiEndpoints"] = epsJSON

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := atomicWriteConfig(configPath, out); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	slog.Info("endpoints persisted to config.json", "count", len(eps))
	return nil
}

// persistPricingConfig writes apiEndpoints and salePriceX10000 to config.json in one atomic write.
func persistPricingConfig(eps []ApiEndpoint, salePriceX10000 int64) error {
	persistMu.Lock()
	defer persistMu.Unlock()

	configPath := filepath.Join(getRootDir(), "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		data = []byte("{}")
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		raw = map[string]json.RawMessage{}
	}

	epsJSON, err := json.Marshal(eps)
	if err != nil {
		return fmt.Errorf("marshal endpoints: %w", err)
	}
	raw["apiEndpoints"] = epsJSON

	saleJSON, err := json.Marshal(salePriceX10000)
	if err != nil {
		return fmt.Errorf("marshal salePriceX10000: %w", err)
	}
	raw["salePriceX10000"] = saleJSON

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := atomicWriteConfig(configPath, out); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	slog.Info("pricing config persisted to config.json", "endpoint_count", len(eps), "salePriceX10000", salePriceX10000)
	return nil
}

// GetInviteInviterReward returns the inviter reward quota, or 0 if not positive.
func GetInviteInviterReward() int {
	if App.InviteInviterReward <= 0 {
		return 0
	}
	return App.InviteInviterReward
}

// GetInviteInviteeReward returns the invitee reward quota, or 0 if not positive.
func GetInviteInviteeReward() int {
	if App.InviteInviteeReward <= 0 {
		return 0
	}
	return App.InviteInviteeReward
}

// GetInviteDefaultQuota returns the default quota for registration without invite code.
func GetInviteDefaultQuota() int {
	return App.InviteDefaultQuota
}

// IsInviteEnabled returns whether the invite system is enabled.
func IsInviteEnabled() bool {
	return App.InviteEnabled
}

// persistInviteConfig writes invite config fields to config.json atomically.
func persistInviteConfig(inviterReward, inviteeReward, defaultQuota int, inviteEnabled bool) error {
	persistMu.Lock()
	defer persistMu.Unlock()

	configPath := filepath.Join(getRootDir(), "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		data = []byte("{}")
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		raw = map[string]json.RawMessage{}
	}

	inviterJSON, err := json.Marshal(inviterReward)
	if err != nil {
		return fmt.Errorf("marshal inviterReward: %w", err)
	}
	raw["inviteInviterReward"] = inviterJSON

	inviteeJSON, err := json.Marshal(inviteeReward)
	if err != nil {
		return fmt.Errorf("marshal inviteeReward: %w", err)
	}
	raw["inviteInviteeReward"] = inviteeJSON

	defaultJSON, err := json.Marshal(defaultQuota)
	if err != nil {
		return fmt.Errorf("marshal defaultQuota: %w", err)
	}
	raw["inviteDefaultQuota"] = defaultJSON

	enabledJSON, err := json.Marshal(inviteEnabled)
	if err != nil {
		return fmt.Errorf("marshal inviteEnabled: %w", err)
	}
	raw["inviteEnabled"] = enabledJSON

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := atomicWriteConfig(configPath, out); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	slog.Info("invite config persisted to config.json", "inviterReward", inviterReward, "inviteeReward", inviteeReward, "defaultQuota", defaultQuota, "inviteEnabled", inviteEnabled)
	return nil
}

// SetInviteConfig updates the runtime invite config and persists to config.json.
func SetInviteConfig(inviterReward, inviteeReward, defaultQuota int, inviteEnabled bool) error {
	App.InviteInviterReward = inviterReward
	App.InviteInviteeReward = inviteeReward
	App.InviteDefaultQuota = defaultQuota
	App.InviteEnabled = inviteEnabled
	return persistInviteConfig(inviterReward, inviteeReward, defaultQuota, inviteEnabled)
}

// atomicWriteConfig writes data to a temp file and renames it, with 0600 permissions.
func atomicWriteConfig(configPath string, data []byte) error {
	tmpPath := configPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmpPath, configPath)
}
