import { useEffect } from 'react'
import { getChangelogDismissKey, initStore, useStore } from './store'
import Header from './components/Header'
import SearchBar from './components/SearchBar'
import TaskGrid from './components/TaskGrid'
import InputBar from './components/InputBar'
import DetailModal from './components/DetailModal'
import Lightbox from './components/Lightbox'
import SettingsModal from './components/SettingsModal'
import ConfirmDialog from './components/ConfirmDialog'
import Toast from './components/Toast'
import MaskEditorModal from './components/MaskEditorModal'
import LoginModal from './components/LoginModal'
import AnnouncementModal from './components/AnnouncementModal'
import ChangelogModal from './components/ChangelogModal'
import MigrationModal from './components/MigrationModal'

export default function App() {
  const authUser = useStore((s) => s.authUser)
  const latestChangelog = useStore((s) => s.latestChangelog)
  const dismissedChangelogKeys = useStore((s) => s.dismissedChangelogKeys)
  const showChangelog = useStore((s) => s.showChangelog)
  const setShowChangelog = useStore((s) => s.setShowChangelog)

  useEffect(() => {
    initStore()
  }, [])

  useEffect(() => {
    if (!latestChangelog?.published || !latestChangelog.version.trim()) return
    const key = getChangelogDismissKey(latestChangelog)
    if (!dismissedChangelogKeys.includes(key)) {
      setShowChangelog(true, key)
    }
  }, [latestChangelog, dismissedChangelogKeys, setShowChangelog])

  useEffect(() => {
    const preventPageImageDrag = (e: DragEvent) => {
      if ((e.target as HTMLElement | null)?.closest('img')) {
        e.preventDefault()
      }
    }

    document.addEventListener('dragstart', preventPageImageDrag)
    return () => document.removeEventListener('dragstart', preventPageImageDrag)
  }, [])

  return (
    <>
      <Header />
      <main data-home-main className="safe-area-x max-w-7xl mx-auto pb-48">
        <SearchBar />
        <TaskGrid />
      </main>
      <InputBar />
      <DetailModal />
      <Lightbox />
      <SettingsModal />
      <ConfirmDialog />
      <Toast />
      <AnnouncementModal mode="auto" />
      {showChangelog && <ChangelogModal onClose={() => setShowChangelog(false)} />}
      <MaskEditorModal />
      {!authUser && <LoginModal />}
      {authUser?.needsMigration && <MigrationModal />}
    </>
  )
}
