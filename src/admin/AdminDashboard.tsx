import { useState, useEffect, useCallback } from 'react'
import { adminListUsers, adminUpdateQuota, adminToggleStatus, clearAdminToken, type AdminUser } from './adminApi'

interface Props {
  onLogout: () => void
}

export default function AdminDashboard({ onLogout }: Props) {
  const [users, setUsers] = useState<AdminUser[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [quotaInputs, setQuotaInputs] = useState<Record<string, string>>({})

  const loadUsers = useCallback(async () => {
    try {
      const { users } = await adminListUsers()
      setUsers(users)
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadUsers() }, [loadUsers])

  const handleQuotaDelta = async (userId: string, delta: number) => {
    try {
      await adminUpdateQuota(userId, delta)
      await loadUsers()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  const handleReset = async (userId: string) => {
    try {
      await adminUpdateQuota(userId, 0, true)
      await loadUsers()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  const handleToggleStatus = async (userId: string, currentStatus: string) => {
    const newStatus = currentStatus === 'active' ? 'disabled' : 'active'
    try {
      await adminToggleStatus(userId, newStatus)
      await loadUsers()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  const handleLogout = () => {
    clearAdminToken()
    onLogout()
  }

  const formatTime = (ms: number) => {
    return new Date(ms).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
  }

  const getQuotaDisplay = (user: AdminUser) => {
    if (user.quota === 0) return `${user.usedCount} / 无限制`
    return `${user.usedCount} / ${user.quota}`
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-950 flex items-center justify-center">
        <div className="text-gray-400">加载中...</div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100">
      <header className="border-b border-white/10 px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <h1 className="text-lg font-semibold">管理后台</h1>
          <div className="flex items-center gap-4">
            <button onClick={loadUsers} className="text-sm text-gray-400 hover:text-gray-200">刷新</button>
            <button onClick={handleLogout} className="text-sm text-red-400 hover:text-red-300">退出登录</button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-6">
        {error && (
          <div className="mb-4 rounded-xl bg-red-500/10 px-4 py-3 text-sm text-red-300">{error}</div>
        )}

        <div className="overflow-x-auto rounded-2xl border border-white/10 bg-white/[0.03]">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-white/10 text-left text-gray-400">
                <th className="px-4 py-3 font-medium">用户</th>
                <th className="px-4 py-3 font-medium">注册时间</th>
                <th className="px-4 py-3 font-medium">配额</th>
                <th className="px-4 py-3 font-medium">状态</th>
                <th className="px-4 py-3 font-medium">配额操作</th>
                <th className="px-4 py-3 font-medium">状态操作</th>
              </tr>
            </thead>
            <tbody>
              {users.map(user => (
                <tr key={user.id} className="border-b border-white/5 hover:bg-white/[0.02]">
                  <td className="px-4 py-3">
                    <div className="font-medium">{user.label}</div>
                    <div className="text-xs text-gray-500">{user.role}</div>
                  </td>
                  <td className="px-4 py-3 text-gray-400">{formatTime(user.createdAt)}</td>
                  <td className="px-4 py-3">
                    <span className={user.quota > 0 && user.usedCount >= user.quota ? 'text-red-400' : ''}>
                      {getQuotaDisplay(user)}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                      user.status === 'active'
                        ? 'bg-green-500/10 text-green-400'
                        : 'bg-red-500/10 text-red-400'
                    }`}>
                      {user.status === 'active' ? '正常' : '已禁用'}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <input
                        type="number"
                        value={quotaInputs[user.id] ?? ''}
                        onChange={e => setQuotaInputs(prev => ({ ...prev, [user.id]: e.target.value }))}
                        placeholder="数量"
                        className="w-16 rounded-lg border border-white/10 bg-white/[0.04] px-2 py-1 text-xs text-gray-100 outline-none focus:border-blue-400"
                      />
                      <button
                        onClick={() => {
                          const val = parseInt(quotaInputs[user.id] || '0', 10)
                          if (val > 0) handleQuotaDelta(user.id, val)
                        }}
                        className="rounded-lg bg-blue-600/20 px-2 py-1 text-xs text-blue-400 hover:bg-blue-600/30"
                      >
                        增加
                      </button>
                      <button
                        onClick={() => {
                          const val = parseInt(quotaInputs[user.id] || '0', 10)
                          if (val > 0) handleQuotaDelta(user.id, -val)
                        }}
                        className="rounded-lg bg-orange-600/20 px-2 py-1 text-xs text-orange-400 hover:bg-orange-600/30"
                      >
                        减少
                      </button>
                      <button
                        onClick={() => handleReset(user.id)}
                        className="rounded-lg bg-gray-600/20 px-2 py-1 text-xs text-gray-400 hover:bg-gray-600/30"
                      >
                        重置
                      </button>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => handleToggleStatus(user.id, user.status)}
                      className={`rounded-lg px-3 py-1 text-xs font-medium ${
                        user.status === 'active'
                          ? 'bg-red-600/20 text-red-400 hover:bg-red-600/30'
                          : 'bg-green-600/20 text-green-400 hover:bg-green-600/30'
                      }`}
                    >
                      {user.status === 'active' ? '禁用' : '启用'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {users.length === 0 && (
          <div className="py-12 text-center text-gray-500">暂无用户</div>
        )}
      </main>
    </div>
  )
}
