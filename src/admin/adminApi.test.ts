import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  adminGetPricingConfig,
  adminUpdatePricingConfig,
  clearAdminToken,
  type ApiEndpoint,
  type PricingConfigResponse,
} from './adminApi'

describe('Task 1 — Pricing DTOs and client functions', () => {
  const TEST_TOKEN = 'test-admin-jwt'

  beforeEach(() => {
    // Provide localStorage in Node environment
    const store: Record<string, string> = {}
    vi.stubGlobal('localStorage', {
      getItem: vi.fn((key: string) => store[key] ?? null),
      setItem: vi.fn((key: string, val: string) => { store[key] = val }),
      removeItem: vi.fn((key: string) => { delete store[key] }),
    })
    // Seed with admin token so adminApi functions can find it
    ;(localStorage.setItem as ReturnType<typeof vi.fn>)('gpt-image-playground-admin-token', TEST_TOKEN)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  describe('ApiEndpoint type', () => {
    it('allows costPerImageX10000 as optional number', () => {
      const ep: ApiEndpoint = {
        baseUrl: 'https://api.openai.com',
        apiKey: 'sk-test',
        costPerImageX10000: 1234,
      }
      expect(ep.costPerImageX10000).toBe(1234)

      const withoutCost: ApiEndpoint = {
        baseUrl: 'https://api.openai.com',
        apiKey: 'sk-test',
      }
      expect(withoutCost.costPerImageX10000).toBeUndefined()
    })
  })

  describe('PricingConfigResponse type', () => {
    it('has endpoints, salePriceX10000, and moneyScale', () => {
      const resp: PricingConfigResponse = {
        endpoints: [
          { baseUrl: 'https://api.openai.com', apiKey: 'sk-test', costPerImageX10000: 500 },
        ],
        salePriceX10000: 2000,
        moneyScale: 10000,
        ok: true,
      }
      expect(resp.endpoints).toHaveLength(1)
      expect(resp.salePriceX10000).toBe(2000)
      expect(resp.moneyScale).toBe(10000)
      expect(resp.ok).toBe(true)
    })
  })

  describe('adminGetPricingConfig', () => {
    it('calls GET /api/admin/config/pricing with admin token', async () => {
      const mockResponse: PricingConfigResponse = {
        endpoints: [
          { baseUrl: 'https://api.openai.com', apiKey: 'sk-test', costPerImageX10000: 500 },
        ],
        salePriceX10000: 2000,
        moneyScale: 10000,
      }

      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      )

      const result = await adminGetPricingConfig()

      expect(globalThis.fetch).toHaveBeenCalledOnce()
      const [url, init] = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]
      expect(url).toContain('/api/admin/config/pricing')
      expect(init.method || 'GET').toBe('GET')
      expect(init.headers).toBeDefined()
      // Verify the Authorization header carries the token
      const headers = init.headers
      const authHeader = headers instanceof Headers ? headers.get('Authorization') : headers['Authorization']
      expect(authHeader).toBe(`Bearer ${TEST_TOKEN}`)
      expect(result).toEqual(mockResponse)
    })
  })

  describe('adminUpdatePricingConfig', () => {
    it('sends PUT with endpoints and salePriceX10000 in body', async () => {
      const endpoints: ApiEndpoint[] = [
        { baseUrl: 'https://api.openai.com', apiKey: 'sk-test', costPerImageX10000: 500 },
      ]
      const salePriceX10000 = 2000

      const mockResponse: PricingConfigResponse = {
        endpoints,
        salePriceX10000,
        moneyScale: 10000,
        ok: true,
      }

      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      )

      const result = await adminUpdatePricingConfig(endpoints, salePriceX10000)

      expect(globalThis.fetch).toHaveBeenCalledOnce()
      const [url, init] = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]
      expect(url).toContain('/api/admin/config/pricing')
      expect(init.method).toBe('PUT')
      const body = JSON.parse(init.body as string)
      expect(body.endpoints).toEqual(endpoints)
      expect(body.salePriceX10000).toBe(2000)
      expect(result).toEqual(mockResponse)
    })

    it('includes Authorization header in PUT request', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
        new Response(JSON.stringify({ endpoints: [], salePriceX10000: 0, moneyScale: 10000 }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      )

      await adminUpdatePricingConfig([], 0)

      const [, init] = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]
      const headers = init.headers
      const authHeader = headers instanceof Headers ? headers.get('Authorization') : headers['Authorization']
      expect(authHeader).toBe(`Bearer ${TEST_TOKEN}`)
    })
  })
})
