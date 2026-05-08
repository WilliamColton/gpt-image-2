package service

import (
	"fmt"
	"strings"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/util"

	"github.com/golang-jwt/jwt/v5"
)

func SignToken(userID, role, jwtSecret string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func VerifyToken(tokenStr, jwtSecret string) (sub, role string, err error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return "", "", fmt.Errorf("登录状态无效")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("登录状态无效")
	}
	subVal, ok := claims["sub"].(string)
	if !ok {
		return "", "", fmt.Errorf("登录状态无效")
	}
	roleVal, _ := claims["role"].(string)
	if roleVal != "admin" {
		roleVal = "user"
	}
	return subVal, roleVal, nil
}

func FindUserByID(id string) (*User, error) {
	row := database.DB.QueryRow("SELECT id, label, role, status, apikey_cipher, quota, used_count FROM users WHERE id = ?", id)
	u := &User{}
	err := row.Scan(&u.ID, &u.Label, &u.Role, &u.Status, &u.ApikeyCipher, &u.Quota, &u.UsedCount)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func LoginWithApikey(apikey string) (token string, user *AuthUser, err error) {
	normalized := strings.TrimSpace(apikey)
	if normalized == "" {
		return "", nil, fmt.Errorf("请输入 apikey")
	}
	apikeyHash := util.HashApikey(normalized)
	now := time.Now().UnixMilli()

	u := &User{}
	err = database.DB.QueryRow("SELECT id, label, role, status, apikey_cipher, quota, used_count FROM users WHERE apikey_hash = ?", apikeyHash).Scan(&u.ID, &u.Label, &u.Role, &u.Status, &u.ApikeyCipher, &u.Quota, &u.UsedCount)

	if err != nil {
		userID := util.GenerateID()
		encrypted := util.EncryptApikey(normalized, config.App.ApikeyEncryptionSecret)
		label := fmt.Sprintf("user-%s", userID[:8])
		_, insErr := database.DB.Exec(`
			INSERT INTO users (id, label, role, apikey_hash, apikey_cipher, status, created_at, last_login_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, userID, label, "user", apikeyHash, encrypted, "active", now, now)
		if insErr != nil {
			return "", nil, fmt.Errorf("登录失败")
		}
		u, err = FindUserByID(userID)
		if err != nil {
			return "", nil, fmt.Errorf("登录失败")
		}
	} else {
		if u.Status == "disabled" {
			return "", nil, fmt.Errorf("该 apikey 已被禁用")
		}
		database.DB.Exec("UPDATE users SET last_login_at = ? WHERE id = ?", now, u.ID)
		u, _ = FindUserByID(u.ID)
	}

	if u == nil {
		return "", nil, fmt.Errorf("登录失败")
	}

	token, err = SignToken(u.ID, u.Role, config.App.JWTSecret)
	if err != nil {
		return "", nil, fmt.Errorf("登录失败")
	}
	var imageCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM images WHERE user_id = ? AND source = 'generated'", u.ID).Scan(&imageCount)

	return token, &AuthUser{ID: u.ID, Label: u.Label, Role: u.Role, ImageCount: imageCount}, nil
}

func CountGeneratedImages(userID string) int {
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM images WHERE user_id = ? AND source = 'generated'", userID).Scan(&count)
	return count
}

// ListAllUsers returns all users for admin view.
func ListAllUsers() ([]AdminUser, error) {
	rows, err := database.DB.Query("SELECT id, label, role, status, quota, used_count, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Label, &u.Role, &u.Status, &u.Quota, &u.UsedCount, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdateUserQuota adjusts the user's quota by delta and optionally resets used_count.
// delta > 0 adds quota, delta < 0 reduces quota. If resetUsedCount is true, used_count is set to 0.
func UpdateUserQuota(userID string, delta int, resetUsedCount bool) error {
	if resetUsedCount {
		_, err := database.DB.Exec("UPDATE users SET quota = MAX(0, quota + ?), used_count = 0 WHERE id = ?", delta, userID)
		return err
	}
	_, err := database.DB.Exec("UPDATE users SET quota = MAX(0, quota + ?) WHERE id = ?", delta, userID)
	return err
}

// SetUserStatus sets a user's status to "active" or "disabled".
func SetUserStatus(userID, status string) error {
	_, err := database.DB.Exec("UPDATE users SET status = ? WHERE id = ?", status, userID)
	return err
}

// IncrementUsedCount adds count to the user's used_count.
func IncrementUsedCount(userID string, count int) error {
	_, err := database.DB.Exec("UPDATE users SET used_count = used_count + ? WHERE id = ?", count, userID)
	return err
}

// CheckQuota returns nil if the user can generate, or an error message if quota is exceeded.
// quota=0 means unlimited.
func CheckQuota(userID string) error {
	u, err := FindUserByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	if u.Quota > 0 && u.UsedCount >= u.Quota {
		return fmt.Errorf("配额已用完")
	}
	return nil
}
