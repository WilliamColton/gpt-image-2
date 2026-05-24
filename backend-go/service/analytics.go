package service

import (
	"fmt"
	"strings"
	"time"

	"gpt-image-playground/backend/database"
)

// AnalyticsRange describes a time range for analytics queries.
type AnalyticsRange struct {
	Label string
	From  int64
	To    int64
}

// AnalyticsMeta carries metadata for analytics API responses.
type AnalyticsMeta struct {
	Range      string `json:"range"`
	From       int64  `json:"from"`
	To         int64  `json:"to"`
	MoneyScale int64  `json:"moneyScale"`
}

// BillingSummary holds aggregated financial totals for a time range.
type BillingSummary struct {
	RevenueX10000 int64 `json:"revenueX10000"`
	CostX10000    int64 `json:"costX10000"`
	ProfitX10000  int64 `json:"profitX10000"`
	SuccessImages int   `json:"successImages"`
}

// BillingTrendPoint holds aggregated data for a single date bucket.
type BillingTrendPoint struct {
	Bucket         string `json:"bucket"`
	RevenueX10000  int64  `json:"revenueX10000"`
	CostX10000     int64  `json:"costX10000"`
	ProfitX10000   int64  `json:"profitX10000"`
	SuccessImages  int    `json:"successImages"`
}

// BillingEndpointRow holds aggregated data grouped by endpoint.
type BillingEndpointRow struct {
	EndpointBaseURL string `json:"endpointBaseUrl"`
	EndpointLabel   string `json:"endpointLabel"`
	SuccessImages   int    `json:"successImages"`
	RevenueX10000   int64  `json:"revenueX10000"`
	CostX10000      int64  `json:"costX10000"`
	ProfitX10000    int64  `json:"profitX10000"`
	ProfitRateBps   int64  `json:"profitRateBps"`
}

// BillingUserRow holds aggregated data grouped by user.
type BillingUserRow struct {
	UserID        string `json:"userId"`
	UserLabel     string `json:"userLabel"`
	SuccessImages int    `json:"successImages"`
	RevenueX10000 int64  `json:"revenueX10000"`
	CostX10000    int64  `json:"costX10000"`
	ProfitX10000  int64  `json:"profitX10000"`
	ProfitRateBps int64  `json:"profitRateBps"`
}

// ParseAnalyticsRange parses a range query value and returns the corresponding time window.
// Supported values: "today", "7d", "30d", "all". Empty string defaults to "7d".
// now is the reference time (typically time.Now()).
func ParseAnalyticsRange(value string, now time.Time) (AnalyticsRange, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "7d"
	}

	if value != "today" && value != "7d" && value != "30d" && value != "all" {
		return AnalyticsRange{}, fmt.Errorf("unsupported range: %q", value)
	}

	r := AnalyticsRange{Label: value, To: now.UnixMilli()}

	switch value {
	case "today":
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		r.From = startOfDay.UnixMilli()
	case "7d":
		r.From = now.Add(-7 * 24 * time.Hour).UnixMilli()
	case "30d":
		r.From = now.Add(-30 * 24 * time.Hour).UnixMilli()
	case "all":
		r.From = 0
	}

	return r, nil
}

// newAnalyticsMeta constructs AnalyticsMeta from a range.
func newAnalyticsMeta(r AnalyticsRange) AnalyticsMeta {
	return AnalyticsMeta{
		Range:      r.Label,
		From:       r.From,
		To:         r.To,
		MoneyScale: MoneyScale,
	}
}

// GetBillingSummary returns aggregated totals for the given time range.
func GetBillingSummary(r AnalyticsRange) (BillingSummary, AnalyticsMeta, error) {
	meta := newAnalyticsMeta(r)

	type sumResult struct {
		RevenueX10000 int64
		CostX10000    int64
		ProfitX10000  int64
		SuccessImages int
	}

	var result sumResult
	q := database.DB.Model(&database.BillingRecord{})
	if r.From > 0 {
		q = q.Where("created_at >= ?", r.From)
	}
	q = q.Where("created_at <= ?", r.To)

	if err := q.Select(
		"COALESCE(SUM(revenue_x10000), 0) as revenue_x10000",
		"COALESCE(SUM(cost_x10000), 0) as cost_x10000",
		"COALESCE(SUM(profit_x10000), 0) as profit_x10000",
		"COALESCE(SUM(success_image_count), 0) as success_images",
	).Scan(&result).Error; err != nil {
		return BillingSummary{}, meta, fmt.Errorf("GetBillingSummary: %w", err)
	}

	return BillingSummary{
		RevenueX10000: result.RevenueX10000,
		CostX10000:    result.CostX10000,
		ProfitX10000:  result.ProfitX10000,
		SuccessImages: result.SuccessImages,
	}, meta, nil
}

