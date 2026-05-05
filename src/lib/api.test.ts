import { describe, expect, it } from 'vitest'
import { normalizeBaseUrl } from './api'

describe('normalizeBaseUrl', () => {
  it('trims trailing slashes', () => {
    expect(normalizeBaseUrl('https://api.example.com/')).toBe('https://api.example.com')
  })

  it('returns empty string as-is', () => {
    expect(normalizeBaseUrl('')).toBe('')
  })
})
