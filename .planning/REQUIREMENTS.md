# Requirements: GPT Image Playground — 后端接管 API 调用

**Defined:** 2026-05-05
**Core Value:** 连接稳定性 — 页面刷新不中断图片生成任务

## v1 Requirements

### 后端任务执行

- [ ] **TASK-01**: 后端接收图片生成任务，独立调用 OpenAI `images/generations` API，结果存入数据库
- [ ] **TASK-02**: 后端接收图片编辑任务，独立调用 OpenAI `images/edits` API (multipart)，结果存入数据库
- [ ] **TASK-03**: 后端接收 Responses API 任务，独立调用 OpenAI `/v1/responses`，结果存入数据库
- [ ] **TASK-04**: 任务有完整生命周期状态 (pending → processing → completed/failed)
- [ ] **TASK-05**: `GET /api/tasks/:id` 返回任务状态和结果 (含 base64 图片、实际参数、改写提示词)
- [ ] **TASK-06**: 多图生成时后端并发调用，合并结果

### 前端适配

- [ ] **FE-01**: `callImageApi` 改为提交任务到后端，轮询等待结果
- [ ] **FE-02**: 移除前端直接向 OpenAI 发送请求和 API Key 的逻辑
- [ ] **FE-03**: 保持 Images API 和 Responses API 两种模式的切换能力
- [ ] **FE-04**: 保持 Codex CLI 兼容模式的行为 (quality=auto, 并发单图请求, 提示词前缀)
- [ ] **FE-05**: 任务卡片显示实时状态，页面刷新后状态不丢失

### 配置

- [ ] **CFG-01**: 后端支持通过环境变量配置 `OPENAI_API_KEY`
- [ ] **CFG-02**: 后端支持通过环境变量配置 `OPENAI_BASE_URL`
- [ ] **CFG-03**: 前端设置中 API Key 改为可选 (后端已有 key 时可不填)

## Out of Scope

| Feature | Reason |
|---------|--------|
| WebSocket/SSE 实时推送 | 轮询足够，复杂度不值得 |
| 任务队列 (Redis/RabbitMQ) | 单机场景下 goroutine 并发足够 |
| 多 API Key 管理 | 超出当前范围 |
| 任务取消/重试 | v1 不需要 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| TASK-01 | Phase 1 | Pending |
| TASK-02 | Phase 1 | Pending |
| TASK-03 | Phase 1 | Pending |
| TASK-04 | Phase 1 | Pending |
| TASK-05 | Phase 1 | Pending |
| TASK-06 | Phase 1 | Pending |
| FE-01 | Phase 2 | Pending |
| FE-02 | Phase 2 | Pending |
| FE-03 | Phase 2 | Pending |
| FE-04 | Phase 2 | Pending |
| FE-05 | Phase 2 | Pending |
| CFG-01 | Phase 1 | Pending |
| CFG-02 | Phase 1 | Pending |
| CFG-03 | Phase 2 | Pending |

**Coverage:**
- v1 requirements: 14 total
- Mapped to phases: 14
- Unmapped: 0 ✓

---
*Requirements defined: 2026-05-05*
*Last updated: 2026-05-05 after architecture correction*
