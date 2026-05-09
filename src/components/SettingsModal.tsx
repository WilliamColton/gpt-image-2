import { useState } from 'react'
import { useStore, logout } from '../store'
import { useCloseOnEscape } from '../hooks/useCloseOnEscape'
import { redeemCode, getMe } from '../lib/backendApi'

export default function SettingsModal() {
  const showSettings = useStore((s) => s.showSettings)
  const setShowSettings = useStore((s) => s.setShowSettings)
  const authUser = useStore((s) => s.authUser)
  const setConfirmDialog = useStore((s) => s.setConfirmDialog)

  const [redeemValue, setRedeemValue] = useState('')
  const [redeemLoading, setRedeemLoading] = useState(false)
  const [redeemError, setRedeemError] = useState('')
  const [redeemSuccess, setRedeemSuccess] = useState('')

  const handleClose = () => setShowSettings(false)
  useCloseOnEscape(showSettings, handleClose)

  const handleRedeem = async () => {
    if (!redeemValue.trim()) return
    setRedeemLoading(true)
    setRedeemError('')
    setRedeemSuccess('')
    try {
      await redeemCode(redeemValue.trim())
      setRedeemSuccess('兑换成功')
      setRedeemValue('')
      // Refresh user info
      const { user } = await getMe()
      useStore.getState().setAuthUser(user)
    } catch (err) {
      setRedeemError(err instanceof Error ? err.message : String(err))
    } finally {
      setRedeemLoading(false)
    }
  }

  if (!showSettings) return null

  const quotaDisplay = authUser
    ? authUser.quota === 0
      ? `${authUser.usedCount} / 无限制`
      : `${authUser.usedCount} / ${authUser.quota}`
    : ''

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/30 backdrop-blur-sm animate-overlay-in" onClick={handleClose} />
      <div className="relative z-10 w-full max-w-md rounded-3xl border border-white/50 bg-white/95 p-5 shadow-2xl ring-1 ring-black/5 animate-modal-in dark:border-white/[0.08] dark:bg-gray-900/95 dark:ring-white/10 overflow-y-auto max-h-[85vh] custom-scrollbar">
        <div className="mb-5 flex items-center justify-between gap-4">
          <h3 className="text-base font-semibold text-gray-800 dark:text-gray-100 flex items-center gap-2">设置</h3>
          <button
            onClick={handleClose}
            className="rounded-full p-1 text-gray-400 transition hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-white/[0.06] dark:hover:text-gray-200"
            aria-label="关闭"
          >
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="space-y-6">
          <section>
            <div className="rounded-2xl border border-gray-200/70 bg-gray-50/70 p-4 text-sm text-gray-600 dark:border-white/[0.08] dark:bg-white/[0.03] dark:text-gray-300">
              <div className="flex items-center justify-between">
                <span className="text-gray-400">已生成的图片数</span>
                <span>{authUser?.imageCount ?? 0} 张</span>
              </div>
              <div className="mt-2 flex items-center justify-between">
                <span className="text-gray-400">配额</span>
                <span>{quotaDisplay}</span>
              </div>
            </div>
          </section>

          <section>
            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">兑换码</h4>
            <div className="flex gap-2">
              <input
                value={redeemValue}
                onChange={(e) => setRedeemValue(e.target.value)}
                placeholder="输入兑换码增加配额"
                className="flex-1 rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm text-gray-800 outline-none transition focus:border-blue-400 dark:border-white/[0.08] dark:bg-white/[0.04] dark:text-gray-100"
              />
              <button
                onClick={handleRedeem}
                disabled={!redeemValue.trim() || redeemLoading}
                className="rounded-xl bg-blue-600 px-4 py-2 text-sm text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {redeemLoading ? '兑换中...' : '兑换'}
              </button>
            </div>
            {redeemError && <div className="mt-2 text-sm text-red-500 dark:text-red-400">{redeemError}</div>}
            {redeemSuccess && <div className="mt-2 text-sm text-green-600 dark:text-green-400">{redeemSuccess}</div>}
          </section>

          <section className="pt-6 border-t border-gray-100 dark:border-white/[0.08]">
            <button
              onClick={() =>
                setConfirmDialog({
                  title: '退出登录',
                  message: '确定要退出登录吗？',
                  action: () => logout(),
                })
              }
              className="w-full rounded-xl border border-gray-200/80 bg-gray-50/50 px-4 py-2.5 text-sm text-gray-500 transition hover:bg-gray-100/80 dark:border-white/[0.08] dark:bg-white/[0.03] dark:text-gray-400 dark:hover:bg-white/[0.06]"
            >
              退出登录
            </button>
          </section>
        </div>
      </div>
    </div>
  )
}
