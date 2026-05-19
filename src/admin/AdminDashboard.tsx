import { useState, useEffect, useCallback } from 'react'
import {
  adminListUsers, adminUpdateQuota, adminToggleStatus, adminDeleteUser, clearAdminToken,
  adminCreateCodes, adminListCodes, adminDeleteUsers, adminDeleteCodes,
  adminGetEndpoints, adminUpdateEndpoints, adminGetAnnouncement, adminUpdateAnnouncement,
  type AdminUser, type RedemptionCode, type ApiEndpoint,
} from './adminApi'
import { copyTextToClipboard } from '../lib/clipboard'
import { useStore } from '../store'
import Toast from '../components/Toast'

interface Props {
  onLogout: () => void
}

type Tab = 'users' | 'codes' | 'config' | 'announcement'

export default function AdminDashboard({ onLogout }: Props) {
  const [tab, setTab] = useState<Tab>('users')
  const [users, setUsers] = useState<AdminUser[]>([])
  const [codes, setCodes] = useState<RedemptionCode[]>([])
  const [loading, setLoading] = useState(true)
  const [quotaModal, setQuotaModal] = useState<{ user: AdminUser; mode: 'increase' | 'decrease' | 'set' } | null>(null)
  const [quotaValue, setQuotaValue] = useState('')
  const [confirmModal, setConfirmModal] = useState<{ user: AdminUser; action: 'disable' | 'enable' | 'delete' } | null>(null)
  const [selectedUserIds, setSelectedUserIds] = useState<Set<string>>(new Set())
  const [selectedCodeIds, setSelectedCodeIds] = useState<Set<string>>(new Set())
  const [batchConfirm, setBatchConfirm] = useState<{ type: 'users' | 'codes'; count: number } | null>(null)
  const toast = useStore((s) => s.showToast)
  const [codeQuota, setCodeQuota] = useState('')
  const [codeCount, setCodeCount] = useState('1')
  const [creating, setCreating] = useState(false)
  const [codeFilter, setCodeFilter] = useState<number | null>(null)
  const [endpoints, setEndpoints] = useState<ApiEndpoint[]>([])
  const [endpointsLoading, setEndpointsLoading] = useState(false)
  const [endpointsSaving, setEndpointsSaving] = useState(false)
  const [visibleKeys, setVisibleKeys] = useState<Set<number>>(new Set())
  const [announcementContent, setAnnouncementContent] = useState('')
  const [announcementEnabled, setAnnouncementEnabled] = useState(false)
  const [announcementLoading, setAnnouncementLoading] = useState(false)
  const [announcementSaving, setAnnouncementSaving] = useState(false)

  const loadUsers = useCallback(async () => {
    try {
      const { users } = await adminListUsers()
      setUsers(users)
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    } finally {
      setLoading(false)
    }
  }, [toast])

  const loadCodes = useCallback(async () => {
    try {
      const { codes } = await adminListCodes()
      setCodes(codes || [])
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    } finally {
      setLoading(false)
    }
  }, [toast])

  const loadEndpoints = useCallback(async () => {
    setEndpointsLoading(true)
    try {
      const { endpoints } = await adminGetEndpoints()
      setEndpoints(endpoints || [])
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    } finally {
      setEndpointsLoading(false)
    }
  }, [toast])

  const loadAnnouncement = useCallback(async () => {
    setAnnouncementLoading(true)
    try {
      const announcement = await adminGetAnnouncement()
      setAnnouncementContent(announcement.content || '')
      setAnnouncementEnabled(announcement.enabled)
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    } finally {
      setAnnouncementLoading(false)
    }
  }, [toast])

  useEffect(() => {
    if (tab === 'users') loadUsers()
    else if (tab === 'codes') loadCodes()
    else if (tab === 'config') loadEndpoints()
    else if (tab === 'announcement') loadAnnouncement()
  }, [tab, loadUsers, loadCodes, loadEndpoints, loadAnnouncement])

  const handleQuotaSubmit = async () => {
    if (!quotaModal) return
    const val = parseInt(quotaValue, 10)
    if (!val || val < 0) { toast('请输入有效的数值', 'error'); return }
    try {
      if (quotaModal.mode === 'set') {
        await adminUpdateQuota(quotaModal.user.id, val, false, 'set')
      } else {
        const delta = quotaModal.mode === 'increase' ? val : -val
        await adminUpdateQuota(quotaModal.user.id, delta)
      }
      setQuotaModal(null); setQuotaValue(''); await loadUsers(); toast('配额已更新', 'success')
    } catch (err) { toast(err instanceof Error ? err.message : String(err), 'error') }
  }

  const handleReset = async (userId: string) => {
    try { await adminUpdateQuota(userId, 0, true); await loadUsers(); toast('已重置使用计数', 'success') }
    catch (err) { toast(err instanceof Error ? err.message : String(err), 'error') }
  }

  const handleToggleStatus = (user: AdminUser) => { setConfirmModal({ user, action: user.status === 'active' ? 'disable' : 'enable' }) }
  const handleDeleteUser = (user: AdminUser) => { setConfirmModal({ user, action: 'delete' }) }

  const handleConfirmAction = async () => {
    if (!confirmModal) return
    try {
      if (confirmModal.action === 'delete') { await adminDeleteUser(confirmModal.user.id) }
      else { await adminToggleStatus(confirmModal.user.id, confirmModal.action === 'disable' ? 'disabled' : 'active') }
      const action = confirmModal.action; setConfirmModal(null); await loadUsers()
      if (action === 'delete') toast('用户已删除', 'success')
      else if (action === 'disable') toast('用户已禁用', 'success')
      else toast('用户已启用', 'success')
    } catch (err) { toast(err instanceof Error ? err.message : String(err), 'error') }
  }

  const handleCreateCodes = async () => {
    const quota = parseInt(codeQuota, 10); const count = parseInt(codeCount, 10) || 1
    if (!quota || quota <= 0) { toast('配额必须大于 0', 'error'); return }
    setCreating(true)
    try {
      const { codes: newCodes } = await adminCreateCodes(quota, count); await loadCodes()
      const text = newCodes.map(c => c.code).join('\n'); await copyTextToClipboard(text)
      toast(`已创建 ${newCodes.length} 个兑换码并复制到剪贴板`, 'success')
    } catch (err) { toast(err instanceof Error ? err.message : String(err), 'error') }
    finally { setCreating(false) }
  }

  const filteredCodes = codeFilter !== null ? codes.filter(c => c.quota === codeFilter) : codes

  const handleCopyUnused = async () => {
    const unused = filteredCodes.filter(c => !c.usedBy)
    if (unused.length === 0) { toast('没有未使用的兑换码', 'error'); return }
    const text = unused.map(c => c.code).join('\n')
    try { await copyTextToClipboard(text); toast(`已复制 ${unused.length} 个未使用码`, 'success') }
    catch { toast('复制到剪贴板失败', 'error') }
  }

  const handleCopyCode = async (code: string) => {
    try { await copyTextToClipboard(code); toast(`已复制: ${code}`, 'success') }
    catch { toast('复制到剪贴板失败', 'error') }
  }

  const handleLogout = () => { clearAdminToken(); onLogout() }

  const toggleUserSelect = (id: string) => { setSelectedUserIds(prev => { const next = new Set(prev); if (next.has(id)) next.delete(id); else next.add(id); return next }) }
  const toggleAllUsers = () => { if (selectedUserIds.size === users.length) setSelectedUserIds(new Set()); else setSelectedUserIds(new Set(users.map(u => u.id))) }
  const toggleCodeSelect = (id: string) => { setSelectedCodeIds(prev => { const next = new Set(prev); if (next.has(id)) next.delete(id); else next.add(id); return next }) }
  const toggleAllCodes = () => { if (selectedCodeIds.size === filteredCodes.length) setSelectedCodeIds(new Set()); else setSelectedCodeIds(new Set(filteredCodes.map(c => c.id))) }

  const handleBatchDelete = async () => {
    if (!batchConfirm) return
    try {
      if (batchConfirm.type === 'users') {
        const count = selectedUserIds.size; await adminDeleteUsers(Array.from(selectedUserIds)); setSelectedUserIds(new Set()); await loadUsers(); toast(`已删除 ${count} 个用户`, 'success')
      } else {
        const count = selectedCodeIds.size; await adminDeleteCodes(Array.from(selectedCodeIds)); setSelectedCodeIds(new Set()); await loadCodes(); toast(`已删除 ${count} 个兑换码`, 'success')
      }
    } catch (err) { toast(err instanceof Error ? err.message : String(err), 'error') }
    finally { setBatchConfirm(null) }
  }

  const handleAddEndpoint = () => { setEndpoints([...endpoints, { baseUrl: '', apiKey: '', priority: 0 }]) }
  const handleRemoveEndpoint = (index: number) => { setEndpoints(endpoints.filter((_, i) => i !== index)) }
  const handleEndpointChange = (index: number, field: 'baseUrl' | 'apiKey' | 'maxConcurrency' | 'priority', value: string) => {
    const updated = [...endpoints]
    if (field === 'maxConcurrency' || field === 'priority') {
      updated[index] = { ...updated[index], [field]: value === '' ? undefined : parseInt(value, 10) || 0 }
    } else {
      updated[index] = { ...updated[index], [field]: value }
    }
    setEndpoints(updated)
  }
  const handleSaveEndpoints = async () => {
    const valid = endpoints.filter(e => e.baseUrl.trim())
    if (valid.length === 0) { toast('至少需要一个 API 端点', 'error'); return }
    setEndpointsSaving(true)
    try { await adminUpdateEndpoints(valid); toast('端点配置已保存', 'success'); const { endpoints: saved } = await adminGetEndpoints(); setEndpoints(saved || []) }
    catch (err) { toast(err instanceof Error ? err.message : String(err), 'error') }
    finally { setEndpointsSaving(false) }
  }

  const handleSaveAnnouncement = async () => {
    setAnnouncementSaving(true)
    try { await adminUpdateAnnouncement(announcementContent, announcementEnabled); toast('公告已保存', 'success') }
    catch (err) { toast(err instanceof Error ? err.message : String(err), 'error') }
    finally { setAnnouncementSaving(false) }
  }

  const formatTime = (ms: number | null) => {
    if (!ms) return '-'
    return new Date(ms).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
  }
  const getQuotaDisplay = (user: AdminUser) => { if (user.quota === 0) return `${user.usedCount} / 无限制`; return `${user.usedCount} / ${user.quota}` }

  if (loading) {
    return (<div className="min-h-screen bg-gray-950 flex items-center justify-center"><div className="text-gray-400">加载中...</div></div>)
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
        <div className="mb-6 flex gap-2">
          <button onClick={() => { setTab('users'); setSelectedUserIds(new Set()); setSelectedCodeIds(new Set()) }} className={`rounded-xl px-4 py-2 text-sm font-medium transition ${tab === 'users' ? 'bg-blue-600 text-white' : 'bg-white/5 text-gray-400 hover:bg-white/10'}`}>用户管理</button>
          <button onClick={() => { setTab('codes'); setSelectedUserIds(new Set()); setSelectedCodeIds(new Set()) }} className={`rounded-xl px-4 py-2 text-sm font-medium transition ${tab === 'codes' ? 'bg-blue-600 text-white' : 'bg-white/5 text-gray-400 hover:bg-white/10'}`}>兑换码管理</button>
          <button onClick={() => { setTab('config'); setSelectedUserIds(new Set()); setSelectedCodeIds(new Set()) }} className={`rounded-xl px-4 py-2 text-sm font-medium transition ${tab === 'config' ? 'bg-blue-600 text-white' : 'bg-white/5 text-gray-400 hover:bg-white/10'}`}>系统配置</button>
          <button onClick={() => { setTab('announcement'); setSelectedUserIds(new Set()); setSelectedCodeIds(new Set()) }} className={`rounded-xl px-4 py-2 text-sm font-medium transition ${tab === 'announcement' ? 'bg-blue-600 text-white' : 'bg-white/5 text-gray-400 hover:bg-white/10'}`}>公告管理</button>
        </div>

        {tab === 'users' && (
          <>
            {selectedUserIds.size > 0 && (
              <div className="mb-4 flex items-center gap-3">
                <span className="text-sm text-gray-400">已选 {selectedUserIds.size} 项</span>
                <button onClick={() => setBatchConfirm({ type: 'users', count: selectedUserIds.size })} className="rounded-xl bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700">批量删除</button>
                <button onClick={() => setSelectedUserIds(new Set())} className="rounded-xl bg-white/5 px-4 py-2 text-sm text-gray-300 hover:bg-white/10">取消选择</button>
              </div>
            )}
            <div className="overflow-x-auto rounded-2xl border border-white/10 bg-white/[0.03]">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-white/10 text-left text-gray-400">
                    <th className="w-10 px-4 py-3"><input type="checkbox" checked={users.length > 0 && selectedUserIds.size === users.length} onChange={toggleAllUsers} className="accent-blue-500" /></th>
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
                      <td className="px-4 py-3"><input type="checkbox" checked={selectedUserIds.has(user.id)} onChange={() => toggleUserSelect(user.id)} className="accent-blue-500" /></td>
                      <td className="px-4 py-3"><div className="font-medium">{user.label}</div><div className="text-xs text-gray-500">{user.role}</div></td>
                      <td className="px-4 py-3 text-gray-400">{formatTime(user.createdAt)}</td>
                      <td className="px-4 py-3"><span className={user.quota > 0 && user.usedCount >= user.quota ? 'text-red-400' : ''}>{getQuotaDisplay(user)}</span></td>
                      <td className="px-4 py-3"><span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${user.status === 'active' ? 'bg-green-500/10 text-green-400' : 'bg-red-500/10 text-red-400'}`}>{user.status === 'active' ? '正常' : '已禁用'}</span></td>
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
                          <button onClick={() => handleToggleStatus(user)} className={`rounded-lg px-3 py-1 text-xs font-medium ${user.status === 'active' ? 'bg-orange-600/20 text-orange-400 hover:bg-orange-600/30' : 'bg-green-600/20 text-green-400 hover:bg-green-600/30'}`}>{user.status === 'active' ? '禁用' : '启用'}</button>
                          <button onClick={() => handleDeleteUser(user)} className="rounded-lg bg-red-600/20 px-3 py-1 text-xs font-medium text-red-400 hover:bg-red-600/30">删除</button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {users.length === 0 && <div className="py-12 text-center text-gray-500">暂无用户</div>}
          </>
        )}

        {tab === 'codes' && (
          <>
            <div className="mb-6 rounded-2xl border border-white/10 bg-white/[0.03] p-4">
              <h3 className="text-sm font-medium text-gray-300 mb-3">创建兑换码</h3>
              <div className="flex flex-wrap items-end gap-3">
                <div><label className="block text-xs text-gray-500 mb-1">每张码的图片数</label><input type="number" value={codeQuota} onChange={e => setCodeQuota(e.target.value)} placeholder="如 100" className="w-32 rounded-lg border border-white/10 bg-white/[0.04] px-3 py-2 text-sm text-gray-100 outline-none focus:border-blue-400" /></div>
                <div><label className="block text-xs text-gray-500 mb-1">数量</label><input type="number" value={codeCount} onChange={e => setCodeCount(e.target.value)} placeholder="1" className="w-24 rounded-lg border border-white/10 bg-white/[0.04] px-3 py-2 text-sm text-gray-100 outline-none focus:border-blue-400" /></div>
                <button onClick={handleCreateCodes} disabled={creating || !codeQuota.trim()} className="rounded-xl bg-blue-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50">{creating ? '创建中...' : '创建并复制'}</button>
                <button onClick={handleCopyUnused} className="rounded-xl bg-white/5 px-4 py-2 text-sm text-gray-300 transition hover:bg-white/10">复制全部未使用码</button>
              </div>
            </div>

            {codes.length > 0 && (
              <div className="mb-4 flex flex-wrap items-center gap-2">
                <span className="text-xs text-gray-500">按额度筛选:</span>
                <button onClick={() => setCodeFilter(null)} className={`rounded-lg px-3 py-1 text-xs font-medium transition ${codeFilter === null ? 'bg-blue-600 text-white' : 'bg-white/5 text-gray-400 hover:bg-white/10'}`}>全部</button>
                {Array.from(new Set(codes.map(c => c.quota))).sort((a, b) => a - b).map(q => (
                  <button key={q} onClick={() => setCodeFilter(q)} className={`rounded-lg px-3 py-1 text-xs font-medium transition ${codeFilter === q ? 'bg-blue-600 text-white' : 'bg-white/5 text-gray-400 hover:bg-white/10'}`}>{q} 张</button>
                ))}
              </div>
            )}

            {selectedCodeIds.size > 0 && (
              <div className="mb-4 flex items-center gap-3">
                <span className="text-sm text-gray-400">已选 {selectedCodeIds.size} 项</span>
                <button onClick={() => setBatchConfirm({ type: 'codes', count: selectedCodeIds.size })} className="rounded-xl bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700">批量删除</button>
                <button onClick={() => setSelectedCodeIds(new Set())} className="rounded-xl bg-white/5 px-4 py-2 text-sm text-gray-300 hover:bg-white/10">取消选择</button>
              </div>
            )}

            <div className="overflow-x-auto rounded-2xl border border-white/10 bg-white/[0.03]">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-white/10 text-left text-gray-400">
                    <th className="w-10 px-4 py-3"><input type="checkbox" checked={filteredCodes.length > 0 && selectedCodeIds.size === filteredCodes.length} onChange={toggleAllCodes} className="accent-blue-500" /></th>
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
                      <td className="px-4 py-3"><input type="checkbox" checked={selectedCodeIds.has(code.id)} onChange={() => toggleCodeSelect(code.id)} className="accent-blue-500" /></td>
                      <td className="px-4 py-3"><button onClick={() => handleCopyCode(code.code)} className="font-mono text-xs bg-white/5 px-2 py-0.5 rounded hover:bg-white/10 transition cursor-pointer" title="点击复制">{code.code}</button></td>
                      <td className="px-4 py-3">{code.quota}</td>
                      <td className="px-4 py-3"><span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${code.usedBy ? 'bg-gray-500/10 text-gray-400' : 'bg-green-500/10 text-green-400'}`}>{code.usedBy ? '已使用' : '未使用'}</span></td>
                      <td className="px-4 py-3 text-gray-400 text-xs">{code.usedBy || '-'}</td>
                      <td className="px-4 py-3 text-gray-400 text-xs">{formatTime(code.usedAt)}</td>
                      <td className="px-4 py-3 text-gray-400 text-xs">{formatTime(code.createdAt)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {filteredCodes.length === 0 && <div className="py-12 text-center text-gray-500">暂无兑换码</div>}
          </>
        )}

        {tab === 'config' && (
          <div className="rounded-2xl border border-white/10 bg-white/[0.03] p-5">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-sm font-medium text-gray-200">API 端点池</h3>
                <p className="text-xs text-gray-500 mt-1">配置 OpenAI API 端点，支持按优先级调度和多端点故障转移。优先级数值越大越优先，同优先级按列表顺序尝试。</p>
              </div>
              <button onClick={handleAddEndpoint} className="rounded-xl bg-blue-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-blue-700">添加端点</button>
            </div>
            {endpointsLoading ? (
              <div className="py-8 text-center text-gray-500">加载中...</div>
            ) : endpoints.length === 0 ? (
              <div className="py-8 text-center text-gray-500">暂未配置任何端点</div>
            ) : (
              <div className="space-y-3">
                {endpoints.map((ep, i) => (
                  <div key={i} className="flex items-start gap-3 rounded-xl border border-white/10 bg-white/[0.02] p-4">
                    <div className="flex-1 space-y-2">
                      <div><label className="block text-xs text-gray-500 mb-1">Base URL</label><input value={ep.baseUrl} onChange={e => handleEndpointChange(i, 'baseUrl', e.target.value)} placeholder="https://api.openai.com/v1" className="w-full rounded-lg border border-white/10 bg-white/[0.04] px-3 py-2 text-sm text-gray-100 outline-none focus:border-blue-400 font-mono" /></div>
                      <div>
                        <label className="block text-xs text-gray-500 mb-1">API Key</label>
                        <div className="relative">
                          <input type={visibleKeys.has(i) ? 'text' : 'password'} value={ep.apiKey} onChange={e => handleEndpointChange(i, 'apiKey', e.target.value)} placeholder="sk-..." className="w-full rounded-lg border border-white/10 bg-white/[0.04] px-3 py-2 pr-10 text-sm text-gray-100 outline-none focus:border-blue-400 font-mono" />
                          <button type="button" onClick={() => setVisibleKeys(prev => { const next = new Set(prev); if (next.has(i)) next.delete(i); else next.add(i); return next })} className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-gray-500 hover:text-gray-300 transition">
                            {visibleKeys.has(i) ? (
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                            ) : (
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" /></svg>
                            )}
                          </button>
                        </div>
                      </div>
                      <div className="flex flex-wrap gap-3">
                        <div><label className="block text-xs text-gray-500 mb-1">最大并发数（0 = 无限制）</label><input type="number" min="0" value={ep.maxConcurrency ?? ''} onChange={e => handleEndpointChange(i, 'maxConcurrency', e.target.value)} placeholder="0" className="w-32 rounded-lg border border-white/10 bg-white/[0.04] px-3 py-2 text-sm text-gray-100 outline-none focus:border-blue-400 font-mono" /></div>
                        <div><label className="block text-xs text-gray-500 mb-1">优先级（越大越优先）</label><input type="number" min="0" value={ep.priority ?? ''} onChange={e => handleEndpointChange(i, 'priority', e.target.value)} placeholder="0" className="w-32 rounded-lg border border-white/10 bg-white/[0.04] px-3 py-2 text-sm text-gray-100 outline-none focus:border-blue-400 font-mono" /></div>
                      </div>
                    </div>
                    <button onClick={() => handleRemoveEndpoint(i)} className="mt-6 rounded-lg bg-red-600/20 px-3 py-2 text-xs font-medium text-red-400 hover:bg-red-600/30 transition">删除</button>
                  </div>
                ))}
              </div>
            )}
            {endpoints.length > 0 && (
              <div className="mt-4 flex justify-end">
                <button onClick={handleSaveEndpoints} disabled={endpointsSaving} className="rounded-xl bg-green-600 px-5 py-2 text-sm font-medium text-white transition hover:bg-green-700 disabled:cursor-not-allowed disabled:opacity-50">{endpointsSaving ? '保存中...' : '保存配置'}</button>
              </div>
            )}
          </div>
        )}

        {tab === 'announcement' && (
          <div className="rounded-2xl border border-white/10 bg-white/[0.03] p-5">
            <div className="mb-5">
              <h3 className="text-sm font-medium text-gray-200">站点公告</h3>
              <p className="mt-1 text-xs text-gray-500">启用后，用户首次打开或公告更新后会看到一次弹窗，之后可在右上角问号菜单中查看。</p>
            </div>
            {announcementLoading ? (
              <div className="py-8 text-center text-gray-500">加载中...</div>
            ) : (
              <>
                <label className="mb-4 flex cursor-pointer items-center gap-3">
                  <span className="relative inline-flex h-6 w-11 items-center rounded-full bg-white/10 transition">
                    <input type="checkbox" checked={announcementEnabled} onChange={e => setAnnouncementEnabled(e.target.checked)} className="peer sr-only" />
                    <span className="absolute inset-0 rounded-full transition peer-checked:bg-blue-600" />
                    <span className="absolute left-0.5 top-0.5 h-5 w-5 rounded-full bg-white transition-transform peer-checked:translate-x-5" />
                  </span>
                  <span className="text-sm text-gray-300">启用公告</span>
                </label>
                <div className="mb-4">
                  <label className="mb-1 block text-xs text-gray-500">公告内容</label>
                  <textarea
                    value={announcementContent}
                    onChange={e => setAnnouncementContent(e.target.value)}
                    rows={8}
                    placeholder="输入公告内容..."
                    className="w-full resize-y rounded-xl border border-white/10 bg-white/[0.04] px-4 py-3 text-sm text-gray-100 outline-none transition focus:border-blue-400"
                  />
                </div>
                <div className="flex justify-end">
                  <button onClick={handleSaveAnnouncement} disabled={announcementSaving} className="rounded-xl bg-green-600 px-5 py-2 text-sm font-medium text-white transition hover:bg-green-700 disabled:cursor-not-allowed disabled:opacity-50">
                    {announcementSaving ? '保存中...' : '保存公告'}
                  </button>
                </div>
              </>
            )}
          </div>
        )}
      </main>

      {quotaModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/50" onClick={() => setQuotaModal(null)} />
          <div className="relative z-10 w-full max-w-sm rounded-2xl border border-white/10 bg-gray-900 p-5 shadow-2xl">
            <h3 className="text-sm font-semibold text-gray-100 mb-4">
              {quotaModal.mode === 'increase' && `增加配额 — ${quotaModal.user.label}`}
              {quotaModal.mode === 'decrease' && `减少配额 — ${quotaModal.user.label}`}
              {quotaModal.mode === 'set' && `设定配额 — ${quotaModal.user.label}`}
            </h3>
            <div className="mb-2 text-xs text-gray-400">当前配额: {getQuotaDisplay(quotaModal.user)}</div>
            <input type="number" value={quotaValue} onChange={e => setQuotaValue(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleQuotaSubmit()} placeholder={quotaModal.mode === 'set' ? '输入目标值' : '输入数量'} autoFocus className="w-full rounded-xl border border-white/10 bg-white/[0.04] px-3 py-2.5 text-sm text-gray-100 outline-none focus:border-blue-400 mb-4" />
            <div className="flex justify-end gap-2">
              <button onClick={() => setQuotaModal(null)} className="rounded-xl px-4 py-2 text-sm text-gray-400 hover:bg-white/5">取消</button>
              <button onClick={handleQuotaSubmit} className="rounded-xl bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">确认</button>
            </div>
          </div>
        </div>
      )}

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
              <button onClick={handleConfirmAction} className={`rounded-xl px-4 py-2 text-sm font-medium text-white ${confirmModal.action === 'delete' ? 'bg-red-600 hover:bg-red-700' : confirmModal.action === 'disable' ? 'bg-orange-600 hover:bg-orange-700' : 'bg-green-600 hover:bg-green-700'}`}>确认</button>
            </div>
          </div>
        </div>
      )}

      {batchConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/50" onClick={() => setBatchConfirm(null)} />
          <div className="relative z-10 w-full max-w-sm rounded-2xl border border-white/10 bg-gray-900 p-5 shadow-2xl">
            <h3 className="text-sm font-semibold text-gray-100 mb-3">批量删除{batchConfirm.type === 'users' ? '用户' : '兑换码'}</h3>
            <p className="text-sm text-gray-400 mb-5">确定要删除选中的 {batchConfirm.count} 个{batchConfirm.type === 'users' ? '用户' : '兑换码'}吗？此操作不可恢复。</p>
            <div className="flex justify-end gap-2">
              <button onClick={() => setBatchConfirm(null)} className="rounded-xl px-4 py-2 text-sm text-gray-400 hover:bg-white/5">取消</button>
              <button onClick={handleBatchDelete} className="rounded-xl bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700">确认删除</button>
            </div>
          </div>
        </div>
      )}
      <Toast />
    </div>
  )
}
