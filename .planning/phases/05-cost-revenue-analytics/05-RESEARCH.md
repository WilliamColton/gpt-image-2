<user_constraints>
## User Constraints (from CONTEXT.md) [CITED: .planning/phases/05-cost-revenue-analytics/05-CONTEXT.md]

### Locked Decisions

### 记账口径
- **D-01:** 一条任务成功生成多张图片时，只按最终成功保存到本地并进入 `outputImages` 的图片张数入账。
- **D-02:** 部分成功时，成功保存几张就记几张；失败或未保存成功的图片不产生收入和成本。
- **D-03:** failover 过程中，成本只归属最终成功生成图片的端点；失败端点不计成本。

### 定价配置
- **D-04:** 每个 API 端点配置当前成本价，单位元/张。
- **D-05:** 管理后台系统配置中新增全局售价，单位元/张。
- **D-06:** 每条成功记账记录必须保存当时的成本价和售价快照；后续修改端点成本价或全局售价不得改变历史统计。
- **D-07:** 成本价和售价输入最多支持 4 位小数，规划实现时应避免浮点金额误差（例如用整数最小单位或等价方式持久化）。

### 图表范围
- **D-08:** 新增管理后台图表页面，第一版同时包含核心总览、趋势、端点拆分和用户拆分。
- **D-09:** 核心指标至少包括总收入、总成本、利润、成功图片数。
- **D-10:** 时间筛选支持今日、7 天、30 天、全部。
- **D-11:** 第一版只展示图表和表格，不做 CSV 导出，不做点击钻取账单明细。

### 数据保留
- **D-12:** 使用单独账单/用量统计表记录成本收益明细，而不是只扩展任务表或只存汇总表。
- **D-13:** 每条账单记录至少需要保留 taskId、userId、用户可读标识快照、成功端点标识/URL、成功图片张数、成本价快照、售价快照、成本、收入、利润、创建时间。
- **D-14:** 管理员删除任务或用户时，历史成本收益统计保留，不随任务/用户删除而丢失。

### Claude's Discretion
- 图表库或是否用现有组件手写简单图表由 planner/researcher 决定，但必须符合现有 React + Tailwind + admin UI 风格。
- 管理 API 的具体路由命名、统计查询 DTO、数据库索引设计由 planner 决定。

### Deferred Ideas (OUT OF SCOPE)
- CSV/PDF 导出。
- 点击图表或表格钻取账单明细。
- 按用户或套餐设置不同售价。
- 支付、充值、发票或用户端价格展示。
- 失败端点尝试成本记账。
</user_constraints>

<phase_requirements>
## Phase Requirements [CITED: .planning/ROADMAP.md]

| ID | Description | Research Support |
|----|-------------|------------------|
| COST-01 | 端点成本价配置 | `backend-go/config/config.go` 的 `ApiEndpoint`、`backend-go/handler/admin.go` 的端点保存、`src/admin/AdminDashboard.tsx` 的系统配置页 |
| COST-02 | 全局售价配置 | `backend-go/config/config.go` 的全局 config、`backend-go/handler/admin.go` 的 admin 配置写入、`src/admin/AdminDashboard.tsx` 的系统配置页 |
| COST-03 | 按成功保存的图片记账 | `backend-go/handler/generate.go` 的成功路径、`backend-go/service/openai.go` 的 endpoint 归属、`backend-go/service/image.go` 的保存结果 |
| COST-04 | 成本收益统计页与图表 | `src/admin/AdminDashboard.tsx` 的新 tab、`src/admin/adminApi.ts` 的统计接口、`backend-go/database/models.go` 的账单表、`backend-go/service/*.go` 的聚合查询 |
</phase_requirements>

# Phase 05：成本与收益统计 - Research

**Researched:** 2026-05-22
**Confidence:** HIGH

