---
status: issues_found
phase: "01"
phase_name: "检查系统 bug"
depth: standard
files_reviewed: 114
findings:
  critical: 11
  warning: 5
  info: 0
  total: 16
---

# Code Review: Phase 01 — 检查系统 bug

审查范围来自阶段 1 的 git diff 回退范围，共 114 个现存文件。首次全量 reviewer 未产出报告，因此本报告基于后端认证、后端任务/计费、前端状态、前端管理四组标准深度审查结果，并抽查了关键源码位置。

## Critical Findings

### CR-01 — 客户端可篡改任务生命周期字段并绕过配额

- **位置:** `backend-go/handler/tasks.go:26`
- **问题:** `TasksUpdate` 接受任意请求体并允许认证用户写入 `status`、`outputImages`、`error`、`finishedAt`、`actualParams` 等任务生命周期字段；任务不存在时还会从客户端 payload 创建新任务。
- **影响:** 用户可以把 `queued/running` 任务改成 `done/error`，使配额检查不再统计该任务，然后继续提交更多生成任务；也可以伪造完成任务和输出图元数据，破坏任务、计费和配额一致性。
- **建议:** 对外任务更新接口只允许修改明确的用户字段，例如 `isFavorite`；禁止 PUT 创建缺失任务。任务状态、输出图、错误、完成时间和实际参数应只由后端生成流程内部服务写入，并校验合法状态流转。

### CR-02 — 生成成功后的计费、扣量和任务完成状态不是原子操作

- **位置:** `backend-go/handler/generate.go:168`
- **问题:** billing 写入、`used_count` 增量、任务标记 `done` 是三段独立数据库操作；任一步失败仅记录日志后继续。
- **影响:** 可能出现任务已完成但无账单或未扣配额，也可能账单/配额已写入但任务仍停留在 `running` 或缺少输出图，导致收入统计、用户配额和任务状态不一致。
- **建议:** 增加事务化的 `FinalizeSuccessfulTask` 服务，在同一个 DB transaction 内写 billing、更新用户 `used_count`、更新任务最终状态；任一步失败时整体回滚并进入可重试或可恢复路径。

### CR-03 — 兑换码登录并发可创建多个有效账号

- **位置:** `backend-go/service/auth.go:144`
- **问题:** `LoginWithCode` 先创建新用户，再用 `WHERE used_by IS NULL` 标记兑换码；如果并发请求同时使用同一未用兑换码，失败的一方只记录 warning，但仍签发 token 并返回新用户。
- **影响:** 单个兑换码可并发生成多个有效账号和配额，造成配额滥发、用户数据污染和运营成本异常。
- **建议:** 将“占用兑换码、创建用户、签发前状态确认”放入同一个数据库事务；`RowsAffected == 0` 必须回滚并返回“兑换码已被使用”。

### CR-04 — 既有用户兑换码消费和配额增加不是原子操作

- **位置:** `backend-go/service/auth.go:177`
- **问题:** `RedeemForUser` 先把兑换码标记为已使用，再更新用户 quota，且没有事务；用户 quota 更新失败或用户并发删除时，兑换码可能已消费但配额未到账。
- **影响:** 用户会丢失兑换码且无法自动恢复，产生数据一致性问题和客服处理成本。
- **建议:** 使用事务包裹兑换码占用和 quota 更新；更新用户 quota 后检查 `RowsAffected`，失败时回滚兑换码占用。

### CR-05 — 管理员 JWT 只信任 role 且默认密钥可预测

- **位置:** `backend-go/middleware/middleware.go:60`, `backend-go/config/config.go:83`
- **问题:** `AdminMiddleware` 只校验 JWT 中的 `role=admin`，不按 `sub` 查询管理员用户是否存在、是否 active；同时缺少 `config.json` 时会继续使用默认 `JWTSecret: "change-me"` 和 `AdminApikey: "change-me-admin-apikey"`。
- **影响:** 默认配置部署会让攻击者可预测密钥并伪造 admin token；已泄露或已签发的 admin token 也无法通过禁用/删除管理员账号吊销。
- **建议:** 启动时拒绝默认 JWT/admin key；`AdminMiddleware` 应按 `sub` 查询管理员用户并校验 `role == "admin"`、`status == "active"`，必要时支持 token 版本或吊销机制。

