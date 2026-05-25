import { useState } from 'react'
import { Bell, BookOpen, Bug, HelpCircle, Palette, Settings } from 'lucide-react'
import { useStore } from '../store'
import AnnouncementModal from './AnnouncementModal'
import AppearanceModal from './AppearanceModal'
import FeedbackModal from './FeedbackModal'
import HelpModal from './HelpModal'

export default function Header() {
  const setShowSettings = useStore((s) => s.setShowSettings)
  const announcement = useStore((s) => s.announcement)
  const latestChangelog = useStore((s) => s.latestChangelog)
  const setShowChangelog = useStore((s) => s.setShowChangelog)
  const authUser = useStore((s) => s.authUser)
  const [showHelp, setShowHelp] = useState(false)
  const [showAnnouncement, setShowAnnouncement] = useState(false)
  const [showAppearance, setShowAppearance] = useState(false)
  const [showFeedback, setShowFeedback] = useState(false)
  const hasAnnouncement = Boolean(announcement?.enabled && announcement.content.trim())
  const version = latestChangelog?.published ? latestChangelog.version.trim() : ''

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
              <BookOpen className="w-5 h-5 text-gray-600 dark:text-gray-400" />
            </button>
          )}
          <button
            onClick={() => setShowFeedback(true)}
            className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-900 transition-colors"
            title="Bug 反馈"
          >
            <Bug className="w-5 h-5 text-gray-600 dark:text-gray-400" />
          </button>
          <button
            onClick={() => setShowAnnouncement(true)}
            disabled={!hasAnnouncement}
            className={`p-2 rounded-lg transition-colors ${
              hasAnnouncement
                ? 'hover:bg-gray-100 dark:hover:bg-gray-900 text-gray-600 dark:text-gray-400'
                : 'cursor-not-allowed text-gray-300 dark:text-gray-700'
            }`}
            title={hasAnnouncement ? '站点公告' : '暂无公告'}
          >
            <Bell className="w-5 h-5" />
          </button>
          <button
            onClick={() => setShowAppearance(true)}
            className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-900 transition-colors"
            title="外观"
          >
            <Palette className="w-5 h-5 text-gray-600 dark:text-gray-400" />
          </button>
          <button
            onClick={() => setShowHelp(true)}
            className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-900 transition-colors"
            title="操作指南"
          >
            <HelpCircle className="w-5 h-5 text-gray-600 dark:text-gray-400" />
          </button>
          {authUser && (
            <span className="max-w-28 truncate px-2 text-sm text-gray-600 dark:text-gray-400">
              {authUser.username || authUser.label || '用户'}
            </span>
          )}
          <button
            onClick={() => setShowSettings(true)}
            className="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-900 transition-colors"
            title="设置"
          >
            <Settings className="w-5 h-5 text-gray-600 dark:text-gray-400" />
          </button>
        </div>
      </div>
      {showFeedback && <FeedbackModal onClose={() => setShowFeedback(false)} />}
      {showHelp && <HelpModal onClose={() => setShowHelp(false)} />}
      {showAppearance && <AppearanceModal onClose={() => setShowAppearance(false)} />}
      {showAnnouncement && <AnnouncementModal mode="manual" onClose={() => setShowAnnouncement(false)} />}
    </header>
  )
}
