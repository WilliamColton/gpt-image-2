# GPT Image Playground

## What This Is

这是一个面向公开用户的 AI 生图网站：用户可以登录、提交文生图或图生图任务、查看历史任务与生成结果；管理员可以管理用户、配额、生图端点、定价、邀请码、公告、更新日志和运营分析。

当前代码库已经具备完整的前后端雏形：React/Vite 前端、Go/Gin 后端、SQLite 存储、本地图片文件存储，以及 OpenAI-compatible 图片 API 多端点 failover。后续 GSD 工作不是从零建设，而是把现有功能继续完善到可稳定公开运营的状态。

## Core Value

公开用户可以稳定、可控、可追踪地完成图片生成，而管理员可以准确管理额度、成本、端点和运营风险。

## Requirements

### Validated

- 已有用户端生图体验：用户可以提交文生图/图生图任务，查看任务状态、历史记录和生成图片结果。
- 已有账号与访问控制能力：支持兑换码登录、用户名密码注册/登录、账号迁移、密码修改、邀请码体系和 JWT 鉴权。
- 已有管理员基础运营能力：管理员可以登录后台，管理用户、用户状态、配额、兑换码、邀请码配置、公告、更新日志和反馈。
- 已有生图端点运营能力：后端支持 OpenAI-compatible 图片 API、多端点配置、优先级、并发槽位、失败切换和 Codex CLI 兼容模式。
- 已有计费与分析雏形：生成成功后记录成本/售价/利润快照，并提供管理员端汇总、趋势、端点和用户维度分析。
- 已有本地持久化能力：SQLite 保存用户、任务、图片元数据、账单和运营内容；文件系统保存上传与生成图片；前端使用 IndexedDB 和内存缓存优化图片访问。

### Active

- [ ] 管理员可以可靠调配公开用户的生图配额，包括普通额度、无限额度、用户状态和批量用户管理。
- [ ] 管理员可以安全、准确地配置生图模型、端点池、并发、成本、售价、邀请码奖励和公开配置。
- [ ] 用户可以在公开服务场景下稳定完成登录、注册、邀请码使用、提交任务、查看结果、复用配置和管理历史图片。
- [ ] 生图任务在并发、失败切换、SSE 更新、轮询降级、图片保存和账单记录上具备可运营的可靠性。
- [ ] 公开上线风险得到处理：默认密钥、CORS、JWT 泄漏、配置文件权限、输入校验和敏感配置暴露不应阻碍上线。
- [ ] 运营者可以追踪成本、收入、用户用量、端点表现、失败原因和用户反馈，用于长期维护。
- [ ] 关键路径具备足够测试覆盖，避免配额、鉴权、任务状态、计费和管理员配置在迭代中回归。
- [ ] 代码结构和维护性足以支撑持续迭代，尤其是大型 AdminDashboard、Zustand store、重复 API client 和前后端类型漂移。

### Out of Scope

- 原生移动 App — 当前以 Web/PWA 公开服务为主，移动端原生体验不是本轮稳定运营的必要条件。
- 微服务化或云对象存储迁移 — 当前单体 Go 后端、SQLite 和本地文件存储足以支撑本轮继续完善；架构迁移另开阶段评估。
- 自动支付、订阅和完整商业化闭环 — 本轮关注配额、成本和运营控制；支付系统需要额外风控、合规和财务流程。
- 多租户企业权限体系 — 当前管理员/普通用户角色已满足基础公开服务运营，复杂组织、团队和 RBAC 暂不纳入本轮。
- 大规模增长营销系统 — 邀请码、公告和反馈已覆盖基础运营；增长自动化、推荐系统和 CRM 集成暂不纳入。

## Context

- 这是一个 brownfield 项目，已有代码和 `.planning/codebase/` 映射文档。GSD 初始化应从现有代码反推能力，而不是假设绿地项目。
- 技术栈是 React 19 + Vite 6 + TypeScript 前端，Go + Gin + GORM + SQLite 后端，OpenAI-compatible 图片 API 集成。
- 管理员能力是项目核心之一：用户明确强调“管理员可以调配用户的生图配额和生图设置”。
- 目标使用场景是公开服务，不只是个人内部工具。因此稳定性、安全性、配额准确性、端点配置、成本追踪和用户体验都属于核心范围。
- 当前项目已有 codebase map 指出若干上线风险和维护风险，包括默认密钥、CORS、JWT query token、输入绑定错误被忽略、配置文件权限、AdminDashboard 和 store 过大、SSE/图片缓存测试缺口等。

## Constraints

- **Brownfield**: 必须尊重现有 React/Go/SQLite/本地文件存储架构，优先增量完善而不是重写。
- **公开服务**: 鉴权、配额、配置、任务记录和成本分析必须按公开用户场景考虑，而不能只按本地个人工具标准处理。
- **成本控制**: 生图 API 调用有真实成本，端点并发、失败重试、售价和账单记录需要保持准确。
- **本地持久化**: 当前依赖 SQLite 和本地文件系统，任何数据清理、迁移或并发优化都要避免破坏已有用户数据。
- **管理员安全**: 管理端可以修改端点、配额、价格和用户状态，必须优先避免默认密钥、越权和敏感配置泄漏。
- **持续迭代**: 后续工作要可分 phase 推进，每个 phase 应覆盖可验证的用户或管理员能力。

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 将项目按 brownfield 初始化 | 代码库已经有完整前后端和 codebase map，需求应从现状反推 | — Pending |
| 本轮重点选择“继续完善” | 用户希望基于已有功能继续打磨，而不是重建核心产品 | — Pending |
| 主要使用者按“公开服务”规划 | 用户选择公开服务，需要比内部工具更重视稳定运营和安全上线 | — Pending |
| 完成标准是“能稳定运营” | 用户明确选择稳定运营作为完成判断，路线图应优先覆盖可靠性、配额、成本和管理能力 | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-24 after initialization*