## 1. Existing patterns to reuse
- 后端已经是 Go + Gin + GORM + SQLite 的单体结构，`database.Init()` 用 `AutoMigrate` 扩表，新增账单表可以沿用同一套路，不需要手写 SQL 迁移 [VERIFIED: backend-go/go.mod, backend-go/database/database.go].
- 运行时可变配置已经通过 `backend-go/config/config.go` 读写 `config.json`，并且 admin 端已经用 `PUT /api/admin/config/endpoints` 原子保存端点数组；全局售价和端点成本最适合继续走同一类配置持久化 [VERIFIED: backend-go/config/config.go, backend-go/handler/admin.go].
- `GenerateImage` 的成功路径已经是“保存图片 → 增加 `used_count` → 更新 task”，成本收益记账应该挂在同一成功路径上，而不是挂在 task 创建或失败分支上 [VERIFIED: backend-go/handler/generate.go].
- `withFailover` 能看到最终成功的 `ep.BaseURL`，但目前只返回 `ImageGenResult`，端点归属没有被带回调用方；而 `queue.go` 也已经把 `baseURL` 当作端点稳定 key [VERIFIED: backend-go/service/openai.go, backend-go/service/queue.go].
- `src/admin/AdminDashboard.tsx` 已经是单页 tabs 壳层，`src/admin/adminApi.ts` 已经集中管理 admin JWT 请求；新增“成本收益统计”更适合作为同壳层新 tab + 新 adminApi 方法，而不是新路由页 [VERIFIED: src/admin/AdminDashboard.tsx, src/admin/adminApi.ts].
- `src/types.ts` 是普通前端共享类型，admin 专用 DTO 目前更适合继续放在 `src/admin/adminApi.ts` 旁边，避免污染用户端类型层 [VERIFIED: src/types.ts, src/admin/adminApi.ts].
- 前端依赖里没有现成图表库；如果不想新增依赖，custom SVG/Canvas 图表最符合当前栈和包审查成本 [VERIFIED: package.json].

## 2. Implementation options / tradeoffs
1. 金额表示
   - `int64` 固定点（`x10000`）放在 DB、config、API 返回里，前端只负责格式化。优点是完全避免浮点误差，最符合 D-07；缺点是前后端都要有统一换算 helper。
   - 字符串十进制放在 API 边界，后端再解析成固定点。优点是表单更直观；缺点是解析/校验代码更多，任何 `parseFloat` 都会破坏精度纪律 [CITED: .planning/phases/05-cost-revenue-analytics/05-CONTEXT.md].

2. 账单粒度
   - 按“每个成功保存的图片事件”写 1 条账单，`successImageCount = 1`。优点是和当前 `used_count`、部分成功、混合 endpoint failover、图片去重都天然兼容；缺点是行数更多，但换来最小的归属歧义。
   - 按 task+endpoint 分组写账单。优点是行数更少；缺点是要先按 endpoint 聚合成功图片，遇到并发多图/部分保存失败时更容易写错归属。

3. 统计 API 形状
   - 分成 `summary` / `trend` / `endpoint-breakdown` / `user-breakdown` 四组接口。优点是 UI 能按卡片独立加载、独立失败重试，符合“某一块失败不清空整页”的 UI 要求；缺点是前端请求数更多。
   - 做一个大接口一次返回全部数据。优点是 snapshot 一致、前端调用少；缺点是任何一块 SQL 出错都会拖垮整页，和本 phase 的 partial failure 目标不一致。

4. 图表实现
   - 先用自定义轻量图表（SVG/Canvas/简单折线柱状）。优点是零新依赖、和现有 React + Tailwind 风格一致；缺点是 tooltip/axis/空态都要自己写。
   - 引入第三方 chart 包。优点是交互更完整；缺点是会新增外部依赖和包合法性审查成本，不符合“只做第一版统计”的范围感。

