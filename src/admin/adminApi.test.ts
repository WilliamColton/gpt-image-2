import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  adminGetBillingSummary,
  adminGetBillingTrend,
  adminGetBillingEndpointBreakdown,
  adminGetBillingUserBreakdown,
  adminGetPricingConfig,
  adminUpdatePricingConfig,
  clearAdminToken,
  type AnalyticsMeta,
  type AnalyticsRange,
  type ApiEndpoint,
  type BillingEndpointBreakdownResponse,
  type BillingEndpointRow,
  type BillingSummary,
  type BillingSummaryResponse,
  type BillingTrendPoint,
  type BillingTrendResponse,
  type BillingUserBreakdownResponse,
  type BillingUserRow,
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
    const setItem = localStorage.setItem as unknown as (key: string, val: string) => void
    setItem('gpt-image-playground-admin-token', TEST_TOKEN)
  })

  afterEach(() => {
    vi.restoreAllMocks()
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

describe('Task 2 — Analytics DTOs and client functions', () => {
  const TEST_TOKEN = 'test-admin-jwt'

  beforeEach(() => {
    const store: Record<string, string> = {}
    vi.stubGlobal('localStorage', {
      getItem: vi.fn((key: string) => store[key] ?? null),
      setItem: vi.fn((key: string, val: string) => { store[key] = val }),
      removeItem: vi.fn((key: string) => { delete store[key] }),
    })
    const setItem = localStorage.setItem as unknown as (key: string, val: string) => void
    setItem('gpt-image-playground-admin-token', TEST_TOKEN)
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  describe('AnalyticsRange type', () => {
    it('accepts only valid range literals', () => {
      const ranges: AnalyticsRange[] = ['today', '7d', '30d', 'all']
      for (const r of ranges) {
        expect(r).toBeTruthy()
      }
    })
  })

  describe('AnalyticsMeta type', () => {
    it('has range, from, to, and moneyScale', () => {
      const meta: AnalyticsMeta = {
        range: '7d',
        from: 1710000000000,
        to: 1710600000000,
        moneyScale: 10000,
      }
      expect(meta.range).toBe('7d')
      expect(meta.moneyScale).toBe(10000)
      expect(meta.from).toBe(1710000000000)
      expect(meta.to).toBe(1710600000000)
    })
  })

  describe('BillingSummary type', () => {
    it('has revenueX10000, costX10000, profitX10000, successImages', () => {
      const summary: BillingSummary = {
        revenueX10000: 123400,
        costX10000: 45600,
        profitX10000: 77800,
        successImages: 321,
      }
      expect(summary.revenueX10000).toBe(123400)
      expect(summary.costX10000).toBe(45600)
      expect(summary.profitX10000).toBe(77800)
      expect(summary.successImages).toBe(321)
    })
  })

  describe('BillingTrendPoint type', () => {
    it('has bucket, money fields, and successImages', () => {
      const point: BillingTrendPoint = {
        bucket: '2026-05-22',
        revenueX10000: 1000,
        costX10000: 400,
        profitX10000: 600,
        successImages: 8,
      }
      expect(point.bucket).toBe('2026-05-22')
      expect(point.profitX10000).toBe(600)
    })
  })

  describe('BillingEndpointRow type', () => {
    it('has endpoint fields and profitRateBps', () => {
      const row: BillingEndpointRow = {
        endpointBaseUrl: 'https://api.openai.com',
        endpointLabel: 'OpenAI',
        successImages: 100,
        revenueX10000: 500000,
        costX10000: 200000,
        profitX10000: 300000,
        profitRateBps: 6000,
      }
      expect(row.endpointBaseUrl).toBe('https://api.openai.com')
      expect(row.profitRateBps).toBe(6000)
    })
  })

  describe('BillingUserRow type', () => {
    it('has user fields and profitRateBps', () => {
      const row: BillingUserRow = {
        userId: 'user-1',
        userLabel: 'Test User',
        successImages: 50,
        revenueX10000: 250000,
        costX10000: 100000,
        profitX10000: 150000,
        profitRateBps: 6000,
      }
      expect(row.userId).toBe('user-1')
      expect(row.profitRateBps).toBe(6000)
    })
  })

  describe('Response wrapper types', () => {
    it('BillingSummaryResponse wraps meta and summary', () => {
      const resp: BillingSummaryResponse = {
        meta: { range: '7d', from: 1, to: 2, moneyScale: 10000 },
        summary: { revenueX10000: 0, costX10000: 0, profitX10000: 0, successImages: 0 },
      }
      expect(resp.meta.range).toBe('7d')
      expect(resp.summary.successImages).toBe(0)
    })

    it('BillingTrendResponse wraps meta and trend array', () => {
      const resp: BillingTrendResponse = {
        meta: { range: '30d', from: 1, to: 2, moneyScale: 10000 },
        trend: [{ bucket: '2026-05-22', revenueX10000: 100, costX10000: 50, profitX10000: 50, successImages: 2 }],
      }
      expect(resp.trend).toHaveLength(1)
      expect(resp.meta.range).toBe('30d')
    })

    it('BillingEndpointBreakdownResponse wraps meta and rows', () => {
      const resp: BillingEndpointBreakdownResponse = {
        meta: { range: 'all', from: 1, to: 2, moneyScale: 10000 },
        rows: [{ endpointBaseUrl: 'https://api.example.com', endpointLabel: 'Example', successImages: 1, revenueX10000: 100, costX10000: 50, profitX10000: 50, profitRateBps: 5000 }],
      }
      expect(resp.rows).toHaveLength(1)
      expect(resp.meta.moneyScale).toBe(10000)
    })

    it('BillingUserBreakdownResponse wraps meta and rows', () => {
      const resp: BillingUserBreakdownResponse = {
        meta: { range: 'today', from: 1, to: 2, moneyScale: 10000 },
        rows: [{ userId: 'u1', userLabel: 'User 1', successImages: 1, revenueX10000: 100, costX10000: 50, profitX10000: 50, profitRateBps: 5000 }],
      }
      expect(resp.rows).toHaveLength(1)
      expect(resp.meta.range).toBe('today')
    })
  })

  describe('adminGetBillingSummary', () => {
    it('calls GET /api/admin/analytics/summary with range query param', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
        new Response(JSON.stringify({
          meta: { range: '7d', from: 1, to: 2, moneyScale: 10000 },
          summary: { revenueX10000: 0, costX10000: 0, profitX10000: 0, successImages: 0 },
        }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
      )

      await adminGetBillingSummary('7d')

      expect(globalThis.fetch).toHaveBeenCalledOnce()
      const [url] = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]
      expect(url).toContain('/api/admin/analytics/summary')
      expect(url).toContain('range=7d')
    })
  })

  describe('adminGetBillingTrend', () => {
    it('calls GET /api/admin/analytics/trend with range query param', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
        new Response(JSON.stringify({
          meta: { range: '30d', from: 1, to: 2, moneyScale: 10000 },
          trend: [],
        }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
      )

      await adminGetBillingTrend('30d')

      const [url] = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]
      expect(url).toContain('/api/admin/analytics/trend')
      expect(url).toContain('range=30d')
    })
  })

  describe('adminGetBillingEndpointBreakdown', () => {
    it('calls GET /api/admin/analytics/endpoints with range query param', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
        new Response(JSON.stringify({
          meta: { range: 'today', from: 1, to: 2, moneyScale: 10000 },
          rows: [],
        }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
      )

      await adminGetBillingEndpointBreakdown('today')

      const [url] = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]
      expect(url).toContain('/api/admin/analytics/endpoints')
      expect(url).toContain('range=today')
    })
  })

  describe('adminGetBillingUserBreakdown', () => {
    it('calls GET /api/admin/analytics/users with range query param', async () => {
      vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
        new Response(JSON.stringify({
          meta: { range: 'all', from: 1, to: 2, moneyScale: 10000 },
          rows: [],
        }), { status: 200, headers: { 'Content-Type': 'application/json' } }),
      )

      await adminGetBillingUserBreakdown('all')

      const [url] = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0]
      expect(url).toContain('/api/admin/analytics/users')
      expect(url).toContain('range=all')
    })
  })
})
