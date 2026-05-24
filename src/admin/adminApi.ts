import type { Announcement, BugFeedback, BugFeedbackStatus, ChangelogEntry, ChangelogEntryPayload } from '../types'

const API_BASE_URL = import.meta.env.VITE_BACKEND_URL?.trim()?.replace(/\/+$/, '') || 'http://localhost:3001'
const ADMIN_TOKEN_KEY = 'gpt-image-playground-admin-token'

export interface AdminUser {
  id: string
  label: string
  username?: string
  role: string
  status: string
  quota: number
  unlimitedQuota: boolean
  usedCount: number
  createdAt: number
}

export interface RedemptionCode {
  id: string
  code: string
  quota: number
  usedBy: string | null
  usedAt: number | null
  createdAt: number
}

export interface ApiEndpoint {
  baseUrl: string
  apiKey: string
  maxConcurrency?: number
  priority?: number
  costPerImageX10000?: number
}

function getAdminToken(): string {
  return localStorage.getItem(ADMIN_TOKEN_KEY) || ''
}

function setAdminToken(token: string) {
  localStorage.setItem(ADMIN_TOKEN_KEY, token)
}

export function clearAdminToken() {
  localStorage.removeItem(ADMIN_TOKEN_KEY)
}

function buildUrl(path: string): string {
  return `${API_BASE_URL}${path.startsWith('/') ? path : `/${path}`}`
}

async function adminRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  const token = getAdminToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  if (options.body && !(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(buildUrl(path), { ...options, headers, cache: 'no-store' })
  if (!response.ok) {
    let message = `HTTP ${response.status}`
    try {
      const payload = await response.json()
      message = payload.error || payload.message || message
    } catch {
      message = await response.text()
    }
    throw new Error(message)
  }
  return response.json() as Promise<T>
}

export async function adminLogin(apikey: string): Promise<{ token: string }> {
  const result = await adminRequest<{ token: string }>('/api/admin/login', {
    method: 'POST',
    body: JSON.stringify({ apikey }),
  })
  setAdminToken(result.token)
  return result
}

export function adminListUsers(): Promise<{ users: AdminUser[] }> {
  return adminRequest('/api/admin/users')
}

export function adminUpdateQuota(userId: string, delta: number, resetUsedCount = false, mode: 'delta' | 'set' = 'delta'): Promise<{ ok: true }> {
  return adminRequest(`/api/admin/users/${encodeURIComponent(userId)}/quota`, {
    method: 'PUT',
    body: JSON.stringify({ delta, resetUsedCount, mode }),
  })
}

export function adminToggleStatus(userId: string, status: 'active' | 'disabled'): Promise<{ ok: true }> {
  return adminRequest(`/api/admin/users/${encodeURIComponent(userId)}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  })
}

export function adminToggleUnlimited(userId: string, unlimited: boolean): Promise<{ ok: true }> {
  return adminRequest(`/api/admin/users/${encodeURIComponent(userId)}/unlimited`, {
    method: 'PUT',
    body: JSON.stringify({ unlimited }),
  })
}

export function isAdminLoggedIn(): boolean {
  return !!getAdminToken()
}

export function adminCreateCodes(quota: number, count: number = 1): Promise<{ codes: RedemptionCode[] }> {
  return adminRequest('/api/admin/codes', {
    method: 'POST',
    body: JSON.stringify({ quota, count }),
  })
}

export function adminListCodes(): Promise<{ codes: RedemptionCode[] }> {
  return adminRequest('/api/admin/codes')
}

export function adminDeleteUser(userId: string): Promise<{ ok: true }> {
  return adminRequest(`/api/admin/users/${encodeURIComponent(userId)}`, {
    method: 'DELETE',
  })
}

export function adminDeleteUsers(ids: string[]): Promise<{ ok: true; deleted: number }> {
  return adminRequest('/api/admin/users', {
    method: 'DELETE',
    body: JSON.stringify({ ids }),
  })
}

export function adminDeleteCodes(ids: string[]): Promise<{ ok: true; deleted: number }> {
  return adminRequest('/api/admin/codes', {
    method: 'DELETE',
    body: JSON.stringify({ ids }),
  })
}

export function adminGetEndpoints(): Promise<{ endpoints: ApiEndpoint[] }> {
  return adminRequest('/api/admin/config/endpoints')
}

export function adminUpdateEndpoints(endpoints: ApiEndpoint[]): Promise<{ ok: true; endpoints: ApiEndpoint[] }> {
  return adminRequest('/api/admin/config/endpoints', {
    method: 'PUT',
    body: JSON.stringify({ endpoints }),
  })
}

// ─── Pricing Configuration ───

export interface PricingConfigResponse {
  endpoints: ApiEndpoint[]
  salePriceX10000: number
  moneyScale: number
  ok?: true
}

export function adminGetPricingConfig(): Promise<PricingConfigResponse> {
  return adminRequest<PricingConfigResponse>('/api/admin/config/pricing')
}

export function adminUpdatePricingConfig(endpoints: ApiEndpoint[], salePriceX10000: number): Promise<PricingConfigResponse> {
  return adminRequest<PricingConfigResponse>('/api/admin/config/pricing', {
    method: 'PUT',
    body: JSON.stringify({ endpoints, salePriceX10000 }),
  })
}

// ─── Billing Analytics ───

export type AnalyticsRange = 'today' | '7d' | '30d' | 'all'

export interface AnalyticsMeta {
  range: AnalyticsRange
  from: number | null
  to: number | null
  moneyScale: number
}

export interface BillingSummary {
  revenueX10000: number
  costX10000: number
  profitX10000: number
  successImages: number
}

export interface BillingSummaryResponse {
  meta: AnalyticsMeta
  summary: BillingSummary
}

export interface BillingTrendPoint {
  bucket: string
  revenueX10000: number
  costX10000: number
  profitX10000: number
  successImages: number
}

export interface BillingTrendResponse {
  meta: AnalyticsMeta
  trend: BillingTrendPoint[]
}

export interface BillingEndpointRow {
  endpointBaseUrl: string
  endpointLabel: string
  successImages: number
  revenueX10000: number
  costX10000: number
  profitX10000: number
  profitRateBps: number
}

export interface BillingEndpointBreakdownResponse {
  meta: AnalyticsMeta
  rows: BillingEndpointRow[]
}

export interface BillingUserRow {
  userId: string
  userLabel: string
  successImages: number
  revenueX10000: number
  costX10000: number
  profitX10000: number
  profitRateBps: number
}

export interface BillingUserBreakdownResponse {
  meta: AnalyticsMeta
  rows: BillingUserRow[]
}

export function adminGetBillingSummary(range: AnalyticsRange): Promise<BillingSummaryResponse> {
  return adminRequest<BillingSummaryResponse>(`/api/admin/analytics/summary?range=${encodeURIComponent(range)}`)
}

export function adminGetBillingTrend(range: AnalyticsRange): Promise<BillingTrendResponse> {
  return adminRequest<BillingTrendResponse>(`/api/admin/analytics/trend?range=${encodeURIComponent(range)}`)
}

export function adminGetBillingEndpointBreakdown(range: AnalyticsRange): Promise<BillingEndpointBreakdownResponse> {
  return adminRequest<BillingEndpointBreakdownResponse>(`/api/admin/analytics/endpoints?range=${encodeURIComponent(range)}`)
}

export function adminGetBillingUserBreakdown(range: AnalyticsRange): Promise<BillingUserBreakdownResponse> {
  return adminRequest<BillingUserBreakdownResponse>(`/api/admin/analytics/users?range=${encodeURIComponent(range)}`)
}

export function adminGetAnnouncement(): Promise<Announcement> {
  return adminRequest('/api/admin/announcement')
}

export function adminUpdateAnnouncement(content: string, enabled: boolean): Promise<Announcement> {
  return adminRequest('/api/admin/announcement', {
    method: 'PUT',
    body: JSON.stringify({ content, enabled }),
  })
}

export function adminListFeedbacks(): Promise<{ feedbacks: BugFeedback[] }> {
  return adminRequest('/api/admin/feedback')
}

export function adminUpdateFeedbackStatus(id: string, status: BugFeedbackStatus): Promise<BugFeedback> {
  return adminRequest(`/api/admin/feedback/${encodeURIComponent(id)}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  })
}

