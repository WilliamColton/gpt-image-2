package handler

import (
	"log/slog"
	"net/http"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/middleware"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

type generateRequest struct {
	TaskID        string             `json:"taskId"`
	Prompt        string             `json:"prompt"`
	Params        service.TaskParams `json:"params"`
	InputImageIDs []string           `json:"inputImageIds"`
	MaskImageID   string             `json:"maskImageId"`
	CodexCli      bool               `json:"codexCli"`
}

func GenerateImage(c *gin.Context) {
	user := middleware.GetAuthUser(c)

	var req generateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// Check quota
	n := req.Params.N
	if n < 1 {
		n = 1
	}
	if err := service.CheckQuota(user.ID, n); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if req.TaskID == "" || req.Prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 taskId 或 prompt"})
		return
	}

	// Create task record
	task := &service.TaskRecord{
		ID:            req.TaskID,
		Prompt:        req.Prompt,
		Params:        req.Params,
		InputImageIDs: req.InputImageIDs,
		Status:        "queued",
		CreatedAt:     time.Now().UnixMilli(),
		ApiMode:       "images",
		CodexCli:      req.CodexCli,
	}
	if req.MaskImageID != "" {
		task.MaskImageID = &req.MaskImageID
	}
	if err := service.UpsertTask(user.ID, task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败"})
		return
	}

	// Execute async — apiKey is empty; withFailover uses each endpoint's own key
	go executeImageGeneration(user.ID, task, req.Params, req.CodexCli, "")

	c.JSON(http.StatusOK, gin.H{"taskId": task.ID, "status": "queued"})
}

func executeImageGeneration(userID string, task *service.TaskRecord, params service.TaskParams, codexCli bool, apiKey string) {
	start := time.Now()

	endpoints := config.GetEndpointPool()
	if len(endpoints) == 0 {
		failTask(userID, task.ID, "未配置任何 API 端点")
		return
	}

	// Load input images
	var imageFiles []service.ImageFileInput
	for _, imgID := range task.InputImageIDs {
		dataURL, err := service.ReadImageDataURLForUser(userID, imgID)
		if err != nil {
			failTask(userID, task.ID, "读取输入图片失败: "+err.Error())
			return
		}
		data, mime, err := service.DataURLToBytes(dataURL)
		if err != nil {
			failTask(userID, task.ID, "解析输入图片失败: "+err.Error())
			return
		}
		imageFiles = append(imageFiles, service.ImageFileInput{Data: data, Mime: mime})
	}

	var maskFile *service.ImageFileInput
	if task.MaskImageID != nil && *task.MaskImageID != "" {
		dataURL, err := service.ReadImageDataURLForUser(userID, *task.MaskImageID)
		if err != nil {
			failTask(userID, task.ID, "读取遮罩图片失败: "+err.Error())
			return
		}
		data, mime, err := service.DataURLToBytes(dataURL)
		if err != nil {
			failTask(userID, task.ID, "解析遮罩图片失败: "+err.Error())
			return
		}
		maskFile = &service.ImageFileInput{Data: data, Mime: mime}
	}

	n := params.N
	if n < 1 {
		n = 1
	}

	// Acquire concurrency slot and execute with failover
	onAcquired := func() {
		task.Status = "running"
		if err := service.UpsertTask(userID, task); err != nil {
			slog.Error("更新任务状态失败", "user_id", userID, "task_id", task.ID, "error", err)
		}
	}

	var result *service.ImageGenResult
	var err error
	if len(imageFiles) > 0 {
		if n > 1 {
			result, err = service.CallImagesEditsConcurrent(task.Prompt, params, imageFiles, maskFile, n, apiKey, onAcquired, endpoints...)
		} else {
			result, err = service.CallImagesEdits(task.Prompt, params, imageFiles, maskFile, codexCli, apiKey, onAcquired, endpoints...)
		}
	} else if codexCli && n > 1 {
		result, err = service.CallImagesGenerationsConcurrent(task.Prompt, params, n, apiKey, onAcquired, endpoints...)
	} else {
		result, err = service.CallImagesGenerations(task.Prompt, params, n, codexCli, apiKey, onAcquired, endpoints...)
	}
	if err != nil {
		failTask(userID, task.ID, err.Error())
		return
	}

	// Save output images and update task
	outputIDs := saveGeneratedImages(userID, result.Images)

	// Increment used_count by number of generated images (per ADMIN-03/05)
	if len(outputIDs) > 0 {
		if err := service.IncrementUsedCount(userID, len(outputIDs)); err != nil {
			slog.Error("更新用户配额计数失败", "user_id", userID, "count", len(outputIDs), "error", err)
		}
	}

	actualParams := mergeActualParams(result)
	actualParamsByImage, revisedPromptByImage := buildPerImageMetadata(outputIDs, result.Images)

	task.OutputImages = outputIDs
	task.Status = "done"
	task.ActualParams = actualParams
	task.ActualParamsByImage = actualParamsByImage
	task.RevisedPromptByImage = revisedPromptByImage
	now := time.Now().UnixMilli()
	task.FinishedAt = &now
	elapsed := now - start.UnixMilli()
	task.Elapsed = &elapsed

	if err := service.UpsertTask(userID, task); err != nil {
		slog.Error("更新任务失败", "user_id", userID, "task_id", task.ID, "error", err)
	}
}

func failTask(userID, taskID, errMsg string) {
	slog.Error("任务失败", "user_id", userID, "task_id", taskID, "error", errMsg)
	task, err := service.GetTask(userID, taskID)
	if err != nil {
		slog.Error("读取失败任务失败", "user_id", userID, "task_id", taskID, "error", err)
		return
	}
	task.Status = "error"
	task.Error = &errMsg
	now := time.Now().UnixMilli()
	task.FinishedAt = &now
	elapsed := now - task.CreatedAt
	task.Elapsed = &elapsed
	service.UpsertTask(userID, task)
}

func saveGeneratedImages(userID string, images []service.GeneratedImage) []string {
	var ids []string
	for _, img := range images {
		saved, err := service.SaveDataURLImage(userID, img.Base64, "generated")
		if err != nil {
			slog.Error("保存生成图片失败", "user_id", userID, "error", err)
			continue
		}
		ids = append(ids, saved.ID)
	}
	return ids
}

func mergeActualParams(result *service.ImageGenResult) map[string]interface{} {
	if result == nil || result.ActualParams == nil {
		return nil
	}
	return result.ActualParams
}

func buildPerImageMetadata(outputIDs []string, images []service.GeneratedImage) (map[string]map[string]interface{}, map[string]string) {
	actualParamsByImage := map[string]map[string]interface{}{}
	revisedPromptByImage := map[string]string{}

	for i, img := range images {
		if i >= len(outputIDs) {
			break
		}
		id := outputIDs[i]
		if img.ActualParams != nil && len(img.ActualParams) > 0 {
			actualParamsByImage[id] = img.ActualParams
		}
		if img.RevisedPrompt != "" {
			revisedPromptByImage[id] = img.RevisedPrompt
		}
	}

	if len(actualParamsByImage) == 0 {
		actualParamsByImage = nil
	}
	if len(revisedPromptByImage) == 0 {
		revisedPromptByImage = nil
	}
	return actualParamsByImage, revisedPromptByImage
}