### CR-06 — 主 JWT 被放入图片 URL 查询参数

- **位置:** `src/lib/backendApi.ts:201`
- **问题:** `getImageUrl` 直接把用户 JWT 放进 `?token=` 查询参数生成图片 URL。
- **影响:** token 会出现在 DOM、浏览器历史、代理/服务端访问日志、复制的图片链接和诊断信息中；一旦泄露，攻击者可在 token 有效期内访问用户账号/API。
- **建议:** 不要生成带主 JWT 的图片 URL。优先用带 `Authorization` header 的 `fetch('/api/images/:id')` 拉取 Blob/Data URL 后赋给 `<img>`；如必须直链，应使用短期、单图片作用域的签名 URL。

### CR-07 — SSE 正常 EOF 不触发轮询 fallback

- **位置:** `backend-go/handler/tasks.go:166`, `src/lib/backendApi.ts:269`
- **问题:** 后端 SSE 连接 10 分钟后主动断开，但 OpenAI 调用可运行更久；前端 `streamTaskStatus` 在 `reader.read()` 返回 `done` 时直接退出，不调用 `onError`，store 不会启动 polling fallback。
- **影响:** 长任务或代理/网络中间断流后，前端任务会永久停在 `queued/running`，用户看不到实际完成/失败状态，并可能重复提交任务造成额外成本。
- **建议:** 让 SSE 超时时间不短于后端生成超时，或服务端在非终态断开前发送明确事件；前端记录是否收到 `done/error`，若流结束但未终态，应触发轮询或重连。

### CR-08 — 图片上传信任 multipart Content-Type 并同源回传

- **位置:** `backend-go/handler/images.go:31`, `backend-go/handler/images.go:62`
- **问题:** 图片上传完全信任 multipart `Content-Type`，下载时把持久化的 MIME 原样作为响应 `Content-Type` 返回。
- **影响:** 认证用户可上传 HTML/SVG 等主动内容，并通过同源 `/api/images/:id?token=...` URL 执行脚本，形成存储型 XSS，威胁用户 token 和 admin token。
- **建议:** 用文件魔数检测真实类型，只允许 `image/png`、`image/jpeg`、`image/webp` 等安全栅格格式，拒绝 `text/html` 和 `image/svg+xml`；响应增加 `X-Content-Type-Options: nosniff`，必要时对非白名单强制 attachment。

### CR-09 — 输入图片上传失败会留下永久 queued 本地任务

- **位置:** `src/store.ts:657`
- **问题:** 普通输入图片上传循环没有 `try/catch`；`uploadImage()` 任一失败会让 `submitTask()` rejected，但本地任务已经插入为 `queued`。
- **影响:** 上传失败时任务不会被标记为 `error`，也没有可靠 toast/详情错误；用户会看到一个永远排队的本地任务，并可能触发未处理 Promise rejection。
- **建议:** 像遮罩上传一样包住输入图片上传流程；失败时调用本地任务更新，将任务标记为 `error` 并填充 `finishedAt`、`elapsed`、错误消息，然后停止后续 `executeTask()`。

### CR-10 — 端点删除/过滤后成本草稿按旧下标错配

- **位置:** `src/admin/AdminDashboard.tsx:408`, `src/admin/AdminDashboard.tsx:470`
- **问题:** 删除端点或过滤空 `baseUrl` 后，`costInputDrafts` 仍按旧数组下标保存；保存时用过滤后的 `valid.map((ep, i) => costInputDrafts[i])` 取成本。
- **影响:** 管理员保存 pricing 后，端点成本可能被写成别的端点成本或 `0`，污染成本、利润统计与后续计费归因。
- **建议:** 删除端点时同步重建/重排 `costInputDrafts`；保存时保留原始下标，例如先构造 `{ ep, originalIndex }` 后再过滤，并使用 `costInputDrafts[originalIndex]`。

### CR-11 — 图片数量输入可绕过 1–4 上限

- **位置:** `src/components/InputBar.tsx:149`, `src/components/InputBar.tsx:461`
- **问题:** 数量输入虽然设置了 `min={1}`、`max={4}`，但 `commitN()` 直接 `Number(nInput)` 后写入 `params.n`，未 clamp 到 1–4，也允许小数/超大值进入提交参数。
- **影响:** 用户或脚本可提交远超 UI 预期的图片数量；在多图路径会触发大量请求，造成额度、成本和后端资源异常消耗。
- **建议:** 在 `commitN()` 和提交前强制整数化并限制 `1 <= n <= 4`；后端也必须在 HTTP 边界校验 `params.n`，不能只依赖 UI。

