import { beforeEach, describe, expect, it, vi } from 'vitest'
import { DEFAULT_PARAMS, DEFAULT_SETTINGS } from './types'
import type { TaskRecord } from './types'
import { editOutputs, submitTask, useStore } from './store'
import { submitGenerateTask, getTasks as fetchTasks, putRemoteTask, uploadImage, streamTaskStatus } from './lib/backendApi'

vi.mock('./lib/backendApi', () => ({
  submitGenerateTask: vi.fn().mockResolvedValue({ taskId: 'task-1', status: 'processing' }),
  submitEditTask: vi.fn().mockResolvedValue({ taskId: 'task-1', status: 'processing' }),
  putRemoteTask: vi.fn().mockResolvedValue({ ok: true }),
  uploadImage: vi.fn().mockResolvedValue({
    id: 'uploaded-1',
    dataUrl: 'data:image/png;base64,abc',
    createdAt: 1,
    source: 'generated' as const,
  }),
  getBackendToken: vi.fn().mockReturnValue('test-token'),
  getMe: vi.fn(),
  getTasks: vi.fn().mockResolvedValue({ tasks: [] }),
  getPublicConfig: vi.fn(),
  getPublicAnnouncement: vi.fn().mockResolvedValue(null),
  getLatestPublicChangelog: vi.fn().mockResolvedValue(null),
  getPublicChangelogEntries: vi.fn().mockResolvedValue({ changelogs: [] }),
  clearBackendToken: vi.fn(),
  clearRemoteTasks: vi.fn(),
  deleteRemoteTask: vi.fn(),
  streamTaskStatus: vi.fn().mockImplementation((_taskId: string, onUpdate: Function) => {
    // Simulate SSE: call onUpdate with done status on next tick
    setTimeout(() => {
      const task = useStore.getState().tasks.find((t: any) => t.id === _taskId)
      if (task) {
        onUpdate({ ...task, status: 'done', outputImages: ['img-1'], finishedAt: Date.now(), elapsed: 1000 })
      }
    }, 10)
    return new AbortController()
  }),
}))

vi.mock('./lib/db', () => ({
  putImage: vi.fn().mockResolvedValue(undefined),
  getImage: vi.fn().mockResolvedValue(null),
  hashDataUrl: vi.fn().mockResolvedValue('hash-1'),
}))

const imageA = { id: 'image-a', dataUrl: 'data:image/png;base64,a' }

function task(overrides: Partial<TaskRecord> = {}): TaskRecord {
  return {
    id: 'task-a',
    prompt: 'prompt',
    params: { ...DEFAULT_PARAMS },
    inputImageIds: [],
    maskTargetImageId: null,
    maskImageId: null,
    outputImages: [],
    status: 'done',
    error: null,
    createdAt: 1,
    finishedAt: 2,
    elapsed: 1,
    ...overrides,
  }
}

describe('announcement state in store', () => {
  beforeEach(() => {
    useStore.setState({ seenAnnouncementUpdatedAt: null })
  })

  it('marks the current announcement version as seen', () => {
    useStore.getState().markAnnouncementSeen(123)

    expect(useStore.getState().seenAnnouncementUpdatedAt).toBe(123)
  })
})

describe('mask draft lifecycle in store actions', () => {
  beforeEach(() => {
    useStore.setState({
      settings: { ...DEFAULT_SETTINGS },
      prompt: 'prompt',
      inputImages: [],
      maskDraft: null,
      maskEditorImageId: null,
      params: { ...DEFAULT_PARAMS },
      tasks: [],
      detailTaskId: null,
      lightboxImageId: null,
      lightboxImageList: [],
      showSettings: false,
      seenAnnouncementUpdatedAt: null,
      toast: null,
      confirmDialog: null,
      showToast: vi.fn(),
      setConfirmDialog: vi.fn(),
    })
  })

  it('clears an existing mask when quick edit-output adds outputs as references', async () => {
    useStore.setState({
      inputImages: [imageA],
      maskDraft: {
        targetImageId: imageA.id,
        maskDataUrl: 'data:image/png;base64,mask',
        updatedAt: 1,
      },
    })

    await editOutputs(task({ outputImages: [imageA.id] }))

    expect(useStore.getState().maskDraft).toBeNull()
  })

  it('clears an invalid mask draft when submit cannot find the mask target image', async () => {
    useStore.setState({
      authUser: { id: 'user-1', label: 'user', role: 'user', imageCount: 0, quota: 0, usedCount: 0 },
      inputImages: [imageA],
      maskDraft: {
        targetImageId: 'missing-image',
        maskDataUrl: 'data:image/png;base64,mask',
        updatedAt: 1,
      },
    })

    await submitTask()

    expect(useStore.getState().maskDraft).toBeNull()
  })
})

