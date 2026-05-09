package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ApiEndpoint struct {
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey"`
}

type Config struct {
	RootDir     string `json:"-"`
	DataDir     string `json:"-"`
	UploadDir   string `json:"-"`
	Port        int    `json:"port"`
	JWTSecret   string `json:"jwtSecret"`
	AdminApikey string `json:"adminApikey"`
	Model       string `json:"model"`
	APIMode                string        `json:"apiMode"`
	Timeout                int           `json:"timeout"`
	CodexCLI               bool          `json:"codexCli"`
	ApiEndpoints           []ApiEndpoint `json:"apiEndpoints"`
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
		APIMode:                "images",
		Timeout:                6000,
		CodexCLI:               true,
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
func (c *Config) GetEndpointPool() []ApiEndpoint {
	return c.ApiEndpoints
}
