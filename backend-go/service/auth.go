package service

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/util"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
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
		slog.Warn("JWT 验证失败", "error", err)
		return "", "", fmt.Errorf("登录状态无效")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		slog.Warn("JWT claims 无效")
		return "", "", fmt.Errorf("登录状态无效")
	}
	subVal, ok := claims["sub"].(string)
	if !ok {
		slog.Warn("JWT 缺少 sub 字段")
		return "", "", fmt.Errorf("登录状态无效")
	}
	roleVal, _ := claims["role"].(string)
	if roleVal != "admin" {
		roleVal = "user"
	}
	return subVal, roleVal, nil
}

func FindUserByID(id string) (*User, error) {
	var u database.User
	if err := database.DB.Where("id = ?", id).First(&u).Error; err != nil {
		slog.Error("查询用户失败", "user_id", id, "error", err)
		return nil, err
	}
	return &User{ID: u.ID, Label: u.Label, Role: u.Role, Status: u.Status, Quota: u.Quota, UsedCount: u.UsedCount}, nil
}

// RedeemCode validates a redemption code exists and is unused.
func RedeemCode(code string) (*RedemptionCode, error) {
	normalized := strings.TrimSpace(strings.ToUpper(code))
	if normalized == "" {
		return nil, fmt.Errorf("请输入兑换码")
	}
	var rc database.RedemptionCode
	err := database.DB.Where("code = ?", normalized).First(&rc).Error
	if err != nil {
		slog.Warn("兑换码查询失败", "code", normalized, "error", err)
		return nil, fmt.Errorf("兑换码无效")
	}
	if rc.UsedBy != nil {
		return nil, fmt.Errorf("该兑换码已被使用")
	}
	return &RedemptionCode{ID: rc.ID, Code: rc.Code, Quota: rc.Quota, UsedBy: rc.UsedBy, UsedAt: rc.UsedAt, CreatedAt: rc.CreatedAt}, nil
}

// LoginWithCode creates a new user via redemption code or logs in an existing user if code was already used.
func LoginWithCode(code string) (token string, user *AuthUser, err error) {
	normalized := strings.TrimSpace(strings.ToUpper(code))
	if normalized == "" {
		return "", nil, fmt.Errorf("请输入兑换码")
	}
	var rc database.RedemptionCode
	if err := database.DB.Where("code = ?", normalized).First(&rc).Error; err != nil {
		slog.Warn("兑换码查询失败", "code", normalized, "error", err)
		return "", nil, fmt.Errorf("兑换码无效")
	}

	now := time.Now().UnixMilli()

	// If code already used — log in the existing user
	if rc.UsedBy != nil {
		existingUser, err := FindUserByID(*rc.UsedBy)
		if err != nil {
			return "", nil, fmt.Errorf("用户不存在")
		}
		if existingUser.Status == "disabled" {
			return "", nil, fmt.Errorf("账号已被禁用")
		}
		// Update last login time
		database.DB.Model(&database.User{}).Where("id = ?", existingUser.ID).Update("last_login_at", now)
		token, err = SignToken(existingUser.ID, existingUser.Role, config.App.JWTSecret)
		if err != nil {
			slog.Error("签发 JWT 失败", "user_id", existingUser.ID, "error", err)
			return "", nil, fmt.Errorf("登录失败")
		}
		var imageCount int64
		database.DB.Model(&database.Image{}).Where("user_id = ? AND source = ?", existingUser.ID, "generated").Count(&imageCount)
		return token, &AuthUser{ID: existingUser.ID, Label: existingUser.Label, Role: existingUser.Role, ImageCount: int(imageCount), Quota: existingUser.Quota, UsedCount: existingUser.UsedCount}, nil
	}

	// Code unused — create new user
	userID := util.GenerateID()
	label := normalized

	newUser := &database.User{
		ID:          userID,
		Label:       label,
		Role:        "user",
		Status:      "active",
		Quota:       rc.Quota,
		UsedCount:   0,
		CreatedAt:   now,
		LastLoginAt: &now,
	}
	if err := database.DB.Create(newUser).Error; err != nil {
		slog.Error("创建用户失败", "user_id", userID, "code", code, "error", err)
		return "", nil, fmt.Errorf("注册失败")
	}

	// Mark code as used
	result := database.DB.Model(&database.RedemptionCode{}).
		Where("id = ? AND used_by IS NULL", rc.ID).
		Updates(map[string]interface{}{"used_by": userID, "used_at": now})
	if result.Error != nil {
		slog.Error("标记兑换码失败", "code_id", rc.ID, "user_id", userID, "error", result.Error)
	} else if result.RowsAffected == 0 {
		slog.Warn("兑换码已被并发使用", "code_id", rc.ID)
	}

	token, err = SignToken(userID, "user", config.App.JWTSecret)
	if err != nil {
		slog.Error("签发 JWT 失败", "user_id", userID, "error", err)
		return "", nil, fmt.Errorf("登录失败")
	}

	var imageCount int64
	database.DB.Model(&database.Image{}).Where("user_id = ? AND source = ?", userID, "generated").Count(&imageCount)

	return token, &AuthUser{ID: userID, Label: label, Role: "user", ImageCount: int(imageCount), Quota: rc.Quota, UsedCount: 0}, nil
}

