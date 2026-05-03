import type { AppSettings, StoredImage, TaskRecord } from '../types'

const API_BASE_URL = import.meta.env.VITE_BACKEND_URL?.trim()?.replace(/\/+$/, '') || 'http://localhost:3001'
const TOKEN_KEY = 'gpt-image-playground-token'

export interface AuthUser {
  id: string
  label: string
  role: 'admin' | 'user'
  imageCount: number
}

export function getBackendToken(): string {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setBackendToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearBackendToken() {
  localStorage.removeItem(TOKEN_KEY)
}

function buildUrl(path: string): string {
  return `${API_BASE_URL}${path.startsWith('/') ? path : `/${path}`}`
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  const token = getBackendToken()
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

export async function loginWithApikey(apikey: string): Promise<{ token: string; user: AuthUser }> {
  const result = await request<{ token: string; user: AuthUser }>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ apikey }),
  })
  setBackendToken(result.token)
  return result
}

export function getMe(): Promise<{ user: AuthUser }> {
  return request('/api/auth/me')
}

export function getPublicConfig(): Promise<AppSettings> {
  return request('/api/config/public')
}

export async function uploadImage(dataUrl: string, source: NonNullable<StoredImage['source']> = 'upload'): Promise<StoredImage> {
  const blob = await dataUrlToBlob(dataUrl)
  const formData = new FormData()
  formData.append('image', blob, `image.${blob.type.split('/')[1] || 'png'}`)
  formData.append('source', source)
  const result = await request<{ id: string; createdAt: number; source: StoredImage['source'] }>('/api/images', {
    method: 'POST',
    body: formData,
  })
  return { id: result.id, dataUrl: getImageUrl(result.id), createdAt: result.createdAt, source: result.source }
}

export function getImageUrl(id: string): string {
  const token = encodeURIComponent(getBackendToken())
  return buildUrl(`/api/images/${encodeURIComponent(id)}?token=${token}`)
}

export function getTasks(): Promise<{ tasks: TaskRecord[] }> {
  return request('/api/tasks')
}

export function putRemoteTask(task: TaskRecord): Promise<{ ok: true }> {
  return request(`/api/tasks/${encodeURIComponent(task.id)}`, {
    method: 'PUT',
    body: JSON.stringify(task),
  })
}

export function deleteRemoteTask(id: string): Promise<{ ok: true }> {
  return request(`/api/tasks/${encodeURIComponent(id)}`, { method: 'DELETE' })
}

export function clearRemoteTasks(): Promise<{ ok: true }> {
  return request('/api/tasks', { method: 'DELETE' })
}

async function dataUrlToBlob(dataUrl: string): Promise<Blob> {
  const response = await fetch(dataUrl)
  return response.blob()
}
