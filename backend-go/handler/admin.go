package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

func AdminLogin(c *gin.Context) {
	var body struct {
		Apikey string `json:"apikey"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Apikey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入管理员密钥"})
		return
	}
	if body.Apikey != config.App.AdminApikey {
		slog.Warn("管理员密钥错误")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "管理员密钥错误"})
		return
	}
	token, err := service.SignToken("admin", "admin", config.App.JWTSecret)
	if err != nil {
		slog.Error("管理员 JWT 签发失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func AdminListUsers(c *gin.Context) {
	users, err := service.ListAllUsers()
	if err != nil {
		slog.Error("获取用户列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func AdminUpdateQuota(c *gin.Context) {
	userID := c.Param("id")
	var body struct {
		Mode           string `json:"mode"`
		Delta          int    `json:"delta"`
		ResetUsedCount bool   `json:"resetUsedCount"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	if body.Mode == "set" {
		if err := service.SetUserQuotaAbs(userID, body.Delta); err != nil {
			slog.Error("设置配额失败", "user_id", userID, "quota", body.Delta, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "设置配额失败"})
			return
		}
	} else {
		if err := service.UpdateUserQuota(userID, body.Delta, body.ResetUsedCount); err != nil {
			slog.Error("更新配额失败", "user_id", userID, "delta", body.Delta, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新配额失败"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminToggleStatus(c *gin.Context) {
	userID := c.Param("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if body.Status != "active" && body.Status != "disabled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "状态值无效"})
		return
	}
	if err := service.SetUserStatus(userID, body.Status); err != nil {
		slog.Error("更新状态失败", "user_id", userID, "status", body.Status, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新状态失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminToggleUnlimited(c *gin.Context) {
	userID := c.Param("id")
	var body struct {
		Unlimited bool `json:"unlimited"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if err := service.SetUserUnlimited(userID, body.Unlimited); err != nil {
		slog.Error("更新无限配额失败", "user_id", userID, "unlimited", body.Unlimited, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新无限配额失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminDeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if err := service.DeleteUser(userID); err != nil {
		slog.Error("删除用户失败", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminCreateCode(c *gin.Context) {
	var body struct {
		Quota int `json:"quota"`
		Count int `json:"count"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if body.Count <= 0 {
		body.Count = 1
	}
	if body.Count > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "单次最多创建 100 个兑换码"})
		return
	}
	var codes []service.RedemptionCode
	for i := 0; i < body.Count; i++ {
		rc, err := service.CreateRedemptionCode(body.Quota)
		if err != nil {
			slog.Error("创建兑换码失败", "quota", body.Quota, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		codes = append(codes, *rc)
	}
	c.JSON(http.StatusOK, gin.H{"codes": codes})
}

func AdminListCodes(c *gin.Context) {
	codes, err := service.ListRedemptionCodes()
	if err != nil {
		slog.Error("获取兑换码列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取兑换码列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"codes": codes})
}

func AdminDeleteUsers(c *gin.Context) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要删除的用户"})
		return
	}
	deleted, err := service.DeleteUsers(body.IDs)
	if err != nil {
		slog.Error("批量删除用户失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量删除用户失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "deleted": deleted})
}

func AdminDeleteCodes(c *gin.Context) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要删除的兑换码"})
		return
	}
	deleted, err := service.DeleteCodes(body.IDs)
	if err != nil {
		slog.Error("批量删除兑换码失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量删除兑换码失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "deleted": deleted})
}

func AdminGetEndpoints(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"endpoints": config.GetEndpointPool()})
}

func AdminUpdateEndpoints(c *gin.Context) {
	var body struct {
		Endpoints []config.ApiEndpoint `json:"endpoints"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	endpoints := make([]config.ApiEndpoint, 0, len(body.Endpoints))
	for i, ep := range body.Endpoints {
		ep.BaseURL = strings.TrimSpace(ep.BaseURL)
		if ep.BaseURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 个端点缺少 baseUrl", i+1)})
			return
		}
		parsed, err := url.ParseRequestURI(ep.BaseURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 个端点 baseUrl 无效", i+1)})
			return
		}
		if ep.MaxConcurrency < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 个端点最大并发数不能小于 0", i+1)})
			return
		}
		if ep.Priority < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 个端点优先级不能小于 0", i+1)})
			return
		}
		if ep.CostPerImageX10000 < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 个端点成本价不能小于 0", i+1)})
			return
		}
		endpoints = append(endpoints, ep)
	}
	config.SetEndpoints(endpoints)
	service.RefreshLimiters()
	slog.Info("API 端点配置已更新", "count", len(body.Endpoints))
	c.JSON(http.StatusOK, gin.H{"ok": true, "endpoints": config.GetEndpointPool()})
}

func AdminGetPricingConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"endpoints":       config.GetEndpointPool(),
		"salePriceX10000": config.GetSalePriceX10000(),
		"moneyScale":      service.MoneyScale,
	})
}

func AdminUpdatePricingConfig(c *gin.Context) {
	var body struct {
		Endpoints       []config.ApiEndpoint `json:"endpoints"`
		SalePriceX10000 int64                `json:"salePriceX10000"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	if body.SalePriceX10000 < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "售价不能小于 0"})
		return
	}

	endpoints := make([]config.ApiEndpoint, 0, len(body.Endpoints))
	for i, ep := range body.Endpoints {
		ep.BaseURL = strings.TrimSpace(ep.BaseURL)
		if ep.BaseURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 个端点缺少 baseUrl", i+1)})
			return
		}
		parsed, err := url.ParseRequestURI(ep.BaseURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 个端点 baseUrl 无效", i+1)})
			return
		}
		if ep.MaxConcurrency < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 个端点最大并发数不能小于 0", i+1)})
			return
		}
		if ep.Priority < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 个端点优先级不能小于 0", i+1)})
			return
		}
		if ep.CostPerImageX10000 < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第 %d 个端点成本价不能小于 0", i+1)})
			return
		}
		endpoints = append(endpoints, ep)
	}

	config.SetPricingConfig(endpoints, body.SalePriceX10000)
	service.RefreshLimiters()
	slog.Info("定价配置已更新", "endpoint_count", len(endpoints), "salePriceX10000", body.SalePriceX10000)
	c.JSON(http.StatusOK, gin.H{
		"ok":              true,
		"endpoints":       config.GetEndpointPool(),
		"salePriceX10000": config.GetSalePriceX10000(),
		"moneyScale":      service.MoneyScale,
	})
}