### 推荐的统计响应形状
```json
{
  "meta": { "range": "7d", "from": 1710000000000, "to": 1710600000000, "moneyScale": 10000 },
  "summary": {
    "revenueX10000": 123400,
    "costX10000": 45600,
    "profitX10000": 77800,
    "successImages": 321
  },
  "trend": [
    { "bucket": "2026-05-22", "revenueX10000": 1000, "costX10000": 400, "profitX10000": 600, "successImages": 8 }
  ],
  "endpointRows": [],
  "userRows": []
}
```
- `trend`、`endpointRows`、`userRows` 都应该是预聚合后的长数组；前端直接渲染图表和表格，不要再在浏览器里做财务聚合 [CITED: .planning/phases/05-cost-revenue-analytics/05-UI-SPEC.md].

## 3. Recommended plan constraints
- 新增单独账单表，不要把历史利润字段塞进 `tasks` 表；账单表必须保存 `taskId`、`userId`、`userLabelSnapshot`、`endpointBaseURLSnapshot`、`successImageCount`、`unitCostX10000`、`unitSaleX10000`、`costX10000`、`revenueX10000`、`profitX10000`、`createdAt` 这类快照字段 [CITED: .planning/phases/05-cost-revenue-analytics/05-CONTEXT.md].
- 不要给账单表挂会级联删除的外键；历史统计必须独立于 `tasks` / `users` / 当前 endpoint 配置存在 [CITED: .planning/phases/05-cost-revenue-analytics/05-CONTEXT.md].
- 成本收益记账要按“实际成功保存的图片事件”执行，不要按请求 `n`、不要按 API 返回结果数、也不要按唯一图片内容计数；`SaveImageBuffer` 会按内容 SHA-256 复用图片 ID，所以 billing 不能把 image ID 当唯一事件键 [VERIFIED: backend-go/service/image.go].
- 现有 `buildPerImageMetadata` 是按 index 把 `outputIDs` 和 `result.Images` 配对；一旦中间有图片保存失败，后面的 metadata 会错位。Phase 5 要把“保存成功记录”改成带完整上下文的 success slice，再由这份 success slice 写 billing [VERIFIED: backend-go/handler/generate.go].
- endpoint 归属必须随 `GeneratedImage` 一起从 `openai.go` 带到 `generate.go`，billing 只读这个快照，不要再反查当前 config 或依赖结果顺序 [VERIFIED: backend-go/service/openai.go, backend-go/handler/generate.go].
- 金额必须统一走固定点 helper，DB、config、统计 DTO、前端显示都不要出现 float64 金额；`moneyScale = 10000` 应该作为显式常量返回给前端 [CITED: .planning/phases/05-cost-revenue-analytics/05-CONTEXT.md].
- admin 写入必须原子化：一个保存动作同时提交端点成本和全局售价，UI 端也要一次性提交整份价格配置，避免“半保存”导致历史快照不一致 [CITED: .planning/phases/05-cost-revenue-analytics/05-UI-SPEC.md].
- 分析查询不要 join 当前 users/config 去算历史总额；要直接按 billing snapshot 聚合，summary/trend/endpoint/user 四块都共享同一个 range 解析逻辑 [CITED: .planning/phases/05-cost-revenue-analytics/05-CONTEXT.md, .planning/phases/05-cost-revenue-analytics/05-UI-SPEC.md].
- 统计页要放回现有 `/admin` tab 壳层里，使用现成 Card/Table/Tabs/Input/Badge 风格，不要新开独立全屏页面 [CITED: .planning/phases/05-cost-revenue-analytics/05-UI-SPEC.md, src/admin/AdminDashboard.tsx].
- 这版不引入新的 chart 依赖，先用自定义轻量图表；只有在 planner 明确需要更强交互时才考虑新包 [VERIFIED: package.json, .planning/phases/05-cost-revenue-analytics/05-UI-SPEC.md].
- 价格字段的后端校验要和前端一致：非空、非负、最多 4 位小数；旧 config.json 里缺少新字段时应按 0 兼容加载，不做破坏性迁移 [CITED: .planning/phases/05-cost-revenue-analytics/05-CONTEXT.md, VERIFIED: backend-go/config/config.go].

## 4. Open questions for planner
- 无；现有约束已经足够把数据模型、记账流、endpoint 归属、admin API 和统计查询形状锁定到可实施范围。
