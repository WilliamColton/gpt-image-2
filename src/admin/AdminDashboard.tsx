import { useState, useEffect, useCallback } from 'react'
import { Eye, EyeOff } from 'lucide-react'
import {
  adminListUsers, adminUpdateQuota, adminToggleStatus, adminDeleteUser, clearAdminToken,
  adminCreateCodes, adminListCodes, adminDeleteUsers, adminDeleteCodes,
  adminGetEndpoints, adminUpdateEndpoints, adminGetAnnouncement, adminUpdateAnnouncement,
  adminListFeedbacks, adminUpdateFeedbackStatus,
  adminListChangelogEntries, adminCreateChangelogEntry, adminUpdateChangelogEntry, adminDeleteChangelogEntry,
  adminGetPricingConfig, adminUpdatePricingConfig,
  adminGetBillingSummary, adminGetBillingTrend,
  adminGetBillingEndpointBreakdown, adminGetBillingUserBreakdown,
  adminResetPassword, adminGetInviteConfig, adminUpdateInviteConfig, adminListInvites,
  type AdminUser, type RedemptionCode, type ApiEndpoint,
  type AnalyticsRange, type AnalyticsMeta, type BillingSummary, type BillingTrendPoint,
  type BillingEndpointRow, type BillingUserRow,
} from './adminApi'
import { formatMoneyInputFromX10000, parseMoneyInputToX10000 } from './moneyFormat'
import { copyTextToClipboard } from '../lib/clipboard'
import { useStore } from '../store'
import { Toaster } from '../components/ui/sonner'
import type { BugFeedback, BugFeedbackStatus, ChangelogEntry, ChangelogEntryPayload, ThemeMode } from '../types'
import Select from '../components/Select'
import { Input } from '../components/ui/input'
import { Textarea } from '../components/ui/textarea'
import { Button } from '../components/ui/button'
import { Tabs, TabsList, TabsTrigger } from '../components/ui/tabs'
import { Switch } from '../components/ui/switch'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '../components/ui/alert-dialog'

interface Props {
  onLogout: () => void
}

type Tab = 'users' | 'codes' | 'config' | 'analytics' | 'announcement' | 'feedback' | 'changelog' | 'invites'

