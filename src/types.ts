// ===== 设置 =====

export type ApiMode = 'images'

export type ThemeMode = 'system' | 'light' | 'dark'

export interface AppSettings {
  model: string
  timeout: number
  apiMode: ApiMode
  codexCli: boolean
  theme: ThemeMode
  inviteEnabled: boolean
}

export const DEFAULT_SETTINGS: AppSettings = {
  model: 'gpt-image-2',
  timeout: 300,
  apiMode: 'images',
  codexCli: false,
  theme: 'system',
  inviteEnabled: true,
}

// ===== 任务参数 =====

export interface TaskParams {
  size: string
  quality: 'auto' | 'low' | 'medium' | 'high'
  output_format: 'png' | 'jpeg' | 'webp'
  output_compression: number | null
  moderation: 'auto' | 'low'
  n: number
}

export const DEFAULT_PARAMS: TaskParams = {
  size: 'auto',
  quality: 'auto',
  output_format: 'png',
  output_compression: null,
  moderation: 'auto',
  n: 1,
}

// ===== 输入图片（UI 层面） =====

export interface InputImage {
  /** IndexedDB image store 的 id（SHA-256 hash） */
  id: string
  /** data URL，用于预览 */
  dataUrl: string
}

export interface MaskDraft {
  targetImageId: string
  maskDataUrl: string
  updatedAt: number
}

// ===== 任务记录 =====

export type TaskStatus = 'queued' | 'running' | 'done' | 'error'

export interface TaskRecord {
  id: string
  prompt: string
  params: TaskParams
  /** API 返回的实际生效参数，用于标记与请求值不一致的情况 */
  actualParams?: Partial<TaskParams>
  /** 输出图片对应的实际生效参数，key 为 outputImages 中的图片 id */
  actualParamsByImage?: Record<string, Partial<TaskParams>>
  /** 输出图片对应的 API 改写提示词，key 为 outputImages 中的图片 id */
  revisedPromptByImage?: Record<string, string>
  /** 输入图片的 image store id 列表 */
  inputImageIds: string[]
  maskTargetImageId?: string | null
  maskImageId?: string | null
  /** 输出图片的 image store id 列表 */
  outputImages: string[]
  status: TaskStatus
  error: string | null
  createdAt: number
  finishedAt: number | null
  /** 总耗时毫秒 */
  elapsed: number | null
  /** 是否收藏 */
  isFavorite?: boolean
}

// ===== IndexedDB 存储的图片 =====

export interface StoredImage {
  id: string
  dataUrl: string
  /** 图片首次存储时间（ms） */
  createdAt?: number
  /** 图片来源：用户上传 / API 生成 / 遮罩 */
  source?: 'upload' | 'generated' | 'mask'
}

export interface Announcement {
  content: string
  enabled: boolean
  updatedAt: number
}

export type BugFeedbackCategory = 'bug' | 'feature'
export type BugFeedbackStatus = 'open' | 'reviewing' | 'resolved'

export interface BugFeedback {
  id: string
  userId: string
  userLabel: string
  category: BugFeedbackCategory
  content: string
  contact: string
  status: BugFeedbackStatus
  createdAt: number
  updatedAt: number
}

export interface CreateBugFeedbackPayload {
  category: BugFeedbackCategory
  content: string
  contact?: string
}

export interface ChangelogEntry {
  id: string
  version: string
  title: string
  content: string
  published: boolean
  createdAt: number
  updatedAt: number
  publishedAt: number | null
}

export interface ChangelogEntryPayload {
  version: string
  title: string
  content: string
  published: boolean
}