// GetBillingTrend returns daily bucketed aggregates for the given time range.
// Buckets are sorted ascending by date.
func GetBillingTrend(r AnalyticsRange) ([]BillingTrendPoint, AnalyticsMeta, error) {
	meta := newAnalyticsMeta(r)

	type bucketRow struct {
		Bucket        string
		RevenueX10000 int64
		CostX10000    int64
		ProfitX10000  int64
		SuccessImages int
	}

	var rows []bucketRow
	// Group by date string extracted from CreatedAt (UnixMilli)
	// SQLite doesn't have native date formatting from Unix millis, so we group by
	// the date part of (CreatedAt / 1000) converted to a date.
	// Using strftime: strftime('%Y-%m-%d', created_at / 1000, 'unixepoch')
	q := database.DB.Model(&database.BillingRecord{}).
		Select(
			`strftime('%Y-%m-%d', created_at / 1000, 'unixepoch') as bucket`,
			"COALESCE(SUM(revenue_x10000), 0) as revenue_x10000",
			"COALESCE(SUM(cost_x10000), 0) as cost_x10000",
			"COALESCE(SUM(profit_x10000), 0) as profit_x10000",
			"COALESCE(SUM(success_image_count), 0) as success_images",
		)
	if r.From > 0 {
		q = q.Where("created_at >= ?", r.From)
	}
	q = q.Where("created_at <= ?", r.To)

	if err := q.Group("bucket").Order("bucket asc").Scan(&rows).Error; err != nil {
		return nil, meta, fmt.Errorf("GetBillingTrend: %w", err)
	}

	points := make([]BillingTrendPoint, len(rows))
	for i, row := range rows {
		points[i] = BillingTrendPoint{
			Bucket:        row.Bucket,
			RevenueX10000: row.RevenueX10000,
			CostX10000:    row.CostX10000,
			ProfitX10000:  row.ProfitX10000,
			SuccessImages: row.SuccessImages,
		}
	}

	return points, meta, nil
}

// GetBillingEndpointBreakdown returns aggregated data grouped by endpoint.
// Results are sorted by profit_x10000 DESC, revenue_x10000 DESC.
func GetBillingEndpointBreakdown(r AnalyticsRange) ([]BillingEndpointRow, AnalyticsMeta, error) {
	meta := newAnalyticsMeta(r)

	type groupRow struct {
		EndpointBaseURL string
		RevenueX10000   int64
		CostX10000      int64
		ProfitX10000    int64
		SuccessImages   int
	}

	var rows []groupRow
	q := database.DB.Model(&database.BillingRecord{}).
		Select(
			"endpoint_base_url_snapshot as endpoint_base_url",
			"COALESCE(SUM(revenue_x10000), 0) as revenue_x10000",
			"COALESCE(SUM(cost_x10000), 0) as cost_x10000",
			"COALESCE(SUM(profit_x10000), 0) as profit_x10000",
			"COALESCE(SUM(success_image_count), 0) as success_images",
		)
	if r.From > 0 {
		q = q.Where("created_at >= ?", r.From)
	}
	q = q.Where("created_at <= ?", r.To)

	if err := q.Group("endpoint_base_url_snapshot").
		Order("profit_x10000 desc, revenue_x10000 desc").
		Scan(&rows).Error; err != nil {
		return nil, meta, fmt.Errorf("GetBillingEndpointBreakdown: %w", err)
	}

	result := make([]BillingEndpointRow, len(rows))
	for i, row := range rows {
		result[i] = BillingEndpointRow{
			EndpointBaseURL: row.EndpointBaseURL,
			EndpointLabel:   row.EndpointBaseURL, // label defaults to base URL
			RevenueX10000:   row.RevenueX10000,
			CostX10000:      row.CostX10000,
			ProfitX10000:    row.ProfitX10000,
			SuccessImages:   row.SuccessImages,
			ProfitRateBps:   calcProfitRateBps(row.ProfitX10000, row.RevenueX10000),
		}
	}

	return result, meta, nil
}

// GetBillingUserBreakdown returns aggregated data grouped by user.
// Results are sorted by profit_x10000 DESC, revenue_x10000 DESC.
func GetBillingUserBreakdown(r AnalyticsRange) ([]BillingUserRow, AnalyticsMeta, error) {
	meta := newAnalyticsMeta(r)

	type groupRow struct {
		UserID             string
		UserLabel          string
		RevenueX10000      int64
		CostX10000         int64
		ProfitX10000       int64
		SuccessImages      int
	}

	var rows []groupRow
	q := database.DB.Model(&database.BillingRecord{}).
		Select(
			"user_id",
			"user_label_snapshot as user_label",
			"COALESCE(SUM(revenue_x10000), 0) as revenue_x10000",
			"COALESCE(SUM(cost_x10000), 0) as cost_x10000",
			"COALESCE(SUM(profit_x10000), 0) as profit_x10000",
			"COALESCE(SUM(success_image_count), 0) as success_images",
		)
	if r.From > 0 {
		q = q.Where("created_at >= ?", r.From)
	}
	q = q.Where("created_at <= ?", r.To)

	if err := q.Group("user_id, user_label_snapshot").
		Order("profit_x10000 desc, revenue_x10000 desc").
		Scan(&rows).Error; err != nil {
		return nil, meta, fmt.Errorf("GetBillingUserBreakdown: %w", err)
	}

	result := make([]BillingUserRow, len(rows))
	for i, row := range rows {
		result[i] = BillingUserRow{
			UserID:        row.UserID,
			UserLabel:     row.UserLabel,
			RevenueX10000: row.RevenueX10000,
			CostX10000:    row.CostX10000,
			ProfitX10000:  row.ProfitX10000,
			SuccessImages: row.SuccessImages,
			ProfitRateBps: calcProfitRateBps(row.ProfitX10000, row.RevenueX10000),
		}
	}

	return result, meta, nil
}

// calcProfitRateBps computes profit rate in basis points (1/10000).
// Returns profitX10000 * 10000 / revenueX10000, or 0 when revenue is 0.
func calcProfitRateBps(profitX10000, revenueX10000 int64) int64 {
	if revenueX10000 <= 0 {
		return 0
	}
	return profitX10000 * 10000 / revenueX10000
}
