import { useState } from 'react'
import { loginWithCode, loginWithPassword } from '../lib/backendApi'
import { bootstrapBackendSession, useStore } from '../store'
import { Tabs, TabsList, TabsTrigger, TabsContent } from './ui/tabs'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Button } from './ui/button'
import { Dialog, DialogContent } from './ui/dialog'
import RegisterModal from './RegisterModal'

export default function LoginModal() {
  const [code, setCode] = useState('')
  const [username, setUsername] = useState('')
  const [pass, setPass] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [showRegister, setShowRegister] = useState(false)

  const handleCodeSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
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

  const handlePasswordLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await loginWithPassword(username, pass)
      await bootstrapBackendSession()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Dialog open modal>
        <DialogContent className="max-w-md" data-no-drag-select hideClose
          onPointerDownOutside={(e) => e.preventDefault()}
          onEscapeKeyDown={(e) => e.preventDefault()}
        >
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100">登录</h2>

          <Tabs defaultValue="code">
            <TabsList className="mt-4 w-full">
              <TabsTrigger value="code" className="flex-1">
                兑换码
              </TabsTrigger>
              <TabsTrigger value="password" className="flex-1">
                密码登录
              </TabsTrigger>
            </TabsList>

            <TabsContent value="code">
              <form onSubmit={handleCodeSubmit}>
                <Input
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  type="text"
                  autoFocus
                  placeholder="请输入兑换码"
                  className="mt-4"
                />
                {error && (
                  <div className="mt-3 rounded-xl bg-red-50 px-3 py-2 text-sm text-red-500 dark:bg-red-500/10 dark:text-red-300">
                    {error}
                  </div>
                )}
                <Button
                  type="submit"
                  disabled={loading || !code.trim()}
                  className="mt-5 w-full"
                >
                  {loading ? '登录中...' : '登录 / 注册'}
                </Button>
              </form>
            </TabsContent>

            <TabsContent value="password">
              <form onSubmit={handlePasswordLogin}>
                <div className="space-y-3">
                  <div>
                    <Label>用户名</Label>
                    <Input
                      type="text"
                      placeholder="用户名"
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      className="mt-1"
                    />
                  </div>
                  <div>
                    <Label>密码</Label>
                    <Input
                      type="password"
                      placeholder="密码"
                      value={pass}
                      onChange={(e) => setPass(e.target.value)}
                      className="mt-1"
                    />
                  </div>
                </div>
                {error && (
                  <div className="mt-3 rounded-xl bg-red-50 px-3 py-2 text-sm text-red-500 dark:bg-red-500/10 dark:text-red-300">
                    {error}
                  </div>
                )}
                <Button
                  type="submit"
                  disabled={loading || !username.trim() || !pass.trim()}
                  className="mt-5 w-full"
                >
                  {loading ? '登录中...' : '登录'}
                </Button>
              </form>
            </TabsContent>
          </Tabs>

          <button
            type="button"
            onClick={() => setShowRegister(true)}
            className="mt-4 text-xs text-blue-600 hover:underline dark:text-blue-400"
          >
            没有账号？立即注册
          </button>
        </DialogContent>
      </Dialog>

      {showRegister && <RegisterModal onClose={() => setShowRegister(false)} />}
    </>
  )
}