export function adminListChangelogEntries(): Promise<{ changelogs: ChangelogEntry[] }> {
  return adminRequest('/api/admin/changelog')
}

export function adminCreateChangelogEntry(payload: ChangelogEntryPayload): Promise<ChangelogEntry> {
  return adminRequest('/api/admin/changelog', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function adminUpdateChangelogEntry(id: string, payload: ChangelogEntryPayload): Promise<ChangelogEntry> {
  return adminRequest(`/api/admin/changelog/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export function adminDeleteChangelogEntry(id: string): Promise<{ ok: true }> {
  return adminRequest(`/api/admin/changelog/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}

// ─── Invite Code Management ───

export function adminResetPassword(userId: string, password: string): Promise<{ ok: true }> {
  return adminRequest(`/api/admin/users/${encodeURIComponent(userId)}/password`, {
    method: 'PUT',
    body: JSON.stringify({ password }),
  })
}

export function adminGetInviteConfig(): Promise<{ inviterReward: number; inviteeReward: number; defaultQuota: number; inviteEnabled: boolean }> {
  return adminRequest('/api/admin/invite-config')
}

export function adminUpdateInviteConfig(inviterReward: number, inviteeReward: number, defaultQuota: number, inviteEnabled: boolean): Promise<{ ok: true; inviterReward: number; inviteeReward: number; defaultQuota: number; inviteEnabled: boolean }> {
  return adminRequest('/api/admin/invite-config', {
    method: 'PUT',
    body: JSON.stringify({ inviterReward, inviteeReward, defaultQuota, inviteEnabled }),
  })
}

export function adminListInvites(): Promise<{ invites: Array<{ username: string; inviteCode: string; usageCount: number }> }> {
  return adminRequest('/api/admin/invites')
}
