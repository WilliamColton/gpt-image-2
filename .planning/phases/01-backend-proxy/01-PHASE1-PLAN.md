---
phase: 1
wave: 1
files_modified:
  - backend-go/config/config.go
  - backend-go/service/openai.go
  - backend-go/service/task.go
  - backend-go/service/models.go
  - backend-go/handler/generate.go
  - backend-go/main.go
autonomous: true
requirements:
  - TASK-01
  - TASK-02
  - TASK-03
  - TASK-04
  - TASK-05
  - TASK-06
  - CFG-01
  - CFG-02
---

# Phase 1: 后端任务执行层

## Goal

Go 后端能接收生成任务、独立调用 OpenAI API、存储结果。前端提交任务后可以断开，后端独立完成。

## Tasks

### Task 1: 扩展配置 — 读取 OpenAI 环境变量

**Read first:** `backend-go/config/config.go`

**Action:**
- `config.go` 新增字段 `OpenAIKey`、`OpenAIModel`、`OpenAIImagesModel`、`OpenAIResponsesModel`
- 从环境变量 `OPENAI_API_KEY`、`OPENAI_BASE_URL`、`OPENAI_IMAGES_MODEL`、`OPENAI_RESPONSES_MODEL` 读取
- `OpenAIImagesModel` 默认 `gpt-image-2`，`OpenAIResponsesModel` 默认 `gpt-5.5`

**Acceptance criteria:**
- `config.go` 包含 `OpenAIKey string` 和 `OpenAIBaseURL string` 字段
- 环境变量 `OPENAI_API_KEY` 被读取
- 缺少 key 时 log 警告但不 fatal

---

### Task 2: 新增 OpenAI API 客户端 service

**Read first:** `src/lib/api.ts` (参考前端调用逻辑), `backend-go/service/models.go`

**Action:** 新建 `backend-go/service/openai.go`

实现 3 个函数：
- `CallImagesGenerations(params, prompt, n, codexCli)` — 调用 `/v1/images/generations`
- `CallImagesEdits(params, prompt, imageFiles, maskFile, codexCli)` — 调用 `/v1/images/edits` (multipart)
- `CallResponsesGenerate(params, prompt, imageDataUrls, maskDataUrl, codexCli)` — 调用 `/v1/responses`

每个函数返回：`[]GeneratedImage`（含 base64、actualParams、revisedPrompt）

**关键逻辑 (从 api.ts 迁移):**
- Images API: `model`, `prompt`, `size`, `output_format`, `moderation`, `quality`（codexCli 时省略）, `output_compression`, `n`
- Edits API: multipart form，`image[]` 字段支持多文件，可选 `mask`
- Responses API: `input` 含 `input_text` + `input_image`，`tools` 含 `image_generation` tool
- Codex CLI: prompt 前加 `Use the following text as the complete prompt. Do not rewrite it:\n`，quality 设为 auto，多图用并发单图请求

**Acceptance criteria:**
- `openai.go` 包含 `CallImagesGenerations`, `CallImagesEdits`, `CallResponsesGenerate` 三个导出函数
- 使用 `config.App.OpenAIBaseURL` 作为 base URL
- 使用 `config.App.OpenAIKey` 作为 Authorization
- 错误时返回 OpenAI 的错误消息

---

### Task 3: 扩展 TaskRecord 模型

**Read first:** `backend-go/service/models.go`, `backend-go/service/task.go`

**Action:**
- `TaskRecord` 新增 `ApiMode string` 字段（`images` / `responses`）
- `TaskRecord` 新增 `CodexCli bool` 字段
- `scanTask` 和 `UpsertTask` 适配新字段
- 数据库 schema 新增 `api_mode TEXT` 和 `codex_cli INTEGER` 列

**Acceptance criteria:**
- `models.go` 的 `TaskRecord` 包含 `ApiMode` 和 `CodexCli` 字段
- `UpsertTask` 写入 `api_mode` 和 `codex_cli` 列

---

### Task 4: 新增任务提交 handler

**Read first:** `backend-go/handler/tasks.go`, `backend-go/service/openai.go`, `src/store.ts` (参考 submitTask 逻辑)

**Action:** 新建 `backend-go/handler/generate.go`

实现 3 个 handler：
- `GenerateImage(c *gin.Context)` — `POST /api/generate`
- `EditImage(c *gin.Context)` — `POST /api/edit`
- `GenerateResponses(c *gin.Context)` — `POST /api/responses-generate`

请求体：
```json
{
  "taskId": "xxx",
  "prompt": "...",
  "params": { "size": "1024x1024", "quality": "high", ... },
  "inputImageIds": ["id1", "id2"],
  "maskImageId": "mask-id",
  "codexCli": false
}
```

流程：
1. 解析请求，从数据库读取输入图片
2. 创建 TaskRecord（status=processing）
3. 启动 goroutine 调用 OpenAI API
4. goroutine 内：调用 API → 存结果图片到数据库 → 更新 TaskRecord（status=completed, outputImages）
5. 立即返回 `{ "taskId": "xxx", "status": "processing" }`

**Acceptance criteria:**
- `POST /api/generate` 返回 `{ "taskId": "...", "status": "processing" }`
- goroutine 内完成 OpenAI 调用后任务 status 变为 `done`
- 失败时 status 变为 `error`，error 字段有值
- 输出图片存入 images 表，关联到任务

---

### Task 5: 注册路由

**Read first:** `backend-go/main.go`

**Action:**
- 在 `main.go` 中注册 3 个新路由：
  - `POST /api/generate` → `handler.GenerateImage`
  - `POST /api/edit` → `handler.EditImage`
  - `POST /api/responses-generate` → `handler.GenerateResponses`
- 路由需要 `AuthMiddleware()`

**Acceptance criteria:**
- `main.go` 包含 `r.POST("/api/generate", middleware.AuthMiddleware(), handler.GenerateImage)`
- 3 个路由全部注册

---

### Task 6: 配置新增 API 端点返回给前端

**Read first:** `backend-go/handler/config.go`

**Action:**
- `/api/config/public` 响应新增 `openaiConfigured: true/false` 字段
- 前端可据此判断是否需要填写 API Key

**Acceptance criteria:**
- `GET /api/config/public` 响应包含 `openaiConfigured` 字段

## must_haves

- API Key 从后端环境变量读取，不经过前端
- goroutine 异步执行，不阻塞 HTTP 响应
- 多图生成（codexCli 模式）用并发 goroutine
- 结果图片存入数据库并关联到任务
