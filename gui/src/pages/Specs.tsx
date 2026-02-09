import { useState, useMemo, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import {
  useSpecs, useSpec, useAddSpec, useSetSpec, useDeleteSpec
} from '@/hooks/useClaribot'
import type { Spec } from '@/types'
import {
  Plus, X, FileText, Search, Eye, Pencil
} from 'lucide-react'
import { MarkdownRenderer } from '@/components/MarkdownRenderer'

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

type StatusFilter = 'all' | 'draft' | 'review' | 'approved' | 'deprecated'

export default function Specs() {
  const { projectId, specId: specIdParam } = useParams<{ projectId?: string; specId?: string }>()
  const navigate = useNavigate()

  const basePath = projectId ? `/projects/${projectId}/specs` : '/specs'

  const { data: specsData } = useSpecs()
  const addSpec = useAddSpec()
  const setSpec = useSetSpec()
  const deleteSpec = useDeleteSpec()

  const [showAdd, setShowAdd] = useState(false)
  const [addForm, setAddForm] = useState({ title: '', content: '' })

  // URL-based spec selection
  const selectedSpecId = specIdParam ? Number(specIdParam) : null
  const setSelectedSpecId = (id: number | null) => {
    if (id !== null) {
      navigate(`${basePath}/${id}`)
    } else {
      navigate(basePath)
    }
  }

  const { data: specDetail } = useSpec(selectedSpecId ?? undefined)
  const selectedSpec: Spec | null = specDetail?.data ?? null
  const [editField, setEditField] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [contentMode, setContentMode] = useState<'preview' | 'edit'>('preview')

  const specItems = useMemo(() => parseItems(specsData?.data), [specsData])

  const filteredItems = useMemo(() => {
    let items = specItems
    if (statusFilter !== 'all') {
      items = items.filter((s: Spec) => s.status === statusFilter)
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase()
      items = items.filter((s: Spec) =>
        s.title.toLowerCase().includes(q) ||
        s.content.toLowerCase().includes(q)
      )
    }
    return items.sort((a: Spec, b: Spec) => a.id - b.id)
  }, [specItems, statusFilter, searchQuery])

  const handleAdd = async () => {
    if (!addForm.title) return
    await addSpec.mutateAsync({
      title: addForm.title,
      content: addForm.content || undefined,
    })
    setAddForm({ title: '', content: '' })
    setShowAdd(false)
  }

  const handleEdit = async (specId: number, field: string, value: string) => {
    await setSpec.mutateAsync({ id: specId, field, value })
    setEditField(null)
  }

  const handleStatusChange = async (specId: number, newStatus: string) => {
    await setSpec.mutateAsync({ id: specId, field: 'status', value: newStatus })
  }

  const isDesktop = useMediaQuery('(min-width: 768px)')

  const detailContent = selectedSpec ? (
    <div className="flex flex-col flex-1 min-h-0">
      <ScrollArea className="flex-1 min-h-0 p-4">
        <div className="space-y-4">
          {/* Title */}
          <div>
            {editField === 'title' ? (
              <div className="space-y-1">
                <Input
                  value={editValue}
                  onChange={e => setEditValue(e.target.value)}
                  className="text-lg font-medium"
                  autoFocus
                />
                <div className="flex gap-1">
                  <Button size="sm" className="min-h-[44px] text-xs" onClick={() => handleEdit(selectedSpec.id, 'title', editValue)}>Save</Button>
                  <Button size="sm" variant="ghost" className="min-h-[44px] text-xs" onClick={() => setEditField(null)}>Cancel</Button>
                </div>
              </div>
            ) : (
              <div className="flex items-start justify-between gap-2">
                <h4 className="text-lg font-medium">{selectedSpec.title}</h4>
                <div className="flex gap-1 shrink-0">
                  <Button variant="ghost" size="sm" className="min-h-[44px] text-xs" onClick={() => { setEditField('title'); setEditValue(selectedSpec.title) }}>
                    <Pencil className="h-3 w-3 mr-1" /> Edit
                  </Button>
                  <Button
                    size="sm"
                    variant="destructive"
                    className="min-h-[44px] text-xs"
                    onClick={() => {
                      if (confirm('Delete this spec?')) {
                        deleteSpec.mutate(selectedSpec.id)
                        setSelectedSpecId(null)
                      }
                    }}
                  >
                    Delete
                  </Button>
                </div>
              </div>
            )}
            <div className="flex items-center gap-2 mt-2">
              <SpecStatusBadge status={selectedSpec.status} />
              <span className="text-xs text-muted-foreground">
                priority: {selectedSpec.priority}
              </span>
              <span className="text-xs text-muted-foreground">
                {new Date(selectedSpec.updated_at).toLocaleDateString()}
              </span>
            </div>
          </div>

          <Separator />

          {/* Status */}
          <div>
            <div className="flex items-center justify-between">
              <h5 className="text-sm font-medium text-muted-foreground">Status</h5>
            </div>
            <div className="flex gap-1 mt-1">
              {(['draft', 'review', 'approved', 'deprecated'] as const).map(s => (
                <Button
                  key={s}
                  size="sm"
                  variant={selectedSpec.status === s ? 'default' : 'outline'}
                  className="min-h-[44px] text-xs"
                  onClick={() => handleStatusChange(selectedSpec.id, s)}
                  disabled={setSpec.isPending}
                >
                  {s}
                </Button>
              ))}
            </div>
          </div>

          {/* Priority */}
          <div>
            <div className="flex items-center justify-between">
              <h5 className="text-sm font-medium text-muted-foreground">Priority</h5>
            </div>
            <div className="flex gap-1 mt-1">
              {([1, 2, 3, 4, 5] as const).map(p => (
                <Button
                  key={p}
                  size="sm"
                  variant={selectedSpec.priority === p ? 'default' : 'outline'}
                  className="min-h-[44px] text-xs w-10"
                  onClick={() => handleEdit(selectedSpec.id, 'priority', String(p))}
                  disabled={setSpec.isPending}
                >
                  {p}
                </Button>
              ))}
            </div>
          </div>

          <Separator />

          {/* Content */}
          <div>
            <div className="flex items-center justify-between">
              <h5 className="text-sm font-medium text-muted-foreground">Content</h5>
              <div className="flex gap-1">
                <Button
                  variant={contentMode === 'preview' ? 'secondary' : 'ghost'}
                  size="sm"
                  className="min-h-[44px] text-xs"
                  onClick={() => setContentMode('preview')}
                >
                  <Eye className="h-3 w-3 mr-1" /> Preview
                </Button>
                <Button
                  variant={contentMode === 'edit' ? 'secondary' : 'ghost'}
                  size="sm"
                  className="min-h-[44px] text-xs"
                  onClick={() => { setContentMode('edit'); setEditValue(selectedSpec.content || '') }}
                >
                  <Pencil className="h-3 w-3 mr-1" /> Edit
                </Button>
              </div>
            </div>
            {contentMode === 'edit' ? (
              <div className="mt-1 space-y-1">
                <Textarea
                  value={editValue}
                  onChange={e => setEditValue(e.target.value)}
                  rows={Math.max(10, (editValue.match(/\n/g) || []).length + 3)}
                  className="font-mono text-sm resize-none overflow-hidden"
                  placeholder="Markdown content..."
                />
                <div className="flex gap-1">
                  <Button size="sm" className="min-h-[44px] text-xs" onClick={async () => { await handleEdit(selectedSpec.id, 'content', editValue); setContentMode('preview') }} disabled={setSpec.isPending}>Save</Button>
                  <Button size="sm" variant="ghost" className="min-h-[44px] text-xs" onClick={() => setContentMode('preview')}>Cancel</Button>
                </div>
              </div>
            ) : (
              <div className="mt-1 bg-muted rounded p-3">
                <MarkdownRenderer content={selectedSpec.content || ''} />
              </div>
            )}
          </div>
        </div>
      </ScrollArea>
    </div>
  ) : null

  return (
    <div className="flex flex-col md:flex-row gap-4 flex-1 min-h-0 h-full overflow-hidden">
      {/* List Panel */}
      <div className={cn(
        "flex flex-col space-y-2 min-w-0",
        isDesktop ? "w-1/3" : "w-full flex-1"
      )}>
        {/* Toolbar */}
        <div className="flex flex-wrap items-center justify-between gap-1">
          <h1 className="text-xl font-bold shrink-0">Specs</h1>
          <Button size="icon" className="h-8 w-8" onClick={() => setShowAdd(!showAdd)}>
            <Plus className="h-4 w-4" />
          </Button>
        </div>

        {/* Search */}
        <div className="relative">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search specs..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="pl-8 h-8 text-sm"
          />
        </div>

        {/* Status Filter */}
        <div className="flex flex-wrap gap-1">
          {(['all', 'draft', 'review', 'approved', 'deprecated'] as const).map(s => (
            <Button
              key={s}
              size="sm"
              variant={statusFilter === s ? 'secondary' : 'ghost'}
              className="text-xs h-7"
              onClick={() => setStatusFilter(s)}
            >
              {s}
              {s !== 'all' && (
                <span className="ml-1 text-muted-foreground">
                  {specItems.filter((sp: Spec) => sp.status === s).length}
                </span>
              )}
            </Button>
          ))}
        </div>

        {/* Add Form */}
        {showAdd && (
          <Card>
            <CardContent className="p-3 space-y-2">
              <Input
                placeholder="Spec title"
                value={addForm.title}
                onChange={e => setAddForm(f => ({ ...f, title: e.target.value }))}
                className="h-8 text-sm"
                autoFocus
              />
              <Textarea
                placeholder="Content (Markdown, optional)"
                value={addForm.content}
                onChange={e => setAddForm(f => ({ ...f, content: e.target.value }))}
                rows={4}
                className="text-sm font-mono"
              />
              <div className="flex gap-2">
                <Button size="sm" className="h-7 text-xs" onClick={handleAdd} disabled={addSpec.isPending}>Add</Button>
                <Button size="sm" variant="ghost" className="h-7 text-xs" onClick={() => setShowAdd(false)}>Cancel</Button>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Spec List */}
        <ScrollArea className="flex-1 border rounded-md">
          <div className="p-2">
            <SpecCards
              items={filteredItems}
              onSelect={s => setSelectedSpecId(s.id)}
              selectedId={selectedSpecId ?? undefined}
            />
            {filteredItems.length === 0 && (
              <p className="text-center text-muted-foreground py-8 text-sm">
                {specItems.length === 0 ? 'No specs yet.' : 'No matching specs.'}
              </p>
            )}
          </div>
        </ScrollArea>
      </div>

      {/* Desktop: Detail Panel */}
      {isDesktop && (
        <div className="w-2/3 border rounded-md flex flex-col min-w-0">
          {selectedSpec ? (
            <>
              <div className="flex items-center justify-between p-4 border-b">
                <h3 className="font-semibold">Spec #{selectedSpec.id}</h3>
                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => { setSelectedSpecId(null); setContentMode('preview'); setEditField(null) }}>
                  <X className="h-4 w-4" />
                </Button>
              </div>
              {detailContent}
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-muted-foreground">
              <div className="text-center">
                <FileText className="h-12 w-12 mx-auto mb-3 opacity-30" />
                <p className="text-sm">Select a spec to view details</p>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Mobile: Full-screen overlay */}
      {!isDesktop && selectedSpec && (
        <div className="fixed inset-0 z-50 bg-background flex flex-col">
          <div className="flex items-center justify-between p-4 border-b">
            <h3 className="font-semibold">Spec #{selectedSpec.id}</h3>
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => { setSelectedSpecId(null); setContentMode('preview'); setEditField(null) }}>
              <X className="h-4 w-4" />
            </Button>
          </div>
          {detailContent}
        </div>
      )}
    </div>
  )
}

// --- Sub-components ---

function SpecStatusBadge({ status }: { status: string }) {
  const variants: Record<string, { variant: any; label: string }> = {
    draft: { variant: 'secondary', label: 'Draft' },
    review: { variant: 'warning', label: 'Review' },
    approved: { variant: 'success', label: 'Approved' },
    deprecated: { variant: 'destructive', label: 'Deprecated' },
  }
  const v = variants[status] || { variant: 'secondary', label: status }
  return <Badge variant={v.variant}>{v.label}</Badge>
}

function SpecTable({ items, onSelect, selectedId }: { items: Spec[]; onSelect: (s: Spec) => void; selectedId?: number }) {
  return (
    <table className="w-full text-sm">
      <thead>
        <tr className="border-b text-left">
          <th className="py-2.5 px-2 w-[50px]">ID</th>
          <th className="py-2.5 px-2">Title</th>
          <th className="py-2.5 px-2 w-[100px]">Status</th>
          <th className="py-2.5 px-2 w-[60px]">Priority</th>
          <th className="py-2.5 px-2 w-[100px]">Updated</th>
        </tr>
      </thead>
      <tbody>
        {items.map((s: Spec) => (
          <tr
            key={s.id}
            className={`border-b cursor-pointer hover:bg-accent ${s.id === selectedId ? 'bg-accent' : ''}`}
            onClick={() => onSelect(s)}
          >
            <td className="py-2.5 px-2 text-muted-foreground">#{s.id}</td>
            <td className="py-2.5 px-2 truncate max-w-[200px]">{s.title}</td>
            <td className="py-2.5 px-2"><SpecStatusBadge status={s.status} /></td>
            <td className="py-2.5 px-2 text-center">{s.priority}</td>
            <td className="py-2.5 px-2 text-muted-foreground text-xs">{new Date(s.updated_at).toLocaleDateString()}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

function SpecCards({ items, onSelect, selectedId }: { items: Spec[]; onSelect: (s: Spec) => void; selectedId?: number }) {
  return (
    <div className="space-y-2">
      {items.map((s: Spec) => (
        <div
          key={s.id}
          className={cn(
            "border rounded-md p-3 cursor-pointer hover:bg-accent",
            s.id === selectedId && "bg-accent border-primary"
          )}
          onClick={() => onSelect(s)}
        >
          <div className="flex items-center justify-between mb-1">
            <span className="text-xs text-muted-foreground">#{s.id}</span>
            <SpecStatusBadge status={s.status} />
          </div>
          <div className="font-medium text-sm truncate">{s.title}</div>
          <div className="flex items-center gap-3 mt-1.5 text-xs text-muted-foreground">
            <span>priority: {s.priority}</span>
            <span>{new Date(s.updated_at).toLocaleDateString()}</span>
          </div>
          {s.content && (
            <div className="mt-1.5 text-xs text-muted-foreground line-clamp-2">
              {s.content.slice(0, 100)}
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

function parseItems(data: any): Spec[] {
  if (!data) return []
  if (Array.isArray(data)) return data
  if (data.items && Array.isArray(data.items)) return data.items
  return []
}
