import { useState } from 'react'
import { adminLogin } from './adminApi'

interface Props {
  onLogin: () => void
}

export default function AdminLogin({ onLogin }: Props) {
  const [apikey, setApikey] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const submit = async (event: React.FormEvent) => {
    event.preventDefault()
    setLoading(true)
    setError('')
    try {
      await adminLogin(apikey)
      onLogin()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-950 flex items-center justify-center p-4">
      <form onSubmit={submit} className="w-full max-w-md rounded-3xl border border-white/20 bg-white p-6 shadow-2xl dark:bg-gray-900">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">管理后台登录</h2>
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
          请输入管理员密钥以访问管理后台。
        </p>
        <input
          value={apikey}
          onChange={(e) => setApikey(e.target.value)}
          type="password"
          autoFocus
          placeholder="请输入管理员密钥"
          className="mt-5 w-full rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm text-gray-800 outline-none transition focus:border-blue-400 dark:border-white/[0.08] dark:bg-white/[0.04] dark:text-gray-100"
        />
        {error && <div className="mt-3 rounded-xl bg-red-50 px-3 py-2 text-sm text-red-500 dark:bg-red-500/10 dark:text-red-300">{error}</div>}
        <button
          type="submit"
          disabled={loading || !apikey.trim()}
          className="mt-5 w-full rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? '登录中...' : '登录'}
        </button>
        <a href="/" className="mt-3 block text-center text-sm text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
          返回首页
        </a>
      </form>
    </div>
  )
}
