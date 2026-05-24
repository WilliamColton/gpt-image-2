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
	"golang.org/x/crypto/bcrypt"
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
	username := ""
	if u.Username != nil {
		username = *u.Username
	}
	isUnlimited := u.UnlimitedQuota != 0
	return &User{
		ID:             u.ID,
		Label:          u.Label,
		Username:       username,
		Role:           u.Role,
		Status:         u.Status,
		Quota:          u.Quota,
		UnlimitedQuota: isUnlimited,
		UsedCount:      u.UsedCount,
		PasswordHash:   u.PasswordHash,
	}, nil
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
		var existingUser database.User
		if err := database.DB.Where("id = ?", *rc.UsedBy).First(&existingUser).Error; err != nil {
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
		return token, dbUserToAuthUser(&existingUser), nil
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

	return token, dbUserToAuthUser(newUser), nil
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
		username := ""
		if u.Username != nil {
			username = *u.Username
		}
		result[i] = AdminUser{ID: u.ID, Label: u.Label, Username: username, Role: u.Role, Status: u.Status, Quota: u.Quota, UnlimitedQuota: u.UnlimitedQuota != 0, UsedCount: u.UsedCount, CreatedAt: u.CreatedAt}
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

// SetUserUnlimited toggles a user's unlimited quota flag.
func SetUserUnlimited(userID string, unlimited bool) error {
	v := 0
	if unlimited {
		v = 1
	}
	err := database.DB.Model(&database.User{}).Where("id = ?", userID).Update("unlimited_quota", v).Error
	if err != nil {
		slog.Error("更新用户无限配额失败", "user_id", userID, "unlimited", unlimited, "error", err)
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

// DeleteUsers removes multiple users by IDs.
func DeleteUsers(ids []string) (int64, error) {
	result := database.DB.Where("id IN ?", ids).Delete(&database.User{})
	if result.Error != nil {
		slog.Error("批量删除用户失败", "count", len(ids), "error", result.Error)
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// DeleteCodes removes multiple redemption codes by IDs.
func DeleteCodes(ids []string) (int64, error) {
	result := database.DB.Where("id IN ?", ids).Delete(&database.RedemptionCode{})
	if result.Error != nil {
		slog.Error("批量删除兑换码失败", "count", len(ids), "error", result.Error)
		return 0, result.Error
	}
	return result.RowsAffected, nil
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
func CheckQuota(userID string, count int) error {
	u, err := FindUserByID(userID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	if u.UnlimitedQuota {
		return nil
	}
	pending := CountPendingImages(userID)
	if u.UsedCount+pending+count > u.Quota {
		remaining := u.Quota - u.UsedCount - pending
		if remaining < 0 {
			remaining = 0
		}
		slog.Warn("用户配额不足", "user_id", userID, "quota", u.Quota, "used_count", u.UsedCount, "pending", pending, "requested", count)
		return fmt.Errorf("配额不足，剩余 %d 张（含进行中任务），本次需要 %d 张", remaining, count)
	}
	return nil
}

// ---------------------------------------------------------------------------
// bcrypt helpers
// ---------------------------------------------------------------------------

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败")
	}
	return string(bytes), nil
}

func checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ---------------------------------------------------------------------------
// Password Authentication
// ---------------------------------------------------------------------------

// LoginWithPassword authenticates a user by username and password.
func LoginWithPassword(username, password string) (token string, user *AuthUser, needsMigrate bool, err error) {
	var u database.User
	if err := database.DB.Where("username = ?", username).First(&u).Error; err != nil {
		return "", nil, false, fmt.Errorf("用户名或密码错误")
	}
	if u.PasswordHash == nil {
		return "", nil, false, fmt.Errorf("该账号尚未设置密码，请使用兑换码登录后设置密码")
	}
	if !checkPassword(*u.PasswordHash, password) {
		return "", nil, false, fmt.Errorf("用户名或密码错误")
	}
	if u.Status == "disabled" {
		return "", nil, false, fmt.Errorf("账号已被禁用")
	}
	now := time.Now().UnixMilli()
	database.DB.Model(&database.User{}).Where("id = ?", u.ID).Update("last_login_at", now)

	token, err = SignToken(u.ID, u.Role, config.App.JWTSecret)
	if err != nil {
		slog.Error("签发 JWT 失败", "user_id", u.ID, "error", err)
		return "", nil, false, fmt.Errorf("登录失败")
	}
	return token, dbUserToAuthUser(&u), false, nil
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

// RegisterUser creates a new user with username, password, and optional invite code.
func RegisterUser(username, password, inviteCode string) (token string, user *AuthUser, err error) {
	if len([]rune(username)) < 3 || len([]rune(username)) > 20 {
		return "", nil, fmt.Errorf("用户名须为 3-20 个字符")
	}
	if len(password) < 8 {
		return "", nil, fmt.Errorf("密码至少需要 8 个字符")
	}

	// Check username uniqueness
	var existing database.User
	if result := database.DB.Where("username = ?", username).First(&existing); result.Error == nil {
		return "", nil, fmt.Errorf("用户名已被使用")
	}

	// Hash password
	hashStr, err := hashPassword(password)
	if err != nil {
		return "", nil, err
	}

	// Validate and lookup invite code
	normalizedInviteCode := strings.TrimSpace(inviteCode)
	if !config.IsInviteEnabled() && normalizedInviteCode != "" {
		return "", nil, fmt.Errorf("邀请功能已关闭")
	}
	var inviter *database.User
	if normalizedInviteCode != "" {
		var inviteOwner database.User
		if err := database.DB.Where("invite_code = ?", normalizedInviteCode).First(&inviteOwner).Error; err != nil {
			return "", nil, fmt.Errorf("邀请码无效")
		}
		inviter = &inviteOwner
	}

	// Determine initial quota
	quota := config.GetInviteDefaultQuota()
	if inviter != nil {
		quota += config.GetInviteInviteeReward()
	}

	userID := util.GenerateID()
	now := time.Now().UnixMilli()

	var inviteCodePtr *string
	if normalizedInviteCode != "" {
		inviteCodePtr = &normalizedInviteCode
	}

	newUser := &database.User{
		ID:           userID,
		Label:        userID[:8],
		Username:     &username,
		PasswordHash: &hashStr,
		Role:         "user",
		Status:       "active",
		Quota:        quota,
		UsedCount:    0,
		CreatedAt:    now,
		LastLoginAt:  &now,
		InvitedBy:    inviteCodePtr,
	}
	if err := database.DB.Create(newUser).Error; err != nil {
		slog.Error("创建用户失败", "user_id", userID, "error", err)
		return "", nil, fmt.Errorf("注册失败")
	}

	// Award inviter quota (atomic)
	if inviter != nil {
		reward := config.GetInviteInviterReward()
		if reward > 0 {
			database.DB.Model(&database.User{}).Where("id = ?", inviter.ID).
				Update("quota", gorm.Expr("quota + ?", reward))
		}
	}

	token, err = SignToken(userID, "user", config.App.JWTSecret)
	if err != nil {
		return "", nil, fmt.Errorf("登录失败")
	}
	return token, dbUserToAuthUser(newUser), nil
}

// ---------------------------------------------------------------------------
// Migration
// ---------------------------------------------------------------------------

// MigrateUser sets a username and password for an existing user (legacy account).
func MigrateUser(userID, username, password string) (*AuthUser, error) {
	if len([]rune(username)) < 3 || len([]rune(username)) > 20 {
		return nil, fmt.Errorf("用户名须为 3-20 个字符")
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("密码至少需要 8 个字符")
	}

	// Username uniqueness check excluding self
	var existing database.User
	if err := database.DB.Where("username = ? AND id != ?", username, userID).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("用户名已被使用")
	}

	hashStr, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	result := database.DB.Model(&database.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{"username": username, "password_hash": hashStr})
	if result.Error != nil {
		slog.Error("迁移用户失败", "user_id", userID, "error", result.Error)
		return nil, fmt.Errorf("迁移失败")
	}

	var u database.User
	if err := database.DB.Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	return dbUserToAuthUser(&u), nil
}

// ---------------------------------------------------------------------------
// Username Change
// ---------------------------------------------------------------------------

// ChangeUsername changes a user's username.
func ChangeUsername(userID, newUsername string) error {
	if len([]rune(newUsername)) < 3 || len([]rune(newUsername)) > 20 {
		return fmt.Errorf("用户名须为 3-20 个字符")
	}

	var existing database.User
	if err := database.DB.Where("username = ? AND id != ?", newUsername, userID).First(&existing).Error; err == nil {
		return fmt.Errorf("用户名已被使用")
	}

	if err := database.DB.Model(&database.User{}).Where("id = ?", userID).
		Update("username", newUsername).Error; err != nil {
		slog.Error("修改用户名失败", "user_id", userID, "error", err)
		return fmt.Errorf("修改用户名失败")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Password Change
// ---------------------------------------------------------------------------

// ChangePassword changes a user's password after verifying the old password.
func ChangePassword(userID, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("密码至少需要 8 个字符")
	}

	var u database.User
	if err := database.DB.Where("id = ?", userID).First(&u).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}
	if u.PasswordHash == nil {
		return fmt.Errorf("该账号尚未设置密码")
	}
	if !checkPassword(*u.PasswordHash, oldPassword) {
		return fmt.Errorf("旧密码不正确")
	}

	hashStr, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := database.DB.Model(&database.User{}).Where("id = ?", userID).
		Update("password_hash", hashStr).Error; err != nil {
		slog.Error("修改密码失败", "user_id", userID, "error", err)
		return fmt.Errorf("修改密码失败")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Invite Code Management
// ---------------------------------------------------------------------------

// InviteRow represents an invite code usage row.
type InviteRow struct {
	Username   string `json:"username"`
	InviteCode string `json:"inviteCode"`
	UsageCount int    `json:"usageCount"`
}

// SetInviteCode assigns an invite code to a user. Duplicate codes are rejected
// by the database unique constraint.
func SetInviteCode(userID, code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("邀请码不能为空")
	}

	result := database.DB.Model(&database.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{"invite_code": code, "invite_code_set_at": time.Now().UnixMilli()})
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "UNIQUE constraint") {
			return fmt.Errorf("该邀请码已被使用")
		}
		slog.Error("设置邀请码失败", "user_id", userID, "error", result.Error)
		return fmt.Errorf("设置邀请码失败")
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}
	return nil
}

// GetInviteCode returns a user's current invite code and when it was set.
func GetInviteCode(userID string) (*string, *int64, error) {
	var u database.User
	if err := database.DB.Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, nil, err
	}
	return u.InviteCode, u.InviteCodeSetAt, nil
}

// ListInvites returns all users with invite codes and their usage counts.
func ListInvites() ([]InviteRow, error) {
	var owners []database.User
	if err := database.DB.Where("invite_code IS NOT NULL").Find(&owners).Error; err != nil {
		slog.Error("查询邀请码列表失败", "error", err)
		return nil, err
	}

	rows := make([]InviteRow, 0, len(owners))
	for _, owner := range owners {
		var count int64
		database.DB.Model(&database.User{}).Where("invited_by = ?", *owner.InviteCode).Count(&count)
		username := ""
		if owner.Username != nil {
			username = *owner.Username
		}
		rows = append(rows, InviteRow{
			Username:   username,
			InviteCode: *owner.InviteCode,
			UsageCount: int(count),
		})
	}
	return rows, nil
}

// InvitedUserRow represents a user who registered with the current user's invite code.
type InvitedUserRow struct {
	Username  string `json:"username"`
	Label     string `json:"label"`
	CreatedAt int64  `json:"createdAt"`
}

// GetInvitedUsers returns users who registered with a given user's invite code.
func GetInvitedUsers(userID string) ([]InvitedUserRow, error) {
	var u database.User
	if err := database.DB.Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, err
	}
	if u.InviteCode == nil || *u.InviteCode == "" {
		return nil, fmt.Errorf("你还没有设置邀请码")
	}

	var invitees []database.User
	if err := database.DB.Where("invited_by = ?", *u.InviteCode).Order("created_at DESC").Find(&invitees).Error; err != nil {
		return nil, err
	}

	rows := make([]InvitedUserRow, 0, len(invitees))
	for _, invitee := range invitees {
		username := ""
		if invitee.Username != nil {
			username = *invitee.Username
		}
		rows = append(rows, InvitedUserRow{
			Username:  username,
			Label:     invitee.Label,
			CreatedAt: invitee.CreatedAt,
		})
	}
	return rows, nil
}

// ---------------------------------------------------------------------------
// Admin Operations
// ---------------------------------------------------------------------------

// AdminResetPassword sets a user's password hash without requiring the old password.
func AdminResetPassword(userID, password string) error {
	if len(password) < 8 {
		return fmt.Errorf("密码至少需要 8 个字符")
	}

	hashStr, err := hashPassword(password)
	if err != nil {
		return err
	}

	result := database.DB.Model(&database.User{}).Where("id = ?", userID).
		Update("password_hash", hashStr)
	if result.Error != nil {
		slog.Error("管理员重置密码失败", "user_id", userID, "error", result.Error)
		return fmt.Errorf("重置密码失败")
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}

	slog.Info("管理员重置密码", "user_id", userID)
	return nil
}

