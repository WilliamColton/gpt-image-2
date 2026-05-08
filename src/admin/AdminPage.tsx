import { useState } from 'react'
import { isAdminLoggedIn } from './adminApi'
import AdminLogin from './AdminLogin'
import AdminDashboard from './AdminDashboard'

export default function AdminPage() {
  const [loggedIn, setLoggedIn] = useState(isAdminLoggedIn())

  if (!loggedIn) {
    return <AdminLogin onLogin={() => setLoggedIn(true)} />
  }

  return <AdminDashboard onLogout={() => {
    import('./adminApi').then(m => m.clearAdminToken())
    setLoggedIn(false)
  }} />
}
