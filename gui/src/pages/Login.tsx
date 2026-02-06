import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { useLogin } from '@/hooks/useAuth'
import { LogIn } from 'lucide-react'

export default function Login() {
  const [password, setPassword] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [error, setError] = useState('')
  const login = useLogin()

  const handleLogin = async () => {
    setError('')

    if (!password) {
      setError('비밀번호를 입력하세요')
      return
    }
    if (totpCode.length !== 6 || !/^\d{6}$/.test(totpCode)) {
      setError('6자리 TOTP 코드를 입력하세요')
      return
    }

    login.mutate(
      { password, totpCode },
      {
        onError: (e: Error) => setError(e.message || '로그인에 실패했습니다'),
      }
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Claribot</CardTitle>
          <CardDescription>로그인하여 계속하세요</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex justify-center">
              <LogIn className="h-12 w-12 text-muted-foreground" />
            </div>
            <div className="space-y-2">
              <Input
                type="password"
                placeholder="비밀번호"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && document.getElementById('totp-input')?.focus()}
              />
              <Input
                id="totp-input"
                type="text"
                inputMode="numeric"
                maxLength={6}
                placeholder="6자리 TOTP 코드"
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                onKeyDown={(e) => e.key === 'Enter' && handleLogin()}
                className="text-center text-lg tracking-widest"
              />
            </div>
            {error && <p className="text-sm text-destructive text-center">{error}</p>}
            <Button className="w-full" onClick={handleLogin} disabled={login.isPending}>
              {login.isPending ? '로그인 중...' : '로그인'}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
