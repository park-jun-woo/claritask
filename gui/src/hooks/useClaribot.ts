import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { projectAPI, taskAPI, messageAPI, scheduleAPI, statusAPI, edgeAPI, health } from '@/api/client'

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
  return useQuery({
    queryKey: ['status'],
    queryFn: statusAPI.get,
    refetchInterval: 15_000,
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

export function useSwitchProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => id === 'none' ? projectAPI.switchNone() : projectAPI.switch(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['status'] })
      qc.invalidateQueries({ queryKey: ['tasks'] })
      qc.invalidateQueries({ queryKey: ['messages'] })
    },
  })
}

export function useDeleteProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => projectAPI.delete(id, true),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['projects'] })
      qc.invalidateQueries({ queryKey: ['status'] })
    },
  })
}

// --- Tasks ---
export function useTasks(all = true) {
  return useQuery({
    queryKey: ['tasks', { all }],
    queryFn: () => taskAPI.list(undefined, all),
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
    mutationFn: (id: number | string) => taskAPI.delete(id, true),
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

// --- Edges ---
export function useEdges(taskId?: number | string, all = true) {
  return useQuery({
    queryKey: ['edges', { taskId, all }],
    queryFn: () => edgeAPI.list(taskId, all),
  })
}

export function useAddEdge() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { fromId: number | string; toId: number | string }) =>
      edgeAPI.add(params.fromId, params.toId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['edges'] }),
  })
}

export function useDeleteEdge() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { fromId: number | string; toId: number | string }) =>
      edgeAPI.delete(params.fromId, params.toId, true),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['edges'] }),
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
    mutationFn: (content: string) => messageAPI.send(content),
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
    mutationFn: (params: { cronExpr: string; message: string; projectId?: string; once?: boolean }) =>
      scheduleAPI.add(params.cronExpr, params.message, params.projectId, params.once),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['schedules'] }),
  })
}

export function useDeleteSchedule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number | string) => scheduleAPI.delete(id, true),
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
