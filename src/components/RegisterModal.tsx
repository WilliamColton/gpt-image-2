import { useState } from 'react'
import { register } from '../lib/backendApi'
import { bootstrapBackendSession, useStore } from '../store'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Button } from './ui/button'

interface RegisterModalProps {
  onClose: () => void
}

export default function RegisterModal({ onClose }: RegisterModalProps) {
  const [inviteCode, setInviteCode] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    const usernameChars = Array.from(username.trim())
    if (usernameChars.length < 3 || usernameChars.length > 20) {
      setError('用户名须为 3-20 个字符')
      return
    }
    if (password.length < 8) {
      setError('密码至少需要 8 个字符')
      return
    }

    setLoading(true)
    try {
      await register(inviteCode.trim(), username.trim(), password)
      await bootstrapBackendSession()
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4">
      <div
        className="absolute inset-0 bg-black/20 dark:bg-black/40 backdrop-blur-md animate-overlay-in"
        onClick={onClose}
      />
      <div
        className="relative z-10 w-full max-w-md rounded-3xl border border-white/50 bg-white/95 p-5 shadow-2xl ring-1 ring-black/5 animate-modal-in dark:border-white/[0.08] dark:bg-gray-900/95 dark:ring-white/10"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100">注册</h2>

        <form onSubmit={handleSubmit} className="mt-5 space-y-4">
          <div>
            <Label>邀请码（可选）</Label>
            <Input
              type="text"
              placeholder="选填，输入邀请码可获得额外配额"
              value={inviteCode}
              onChange={(e) => setInviteCode(e.target.value)}
              className="mt-1"
            />
          </div>
          <div>
            <Label>用户名</Label>
            <Input
              type="text"
              placeholder="3-20 字符，允许中文"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="mt-1"
            />
          </div>
          <div>
            <Label>密码</Label>
            <Input
              type="password"
              placeholder="至少 8 字符"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="mt-1"
            />
          </div>

          {error && (
            <div className="rounded-xl bg-red-50 px-3 py-2 text-sm text-red-500 dark:bg-red-500/10 dark:text-red-300">
              {error}
            </div>
          )}

          <Button type="submit" disabled={loading} className="w-full">
            {loading ? '注册中...' : '注册'}
          </Button>
        </form>
      </div>
    </div>
  )
}
