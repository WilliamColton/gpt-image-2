import { describe, expect, it } from 'vitest'
import loginModalSource from './LoginModal.tsx?raw'
import registerModalSource from './RegisterModal.tsx?raw'

describe('Task 7 — LoginModal Tab switching + RegisterModal creation', () => {

  // ===== LoginModal Tabs =====

  it('LoginModal imports Tabs, TabsList, TabsTrigger, and TabsContent', () => {
    expect(loginModalSource).toContain("import { Tabs, TabsList, TabsTrigger, TabsContent } from './ui/tabs'")
  })

  it('LoginModal imports Input, Label, and Button from ui', () => {
    expect(loginModalSource).toContain("import { Input } from './ui/input'")
    expect(loginModalSource).toContain("import { Label } from './ui/label'")
    expect(loginModalSource).toContain("import { Button } from './ui/button'")
  })

  it('LoginModal imports loginWithPassword from backendApi', () => {
    expect(loginModalSource).toMatch(/import \{.*loginWithPassword.*\} from ['"]\.\.\/lib\/backendApi['"]/)
  })

  it('LoginModal renders Tabs with defaultValue="code"', () => {
    expect(loginModalSource).toMatch(/defaultValue=["']code["']/)
  })

  it('LoginModal has TabsTrigger for 兑换码 and 密码登录', () => {
    expect(loginModalSource).toContain('value="code"')
    const textContent = loginModalSource.replace(/className="[^"]*"/g, '')
    expect(textContent).toContain('兑换码')
    expect(loginModalSource).toContain('value="password"')
    expect(textContent).toContain('密码登录')
  })

  it('LoginModal TabsContent value="code" keeps existing code login (loginWithCode call)', () => {
    expect(loginModalSource).toContain('loginWithCode')
    expect(loginModalSource).toContain('<TabsContent value="code"')
  })

  it('LoginModal TabsContent value="password" has username and password Input fields', () => {
    expect(loginModalSource).toContain('<TabsContent value="password"')
    // Should contain a password type input
    expect(loginModalSource).toMatch(/type=["']password["']/)
  })

  it('LoginModal password tab calls loginWithPassword on submit', () => {
    expect(loginModalSource).toContain('loginWithPassword')
  })

  it('LoginModal password tab button shows disabled state based on username and pass', () => {
    // Button disabled when loading or username/pass empty
    expect(loginModalSource).toContain('username')
    expect(loginModalSource).toContain('pass')
  })

  // ===== Register link =====

  it('LoginModal has 没有账号？立即注册 link at bottom', () => {
    const textContent = loginModalSource.replace(/className="[^"]*"/g, '').replace(/\s+/g, ' ')
    expect(textContent).toContain('没有账号？立即注册')
  })

  it('LoginModal 没有账号？立即注册 link sets showRegister state', () => {
    expect(loginModalSource).toMatch(/setShowRegister/)
  })

  // ===== RegisterModal =====

  it('RegisterModal.tsx file exists and is a React component', () => {
    expect(registerModalSource).toContain('export default function RegisterModal')
  })

  it('RegisterModal imports register, bootstrapBackendSession, useStore', () => {
    expect(registerModalSource).toContain("import { register } from '../lib/backendApi'")
    expect(registerModalSource).toContain("import { bootstrapBackendSession, useStore } from '../store'")
  })

  it('RegisterModal imports Input, Label, Button from ui', () => {
    expect(registerModalSource).toContain("import { Input } from './ui/input'")
    expect(registerModalSource).toContain("import { Label } from './ui/label'")
    expect(registerModalSource).toContain("import { Button } from './ui/button'")
  })

  it('RegisterModal has three fields: inviteCode, username, password', () => {
    expect(registerModalSource).toContain('inviteCode')
    expect(registerModalSource).toContain('username')
    expect(registerModalSource).toContain('password')
  })

  it('RegisterModal invite code field has correct placeholder from UI-SPEC', () => {
    const textContent = registerModalSource.replace(/className="[^"]*"/g, '')
    expect(textContent).toContain('选填，输入邀请码可获得额外配额')
  })

  it('RegisterModal username field has correct placeholder from UI-SPEC', () => {
    const textContent = registerModalSource.replace(/className="[^"]*"/g, '')
    expect(textContent).toContain('3-20 字符，允许中文')
  })

  it('RegisterModal password field has correct placeholder from UI-SPEC', () => {
    const textContent = registerModalSource.replace(/className="[^"]*"/g, '')
    expect(textContent).toContain('至少 8 字符')
  })

  it('RegisterModal calls register() on submit then bootstrapBackendSession', () => {
    expect(registerModalSource).toContain('register(')
    expect(registerModalSource).toContain('bootstrapBackendSession')
  })

  it('RegisterModal has frontend validation for username 3-20 characters', () => {
    expect(registerModalSource).toMatch(/Array\.from/)
  })

  it('RegisterModal uses Dialog component from shadcn', () => {
    expect(registerModalSource).toContain('Dialog')
  })

  it('RegisterModal button shows 立即注册 text', () => {
    const textContent = registerModalSource.replace(/className="[^"]*"/g, '').replace(/\s+/g, ' ')
    expect(textContent).toContain('>立即注册<')
  })
})
