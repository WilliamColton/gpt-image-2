# Phase 3: API 降级机制 - Context

**Gathered:** 2026-05-08
**Status:** Ready for planning

<domain>
## Phase Boundary

为后端添加 API 端点降级机制：配置多个 base URL + API Key 组合，请求失败时自动切换到下一个可用端点。

仅涉及后端 Go 代码，前端无需改动。

</domain>

<decisions>
## Implementation Decisions

### 端点配置结构
- `config.json` 新增 `apiEndpoints` 数组字段，每项包含 `baseUrl` 和 `apiKey`
- 示例: `"apiEndpoints": [{"baseUrl": "https://api.openai.com/v1", "apiKey": "sk-xxx"}, {"baseUrl": "https://api.hi-code.cc/v1", "apiKey": "sk-yyy"}]`
- `apiEndpoints` 必须至少配置一项，不再回退到 `defaults.baseUrl`

### 降级策略
- 即时切换：请求失败后立即尝试下一个端点，不等待
- 可触发切换的错误类型：网络连接错误、HTTP 429 (Too Many Requests)、HTTP 5xx 服务端错误
- 不可切换的错误：4xx 客户端错误（如 401 认证失败、400 参数错误），这些是用户问题而非端点问题
- 所有端点均失败时，返回最后一个遇到的错误

### 调用方 API Key 处理
- 当 `apiEndpoints` 中配置了 `apiKey` 时，使用端点自己的 key
- 当端点未配置 `apiKey`（空字符串）时，使用调用方传入的用户 apiKey

### Claude's Discretion
- 端点选择策略（顺序 vs 随机） — 顺序即可，简单可靠
- 是否记录失败端点便于排查 — 使用简单日志即可
- `defaults.baseUrl` 字段完全移除，不再保留

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 后端配置
- `backend-go/config/config.go` — 当前配置结构（Config、Defaults）
- `backend-go/config.json` — 实际配置文件

### OpenAI 调用层
- `backend-go/service/openai.go` — 当前 OpenAI 调用逻辑（newClient、CallImagesGenerations、CallImagesEdits）

</canonical_refs>

<specifics>
## Specific Ideas

当前 `newClient` 函数直接使用 `config.App.Defaults.BaseURL` 创建客户端：
```go
func newClient(apiKey string) openai.Client {
    return openai.NewClient(
        option.WithBaseURL(config.App.Defaults.BaseURL),
        option.WithAPIKey(apiKey),
        option.WithMaxRetries(0),
    )
}
```

需要将此函数改为接受端点参数，或在调用层实现重试/降级循环。

</specifics>

<deferred>
## Deferred Ideas

- 端点健康检查（定期 ping）
- 端点优先级/权重配置
- 失败计数和临时屏蔽机制

</deferred>

---

*Phase: 03-api-failover*
*Context gathered: 2026-05-08*
