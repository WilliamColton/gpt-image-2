import type { AppSettings, StoredImage, TaskRecord, TaskParams } from '../types'

const API_BASE_URL = import.meta.env.VITE_BACKEND_URL?.trim()?.replace(/\/+$/, '') || 'http://localhost:3001'
const TOKEN_KEY = 'gpt-image-playground-token'

export interface AuthUser {
  id: string
  label: string
  role: 'admin' | 'user'
  imageCount: number
  quota: number
  usedCount: number
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

export async function loginWithCode(code: string): Promise<{ token: string; user: AuthUser }> {
  const result = await request<{ token: string; user: AuthUser }>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ code }),
  })
  setBackendToken(result.token)
  return result
}

export function redeemCode(code: string): Promise<{ ok: true; quota?: number; usedCount?: number }> {
  return request('/api/auth/redeem', {
    method: 'POST',
    body: JSON.stringify({ code }),
  })
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

/** Submit image generation task to backend */
export async function submitGenerateTask(taskId: string, prompt: string, params: TaskParams, inputImageIds: string[], codexCli: boolean): Promise<{ taskId: string; status: string }> {
  return request('/api/generate', {
    method: 'POST',
    body: JSON.stringify({ taskId, prompt, params, inputImageIds, codexCli }),
  })
}

/** Submit image edit task to backend */
export async function submitEditTask(taskId: string, prompt: string, params: TaskParams, inputImageIds: string[], maskImageId: string | null, codexCli: boolean): Promise<{ taskId: string; status: string }> {
  return request('/api/edit', {
    method: 'POST',
    body: JSON.stringify({ taskId, prompt, params, inputImageIds, maskImageId, codexCli }),
  })
}

async function dataUrlToBlob(dataUrl: string): Promise<Blob> {
  const response = await fetch(dataUrl)
  return response.blob()
}

/** Stream task status via SSE. Returns an AbortController to cancel. */
export function streamTaskStatus(
  taskId: string,
  onUpdate: (task: TaskRecord) => void,
  onError?: (err: Error) => void,
): AbortController {
  const controller = new AbortController()

  ;(async () => {
    try {
      const token = getBackendToken()
      const response = await fetch(buildUrl(`/api/tasks/${encodeURIComponent(taskId)}/stream`), {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        signal: controller.signal,
      })
      if (!response.ok || !response.body) {
        throw new Error(`SSE connection failed: ${response.status}`)
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const task = JSON.parse(line.slice(6)) as TaskRecord
              onUpdate(task)
              if (task.status === 'done' || task.status === 'error') {
                reader.cancel()
                return
              }
            } catch { /* ignore parse errors */ }
          }
        }
      }
    } catch (err) {
      if ((err as Error).name !== 'AbortError') {
        onError?.(err as Error)
      }
    }
  })()

  return controller
}
