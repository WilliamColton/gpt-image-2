package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Defaults struct {
	CodexCLI bool   `json:"codexCli"`
	APIMode  string `json:"apiMode"`
	Model    string `json:"model"`
	Timeout  int    `json:"timeout"`
}

type ApiEndpoint struct {
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey"`
}

type Config struct {
	RootDir                string        `json:"-"`
	DataDir                string        `json:"-"`
	UploadDir              string        `json:"-"`
	Port                   int           `json:"port"`
	JWTSecret              string        `json:"jwtSecret"`
	AdminApikey            string        `json:"adminApikey"`
	ApikeyEncryptionSecret string        `json:"apikeyEncryptionSecret"`
	Defaults               Defaults      `json:"defaults"`
	ApiEndpoints           []ApiEndpoint `json:"apiEndpoints"`
}

var App *Config

func Load() error {
	rootDir := getRootDir()
	App = &Config{
		RootDir:                rootDir,
		DataDir:                filepath.Join(rootDir, "data"),
		UploadDir:              filepath.Join(rootDir, "upload"),
		Port:                   3001,
		JWTSecret:              "change-me",
		AdminApikey:            "change-me-admin-apikey",
		ApikeyEncryptionSecret: "change-me-32-bytes-minimum-secret",
		Defaults: Defaults{
			CodexCLI: true,
			APIMode:  "images",
			Model:    "gpt-image-2",
			Timeout:  6000,
		},
	}

	data, err := os.ReadFile(filepath.Join(rootDir, "config.json"))
	if err != nil {
		return nil
	}
	_ = json.Unmarshal(data, App)
	return nil
}

func getRootDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

// GetEndpointPool returns the list of API endpoints for failover.
// apiEndpoints must have at least one entry.
func (c *Config) GetEndpointPool() []ApiEndpoint {
	return c.ApiEndpoints
}
