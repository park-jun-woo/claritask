import { useState, useMemo, useCallback, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useFiles, useFileContent } from '@/hooks/useClaribot'
import { fileAPI } from '@/api/client'
import type { FileEntry } from '@/types'
import {
  ChevronRight, ChevronDown, File, Folder, FolderOpen,
  FileCode, FileText, FileJson, Loader2, AlertCircle, X,
} from 'lucide-react'
import { MarkdownRenderer } from '@/components/MarkdownRenderer'
import { Light as SyntaxHighlighter } from 'react-syntax-highlighter'
import { atomOneLight } from 'react-syntax-highlighter/dist/esm/styles/hljs'

// Register only needed languages
import typescript from 'react-syntax-highlighter/dist/esm/languages/hljs/typescript'
import javascript from 'react-syntax-highlighter/dist/esm/languages/hljs/javascript'
import go from 'react-syntax-highlighter/dist/esm/languages/hljs/go'
import python from 'react-syntax-highlighter/dist/esm/languages/hljs/python'
import json from 'react-syntax-highlighter/dist/esm/languages/hljs/json'
import yaml from 'react-syntax-highlighter/dist/esm/languages/hljs/yaml'
import bash from 'react-syntax-highlighter/dist/esm/languages/hljs/bash'
import css from 'react-syntax-highlighter/dist/esm/languages/hljs/css'
import xml from 'react-syntax-highlighter/dist/esm/languages/hljs/xml'
import sql from 'react-syntax-highlighter/dist/esm/languages/hljs/sql'
import dockerfile from 'react-syntax-highlighter/dist/esm/languages/hljs/dockerfile'
import markdown from 'react-syntax-highlighter/dist/esm/languages/hljs/markdown'

SyntaxHighlighter.registerLanguage('typescript', typescript)
SyntaxHighlighter.registerLanguage('javascript', javascript)
SyntaxHighlighter.registerLanguage('go', go)
SyntaxHighlighter.registerLanguage('python', python)
SyntaxHighlighter.registerLanguage('json', json)
SyntaxHighlighter.registerLanguage('yaml', yaml)
SyntaxHighlighter.registerLanguage('bash', bash)
SyntaxHighlighter.registerLanguage('css', css)
SyntaxHighlighter.registerLanguage('xml', xml)
SyntaxHighlighter.registerLanguage('sql', sql)
SyntaxHighlighter.registerLanguage('dockerfile', dockerfile)
SyntaxHighlighter.registerLanguage('markdown', markdown)

function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(() =>
    typeof window !== 'undefined' ? window.matchMedia(query).matches : false
  )
  useEffect(() => {
    const mql = window.matchMedia(query)
    const handler = (e: MediaQueryListEvent) => setMatches(e.matches)
    mql.addEventListener('change', handler)
    return () => mql.removeEventListener('change', handler)
  }, [query])
  return matches
}

// Extension to language mapping
const extToLanguage: Record<string, string> = {
  '.ts': 'typescript', '.tsx': 'typescript',
  '.js': 'javascript', '.jsx': 'javascript',
  '.go': 'go',
  '.py': 'python',
  '.json': 'json',
  '.yaml': 'yaml', '.yml': 'yaml',
  '.sh': 'bash', '.bash': 'bash',
  '.css': 'css', '.scss': 'css',
  '.html': 'xml', '.htm': 'xml', '.xml': 'xml', '.svg': 'xml',
  '.sql': 'sql',
  '.md': 'markdown',
  '.dockerfile': 'dockerfile',
}

function getLanguage(filename: string): string | undefined {
  const ext = '.' + filename.split('.').pop()?.toLowerCase()
  if (filename.toLowerCase() === 'dockerfile') return 'dockerfile'
  if (filename.toLowerCase() === 'makefile') return 'bash'
  return extToLanguage[ext]
}

