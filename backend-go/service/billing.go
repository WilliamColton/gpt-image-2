package service

import (
	"time"

	"gpt-image-playground/backend/database"
	"gpt-image-playground/backend/util"
)

// BillingImageInput carries per-output-image billing data.
type BillingImageInput struct {
	OutputImageID           string
	EndpointBaseURLSnapshot string
	UnitCostX10000          int64
}

// BillingBatchInput carries the batch-level billing context.
type BillingBatchInput struct {
	TaskID            string
	UserID            string
	UserLabelSnapshot string
	UnitSaleX10000    int64
	Images            []BillingImageInput
	CreatedAt         int64
}

// RecordBillingForSuccessfulImages writes one billing_records row
// per successfully saved output image. Each row gets a unique non-empty ID
// from util.GenerateID(). Returns nil and creates zero rows when Images is empty.
// When CreatedAt is 0, defaults to the current UnixMilli time.
// Billing rows are independent of task/user lifetime (no foreign keys).
func RecordBillingForSuccessfulImages(input BillingBatchInput) error {
	if len(input.Images) == 0 {
		return nil
	}

	now := input.CreatedAt
	if now == 0 {
		now = time.Now().UnixMilli()
	}

	records := make([]database.BillingRecord, 0, len(input.Images))
	for _, img := range input.Images {
		records = append(records, database.BillingRecord{
			ID:                      util.GenerateID(),
			TaskID:                  input.TaskID,
			UserID:                  input.UserID,
			UserLabelSnapshot:       input.UserLabelSnapshot,
			EndpointBaseURLSnapshot: img.EndpointBaseURLSnapshot,
			OutputImageID:           img.OutputImageID,
			SuccessImageCount:       1,
			UnitCostX10000:          img.UnitCostX10000,
			UnitSaleX10000:          input.UnitSaleX10000,
			CostX10000:              img.UnitCostX10000,
			RevenueX10000:           input.UnitSaleX10000,
			ProfitX10000:            input.UnitSaleX10000 - img.UnitCostX10000,
			CreatedAt:               now,
		})
	}

	return database.DB.Create(&records).Error
}
