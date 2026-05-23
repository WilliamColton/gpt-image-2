# Phase 06: 账号密码 & 邀请码 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-23
**Phase:** 06-invite-code
**Areas discussed:** 登录方式 (密码认证), 邀请码与兑换码, 奖励规则配置, 前端UI与流程, 管理后台扩展

---

## 登录方式 (密码认证)

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 登录方式共存策略 | 兑换码或密码二选一 | 两种方式都支持 | |
| | 仅密码登录（过渡方案） | 兑换码过渡期可用，最终只有密码登录 | ✓ |
| | 统一入口 | 两个 tab/切换 | |
| | 补充说明 | 默认兑换码 tab 登录，但新用户必须密码 | ✓ (user clarification) |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 密码认证方式 | bcrypt 哈希存储 | 加盐慢哈希，安全 | ✓ |
| | HMAC 签名 | 简单但不如 bcrypt | |
| | 明文存储 | 极不安全 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 登录接口设计 | 新增独立接口 | POST /api/auth/login-password | ✓ |
| | 复用现有接口加 mode | 一个接口两种方式 | |
| | 替换现有接口 | 去掉兑换码登录 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 用户名规则 | 字母数字下划线唯一 | 2-30 字符，严格 | |
| | 宽松规则唯一 | 3-20 字符，允许中文 | ✓ |
| | 仅显示名 | 不用于登录 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 密码规则 | 最少 8 字符 | 简单够用 | ✓ |
| | 中复杂度 | 大小写+数字 | |
| | 高复杂度 | 12 位+特殊字符 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 密码管理 | 设置页改密码 | 旧密码确认 | ✓ |
| | 改密码+管理员重置 | 管理员可设新密码 | ✓ (combined) |
| | 完整密码管理 | 邮箱找回等 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 老用户迁移 | 登录后弹迁移Modal | 不可关闭，必须完成 | ✓ |
| | 注册流程拦截 | 用兑换码走注册 | |
| | 禁止兑换码登录 | 硬迁移 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 注册接口 | 独立注册接口 | POST /api/auth/register | ✓ |
| | 复用 login 自动判断 | 后端自动路由 | |
| | 两步注册验证 | 先验证邀请码 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| Token 策略 | 保持 30 天 JWT | 无 refresh token | ✓ |
| | 7 天+refresh | 短 token+刷新 | |
| | 会话级 token | 关浏览器失效 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 管理员重置密码 | 直接设置新密码 | 输入新密码即生效 | ✓ |
| | 生成临时密码 | 展示临时密码 | |
| | 不可重置 | 用户自行负责 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 登录 UI | 用户名+密码为主 | 纯密码登录页 | |
| | 紧凑单栏 | 一个输入框 | |
| | Tab 切换登录方式 | 默认兑换码 Tab | ✓ |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 记住我 | 默认记住可选短期 | 可选项 | |
| | 始终记住 | localStorage 持久 | ✓ |
| | 始终不记住 | sessionStorage | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 登出行为 | 仅前端清 token | JWT 保持有效 | ✓ |
| | 后端黑名单 | 维护失效 token 表 | |
| | 按时间失效 | 记录登出时间 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 修改密码 UI | 旧密码+新密码+确认 | 三字段 | ✓ |
| | 仅新密码+确认 | 两字段 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 登录失败锁定 | 不做锁定 | bcrypt 缓解 | ✓ |
| | 5 次失败锁定 15 分钟 | 轻量保护 | |
| | 10 次失败+管理员解锁 | 重保护 | |

**Notes:** 用户明确选择了"仅密码登录"的过渡方案，但后续澄清登录 UI 用 Tab 切换且默认兑换码 Tab。这不是矛盾——过渡期内兑换码 Tab 仍然可用（老用户可继续用兑换码登录直到迁移），新用户必须通过注册设置密码。

---

