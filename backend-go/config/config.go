package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Defaults struct {
	BaseURL  string
	CodexCLI bool
	APIMode  string
	Model    string
	Timeout  int
}

type Config struct {
	RootDir                string
	DataDir                string
	UploadDir              string
	Port                   int
	JWTSecret              string
	AdminApikey            string
	ApikeyEncryptionSecret string
	CORSOrigin             string
	OpenAIConfigured       bool
	Defaults               Defaults
}

var App *Config

func readString(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func readNumber(name string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func readBool(name string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}
	return fallback
}

func Load() error {
	_ = godotenv.Load(filepath.Join(getRootDir(), ".env"))

	rootDir := getRootDir()
	App = &Config{
		RootDir:                rootDir,
		DataDir:                filepath.Join(rootDir, "data"),
		UploadDir:              filepath.Join(rootDir, "upload"),
		Port:                   readNumber("PORT", 3001),
		JWTSecret:              readString("JWT_SECRET", "change-me"),
		AdminApikey:            readString("ADMIN_APIKEY", "change-me-admin-apikey"),
		ApikeyEncryptionSecret: readString("APIKEY_ENCRYPTION_SECRET", "change-me-32-bytes-minimum-secret"),
		CORSOrigin:             readString("CORS_ORIGIN", "http://localhost:5173"),
		OpenAIConfigured:       strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) != "",
		Defaults: Defaults{
			BaseURL:  readString("DEFAULT_BASE_URL", "https://api.openai.com/v1"),
			CodexCLI: readBool("DEFAULT_CODEX_CLI", false),
			APIMode:  readString("DEFAULT_API_MODE", "images"),
			Model:    readString("DEFAULT_MODEL", "gpt-image-2"),
			Timeout:  readNumber("DEFAULT_TIMEOUT", 6000),
		},
	}
	return nil
}

func getRootDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}
