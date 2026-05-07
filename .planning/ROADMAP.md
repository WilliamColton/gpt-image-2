# Roadmap: GPT Image Playground — 后端接管 API 调用

**Created:** 2026-05-05
**Granularity:** Coarse

## 架构变更说明

**旧架构:** 浏览器 → 直接调用 OpenAI API → 页面刷新 = 连接丢失

**新架构:** 浏览器 → 提交任务到后端 → 后端独立调用 OpenAI → 前端轮询获取结果

后端完全接管 API 调用，前端只负责提交任务和展示结果。页面刷新不影响正在进行的任务。

## Phases

### Phase 1: 后端任务执行层

**Goal:** Go 后端能接收生成任务、独立调用 OpenAI API、存储结果

**Requirements:** TASK-01, TASK-02, TASK-03, TASK-04, TASK-05, TASK-06, CFG-01, CFG-02

**Success Criteria:**
1. 后端 `POST /api/generate` 接收任务参数，异步调用 OpenAI Images API，结果存入数据库
2. 后端 `POST /api/edit` 接收任务参数和参考图，异步调用 OpenAI Edits API，结果存入数据库
3. 后端 `POST /api/responses-generate` 接收任务参数，异步调用 Responses API，结果存入数据库
4. `GET /api/tasks/:id` 返回任务状态 (pending/processing/completed/failed) 和结果
5. API Key 从后端环境变量读取，不从前端传递
6. 任务结果包含 base64 图片、实际参数、改写提示词

**Plans:**
- 01: 新增任务执行 service (OpenAI API 调用逻辑)
- 02: 新增 generate/edit/responses handler 和路由
- 03: 任务状态管理和图片存储
- 04: 环境变量配置 (API Key, Base URL, Model)

**Depends on:** —

---

### Phase 2: 前端适配

**Goal:** 前端改为提交任务 + 轮询结果，移除直接调用 OpenAI 的逻辑

**Requirements:** FE-01, FE-02, FE-03, FE-04, FE-05, CFG-03

**Success Criteria:**
1. 前端 `callImageApi` 改为提交任务到后端，轮询等待结果
2. 页面刷新后任务卡片显示正确状态，完成后可查看结果
3. Images API / Responses API 模式切换正常工作
4. Codex CLI 兼容模式行为不变
5. 前端设置中 API Key 变为可选 (后端有 key 时无需填写)

**Plans:**
- 05: 改造 api.ts — 提交任务 + 轮询
- 06: 更新前端设置和配置逻辑
- 07: 端到端测试验证

**Depends on:** Phase 1

---

## Summary

| Phase | Name | Requirements | Plans |
|-------|------|--------------|-------|
| 1 | 后端任务执行层 | 8 | 4 |
| 2 | 前端适配 | 6 | 3 |

**Total:** 14 requirements, 7 plans

---
*Roadmap created: 2026-05-05*
