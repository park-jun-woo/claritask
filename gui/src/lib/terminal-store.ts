import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'

export interface TerminalSession {
  term: Terminal
  ws: WebSocket | null
  fitAddon: FitAddon
  containerEl: HTMLDivElement
  projectId?: string
  onDataDisposable: { dispose(): void } | null
  onResizeDisposable: { dispose(): void } | null
  reconnectTimer: ReturnType<typeof setTimeout> | null
  destroyed: boolean
}

// Module-level store: sessions persist across React component mounts/unmounts
const sessions = new Map<string, TerminalSession>()

export function getSession(key: string): TerminalSession | undefined {
  return sessions.get(key)
}

export function createSession(key: string, projectId?: string): TerminalSession {
  // Clean up existing session for same key if any
  destroySession(key)

  const containerEl = document.createElement('div')
  containerEl.style.height = '100%'
  containerEl.style.width = '100%'

  const term = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    theme: {
      background: '#1a1b26',
      foreground: '#c0caf5',
      cursor: '#c0caf5',
      selectionBackground: '#33467c',
    },
  })

  const fitAddon = new FitAddon()
  const webLinksAddon = new WebLinksAddon()
  term.loadAddon(fitAddon)
  term.loadAddon(webLinksAddon)
  term.open(containerEl)
  fitAddon.fit()

  const session: TerminalSession = {
    term,
    ws: null,
    fitAddon,
    containerEl,
    projectId,
    onDataDisposable: null,
    onResizeDisposable: null,
    reconnectTimer: null,
    destroyed: false,
  }
  sessions.set(key, session)

  // Connect WebSocket
  connectWS(session, key)

  return session
}

// connectWS creates a new WebSocket connection for an existing TerminalSession.
function connectWS(session: TerminalSession, key: string) {
  if (session.destroyed) return

  // Dispose previous listeners to prevent duplicates
  if (session.onDataDisposable) {
    session.onDataDisposable.dispose()
    session.onDataDisposable = null
  }
  if (session.onResizeDisposable) {
    session.onResizeDisposable.dispose()
    session.onResizeDisposable = null
  }

  // Detach old WS handlers before closing so onclose doesn't trigger reconnect
  const oldWs = session.ws
  if (oldWs) {
    oldWs.onclose = null
    oldWs.onerror = null
    oldWs.onmessage = null
    if (oldWs.readyState === WebSocket.OPEN || oldWs.readyState === WebSocket.CONNECTING) {
      oldWs.close()
    }
  }

  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  let wsUrl = `${protocol}//${location.host}/api/terminal/ws?cols=${session.term.cols}&rows=${session.term.rows}`
  if (session.projectId) {
    wsUrl += `&project_id=${encodeURIComponent(session.projectId)}`
  }
  const ws = new WebSocket(wsUrl)
  ws.binaryType = 'arraybuffer'
  session.ws = ws

  ws.onopen = () => {
    // Send resize to sync terminal size
    if (ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'resize', cols: session.term.cols, rows: session.term.rows }))
    }
  }

  ws.onmessage = (e) => {
    if (e.data instanceof ArrayBuffer) {
      session.term.write(new Uint8Array(e.data))
    } else if (typeof e.data === 'string') {
      // Handle text messages (control messages like "attached")
      try {
        const msg = JSON.parse(e.data)
        if (msg.type === 'pong') return
      } catch {
        // ignore
      }
    }
  }

  ws.onclose = () => {
    // Only reconnect if this is still the active WS (not replaced by a newer one)
    if (session.destroyed || session.ws !== ws) return
    session.term.write('\r\n\x1b[33m[Disconnected - reconnecting...]\x1b[0m\r\n')
    scheduleReconnect(session, key)
  }

  ws.onerror = () => {
    // onclose will fire after onerror, so reconnect is handled there
  }

  // Terminal input → WebSocket
  session.onDataDisposable = session.term.onData((data) => {
    if (session.ws && session.ws.readyState === WebSocket.OPEN) {
      session.ws.send(new TextEncoder().encode(data))
    }
  })

  // Terminal resize → WebSocket
  session.onResizeDisposable = session.term.onResize(({ cols, rows }) => {
    if (session.ws && session.ws.readyState === WebSocket.OPEN) {
      session.ws.send(JSON.stringify({ type: 'resize', cols, rows }))
    }
  })
}

// scheduleReconnect sets up auto-reconnect after a delay.
function scheduleReconnect(session: TerminalSession, key: string) {
  if (session.destroyed) return
  if (session.reconnectTimer) {
    clearTimeout(session.reconnectTimer)
  }
  session.reconnectTimer = setTimeout(() => {
    session.reconnectTimer = null
    if (!session.destroyed) {
      connectWS(session, key)
    }
  }, 3000)
}

// ensureConnected checks if the session's WS is connected, and reconnects if not.
export function ensureConnected(key: string): void {
  const session = sessions.get(key)
  if (!session || session.destroyed) return

  if (!session.ws || session.ws.readyState === WebSocket.CLOSED || session.ws.readyState === WebSocket.CLOSING) {
    // Cancel any pending reconnect and connect immediately
    if (session.reconnectTimer) {
      clearTimeout(session.reconnectTimer)
      session.reconnectTimer = null
    }
    connectWS(session, key)
  }
}

// isConnected returns whether the session's WS is currently open.
export function isConnected(key: string): boolean {
  const session = sessions.get(key)
  if (!session) return false
  return session.ws !== null && session.ws.readyState === WebSocket.OPEN
}

// sendInput sends raw key data to the PTY via WebSocket (for control pad buttons).
export function sendInput(key: string, data: string): boolean {
  const session = sessions.get(key)
  if (!session || !session.ws || session.ws.readyState !== WebSocket.OPEN) return false
  session.ws.send(new TextEncoder().encode(data))
  return true
}

// destroySession sends close to the PTY and cleans up everything.
export function destroySession(key: string) {
  const session = sessions.get(key)
  if (!session) return
  session.destroyed = true

  if (session.reconnectTimer) {
    clearTimeout(session.reconnectTimer)
    session.reconnectTimer = null
  }

  // Send explicit close to kill PTY on server
  if (session.ws && session.ws.readyState === WebSocket.OPEN) {
    try {
      session.ws.send(JSON.stringify({ type: 'close' }))
    } catch {
      // ignore
    }
  }
  if (session.ws) {
    session.ws.close()
  }

  if (session.onDataDisposable) {
    session.onDataDisposable.dispose()
  }
  if (session.onResizeDisposable) {
    session.onResizeDisposable.dispose()
  }
  session.term.dispose()
  if (session.containerEl.parentElement) {
    session.containerEl.parentElement.removeChild(session.containerEl)
  }
  sessions.delete(key)
}
