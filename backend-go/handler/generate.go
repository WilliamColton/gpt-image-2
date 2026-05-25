package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"gpt-image-playground/backend/config"
	"gpt-image-playground/backend/middleware"
	"gpt-image-playground/backend/service"

	"github.com/gin-gonic/gin"
)

// activeTask pairs a cancel function with the owning user so per-user
// operations (clear) do not interfere with other users' active generation.
type activeTask struct {
	cancel context.CancelFunc
	userID string
}

// activeGenerationTasks tracks cancellable generation goroutines by task ID.
var activeGenerationTasks sync.Map // taskID -> activeTask

// CancelGeneration cancels a running generation task. Returns true if found.
func CancelGeneration(taskID string) bool {
	if v, ok := activeGenerationTasks.LoadAndDelete(taskID); ok {
		at := v.(activeTask)
		at.cancel()
		return true
	}
	return false
}

// CancelUserGenerations cancels all generation tasks belonging to a user.
func CancelUserGenerations(userID string) {
	activeGenerationTasks.Range(func(key, value interface{}) bool {
		at := value.(activeTask)
		if at.userID == userID {
			at.cancel()
			activeGenerationTasks.Delete(key)
		}
		return true
	})
}

type generateRequest struct {
	TaskID        string             `json:"taskId"`
	Prompt        string             `json:"prompt"`
	Params        service.TaskParams `json:"params"`
	InputImageIDs []string           `json:"inputImageIds"`
	MaskImageID   string             `json:"maskImageId"`
	CodexCli      bool               `json:"codexCli"`
}

// savedGeneratedImage pairs a successfully saved output image ID with the
// generated image that produced it. This pairing survives partial save
// failures and provides the data needed for billing and per-image metadata.
type savedGeneratedImage struct {
	OutputImageID string
	Generated     service.GeneratedImage
}

func GenerateImage(c *gin.Context) {
	user := middleware.GetAuthUser(c)

	var req generateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	req.Params.N = service.NormalizeTaskN(req.Params.N)
	n := req.Params.N

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

	// Atomically check quota and create task in a single transaction
	qResult, err := service.CheckQuotaAndCreateTask(user.ID, task, n)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败"})
		return
	}
	if !qResult.Allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": qResult.Error})
		return
	}

	// Execute async — apiKey is empty; withFailover uses each endpoint's own key
	ctx, cancel := context.WithCancel(context.Background())
	activeGenerationTasks.Store(task.ID, activeTask{cancel: cancel, userID: user.ID})
	go executeImageGeneration(ctx, user.ID, user.Label, task, req.Params, req.CodexCli, "")

	c.JSON(http.StatusOK, gin.H{"taskId": task.ID, "status": "queued"})
}

