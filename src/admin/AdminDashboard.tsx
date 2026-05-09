import { useState, useEffect, useCallback } from 'react'
import {
  adminListUsers, adminUpdateQuota, adminToggleStatus, adminDeleteUser, clearAdminToken,
  adminCreateCodes, adminListCodes, adminDeleteUsers, adminDeleteCodes,
  type AdminUser, type RedemptionCode,
} from './adminApi'
import { copyTextToClipboard } from '../lib/clipboard'

interface Props {
  onLogout: () => void
}

type Tab = 'users' | 'codes'

export default function AdminDashboard({ onLogout }: Props) {
  const [tab, setTab] = useState<Tab>('users')
  const [users, setUsers] = useState<AdminUser[]>([])
  const [codes, setCodes] = useState<RedemptionCode[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Quota modal
  const [quotaModal, setQuotaModal] = useState<{ user: AdminUser; mode: 'increase' | 'decrease' | 'set' } | null>(null)
  const [quotaValue, setQuotaValue] = useState('')

  // Confirm modal
  const [confirmModal, setConfirmModal] = useState<{
    user: AdminUser
    action: 'disable' | 'enable' | 'delete'
  } | null>(null)

  // Selection state
  const [selectedUserIds, setSelectedUserIds] = useState<Set<string>>(new Set())
  const [selectedCodeIds, setSelectedCodeIds] = useState<Set<string>>(new Set())

  // Batch confirm modal
  const [batchConfirm, setBatchConfirm] = useState<{
    type: 'users' | 'codes'
    count: number
  } | null>(null)

  // Code creation form
  const [codeQuota, setCodeQuota] = useState('')
  const [codeCount, setCodeCount] = useState('1')
  const [creating, setCreating] = useState(false)
  const [codeFilter, setCodeFilter] = useState<number | null>(null)

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

  const loadCodes = useCallback(async () => {
    try {
      const { codes } = await adminListCodes()
      setCodes(codes || [])
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (tab === 'users') loadUsers()
    else loadCodes()
  }, [tab, loadUsers, loadCodes])

  const handleQuotaSubmit = async () => {
    if (!quotaModal) return
    const val = parseInt(quotaValue, 10)
    if (!val || val < 0) {
      setError('请输入有效的数值')
      return
    }
    try {
      if (quotaModal.mode === 'set') {
        await adminUpdateQuota(quotaModal.user.id, val, false, 'set')
      } else {
        const delta = quotaModal.mode === 'increase' ? val : -val
        await adminUpdateQuota(quotaModal.user.id, delta)
      }
      setQuotaModal(null)
      setQuotaValue('')
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

  const handleToggleStatus = (user: AdminUser) => {
    setConfirmModal({ user, action: user.status === 'active' ? 'disable' : 'enable' })
  }

  const handleDeleteUser = (user: AdminUser) => {
    setConfirmModal({ user, action: 'delete' })
  }

  const handleConfirmAction = async () => {
    if (!confirmModal) return
    try {
      if (confirmModal.action === 'delete') {
        await adminDeleteUser(confirmModal.user.id)
      } else {
        const newStatus = confirmModal.action === 'disable' ? 'disabled' : 'active'
        await adminToggleStatus(confirmModal.user.id, newStatus)
      }
      setConfirmModal(null)
      await loadUsers()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  const handleCreateCodes = async () => {
    const quota = parseInt(codeQuota, 10)
    const count = parseInt(codeCount, 10) || 1
    if (!quota || quota <= 0) {
      setError('配额必须大于 0')
      return
    }
    setCreating(true)
    setError('')
    try {
      const { codes: newCodes } = await adminCreateCodes(quota, count)
      await loadCodes()
      const text = newCodes.map(c => c.code).join('\n')
      await copyTextToClipboard(text)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setCreating(false)
    }
  }

  const filteredCodes = codeFilter !== null ? codes.filter(c => c.quota === codeFilter) : codes

  const handleCopyUnused = async () => {
    const unused = filteredCodes.filter(c => !c.usedBy)
    if (unused.length === 0) {
      setError('没有未使用的兑换码')
      return
    }
    const text = unused.map(c => c.code).join('\n')
    try {
      await copyTextToClipboard(text)
    } catch {
      setError('复制到剪贴板失败')
    }
  }

  const handleLogout = () => {
    clearAdminToken()
    onLogout()
  }

  const toggleUserSelect = (id: string) => {
    setSelectedUserIds(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const toggleAllUsers = () => {
    if (selectedUserIds.size === users.length) {
      setSelectedUserIds(new Set())
    } else {
      setSelectedUserIds(new Set(users.map(u => u.id)))
    }
  }

  const toggleCodeSelect = (id: string) => {
    setSelectedCodeIds(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const toggleAllCodes = () => {
    if (selectedCodeIds.size === filteredCodes.length) {
      setSelectedCodeIds(new Set())
    } else {
      setSelectedCodeIds(new Set(filteredCodes.map(c => c.id)))
    }
  }

  const handleBatchDelete = async () => {
    if (!batchConfirm) return
    try {
      if (batchConfirm.type === 'users') {
        await adminDeleteUsers(Array.from(selectedUserIds))
        setSelectedUserIds(new Set())
        await loadUsers()
      } else {
        await adminDeleteCodes(Array.from(selectedCodeIds))
        setSelectedCodeIds(new Set())
        await loadCodes()
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setBatchConfirm(null)
    }
  }

  const formatTime = (ms: number | null) => {
    if (!ms) return '-'
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
            <button onClick={() => { loadUsers(); loadCodes() }} className="text-sm text-gray-400 hover:text-gray-200">刷新</button>
            <button onClick={handleLogout} className="text-sm text-red-400 hover:text-red-300">退出登录</button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-6">
        {/* Tab switcher */}
        <div className="mb-6 flex gap-2">
          <button
            onClick={() => { setTab('users'); setSelectedUserIds(new Set()); setSelectedCodeIds(new Set()) }}
            className={`rounded-xl px-4 py-2 text-sm font-medium transition ${
              tab === 'users'
                ? 'bg-blue-600 text-white'
                : 'bg-white/5 text-gray-400 hover:bg-white/10'
            }`}
          >
            用户管理
          </button>
          <button
            onClick={() => { setTab('codes'); setSelectedUserIds(new Set()); setSelectedCodeIds(new Set()) }}
            className={`rounded-xl px-4 py-2 text-sm font-medium transition ${
              tab === 'codes'
                ? 'bg-blue-600 text-white'
                : 'bg-white/5 text-gray-400 hover:bg-white/10'
            }`}
          >
            兑换码管理
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded-xl bg-red-500/10 px-4 py-3 text-sm text-red-300">{error}</div>
        )}

        {/* Users tab */}
        {tab === 'users' && (
          <>
            {selectedUserIds.size > 0 && (
              <div className="mb-4 flex items-center gap-3">
                <span className="text-sm text-gray-400">已选 {selectedUserIds.size} 项</span>
                <button
                  onClick={() => setBatchConfirm({ type: 'users', count: selectedUserIds.size })}
                  className="rounded-xl bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
                >
                  批量删除
                </button>
                <button
                  onClick={() => setSelectedUserIds(new Set())}
                  className="rounded-xl bg-white/5 px-4 py-2 text-sm text-gray-300 hover:bg-white/10"
                >
                  取消选择
                </button>
              </div>
            )}

            <div className="overflow-x-auto rounded-2xl border border-white/10 bg-white/[0.03]">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-white/10 text-left text-gray-400">
                    <th className="w-10 px-4 py-3">
                      <input
                        type="checkbox"
                        checked={users.length > 0 && selectedUserIds.size === users.length}
                        onChange={toggleAllUsers}
                        className="accent-blue-500"
                      />
                    </th>
                    <th className="px-4 py-3 font-medium">用户</th>
                    <th className="px-4 py-3 font-medium">注册时间</th>
                    <th className="px-4 py-3 font-medium">配额</th>
                    <th className="px-4 py-3 font-medium">状态</th>
                    <th className="px-4 py-3 font-medium">配额操作</th>
                    <th className="px-4 py-3 font-medium">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map(user => (
                    <tr key={user.id} className={`border-b border-white/5 hover:bg-white/[0.02] ${selectedUserIds.has(user.id) ? 'bg-blue-500/5' : ''}`}>
                      <td className="px-4 py-3">
                        <input
                          type="checkbox"
                          checked={selectedUserIds.has(user.id)}
                          onChange={() => toggleUserSelect(user.id)}
                          className="accent-blue-500"
                        />
                      </td>
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
                          <button onClick={() => { setQuotaModal({ user, mode: 'increase' }); setQuotaValue('') }} className="rounded-lg bg-blue-600/20 px-2 py-1 text-xs text-blue-400 hover:bg-blue-600/30">增加</button>
                          <button onClick={() => { setQuotaModal({ user, mode: 'decrease' }); setQuotaValue('') }} className="rounded-lg bg-orange-600/20 px-2 py-1 text-xs text-orange-400 hover:bg-orange-600/30">减少</button>
                          <button onClick={() => { setQuotaModal({ user, mode: 'set' }); setQuotaValue('') }} className="rounded-lg bg-purple-600/20 px-2 py-1 text-xs text-purple-400 hover:bg-purple-600/30">设定</button>
                          <button onClick={() => handleReset(user.id)} className="rounded-lg bg-gray-600/20 px-2 py-1 text-xs text-gray-400 hover:bg-gray-600/30">重置</button>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => handleToggleStatus(user)}
                            className={`rounded-lg px-3 py-1 text-xs font-medium ${
                              user.status === 'active'
                                ? 'bg-orange-600/20 text-orange-400 hover:bg-orange-600/30'
                                : 'bg-green-600/20 text-green-400 hover:bg-green-600/30'
                            }`}
                          >
                            {user.status === 'active' ? '禁用' : '启用'}
                          </button>
                          <button
                            onClick={() => handleDeleteUser(user)}
                            className="rounded-lg bg-red-600/20 px-3 py-1 text-xs font-medium text-red-400 hover:bg-red-600/30"
                          >
                            删除
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {users.length === 0 && (
              <div className="py-12 text-center text-gray-500">暂无用户</div>
            )}
          </>
        )}

        {/* Codes tab */}
        {tab === 'codes' && (
          <>
            {/* Create form */}
            <div className="mb-6 rounded-2xl border border-white/10 bg-white/[0.03] p-4">
              <h3 className="text-sm font-medium text-gray-300 mb-3">创建兑换码</h3>
              <div className="flex flex-wrap items-end gap-3">
                <div>
                  <label className="block text-xs text-gray-500 mb-1">每张码的图片数</label>
                  <input
                    type="number"
                    value={codeQuota}
                    onChange={e => setCodeQuota(e.target.value)}
                    placeholder="如 100"
                    className="w-32 rounded-lg border border-white/10 bg-white/[0.04] px-3 py-2 text-sm text-gray-100 outline-none focus:border-blue-400"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">数量</label>
                  <input
                    type="number"
                    value={codeCount}
                    onChange={e => setCodeCount(e.target.value)}
                    placeholder="1"
                    className="w-24 rounded-lg border border-white/10 bg-white/[0.04] px-3 py-2 text-sm text-gray-100 outline-none focus:border-blue-400"
                  />
                </div>
                <button
                  onClick={handleCreateCodes}
                  disabled={creating || !codeQuota.trim()}
                  className="rounded-xl bg-blue-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {creating ? '创建中...' : '创建并复制'}
                </button>
                <button
                  onClick={handleCopyUnused}
                  className="rounded-xl bg-white/5 px-4 py-2 text-sm text-gray-300 transition hover:bg-white/10"
                >
                  复制全部未使用码
                </button>
              </div>
            </div>

            {/* Filter by quota */}
            {codes.length > 0 && (
              <div className="mb-4 flex flex-wrap items-center gap-2">
                <span className="text-xs text-gray-500">按额度筛选:</span>
                <button
                  onClick={() => setCodeFilter(null)}
                  className={`rounded-lg px-3 py-1 text-xs font-medium transition ${codeFilter === null ? 'bg-blue-600 text-white' : 'bg-white/5 text-gray-400 hover:bg-white/10'}`}
                >
                  全部
                </button>
                {Array.from(new Set(codes.map(c => c.quota))).sort((a, b) => a - b).map(q => (
                  <button
                    key={q}
                    onClick={() => setCodeFilter(q)}
                    className={`rounded-lg px-3 py-1 text-xs font-medium transition ${codeFilter === q ? 'bg-blue-600 text-white' : 'bg-white/5 text-gray-400 hover:bg-white/10'}`}
                  >
                    {q} 张
                  </button>
                ))}
              </div>
            )}

            {/* Codes action bar */}
            {selectedCodeIds.size > 0 && (
              <div className="mb-4 flex items-center gap-3">
                <span className="text-sm text-gray-400">已选 {selectedCodeIds.size} 项</span>
                <button
                  onClick={() => setBatchConfirm({ type: 'codes', count: selectedCodeIds.size })}
                  className="rounded-xl bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
                >
                  批量删除
                </button>
                <button
                  onClick={() => setSelectedCodeIds(new Set())}
                  className="rounded-xl bg-white/5 px-4 py-2 text-sm text-gray-300 hover:bg-white/10"
                >
                  取消选择
                </button>
              </div>
            )}

            {/* Codes table */}
            <div className="overflow-x-auto rounded-2xl border border-white/10 bg-white/[0.03]">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-white/10 text-left text-gray-400">
                    <th className="w-10 px-4 py-3">
                      <input
                        type="checkbox"
                        checked={filteredCodes.length > 0 && selectedCodeIds.size === filteredCodes.length}
                        onChange={toggleAllCodes}
                        className="accent-blue-500"
                      />
                    </th>
                    <th className="px-4 py-3 font-medium">兑换码</th>
                    <th className="px-4 py-3 font-medium">图片数</th>
                    <th className="px-4 py-3 font-medium">状态</th>
                    <th className="px-4 py-3 font-medium">使用者</th>
                    <th className="px-4 py-3 font-medium">使用时间</th>
                    <th className="px-4 py-3 font-medium">创建时间</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredCodes.map(code => (
                    <tr key={code.id} className={`border-b border-white/5 hover:bg-white/[0.02] ${selectedCodeIds.has(code.id) ? 'bg-blue-500/5' : ''}`}>
                      <td className="px-4 py-3">
                        <input
                          type="checkbox"
                          checked={selectedCodeIds.has(code.id)}
                          onChange={() => toggleCodeSelect(code.id)}
                          className="accent-blue-500"
                        />
                      </td>
                      <td className="px-4 py-3">
                        <span className="font-mono text-xs bg-white/5 px-2 py-0.5 rounded">{code.code}</span>
                      </td>
                      <td className="px-4 py-3">{code.quota}</td>
                      <td className="px-4 py-3">
                        <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${
                          code.usedBy
                            ? 'bg-gray-500/10 text-gray-400'
                            : 'bg-green-500/10 text-green-400'
                        }`}>
                          {code.usedBy ? '已使用' : '未使用'}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-gray-400 text-xs">{code.usedBy || '-'}</td>
                      <td className="px-4 py-3 text-gray-400 text-xs">{formatTime(code.usedAt)}</td>
                      <td className="px-4 py-3 text-gray-400 text-xs">{formatTime(code.createdAt)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {filteredCodes.length === 0 && (
              <div className="py-12 text-center text-gray-500">暂无兑换码</div>
            )}
          </>
        )}
      </main>

      {/* Quota Modal */}
      {quotaModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/50" onClick={() => setQuotaModal(null)} />
          <div className="relative z-10 w-full max-w-sm rounded-2xl border border-white/10 bg-gray-900 p-5 shadow-2xl">
            <h3 className="text-sm font-semibold text-gray-100 mb-4">
              {quotaModal.mode === 'increase' && `增加配额 — ${quotaModal.user.label}`}
              {quotaModal.mode === 'decrease' && `减少配额 — ${quotaModal.user.label}`}
              {quotaModal.mode === 'set' && `设定配额 — ${quotaModal.user.label}`}
            </h3>
            <div className="mb-2 text-xs text-gray-400">
              当前配额: {getQuotaDisplay(quotaModal.user)}
            </div>
            <input
              type="number"
              value={quotaValue}
              onChange={e => setQuotaValue(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && handleQuotaSubmit()}
              placeholder={quotaModal.mode === 'set' ? '输入目标值' : '输入数量'}
              autoFocus
              className="w-full rounded-xl border border-white/10 bg-white/[0.04] px-3 py-2.5 text-sm text-gray-100 outline-none focus:border-blue-400 mb-4"
            />
            <div className="flex justify-end gap-2">
              <button onClick={() => setQuotaModal(null)} className="rounded-xl px-4 py-2 text-sm text-gray-400 hover:bg-white/5">取消</button>
              <button onClick={handleQuotaSubmit} className="rounded-xl bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">确认</button>
            </div>
          </div>
        </div>
      )}

      {/* Confirm Modal */}
      {confirmModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/50" onClick={() => setConfirmModal(null)} />
          <div className="relative z-10 w-full max-w-sm rounded-2xl border border-white/10 bg-gray-900 p-5 shadow-2xl">
            <h3 className="text-sm font-semibold text-gray-100 mb-3">
              {confirmModal.action === 'delete' && `删除用户 — ${confirmModal.user.label}`}
              {confirmModal.action === 'disable' && `禁用用户 — ${confirmModal.user.label}`}
              {confirmModal.action === 'enable' && `启用用户 — ${confirmModal.user.label}`}
            </h3>
            <p className="text-sm text-gray-400 mb-5">
              {confirmModal.action === 'delete' && '删除后该用户及其数据将无法恢复，确定要删除吗？'}
              {confirmModal.action === 'disable' && '禁用后该用户将无法登录，确定要禁用吗？'}
              {confirmModal.action === 'enable' && '确定要启用该用户吗？'}
            </p>
            <div className="flex justify-end gap-2">
              <button onClick={() => setConfirmModal(null)} className="rounded-xl px-4 py-2 text-sm text-gray-400 hover:bg-white/5">取消</button>
              <button
                onClick={handleConfirmAction}
                className={`rounded-xl px-4 py-2 text-sm font-medium text-white ${
                  confirmModal.action === 'delete'
                    ? 'bg-red-600 hover:bg-red-700'
                    : confirmModal.action === 'disable'
                    ? 'bg-orange-600 hover:bg-orange-700'
                    : 'bg-green-600 hover:bg-green-700'
                }`}
              >
                确认
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Batch Confirm Modal */}
      {batchConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/50" onClick={() => setBatchConfirm(null)} />
          <div className="relative z-10 w-full max-w-sm rounded-2xl border border-white/10 bg-gray-900 p-5 shadow-2xl">
            <h3 className="text-sm font-semibold text-gray-100 mb-3">
              批量删除{batchConfirm.type === 'users' ? '用户' : '兑换码'}
            </h3>
            <p className="text-sm text-gray-400 mb-5">
              确定要删除选中的 {batchConfirm.count} 个{batchConfirm.type === 'users' ? '用户' : '兑换码'}吗？此操作不可恢复。
            </p>
            <div className="flex justify-end gap-2">
              <button onClick={() => setBatchConfirm(null)} className="rounded-xl px-4 py-2 text-sm text-gray-400 hover:bg-white/5">取消</button>
              <button
                onClick={handleBatchDelete}
                className="rounded-xl bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700"
              >
                确认删除
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
