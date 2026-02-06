import { useState } from 'react'
import { QRCodeSVG } from 'qrcode.react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { useSetup, useSetupVerify } from '@/hooks/useAuth'
import { Shield, QrCode, KeyRound } from 'lucide-react'

type Step = 1 | 2 | 3

export default function Setup() {
  const [step, setStep] = useState<Step>(1)
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [totpURI, setTotpURI] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [error, setError] = useState('')
  const setup = useSetup()
  const setupVerify = useSetupVerify()

  const handlePasswordSubmit = () => {
    setError('')

    if (password.length < 4) {
      setError('비밀번호는 최소 4자 이상이어야 합니다')
      return
    }
    if (password !== confirmPassword) {
      setError('비밀번호가 일치하지 않습니다')
      return
    }

    setup.mutate(
      { password },
      {
        onSuccess: (result) => {
          if (result.totp_uri) {
            setTotpURI(result.totp_uri)
            setStep(2)
          }
        },
        onError: (e: Error) => setError(e.message || '설정 중 오류가 발생했습니다'),
      }
    )
  }

  const handleVerifySubmit = () => {
    setError('')

    if (totpCode.length !== 6 || !/^\d{6}$/.test(totpCode)) {
      setError('6자리 숫자를 입력하세요')
      return
    }

    setupVerify.mutate(
      { password, totpCode },
      {
        onError: (e: Error) => setError(e.message || 'TOTP 검증에 실패했습니다'),
      }
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Claribot Setup</CardTitle>
          <CardDescription>
            {step === 1 && '관리자 비밀번호를 설정하세요'}
            {step === 2 && '인증 앱으로 QR 코드를 스캔하세요'}
            {step === 3 && 'TOTP 코드를 입력하세요'}
          </CardDescription>
          {/* Step indicator */}
          <div className="flex items-center justify-center gap-2 pt-2">
            {[1, 2, 3].map((s) => (
              <div
                key={s}
                className={`h-2 w-8 rounded-full transition-colors ${
                  s <= step ? 'bg-primary' : 'bg-muted'
                }`}
              />
            ))}
          </div>
        </CardHeader>
        <CardContent>
          {/* Step 1: Password */}
          {step === 1 && (
            <div className="space-y-4">
              <div className="flex justify-center">
                <Shield className="h-12 w-12 text-muted-foreground" />
              </div>
              <div className="space-y-2">
                <Input
                  type="password"
                  placeholder="비밀번호"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handlePasswordSubmit()}
                />
                <Input
                  type="password"
                  placeholder="비밀번호 확인"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handlePasswordSubmit()}
                />
              </div>
              {error && <p className="text-sm text-destructive text-center">{error}</p>}
              <Button className="w-full" onClick={handlePasswordSubmit} disabled={setup.isPending}>
                {setup.isPending ? '처리 중...' : '다음'}
              </Button>
            </div>
          )}

          {/* Step 2: QR Code */}
          {step === 2 && (
            <div className="space-y-4">
              <div className="flex justify-center">
                <QrCode className="h-12 w-12 text-muted-foreground" />
              </div>
              <div className="flex justify-center p-4 bg-white rounded-lg">
                <QRCodeSVG value={totpURI} size={200} />
              </div>
              <p className="text-xs text-muted-foreground text-center">
                Google Authenticator 또는 다른 TOTP 앱으로 스캔하세요
              </p>
              <Button className="w-full" onClick={() => setStep(3)}>
                다음
              </Button>
            </div>
          )}

          {/* Step 3: TOTP Verification */}
          {step === 3 && (
            <div className="space-y-4">
              <div className="flex justify-center">
                <KeyRound className="h-12 w-12 text-muted-foreground" />
              </div>
              <Input
                type="text"
                inputMode="numeric"
                maxLength={6}
                placeholder="6자리 코드 입력"
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                onKeyDown={(e) => e.key === 'Enter' && handleVerifySubmit()}
                className="text-center text-lg tracking-widest"
              />
              {error && <p className="text-sm text-destructive text-center">{error}</p>}
              <Button className="w-full" onClick={handleVerifySubmit} disabled={setupVerify.isPending}>
                {setupVerify.isPending ? '검증 중...' : '완료'}
              </Button>
              <Button variant="ghost" className="w-full" onClick={() => setStep(2)}>
                QR 코드 다시 보기
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
