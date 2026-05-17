package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"gpt-image-playground/backend/config"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type GeneratedImage struct {
	Base64        string
	ActualParams  map[string]interface{}
	RevisedPrompt string
}

type ImageGenResult struct {
	Images       []GeneratedImage
	ActualParams map[string]interface{}
}

type ImageFileInput struct {
	Data []byte
	Mime string
}

type namedReader struct {
	io.Reader
	name        string
	contentType string
}

func (n namedReader) Name() string        { return n.name }
func (n namedReader) ContentType() string { return n.contentType }

func codexPrompt(prompt string) string {
	return "Use the following text as the complete prompt. Do not rewrite it:\n" + prompt
}

func newClient(apiKey, baseURL string) openai.Client {
	return openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(apiKey),
		option.WithMaxRetries(0),
	)
}

// withFailover executes the given function against the endpoint pool.
// It tries each endpoint in order. If the function returns an error,
// it tries the next endpoint. Returns the last error if all endpoints fail.
func withFailover(
	endpoints []config.ApiEndpoint,
	callerAPIKey string,
	onAcquired func(),
	fn func(apiKey, baseURL string) (*ImageGenResult, error),
) (*ImageGenResult, error) {
	if len(endpoints) == 0 {
		slog.Error("failover: 未配置任何 API 端点")
		return nil, fmt.Errorf("no endpoints configured")
	}

	var lastErr error
	var acquireOnce sync.Once
	markAcquired := func() {
		if onAcquired != nil {
			acquireOnce.Do(onAcquired)
		}
	}
	for start := 0; start < len(endpoints); {
		epIdx, release := AcquireSlotFrom(endpoints, start, markAcquired)
		ep := endpoints[epIdx]
		apiKey := ep.APIKey
		if apiKey == "" {
			apiKey = callerAPIKey
		}

		result, err := func() (*ImageGenResult, error) {
			defer release()
			return fn(apiKey, ep.BaseURL)
		}()
		if err == nil {
			return result, nil
		}

		lastErr = err
		start = epIdx + 1
		slog.Warn("failover: 端点失败，尝试下一个", "base_url", ep.BaseURL, "error", err)
	}

	if lastErr == nil {
		return nil, fmt.Errorf("所有端点均失败")
	}
	return nil, fmt.Errorf("所有端点均失败，最后错误: %w", lastErr)
}

// callImagesGenerationsOnce calls /v1/images/generations against a single endpoint.
func callImagesGenerationsOnce(prompt string, params TaskParams, n int, codexCli bool, apiKey string, baseURL string) (*ImageGenResult, error) {
	client := newClient(apiKey, baseURL)

	actualPrompt := prompt
	if codexCli {
		actualPrompt = codexPrompt(prompt)
	}

	p := openai.ImageGenerateParams{
		Model:        openai.ImageModel(config.App.Model),
		Prompt:       actualPrompt,
		Size:         openai.ImageGenerateParamsSize(params.Size),
		OutputFormat: openai.ImageGenerateParamsOutputFormat(params.OutputFormat),
		Moderation:   openai.ImageGenerateParamsModeration(params.Moderation),
	}
	if n > 1 {
		p.N = openai.Int(int64(n))
	}
	if params.OutputFormat != "png" && params.OutputCompression != nil {
		p.OutputCompression = openai.Int(int64(*params.OutputCompression))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	resp, err := client.Images.Generate(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("错误: %w", err)
	}

	return convertImagesResponse(resp, params), nil
}

// CallImagesGenerations calls /v1/images/generations with failover across endpoints.
func CallImagesGenerations(prompt string, params TaskParams, n int, codexCli bool, apiKey string, onAcquired func(), endpoints ...config.ApiEndpoint) (*ImageGenResult, error) {
	return withFailover(endpoints, apiKey, onAcquired, func(epKey, epURL string) (*ImageGenResult, error) {
		return callImagesGenerationsOnce(prompt, params, n, codexCli, epKey, epURL)
	})
}

// callImagesEditsOnce calls /v1/images/edits against a single endpoint.
func callImagesEditsOnce(prompt string, params TaskParams, imageFiles []ImageFileInput, maskFile *ImageFileInput, codexCli bool, apiKey string, baseURL string) (*ImageGenResult, error) {
	client := newClient(apiKey, baseURL)

	actualPrompt := prompt
	if codexCli {
		actualPrompt = codexPrompt(prompt)
	}

	// Build image readers with filenames (SDK derives MIME from extension)
	var readers []io.Reader
	for i, img := range imageFiles {
		mimeToExt := map[string]string{
			"image/jpeg": "jpg",
			"image/jpg":  "jpg",
			"image/png":  "png",
			"image/webp": "webp",
			"image/gif":  "gif",
			"image/bmp":  "bmp",
			"image/tiff": "tiff",
			"image/svg":  "svg",
		}
		ext := "png"
		if e, ok := mimeToExt[img.Mime]; ok {
			ext = e
		}
		readers = append(readers, namedReader{Reader: bytes.NewReader(img.Data), name: fmt.Sprintf("input-%d.%s", i+1, ext), contentType: img.Mime})
	}

	p := openai.ImageEditParams{
		Model:        config.App.Model,
		Prompt:       actualPrompt,
		Size:         openai.ImageEditParamsSize(params.Size),
		OutputFormat: openai.ImageEditParamsOutputFormat(params.OutputFormat),
	}
	//p.SetExtraFields(map[string]any{
	//	"moderation": params.Moderation,
	//})
	p.Image = openai.ImageEditParamsImageUnion{OfFileArray: readers}
	if maskFile != nil {
		p.Mask = namedReader{Reader: bytes.NewReader(maskFile.Data), name: "mask.png", contentType: maskFile.Mime}
	}
	if params.OutputFormat != "png" && params.OutputCompression != nil {
		p.OutputCompression = openai.Int(int64(*params.OutputCompression))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	resp, err := client.Images.Edit(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("错误: %w", err)
	}

	return convertImagesResponse(resp, params), nil
}

// CallImagesEdits calls /v1/images/edits with failover across endpoints.
func CallImagesEdits(prompt string, params TaskParams, imageFiles []ImageFileInput, maskFile *ImageFileInput, codexCli bool, apiKey string, onAcquired func(), endpoints ...config.ApiEndpoint) (*ImageGenResult, error) {
	return withFailover(endpoints, apiKey, onAcquired, func(epKey, epURL string) (*ImageGenResult, error) {
		return callImagesEditsOnce(prompt, params, imageFiles, maskFile, codexCli, epKey, epURL)
	})
}

// CallImagesEditsConcurrent 图生图多图并发调用
func CallImagesEditsConcurrent(prompt string, params TaskParams, imageFiles []ImageFileInput, maskFile *ImageFileInput, n int, apiKey string, onAcquired func(), endpoints ...config.ApiEndpoint) (*ImageGenResult, error) {
	var wg sync.WaitGroup
	var acquireOnce sync.Once
	results := make([]*ImageGenResult, n)
	errs := make([]error, n)

	markAcquired := func() {
		if onAcquired != nil {
			acquireOnce.Do(onAcquired)
		}
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = CallImagesEdits(prompt, params, imageFiles, maskFile, false, apiKey, markAcquired, endpoints...)
		}(i)
	}
	wg.Wait()

	return mergeConcurrentResults(results, errs)
}

func mergeConcurrentResults(results []*ImageGenResult, errs []error) (*ImageGenResult, error) {
	var allImages []GeneratedImage
	var firstActual map[string]interface{}
	var lastErr error
	for i, r := range results {
		if errs[i] != nil {
			lastErr = errs[i]
			continue
		}
		if r == nil {
			continue
		}
		if firstActual == nil && r.ActualParams != nil {
			firstActual = r.ActualParams
		}
		allImages = append(allImages, r.Images...)
	}

	if len(allImages) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("所有并发请求均失败: %w", lastErr)
		}
		return nil, fmt.Errorf("所有并发请求均失败")
	}

	merged := map[string]interface{}{}
	for k, v := range firstActual {
		merged[k] = v
	}
	merged["n"] = len(allImages)

	return &ImageGenResult{Images: allImages, ActualParams: merged}, nil
}