function isMarkdown(filename: string): boolean {
  return filename.toLowerCase().endsWith('.md')
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(i > 0 ? 1 : 0)} ${units[i]}`
}

function getFileIcon(entry: FileEntry) {
  if (entry.type === 'dir') return null
  const ext = entry.ext || ''
  if (['.ts', '.tsx', '.js', '.jsx', '.go', '.py', '.sh'].includes(ext)) {
    return <FileCode className="h-4 w-4 text-blue-500 shrink-0" />
  }
  if (['.json', '.yaml', '.yml'].includes(ext)) {
    return <FileJson className="h-4 w-4 text-yellow-600 shrink-0" />
  }
  if (['.md', '.txt', '.log'].includes(ext)) {
    return <FileText className="h-4 w-4 text-gray-500 shrink-0" />
  }
  return <File className="h-4 w-4 text-gray-400 shrink-0" />
}

// Build path for a child entry
function buildPath(parentPath: string, name: string): string {
  if (parentPath === '.' || parentPath === '') return name
  return parentPath + '/' + name
}

// --- Tree Node Component ---

interface TreeEntry extends FileEntry {
  fullPath: string
  children?: TreeEntry[]
}

function TreeNode({
  entry,
  depth,
  expandedPaths,
  onToggle,
  onSelect,
  selectedPath,
  onLoadChildren,
}: {
  entry: TreeEntry
  depth: number
  expandedPaths: Set<string>
  onToggle: (path: string) => void
  onSelect: (path: string) => void
  selectedPath: string | null
  onLoadChildren: (path: string) => void
}) {
  const isDir = entry.type === 'dir'
  const isExpanded = expandedPaths.has(entry.fullPath)
  const isSelected = selectedPath === entry.fullPath

  const handleClick = () => {
    if (isDir) {
      onToggle(entry.fullPath)
      if (!isExpanded) {
        onLoadChildren(entry.fullPath)
      }
    } else {
      onSelect(entry.fullPath)
    }
  }

  return (
    <>
      <button
        onClick={handleClick}
        className={cn(
          "flex items-center gap-1 w-full text-left text-sm py-1.5 px-2 rounded hover:bg-accent transition-colors min-h-[32px]",
          isSelected && "bg-accent text-accent-foreground font-medium"
        )}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
      >
        {isDir ? (
          <>
            {isExpanded ? (
              <ChevronDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            ) : (
              <ChevronRight className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            )}
            {isExpanded ? (
              <FolderOpen className="h-4 w-4 text-amber-500 shrink-0" />
            ) : (
              <Folder className="h-4 w-4 text-amber-500 shrink-0" />
            )}
          </>
        ) : (
          <>
            <span className="w-3.5 shrink-0" />
            {getFileIcon(entry)}
          </>
        )}
        <span className="truncate">{entry.name}</span>
      </button>
      {isDir && isExpanded && entry.children && (
        <>
          {entry.children
            .sort((a, b) => {
              if ((a.type === 'dir') !== (b.type === 'dir')) return a.type === 'dir' ? -1 : 1
              return a.name.localeCompare(b.name)
            })
            .map((child) => (
              <TreeNode
                key={child.fullPath}
                entry={child}
                depth={depth + 1}
                expandedPaths={expandedPaths}
                onToggle={onToggle}
                onSelect={onSelect}
                selectedPath={selectedPath}
                onLoadChildren={onLoadChildren}
              />
            ))}
        </>
      )}
    </>
  )
}

// --- File Content Viewer ---

function FileContentViewer({ path }: { path: string }) {
  const { data, isLoading, error } = useFileContent(path)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-full">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full gap-2 text-destructive">
        <AlertCircle className="h-5 w-5" />
        <span className="text-sm">Failed to load file</span>
      </div>
    )
  }

  const fileData = data?.data as { path: string; content: string; size: number; ext: string; binary: boolean } | undefined
  if (!fileData) return null

  const fileName = path.split('/').pop() || path
  const pathParts = path.split('/')

  return (
    <div className="flex flex-col h-full">
      {/* File info header */}
      <div className="border-b px-4 py-3 space-y-1 shrink-0">
        <div className="flex items-center gap-2 flex-wrap">
          <h3 className="font-semibold text-sm">{fileName}</h3>
          <Badge variant="secondary" className="text-xs h-5">
            {formatFileSize(fileData.size)}
          </Badge>
        </div>
        {/* Breadcrumb path */}
        <div className="flex items-center gap-1 text-xs text-muted-foreground flex-wrap">
          {pathParts.map((part, i) => (
            <span key={i} className="flex items-center gap-1">
              {i > 0 && <span>/</span>}
              <span className={i === pathParts.length - 1 ? 'text-foreground font-medium' : ''}>
                {part}
              </span>
            </span>
          ))}
        </div>
      </div>

      {/* File content */}
      <ScrollArea className="flex-1">
        {fileData.binary ? (
          <div className="flex items-center justify-center h-48 text-muted-foreground">
            <div className="text-center space-y-2">
              <AlertCircle className="h-8 w-8 mx-auto" />
              <p className="text-sm">Binary file - cannot display</p>
            </div>
          </div>
        ) : isMarkdown(fileName) ? (
          <div className="p-4">
            <MarkdownRenderer content={fileData.content} />
          </div>
        ) : getLanguage(fileName) ? (
          <SyntaxHighlighter
            language={getLanguage(fileName)}
            style={atomOneLight}
            showLineNumbers
            lineNumberStyle={{ minWidth: '3em', paddingRight: '1em', color: '#999', userSelect: 'none' }}
            customStyle={{ margin: 0, padding: '1rem', fontSize: '13px', background: 'transparent' }}
          >
            {fileData.content}
          </SyntaxHighlighter>
        ) : (
          <pre className="p-4 text-sm font-mono whitespace-pre-wrap break-all">
            {fileData.content}
          </pre>
        )}
      </ScrollArea>
    </div>
  )
}

// --- Main Files Page ---

// Convert API items to TreeEntry with fullPath
function toTreeEntries(items: FileEntry[], parentPath: string): TreeEntry[] {
  return items.map(({ children: _, ...item }) => ({
    ...item,
    fullPath: buildPath(parentPath, item.name),
  }))
}

export default function Files() {
  const isDesktop = useMediaQuery('(min-width: 768px)')
  const [selectedPath, setSelectedPath] = useState<string | null>(null)
  const [expandedPaths, setExpandedPaths] = useState<Set<string>>(new Set())
  const [loadedPaths, setLoadedPaths] = useState<Set<string>>(new Set())

  // Load root directory
  const { data: rootData, isLoading: rootLoading } = useFiles()

  // Dynamically load subdirectories
  const [childrenMap, setChildrenMap] = useState<Record<string, TreeEntry[]>>({})

  const rootEntries: TreeEntry[] = useMemo(() => {
    const raw = rootData?.data as { items?: FileEntry[]; path?: string } | undefined
    const items = raw?.items
    if (!items || !Array.isArray(items)) return []
    return toTreeEntries(items, '.').sort((a, b) => {
      if ((a.type === 'dir') !== (b.type === 'dir')) return a.type === 'dir' ? -1 : 1
      return a.name.localeCompare(b.name)
    })
  }, [rootData])

  // Merge loaded children into entries recursively
  const mergedEntries = useMemo(() => {
    function mergeChildren(entries: TreeEntry[]): TreeEntry[] {
      return entries.map(entry => {
        if (entry.type !== 'dir') return entry
        const loaded = childrenMap[entry.fullPath]
        if (loaded) {
          return { ...entry, children: mergeChildren(loaded) }
        }
        if (entry.children) {
          return { ...entry, children: mergeChildren(entry.children) }
        }
        return entry
      })
    }
    return mergeChildren(rootEntries)
  }, [rootEntries, childrenMap])

  const handleToggle = useCallback((path: string) => {
    setExpandedPaths(prev => {
      const next = new Set(prev)
      if (next.has(path)) {
        next.delete(path)
      } else {
        next.add(path)
      }
      return next
    })
  }, [])

  const handleLoadChildren = useCallback(async (dirPath: string) => {
    if (loadedPaths.has(dirPath)) return
    setLoadedPaths(prev => new Set(prev).add(dirPath))
    try {
      const res = await fileAPI.list(dirPath)
      const raw = res?.data as { items?: FileEntry[] } | undefined
      const items = raw?.items
      if (items && Array.isArray(items)) {
        setChildrenMap(prev => ({ ...prev, [dirPath]: toTreeEntries(items, dirPath) }))
      }
    } catch {
      // Silently fail
    }
  }, [loadedPaths])

  const handleSelectFile = useCallback((path: string) => {
    setSelectedPath(path)
    if (!isDesktop) {
      window.history.pushState({ fileView: true }, '')
    }
  }, [isDesktop])

  // Mobile back button: close file viewer instead of leaving page
  useEffect(() => {
    const handlePopState = (e: PopStateEvent) => {
      if (selectedPath && !isDesktop) {
        e.preventDefault()
        setSelectedPath(null)
      }
    }
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [selectedPath, isDesktop])

  return (
    <div className="flex flex-col flex-1 min-h-0 h-full overflow-hidden">
      <div className="px-3 py-2 sm:px-4 sm:py-3 md:px-6 md:py-4 shrink-0">
        <h1 className="text-2xl md:text-3xl font-bold">Files</h1>
      </div>

      <div className={cn(
        "flex-1 flex overflow-hidden px-3 pb-3 sm:px-4 sm:pb-4 md:px-6 md:pb-6",
        isDesktop ? "gap-3" : "gap-0"
      )}>
        {/* Left: File Tree */}
        <Card className={cn(
          "shrink-0 flex flex-col overflow-hidden",
          isDesktop ? "w-[300px]" : "w-full flex-1"
        )}>
          <CardContent className="p-0 flex-1 overflow-hidden">
            <ScrollArea className="h-full">
              <div className="py-2">
                {rootLoading ? (
                  <div className="flex items-center justify-center py-8">
                    <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
                  </div>
                ) : mergedEntries.length === 0 ? (
                  <div className="text-center py-8 text-muted-foreground text-sm">
                    No files found
                  </div>
                ) : (
                  mergedEntries.map((entry) => (
                    <TreeNode
                      key={entry.fullPath}
                      entry={entry}
                      depth={0}
                      expandedPaths={expandedPaths}
                      onToggle={handleToggle}
                      onSelect={handleSelectFile}
                      selectedPath={selectedPath}
                      onLoadChildren={handleLoadChildren}
                    />
                  ))
                )}
              </div>
            </ScrollArea>
          </CardContent>
        </Card>

        {/* Desktop: File Content */}
        {isDesktop && (
          <Card className="flex-1 flex flex-col overflow-hidden">
            <CardContent className="p-0 flex-1 overflow-hidden">
              {selectedPath ? (
                <FileContentViewer path={selectedPath} />
              ) : (
                <div className="flex items-center justify-center h-full text-muted-foreground">
                  <div className="text-center space-y-2">
                    <FileText className="h-12 w-12 mx-auto opacity-30" />
                    <p className="text-sm">Select a file to view its contents</p>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        )}
      </div>

      {/* Mobile: Full-screen overlay */}
      {!isDesktop && selectedPath && (
        <div className="fixed inset-0 z-50 bg-background flex flex-col">
          <div className="flex items-center justify-between p-4 border-b">
            <h3 className="font-semibold text-sm truncate">
              {selectedPath.split('/').pop() || selectedPath}
            </h3>
            <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0" onClick={() => { window.history.back() }}>
              <X className="h-4 w-4" />
            </Button>
          </div>
          <div className="flex-1 overflow-hidden">
            <FileContentViewer path={selectedPath} />
          </div>
        </div>
      )}
    </div>
  )
}
