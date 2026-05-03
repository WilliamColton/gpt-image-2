package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/util"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() error {
	dbPath := filepath.Join(config.App.DataDir, "app.sqlite")
	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	DB.SetMaxOpenConns(1)

	if err := initSchema(); err != nil {
		return err
	}
	if err := initAdmin(); err != nil {
		return err
	}
	return nil
}

func initSchema() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			label TEXT NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('admin', 'user')),
			apikey_hash TEXT NOT NULL UNIQUE,
			apikey_cipher TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),
			created_at INTEGER NOT NULL,
			last_login_at INTEGER
		);

		CREATE TABLE IF NOT EXISTS images (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			file_path TEXT NOT NULL,
			mime TEXT NOT NULL,
			size INTEGER NOT NULL,
			sha256 TEXT NOT NULL,
			source TEXT NOT NULL CHECK (source IN ('upload', 'generated', 'mask')),
			created_at INTEGER NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			prompt TEXT NOT NULL,
			params_json TEXT NOT NULL,
			actual_params_json TEXT,
			actual_params_by_image_json TEXT,
			revised_prompt_by_image_json TEXT,
			input_image_ids_json TEXT NOT NULL,
			mask_target_image_id TEXT,
			mask_image_id TEXT,
			output_image_ids_json TEXT NOT NULL,
			status TEXT NOT NULL CHECK (status IN ('running', 'done', 'error')),
			error TEXT,
			is_favorite INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			finished_at INTEGER,
			elapsed INTEGER,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);
	`)
	return err
}

func initAdmin() error {
	adminHash := util.HashApikey(config.App.AdminApikey)
	var existingID string
	err := DB.QueryRow("SELECT id FROM users WHERE role = ? LIMIT 1", "admin").Scan(&existingID)
	if err == nil {
		return nil
	}
	cipher := util.EncryptApikey(config.App.AdminApikey, config.App.ApikeyEncryptionSecret)
	_, err = DB.Exec(`
		INSERT INTO users (id, label, role, apikey_hash, apikey_cipher, status, created_at, last_login_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NULL)
	`, util.GenerateID(), "admin", "admin", adminHash, cipher, "active", time.Now().UnixMilli())
	return err
}
