package service

import (
	"gpt-image-playground/backend/database"

	_ "golang.org/x/crypto/bcrypt"
)

type User struct {
	ID              string  `json:"id"`
	Label           string  `json:"label"`
	Username        string  `json:"username,omitempty"`
	Role            string  `json:"role"`
	Status          string  `json:"-"`
	Quota           int     `json:"quota"`
	UsedCount       int     `json:"usedCount"`
	PasswordHash    *string `json:"-"`
	InviteCode      *string `json:"inviteCode,omitempty"`
	InviteCodeSetAt *int64  `json:"-"`
}

type AdminUser struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Username  string `json:"username,omitempty"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	Quota     int    `json:"quota"`
	UsedCount int    `json:"usedCount"`
	CreatedAt int64  `json:"createdAt"`
}

type AuthUser struct {
	ID             string `json:"id"`
	Username       string `json:"username,omitempty"`
	Label          string `json:"label"`
	Role           string `json:"role"`
	ImageCount     int    `json:"imageCount"`
	Quota          int    `json:"quota"`
	UsedCount      int    `json:"usedCount"`
	NeedsMigration bool   `json:"needsMigration"`
}

// dbUserToAuthUser converts a database.User to a service.AuthUser.
func dbUserToAuthUser(u *database.User) *AuthUser {
	username := ""
	if u.Username != nil {
		username = *u.Username
	}
	return &AuthUser{
		ID:             u.ID,
		Username:       username,
		Label:          u.Label,
		Role:           u.Role,
		ImageCount:     0,
		Quota:          u.Quota,
		UsedCount:      u.UsedCount,
		NeedsMigration: u.PasswordHash == nil,
	}
}

type RedemptionCode struct {
	ID        string  `json:"id"`
	Code      string  `json:"code"`
	Quota     int     `json:"quota"`
	UsedBy    *string `json:"usedBy,omitempty"`
	UsedAt    *int64  `json:"usedAt,omitempty"`
	CreatedAt int64   `json:"createdAt"`
}

type AppConfig struct {
	CodexCLI      bool   `json:"codexCli"`
	APIMode       string `json:"apiMode"`
	Model         string `json:"model"`
	Timeout       int    `json:"timeout"`
	InviteEnabled bool   `json:"inviteEnabled"`
}

type Image struct {
	ID        string `json:"id"`
	UserID    string `json:"userId,omitempty"`
	FilePath  string `json:"filePath,omitempty"`
	Mime      string `json:"mime"`
	Size      int64  `json:"size"`
	Sha256    string `json:"sha256,omitempty"`
	Source    string `json:"source"`
	CreatedAt int64  `json:"createdAt"`
}

type TaskParams struct {
	Size              string  `json:"size"`
	Quality           string  `json:"quality"`
	OutputFormat      string  `json:"output_format"`
	OutputCompression *int    `json:"output_compression"`
	Moderation        string  `json:"moderation"`
	N                 int     `json:"n"`
}

type TaskRecord struct {
	ID                   string              `json:"id"`
	Prompt               string              `json:"prompt"`
	Params               interface{}         `json:"params"`
	ActualParams         interface{}         `json:"actualParams,omitempty"`
	ActualParamsByImage  interface{}         `json:"actualParamsByImage,omitempty"`
	RevisedPromptByImage interface{}        `json:"revisedPromptByImage,omitempty"`
	InputImageIDs        []string            `json:"inputImageIds"`
	MaskTargetImageID    *string             `json:"maskTargetImageId"`
	MaskImageID          *string             `json:"maskImageId"`
	OutputImages         []string            `json:"outputImages"`
	Status               string              `json:"status"`
	Error                *string             `json:"error"`
	IsFavorite           bool                `json:"isFavorite"`
	CreatedAt            int64               `json:"createdAt"`
	FinishedAt           *int64              `json:"finishedAt"`
	Elapsed              *int64              `json:"elapsed"`
	ApiMode              string              `json:"apiMode,omitempty"`
	CodexCli             bool                `json:"codexCli,omitempty"`
}
