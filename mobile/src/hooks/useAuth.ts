import { useState, useEffect, useCallback, useRef } from 'react'
import { authAPI, setServerUrl, setOnAuthError, setAuthToken } from '../api/client'
import { saveToken, getToken, deleteToken, saveServerUrl, getServerUrl, deleteServerUrl } from '../storage/keychain'
import { setCachedServerUrl, getCachedServerUrl, clearCachedServerUrl } from '../storage/cache'

export type AuthState = 'loading' | 'no_server' | 'needs_setup' | 'needs_login' | 'authenticated'

interface UseAuthReturn {
  state: AuthState
  serverUrl: string | null
  isLoading: boolean
  error: string | null
  connectServer: (url: string) => Promise<boolean>
  login: (password: string, totpCode: string) => Promise<void>
  logout: () => Promise<void>
  disconnectServer: () => Promise<void>
  checkAuth: () => Promise<void>
}

export function useAuth(): UseAuthReturn {
  const [state, setState] = useState<AuthState>('loading')
  const [serverUrl, setServerUrlState] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const refreshTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Set up auth error callback to redirect to login
  useEffect(() => {
    setOnAuthError(() => {
      setState('needs_login')
    })
  }, [])

  // Initialize: check saved server URL and auth state
  useEffect(() => {
    initialize()
    return () => {
      if (refreshTimerRef.current) {
        clearTimeout(refreshTimerRef.current)
      }
    }
  }, [])

  const initialize = useCallback(async () => {
    setIsLoading(true)
    try {
      // Try to restore server URL from keychain
      const savedUrl = await getServerUrl()
      if (!savedUrl) {
        // Check cache as fallback
        const cachedUrl = getCachedServerUrl()
        if (!cachedUrl) {
          setState('no_server')
          setIsLoading(false)
          return
        }
        setServerUrl(cachedUrl)
        setServerUrlState(cachedUrl)
      } else {
        setServerUrl(savedUrl)
        setServerUrlState(savedUrl)
      }

      // Check if we have a stored token
      const token = await getToken()
      if (token) {
        setAuthToken(token)
        // Validate token by calling auth status
        try {
          const status = await authAPI.status()
          if (status.is_authenticated) {
            setState('authenticated')
            scheduleTokenRefresh()
          } else if (!status.setup_completed) {
            setState('needs_setup')
          } else {
            // Token expired/invalid
            await deleteToken()
            setState('needs_login')
          }
        } catch {
          // Server unreachable or token invalid
          await deleteToken()
          setState('needs_login')
        }
      } else {
        // No token - check if setup is needed
        try {
          const status = await authAPI.status()
          if (!status.setup_completed) {
            setState('needs_setup')
          } else {
            setState('needs_login')
          }
        } catch {
          setState('needs_login')
        }
      }
    } catch {
      setState('no_server')
    } finally {
      setIsLoading(false)
    }
  }, [])

  // Connect to a server URL and verify it's reachable
  const connectServer = useCallback(async (url: string): Promise<boolean> => {
    setError(null)
    setIsLoading(true)
    try {
      const normalizedUrl = url.replace(/\/+$/, '')
      setServerUrl(normalizedUrl)

      // Verify server is reachable via health endpoint
      const res = await fetch(`${normalizedUrl}/api/health`)
      if (!res.ok) {
        throw new Error('Server not reachable')
      }

      // Save server URL
      await saveServerUrl(normalizedUrl)
      setCachedServerUrl(normalizedUrl)
      setServerUrlState(normalizedUrl)

      // Check auth status
      const status = await authAPI.status()
      if (!status.setup_completed) {
        setState('needs_setup')
      } else {
        setState('needs_login')
      }
      return true
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to connect to server')
      setState('no_server')
      return false
    } finally {
      setIsLoading(false)
    }
  }, [])

  // Login with password + TOTP
  const login = useCallback(async (password: string, totpCode: string) => {
    setError(null)
    setIsLoading(true)
    try {
      const result = await authAPI.login(password, totpCode)
      if (result.token) {
        await saveToken(result.token)
        setAuthToken(result.token)
        setState('authenticated')
        scheduleTokenRefresh()
      } else {
        throw new Error('No token received from server')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
      throw err
    } finally {
      setIsLoading(false)
    }
  }, [])

  // Logout
  const logout = useCallback(async () => {
    try {
      await authAPI.logout()
    } catch {
      // Ignore logout errors - we clear local state regardless
    }
    await deleteToken()
    if (refreshTimerRef.current) {
      clearTimeout(refreshTimerRef.current)
      refreshTimerRef.current = null
    }
    setState('needs_login')
  }, [])

  // Disconnect from server (logout + clear server URL)
  const disconnectServer = useCallback(async () => {
    try {
      await authAPI.logout()
    } catch {
      // Ignore errors
    }
    await deleteToken()
    await deleteServerUrl()
    clearCachedServerUrl()
    setServerUrl('')
    setServerUrlState(null)
    if (refreshTimerRef.current) {
      clearTimeout(refreshTimerRef.current)
      refreshTimerRef.current = null
    }
    setState('no_server')
  }, [])

  // Check auth status (manual refresh)
  const checkAuth = useCallback(async () => {
    setIsLoading(true)
    try {
      const token = await getToken()
      if (!token) {
        setState('needs_login')
        return
      }
      const status = await authAPI.status()
      if (status.is_authenticated) {
        setState('authenticated')
      } else {
        await deleteToken()
        setState('needs_login')
      }
    } catch {
      await deleteToken()
      setState('needs_login')
    } finally {
      setIsLoading(false)
    }
  }, [])

  // Schedule periodic token validity check (every 5 minutes)
  const scheduleTokenRefresh = useCallback(() => {
    if (refreshTimerRef.current) {
      clearTimeout(refreshTimerRef.current)
    }
    refreshTimerRef.current = setTimeout(async () => {
      try {
        const token = await getToken()
        if (!token) {
          setState('needs_login')
          return
        }
        const status = await authAPI.status()
        if (!status.is_authenticated) {
          await deleteToken()
          setState('needs_login')
        } else {
          scheduleTokenRefresh() // reschedule
        }
      } catch {
        // Network error - keep current state, retry later
        scheduleTokenRefresh()
      }
    }, 5 * 60 * 1000) // 5 minutes
  }, [])

  return {
    state,
    serverUrl,
    isLoading,
    error,
    connectServer,
    login,
    logout,
    disconnectServer,
    checkAuth,
  }
}
