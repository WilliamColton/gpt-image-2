---
phase: 2
plan: 7
type: execute
wave: 2
depends_on:
  - 5
files_modified:
  - src/store.test.ts
  - src/lib/api.test.ts
autonomous: true
requirements:
  - FE-03
  - FE-04
  - FE-05
---

# Plan 07: 端到端测试验证

## Goal

验证前端改造后功能完整性：Images API / Responses API 模式切换正常、Codex CLI 兼容模式行为不变、任务卡片显示实时状态、页面刷新后状态不丢失。

## Tasks

### Task 1: 更新 store 测试 — submitTask 后端提交流程

**Read first:** `src/store.test.ts`

**Action:**
- 添加测试：`submitTask` 成功提交后调用 `submitGenerateTask`（而非 `callImageApi`）
- 添加测试：轮询逻辑在任务 status 变为 `done` 时正确更新任务
- 添加测试：轮询超时后任务 status 变为 `error`
- 添加测试：任务 status 为 `error` 时正确更新错误信息
- Mock `backendApi.submitGenerateTask` 和 `backendApi.getTasks`

**Acceptance criteria:**
- 测试验证 `submitGenerateTask` 被调用（而非 `callImageApi`）
- 测试验证轮询完成后任务状态更新为 `done`
- 测试验证超时后任务状态更新为 `error`
- 所有测试通过

---

### Task 2: 更新 api 测试 — 移除 OpenAI 直接调用测试

**Read first:** `src/lib/api.test.ts`

**Action:**
- 删除或更新针对 `callImageApi`, `callImagesApi`, `callResponsesImageApi` 的测试（这些函数已移除）
- 保留图片工具函数的测试（如 blob 转换、大小计算等）
- 如果有新的工具函数需要测试，添加相应测试用例

**Acceptance criteria:**
- `api.test.ts` 不包含对已删除函数的测试
- 保留的测试全部通过
- 无编译错误

---

### Task 3: 手动验证 — Images API 模式

**Read first:** 无

**Action:** 手动测试以下场景：
1. 确保后端 `OPENAI_API_KEY` 已配置
2. 启动后端 (`go run main.go`)
3. 启动前端 (`npm run dev`)
4. 使用 API Key 登录
5. 选择 Images API 模式
6. 输入提示词，点击生成
7. 观察任务卡片从 running 状态变为 done
8. 确认输出图片正确显示
9. 刷新页面，确认任务状态和图片不丢失

**Acceptance criteria:**
- 任务提交后立即显示 running 状态卡片
- 后端独立完成 API 调用（前端不直接请求 OpenAI）
- 任务完成后图片正确显示
- 页面刷新后任务和图片仍在

---

### Task 4: 手动验证 — Responses API 模式

**Read first:** 无

**Action:** 手动测试以下场景：
1. 切换到 Responses API 模式
2. 输入提示词，点击生成
3. 观察任务从 running 变为 done
4. 确认输出图片正确显示
5. 确认 `revisedPromptByImage` 数据正确（如果有改写）

**Acceptance criteria:**
- Responses API 模式下任务正常提交和完成
- 输出图片和元数据正确

---

### Task 5: 手动验证 — Codex CLI 兼容模式

**Read first:** 无

**Action:** 手动测试以下场景：
1. 开启 Codex CLI 模式
2. 设置 n > 1（多图生成）
3. 输入提示词，点击生成
4. 确认后端使用并发单图请求（查看后端日志）
5. 确认所有输出图片正确显示

**Acceptance criteria:**
- Codex CLI 模式下 quality 参数为 auto
- 多图生成使用并发请求（后端处理）
- 所有输出图片正确显示

---

### Task 6: 手动验证 — 编辑模式（带遮罩）

**Read first:** 无

**Action:** 手动测试以下场景：
1. 上传一张参考图
2. 打开遮罩编辑器，绘制遮罩
3. 输入提示词，点击生成
4. 确认后端正确接收 maskImageId
5. 确认输出图片为编辑结果

**Acceptance criteria:**
- 遮罩编辑后任务正确提交到 `/api/edit`
- maskImageId 正确传递给后端
- 输出图片为编辑后的结果

## must_haves

- Images API 和 Responses API 两种模式都能正常工作
- Codex CLI 兼容模式行为不变
- 任务状态实时更新（轮询机制）
- 页面刷新后状态不丢失
- 遮罩编辑功能正常
