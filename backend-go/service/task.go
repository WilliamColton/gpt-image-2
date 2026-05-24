package service

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"gpt-image-playground/backend/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func toTaskModel(userID string, t *TaskRecord) *database.Task {
	paramsJSON, _ := json.Marshal(t.Params)
	inputIDsJSON, _ := json.Marshal(t.InputImageIDs)
	outputIDsJSON, _ := json.Marshal(t.OutputImages)

	var actualParams, actualParamsByImage, revisedPromptByImage *string
	if b, err := json.Marshal(t.ActualParams); err == nil && t.ActualParams != nil {
		s := string(b)
		actualParams = &s
	}
	if b, err := json.Marshal(t.ActualParamsByImage); err == nil && t.ActualParamsByImage != nil {
		s := string(b)
		actualParamsByImage = &s
	}
	if b, err := json.Marshal(t.RevisedPromptByImage); err == nil && t.RevisedPromptByImage != nil {
		s := string(b)
		revisedPromptByImage = &s
	}

	isFav := 0
	if t.IsFavorite {
		isFav = 1
	}
	codex := 0
	if t.CodexCli {
		codex = 1
	}

	return &database.Task{
		ID:                       t.ID,
		UserID:                   userID,
		Prompt:                   t.Prompt,
		ParamsJSON:               string(paramsJSON),
		ActualParamsJSON:         actualParams,
		ActualParamsByImageJSON:  actualParamsByImage,
		RevisedPromptByImageJSON: revisedPromptByImage,
		InputImageIDsJSON:        string(inputIDsJSON),
		MaskTargetImageID:        t.MaskTargetImageID,
		MaskImageID:              t.MaskImageID,
		OutputImageIDsJSON:       string(outputIDsJSON),
		Status:                   t.Status,
		Error:                    t.Error,
		IsFavorite:               isFav,
		CreatedAt:                t.CreatedAt,
		FinishedAt:               t.FinishedAt,
		Elapsed:                  t.Elapsed,
		ApiMode:                  &t.ApiMode,
		CodexCli:                 codex,
	}
}

func toTaskRecord(m *database.Task) TaskRecord {
	t := TaskRecord{
		ID:                m.ID,
		Prompt:            m.Prompt,
		MaskTargetImageID: m.MaskTargetImageID,
		MaskImageID:       m.MaskImageID,
		Status:            m.Status,
		Error:             m.Error,
		CreatedAt:         m.CreatedAt,
		FinishedAt:        m.FinishedAt,
		Elapsed:           m.Elapsed,
	}

	json.Unmarshal([]byte(m.ParamsJSON), &t.Params)
	json.Unmarshal([]byte(m.InputImageIDsJSON), &t.InputImageIDs)
	json.Unmarshal([]byte(m.OutputImageIDsJSON), &t.OutputImages)
	if t.InputImageIDs == nil {
		t.InputImageIDs = []string{}
	}
	if t.OutputImages == nil {
		t.OutputImages = []string{}
	}

	if m.ActualParamsJSON != nil {
		json.Unmarshal([]byte(*m.ActualParamsJSON), &t.ActualParams)
	}
	if m.ActualParamsByImageJSON != nil {
		json.Unmarshal([]byte(*m.ActualParamsByImageJSON), &t.ActualParamsByImage)
	}
	if m.RevisedPromptByImageJSON != nil {
		json.Unmarshal([]byte(*m.RevisedPromptByImageJSON), &t.RevisedPromptByImage)
	}
	if m.ApiMode != nil {
		t.ApiMode = *m.ApiMode
	}
	t.IsFavorite = m.IsFavorite == 1
	t.CodexCli = m.CodexCli == 1
	return t
}

func ListTasks(userID string) ([]TaskRecord, error) {
	var models []database.Task
	if err := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&models).Error; err != nil {
		slog.Error("查询任务列表失败", "user_id", userID, "error", err)
		return nil, err
	}
	tasks := make([]TaskRecord, len(models))
	for i, m := range models {
		tasks[i] = toTaskRecord(&m)
	}
	return tasks, nil
}

