import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { projectAPI, taskAPI, specAPI, messageAPI, scheduleAPI, statusAPI, health } from '@/api/client'
import type { StatusResponse } from '@/types'

// --- Health ---
export function useHealth() {
  return useQuery({
    queryKey: ['health'],
    queryFn: health,
    refetchInterval: 30_000,
    retry: false,
  })
}

// --- Status ---
export function useStatus() {
  return useQuery<StatusResponse>({
    queryKey: ['status'],
    queryFn: statusAPI.get,
    refetchInterval: (query) => {
      const data = query.state.data
      if (data?.cycle_status?.status === 'running') return 5_000
      return 15_000
    },
  })
}

// --- Projects ---
export function useProjects(all = true) {
  return useQuery({
    queryKey: ['projects', { all }],
    queryFn: () => projectAPI.list(all),
  })
}

export function useProject(id?: string) {
  return useQuery({
    queryKey: ['project', id],
    queryFn: () => projectAPI.get(id),
    enabled: !!id,
  })
}

export function useProjectStats() {
  return useQuery({
    queryKey: ['projectStats'],
    queryFn: projectAPI.stats,
    refetchInterval: 30_000,
  })
}

export function useSwitchProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => id === 'none' ? projectAPI.switchNone() : projectAPI.switch(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['status'] })
      qc.invalidateQueries({ queryKey: ['tasks'] })
      qc.invalidateQueries({ queryKey: ['task'] })
      qc.invalidateQueries({ queryKey: ['messages'] })
      qc.invalidateQueries({ queryKey: ['specs'] })
      qc.invalidateQueries({ queryKey: ['spec'] })
      qc.invalidateQueries({ queryKey: ['project'] })
    },
  })
}

export function useDeleteProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => projectAPI.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['projects'] })
      qc.invalidateQueries({ queryKey: ['status'] })
    },
  })
}

export function useUpdateProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: { description?: string; parallel?: number } }) =>
      projectAPI.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['projects'] })
      qc.invalidateQueries({ queryKey: ['project'] })
      qc.invalidateQueries({ queryKey: ['projectStats'] })
    },
  })
}

// --- Tasks ---
export function useTasks(all = true) {
  return useQuery({
    queryKey: ['tasks', { all, tree: true }],
    queryFn: () => taskAPI.list(undefined, false, true),
    refetchInterval: 15_000,
  })
}

export function useTask(id?: number | string) {
  return useQuery({
    queryKey: ['task', id],
    queryFn: () => taskAPI.get(id!),
    enabled: id !== undefined,
  })
}

export function useAddTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { title: string; parentId?: number; spec?: string }) =>
      taskAPI.add(params.title, params.parentId, params.spec),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tasks'] }),
  })
}

export function useSetTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { id: number | string; field: string; value: string }) =>
      taskAPI.set(params.id, params.field, params.value),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] })
      qc.invalidateQueries({ queryKey: ['task'] })
    },
  })
}

export function useDeleteTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number | string) => taskAPI.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tasks'] }),
  })
}

export function useTaskPlan() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id?: number | string) => id !== undefined ? taskAPI.plan(id) : taskAPI.planAll(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tasks'] }),
  })
}

export function useTaskRun() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id?: number | string) => id !== undefined ? taskAPI.run(id) : taskAPI.runAll(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tasks'] }),
  })
}

export function useTaskCycle() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => taskAPI.cycle(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tasks'] }),
  })
}

export function useTaskStop() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => taskAPI.stop(),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] })
      qc.invalidateQueries({ queryKey: ['status'] })
    },
  })
}

// --- Messages ---
export function useMessages(all = true) {
  return useQuery({
    queryKey: ['messages', { all }],
    queryFn: () => messageAPI.list(all),
    refetchInterval: 10_000,
  })
}

export function useMessage(id?: number | string) {
  return useQuery({
    queryKey: ['message', id],
    queryFn: () => messageAPI.get(id!),
    enabled: id !== undefined,
    refetchInterval: 5_000,
  })
}

export function useMessageStatus() {
  return useQuery({
    queryKey: ['messageStatus'],
    queryFn: messageAPI.status,
    refetchInterval: 5_000,
  })
}

export function useSendMessage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ content, projectId }: { content: string; projectId?: string }) =>
      messageAPI.send(content, projectId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['messages'] })
      qc.invalidateQueries({ queryKey: ['messageStatus'] })
    },
  })
}

// --- Schedules ---
export function useSchedules(all = true) {
  return useQuery({
    queryKey: ['schedules', { all }],
    queryFn: () => scheduleAPI.list(all),
  })
}

export function useScheduleRuns(scheduleId: number | string) {
  return useQuery({
    queryKey: ['scheduleRuns', scheduleId],
    queryFn: () => scheduleAPI.runs(scheduleId),
    enabled: !!scheduleId,
  })
}

export function useAddSchedule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { cronExpr: string; message: string; projectId?: string; once?: boolean; type?: 'claude' | 'bash' }) =>
      scheduleAPI.add(params.cronExpr, params.message, params.projectId, params.once, params.type),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedules'] }),
  })
}

export function useDeleteSchedule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number | string) => scheduleAPI.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedules'] }),
  })
}

export function useToggleSchedule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { id: number | string; enable: boolean }) =>
      params.enable ? scheduleAPI.enable(params.id) : scheduleAPI.disable(params.id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedules'] }),
  })
}

// --- Specs ---
export function useSpecs(all = true) {
  return useQuery({
    queryKey: ['specs', { all }],
    queryFn: () => specAPI.list(all),
  })
}

export function useSpec(id?: number | string) {
  return useQuery({
    queryKey: ['spec', id],
    queryFn: () => specAPI.get(id!),
    enabled: id !== undefined,
  })
}

export function useAddSpec() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { title: string; content?: string }) =>
      specAPI.add(params.title, params.content),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['specs'] }),
  })
}

export function useSetSpec() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { id: number | string; field: string; value: string }) =>
      specAPI.set(params.id, params.field, params.value),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['specs'] })
      qc.invalidateQueries({ queryKey: ['spec'] })
    },
  })
}

export function useDeleteSpec() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number | string) => specAPI.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['specs'] }),
  })
}
