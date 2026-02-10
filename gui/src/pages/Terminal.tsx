import { useEffect, useRef, useState, useCallback, memo } from 'react'
import { useParams } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { getSession, createSession, ensureConnected, isConnected, sendInput } from '@/lib/terminal-store'
import '@xterm/xterm/css/xterm.css'

const CONTROL_BUTTONS = [
  // Row 1: arrows + common responses
  [
    { label: '\u2191', data: '\x1b[A', className: 'col-start-2' },
    { label: 'y', data: 'y', className: 'col-start-4' },
    { label: 'n', data: 'n' },
    { label: 'Ctrl+C', data: '\x03' },
  ],
  // Row 2: arrows + enter/tab
  [
    { label: '\u2190', data: '\x1b[D' },
    { label: '\u2193', data: '\x1b[B' },
    { label: '\u2192', data: '\x1b[C' },
    { label: 'Enter', data: '\r' },
    { label: 'Tab', data: '\t' },
  ],
] as const

const Terminal = memo(function Terminal() {
  const wrapperRef = useRef<HTMLDivElement>(null)
  const [connected, setConnected] = useState(false)
  const { projectId } = useParams<{ projectId?: string }>()
  const sessionKey = projectId || '__global__'

  const handleControlButton = useCallback((data: string) => {
    sendInput(sessionKey, data)
  }, [sessionKey])

  useEffect(() => {
    const wrapper = wrapperRef.current
    if (!wrapper) return

    // Reuse existing session or create new one
    let session = getSession(sessionKey)
    if (!session) {
      session = createSession(sessionKey, projectId)
    } else {
      // Ensure WS is connected for existing session
      ensureConnected(sessionKey)
    }

    const { containerEl, fitAddon, term } = session

    // Attach container to DOM
    wrapper.appendChild(containerEl)

    // Debounced fit to prevent flicker loops
    let fitTimer: ReturnType<typeof setTimeout> | null = null
    const debouncedFit = () => {
      if (fitTimer) clearTimeout(fitTimer)
      fitTimer = setTimeout(() => fitAddon.fit(), 100)
    }

    // Initial fit after DOM attach
    requestAnimationFrame(() => fitAddon.fit())

    // Poll connection state every second
    setConnected(isConnected(sessionKey))
    const connPoller = setInterval(() => {
      setConnected(isConnected(sessionKey))
    }, 1000)

    // Resize: debounced
    window.addEventListener('resize', debouncedFit)
    const resizeObserver = new ResizeObserver(debouncedFit)
    resizeObserver.observe(wrapper)

    term.focus()

    return () => {
      if (fitTimer) clearTimeout(fitTimer)
      clearInterval(connPoller)
      window.removeEventListener('resize', debouncedFit)
      resizeObserver.disconnect()
      // Detach from DOM but keep session alive
      if (containerEl.parentElement) {
        containerEl.parentElement.removeChild(containerEl)
      }
    }
  }, [sessionKey]) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2 border-b shrink-0">
        <div className="flex items-center gap-2">
          <h1 className="text-lg font-semibold">Terminal</h1>
          {projectId && (
            <Badge variant="outline" className="text-xs font-mono">
              {projectId}
            </Badge>
          )}
        </div>
        <Badge variant={connected ? 'outline' : 'secondary'}>
          {connected ? 'Connected' : 'Reconnecting...'}
        </Badge>
      </div>
      <div
        ref={wrapperRef}
        className="flex-1 min-h-0"
        style={{ backgroundColor: '#1a1b26', padding: 4 }}
      />
      {/* Control Pad */}
      <div className="border-t bg-muted/50 px-3 py-2 shrink-0">
        {CONTROL_BUTTONS.map((row, ri) => (
          <div key={ri} className="grid grid-cols-6 gap-1.5 mb-1.5 last:mb-0">
            {row.map((btn) => (
              <Button
                key={btn.label}
                variant="outline"
                size="sm"
                className={`h-10 text-sm font-mono active:scale-95 transition-transform ${'className' in btn ? btn.className : ''}`}
                onPointerDown={(e) => {
                  e.preventDefault()
                  handleControlButton(btn.data)
                }}
              >
                {btn.label}
              </Button>
            ))}
          </div>
        ))}
      </div>
    </div>
  )
})

export default Terminal
