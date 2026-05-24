import { useEffect, useState } from 'react'
import { X } from 'lucide-react'
import { useStore, logout } from '../store'
import { useCloseOnEscape } from '../hooks/useCloseOnEscape'
import { redeemCode, getMe, setInviteCode, getInviteCode, changePassword } from '../lib/backendApi'
import { Separator } from './ui/separator'
import { Input } from './ui/input'
import { Button } from './ui/button'
import { Dialog, DialogContent } from './ui/dialog'

export default function SettingsModal() {
  const showSettings = useStore((s) => s.showSettings)
  const setShowSettings = useStore((s) => s.setShowSettings)
  const authUser = useStore((s) => s.authUser)
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

  return (
    <Dialog open={showSettings} onOpenChange={(open) => { if (!open) setShowSettings(false) }}>
      <DialogContent className="max-w-md max-h-[85vh] overflow-y-auto custom-scrollbar" data-no-drag-select hideClose>
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
      </DialogContent>
    </Dialog>
  )
}
