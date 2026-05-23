# Phase 09: 操作日志 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-23
**Phase:** 09-操作日志
**Areas discussed:** 日志范围, 存储与保留策略, 查看与过滤, 性能与级别控制

---

## 日志范围

| Option | Description | Selected |
|--------|-------------|----------|
| 仅关键业务事件 | 登录/生成/管理操作/配额事件，每条有业务意义 | ✓ |
| HTTP 请求 + 业务事件 | 在业务事件基础上增加所有 HTTP 请求日志 | |
| 仅管理员操作 | 只记录管理后台变更操作 | |

**User's choice:** 仅关键业务事件（推荐）

| Option | Description | Selected |
|--------|-------------|----------|
| 认证事件 | 登录/注册/迁移/改密（成功/失败） | ✓ |
| 生成事件 | 任务提交/成功/失败 | ✓ |
| 管理操作 | 配额/用户/配置变更 | ✓ |
| 配额事件 | 兑换/邀请奖励/已用递增 | ✓ |

**User's choice:** 全部四类事件

| Option | Description | Selected |
|--------|-------------|----------|
| 精简 | event_type, user_id, user_label, message, severity, created_at | |
| 标准 | 精简 + IP, details_json | ✓ |
| 最小化 | event_type, message, created_at | |

**User's choice:** 标准字段方案

---

## 存储与保留策略

| Option | Description | Selected |
|--------|-------------|----------|
| SQLite 新表 | audit_logs 表通过 GORM 写入 | ✓ |
| 独立日志文件 | 写文本/JSON 文件 | |
| stdout 重定向 | 不改现有代码 | |

**User's choice:** SQLite 新表（推荐）

| Option | Description | Selected |
|--------|-------------|----------|
| 手动清理 | 管理员在 Tab 点击按钮清理 N 天前或清空 | ✓ |
| 自动清理 | 定时/启动时自动删除 N 天前日志 | |
| 无限保留 | 不清理 | |

**User's choice:** 手动清理（推荐）

---

## 查看与过滤

| Option | Description | Selected |
|--------|-------------|----------|
| 管理后台 Tab | 新增"操作日志"Tab | ✓ |
| 独立页面 | /logs 独立页面 | |
| 仅 API | 无前端 UI | |

**User's choice:** 管理后台新 Tab（推荐）

| Option | Description | Selected |
|--------|-------------|----------|
| 事件类型筛选 | 下拉选择事件类型 | ✓ |
| 严重级别筛选 | INFO/WARN/ERROR 下拉 | ✓ |
| 用户搜索 | 文本搜索用户名 | ✓ |
| 关键词搜索 | message 内容搜索 | ✓ |
| 时间范围筛选 | 今日/7天/30天/全部 | ✓ |

**User's choice:** 全选

| Option | Description | Selected |
|--------|-------------|----------|
| 表格列表 | 与现有 admin Table 风格一致 | ✓ |
| 终端风格 | 纯文本行 | |
| 卡片式 | 每条一个卡片 | |

**User's choice:** 表格列表（推荐）

| Option | Description | Selected |
|--------|-------------|----------|
| 分页 | 后端 page/pageSize | |
| 一次加载 | 全量加载到前端 | ✓ |

**User's choice:** 一次性全部加载

---

## 性能与级别控制

| Option | Description | Selected |
|--------|-------------|----------|
| 全部级别入库 | INFO/WARN/ERROR 都存 | ✓ |
| 仅 WARN/ERROR | INFO 仅 stdout | |
| 仅 ERROR | | |

**User's choice:** 全部级别入库（推荐）

| Option | Description | Selected |
|--------|-------------|----------|
| 同步写入 | 事件触发时直接 INSERT | ✓ |
| 异步批量 | channel + goroutine 批量写 | |
| stdout 桥接 | 先 slog 再消费写入 | |

**User's choice:** 同步写入（推荐）

| Option | Description | Selected |
|--------|-------------|----------|
| 不做限流 | 配额制自然控制日志量 | ✓ |
| 按用户限流 | 每秒 N 条上限 | |

**User's choice:** 不做限流（推荐）

---

## Claude's Discretion

- audit_logs 表具体字段和 GORM 模型定义
- 各事件类型在 details_json 中的专属字段结构
- 管理后台日志 Tab 的具体 UI 布局和样式
- 后端 handler/service 分层和日志写入辅助函数签名
- 清理 API 返回的清理条数

## Deferred Ideas

- HTTP 请求日志入库（当前仅 stdout）
- 日志导出（CSV/PDF）
- 自动清理策略
- 实时日志流（SSE/WebSocket）
- 日志告警/通知
- 按用户差异化保留策略