func convertImagesResponse(resp *openai.ImagesResponse, params TaskParams) *ImageGenResult {
	actualParams := map[string]interface{}{}
	if resp.Size != "" {
		actualParams["size"] = string(resp.Size)
	}
	if resp.Quality != "" {
		actualParams["quality"] = string(resp.Quality)
	}
	if resp.OutputFormat != "" {
		actualParams["output_format"] = string(resp.OutputFormat)
	}

	images := make([]GeneratedImage, 0, len(resp.Data))
	for _, item := range resp.Data {
		if item.B64JSON == "" {
			continue
		}
		b64 := item.B64JSON
		if !strings.HasPrefix(b64, "data:") {
			mime := "image/png"
			of := string(resp.OutputFormat)
			if of == "jpeg" {
				mime = "image/jpeg"
			} else if of == "webp" {
				mime = "image/webp"
			}
			b64 = "data:" + mime + ";base64," + b64
		}

		imgActual := map[string]interface{}{}
		for k, v := range actualParams {
			imgActual[k] = v
		}

		images = append(images, GeneratedImage{
			Base64:        b64,
			ActualParams:  imgActual,
			RevisedPrompt: item.RevisedPrompt,
		})
	}

	return &ImageGenResult{Images: images, ActualParams: actualParams}
}

// CallImagesGenerationsConcurrent Codex CLI 模式下多图并发调用
func CallImagesGenerationsConcurrent(prompt string, params TaskParams, n int, apiKey string, onAcquired func(), endpoints ...config.ApiEndpoint) (*ImageGenResult, error) {
	var wg sync.WaitGroup
	var acquireOnce sync.Once
	results := make([]*ImageGenResult, n)
	errs := make([]error, n)

	markAcquired := func() {
		if onAcquired != nil {
			acquireOnce.Do(onAcquired)
		}
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			singleParams := params
			singleParams.Quality = "auto"
			results[idx], errs[idx] = CallImagesGenerations(prompt, singleParams, 1, true, apiKey, markAcquired, endpoints...)
		}(i)
	}
	wg.Wait()

	return mergeConcurrentResults(results, errs)
}

// DataURLToBytes 将 data URL 转为 []byte
func DataURLToBytes(dataURL string) ([]byte, string, error) {
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("无效的 data URL")
	}

	mime := "image/png"
	if idx := strings.Index(parts[0], ":"); idx >= 0 {
		if end := strings.Index(parts[0], ";"); end > idx {
			mime = parts[0][idx+1 : end]
		}
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", fmt.Errorf("base64 解码失败: %w", err)
	}
	return data, mime, nil
}
