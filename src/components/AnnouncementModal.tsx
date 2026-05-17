import { useState } from 'react'
import { useStore } from '../store'

export default function AnnouncementModal() {
  const announcement = useStore((s) => s.announcement)
  const [dismissedAnnouncementUpdatedAt, setDismissedAnnouncementUpdatedAt] = useState<number | null>(null)

  if (!announcement?.enabled) return null
  if (!announcement.content.trim()) return null
  if (announcement.updatedAt === dismissedAnnouncementUpdatedAt) return null

  return (
    <div className="fixed inset-0 z-[130] flex items-center justify-center bg-gray-950/70 p-4 backdrop-blur-sm">
      <div className="w-full max-w-lg rounded-3xl border border-white/20 bg-white p-6 shadow-2xl dark:bg-gray-900">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">公告</h2>
        <div className="mt-4 max-h-[60vh] overflow-y-auto whitespace-pre-wrap text-sm leading-relaxed text-gray-600 dark:text-gray-300 custom-scrollbar">
          {announcement.content}
        </div>
        <button
          type="button"
          onClick={() => setDismissedAnnouncementUpdatedAt(announcement.updatedAt)}
          className="mt-5 w-full rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-blue-700"
        >
          我知道了
        </button>
      </div>
    </div>
  )
}