func executeImageGeneration(ctx context.Context, userID string, userLabel string, task *service.TaskRecord, params service.TaskParams, codexCli bool, apiKey string) {
	defer activeGenerationTasks.Delete(task.ID)

	// Recover from panics to avoid crashing the whole server
	defer func() {
		if r := recover(); r != nil {
			slog.Error("生成任务 panic", "user_id", userID, "task_id", task.ID, "panic", r)
			failTask(userID, task.ID, fmt.Sprintf("服务内部错误: %v", r))
		}
	}()

	start := time.Now()

	endpoints := config.GetEndpointPool()
	if len(endpoints) == 0 {
		failTask(userID, task.ID, "未配置任何 API 端点")
		return
	}

	// Check if task was already cancelled before doing expensive work
	select {
	case <-ctx.Done():
		slog.Warn("任务已取消", "user_id", userID, "task_id", task.ID)
		return
	default:
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

	params.N = service.NormalizeTaskN(params.N)
	n := params.N

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

	// Save output images — returns a saved-success slice that pairs each
	// successfully saved output image ID with the generated image it came from.
	saved := saveGeneratedImagesWithAttribution(userID, result.Images)

	// Build outputIDs from the saved-success slice for task storage
	outputIDs := make([]string, 0, len(saved))
	for _, s := range saved {
		outputIDs = append(outputIDs, s.OutputImageID)
	}

	actualParams := mergeActualParams(result)
	actualParamsByImage, revisedPromptByImage := buildPerImageMetadataFromSaved(saved)

	task.OutputImages = outputIDs
	task.Status = "done"
	task.ActualParams = actualParams
	task.ActualParamsByImage = actualParamsByImage
	task.RevisedPromptByImage = revisedPromptByImage
	now := time.Now().UnixMilli()
	task.FinishedAt = &now
	elapsed := now - start.UnixMilli()
	task.Elapsed = &elapsed

	// Atomically record billing, increment used_count, and update task in one transaction
	if len(saved) > 0 {
		billingInput := buildBillingInput(task.ID, userID, userLabel, saved)
		if err := service.FinalizeSuccessfulTask(userID, task, billingInput, len(outputIDs)); err != nil {
			slog.Error("事务写入失败", "user_id", userID, "task_id", task.ID, "error", err)
			for _, s := range saved {
				service.DeleteImageForUser(userID, s.OutputImageID)
			}
			failTask(userID, task.ID, "任务状态写入失败，请联系管理员")
			return
		}
	} else {
		// No images to bill for, but still need to update task status
		if err := service.UpsertTask(userID, task); err != nil {
			slog.Error("更新任务失败", "user_id", userID, "task_id", task.ID, "error", err)
		}
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

// saveGeneratedImagesWithAttribution saves each generated image and returns
// a slice that pairs each saved output image ID with the generated image
// that produced it. If a save fails, that entry is skipped — the pairing
// never shifts and later entries are not affected.
func saveGeneratedImagesWithAttribution(userID string, images []service.GeneratedImage) []savedGeneratedImage {
	var saved []savedGeneratedImage
	for _, img := range images {
		result, err := service.SaveDataURLImage(userID, img.Base64, "generated")
		if err != nil {
			slog.Error("保存生成图片失败", "user_id", userID, "error", err)
			continue
		}
		saved = append(saved, savedGeneratedImage{
			OutputImageID: result.ID,
			Generated:     img,
		})
	}
	return saved
}

// buildBillingInput constructs a BillingBatchInput from the saved-success
// slice. It captures the current global sale price as an immutable snapshot
// so historical reports are unaffected by later price changes.
func buildBillingInput(taskID, userID, userLabel string, saved []savedGeneratedImage) service.BillingBatchInput {
	images := make([]service.BillingImageInput, 0, len(saved))
	for _, s := range saved {
		images = append(images, service.BillingImageInput{
			OutputImageID:           s.OutputImageID,
			EndpointBaseURLSnapshot: s.Generated.EndpointBaseURL,
			UnitCostX10000:          s.Generated.UnitCostX10000,
		})
	}
	return service.BillingBatchInput{
		TaskID:            taskID,
		UserID:            userID,
		UserLabelSnapshot: userLabel,
		UnitSaleX10000:    config.GetSalePriceX10000(),
		Images:            images,
	}
}

func mergeActualParams(result *service.ImageGenResult) map[string]interface{} {
	if result == nil || result.ActualParams == nil {
		return nil
	}
	return result.ActualParams
}

// buildPerImageMetadataFromSaved builds metadata maps from the saved-success
// slice. Each entry uses the saved image ID as key and the corresponding
// generated image's data. This replaces the old index-based buildPerImageMetadata
// which could shift when earlier images failed to save.
func buildPerImageMetadataFromSaved(saved []savedGeneratedImage) (map[string]map[string]interface{}, map[string]string) {
	actualParamsByImage := map[string]map[string]interface{}{}
	revisedPromptByImage := map[string]string{}

	for _, s := range saved {
		if s.Generated.ActualParams != nil && len(s.Generated.ActualParams) > 0 {
			actualParamsByImage[s.OutputImageID] = s.Generated.ActualParams
		}
		if s.Generated.RevisedPrompt != "" {
			revisedPromptByImage[s.OutputImageID] = s.Generated.RevisedPrompt
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
