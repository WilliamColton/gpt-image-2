import type { Announcement } from '../types'

const API_BASE_URL = import.meta.env.VITE_BACKEND_URL?.trim()?.replace(/\/+$/, '') || 'http://localhost:3001'
const ADMIN_TOKEN_KEY = 'gpt-image-playground-admin-token'

export interface AdminUser {
  id: string
  label: string
  role: string
  status: string
  quota: number
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

export function adminGetAnnouncement(): Promise<Announcement> {
  return adminRequest('/api/admin/announcement')
}

export function adminUpdateAnnouncement(content: string, enabled: boolean): Promise<Announcement> {
  return adminRequest('/api/admin/announcement', {
    method: 'PUT',
    body: JSON.stringify({ content, enabled }),
  })
}
