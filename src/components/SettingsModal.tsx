import { useEffect, useState } from 'react'
import { X } from 'lucide-react'
import { useStore, logout } from '../store'
import { useCloseOnEscape } from '../hooks/useCloseOnEscape'
import { redeemCode, getMe, setInviteCode, getInviteCode, changePassword } from '../lib/backendApi'
import type { ThemeMode } from '../types'
import { Separator } from './ui/separator'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Button } from './ui/button'

export default function SettingsModal() {
  const showSettings = useStore((s) => s.showSettings)
  const setShowSettings = useStore((s) => s.setShowSettings)
  const authUser = useStore((s) => s.authUser)
  const settings = useStore((s) => s.settings)
  const setSettings = useStore((s) => s.setSettings)
  const setConfirmDialog = useStore((s) => s.setConfirmDialog)
  const showToast = useStore((s) => s.showToast)

  const [redeemValue, setRedeemValue] = useState('')
  const [redeemLoading, setRedeemLoading] = useState(false)
  const [redeemError, setRedeemError] = useState('')
  const [redeemSuccess, setRedeemSuccess] = useState('')

  // Invite code section state
  const [inviteCode, setInviteCode] = useState<string | null>(null)
  const [showModifyInvite, setShowModifyInvite] = useState(false)
  const [newInviteCode, setNewInviteCode] = useState('')
  const [inviteError, setInviteError] = useState('')
  const [inviteSuccess, setInviteSuccess] = useState('')

  // Change password section state
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmNewPassword, setConfirmNewPassword] = useState('')
  const [pwLoading, setPwLoading] = useState(false)
  const [pwError, setPwError] = useState('')
  const [pwSuccess, setPwSuccess] = useState('')

  // Fetch invite code when settings modal opens
  useEffect(() => {
    if (showSettings) {
      getInviteCode()
        .then((res) => setInviteCode(res.code))
        .catch(() => setInviteCode(null))
    }
  }, [showSettings])

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

  const handleCopyInvite = () => {
    if (inviteCode) {
      navigator.clipboard.writeText(inviteCode)
      showToast('已复制', 'success')
    }
  }

  const handleSaveInvite = async () => {
    if (!newInviteCode.trim()) return
    setInviteError('')
    setInviteSuccess('')
    try {
      await setInviteCode(newInviteCode.trim())
      setInviteCode(newInviteCode.trim())
      setShowModifyInvite(false)
      setNewInviteCode('')
      showToast('邀请码已更新', 'success')
    } catch (err) {
      setInviteError(err instanceof Error ? err.message : String(err))
    }
  }

  const handleChangePassword = async () => {
    setPwError('')
    setPwSuccess('')
    if (newPassword.length < 8) {
      setPwError('密码至少需要 8 个字符')
      return
    }
    if (newPassword !== confirmNewPassword) {
      setPwError('两次输入的密码不一致')
      return
    }
    setPwLoading(true)
    try {
      await changePassword(oldPassword, newPassword, confirmNewPassword)
      setPwSuccess('密码已修改')
      setOldPassword('')
      setNewPassword('')
      setConfirmNewPassword('')
      showToast('密码已修改', 'success')
    } catch (err) {
      setPwError(err instanceof Error ? err.message : String(err))
    } finally {
      setPwLoading(false)
    }
  }

  if (!showSettings) return null

  const quotaDisplay = authUser
    ? authUser.quota === 0
      ? `${authUser.usedCount} / 无限制`
      : `${authUser.usedCount} / ${authUser.quota}`
    : ''

  const themeOptions: Array<{ value: ThemeMode; label: string }> = [
    { value: 'system', label: '跟随系统' },
    { value: 'light', label: '浅色' },
    { value: 'dark', label: '深色' },
  ]

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
            <X className="h-5 w-5" />
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
            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">外观</h4>
            <div className="grid grid-cols-3 gap-2 rounded-2xl bg-gray-100/70 p-1 dark:bg-white/[0.04]">
              {themeOptions.map((option) => {
                const active = settings.theme === option.value
                return (
                  <button
                    key={option.value}
                    type="button"
                    onClick={() => setSettings({ theme: option.value })}
                    className={`rounded-xl px-3 py-1.5 text-xs font-medium transition ${
                      active
                        ? 'bg-white text-gray-900 shadow-sm dark:bg-white/[0.12] dark:text-gray-100'
                        : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
                    }`}
                  >
                    {option.label}
                  </button>
                )
              })}
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

          {/* Invite Code Section */}
          <section className="space-y-3">
            <Separator />
            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300">邀请码</h4>
            {inviteCode ? (
              <div className="flex items-center gap-2">
                <code className="flex-1 font-mono text-xs bg-white/50 dark:bg-white/[0.04] px-3 py-2 rounded-xl border border-gray-200 dark:border-white/[0.08] truncate">{inviteCode}</code>
                <Button variant="outline" size="sm" onClick={handleCopyInvite}>复制</Button>
                <Button variant="ghost" size="sm" onClick={() => setShowModifyInvite(true)}>修改</Button>
              </div>
            ) : (
              <div className="text-sm text-gray-400 dark:text-gray-500">未设置</div>
            )}
            {showModifyInvite && (
              <div className="flex gap-2">
                <Input type="text" placeholder="输入新的邀请码" value={newInviteCode} onChange={(e) => setNewInviteCode(e.target.value)} />
                <Button variant="default" size="sm" onClick={handleSaveInvite}>确认</Button>
                <Button variant="ghost" size="sm" onClick={() => setShowModifyInvite(false)}>取消</Button>
              </div>
            )}
            {inviteError && <div className="text-sm text-red-500 dark:text-red-400">{inviteError}</div>}
            {inviteSuccess && <div className="text-sm text-green-600 dark:text-green-400">{inviteSuccess}</div>}
          </section>

          {/* Change Password Section */}
          <section className="space-y-3">
            <Separator />
            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300">修改密码</h4>
            <div className="space-y-3">
              <Input type="password" placeholder="输入旧密码" value={oldPassword} onChange={(e) => setOldPassword(e.target.value)} />
              <Input type="password" placeholder="至少 8 字符" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} />
              <Input type="password" placeholder="再次输入新密码" value={confirmNewPassword} onChange={(e) => setConfirmNewPassword(e.target.value)} />
            </div>
            <Button onClick={handleChangePassword} disabled={pwLoading} className="w-full">
              {pwLoading ? '修改中...' : '修改密码'}
            </Button>
            {pwError && <div className="text-sm text-red-500 dark:text-red-400">{pwError}</div>}
            {pwSuccess && <div className="text-sm text-green-600 dark:text-green-400">{pwSuccess}</div>}
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