func GetTask(userID, taskID string) (*TaskRecord, error) {
	var m database.Task
	err := database.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("任务不存在")
		}
		slog.Error("查询任务失败", "user_id", userID, "task_id", taskID, "error", err)
		return nil, fmt.Errorf("任务不存在")
	}
	t := toTaskRecord(&m)
	return &t, nil
}

func UpsertTask(userID string, task *TaskRecord) error {
	model := toTaskModel(userID, task)
	err := database.DB.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(model).Error
	if err != nil {
		slog.Error("保存任务失败", "user_id", userID, "task_id", task.ID, "error", err)
	}
	return err
}

// QuotaCheckResult carries the result of a transactional quota check + task creation.
type QuotaCheckResult struct {
	Allowed bool
	Error   string
	Task    *TaskRecord // valid only when Allowed
}

// CheckQuotaAndCreateTask atomically checks quota and creates the queued task record
// in a single database transaction, preventing concurrent requests from bypassing
// the quota limit.
func CheckQuotaAndCreateTask(userID string, task *TaskRecord, n int) (*QuotaCheckResult, error) {
	result := &QuotaCheckResult{}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var u database.User
		if err := tx.Where("id = ?", userID).First(&u).Error; err != nil {
			result.Error = "用户不存在"
			return nil
		}

		if u.UnlimitedQuota == 0 {
			pending := countPendingImagesTx(tx, userID)
			if u.UsedCount+pending+n > u.Quota {
				remaining := u.Quota - u.UsedCount - pending
				if remaining < 0 {
					remaining = 0
				}
				slog.Warn("用户配额不足", "user_id", userID, "quota", u.Quota, "used_count", u.UsedCount, "pending", pending, "requested", n)
				result.Error = fmt.Sprintf("配额不足，剩余 %d 张（含进行中任务），本次需要 %d 张", remaining, n)
				return nil
			}
		}

		model := toTaskModel(userID, task)
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(model).Error; err != nil {
			slog.Error("保存任务失败", "user_id", userID, "task_id", task.ID, "error", err)
			return err
		}

		result.Allowed = true
		result.Task = task
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !result.Allowed && result.Error == "" {
		result.Error = "任务创建失败"
	}
	return result, nil
}

// countPendingImagesTx is the transactional variant of CountPendingImages.
func countPendingImagesTx(tx *gorm.DB, userID string) int {
	var tasks []database.Task
	if err := tx.Where("user_id = ? AND status IN ?", userID, []string{"queued", "running"}).Find(&tasks).Error; err != nil {
		slog.Error("查询未完成任务失败", "user_id", userID, "error", err)
		return 0
	}
	total := 0
	for _, t := range tasks {
		var params struct {
			N int `json:"n"`
		}
		if err := json.Unmarshal([]byte(t.ParamsJSON), &params); err != nil || params.N < 1 {
			total += 1
		} else {
			total += params.N
		}
	}
	return total
}

func DeleteTask(userID, taskID string) {
	if err := database.DB.Where("id = ? AND user_id = ?", taskID, userID).Delete(&database.Task{}).Error; err != nil {
		slog.Error("删除任务失败", "user_id", userID, "task_id", taskID, "error", err)
	}
}

func ClearTasks(userID string) {
	if err := database.DB.Where("user_id = ?", userID).Delete(&database.Task{}).Error; err != nil {
		slog.Error("清空任务失败", "user_id", userID, "error", err)
	}
}

// CountPendingImages returns the total number of requested images across all unfinished tasks.
func CountPendingImages(userID string) int {
	var tasks []database.Task
	if err := database.DB.Where("user_id = ? AND status IN ?", userID, []string{"queued", "running"}).Find(&tasks).Error; err != nil {
		slog.Error("查询未完成任务失败", "user_id", userID, "error", err)
		return 0
	}
	total := 0
	for _, t := range tasks {
		var params struct {
			N int `json:"n"`
		}
		if err := json.Unmarshal([]byte(t.ParamsJSON), &params); err != nil || params.N < 1 {
			total += 1
		} else {
			total += params.N
		}
	}
	return total
}
