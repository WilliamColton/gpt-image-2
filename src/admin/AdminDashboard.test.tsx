import { describe, expect, it } from 'vitest'
import source from './AdminDashboard.tsx?raw'

// Source-check tests — validate acceptance criteria strings are present in AdminDashboard.tsx
// Uses Vite's ?raw import to avoid Node.js fs/path that would fail tsc -b.

describe('Task 2 — Endpoint cost and global sale price controls', () => {
  it('contains endpoint cost label', () => {
    expect(source).toContain('成本价（元/张）')
  })

  it('contains global sale price label', () => {
    expect(source).toContain('全局售价（元/张）')
  })

  it('contains step="0.0001" for precise money input', () => {
    expect(source).toContain('step="0.0001"')
  })

  it('contains help text for 4 decimal places', () => {
    expect(source).toContain('支持 4 位小数')
  })

  it('contains inline error message for invalid price', () => {
    expect(source).toContain('请输入非负数字，最多 4 位小数')
  })

  it('does NOT use parseFloat anywhere', () => {
    expect(source).not.toContain('parseFloat')
  })
})

describe('Task 3 — Save all price fields atomically', () => {
  it('contains handleSavePricingConfig', () => {
    expect(source).toContain('handleSavePricingConfig')
  })

  it('calls adminUpdatePricingConfig', () => {
    expect(source).toContain('adminUpdatePricingConfig')
  })

  it('contains save CTA text', () => {
    expect(source).toContain('保存价格配置')
  })

  it('contains success toast text', () => {
    expect(source).toContain('价格配置已保存')
  })
})
