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