describe('submitTask backend submission flow', () => {
  beforeEach(() => {
    vi.mocked(submitGenerateTask).mockReset()
    vi.mocked(submitGenerateTask).mockResolvedValue({ taskId: 'task-1', status: 'processing' })
    vi.mocked(fetchTasks).mockReset()
    vi.mocked(fetchTasks).mockResolvedValue({ tasks: [] })
    vi.mocked(putRemoteTask).mockClear()
    vi.mocked(uploadImage).mockClear()

    useStore.setState({
      settings: { ...DEFAULT_SETTINGS, timeout: 60 },
      prompt: 'test prompt',
      inputImages: [],
      maskDraft: null,
      maskEditorImageId: null,
      params: { ...DEFAULT_PARAMS },
      tasks: [],
      detailTaskId: null,
      lightboxImageId: null,
      lightboxImageList: [],
      showSettings: false,
      seenAnnouncementUpdatedAt: null,
      toast: null,
      confirmDialog: null,
      showToast: vi.fn(),
      setConfirmDialog: vi.fn(),
      authUser: { id: 'user-1', label: 'test', role: 'user', imageCount: 0, quota: 0, usedCount: 0 },
    })
  })

  it('creates a queued task immediately', () => {
    submitTask()
    const tasks = useStore.getState().tasks
    expect(tasks.length).toBe(1)
    expect(tasks[0].status).toBe('queued')
    expect(tasks[0].prompt).toBe('test prompt')
  })

  it('calls submitGenerateTask to submit to backend', async () => {
    // Mock fetchTasks to return done on first poll
    vi.mocked(fetchTasks).mockResolvedValue({
      tasks: [{
        id: useStore.getState().tasks[0]?.id || 'task-1',
        status: 'done',
        outputImages: ['img-1'],
        prompt: 'test prompt',
        params: { ...DEFAULT_PARAMS },
        inputImageIds: [],
        maskTargetImageId: null,
        maskImageId: null,
        error: null,
        createdAt: Date.now(),
        finishedAt: Date.now(),
        elapsed: 1000,
      }],
    })

    submitTask()
    await vi.waitFor(() => {
      expect(submitGenerateTask).toHaveBeenCalledWith(
        expect.any(String),
        'test prompt',
        expect.objectContaining({}),
        [],
        false,
      )
    })
  })

  it('polls for completion and updates task to done', async () => {
    submitTask()

    await vi.waitFor(() => {
      const tasks = useStore.getState().tasks
      expect(tasks[0].status).toBe('done')
    }, { timeout: 5000 })
  })

  it('updates task to error when backend returns error status', async () => {
    // Override SSE mock to return error
    vi.mocked(streamTaskStatus).mockImplementation((_taskId: string, onUpdate: Function) => {
      setTimeout(() => {
        const task = useStore.getState().tasks.find((t: any) => t.id === _taskId)
        if (task) {
          onUpdate({ ...task, status: 'error', error: 'Generation failed', outputImages: [], finishedAt: Date.now(), elapsed: 1000 })
        }
      }, 10)
      return new AbortController()
    })

    submitTask()

    await vi.waitFor(() => {
      const tasks = useStore.getState().tasks
      expect(tasks[0].status).toBe('error')
      expect(tasks[0].error).toBe('Generation failed')
    }, { timeout: 5000 })
  })

  it('does not submit when prompt is empty', async () => {
    useStore.setState({ prompt: '   ' })

    submitTask()

    expect(submitGenerateTask).not.toHaveBeenCalled()
    expect(useStore.getState().tasks.length).toBe(0)
  })

  it('does not submit when not authenticated', async () => {
    useStore.setState({ authUser: null })

    submitTask()

    expect(submitGenerateTask).not.toHaveBeenCalled()
    expect(useStore.getState().tasks.length).toBe(0)
  })
})