export default function AdminDashboard({ onLogout }: Props) {
  const [tab, setTab] = useState<Tab>('users')
  const [users, setUsers] = useState<AdminUser[]>([])
  const [codes, setCodes] = useState<RedemptionCode[]>([])
  const [loading, setLoading] = useState(true)
  const [quotaModal, setQuotaModal] = useState<{ user: AdminUser; mode: 'increase' | 'decrease' | 'set' } | null>(null)
  const [quotaValue, setQuotaValue] = useState('')
  const [quotaConfirm, setQuotaConfirm] = useState<{ user: AdminUser; mode: 'increase' | 'decrease' | 'set'; value: number } | null>(null)
  const [confirmModal, setConfirmModal] = useState<{ user: AdminUser; action: 'disable' | 'enable' | 'delete' } | null>(null)
  const [selectedUserIds, setSelectedUserIds] = useState<Set<string>>(new Set())
  const [selectedCodeIds, setSelectedCodeIds] = useState<Set<string>>(new Set())
  const [batchConfirm, setBatchConfirm] = useState<{ type: 'users' | 'codes'; count: number } | null>(null)
  const toast = useStore((s) => s.showToast)
  const settings = useStore((s) => s.settings)
  const setSettings = useStore((s) => s.setSettings)
  const [codeQuota, setCodeQuota] = useState('')
  const [codeCount, setCodeCount] = useState('1')
  const [creating, setCreating] = useState(false)
  const [codeFilter, setCodeFilter] = useState<number | null>(null)
  const [endpoints, setEndpoints] = useState<ApiEndpoint[]>([])
  const [endpointsLoading, setEndpointsLoading] = useState(false)
  const [pricingSaving, setPricingSaving] = useState(false)
  const [salePriceInput, setSalePriceInput] = useState('')
  const [costInputDrafts, setCostInputDrafts] = useState<Record<number, string>>({})
  const [priceErrors, setPriceErrors] = useState<Record<number, string | null>>({})
  const [visibleKeys, setVisibleKeys] = useState<Set<number>>(new Set())
  const [announcementContent, setAnnouncementContent] = useState('')
  const [announcementEnabled, setAnnouncementEnabled] = useState(false)
  const [announcementLoading, setAnnouncementLoading] = useState(false)
  const [announcementSaving, setAnnouncementSaving] = useState(false)
  const [feedbacks, setFeedbacks] = useState<BugFeedback[]>([])
  const [feedbacksLoading, setFeedbacksLoading] = useState(false)
  const [feedbackUpdatingId, setFeedbackUpdatingId] = useState<string | null>(null)
  const [changelogs, setChangelogs] = useState<ChangelogEntry[]>([])
  const [changelogLoading, setChangelogLoading] = useState(false)
  const [changelogSaving, setChangelogSaving] = useState(false)
  const [editingChangelogId, setEditingChangelogId] = useState<string | null>(null)
  const [changelogVersion, setChangelogVersion] = useState('')
  const [changelogTitle, setChangelogTitle] = useState('')
  const [changelogContent, setChangelogContent] = useState('')
  const [changelogPublished, setChangelogPublished] = useState(false)
  const [deleteChangelogId, setDeleteChangelogId] = useState<string | null>(null)

  // ─── Invite code state ───
  const [inviterReward, setInviterReward] = useState(0)
  const [inviteeReward, setInviteeReward] = useState(0)
  const [defaultQuota, setDefaultQuota] = useState(0)
  const [inviteConfigLoading, setInviteConfigLoading] = useState(false)
  const [inviteConfigSaving, setInviteConfigSaving] = useState(false)
  const [inviteRows, setInviteRows] = useState<Array<{username:string;inviteCode:string;usageCount:number}>>([])
  const [inviteRowsLoading, setInviteRowsLoading] = useState(false)
  const [resetPasswordModal, setResetPasswordModal] = useState<{userId:string;label:string} | null>(null)
  const [resetQuotaConfirm, setResetQuotaConfirm] = useState<string | null>(null)
  const [resetPasswordValue, setResetPasswordValue] = useState('')
  const [resetPasswordConfirm, setResetPasswordConfirm] = useState<{userId:string;label:string;value:string} | null>(null)

  // ─── Analytics state ───
  const [analyticsRange, setAnalyticsRange] = useState<AnalyticsRange>('7d')
  const [summary, setSummary] = useState<BillingSummary | null>(null)
  const [summaryMeta, setSummaryMeta] = useState<AnalyticsMeta | null>(null)
  const [summaryLoading, setSummaryLoading] = useState(false)
  const [summaryError, setSummaryError] = useState('')
  const [trend, setTrend] = useState<BillingTrendPoint[]>([])
  const [trendMeta, setTrendMeta] = useState<AnalyticsMeta | null>(null)
  const [trendLoading, setTrendLoading] = useState(false)
  const [trendError, setTrendError] = useState('')
  const [endpointRows, setEndpointRows] = useState<BillingEndpointRow[]>([])
  const [endpointMeta, setEndpointMeta] = useState<AnalyticsMeta | null>(null)
  const [endpointLoading, setEndpointLoading] = useState(false)
  const [endpointError, setEndpointError] = useState('')
  const [userRows, setUserRows] = useState<BillingUserRow[]>([])
  const [userMeta, setUserMeta] = useState<AnalyticsMeta | null>(null)
  const [userLoading, setUserLoading] = useState(false)
  const [userError, setUserError] = useState('')

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

  const loadPricingConfig = useCallback(async () => {
    setEndpointsLoading(true)
    try {
      const pricing = await adminGetPricingConfig()
      const eps = pricing.endpoints || []
      setEndpoints(eps)
      setSalePriceInput(formatMoneyInputFromX10000(pricing.salePriceX10000))
      const drafts: Record<number, string> = {}
      eps.forEach((ep, i) => {
        drafts[i] = formatMoneyInputFromX10000(ep.costPerImageX10000 ?? 0)
      })
      setCostInputDrafts(drafts)
      setPriceErrors({})
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

  const loadFeedbacks = useCallback(async () => {
    setFeedbacksLoading(true)
    try {
      const { feedbacks } = await adminListFeedbacks()
      setFeedbacks(feedbacks || [])
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    } finally {
      setFeedbacksLoading(false)
      setLoading(false)
    }
  }, [toast])

  const loadChangelogs = useCallback(async () => {
    setChangelogLoading(true)
    try {
      const { changelogs } = await adminListChangelogEntries()
      setChangelogs(changelogs || [])
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    } finally {
      setChangelogLoading(false)
      setLoading(false)
    }
  }, [toast])

  const loadInviteConfig = useCallback(async () => {
    setInviteConfigLoading(true)
    try {
      const cfg = await adminGetInviteConfig()
      setInviterReward(cfg.inviterReward)
      setInviteeReward(cfg.inviteeReward)
      setDefaultQuota(cfg.defaultQuota)
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    } finally {
      setInviteConfigLoading(false)
    }
  }, [toast])

  const loadInviteRows = useCallback(async () => {
    setInviteRowsLoading(true)
    try {
      const { invites } = await adminListInvites()
      setInviteRows(invites || [])
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    } finally {
      setInviteRowsLoading(false)
    }
  }, [toast])

  // ─── Analytics loaders ───
  const loadAnalyticsSummary = useCallback(async (range: AnalyticsRange = analyticsRange) => {
    setSummaryLoading(true)
    setSummaryError('')
    try {
      const res = await adminGetBillingSummary(range)
      setSummary(res.summary)
      setSummaryMeta(res.meta)
    } catch (err) {
      setSummaryError(err instanceof Error ? err.message : String(err))
    } finally {
      setSummaryLoading(false)
    }
  }, [analyticsRange])

  const loadAnalyticsTrend = useCallback(async (range: AnalyticsRange = analyticsRange) => {
    setTrendLoading(true)
    setTrendError('')
    try {
      const res = await adminGetBillingTrend(range)
      setTrend(res.trend)
      setTrendMeta(res.meta)
    } catch (err) {
      setTrendError(err instanceof Error ? err.message : String(err))
    } finally {
      setTrendLoading(false)
    }
  }, [analyticsRange])

  const loadAnalyticsEndpointBreakdown = useCallback(async (range: AnalyticsRange = analyticsRange) => {
    setEndpointLoading(true)
    setEndpointError('')
    try {
      const res = await adminGetBillingEndpointBreakdown(range)
      setEndpointRows(res.rows)
      setEndpointMeta(res.meta)
    } catch (err) {
      setEndpointError(err instanceof Error ? err.message : String(err))
    } finally {
      setEndpointLoading(false)
    }
  }, [analyticsRange])

  const loadAnalyticsUserBreakdown = useCallback(async (range: AnalyticsRange = analyticsRange) => {
    setUserLoading(true)
    setUserError('')
    try {
      const res = await adminGetBillingUserBreakdown(range)
      setUserRows(res.rows)
      setUserMeta(res.meta)
    } catch (err) {
      setUserError(err instanceof Error ? err.message : String(err))
    } finally {
      setUserLoading(false)
    }
  }, [analyticsRange])

  useEffect(() => {
    if (tab === 'users') loadUsers()
    else if (tab === 'codes') loadCodes()
    else if (tab === 'config') loadPricingConfig()
    else if (tab === 'analytics') { loadAnalyticsSummary(analyticsRange); loadAnalyticsTrend(analyticsRange); loadAnalyticsEndpointBreakdown(analyticsRange); loadAnalyticsUserBreakdown(analyticsRange) }
    else if (tab === 'announcement') loadAnnouncement()
    else if (tab === 'feedback') loadFeedbacks()
    else if (tab === 'changelog') loadChangelogs()
    else if (tab === 'invites') { loadInviteConfig(); loadInviteRows() }
  }, [tab, loadUsers, loadCodes, loadPricingConfig, loadAnnouncement, loadFeedbacks, loadChangelogs, loadInviteConfig, loadInviteRows, loadAnalyticsSummary, loadAnalyticsTrend, loadAnalyticsEndpointBreakdown, loadAnalyticsUserBreakdown, analyticsRange])

  const handleQuotaConfirm = () => {
    if (!quotaModal) return
    const val = parseInt(quotaValue, 10)
    if (!val || val < 0) { toast('请输入有效的数值', 'error'); return }
    setQuotaConfirm({ ...quotaModal, value: val })
    setQuotaModal(null)
  }

  const handleQuotaSubmit = async () => {
    if (!quotaConfirm) return
    const { user, mode, value } = quotaConfirm
    setQuotaConfirm(null)
    try {
      if (mode === 'set') {
        await adminUpdateQuota(user.id, value, false, 'set')
      } else {
        const delta = mode === 'increase' ? value : -value
        await adminUpdateQuota(user.id, delta)
      }
      setQuotaValue(''); await loadUsers(); toast('配额已更新', 'success')
    } catch (err) { toast(err instanceof Error ? err.message : String(err), 'error') }
  }

  const handleReset = async (userId: string) => {
    setResetQuotaConfirm(userId)
  }

  const handleResetConfirm = async () => {
    if (!resetQuotaConfirm) return
    try { await adminUpdateQuota(resetQuotaConfirm, 0, true); await loadUsers(); toast('已重置使用计数', 'success') }
    catch (err) { toast(err instanceof Error ? err.message : String(err), 'error') }
    finally { setResetQuotaConfirm(null) }
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

  const handleAddEndpoint = () => {
    const newIndex = endpoints.length
    setEndpoints([...endpoints, { baseUrl: '', apiKey: '', priority: 0 }])
    setCostInputDrafts(prev => ({ ...prev, [newIndex]: '0' }))
  }
  const handleRemoveEndpoint = (index: number) => {
    setEndpoints(endpoints.filter((_, i) => i !== index))
    setCostInputDrafts(prev => {
      const next = { ...prev }
      delete next[index]
      return next
    })
    setPriceErrors(prev => {
      const next = { ...prev }
      delete next[index]
      return next
    })
  }
  const handleEndpointCostChange = (index: number, rawValue: string) => {
    setCostInputDrafts(prev => ({ ...prev, [index]: rawValue }))
    // Validate inline
    if (rawValue.trim() === '') {
      setPriceErrors(prev => ({ ...prev, [index]: '请输入非负数字，最多 4 位小数' }))
    } else if (parseMoneyInputToX10000(rawValue) === null) {
      setPriceErrors(prev => ({ ...prev, [index]: '请输入非负数字，最多 4 位小数' }))
    } else {
      setPriceErrors(prev => {
        const next = { ...prev }
        delete next[index]
        return next
      })
    }
  }
  const handleEndpointChange = (index: number, field: 'baseUrl' | 'apiKey' | 'maxConcurrency' | 'priority', value: string) => {
    const updated = [...endpoints]
    if (field === 'maxConcurrency' || field === 'priority') {
      updated[index] = { ...updated[index], [field]: value === '' ? undefined : parseInt(value, 10) || 0 }
    } else {
      updated[index] = { ...updated[index], [field]: value }
    }
    setEndpoints(updated)
  }
  const handleSavePricingConfig = async () => {
    // Validate at least one endpoint
    const valid = endpoints.filter(e => e.baseUrl.trim())
    if (valid.length === 0) { toast('至少需要一个 API 端点', 'error'); return }

    // Validate all cost inputs
    const parseErrors: Record<number, string> = {}
    for (let i = 0; i < endpoints.length; i++) {
      const draft = costInputDrafts[i] ?? ''
      if (parseMoneyInputToX10000(draft) === null) {
        parseErrors[i] = '请输入非负数字，最多 4 位小数'
      }
    }

    // Validate sale price
    const salePriceParsed = parseMoneyInputToX10000(salePriceInput)

    if (Object.keys(parseErrors).length > 0 || salePriceParsed === null) {
      setPriceErrors(prev => ({ ...prev, ...parseErrors }))
      if (salePriceParsed === null) {
        toast('请输入有效的全局售价，最多 4 位小数', 'error')
      }
      return
    }

    // Build endpoints with cost values
    const pricedEndpoints = valid.map((ep, i) => {
      const costVal = parseMoneyInputToX10000(costInputDrafts[i] ?? '0') ?? 0
      return {
        ...ep,
        costPerImageX10000: costVal,
        maxConcurrency: ep.maxConcurrency,
        priority: ep.priority,
      }
    })

    setPricingSaving(true)
    try {
      await adminUpdatePricingConfig(pricedEndpoints, salePriceParsed!)
      toast('配置已保存', 'success')
      await loadPricingConfig()
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    } finally {
      setPricingSaving(false)
    }
  }

  const handleSaveAnnouncement = async () => {
    setAnnouncementSaving(true)
    try { await adminUpdateAnnouncement(announcementContent, announcementEnabled); toast('公告已保存', 'success') }
    catch (err) { toast(err instanceof Error ? err.message : String(err), 'error') }
    finally { setAnnouncementSaving(false) }
  }

  const handleUpdateFeedbackStatus = async (id: string, status: BugFeedbackStatus) => {
    setFeedbackUpdatingId(id)
    try {
      const updated = await adminUpdateFeedbackStatus(id, status)
      setFeedbacks(feedbacks.map(feedback => feedback.id === id ? updated : feedback))
      toast('反馈状态已更新', 'success')
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    } finally {
      setFeedbackUpdatingId(null)
    }
  }

  const resetChangelogForm = () => {
    setEditingChangelogId(null)
    setChangelogVersion('')
    setChangelogTitle('')
    setChangelogContent('')
    setChangelogPublished(false)
  }

  const openChangelogEditor = (entry: ChangelogEntry) => {
    setEditingChangelogId(entry.id)
    setChangelogVersion(entry.version)
    setChangelogTitle(entry.title)
    setChangelogContent(entry.content)
    setChangelogPublished(entry.published)
  }

  const handleSaveChangelog = async () => {
    if (changelogPublished && !changelogVersion.trim()) {
      toast('发布更新日志前请填写版本号', 'error')
      return
    }
    const payload: ChangelogEntryPayload = {
      version: changelogVersion,
      title: changelogTitle,
      content: changelogContent,
      published: changelogPublished,
    }
    setChangelogSaving(true)
    try {
      if (editingChangelogId) {
        await adminUpdateChangelogEntry(editingChangelogId, payload)
        toast('更新日志已保存', 'success')
      } else {
        await adminCreateChangelogEntry(payload)
        toast('更新日志已创建', 'success')
      }
      resetChangelogForm()
      await loadChangelogs()
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    } finally {
      setChangelogSaving(false)
    }
  }

  const handleDeleteChangelog = async () => {
    if (!deleteChangelogId) return
    try {
      await adminDeleteChangelogEntry(deleteChangelogId)
      if (editingChangelogId === deleteChangelogId) resetChangelogForm()
      setDeleteChangelogId(null)
      await loadChangelogs()
      toast('更新日志已删除', 'success')
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    }
  }

  const handleSaveInviteConfig = async () => {
    setInviteConfigSaving(true)
    try {
      await adminUpdateInviteConfig(inviterReward, inviteeReward, defaultQuota)
      toast('配置已保存', 'success')
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    } finally {
      setInviteConfigSaving(false)
    }
  }

  const handleResetPasswordConfirm = () => {
    if (!resetPasswordModal) return
    if (resetPasswordValue.length < 8) {
      toast('密码至少需要 8 个字符', 'error')
      return
    }
    setResetPasswordConfirm({ ...resetPasswordModal, value: resetPasswordValue.trim() })
    setResetPasswordModal(null)
  }

  const handleResetPassword = async () => {
    if (!resetPasswordConfirm) return
    try {
      await adminResetPassword(resetPasswordConfirm.userId, resetPasswordConfirm.value)
      toast('密码已重置', 'success')
      setResetPasswordConfirm(null)
      setResetPasswordValue('')
    } catch (err) {
      toast(err instanceof Error ? err.message : String(err), 'error')
    }
  }

  // Re-load all analytics when range changes while already on analytics tab
  useEffect(() => {
    if (tab === 'analytics') {
      loadAnalyticsSummary(analyticsRange)
      loadAnalyticsTrend(analyticsRange)
      loadAnalyticsEndpointBreakdown(analyticsRange)
      loadAnalyticsUserBreakdown(analyticsRange)
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [analyticsRange])

  // ─── Money formatting for analytics (uses response meta.moneyScale) ───
  const formatMoneyX10000 = useCallback((value: number, moneyScale?: number): string => {
    const scale = moneyScale && moneyScale > 0 ? moneyScale : 10000
    const negative = value < 0
    const abs = negative ? -value : value
    const integerPart = Math.floor(abs / scale)
    const fractionalPart = abs % scale

    if (fractionalPart === 0) return negative ? `-${integerPart}` : `${integerPart}`

    let fracStr = fractionalPart.toString()
    // Scale fractional digits based on moneyScale (e.g. 10000 = 4 digits)
    const fracDigits = scale.toString().length - 1
    fracStr = fracStr.padStart(fracDigits, '0').replace(/0+$/, '')
    if (fracStr === '') return negative ? `-${integerPart}` : `${integerPart}`
    return negative ? `-${integerPart}.${fracStr}` : `${integerPart}.${fracStr}`
  }, [])

  const formatMoneyInAnalytics = useCallback((value: number): string => {
    // Prefer summary meta moneyScale, fallback to trend meta, fallback to 10000
    const scale = summaryMeta?.moneyScale ?? trendMeta?.moneyScale ?? 10000
    return formatMoneyX10000(value, scale)
  }, [summaryMeta, trendMeta, formatMoneyX10000])

  const formatProfitClass = (value: number): string => {
    if (value > 0) return 'text-green-800 dark:text-green-400 tabular-nums'
    if (value < 0) return 'text-red-500 dark:text-red-400 tabular-nums'
    return 'text-gray-500 dark:text-gray-400 tabular-nums'
  }

  const formatProfitRate = (bps: number): string => {
    const pct = bps / 100
    return pct.toFixed(2) + '%'
  }

  const renderSkeletonRows = (count: number, cols: number) => {
    return Array.from({ length: count }).map((_, r) => (
      <div key={r} className="flex gap-4 animate-pulse">
        {Array.from({ length: cols }).map((_, c) => (
          <div key={c} className="flex-1 h-4 bg-gray-100 dark:bg-white/5 rounded" />
        ))}
      </div>
    ))
  }

  const formatTime = (ms: number | null) => {
    if (!ms) return '-'
    return new Date(ms).toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
  }
  const getQuotaDisplay = (user: AdminUser) => { if (user.quota === 0) return `${user.usedCount} / 无限制`; return `${user.usedCount} / ${user.quota}` }
  const getFeedbackCategoryLabel = (feedback: BugFeedback) => feedback.category === 'feature' ? '功能建议' : 'Bug 反馈'
  const getFeedbackStatusLabel = (status: BugFeedbackStatus) => status === 'resolved' ? '已解决' : status === 'reviewing' ? '处理中' : '待处理'
  const getFeedbackStatusClass = (status: BugFeedbackStatus) => status === 'resolved' ? 'bg-green-500/10 text-green-800 dark:text-green-400' : status === 'reviewing' ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400' : 'bg-orange-500/10 text-orange-600 dark:text-orange-400'
  const getChangelogTitle = (entry: ChangelogEntry) => entry.title || '更新日志'

  const themeOptions: Array<{ value: ThemeMode; label: string }> = [
    { value: 'system', label: '跟随系统' },
    { value: 'light', label: '浅色' },
    { value: 'dark', label: '深色' },
  ]

  if (loading){
    return (<div className="min-h-screen bg-gray-50 dark:bg-gray-950 flex items-center justify-center"><div className="text-gray-500 dark:text-gray-400">加载中...</div></div>)
  }

  return (
    <div className="min-h-screen bg-gray-50 text-gray-900 dark:bg-gray-950 dark:text-gray-100">
      <header className="border-b border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <h1 className="text-lg font-semibold text-gray-900 dark:text-gray-100">管理后台</h1>
          <div className="flex items-center gap-4">
            <Select
              value={settings.theme}
              onChange={(value) => setSettings({ theme: value as ThemeMode })}
              options={themeOptions}
              className="h-8 w-28"
            />
            <button onClick={() => {
              if (tab === 'analytics') {
                loadAnalyticsSummary(analyticsRange)
                loadAnalyticsTrend(analyticsRange)
                loadAnalyticsEndpointBreakdown(analyticsRange)
                loadAnalyticsUserBreakdown(analyticsRange)
                toast('统计已刷新', 'success')
              } else {
                loadUsers(); loadCodes(); loadFeedbacks(); loadChangelogs()
              }
            }} className="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">刷新</button>
            <button onClick={handleLogout} className="text-sm text-red-500 dark:text-red-400 hover:text-red-600 dark:hover:text-red-300">退出登录</button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-6">
        <Tabs value={tab} onValueChange={(v) => { setTab(v as Tab); setSelectedUserIds(new Set()); setSelectedCodeIds(new Set()) }}>
          <TabsList className="mb-6 flex-wrap h-auto bg-gray-100 dark:bg-gray-100 dark:bg-white/5">
            <TabsTrigger value="users" className="data-[state=active]:bg-blue-600 data-[state=active]:text-white">用户管理</TabsTrigger>
            <TabsTrigger value="codes" className="data-[state=active]:bg-blue-600 data-[state=active]:text-white">兑换码管理</TabsTrigger>
            <TabsTrigger value="config" className="data-[state=active]:bg-blue-600 data-[state=active]:text-white">系统配置</TabsTrigger>
            <TabsTrigger value="analytics" className="data-[state=active]:bg-blue-600 data-[state=active]:text-white">成本收益统计</TabsTrigger>
            <TabsTrigger value="announcement" className="data-[state=active]:bg-blue-600 data-[state=active]:text-white">公告管理</TabsTrigger>
            <TabsTrigger value="feedback" className="data-[state=active]:bg-blue-600 data-[state=active]:text-white">反馈管理</TabsTrigger>
            <TabsTrigger value="changelog" className="data-[state=active]:bg-blue-600 data-[state=active]:text-white">更新日志</TabsTrigger>
            <TabsTrigger value="invites" className="data-[state=active]:bg-blue-600 data-[state=active]:text-white">邀请码设置</TabsTrigger>
          </TabsList>
        </Tabs>

        {tab === 'users' && (
          <>
            {selectedUserIds.size > 0 && (
              <div className="mb-4 flex items-center gap-3">
                <span className="text-sm text-gray-500 dark:text-gray-400">已选 {selectedUserIds.size} 项</span>
                <button onClick={() => setBatchConfirm({ type: 'users', count: selectedUserIds.size })} className="rounded-xl bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700">批量删除</button>
                <button onClick={() => setSelectedUserIds(new Set())} className="rounded-xl bg-gray-100 dark:bg-white/5 px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-white/10">取消选择</button>
              </div>
            )}
            <div className="overflow-x-auto rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03]">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] text-left text-gray-500 dark:text-gray-400">
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
                    <tr key={user.id} className={`border-b border-gray-200/70 dark:border-gray-200/50 dark:border-white/5 hover:bg-gray-50 dark:hover:bg-white/[0.02] ${selectedUserIds.has(user.id) ? 'bg-blue-50 dark:bg-blue-500/5' : ''}`}>
                      <td className="px-4 py-3"><input type="checkbox" checked={selectedUserIds.has(user.id)} onChange={() => toggleUserSelect(user.id)} className="accent-blue-500" /></td>
                      <td className="px-4 py-3"><div className="font-medium">{user.label}</div><div className="text-xs text-gray-500">{user.role}</div></td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400">{formatTime(user.createdAt)}</td>
                      <td className="px-4 py-3"><span className={user.quota > 0 && user.usedCount >= user.quota ? 'text-red-500 dark:text-red-400' : ''}>{getQuotaDisplay(user)}</span></td>
                      <td className="px-4 py-3"><span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${user.status === 'active' ? 'bg-green-500/10 text-green-800 dark:text-green-400' : 'bg-red-500/10 text-red-500 dark:text-red-400'}`}>{user.status === 'active' ? '正常' : '已禁用'}</span></td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <button onClick={() => { setQuotaModal({ user, mode: 'increase' }); setQuotaValue('') }} className="rounded-lg bg-blue-600/20 px-2 py-1 text-xs text-blue-600 dark:text-blue-400 hover:bg-blue-600/30">增加</button>
                          <button onClick={() => { setQuotaModal({ user, mode: 'decrease' }); setQuotaValue('') }} className="rounded-lg bg-orange-600/20 px-2 py-1 text-xs text-orange-600 dark:text-orange-400 hover:bg-orange-600/30">减少</button>
                          <button onClick={() => { setQuotaModal({ user, mode: 'set' }); setQuotaValue('') }} className="rounded-lg bg-purple-600/20 px-2 py-1 text-xs text-purple-600 dark:text-purple-400 hover:bg-purple-600/30">设定</button>
                          <button onClick={() => handleReset(user.id)} className="rounded-lg bg-gray-600/20 px-2 py-1 text-xs text-gray-500 dark:text-gray-400 hover:bg-gray-600/30">重置</button>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          <button onClick={() => { setResetPasswordModal({ userId: user.id, label: user.label }); setResetPasswordValue('') }}
                            className="rounded-lg bg-blue-600/20 px-3 py-1 text-xs font-medium text-blue-600 dark:text-blue-400 hover:bg-blue-600/30">
                            重置密码
                          </button>
                          <button onClick={() => handleToggleStatus(user)} className={`rounded-lg px-3 py-1 text-xs font-medium ${user.status === 'active' ? 'bg-orange-600/20 text-orange-600 dark:text-orange-400 hover:bg-orange-600/30' : 'bg-green-600/20 text-green-800 dark:text-green-400 hover:bg-green-600/30'}`}>{user.status === 'active' ? '禁用' : '启用'}</button>
                          <button onClick={() => handleDeleteUser(user)} className="rounded-lg bg-red-600/20 px-3 py-1 text-xs font-medium text-red-500 dark:text-red-400 hover:bg-red-600/30">删除</button>
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
            <div className="mb-6 rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-4">
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-100 mb-3">创建兑换码</h3>
              <div className="flex flex-wrap items-end gap-3">
                <div><label className="block text-xs text-gray-500 mb-1">每张码的图片数</label><input type="number" value={codeQuota} onChange={e => setCodeQuota(e.target.value)} placeholder="如 100" className="w-32 rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 text-sm text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400" /></div>
                <div><label className="block text-xs text-gray-500 mb-1">数量</label><input type="number" value={codeCount} onChange={e => setCodeCount(e.target.value)} placeholder="1" className="w-24 rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 text-sm text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400" /></div>
                <button onClick={handleCreateCodes} disabled={creating || !codeQuota.trim()} className="rounded-xl bg-blue-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50">{creating ? '创建中...' : '创建并复制'}</button>
                <button onClick={handleCopyUnused} className="rounded-xl bg-gray-100 dark:bg-white/5 px-4 py-2 text-sm text-gray-700 dark:text-gray-300 transition hover:bg-gray-200 dark:hover:bg-white/10">复制全部未使用码</button>
              </div>
            </div>

            {codes.length > 0 && (
              <div className="mb-4 flex flex-wrap items-center gap-2">
                <span className="text-xs text-gray-500">按额度筛选:</span>
                <button onClick={() => setCodeFilter(null)} className={`rounded-lg px-3 py-1 text-xs font-medium transition ${codeFilter === null ? 'bg-blue-600 text-white' : 'bg-gray-100 dark:bg-white/5 text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-white/10'}`}>全部</button>
                {Array.from(new Set(codes.map(c => c.quota))).sort((a, b) => a - b).map(q => (
                  <button key={q} onClick={() => setCodeFilter(q)} className={`rounded-lg px-3 py-1 text-xs font-medium transition ${codeFilter === q ? 'bg-blue-600 text-white' : 'bg-gray-100 dark:bg-white/5 text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-white/10'}`}>{q} 张</button>
                ))}
              </div>
            )}

            {selectedCodeIds.size > 0 && (
              <div className="mb-4 flex items-center gap-3">
                <span className="text-sm text-gray-500 dark:text-gray-400">已选 {selectedCodeIds.size} 项</span>
                <button onClick={() => setBatchConfirm({ type: 'codes', count: selectedCodeIds.size })} className="rounded-xl bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700">批量删除</button>
                <button onClick={() => setSelectedCodeIds(new Set())} className="rounded-xl bg-gray-100 dark:bg-white/5 px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-white/10">取消选择</button>
              </div>
            )}

            <div className="overflow-x-auto rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03]">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] text-left text-gray-500 dark:text-gray-400">
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
                    <tr key={code.id} className={`border-b border-gray-200/70 dark:border-gray-200/50 dark:border-white/5 hover:bg-gray-50 dark:hover:bg-white/[0.02] ${selectedCodeIds.has(code.id) ? 'bg-blue-50 dark:bg-blue-500/5' : ''}`}>
                      <td className="px-4 py-3"><input type="checkbox" checked={selectedCodeIds.has(code.id)} onChange={() => toggleCodeSelect(code.id)} className="accent-blue-500" /></td>
                      <td className="px-4 py-3"><button onClick={() => handleCopyCode(code.code)} className="font-mono text-xs bg-gray-100 dark:bg-white/5 px-2 py-0.5 rounded hover:bg-gray-200 dark:hover:bg-white/10 transition cursor-pointer" title="点击复制">{code.code}</button></td>
                      <td className="px-4 py-3">{code.quota}</td>
                      <td className="px-4 py-3"><span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${code.usedBy ? 'bg-gray-500/10 text-gray-500 dark:text-gray-400' : 'bg-green-500/10 text-green-800 dark:text-green-400'}`}>{code.usedBy ? '已使用' : '未使用'}</span></td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">{code.usedBy || '-'}</td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">{formatTime(code.usedAt)}</td>
                      <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">{formatTime(code.createdAt)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {filteredCodes.length === 0 && <div className="py-12 text-center text-gray-500">暂无兑换码</div>}
          </>
        )}

        {tab === 'config' && (
          <div className="space-y-5">
            <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h3 className="text-sm font-medium text-gray-800 dark:text-gray-100">API 端点池</h3>
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
                    <div key={i} className="flex items-start gap-3 rounded-xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/70 dark:bg-white/[0.02] p-4">
                      <div className="flex-1 space-y-2">
                        <div><label className="block text-xs text-gray-500 mb-1">Base URL</label><input value={ep.baseUrl} onChange={e => handleEndpointChange(i, 'baseUrl', e.target.value)} placeholder="https://api.openai.com/v1" className="w-full rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 text-sm text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400 font-mono" /></div>
                        <div>
                          <label className="block text-xs text-gray-500 mb-1">API Key</label>
                          <div className="relative">
                            <input type={visibleKeys.has(i) ? 'text' : 'password'} value={ep.apiKey} onChange={e => handleEndpointChange(i, 'apiKey', e.target.value)} placeholder="sk-..." className="w-full rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 pr-10 text-sm text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400 font-mono" />
                            <button type="button" onClick={() => setVisibleKeys(prev => { const next = new Set(prev); if (next.has(i)) next.delete(i); else next.add(i); return next })} className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-gray-500 hover:text-gray-700 dark:text-gray-300 transition">
                              {visibleKeys.has(i) ? (
                                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                              ) : (
                                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" /></svg>
                              )}
                            </button>
                          </div>
                        </div>
                        <div className="flex flex-wrap gap-3">
                          <div><label className="block text-xs text-gray-500 mb-1">最大并发数（0 = 无限制）</label><input type="number" min="0" value={ep.maxConcurrency ?? ''} onChange={e => handleEndpointChange(i, 'maxConcurrency', e.target.value)} placeholder="0" className="w-32 rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 text-sm text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400 font-mono" /></div>
                          <div><label className="block text-xs text-gray-500 mb-1">优先级（越大越优先）</label><input type="number" min="0" value={ep.priority ?? ''} onChange={e => handleEndpointChange(i, 'priority', e.target.value)} placeholder="0" className="w-32 rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 text-sm text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400 font-mono" /></div>
                          <div>
                            <label className="block text-xs text-gray-500 mb-1">成本价（元/张）</label>
                            <input type="number" min="0" step="0.0001" value={costInputDrafts[i] ?? '0'} onChange={e => handleEndpointCostChange(i, e.target.value)} placeholder="0" className="w-36 rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 text-sm text-right text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400 tabular-nums" />
                            {priceErrors[i] && <p className="mt-1 text-xs text-red-500 dark:text-red-400">{priceErrors[i]}</p>}
                          </div>
                        </div>
                      </div>
                      <button onClick={() => handleRemoveEndpoint(i)} className="mt-6 rounded-lg bg-red-600/20 px-3 py-2 text-xs font-medium text-red-500 dark:text-red-400 hover:bg-red-600/30 transition">删除</button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Global Sale Price Card */}
            <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
              <div>
                <h3 className="text-sm font-medium text-gray-800 dark:text-gray-100">全局售价（元/张）</h3>
                <p className="text-xs text-gray-500 mt-1">支持 4 位小数</p>
              </div>
              <div className="mt-4">
                <label className="block text-xs text-gray-500 mb-1">全局售价（元/张）</label>
                <input
                  type="number"
                  min="0"
                  step="0.0001"
                  value={salePriceInput}
                  onChange={e => setSalePriceInput(e.target.value)}
                  placeholder="0"
                  className="w-48 rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 text-sm text-right text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400 tabular-nums"
                />
                {(() => {
                  const saleErr = (() => {
                    if (salePriceInput.trim() === '') return '请输入非负数字，最多 4 位小数'
                    if (parseMoneyInputToX10000(salePriceInput) === null) return '请输入非负数字，最多 4 位小数'
                    return null
                  })()
                  return saleErr ? <p className="mt-1 text-xs text-red-500 dark:text-red-400">{saleErr}</p> : null
                })()}
              </div>
            </div>

            {/* Save button visible when there are endpoints */}
            {endpoints.length > 0 && (() => {
              const anyCostInvalid = endpoints.some((_, i) => {
                const draft = costInputDrafts[i] ?? ''
                return draft.trim() === '' || parseMoneyInputToX10000(draft) === null
              })
              const saleInvalid = parseMoneyInputToX10000(salePriceInput) === null
              const hasInvalid = anyCostInvalid || saleInvalid
              const hasErrors = Object.keys(priceErrors).length > 0
              return (
              <div className="flex justify-end">
                <button
                  onClick={handleSavePricingConfig}
                  disabled={pricingSaving || hasInvalid || hasErrors}
                  className="rounded-xl bg-green-600 px-5 py-2 text-sm font-medium text-white transition hover:bg-green-700 disabled:cursor-not-allowed disabled:opacity-50"
                >{pricingSaving ? '保存中...' : '保存配置'}</button>
              </div>
            )})()}
          </div>
        )}

        {/* ─── Analytics Tab ─── */}
        {tab === 'analytics' && (
          <div className="space-y-6">
            {/* ─── KPI Cards ─── */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              {summaryLoading && !summary ? (
                <>
                  {['总收入', '总成本', '利润', '成功图片数'].map(label => (
                    <div key={label} className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5 animate-pulse">
                      <div className="h-3 w-16 bg-gray-100 dark:bg-white/5 rounded mb-3" />
                      <div className="h-7 w-24 bg-gray-100 dark:bg-white/5 rounded" />
                    </div>
                  ))}
                </>
              ) : summary ? (
                <>
                  <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
                    <div className="text-xs text-gray-500 mb-1">总收入</div>
                    <div className="text-[28px] font-semibold tabular-nums leading-tight text-blue-600 dark:text-blue-400">{formatMoneyInAnalytics(summary.revenueX10000)}</div>
                    <div className="text-xs text-gray-500 mt-1">元</div>
                  </div>
                  <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
                    <div className="text-xs text-gray-500 mb-1">总成本</div>
                    <div className="text-[28px] font-semibold tabular-nums leading-tight text-amber-600 dark:text-amber-400">{formatMoneyInAnalytics(summary.costX10000)}</div>
                    <div className="text-xs text-gray-500 mt-1">元</div>
                  </div>
                  <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
                    <div className="text-xs text-gray-500 mb-1">利润</div>
                    <div className={`text-[28px] font-semibold tabular-nums leading-tight ${formatProfitClass(summary.profitX10000)}`}>{formatMoneyInAnalytics(summary.profitX10000)}</div>
                    <div className="text-xs text-gray-500 mt-1">元</div>
                  </div>
                  <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
                    <div className="text-xs text-gray-500 mb-1">成功图片数</div>
                    <div className="text-[28px] font-semibold tabular-nums leading-tight text-violet-600 dark:text-violet-400">{summary.successImages.toLocaleString()}</div>
                    <div className="text-xs text-gray-500 mt-1">张</div>
                  </div>
                </>
              ) : null}
            </div>

            {summaryError && !summary && (
              <div className="rounded-2xl border border-red-500/30 bg-red-500/[0.04] p-4">
                <p className="text-sm text-red-500 dark:text-red-400">{summaryError}</p>
                <button onClick={() => { loadAnalyticsSummary(analyticsRange); loadAnalyticsTrend(analyticsRange); loadAnalyticsEndpointBreakdown(analyticsRange); loadAnalyticsUserBreakdown(analyticsRange) }} className="mt-2 rounded-xl bg-red-600/20 px-4 py-2 text-sm font-medium text-red-500 dark:text-red-400 hover:bg-red-600/30">刷新统计</button>
              </div>
            )}

            {/* ─── Trend Card with Range Filters and Chart ─── */}
            <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
              <div className="flex flex-wrap items-center justify-between gap-3 mb-5">
                <div className="flex gap-2">
                  {(['today', '7d', '30d', 'all'] as const).map(r => {
                    const labelMap: Record<string, string> = { today: '今日', '7d': '7天', '30d': '30天', all: '全部' }
                    return (
                      <button
                        key={r}
                        onClick={() => setAnalyticsRange(r)}
                        className={`rounded-xl px-4 py-2 text-sm font-medium transition ${analyticsRange === r ? 'bg-blue-600 text-white' : 'bg-gray-100 dark:bg-white/5 text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-white/10'}`}
                      >{labelMap[r]}</button>
                    )
                  })}
                </div>
                <button
                  onClick={() => { loadAnalyticsSummary(analyticsRange); loadAnalyticsTrend(analyticsRange); loadAnalyticsEndpointBreakdown(analyticsRange); loadAnalyticsUserBreakdown(analyticsRange); toast('统计已刷新', 'success') }}
                  className="rounded-xl bg-gray-100 dark:bg-white/5 px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-white/10 transition"
                >刷新统计</button>
              </div>

              {trendLoading ? (
                <div className="space-y-2">{renderSkeletonRows(4, 4)}</div>
              ) : trendError && trend.length === 0 ? (
                <div className="py-8 text-center">
                  <p className="text-sm text-red-500 dark:text-red-400 mb-3">{trendError}</p>
                  <button onClick={() => { loadAnalyticsSummary(analyticsRange); loadAnalyticsTrend(analyticsRange); loadAnalyticsEndpointBreakdown(analyticsRange); loadAnalyticsUserBreakdown(analyticsRange) }} className="rounded-xl bg-red-600/20 px-4 py-2 text-sm font-medium text-red-500 dark:text-red-400 hover:bg-red-600/30">刷新统计</button>
                </div>
              ) : trend.length > 0 ? (
                <div className="space-y-6">
                  {/* Money trend chart (revenue, cost, profit) */}
                  <div className="relative w-full h-64">
                    <svg viewBox="0 0 800 240" className="w-full h-full" role="img" aria-label="收入/成本/利润趋势图">
                      <defs>
                        <linearGradient id="revArea" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stopColor="#2563eb" stopOpacity="0.15" /><stop offset="100%" stopColor="#2563eb" stopOpacity="0" /></linearGradient>
                        <linearGradient id="costArea" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stopColor="#f59e0b" stopOpacity="0.15" /><stop offset="100%" stopColor="#f59e0b" stopOpacity="0" /></linearGradient>
                        <linearGradient id="profitArea" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stopColor="#059669" stopOpacity="0.15" /><stop offset="100%" stopColor="#059669" stopOpacity="0" /></linearGradient>
                      </defs>
                      {/* Grid lines */}
                      {[0, 1, 2, 3, 4].map(i => (
                        <line key={`grid-${i}`} x1={60} y1={20 + i * 50} x2={780} y2={20 + i * 50} stroke="currentColor" strokeOpacity={0.06} strokeWidth={1} className="text-black dark:text-white" />
                      ))}
                      {/* Money axes labels */}
                      {(() => {
                        const moneyVals = summary ? [summary.revenueX10000, summary.costX10000, Math.abs(summary.profitX10000)] : trend.length > 0 ? trend.flatMap(p => [Math.abs(p.revenueX10000), Math.abs(p.costX10000), Math.abs(p.profitX10000)]) : [0]
                        const maxMoney = Math.max(...moneyVals, 1)
                        const step = maxMoney / 4
                        return [0, 1, 2, 3, 4].map(i => (
                          <text key={`yl-${i}`} x={56} y={24 + i * 50} textAnchor="end" fill="#9ca3af" className="text-[10px] dark:fill-gray-500">{step > 0 ? formatMoneyInAnalytics(Math.round((4 - i) * step)) : '0'}</text>
                        ))
                      })()}
                      {/* Revenue line */}
                      <polyline
                        points={trend.map((p, i) => {
                          const maxM = Math.max(...trend.flatMap(p => [Math.abs(p.revenueX10000), Math.abs(p.costX10000), Math.abs(p.profitX10000)]), 1)
                          const x = 60 + i * (720 / Math.max(trend.length - 1, 1))
                          const y = 20 + 200 * (1 - p.revenueX10000 / maxM)
                          return `${x},${y}`
                        }).join(' ')}
                        fill="url(#revArea)" stroke="#2563eb" strokeWidth={2} strokeLinejoin="round" strokeLinecap="round"
                      />
                      {/* Cost line */}
                      <polyline
                        points={trend.map((p, i) => {
                          const maxM = Math.max(...trend.flatMap(p => [Math.abs(p.revenueX10000), Math.abs(p.costX10000), Math.abs(p.profitX10000)]), 1)
                          const x = 60 + i * (720 / Math.max(trend.length - 1, 1))
                          const y = 20 + 200 * (1 - p.costX10000 / maxM)
                          return `${x},${y}`
                        }).join(' ')}
                        fill="url(#costArea)" stroke="#f59e0b" strokeWidth={2} strokeLinejoin="round" strokeLinecap="round"
                      />
                      {/* Profit line */}
                      <polyline
                        points={trend.map((p, i) => {
                          const maxM = Math.max(...trend.flatMap(p => [Math.abs(p.revenueX10000), Math.abs(p.costX10000), Math.abs(p.profitX10000)]), 1)
                          const x = 60 + i * (720 / Math.max(trend.length - 1, 1))
                          const y = 20 + 200 * (1 - p.profitX10000 / maxM)
                          return `${x},${y}`
                        }).join(' ')}
                        fill="url(#profitArea)" stroke="#059669" strokeWidth={2} strokeLinejoin="round" strokeLinecap="round"
                      />
                      {/* X-axis labels */}
                      {trend.map((p, i) => (
                        <text key={`xl-${i}`} x={60 + i * (720 / Math.max(trend.length - 1, 1))} y={234} textAnchor="middle" fill="#9ca3af" className="text-[9px] dark:fill-gray-500">
                          {p.bucket.length > 10 ? p.bucket.slice(-5) : p.bucket}
                        </text>
                      ))}
                    </svg>
                    {/* Legend */}
                    <div className="flex flex-wrap justify-center gap-4 mt-2">
                      <span className="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400"><span className="w-3 h-3 rounded-full bg-blue-600 inline-block" /> 收入</span>
                      <span className="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400"><span className="w-3 h-3 rounded-full bg-amber-500 inline-block" /> 成本</span>
                      <span className="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400"><span className="w-3 h-3 rounded-full bg-emerald-600 inline-block" /> 利润</span>
                    </div>
                  </div>

                  {/* Success images chart (bar chart) */}
                  {trend.some(p => p.successImages > 0) && (
                    <div className="relative w-full h-32 mt-4 border-t border-gray-200/70 dark:border-gray-200/50 dark:border-white/5 pt-4">
                      <div className="flex items-end gap-1 h-24 px-4">
                        {trend.map((p, i) => {
                          const maxImgCount = Math.max(...trend.map(p => p.successImages), 1)
                          const hPct = (p.successImages / maxImgCount) * 100
                          return (
                            <div key={i} className="flex-1 flex flex-col items-center justify-end" title={`${p.bucket}: ${p.successImages} 张`}>
                              <span className="text-[9px] text-gray-500 mb-1">{p.successImages}</span>
                              <div className="w-full max-w-[24px] rounded-t-sm bg-violet-500/60" style={{ height: `${Math.max(hPct, 2)}%` }} />
                            </div>
                          )
                        })}
                      </div>
                      <div className="flex items-center gap-1.5 justify-end mt-1">
                        <span className="flex items-center gap-1 text-xs text-gray-500"><span className="w-2.5 h-2.5 rounded-sm bg-violet-500/60 inline-block" /> 成功图片数</span>
                      </div>
                    </div>
                  )}
                </div>
              ) : !trendLoading && !trendError ? (
                // Empty data
                <div className="py-10 text-center">
                  <h4 className="text-base font-medium text-gray-500 dark:text-gray-400 mb-2">暂无成本收益数据</h4>
                  <p className="text-sm text-gray-500 max-w-md mx-auto">完成图片生成并保存配置后，这里将显示收入、成本、利润和成功图片数。</p>
                </div>
              ) : null}
            </div>

            {/* ─── Endpoint & User Breakdown Tables ─── */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              {/* Endpoint Breakdown */}
              <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
                <h3 className="text-base font-medium text-gray-800 dark:text-gray-100 mb-4">端点拆分</h3>
                {endpointLoading ? (
                  <div className="space-y-2">{renderSkeletonRows(4, 6)}</div>
                ) : endpointError && endpointRows.length === 0 ? (
                  <div className="py-6 text-center">
                    <p className="text-sm text-red-500 dark:text-red-400 mb-3">统计数据加载失败，请点击"刷新统计"重试；保存配置失败时，请检查金额是否为数字且最多 4 位小数。</p>
                    <button onClick={() => { loadAnalyticsSummary(analyticsRange); loadAnalyticsTrend(analyticsRange); loadAnalyticsEndpointBreakdown(analyticsRange); loadAnalyticsUserBreakdown(analyticsRange) }} className="rounded-xl bg-red-600/20 px-4 py-2 text-sm font-medium text-red-500 dark:text-red-400 hover:bg-red-600/30">刷新统计</button>
                  </div>
                ) : endpointRows.length > 0 ? (
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] text-left text-gray-500 dark:text-gray-400">
                          <th className="px-3 py-2 font-medium text-xs">端点标识</th>
                          <th className="px-3 py-2 font-medium text-xs text-right">成功图片数</th>
                          <th className="px-3 py-2 font-medium text-xs text-right">收入</th>
                          <th className="px-3 py-2 font-medium text-xs text-right">成本</th>
                          <th className="px-3 py-2 font-medium text-xs text-right">利润</th>
                          <th className="px-3 py-2 font-medium text-xs text-right">利润率</th>
                        </tr>
                      </thead>
                      <tbody>
                        {endpointRows.map((row, i) => {
                          const moneyScale = endpointMeta?.moneyScale ?? 10000
                          return (
                            <tr key={i} className="border-b border-gray-200/70 dark:border-gray-200/50 dark:border-white/5 hover:bg-gray-50 dark:hover:bg-white/[0.02]">
                              <td className="px-3 py-2.5 text-xs font-mono text-gray-700 dark:text-gray-300 max-w-[160px] truncate" title={row.endpointLabel || row.endpointBaseUrl}>{row.endpointLabel || row.endpointBaseUrl}</td>
                              <td className="px-3 py-2.5 text-xs text-right tabular-nums text-gray-700 dark:text-gray-300">{row.successImages.toLocaleString()}</td>
                              <td className="px-3 py-2.5 text-xs text-right tabular-nums text-blue-600 dark:text-blue-400">{formatMoneyX10000(row.revenueX10000, moneyScale)}</td>
                              <td className="px-3 py-2.5 text-xs text-right tabular-nums text-amber-600 dark:text-amber-400">{formatMoneyX10000(row.costX10000, moneyScale)}</td>
                              <td className={`px-3 py-2.5 text-xs text-right ${formatProfitClass(row.profitX10000)}`}>{formatMoneyX10000(row.profitX10000, moneyScale)}</td>
                              <td className="px-3 py-2.5 text-xs text-right tabular-nums text-gray-700 dark:text-gray-300">{formatProfitRate(row.profitRateBps)}</td>
                            </tr>
                          )
                        })}
                      </tbody>
                    </table>
                  </div>
                ) : !endpointLoading && !endpointError ? (
                  <div className="py-6 text-center text-sm text-gray-500">暂无数据</div>
                ) : null}
              </div>

              {/* User Breakdown */}
              <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
                <h3 className="text-base font-medium text-gray-800 dark:text-gray-100 mb-4">用户拆分</h3>
                {userLoading ? (
                  <div className="space-y-2">{renderSkeletonRows(4, 6)}</div>
                ) : userError && userRows.length === 0 ? (
                  <div className="py-6 text-center">
                    <p className="text-sm text-red-500 dark:text-red-400 mb-3">统计数据加载失败，请点击"刷新统计"重试；保存配置失败时，请检查金额是否为数字且最多 4 位小数。</p>
                    <button onClick={() => { loadAnalyticsSummary(analyticsRange); loadAnalyticsTrend(analyticsRange); loadAnalyticsEndpointBreakdown(analyticsRange); loadAnalyticsUserBreakdown(analyticsRange) }} className="rounded-xl bg-red-600/20 px-4 py-2 text-sm font-medium text-red-500 dark:text-red-400 hover:bg-red-600/30">刷新统计</button>
                  </div>
                ) : userRows.length > 0 ? (
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] text-left text-gray-500 dark:text-gray-400">
                          <th className="px-3 py-2 font-medium text-xs">用户标识</th>
                          <th className="px-3 py-2 font-medium text-xs text-right">成功图片数</th>
                          <th className="px-3 py-2 font-medium text-xs text-right">收入</th>
                          <th className="px-3 py-2 font-medium text-xs text-right">成本</th>
                          <th className="px-3 py-2 font-medium text-xs text-right">利润</th>
                          <th className="px-3 py-2 font-medium text-xs text-right">利润率</th>
                        </tr>
                      </thead>
                      <tbody>
                        {userRows.map((row, i) => {
                          const moneyScale = userMeta?.moneyScale ?? 10000
                          return (
                            <tr key={i} className="border-b border-gray-200/70 dark:border-gray-200/50 dark:border-white/5 hover:bg-gray-50 dark:hover:bg-white/[0.02]">
                              <td className="px-3 py-2.5 text-xs text-gray-700 dark:text-gray-300 max-w-[160px] truncate" title={row.userLabel || row.userId}>{row.userLabel || row.userId}</td>
                              <td className="px-3 py-2.5 text-xs text-right tabular-nums text-gray-700 dark:text-gray-300">{row.successImages.toLocaleString()}</td>
                              <td className="px-3 py-2.5 text-xs text-right tabular-nums text-blue-600 dark:text-blue-400">{formatMoneyX10000(row.revenueX10000, moneyScale)}</td>
                              <td className="px-3 py-2.5 text-xs text-right tabular-nums text-amber-600 dark:text-amber-400">{formatMoneyX10000(row.costX10000, moneyScale)}</td>
                              <td className={`px-3 py-2.5 text-xs text-right ${formatProfitClass(row.profitX10000)}`}>{formatMoneyX10000(row.profitX10000, moneyScale)}</td>
                              <td className="px-3 py-2.5 text-xs text-right tabular-nums text-gray-700 dark:text-gray-300">{formatProfitRate(row.profitRateBps)}</td>
                            </tr>
                          )
                        })}
                      </tbody>
                    </table>
                  </div>
                ) : !userLoading && !userError ? (
                  <div className="py-6 text-center text-sm text-gray-500">暂无数据</div>
                ) : null}
              </div>
            </div>
          </div>
        )}

        {tab === 'changelog' && (
          <div className="space-y-5">
            <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
              <div className="mb-5 flex items-center justify-between gap-4">
                <div>
                  <h3 className="text-sm font-medium text-gray-800 dark:text-gray-100">{editingChangelogId ? '编辑更新日志' : '新增更新日志'}</h3>
                  <p className="mt-1 text-xs text-gray-500">普通前端显示的版本号会以最新已发布日志为准。关闭发布后该版本不会出现在前端。</p>
                </div>
                {editingChangelogId && <button onClick={resetChangelogForm} className="rounded-xl bg-gray-100 dark:bg-white/5 px-4 py-2 text-sm text-gray-700 dark:text-gray-300 transition hover:bg-gray-200 dark:hover:bg-white/10">取消编辑</button>}
              </div>
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <label className="mb-1 block text-xs text-gray-500">版本号</label>
                  <input value={changelogVersion} onChange={e => setChangelogVersion(e.target.value)} placeholder="如 0.2.16" className="w-full rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 text-sm text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400" />
                </div>
                <div>
                  <label className="mb-1 block text-xs text-gray-500">标题（可选）</label>
                  <input value={changelogTitle} onChange={e => setChangelogTitle(e.target.value)} placeholder="如 反馈机制优化" className="w-full rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 text-sm text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400" />
                </div>
              </div>
              <div className="mt-4">
                <label className="mb-1 block text-xs text-gray-500">更新日志内容</label>
                <Textarea value={changelogContent} onChange={e => setChangelogContent(e.target.value)} rows={8} placeholder="输入本次更新内容..." className="w-full resize-y" />
              </div>
              <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
                <div className="flex items-center gap-3">
                  <Switch checked={changelogPublished} onCheckedChange={setChangelogPublished} />
                  <span className="text-sm text-gray-700 dark:text-gray-300">发布到前端</span>
                </div>
                <button onClick={handleSaveChangelog} disabled={changelogSaving} className="rounded-xl bg-green-600 px-5 py-2 text-sm font-medium text-white transition hover:bg-green-700 disabled:cursor-not-allowed disabled:opacity-50">
                  {changelogSaving ? '保存中...' : editingChangelogId ? '保存更新日志' : '创建更新日志'}
                </button>
              </div>
            </div>

            <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
              <div className="mb-5 flex items-center justify-between gap-4">
                <div>
                  <h3 className="text-sm font-medium text-gray-800 dark:text-gray-100">历史版本</h3>
                  <p className="mt-1 text-xs text-gray-500">最新已发布日志会成为普通前端显示的版本号来源。</p>
                </div>
                <button onClick={loadChangelogs} className="rounded-xl bg-gray-100 dark:bg-white/5 px-4 py-2 text-sm text-gray-700 dark:text-gray-300 transition hover:bg-gray-200 dark:hover:bg-white/10">刷新</button>
              </div>
              {changelogLoading ? (
                <div className="py-8 text-center text-gray-500">加载中...</div>
              ) : changelogs.length === 0 ? (
                <div className="py-12 text-center text-gray-500">暂无更新日志</div>
              ) : (
                <div className="space-y-3">
                  {changelogs.map(entry => (
                    <div key={entry.id} className={`rounded-xl border p-4 transition ${editingChangelogId === entry.id ? 'border-blue-500/60 bg-blue-50 dark:bg-blue-500/5' : 'border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/70 dark:bg-white/[0.02]'}`}>
                      <div className="mb-3 flex flex-wrap items-start justify-between gap-3">
                        <div>
                          <div className="flex flex-wrap items-center gap-2">
                            <span className="font-mono text-sm font-medium text-gray-900 dark:text-gray-100">v{entry.version || '-'}</span>
                            <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${entry.published ? 'bg-green-500/10 text-green-800 dark:text-green-400' : 'bg-gray-500/10 text-gray-500 dark:text-gray-400'}`}>{entry.published ? '已发布' : '草稿'}</span>
                          </div>
                          <div className="mt-1 text-sm text-gray-700 dark:text-gray-300">{getChangelogTitle(entry)}</div>
                        </div>
                        <div className="flex items-center gap-2">
                          <button onClick={() => openChangelogEditor(entry)} className="rounded-lg bg-blue-600/20 px-3 py-1 text-xs font-medium text-blue-600 dark:text-blue-400 hover:bg-blue-600/30">编辑</button>
                          <button onClick={() => setDeleteChangelogId(entry.id)} className="rounded-lg bg-red-600/20 px-3 py-1 text-xs font-medium text-red-500 dark:text-red-400 hover:bg-red-600/30">删除</button>
                        </div>
                      </div>
                      <div className="mb-3 line-clamp-3 whitespace-pre-wrap break-words text-sm leading-6 text-gray-500 dark:text-gray-400">{entry.content || '暂无内容'}</div>
                      <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-gray-500">
                        <span>发布时间：{formatTime(entry.publishedAt)}</span>
                        <span>更新时间：{formatTime(entry.updatedAt)}</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {tab === 'feedback' && (
          <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
            <div className="mb-5 flex items-center justify-between gap-4">
              <div>
                <h3 className="text-sm font-medium text-gray-800 dark:text-gray-100">反馈管理</h3>
                <p className="mt-1 text-xs text-gray-500">查看用户提交的 Bug 反馈和功能建议。</p>
              </div>
              <button onClick={loadFeedbacks} className="rounded-xl bg-gray-100 dark:bg-white/5 px-4 py-2 text-sm text-gray-700 dark:text-gray-300 transition hover:bg-gray-200 dark:hover:bg-white/10">刷新</button>
            </div>
            {feedbacksLoading ? (
              <div className="py-8 text-center text-gray-500">加载中...</div>
            ) : feedbacks.length === 0 ? (
              <div className="py-12 text-center text-gray-500">暂无反馈</div>
            ) : (
              <div className="space-y-3">
                {feedbacks.map(feedback => (
                  <div key={feedback.id} className="rounded-xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/70 dark:bg-white/[0.02] p-4">
                    <div className="mb-3 flex flex-wrap items-start justify-between gap-3">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${feedback.category === 'feature' ? 'bg-purple-500/10 text-purple-600 dark:text-purple-400' : 'bg-red-500/10 text-red-500 dark:text-red-400'}`}>{getFeedbackCategoryLabel(feedback)}</span>
                        <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-medium ${getFeedbackStatusClass(feedback.status)}`}>{getFeedbackStatusLabel(feedback.status)}</span>
                        <span className="text-xs text-gray-500">{formatTime(feedback.createdAt)}</span>
                      </div>
                      <select
                        value={feedback.status}
                        onChange={e => handleUpdateFeedbackStatus(feedback.id, e.target.value as BugFeedbackStatus)}
                        disabled={feedbackUpdatingId === feedback.id}
                        className="rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-gray-900 px-2 py-1 text-xs text-gray-800 dark:text-gray-300 outline-none transition focus:border-blue-400 disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        <option value="open">待处理</option>
                        <option value="reviewing">处理中</option>
                        <option value="resolved">已解决</option>
                      </select>
                    </div>
                    <div className="mb-3 text-sm leading-6 text-gray-800 dark:text-gray-300 whitespace-pre-wrap break-words">{feedback.content}</div>
                    <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-gray-500">
                      <span>用户：{feedback.userLabel || feedback.userId}</span>
                      <span>联系方式：{feedback.contact || '-'}</span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {tab === 'announcement' && (
          <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
            <div className="mb-5">
              <h3 className="text-sm font-medium text-gray-800 dark:text-gray-100">站点公告</h3>
              <p className="mt-1 text-xs text-gray-500">启用后，用户首次打开或公告更新后会看到一次弹窗，之后可在右上角问号菜单中查看。</p>
            </div>
            {announcementLoading ? (
              <div className="py-8 text-center text-gray-500">加载中...</div>
            ) : (
              <>
                <div className="flex items-center gap-3 mb-4">
                  <Switch checked={announcementEnabled} onCheckedChange={setAnnouncementEnabled} />
                  <span className="text-sm text-gray-700 dark:text-gray-300">启用公告</span>
                </div>
                <div className="mb-4">
                  <label className="mb-1 block text-xs text-gray-500">公告内容</label>
                  <Textarea
                    value={announcementContent}
                    onChange={e => setAnnouncementContent(e.target.value)}
                    rows={8}
                    placeholder="输入公告内容..."
                    className="w-full resize-y"
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

        {tab === 'invites' && (
          <div className="space-y-5">
            {/* 奖励配置区 */}
            <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
              <h3 className="text-sm font-medium text-gray-800 dark:text-gray-100 mb-4">奖励配置</h3>
              {inviteConfigLoading ? (<div className="py-8 text-center text-gray-500">加载中...</div>) : (
                <div className="flex flex-wrap items-end gap-4">
                  <div>
                    <label className="block text-xs text-gray-500 mb-1">邀请人奖励配额（张）</label>
                    <input type="number" min="0" value={inviterReward} onChange={e => setInviterReward(parseInt(e.target.value)||0)}
                      className="w-36 rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 text-sm text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400" />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-500 mb-1">被邀请人奖励配额（张）</label>
                    <input type="number" min="0" value={inviteeReward} onChange={e => setInviteeReward(parseInt(e.target.value)||0)}
                      className="w-36 rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 text-sm text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400" />
                  </div>
                  <div>
                    <label className="block text-xs text-gray-500 mb-1">默认注册配额（张）</label>
                    <input type="number" min="0" value={defaultQuota} onChange={e => setDefaultQuota(parseInt(e.target.value)||0)}
                      className="w-36 rounded-lg border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white dark:bg-white/[0.04] px-3 py-2 text-sm text-gray-900 dark:text-gray-100 outline-none focus:border-blue-400" />
                  </div>
                  <button onClick={handleSaveInviteConfig} disabled={inviteConfigSaving}
                    className="rounded-xl bg-green-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-green-700 disabled:cursor-not-allowed disabled:opacity-50">
                    {inviteConfigSaving ? '保存中...' : '保存配置'}
                  </button>
                </div>
              )}
            </div>

            {/* 邀请码使用列表区 */}
            <div className="rounded-2xl border border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] bg-white/80 dark:bg-white/[0.03] p-5">
              <h3 className="text-sm font-medium text-gray-800 dark:text-gray-100 mb-4">邀请码使用情况</h3>
              {inviteRowsLoading ? (<div className="py-8 text-center text-gray-500">加载中...</div>) : inviteRows.length === 0 ? (
                <div className="py-12 text-center text-gray-500">暂无邀请码使用记录</div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-gray-200/70 dark:border-gray-200/70 dark:border-white/[0.08] text-left text-gray-500 dark:text-gray-400">
                        <th className="px-4 py-3 font-medium text-xs">用户</th>
                        <th className="px-4 py-3 font-medium text-xs">邀请码</th>
                        <th className="px-4 py-3 font-medium text-xs text-right">使用次数</th>
                      </tr>
                    </thead>
                    <tbody>
                      {inviteRows.map((row, i) => (
                        <tr key={i} className="border-b border-gray-200/70 dark:border-gray-200/50 dark:border-white/5 hover:bg-gray-50 dark:hover:bg-white/[0.02]">
                          <td className="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">{row.username}</td>
                          <td className="px-4 py-3"><span className="font-mono text-xs bg-gray-100 dark:bg-white/5 px-2 py-0.5 rounded">{row.inviteCode || '-'}</span></td>
                          <td className="px-4 py-3 text-sm text-right tabular-nums text-gray-700 dark:text-gray-300">{row.usageCount}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        )}

      </main>

      {quotaModal && (
        <AlertDialog open onOpenChange={(open) => { if (!open) setQuotaModal(null) }}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>
                {quotaModal.mode === 'increase' && `增加配额 — ${quotaModal.user.label}`}
                {quotaModal.mode === 'decrease' && `减少配额 — ${quotaModal.user.label}`}
                {quotaModal.mode === 'set' && `设定配额 — ${quotaModal.user.label}`}
              </AlertDialogTitle>
              <AlertDialogDescription>当前配额: {getQuotaDisplay(quotaModal.user)}</AlertDialogDescription>
            </AlertDialogHeader>
            <Input type="number" value={quotaValue} onChange={e => setQuotaValue(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleQuotaConfirm()} placeholder={quotaModal.mode === 'set' ? '输入目标值' : '输入数量'} autoFocus className="w-full mb-4" />
            <AlertDialogFooter>
              <AlertDialogCancel onClick={() => setQuotaModal(null)}>取消</AlertDialogCancel>
              <AlertDialogAction onClick={handleQuotaConfirm}>确认</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}

      {quotaConfirm && (
        <AlertDialog open onOpenChange={(open) => { if (!open) setQuotaConfirm(null) }}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>
                {quotaConfirm.mode === 'increase' && `增加配额 — ${quotaConfirm.user.label}`}
                {quotaConfirm.mode === 'decrease' && `减少配额 — ${quotaConfirm.user.label}`}
                {quotaConfirm.mode === 'set' && `设定配额 — ${quotaConfirm.user.label}`}
              </AlertDialogTitle>
              <AlertDialogDescription>
                当前配额: {getQuotaDisplay(quotaConfirm.user)}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <p className="text-sm text-gray-600 dark:text-gray-300 -mt-2">
              {quotaConfirm.mode === 'increase' && `确定为该用户增加 ${quotaConfirm.value} 张图片配额吗？`}
              {quotaConfirm.mode === 'decrease' && `确定为该用户减少 ${quotaConfirm.value} 张图片配额吗？`}
              {quotaConfirm.mode === 'set' && `确定将该用户配额设为 ${quotaConfirm.value} 张吗？`}
            </p>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={() => setQuotaConfirm(null)}>取消</AlertDialogCancel>
              <AlertDialogAction onClick={handleQuotaSubmit}>确认操作</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}

      {resetQuotaConfirm && (
        <AlertDialog open onOpenChange={(open) => { if (!open) setResetQuotaConfirm(null) }}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>重置配额</AlertDialogTitle>
              <AlertDialogDescription>确定要重置该用户的使用计数吗？</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={() => setResetQuotaConfirm(null)}>取消</AlertDialogCancel>
              <AlertDialogAction onClick={handleResetConfirm}>确认重置</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}

      {confirmModal && (
        <AlertDialog open onOpenChange={(open) => { if (!open) setConfirmModal(null) }}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>
                {confirmModal.action === 'delete' && `删除用户 — ${confirmModal.user.label}`}
                {confirmModal.action === 'disable' && `禁用用户 — ${confirmModal.user.label}`}
                {confirmModal.action === 'enable' && `启用用户 — ${confirmModal.user.label}`}
              </AlertDialogTitle>
              <AlertDialogDescription>
                {confirmModal.action === 'delete' && '删除后该用户及其数据将无法恢复，确定要删除吗？'}
                {confirmModal.action === 'disable' && '禁用后该用户将无法登录，确定要禁用吗？'}
                {confirmModal.action === 'enable' && '确定要启用该用户吗？'}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={() => setConfirmModal(null)}>取消</AlertDialogCancel>
              <AlertDialogAction onClick={handleConfirmAction}
                variant={confirmModal.action === 'delete' ? 'destructive' : confirmModal.action === 'disable' ? undefined : 'default'}
                className={confirmModal.action === 'disable' ? 'bg-orange-600 hover:bg-orange-700' : confirmModal.action === 'enable' ? 'bg-green-600 hover:bg-green-700' : ''}
              >确认</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}

      {deleteChangelogId && (
        <AlertDialog open onOpenChange={(open) => { if (!open) setDeleteChangelogId(null) }}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>删除更新日志</AlertDialogTitle>
              <AlertDialogDescription>确定要删除这条更新日志吗？此操作不可恢复。</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={() => setDeleteChangelogId(null)}>取消</AlertDialogCancel>
              <AlertDialogAction onClick={handleDeleteChangelog} variant="destructive">确认删除</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}

      {batchConfirm && (
        <AlertDialog open onOpenChange={(open) => { if (!open) setBatchConfirm(null) }}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>批量删除{batchConfirm.type === 'users' ? '用户' : '兑换码'}</AlertDialogTitle>
              <AlertDialogDescription>确定要删除选中的 {batchConfirm.count} 个{batchConfirm.type === 'users' ? '用户' : '兑换码'}吗？此操作不可恢复。</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={() => setBatchConfirm(null)}>取消</AlertDialogCancel>
              <AlertDialogAction onClick={handleBatchDelete} variant="destructive">确认删除</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}

      {resetPasswordModal && (
        <AlertDialog open onOpenChange={(open) => { if (!open) setResetPasswordModal(null) }}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>重置密码 — {resetPasswordModal.label}</AlertDialogTitle>
            </AlertDialogHeader>
            <Input type="password" value={resetPasswordValue} onChange={e => setResetPasswordValue(e.target.value)}
              placeholder="输入新密码（至少 8 字符）" autoFocus className="w-full mb-4" />
            <AlertDialogFooter>
              <AlertDialogCancel onClick={() => setResetPasswordModal(null)}>取消</AlertDialogCancel>
              <AlertDialogAction onClick={handleResetPasswordConfirm} disabled={resetPasswordValue.length < 8}>确认</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}

      {resetPasswordConfirm && (
        <AlertDialog open onOpenChange={(open) => { if (!open) setResetPasswordConfirm(null) }}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>重置密码 — {resetPasswordConfirm.label}</AlertDialogTitle>
              <AlertDialogDescription>
                确定要将该用户的密码重置为输入的新密码吗？
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={() => setResetPasswordConfirm(null)}>取消</AlertDialogCancel>
              <AlertDialogAction onClick={handleResetPassword}>确认重置</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
      <Toaster />
    </div>
  )
}
