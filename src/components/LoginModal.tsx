import { useState } from 'react'
import { loginWithApikey } from '../lib/backendApi'
import { bootstrapBackendSession, useStore } from '../store'

export default function LoginModal() {
  const [apikey, setApikey] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const openAIConfigured = useStore((s) => s.settings.openAIConfigured)

  const submit = async (event: React.FormEvent) => {
    event.preventDefault()
    setLoading(true)
    setError('')
    try {
      await loginWithApikey(apikey)
      useStore.getState().setSettings({ apiKey: apikey })
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
        <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">输入 apikey 登录</h2>
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
          如果这个 apikey 第一次使用，系统会自动注册并创建独立的数据空间。
        </p>
        <input
          value={apikey}
          onChange={(e) => setApikey(e.target.value)}
          type="text"
          autoFocus
          placeholder="请输入你的 apikey"
          className="mt-5 w-full rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm text-gray-800 outline-none transition focus:border-blue-400 dark:border-white/[0.08] dark:bg-white/[0.04] dark:text-gray-100"
        />
        {error && <div className="mt-3 rounded-xl bg-red-50 px-3 py-2 text-sm text-red-500 dark:bg-red-500/10 dark:text-red-300">{error}</div>}
        <button
          type="submit"
          disabled={loading || !apikey.trim()}
          className="mt-5 w-full rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? '登录中...' : '登录 / 自动注册'}
        </button>
        <div className="mt-3 text-xs text-gray-400 dark:text-gray-500">
          {openAIConfigured
            ? '后端已配置 API Key，无需在前端设置'
            : '后端未配置 API Key，请联系管理员配置环境变量 OPENAI_API_KEY'}
        </div>
      </form>
    </div>
  )
}