## 邀请码与兑换码

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 用户生成邀请码 | 管理员配额+消耗 | 限制生成数量 | |
| | 用户无限制生成 | 任意自定义 | ✓ |
| | 仅管理员生成 | 管理员控制 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 奖励形式 | 双方配额奖励 | 邀请人和被邀请人都获得 | ✓ |
| | 仅新用户奖励 | 只有被邀请人获得 | |
| | 持续分成 | 后续生成量分成 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 邀请码格式 | 类兑换码随机串 | 8-16 位字母数字 | |
| | 用户专属链接 | URL 形式 | |
| | 用户自定义码 | 如 'WILLIAM2026' | ✓ |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 与兑换码关系 | 兑换码保持 | 纯充值配额 | ✓ |
| | 兑换码可当邀请码 | 多功能 | |
| | 全部合并 | 统一为邀请码 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 邀请码唯一性 | 全局唯一 | 先到先得 | ✓ |
| | 系统自动生成 | 不可改 | |
| | 允许重复 | 需额外输入邀请人用户 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 邀请码修改 | 随时可改 | 旧码立即失效 | ✓ |
| | 不可修改 | 设后不变 | |
| | 仅管理员可改 | 管理员控制 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 被邀请人奖励发放 | 注册时自动获得 | 直接到账 | ✓ |
| | 手动兑换 | 设置页操作 | |
| | 管理员手动发放 | 后台操作 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 邀请人奖励发放 | 注册时自动发放 | 被邀请人注册成功触发 | ✓ |
| | 条件触发 | 满足条件才发 | |
| | 管理员手动发放 | 后台操作 | |

---

## 奖励规则配置

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 配置位置 | 新邀请码设置 tab | 管理后台新增 | ✓ |
| | 嵌入系统配置 | 现有 tab 内 | |
| | 配置文件 | config.json | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 默认注册配额 | 无邀请码默认 0 | 不能不带邀请码注册 | |
| | 无邀请码有默认配额 | 管理员设定 | ✓ |
| | 需审核 | 审核开通 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 奖励区分 | 分别设置两个值 | 邀请人和被邀请人独立 | ✓ |
| | 统一一个值 | 双方相同 | |
| | 硬编码 | 代码中写死 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 后台可见性 | 查看所有邀请码使用 | 列表展示 | ✓ |
| | 不展示 | 不公开 | |
| | 可禁用邀请码 | 管理操作 | |

---

## 前端 UI 与流程

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 注册入口 | 登录页底部链接 | "没有邀请码？注册" | ✓ |
| | 旁边按钮 | 切换注册 | |
| | 首页两个入口 | 独立入口 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 注册表单字段 | 邀请码(可选)+用户名+密码 | 3 字段 | ✓ |
| | 加邮箱字段 | 4 字段 | |
| | 仅用户名+密码 | 2 字段 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 迁移 Modal | 用户名+密码 | 2 字段 | |
| | 加确认密码 | 3 字段 | ✓ |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 邀请码展示管理 | 邀请码+复制+修改 | 设置页 | ✓ |
| | 仅展示 | 不可操作 | |
| | 首页展示 | 醒目位置 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 设置页布局 | 独立区域分块展示 | 清晰分区 | ✓ |
| | 不分块 | 混在一起 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 迁移判断 | 返回 needsMigration 标志 | 登录接口返回 | ✓ |
| | 到主页再弹 | 延迟判断 | |
| | getMe 判断 | 获取用户信息判断 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 注册成功后 | 注册即自动登录 | 返回 token+user | ✓ |
| | 跳转登录页 | 手动登录 | |
| | 提示+切换表单 | 半自动 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 用户名展示 | 顶部显示用户名 | 替换 label | ✓ |
| | 保持不变 | 显示 label | |

---

## 管理后台扩展

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 密码重置 UI 位置 | 用户列表加按钮 | 操作列新增 | ✓ |
| | 新 tab 中管理 | 邀请码 tab | |
| | 无 UI | 无前端操作 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 邀请码 tab 内容 | 奖励配置 + 使用列表 | 双区域 | ✓ |
| | 仅配置 | 无列表 | |
| | 完整管理 | +筛选+导出 | |

| Question | Option | Description | Selected |
|----------|--------|-------------|----------|
| 新增管理 API | 新增独立路由 | PUT password, GET|PUT invite-config, GET invites | ✓ |
| | 扩展用户路由 | 在现有路由上扩展 | |

---

## Claude's Discretion

- bcrypt 库选择（推荐 `golang.org/x/crypto/bcrypt`）
- 数据库迁移方案：新增字段 password_hash, username, invite_code, invite_code_set_at
- Admin API DTO 结构设计
- 邀请奖励配置的存储方式（数据库新表或 config.json 扩展）
- 具体 UI 样式（遵循现有 glass morphism 设计系统，使用现有 UI 组件）
- 错误提示措辞
