package service

import (
	"database/sql"
	"encoding/json"

	"gpt-image-playground/backend/database"
)

func parseJSON[T any](value sql.NullString, fallback T) T {
	if !value.Valid {
		return fallback
	}
	var result T
	if err := json.Unmarshal([]byte(value.String), &result); err != nil {
		return fallback
	}
	return result
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func ListTasks(userID string) ([]TaskRecord, error) {
	rows, err := database.DB.Query("SELECT * FROM tasks WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks = make([]TaskRecord, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}
	return tasks, rows.Err()
}

func scanTask(scanner interface{ Scan(...interface{}) error }) (*TaskRecord, error) {
	t := &TaskRecord{}
	var paramsJSON, inputImageIDsJSON, outputImageIDsJSON string
	var actualParamsJSON, actualParamsByImageJSON, revisedPromptByImageJSON sql.NullString
	var isFavorite int

	var dummyUserID string
	err := scanner.Scan(
		&t.ID, &dummyUserID, &t.Prompt,
		&paramsJSON, &actualParamsJSON, &actualParamsByImageJSON,
		&revisedPromptByImageJSON, &inputImageIDsJSON,
		&t.MaskTargetImageID, &t.MaskImageID,
		&outputImageIDsJSON, &t.Status, &t.Error,
		&isFavorite, &t.CreatedAt, &t.FinishedAt, &t.Elapsed,
	)
	if err != nil {
		return nil, err
	}

	t.Params = parseJSON[interface{}](sql.NullString{String: paramsJSON, Valid: true}, struct{}{})
	t.InputImageIDs = parseJSON[[]string](sql.NullString{String: inputImageIDsJSON, Valid: true}, []string{})
	t.OutputImages = parseJSON[[]string](sql.NullString{String: outputImageIDsJSON, Valid: true}, []string{})
	t.IsFavorite = isFavorite == 1

	if v := parseJSON[interface{}](actualParamsJSON, nil); v != nil {
		t.ActualParams = v
	}
	if v := parseJSON[interface{}](actualParamsByImageJSON, nil); v != nil {
		t.ActualParamsByImage = v
	}
	if v := parseJSON[interface{}](revisedPromptByImageJSON, nil); v != nil {
		t.RevisedPromptByImage = v
	}
	return t, nil
}

func GetTask(userID, taskID string) (*TaskRecord, error) {
	row := database.DB.QueryRow("SELECT * FROM tasks WHERE id = ? AND user_id = ?", taskID, userID)
	t, err := scanTask(row)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func UpsertTask(userID string, task *TaskRecord) error {
	paramsJSON, _ := json.Marshal(task.Params)
	var actualParamsJSON, actualParamsByImageJSON, revisedPromptByImageJSON []byte
	if task.ActualParams != nil {
		actualParamsJSON, _ = json.Marshal(task.ActualParams)
	}
	if task.ActualParamsByImage != nil {
		actualParamsByImageJSON, _ = json.Marshal(task.ActualParamsByImage)
	}
	if task.RevisedPromptByImage != nil {
		revisedPromptByImageJSON, _ = json.Marshal(task.RevisedPromptByImage)
	}
	inputImageIDsJSON, _ := json.Marshal(task.InputImageIDs)
	outputImageIDsJSON, _ := json.Marshal(task.OutputImages)

	isFavorite := 0
	if task.IsFavorite {
		isFavorite = 1
	}

	_, err := database.DB.Exec(`
		INSERT INTO tasks (
			id, user_id, prompt, params_json, actual_params_json, actual_params_by_image_json,
			revised_prompt_by_image_json, input_image_ids_json, mask_target_image_id, mask_image_id,
			output_image_ids_json, status, error, is_favorite, created_at, finished_at, elapsed
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			prompt = excluded.prompt,
			params_json = excluded.params_json,
			actual_params_json = excluded.actual_params_json,
			actual_params_by_image_json = excluded.actual_params_by_image_json,
			revised_prompt_by_image_json = excluded.revised_prompt_by_image_json,
			input_image_ids_json = excluded.input_image_ids_json,
			mask_target_image_id = excluded.mask_target_image_id,
			mask_image_id = excluded.mask_image_id,
			output_image_ids_json = excluded.output_image_ids_json,
			status = excluded.status,
			error = excluded.error,
			is_favorite = excluded.is_favorite,
			finished_at = excluded.finished_at,
			elapsed = excluded.elapsed
	`,
		task.ID, userID, task.Prompt,
		string(paramsJSON),
		nullBytes(actualParamsJSON),
		nullBytes(actualParamsByImageJSON),
		nullBytes(revisedPromptByImageJSON),
		string(inputImageIDsJSON),
		task.MaskTargetImageID,
		task.MaskImageID,
		string(outputImageIDsJSON),
		task.Status,
		task.Error,
		isFavorite,
		task.CreatedAt,
		task.FinishedAt,
		task.Elapsed,
	)
	return err
}

func nullBytes(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}

func DeleteTask(userID, taskID string) {
	database.DB.Exec("DELETE FROM tasks WHERE id = ? AND user_id = ?", taskID, userID)
}

func ClearTasks(userID string) {
	database.DB.Exec("DELETE FROM tasks WHERE user_id = ?", userID)
}
