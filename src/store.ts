import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type {
  AppSettings,
  TaskParams,
  InputImage,
  MaskDraft,
  TaskRecord,
} from './types'
import { DEFAULT_SETTINGS, DEFAULT_PARAMS } from './types'
import {
  clearBackendToken,
  clearRemoteTasks,
  deleteRemoteTask,
  getBackendToken,
  getMe,
  getPublicConfig,
  getTasks as fetchTasks,
  putRemoteTask,
  submitGenerateTask,
  submitEditTask,
  uploadImage,
  type AuthUser,
} from './lib/backendApi'
import {
  hashDataUrl,
  getImage,
  putImage,
} from './lib/db'
import { validateMaskMatchesImage } from './lib/canvasImage'
import { orderInputImagesForMask } from './lib/mask'
import { normalizeImageSize } from './lib/size'

// ===== Image cache =====
// 内存缓存，id → dataUrl，避免每次从 IndexedDB 读取

const imageCache = new Map<string, string>()

// ===== Global polling =====

let pollTimer: ReturnType<typeof setInterval> | null = null
const POLL_INTERVAL = 5000

function startPolling() {
  if (pollTimer) return
  pollTimer = setInterval(pollRunningTasks, POLL_INTERVAL)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function pollRunningTasks() {
  const runningTasks = useStore.getState().tasks.filter(t => t.status === 'running')
  if (runningTasks.length === 0) {
    stopPolling()
    return
  }

  try {
    const { tasks: remoteTasks } = await fetchTasks()
    for (const local of runningTasks) {
      const remote = remoteTasks.find(t => t.id === local.id)
      if (!remote) continue

      if (remote.status === 'done') {
        for (const imgId of remote.outputImages || []) {
          await setCacheFromIdbOrRemote(imgId)
        }
        const { tasks: currentTasks, setTasks } = useStore.getState()
        setTasks(currentTasks.map(t => t.id === local.id ? { ...t, ...remote } : t))
        useStore.getState().showToast(`生成完成，共 ${(remote.outputImages || []).length} 张图片`, 'success')
        if (local.maskImageId) useStore.getState().clearMaskDraft()
      } else if (remote.status === 'error') {
        updateTaskInStore(local.id, {
          status: 'error',
          error: remote.error || 'Unknown error',
          finishedAt: Date.now(),
          elapsed: Date.now() - local.createdAt,
        })
        useStore.getState().setDetailTaskId(local.id)
      }
    }
  } catch {
    // ignore poll errors, will retry next interval
  }
}

export function getCachedImage(id: string): string | undefined {
  return imageCache.get(id)
}

async function setCacheFromIdbOrRemote(id: string) {
  try {
    const stored = await getImage(id)
    if (stored?.dataUrl) {
      imageCache.set(id, stored.dataUrl)
      return
    }
  } catch { /* ignore */ }
  imageCache.set(id, getRemoteImageDataUrl(id))
}

export async function ensureImageCached(id: string): Promise<string | undefined> {
  const cached = imageCache.get(id)
  // 内存缓存是 dataUrl 直接返回；是后端 URL 则继续查 IDB
  if (cached && !cached.startsWith('http')) return cached
  // 优先从 IndexedDB 读取（避免 CORS 跨域请求）
  try {
    const stored = await getImage(id)
    if (stored?.dataUrl) {
      imageCache.set(id, stored.dataUrl)
      return stored.dataUrl
    }
  } catch { /* ignore IDB errors */ }
  if (cached) return cached
  const url = getRemoteImageDataUrl(id)
  imageCache.set(id, url)
  return url
}

/** 从后端 fetch 图片并转为 base64 dataUrl，存入缓存和 IDB */
async function fetchAndCacheImage(id: string): Promise<string | undefined> {
  const url = getRemoteImageDataUrl(id)
  try {
    const resp = await fetch(url)
    if (!resp.ok) return undefined
    const blob = await resp.blob()
    const dataUrl = await new Promise<string>((resolve, reject) => {
      const reader = new FileReader()
      reader.onloadend = () => resolve(reader.result as string)
      reader.onerror = reject
      reader.readAsDataURL(blob)
    })
    imageCache.set(id, dataUrl)
    putImage({ id, dataUrl, createdAt: Date.now(), source: 'upload' }).catch(() => {})
    return dataUrl
  } catch {
    return undefined
  }
}

// ===== Store 类型 =====

interface AppState {
  // 认证
  authUser: AuthUser | null
  setAuthUser: (user: AuthUser | null) => void

  // 设置
  settings: AppSettings
  setSettings: (s: Partial<AppSettings>) => void
  dismissedCodexCliPrompts: string[]
  dismissCodexCliPrompt: (key: string) => void

  // 输入
  prompt: string
  setPrompt: (p: string) => void
  inputImages: InputImage[]
  addInputImage: (img: InputImage) => void
  removeInputImage: (idx: number) => void
  clearInputImages: () => void
  setInputImages: (imgs: InputImage[]) => void
  maskDraft: MaskDraft | null
  setMaskDraft: (draft: MaskDraft | null) => void
  clearMaskDraft: () => void
  maskEditorImageId: string | null
  setMaskEditorImageId: (id: string | null) => void

  // 参数
  params: TaskParams
  setParams: (p: Partial<TaskParams>) => void

  // 任务列表
  tasks: TaskRecord[]
  setTasks: (t: TaskRecord[]) => void

  // 搜索和筛选
  searchQuery: string
  setSearchQuery: (q: string) => void
  filterStatus: 'all' | 'running' | 'done' | 'error'
  setFilterStatus: (status: AppState['filterStatus']) => void
  filterFavorite: boolean
  setFilterFavorite: (f: boolean) => void

  // 多选
  selectedTaskIds: string[]
  setSelectedTaskIds: (ids: string[] | ((prev: string[]) => string[])) => void
  toggleTaskSelection: (id: string, force?: boolean) => void
  clearSelection: () => void

  // UI
  detailTaskId: string | null
  setDetailTaskId: (id: string | null) => void
  lightboxImageId: string | null
  lightboxImageList: string[]
  setLightboxImageId: (id: string | null, list?: string[]) => void
  showSettings: boolean
  setShowSettings: (v: boolean) => void

  // Toast
  toast: { message: string; type: 'info' | 'success' | 'error' } | null
  showToast: (message: string, type?: 'info' | 'success' | 'error') => void

  // Confirm dialog
  confirmDialog: {
    title: string
    message: string
    confirmText?: string
    messageAlign?: 'left' | 'center'
    tone?: 'danger' | 'warning'
    action: () => void
    cancelAction?: () => void
  } | null
  setConfirmDialog: (d: AppState['confirmDialog']) => void
}

export const useStore = create<AppState>()(
  persist(
    (set, get) => ({
      // Auth
      authUser: null,
      setAuthUser: (authUser) => set({ authUser }),

      // Settings
      settings: { ...DEFAULT_SETTINGS },
      setSettings: (s) => set((st) => ({
        settings: {
          ...st.settings,
          ...s,
          apiMode:
            s.apiMode === 'images'
              ? s.apiMode
              : st.settings.apiMode ?? DEFAULT_SETTINGS.apiMode,
          codexCli: s.codexCli ?? st.settings.codexCli ?? DEFAULT_SETTINGS.codexCli,
        },
      })),
      dismissedCodexCliPrompts: [],
      dismissCodexCliPrompt: (key) => set((st) => ({
        dismissedCodexCliPrompts: st.dismissedCodexCliPrompts.includes(key)
          ? st.dismissedCodexCliPrompts
          : [...st.dismissedCodexCliPrompts, key],
      })),

      // Input
      prompt: '',
      setPrompt: (prompt) => set({ prompt }),
      inputImages: [],
      addInputImage: (img) =>
        set((s) => {
          if (s.inputImages.find((i) => i.id === img.id)) return s
          return { inputImages: [...s.inputImages, img] }
        }),
      removeInputImage: (idx) =>
        set((s) => {
          const removed = s.inputImages[idx]
          const shouldClearMask = removed?.id === s.maskDraft?.targetImageId
          return {
            inputImages: s.inputImages.filter((_, i) => i !== idx),
            ...(shouldClearMask ? { maskDraft: null, maskEditorImageId: null } : {}),
          }
        }),
      clearInputImages: () =>
        set((s) => {
          for (const img of s.inputImages) imageCache.delete(img.id)
          return { inputImages: [], maskDraft: null, maskEditorImageId: null }
        }),
      setInputImages: (imgs) =>
        set((s) => {
          const shouldClearMask =
            Boolean(s.maskDraft) && !imgs.some((img) => img.id === s.maskDraft?.targetImageId)
          return {
            inputImages: imgs,
            ...(shouldClearMask ? { maskDraft: null, maskEditorImageId: null } : {}),
          }
        }),
      maskDraft: null,
      setMaskDraft: (maskDraft) => set({ maskDraft }),
      clearMaskDraft: () => set({ maskDraft: null }),
      maskEditorImageId: null,
      setMaskEditorImageId: (maskEditorImageId) => set({ maskEditorImageId }),

      // Params
      params: { ...DEFAULT_PARAMS },
      setParams: (p) => set((s) => ({ params: { ...s.params, ...p } })),

      // Tasks
      tasks: [],
      setTasks: (tasks) => set({ tasks }),

      // Search & Filter
      searchQuery: '',
      setSearchQuery: (searchQuery) => set({ searchQuery }),
      filterStatus: 'all',
      setFilterStatus: (filterStatus) => set({ filterStatus }),
      filterFavorite: false,
      setFilterFavorite: (filterFavorite) => set({ filterFavorite }),

      // Selection
      selectedTaskIds: [],
      setSelectedTaskIds: (updater) => set((s) => ({
        selectedTaskIds: typeof updater === 'function' ? updater(s.selectedTaskIds) : updater
      })),
      toggleTaskSelection: (id, force) => set((s) => {
        const isSelected = s.selectedTaskIds.includes(id)
        const shouldSelect = force !== undefined ? force : !isSelected
        if (shouldSelect === isSelected) return s
        return {
          selectedTaskIds: shouldSelect
            ? [...s.selectedTaskIds, id]
            : s.selectedTaskIds.filter((x) => x !== id)
        }
      }),
      clearSelection: () => set({ selectedTaskIds: [] }),

      // UI
      detailTaskId: null,
      setDetailTaskId: (detailTaskId) => set({ detailTaskId }),
      lightboxImageId: null,
      lightboxImageList: [],
      setLightboxImageId: (lightboxImageId, list) =>
        set({ lightboxImageId, lightboxImageList: list ?? (lightboxImageId ? [lightboxImageId] : []) }),
      showSettings: false,
      setShowSettings: (showSettings) => set({ showSettings }),

      // Toast
      toast: null,
      showToast: (message, type = 'info') => {
        set({ toast: { message, type } })
        setTimeout(() => {
          set((s) => (s.toast?.message === message ? { toast: null } : s))
        }, 3000)
      },

      // Confirm
      confirmDialog: null,
      setConfirmDialog: (confirmDialog) => set({ confirmDialog }),
    }),
    {
      name: 'gpt-image-playground',
      partialize: (state) => ({
        settings: state.settings,
        authUser: state.authUser,
        params: state.params,
        dismissedCodexCliPrompts: state.dismissedCodexCliPrompts,
      }),
    },
  ),
)

export async function bootstrapBackendSession() {
  if (!getBackendToken()) return
  const [{ user }, { tasks }, publicConfig] = await Promise.all([
    getMe(),
    fetchTasks(),
    getPublicConfig(),
  ])
  useStore.getState().setAuthUser(user)
  useStore.getState().setTasks(tasks)
  useStore.getState().setSettings({ ...publicConfig, apiKey: useStore.getState().settings.apiKey })
  imageCache.clear()
  for (const task of tasks) {
    for (const id of task.inputImageIds || []) await setCacheFromIdbOrRemote(id)
    if (task.maskImageId) await setCacheFromIdbOrRemote(task.maskImageId)
    for (const id of task.outputImages || []) await setCacheFromIdbOrRemote(id)
  }

  // Resume polling if any tasks are still running
  if (tasks.some(t => t.status === 'running')) {
    startPolling()
  }
}

export async function logout() {
  stopPolling()
  clearBackendToken()
  imageCache.clear()
  useStore.getState().setAuthUser(null)
  useStore.getState().setTasks([])
}

function getRemoteImageDataUrl(id: string): string {
  return `${import.meta.env.VITE_BACKEND_URL?.trim()?.replace(/\/+$/, '') || 'http://localhost:3001'}/api/images/${encodeURIComponent(id)}?token=${encodeURIComponent(getBackendToken())}`
}

// ===== Actions =====

let uid = 0
function genId(): string {
  return Date.now().toString(36) + (++uid).toString(36) + Math.random().toString(36).slice(2, 6)
}

export function getCodexCliPromptKey(settings: AppSettings): string {
  return `${settings.baseUrl}`
}

export function showCodexCliPrompt(force = false, reason = '接口返回的提示词已被改写') {
  const state = useStore.getState()
  const settings = state.settings
  const promptKey = getCodexCliPromptKey(settings)
  if (!force && (settings.codexCli || state.dismissedCodexCliPrompts.includes(promptKey))) return

  state.setConfirmDialog({
    title: '检测到 Codex CLI API',
    message: `${reason}，当前 API 来源很可能是 Codex CLI。\n\n是否开启 Codex CLI 兼容模式？开启后会禁用在此处无效的质量参数，并在 Images API 多图生成时使用并发请求，解决该 API 数量参数无效的问题。同时，提示词文本开头会加入简短的不改写要求，避免模型重写提示词，偏离原意。`,
    confirmText: '开启',
    action: () => {
      const state = useStore.getState()
      state.dismissCodexCliPrompt(promptKey)
      state.setSettings({ codexCli: true })
    },
    cancelAction: () => useStore.getState().dismissCodexCliPrompt(promptKey),
  })
}

/** 初始化：从 IndexedDB 加载任务和图片缓存，清理孤立图片 */
export async function initStore() {
  if (getBackendToken()) {
    try {
      await bootstrapBackendSession()
    } catch {
      clearBackendToken()
      useStore.getState().setAuthUser(null)
      useStore.getState().setTasks([])
    }
  }
}

/** 提交新任务 */
export async function submitTask(options: { allowFullMask?: boolean } = {}) {
  const { settings, prompt, inputImages, maskDraft, params, showToast, setConfirmDialog } =
    useStore.getState()

  if (!useStore.getState().authUser) {
    showToast('请先输入 apikey 登录', 'error')
    return
  }

  if (!prompt.trim()) {
    showToast('请输入提示词', 'error')
    return
  }

  let orderedInputImages = inputImages
  let maskTargetImageId: string | null = null

  if (maskDraft) {
    try {
      orderedInputImages = orderInputImagesForMask(inputImages, maskDraft.targetImageId)
      const coverage = await validateMaskMatchesImage(maskDraft.maskDataUrl, orderedInputImages[0].dataUrl)
      if (coverage === 'full' && !options.allowFullMask) {
        setConfirmDialog({
          title: '确认编辑整张图片？',
          message: '当前遮罩覆盖了整张图片，提交后可能会重绘全部内容。是否继续？',
          confirmText: '继续提交',
          tone: 'warning',
          action: () => {
            void submitTask({ allowFullMask: true })
          },
        })
        return
      }
      maskTargetImageId = maskDraft.targetImageId
    } catch (err) {
      if (!inputImages.some((img) => img.id === maskDraft.targetImageId)) {
        useStore.getState().clearMaskDraft()
      }
      showToast(err instanceof Error ? err.message : String(err), 'error')
      return
    }
  }

  const normalizedParams = {
    ...params,
    size: normalizeImageSize(params.size) || DEFAULT_PARAMS.size,
    quality: settings.codexCli ? DEFAULT_PARAMS.quality : params.quality,
  }
  if (normalizedParams.size !== params.size || normalizedParams.quality !== params.quality) {
    useStore.getState().setParams({ size: normalizedParams.size, quality: normalizedParams.quality })
  }

  // Show task UI immediately — uploads happen below
  const taskId = genId()
  const task: TaskRecord = {
    id: taskId,
    prompt: prompt.trim(),
    params: normalizedParams,
    inputImageIds: orderedInputImages.map((i) => i.id),
    maskTargetImageId,
    maskImageId: null,
    outputImages: [],
    status: 'running',
    error: null,
    createdAt: Date.now(),
    finishedAt: null,
    elapsed: null,
  }

  const latestTasks = useStore.getState().tasks
  useStore.getState().setTasks([task, ...latestTasks])
  putRemoteTask(task).catch(() => {})

  // --- Upload images (runs after UI is visible) ---

  let maskImageId: string | null = null
  if (maskDraft) {
    try {
      const maskUploaded = await uploadImage(maskDraft.maskDataUrl, 'mask')
      maskImageId = maskUploaded.id
      putImage({ id: maskImageId, dataUrl: maskDraft.maskDataUrl, createdAt: maskUploaded.createdAt, source: 'mask' }).catch(() => {})
      imageCache.set(maskImageId, maskDraft.maskDataUrl)
      updateTaskInStore(taskId, { maskImageId })
    } catch (err) {
      if (!inputImages.some((img) => img.id === maskDraft.targetImageId)) {
        useStore.getState().clearMaskDraft()
      }
      updateTaskInStore(taskId, { status: 'error', error: err instanceof Error ? err.message : String(err), finishedAt: Date.now() })
      return
    }
  }

  for (const img of orderedInputImages) {
    if (!img.dataUrl.startsWith('http')) {
      const originalDataUrl = img.dataUrl
      const uploaded = await uploadImage(originalDataUrl, 'upload')
      putImage({ id: uploaded.id, dataUrl: originalDataUrl, createdAt: uploaded.createdAt, source: 'upload' }).catch(() => {})
      imageCache.delete(img.id)
      imageCache.set(uploaded.id, originalDataUrl)
      img.id = uploaded.id
      img.dataUrl = getRemoteImageDataUrl(uploaded.id)
    }
  }

  // Update task with final input image IDs (after upload)
  updateTaskInStore(taskId, { inputImageIds: orderedInputImages.map((i) => i.id) })

  executeTask(taskId)
}

async function executeTask(taskId: string) {
  const task = useStore.getState().tasks.find((t) => t.id === taskId)
  if (!task) return

  const settings = useStore.getState().settings

  try {
    // Submit task to backend based on apiMode
    if (task.inputImageIds.length > 0 && task.maskImageId) {
      await submitEditTask(taskId, task.prompt, task.params, task.inputImageIds, task.maskImageId, settings.codexCli)
    } else {
      await submitGenerateTask(taskId, task.prompt, task.params, task.inputImageIds, settings.codexCli)
    }

    // Global polling handles completion — start it if not already running
    startPolling()
  } catch (err) {
    updateTaskInStore(taskId, {
      status: 'error',
      error: err instanceof Error ? err.message : String(err),
      finishedAt: Date.now(),
      elapsed: Date.now() - task.createdAt,
    })
    useStore.getState().setDetailTaskId(taskId)
  }
}

export function updateTaskInStore(taskId: string, patch: Partial<TaskRecord>) {
  const { tasks, setTasks } = useStore.getState()
  const updated = tasks.map((t) =>
    t.id === taskId ? { ...t, ...patch } : t,
  )
  setTasks(updated)
  const task = updated.find((t) => t.id === taskId)
  if (task) putRemoteTask(task)
}

/** 复用配置 */
export async function reuseConfig(task: TaskRecord) {
  const { setPrompt, setParams, setInputImages, setMaskDraft, clearMaskDraft, showToast } = useStore.getState()
  setPrompt(task.prompt)
  setParams(task.params)

  // 恢复输入图片
  const imgs: InputImage[] = []
  for (const imgId of task.inputImageIds) {
    const dataUrl = await ensureImageCached(imgId)
    if (dataUrl) {
      imgs.push({ id: imgId, dataUrl })
    }
  }
  setInputImages(imgs)
  const maskTargetImageId = task.maskTargetImageId ?? (task.maskImageId ? task.inputImageIds[0] : null)
  if (maskTargetImageId && task.maskImageId && imgs.some((img) => img.id === maskTargetImageId)) {
    const maskDataUrl = await ensureImageCached(task.maskImageId)
    if (maskDataUrl) {
      setMaskDraft({
        targetImageId: maskTargetImageId,
        maskDataUrl,
        updatedAt: Date.now(),
      })
    } else {
      clearMaskDraft()
    }
  } else {
    clearMaskDraft()
  }
  showToast('已复用配置到输入框', 'success')
}

/** 编辑输出：将输出图加入输入 */
export async function editOutputs(task: TaskRecord) {
  const { inputImages, addInputImage, clearMaskDraft, showToast } = useStore.getState()
  if (!task.outputImages?.length) return

  clearMaskDraft()
  let added = 0
  for (const imgId of task.outputImages) {
    if (inputImages.find((i) => i.id === imgId)) continue
    const dataUrl = await ensureImageCached(imgId)
    if (dataUrl) {
      addInputImage({ id: imgId, dataUrl })
      added++
    }
  }
  showToast(`已添加 ${added} 张输出图到输入`, 'success')
}

/** 删除多条任务 */
export async function removeMultipleTasks(taskIds: string[]) {
  const { tasks, setTasks, inputImages, showToast, clearSelection, selectedTaskIds } = useStore.getState()
  
  if (!taskIds.length) return

  const toDelete = new Set(taskIds)
  const remaining = tasks.filter(t => !toDelete.has(t.id))

  // 收集所有被删除任务的关联图片
  const deletedImageIds = new Set<string>()
  for (const t of tasks) {
    if (toDelete.has(t.id)) {
      for (const id of t.inputImageIds || []) deletedImageIds.add(id)
      if (t.maskImageId) deletedImageIds.add(t.maskImageId)
      for (const id of t.outputImages || []) deletedImageIds.add(id)
    }
  }

  setTasks(remaining)
  for (const id of taskIds) {
    await deleteRemoteTask(id)
  }

  // 找出其他任务仍引用的图片
  const stillUsed = new Set<string>()
  for (const t of remaining) {
    for (const id of t.inputImageIds || []) stillUsed.add(id)
    if (t.maskImageId) stillUsed.add(t.maskImageId)
    for (const id of t.outputImages || []) stillUsed.add(id)
  }
  for (const img of inputImages) stillUsed.add(img.id)

  // 删除孤立图片
  for (const imgId of deletedImageIds) {
    if (!stillUsed.has(imgId)) {
      imageCache.delete(imgId)
    }
  }

  // 如果删除的任务在选中列表中，则移除
  const newSelection = selectedTaskIds.filter(id => !toDelete.has(id))
  if (newSelection.length !== selectedTaskIds.length) {
    useStore.getState().setSelectedTaskIds(newSelection)
  }

  showToast(`已删除 ${taskIds.length} 条记录`, 'success')
}

/** 删除单条任务 */
export async function removeTask(task: TaskRecord) {
  const { tasks, setTasks, inputImages, showToast } = useStore.getState()

  // 收集此任务关联的图片
  const taskImageIds = new Set([
    ...(task.inputImageIds || []),
    ...(task.maskImageId ? [task.maskImageId] : []),
    ...(task.outputImages || []),
  ])

  // 从列表移除
  const remaining = tasks.filter((t) => t.id !== task.id)
  setTasks(remaining)
  await deleteRemoteTask(task.id)

  // 找出其他任务仍引用的图片
  const stillUsed = new Set<string>()
  for (const t of remaining) {
    for (const id of t.inputImageIds || []) stillUsed.add(id)
    if (t.maskImageId) stillUsed.add(t.maskImageId)
    for (const id of t.outputImages || []) stillUsed.add(id)
  }
  for (const img of inputImages) stillUsed.add(img.id)

  // 删除孤立图片
  for (const imgId of taskImageIds) {
    if (!stillUsed.has(imgId)) {
      imageCache.delete(imgId)
    }
  }

  showToast('记录已删除', 'success')
}


/** 添加图片到输入（文件上传）—— 仅放入内存缓存，不写 IndexedDB */
export async function addImageFromFile(file: File): Promise<void> {
  if (!file.type.startsWith('image/')) return
  const dataUrl = await fileToDataUrl(file)
  const id = await hashDataUrl(dataUrl)
  imageCache.set(id, dataUrl)
  useStore.getState().addInputImage({ id, dataUrl })
}

/** 添加图片到输入（右键菜单）—— 支持 data/blob/http URL */
export async function addImageFromUrl(src: string): Promise<void> {
  const res = await fetch(src)
  const blob = await res.blob()
  if (!blob.type.startsWith('image/')) throw new Error('不是有效的图片')
  const dataUrl = await blobToDataUrl(blob)
  const id = await hashDataUrl(dataUrl)
  imageCache.set(id, dataUrl)
  useStore.getState().addInputImage({ id, dataUrl })
}

function fileToDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

function blobToDataUrl(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = reject
    reader.readAsDataURL(blob)
  })
}
