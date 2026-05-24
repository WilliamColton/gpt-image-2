import { useEffect, useRef, useState } from 'react'
import { AlertCircle, Check, Clock, Edit3, ImageIcon, Loader2, RotateCcw, Star, Trash2, X } from 'lucide-react'
import type { TaskRecord } from '../types'
import { useStore, getCachedImage, ensureImageCached, updateTaskInStore } from '../store'
import { formatImageRatio } from '../lib/size'
import { ParamValue } from '../lib/paramDisplay'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { Card } from './ui/card'

interface Props {
  task: TaskRecord
  onReuse: () => void
  onEditOutputs: () => void
  onDelete: () => void
  onClick: (e: React.MouseEvent | React.TouchEvent) => void
  isSelected?: boolean
}

export default function TaskCard({
  task,
  onReuse,
  onEditOutputs,
  onDelete,
  onClick,
  isSelected,
}: Props) {
  const [thumbSrc, setThumbSrc] = useState<string>('')
  const [coverRatio, setCoverRatio] = useState<string>('')
  const [coverSize, setCoverSize] = useState<string>('')
  const [now, setNow] = useState(Date.now())
  const [swipeOffset, setSwipeOffset] = useState(0)
  const [isSwiping, setIsSwiping] = useState(false)
  const [swipeStartedSelected, setSwipeStartedSelected] = useState(false)
  const [swipeActionActive, setSwipeActionActive] = useState(false)
  const toggleTaskSelection = useStore((s) => s.toggleTaskSelection)
  const touchStartRef = useRef<{ x: number; y: number } | null>(null)
  const swipeResetTimerRef = useRef<number | null>(null)
  const suppressClickUntilRef = useRef(0)
  const horizontalSwipeRef = useRef(false)

  const handleTouchStart = (e: React.TouchEvent) => {
    if (swipeResetTimerRef.current != null) {
      window.clearTimeout(swipeResetTimerRef.current)
      swipeResetTimerRef.current = null
    }
    touchStartRef.current = { x: e.touches[0].clientX, y: e.touches[0].clientY }
    horizontalSwipeRef.current = false
    setSwipeStartedSelected(Boolean(isSelected))
    setSwipeActionActive(false)
    setIsSwiping(true)
  }

  const handleTouchMove = (e: React.TouchEvent) => {
    if (!touchStartRef.current) return
    const deltaX = e.touches[0].clientX - touchStartRef.current.x
    const deltaY = e.touches[0].clientY - touchStartRef.current.y

    if (Math.abs(deltaX) > Math.abs(deltaY) && Math.abs(deltaX) > 10) {
      horizontalSwipeRef.current = true
      e.preventDefault()
      const boundedOffset = Math.max(-60, Math.min(60, deltaX))
      setSwipeOffset(boundedOffset)
      setSwipeActionActive(Math.abs(deltaX) >= 40)
    }
  }

  const handleTouchEnd = (e: React.TouchEvent) => {
    setIsSwiping(false)
    setSwipeOffset(0)

    if (!touchStartRef.current) return
    const deltaX = e.changedTouches[0].clientX - touchStartRef.current.x
    touchStartRef.current = null
    const isSwipeAction = horizontalSwipeRef.current && Math.abs(deltaX) > 40
    horizontalSwipeRef.current = false
    setSwipeActionActive(isSwipeAction)
    swipeResetTimerRef.current = window.setTimeout(() => {
      setSwipeActionActive(false)
      swipeResetTimerRef.current = null
    }, 220)

    if (isSwipeAction) {
      suppressClickUntilRef.current = Date.now() + 350
      e.preventDefault()
      e.stopPropagation()
      toggleTaskSelection(task.id)
    }
  }

  const handleTouchCancel = () => {
    touchStartRef.current = null
    horizontalSwipeRef.current = false
    setIsSwiping(false)
    setSwipeOffset(0)
    setSwipeActionActive(false)
  }

  useEffect(() => () => {
    if (swipeResetTimerRef.current != null) {
      window.clearTimeout(swipeResetTimerRef.current)
    }
  }, [])

  useEffect(() => {
    if (task.status !== 'running' && task.status !== 'queued') return
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [task.status])

  useEffect(() => {
    setCoverRatio('')
    setCoverSize('')

    if (task.outputImages?.[0]) {
      const cached = getCachedImage(task.outputImages[0])
      if (cached) {
        setThumbSrc(cached)
      } else {
        ensureImageCached(task.outputImages[0]).then((url) => {
          if (url) setThumbSrc(url)
        })
      }
    }
  }, [task.outputImages])

  useEffect(() => {
    if (!thumbSrc) return

    let cancelled = false
    const image = new Image()
    image.onload = () => {
      if (!cancelled && image.naturalWidth > 0 && image.naturalHeight > 0) {
        setCoverRatio(formatImageRatio(image.naturalWidth, image.naturalHeight))
        setCoverSize(`${image.naturalWidth}×${image.naturalHeight}`)
      }
    }
    image.src = thumbSrc
    if (image.complete && image.naturalWidth > 0 && image.naturalHeight > 0) {
      setCoverRatio(formatImageRatio(image.naturalWidth, image.naturalHeight))
      setCoverSize(`${image.naturalWidth}×${image.naturalHeight}`)
    }

    return () => {
      cancelled = true
    }
  }, [thumbSrc])

  const duration = (() => {
    let seconds: number
    if (task.status === 'running' || task.status === 'queued') {
      seconds = Math.floor((now - task.createdAt) / 1000)
    } else if (task.elapsed != null) {
      seconds = Math.floor(task.elapsed / 1000)
    } else {
      return '00:00'
    }
    const mm = String(Math.floor(seconds / 60)).padStart(2, '0')
    const ss = String(seconds % 60).padStart(2, '0')
    return `${mm}:${ss}`
  })()
  const aggregateActualParams = task.outputImages?.length
    ? { ...task.actualParams, n: task.outputImages.length }
    : task.actualParams
  const isSwipeReady = Math.abs(swipeOffset) >= 40
  const showSwipeAction = isSwipeReady || swipeActionActive
  const swipeBgClass = showSwipeAction
    ? swipeStartedSelected
      ? 'bg-gray-500 dark:bg-gray-600'
      : 'bg-blue-500'
    : 'bg-gray-200 dark:bg-gray-700'

  return (
    <div className="relative rounded-xl">
      <div
        className={`absolute inset-0 rounded-xl flex items-center transition-opacity duration-200 pointer-events-none ${
          isSwiping || swipeOffset || swipeActionActive ? 'opacity-100' : 'opacity-0'
        } ${swipeBgClass} ${
          swipeOffset > 0 ? 'justify-start pl-6' : 'justify-end pr-6'
        }`}
      >
        {swipeStartedSelected && showSwipeAction ? (
          <X className={`h-8 w-8 transition-transform duration-150 ${showSwipeAction ? 'scale-110 text-white' : 'scale-90 text-white/60'}`} />
        ) : (
          <Check className={`h-8 w-8 transition-transform duration-150 ${showSwipeAction ? 'scale-110 text-white' : 'scale-90 text-white/60'}`} strokeWidth={3} />
        )}
      </div>

      <Card
        className={`relative overflow-hidden rounded-xl cursor-pointer duration-200 hover:shadow-lg dark:hover:bg-gray-800/80 ${
          !isSwiping ? 'transition-[box-shadow,border-color,background-color,transform]' : 'transition-[box-shadow,border-color,background-color]'
        } ${
          task.status === 'running'
            ? 'border-blue-400 generating'
            : task.status === 'queued'
            ? 'border-yellow-400'
            : isSelected
            ? 'border-blue-500 shadow-md ring-2 ring-blue-500/50'
            : 'hover:border-gray-300 dark:hover:border-white/[0.18]'
        }`}
        style={{
          transform: swipeOffset ? `translateX(${swipeOffset}px)` : undefined,
        }}
        onClick={(e) => {
          if (Date.now() < suppressClickUntilRef.current) {
            e.preventDefault()
            e.stopPropagation()
            return
          }
          onClick(e)
        }}
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
        onTouchCancel={handleTouchCancel}
      >
        {isSelected && (
          <div className="absolute right-2 top-2 z-10 flex h-5 w-5 items-center justify-center rounded-full bg-blue-500 shadow-sm">
            <Check className="h-3 w-3 text-white" strokeWidth={3} />
          </div>
        )}
        <div className="flex h-40">
          <div className="w-40 min-w-[10rem] h-full bg-gray-100 dark:bg-black/20 relative flex items-center justify-center overflow-hidden flex-shrink-0">
            {task.status === 'running' && (
              <div className="flex flex-col items-center gap-2">
                <Loader2 className="w-8 h-8 text-blue-400 animate-spin" />
                <span className="text-xs text-gray-400 dark:text-gray-500">生成中...</span>
              </div>
            )}
            {task.status === 'queued' && (
              <div className="flex flex-col items-center gap-2">
                <Clock className="w-8 h-8 text-yellow-400" />
                <span className="text-xs text-yellow-400 dark:text-yellow-500">排队中...</span>
              </div>
            )}
            {task.status === 'error' && (
              <div className="flex flex-col items-center gap-1 px-2">
                <AlertCircle className="w-7 h-7 text-red-400" />
                <span className="text-xs text-red-400 text-center leading-tight">失败</span>
              </div>
            )}
            {task.status === 'done' && thumbSrc && (
              <>
                <img src={thumbSrc} className="w-full h-full object-cover" loading="lazy" alt="" />
                {task.outputImages.length > 1 && (
                  <Badge className="absolute bottom-1 right-1 border-0 bg-black/60 px-1.5 py-0.5 text-xs text-white">
                    {task.outputImages.length}
                  </Badge>
                )}
              </>
            )}
            {task.status === 'done' && !thumbSrc && (
              <ImageIcon className="h-8 w-8 text-gray-300" strokeWidth={1.5} />
            )}
            <div className="absolute top-1.5 left-1.5 flex items-center gap-1">
              {task.status !== 'done' || !coverRatio || !coverSize ? (
                <span className="flex items-center gap-1 bg-black/50 text-white text-[10px] sm:text-xs px-1.5 py-0.5 rounded backdrop-blur-sm font-mono">
                  <Clock className="h-3 w-3" />
                  {duration}
                </span>
              ) : (
                <>
                  <span className="bg-black/50 text-white text-[10px] sm:text-xs px-1.5 py-0.5 rounded backdrop-blur-sm font-mono">
                    {coverRatio}
                  </span>
                  <span className="bg-black/50 text-white/90 text-[10px] sm:text-xs px-1.5 py-0.5 rounded backdrop-blur-sm font-medium">
                    {coverSize}
                  </span>
                </>
              )}
            </div>
          </div>

          <div className="flex-1 p-3 flex flex-col min-w-0">
            <div className="flex-1 min-h-0 mb-2">
              <p className="text-sm text-gray-700 dark:text-gray-300 leading-relaxed line-clamp-3">
                {task.prompt || '(无提示词)'}
              </p>
            </div>
            <div className="mt-auto flex flex-col gap-1.5">
              <div className="flex overflow-x-auto hide-scrollbar gap-1.5 whitespace-nowrap mask-edge-r min-w-0 pr-2">
                <ParamValue task={task} paramKey="size" className="text-xs px-1.5 py-0.5 rounded flex-shrink-0" />
                <ParamValue task={task} paramKey="output_format" className="text-xs px-1.5 py-0.5 rounded flex-shrink-0" />
                <ParamValue task={task} paramKey="n" className="text-xs px-1.5 py-0.5 rounded flex-shrink-0" actualParams={aggregateActualParams} />
                {task.maskImageId && (
                  <Badge className="flex-shrink-0 px-1.5 py-0.5 text-xs">mask</Badge>
                )}
              </div>
              <div className="flex flex-shrink-0 justify-end gap-1" onClick={(e) => e.stopPropagation()}>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => updateTaskInStore(task.id, { isFavorite: !task.isFavorite })}
                  className={`h-7 w-7 ${task.isFavorite ? 'text-yellow-400 hover:bg-yellow-50 dark:hover:bg-yellow-500/10' : 'text-gray-400 hover:text-yellow-400 hover:bg-yellow-50 dark:hover:bg-yellow-500/10'}`}
                  title={task.isFavorite ? '取消收藏' : '收藏记录'}
                >
                  <Star className="h-4 w-4" fill={task.isFavorite ? 'currentColor' : 'none'} />
                </Button>
                <Button type="button" variant="ghost" size="icon" onClick={onReuse} className="h-7 w-7 text-gray-400 hover:text-blue-500" title="复用配置">
                  <RotateCcw className="h-4 w-4" />
                </Button>
                <Button type="button" variant="ghost" size="icon" onClick={onEditOutputs} disabled={!task.outputImages?.length} className="h-7 w-7 text-gray-400 hover:text-green-500" title="编辑输出">
                  <Edit3 className="h-4 w-4" />
                </Button>
                <Button type="button" variant="ghost" size="icon" onClick={onDelete} className="h-7 w-7 text-gray-400 hover:text-red-500" title="删除记录">
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>
        </div>
      </Card>
    </div>
  )
}
