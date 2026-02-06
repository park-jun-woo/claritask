import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { authAPI } from '@/api/client'

export function useAuthStatus() {
  return useQuery({
    queryKey: ['auth'],
    queryFn: authAPI.status,
    refetchInterval: 30_000,
    retry: 1,
  })
}

export function useLogin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { password: string; totpCode: string }) =>
      authAPI.login(params.password, params.totpCode),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['auth'] }),
  })
}

export function useLogout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => authAPI.logout(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['auth'] }),
  })
}

export function useSetup() {
  return useMutation({
    mutationFn: (params: { password: string }) =>
      authAPI.setup(params.password),
  })
}

export function useSetupVerify() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { password: string; totpCode: string }) =>
      authAPI.setupVerify(params.password, params.totpCode),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['auth'] }),
  })
}