func AdminBillingSummary(c *gin.Context) {
	rangeVal := c.Query("range")
	r, err := service.ParseAnalyticsRange(rangeVal, time.Now())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	summary, meta, err := service.GetBillingSummary(r)
	if err != nil {
		slog.Error("获取账单总览失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取账单总览失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"meta":    meta,
		"summary": summary,
	})
}

func AdminBillingTrend(c *gin.Context) {
	rangeVal := c.Query("range")
	r, err := service.ParseAnalyticsRange(rangeVal, time.Now())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	points, meta, err := service.GetBillingTrend(r)
	if err != nil {
		slog.Error("获取账单趋势失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取账单趋势失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"meta":  meta,
		"trend": points,
	})
}

func AdminBillingEndpointBreakdown(c *gin.Context) {
	rangeVal := c.Query("range")
	r, err := service.ParseAnalyticsRange(rangeVal, time.Now())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rows, meta, err := service.GetBillingEndpointBreakdown(r)
	if err != nil {
		slog.Error("获取端点统计失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取端点统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"meta": meta,
		"rows": rows,
	})
}

func AdminBillingUserBreakdown(c *gin.Context) {
	rangeVal := c.Query("range")
	r, err := service.ParseAnalyticsRange(rangeVal, time.Now())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rows, meta, err := service.GetBillingUserBreakdown(r)
	if err != nil {
		slog.Error("获取用户统计失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"meta": meta,
		"rows": rows,
	})
}

// --- New admin handlers (stubs for RED phase) ---

func AdminResetPassword(c *gin.Context) {
	userID := c.Param("id")
	var body struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少需要 8 个字符"})
		return
	}
	if err := service.AdminResetPassword(userID, body.Password); err != nil {
		slog.Error("管理员重置密码失败", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	slog.Info("管理员重置密码", "user_id", userID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func AdminGetInviteConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"inviterReward": config.GetInviteInviterReward(),
		"inviteeReward": config.GetInviteInviteeReward(),
		"defaultQuota":  config.GetInviteDefaultQuota(),
		"inviteEnabled":  config.IsInviteEnabled(),
	})
}

func AdminUpdateInviteConfig(c *gin.Context) {
	var body struct {
		InviterReward int  `json:"inviterReward"`
		InviteeReward int  `json:"inviteeReward"`
		DefaultQuota  int  `json:"defaultQuota"`
		InviteEnabled bool `json:"inviteEnabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	if body.InviterReward < 0 || body.InviteeReward < 0 || body.DefaultQuota < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "奖励值不能为负数"})
		return
	}
	config.SetInviteConfig(body.InviterReward, body.InviteeReward, body.DefaultQuota, body.InviteEnabled)
	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"inviterReward": body.InviterReward,
		"inviteeReward": body.InviteeReward,
		"defaultQuota":  body.DefaultQuota,
		"inviteEnabled":  body.InviteEnabled,
	})
}

func AdminListInvites(c *gin.Context) {
	rows, err := service.ListInvites()
	if err != nil {
		slog.Error("获取邀请码列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取邀请码列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"invites": rows})
}
