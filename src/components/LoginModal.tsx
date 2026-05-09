import { useState } from 'react'
import { loginWithCode } from '../lib/backendApi'
import { bootstrapBackendSession, useStore } from '../store'

export default function LoginModal() {
  const [code, setCode] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const submit = async (event: React.FormEvent) => {
    event.preventDefault()
    setLoading(true)
    setError('')
    try {
      await loginWithCode(code)
      await bootstrapBackendSession()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-gray-950/70 p-4 backdrop-blur-sm">
      <form onSubmit={submit} className="w-full max-w-md rounded-3xl border border-white/20 bg-white p-6 shadow-2xl dark:bg-gray-900">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">输入兑换码登录</h2>
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
          输入兑换码即可注册并获得图片生成配额。已有账户的用户也可输入新兑换码来增加配额。
        </p>
        <input
          value={code}
          onChange={(e) => setCode(e.target.value)}
          type="text"
          autoFocus
          placeholder="请输入兑换码"
          className="mt-5 w-full rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm text-gray-800 outline-none transition focus:border-blue-400 dark:border-white/[0.08] dark:bg-white/[0.04] dark:text-gray-100"
        />
        {error && <div className="mt-3 rounded-xl bg-red-50 px-3 py-2 text-sm text-red-500 dark:bg-red-500/10 dark:text-red-300">{error}</div>}
        <button
          type="submit"
          disabled={loading || !code.trim()}
          className="mt-5 w-full rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? '登录中...' : '登录 / 注册'}
        </button>
      </form>
    </div>
  )
}