// RedeemForUser adds quota to an existing user via redemption code.
func RedeemForUser(userID, code string) error {
	rc, err := RedeemCode(code)
	if err != nil {
		return err
	}

	now := time.Now().UnixMilli()

	// Atomic: mark code as used only if still unused
	result := database.DB.Model(&database.RedemptionCode{}).
		Where("id = ? AND used_by IS NULL", rc.ID).
		Updates(map[string]interface{}{"used_by": userID, "used_at": now})
	if result.Error != nil {
		slog.Error("标记兑换码失败", "code_id", rc.ID, "user_id", userID, "error", result.Error)
		return fmt.Errorf("兑换失败")
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("该兑换码已被使用")
	}

	// Add quota to user
	if err := database.DB.Model(&database.User{}).Where("id = ?", userID).Update("quota", gorm.Expr("quota + ?", rc.Quota)).Error; err != nil {
		slog.Error("更新用户配额失败", "user_id", userID, "quota_delta", rc.Quota, "error", err)
		return fmt.Errorf("兑换失败")
	}

	return nil
}

// CreateRedemptionCode generates a new redemption code with the given quota.
func CreateRedemptionCode(quota int) (*RedemptionCode, error) {
	if quota <= 0 {
		return nil, fmt.Errorf("配额必须大于 0")
	}
	id := util.GenerateID()
	code := generateCode(20)
	now := time.Now().UnixMilli()
	rc := &database.RedemptionCode{
		ID:        id,
		Code:      code,
		Quota:     quota,
		CreatedAt: now,
	}
	if err := database.DB.Create(rc).Error; err != nil {
		slog.Error("创建兑换码失败", "quota", quota, "error", err)
		return nil, fmt.Errorf("创建兑换码失败")
	}
	return &RedemptionCode{ID: rc.ID, Code: rc.Code, Quota: rc.Quota, CreatedAt: rc.CreatedAt}, nil
}

// ListRedemptionCodes returns all redemption codes.
func ListRedemptionCodes() ([]RedemptionCode, error) {
	var codes []database.RedemptionCode
	if err := database.DB.Order("created_at DESC").Find(&codes).Error; err != nil {
		slog.Error("查询兑换码列表失败", "error", err)
		return nil, err
	}
	result := make([]RedemptionCode, len(codes))
	for i, rc := range codes {
		result[i] = RedemptionCode{ID: rc.ID, Code: rc.Code, Quota: rc.Quota, UsedBy: rc.UsedBy, UsedAt: rc.UsedAt, CreatedAt: rc.CreatedAt}
	}
	return result, nil
}

func generateCode(length int) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[n.Int64()]
	}
	return string(b)
}

func CountGeneratedImages(userID string) int {
	var count int64
	database.DB.Model(&database.Image{}).Where("user_id = ? AND source = ?", userID, "generated").Count(&count)
	return int(count)
}

// ListAllUsers returns all users for admin view.
func ListAllUsers() ([]AdminUser, error) {
	var users []database.User
	if err := database.DB.Order("created_at DESC").Find(&users).Error; err != nil {
		slog.Error("查询用户列表失败", "error", err)
		return nil, err
	}
	result := make([]AdminUser, len(users))
	for i, u := range users {
		result[i] = AdminUser{ID: u.ID, Label: u.Label, Role: u.Role, Status: u.Status, Quota: u.Quota, UsedCount: u.UsedCount, CreatedAt: u.CreatedAt}
	}
	return result, nil
}

// UpdateUserQuota adjusts the user's quota by delta and optionally resets used_count.
func UpdateUserQuota(userID string, delta int, resetUsedCount bool) error {
	if resetUsedCount {
		err := database.DB.Model(&database.User{}).Where("id = ?", userID).
			Updates(map[string]interface{}{"quota": gorm.Expr("MAX(0, quota + ?)", delta), "used_count": 0}).Error
		if err != nil {
			slog.Error("更新用户配额失败", "user_id", userID, "delta", delta, "reset", true, "error", err)
		}
		return err
	}
	err := database.DB.Model(&database.User{}).Where("id = ?", userID).
		Update("quota", gorm.Expr("MAX(0, quota + ?)", delta)).Error
	if err != nil {
		slog.Error("更新用户配额失败", "user_id", userID, "delta", delta, "error", err)
	}
	return err
}

// SetUserQuotaAbs sets the user's quota to an absolute value.
func SetUserQuotaAbs(userID string, quota int) error {
	err := database.DB.Model(&database.User{}).Where("id = ?", userID).Update("quota", quota).Error
	if err != nil {
		slog.Error("设置用户配额失败", "user_id", userID, "quota", quota, "error", err)
	}
	return err
}

// SetUserStatus sets a user's status to "active" or "disabled".
func SetUserStatus(userID, status string) error {
	err := database.DB.Model(&database.User{}).Where("id = ?", userID).Update("status", status).Error
	if err != nil {
		slog.Error("更新用户状态失败", "user_id", userID, "status", status, "error", err)
	}
	return err
}

// DeleteUser removes a user from the database.
func DeleteUser(userID string) error {
	result := database.DB.Where("id = ?", userID).Delete(&database.User{})
	if result.Error != nil {
		slog.Error("删除用户失败", "user_id", userID, "error", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}
	return nil
}

// IncrementUsedCount adds count to the user's used_count.
func IncrementUsedCount(userID string, count int) error {
	err := database.DB.Model(&database.User{}).Where("id = ?", userID).
		Update("used_count", gorm.Expr("used_count + ?", count)).Error
	if err != nil {
		slog.Error("更新用户使用计数失败", "user_id", userID, "count", count, "error", err)
	}
	return err
}

// CheckQuota returns nil if the user can generate count images, or an error if quota is exceeded.
// quota=0 means unlimited.
func CheckQuota(userID string, count int) error {
	u, err := FindUserByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	if u.Quota > 0 && u.UsedCount+count > u.Quota {
		remaining := u.Quota - u.UsedCount
		if remaining < 0 {
			remaining = 0
		}
		slog.Warn("用户配额不足", "user_id", userID, "quota", u.Quota, "used_count", u.UsedCount, "requested", count)
		return fmt.Errorf("配额不足，剩余 %d 张，本次需要 %d 张", remaining, count)
	}
	return nil
}

