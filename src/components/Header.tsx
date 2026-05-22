import { useEffect, useRef, useState } from 'react'
import { useStore } from '../store'
import AnnouncementModal from './AnnouncementModal'
import FeedbackModal from './FeedbackModal'
import HelpModal from './HelpModal'

export default function Header() {
  const setShowSettings = useStore((s) => s.setShowSettings)
  const announcement = useStore((s) => s.announcement)
  const latestChangelog = useStore((s) => s.latestChangelog)
  const setShowChangelog = useStore((s) => s.setShowChangelog)
  const [showHelpMenu, setShowHelpMenu] = useState(false)
  const [showHelp, setShowHelp] = useState(false)
  const [showAnnouncement, setShowAnnouncement] = useState(false)
  const [showFeedback, setShowFeedback] = useState(false)
  const helpMenuRef = useRef<HTMLDivElement>(null)
  const hasAnnouncement = Boolean(announcement?.enabled && announcement.content.trim())
  const version = latestChangelog?.published ? latestChangelog.version.trim() : ''

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (helpMenuRef.current && !helpMenuRef.current.contains(e.target as Node)) {
        setShowHelpMenu(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  return (
    <header className="safe-area-top sticky top-0 z-40 bg-white/80 dark:bg-gray-950/80 backdrop-blur border-b border-gray-200 dark:border-white/[0.08]">
      <div className="safe-area-x safe-header-inner max-w-7xl mx-auto flex items-center justify-between">
        <div className="flex items-start gap-2">
          <h1 className="text-lg font-bold tracking-tight text-gray-800 dark:text-gray-100">
            GPT Image Playground
          </h1>
          {version && (
            <button
              onClick={() => setShowChangelog(true)}
              className="mt-0.5 rounded-full bg-blue-50 px-2 py-0.5 font-mono text-xs font-medium text-blue-600 transition hover:bg-blue-100 dark:bg-blue-500/10 dark:text-blue-400 dark:hover:bg-blue-500/20"
              title="查看更新日志"
            >
              v{version}
            </button>
          )}
        </div>
        <div className="flex items-center gap-1">
          {version && (
            <button
              onClick={() => setShowChangelog(true)}
              className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-900 transition-colors"
              title="更新日志"
            >
              <svg
                className="w-5 h-5 text-gray-600 dark:text-gray-400"
                fill="none"
                stroke="currentColor"
                strokeWidth={2}
                strokeLinecap="round"
                strokeLinejoin="round"
                viewBox="0 0 24 24"
              >
                <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" />
                <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z" />
                <path d="M8 7h8" />
                <path d="M8 11h6" />
              </svg>
            </button>
          )}
          <button
            onClick={() => setShowFeedback(true)}
            className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-900 transition-colors"
            title="Bug 反馈"
          >
            <svg
              className="w-5 h-5 text-gray-600 dark:text-gray-400"
              fill="none"
              stroke="currentColor"
              strokeWidth={2}
              strokeLinecap="round"
              strokeLinejoin="round"
              viewBox="0 0 24 24"
            >
              <path d="M8 8h8v8H8z" />
              <path d="M3 13h5" />
              <path d="M16 13h5" />
              <path d="M12 3v5" />
              <path d="M12 16v5" />
              <path d="M5.5 5.5 8 8" />
              <path d="m16 16 2.5 2.5" />
              <path d="M18.5 5.5 16 8" />
              <path d="M8 16l-2.5 2.5" />
            </svg>
          </button>
          <div ref={helpMenuRef} className="relative">
            <button
              onClick={(e) => {
                e.stopPropagation()
                setShowHelpMenu((v) => !v)
              }}
              className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-900 transition-colors"
              title="帮助与公告"
            >
              <svg
                className="w-5 h-5 text-gray-600 dark:text-gray-400"
                fill="none"
                stroke="currentColor"
                strokeWidth={2}
                strokeLinecap="round"
                strokeLinejoin="round"
                viewBox="0 0 24 24"
              >
                <circle cx="12" cy="12" r="10" />
                <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
                <path d="M12 17h.01" />
              </svg>
            </button>
            {showHelpMenu && (
              <div className="absolute right-0 top-full z-50 mt-1.5 w-36 overflow-hidden rounded-xl border border-gray-200/60 bg-white/95 py-1 shadow-[0_8px_30px_rgb(0,0,0,0.12)] ring-1 ring-black/5 backdrop-blur-xl animate-dropdown-down dark:border-white/[0.08] dark:bg-gray-900/95 dark:shadow-[0_8px_30px_rgb(0,0,0,0.3)] dark:ring-white/10">
                <button
                  type="button"
                  onClick={() => {
                    setShowHelpMenu(false)
                    setShowHelp(true)
                  }}
                  className="block w-full px-3 py-2 text-left text-xs text-gray-700 transition-colors hover:bg-gray-50 dark:text-gray-300 dark:hover:bg-white/[0.06]"
                >
                  操作指南
                </button>
                <button
                  type="button"
                  disabled={!hasAnnouncement}
                  onClick={() => {
                    if (!hasAnnouncement) return
                    setShowHelpMenu(false)
                    setShowAnnouncement(true)
                  }}
                  className={`block w-full px-3 py-2 text-left text-xs transition-colors ${
                    hasAnnouncement
                      ? 'text-gray-700 hover:bg-gray-50 dark:text-gray-300 dark:hover:bg-white/[0.06]'
                      : 'cursor-not-allowed text-gray-400 dark:text-gray-600'
                  }`}
                >
                  {hasAnnouncement ? '站点公告' : '暂无公告'}
                </button>
              </div>
            )}
          </div>
          <button
            onClick={() => setShowSettings(true)}
            className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-900 transition-colors"
            title="设置"
          >
            <svg
              className="w-5 h-5 text-gray-600 dark:text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
              />
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
              />
            </svg>
          </button>
        </div>
      </div>
      {showFeedback && <FeedbackModal onClose={() => setShowFeedback(false)} />}
      {showHelp && <HelpModal onClose={() => setShowHelp(false)} />}
      {showAnnouncement && <AnnouncementModal mode="manual" onClose={() => setShowAnnouncement(false)} />}
    </header>
  )
}
