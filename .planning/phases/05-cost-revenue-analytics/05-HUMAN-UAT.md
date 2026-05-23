---
status: partial
phase: 05-cost-revenue-analytics
source: [05-VERIFICATION.md]
started: "2026-05-23T10:55:00+08:00"
updated: "2026-05-23T10:55:00+08:00"
---

## Current Test

[awaiting human testing]

## Tests

### 1. 系统配置页定价输入 UI
expected: 端点成本价和全局售价输入框正确渲染，校验工作正常，原子保存成功并显示 toast
result: [pending]

### 2. 成本收益统计 tab 视觉与数据
expected: 统计 tab 按 UI-SPEC 合同布局显示，KPI 数字格式正确，SVG 趋势图颜色区分清晰，拆分表按利润降序排列
result: [pending]

### 3. 端到端记账流程
expected: 成功保存的图片数量 = billing_records 行数，每行包含 task_id、user_id、user_label_snapshot、endpoint_base_url_snapshot、unit_cost_x10000、unit_sale_x10000 等快照字段
result: [pending]

## Summary

total: 3
passed: 0
issues: 0
pending: 3
skipped: 0
blocked: 0

## Gaps
