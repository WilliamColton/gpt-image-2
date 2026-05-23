# GPT Image Playground

## What This Is

基于 OpenAI 图像生成接口的图片生成与编辑工具。提供简洁的 Web UI，支持文本生成图片、参考图编辑、可视化参数调节、历史记录管理与本地数据导入导出。前端 React + TypeScript，后端 Go (Gin)。

## Core Value

用户可以通过简洁的界面调用 OpenAI 图像生成/编辑 API，生成和管理图片。连接稳定性是核心体验。

## Requirements

### Validated

- ✓ 文本生图 (images/generations + Responses API) — existing
- ✓ 参考图编辑 (images/edits + Responses API 多模态输入) — existing
- ✓ 遮罩编辑 — existing
- ✓ 批量生成 — existing
- ✓ 历史记录管理 (IndexedDB) — existing
- ✓ 响应式布局 + PWA — existing
- ✓ 后端 Auth + 任务/图片存储 — existing

### Active

_All requirements have been validated through Phase 1-5 execution._

### Validated

- ✓ 后端完全接管 OpenAI API 调用 (前端提交任务，后端独立执行) — Validated in Phase 1-2
- ✓ API Key 仅存储在后端 (安全性提升) — Validated in Phase 1
- ✓ 任务状态持久化 (页面刷新不丢失进度) — Validated in Phase 1
- ✓ API 降级机制 (多端点自动切换) — Validated in Phase 3
- ✓ 管理后台 (用户配额、统计、禁用/启用) — Validated in Phase 4
- ✓ 成本与收益统计 (端点成本价、全局售价、记账、分析图表) — Validated in Phase 5

### Out of Scope

- WebSocket/SSE 实时推送 — 代理模式下不需要，用轮询或长连接即可
- 多用户系统改造 — 当前架构已支持

## Context

- 当前架构：前端浏览器直接调用 OpenAI API (`src/lib/api.ts`)，使用 `settings.apiKey` 和 `settings.baseUrl`
- 问题：页面刷新时浏览器 AbortController 取消请求，连接丢失
- 新架构：前端提交任务到后端 → 后端独立调用 OpenAI → 前端轮询获取结果
- Go 后端已有 Auth、图片存储、任务管理等功能 (Gin 框架)

## Constraints

- **Tech stack**: 前端 React 19 + TypeScript + Vite + Tailwind CSS 3, 后端 Go + Gin
- **Compatibility**: 需保持现有功能不变 (Images API + Responses API 两种模式)
- **部署**: 支持 Vercel + Docker 部署方式

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 后端接管而非代理转发 | 代理转发仍依赖前后端连接，刷新照样断；后端独立执行才能解决问题 | Implemented in Phase 1-2 |
| 轮询而非 WebSocket | 轮询简单可靠，现有 task API 已有基础 | Implemented in Phase 2 |

---
*Last updated: 2026-05-23 after Phase 5 completion*