## Warning Findings

### WR-01 — 非 legacy 账号可调用迁移接口覆盖用户名和密码

- **位置:** `backend-go/handler/auth.go:123`, `backend-go/service/auth.go:518`
- **问题:** `/api/auth/migrate` 只要求已登录 token，没有校验当前账号是否真的是 legacy 账号；`MigrateUser` 可覆盖任意已登录用户的 `username` 和 `password_hash`。
- **影响:** 已有密码的普通用户可绕过 `/change-password` 的旧密码校验直接改密码；若 token 泄露，攻击者可永久接管账号。
- **建议:** 在 `MigrateUser` 内读取用户并拒绝 `PasswordHash != nil` 的账号迁移；非 legacy 账号改密码必须走 `ChangePassword` 并校验旧密码。

### WR-02 — bootstrapping 把非认证错误当作登录失效

- **位置:** `src/store.ts:443`, `src/store.ts:532`
- **问题:** `bootstrapBackendSession()` 用 `Promise.all([getMe(), fetchTasks(), getPublicConfig()])`，任何一个请求失败都会被 `initStore()` 当作登录失效处理并清空 token、authUser 和 tasks。
- **影响:** 公共配置临时失败、任务列表偶发网络错误、后端短暂不可达都会把有效用户强制登出，并清空当前前端任务视图。
- **建议:** 将 `getMe()` 与非认证依赖拆开处理；只有明确 401/登录失效时才 `clearBackendToken()`。`getPublicConfig()` 应继续可选降级，`fetchTasks()` 失败应保留 session 并提示或重试。

### WR-03 — 遮罩主图关联未持久化到后端任务

- **位置:** `src/lib/backendApi.ts:234`, `src/store.ts:625`, `src/store.ts:669`
- **问题:** 遮罩编辑任务没有把 `maskTargetImageId` 发送到后端；输入图上传后也没有把本地遮罩目标重映射为上传后的后端图片 ID。
- **影响:** SSE/刷新后远程任务会丢失遮罩主图关联，详情页无法稳定展示遮罩预览，历史记录中的遮罩语义依赖“第一张输入图”的隐式猜测。
- **建议:** `submitEditTask()` 请求体增加 `maskTargetImageId`；上传完成后把遮罩目标更新为对应后端图片 ID，并确保后端任务记录持久化该字段。

### WR-04 — 管理端非 JSON 错误响应会丢失原始错误信息

- **位置:** `src/admin/adminApi.ts:60`
- **问题:** 错误处理先 `await response.json()`，失败后再 `await response.text()`；Fetch body 已被消费，非 JSON 错误响应会变成二次读取异常。
- **影响:** 代理错误、HTML/纯文本 502/401 等场景下，管理端会丢失真实错误信息，排障和用户反馈都不可靠。
- **建议:** 先读取一次文本再尝试 `JSON.parse`，或使用 `response.clone().json()` 后 fallback 到原响应文本。

### WR-05 — 邀请配置更新缺省 inviteEnabled 会静默关闭邀请

- **位置:** `backend-go/handler/admin.go:409`
- **问题:** `AdminUpdateInviteConfig` 使用非指针 `bool InviteEnabled`，请求体缺少 `inviteEnabled` 时默认为 `false`，导致只更新奖励/默认配额时可能意外关闭邀请功能。
- **影响:** 旧版客户端或手工 API 调用只更新数值配置，会静默关闭注册邀请流程，造成线上功能异常。
- **建议:** 将请求字段改为 `*bool`，缺省时保留当前 `config.IsInviteEnabled()`；或要求请求必须显式包含该字段并在缺失时返回 400。

## Notes

- 本报告聚焦高置信问题。由于阶段 1 git diff 回退范围达到 114 个文件，建议优先修复 Critical 项后对修复范围重新运行 `/gsd-code-review 1 --files=...` 或 `/gsd-code-review 1 --fix`。
- 多个问题互相关联：任务 lifecycle 写权限、SSE 断流 fallback、成功完成事务化、配额计数需要作为同一条任务可靠性链路统一修复。
