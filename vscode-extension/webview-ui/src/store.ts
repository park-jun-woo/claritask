import { create } from 'zustand';
import type { Project, Feature, Task, Edge, Expert, ProjectData } from './types';

interface AppState {
  // Data
  project: Project | null;
  features: Feature[];
  tasks: Task[];
  taskEdges: Edge[];
  featureEdges: Edge[];
  experts: Expert[];
  projectExperts: string[];
  context: Record<string, any> | null;
  tech: Record<string, any> | null;
  design: Record<string, any> | null;
  state: Record<string, string>;

  // UI State
  selectedFeatureId: number | null;
  selectedTaskId: number | null;
  selectedExpertId: string | null;
  editingFeatureId: number | null;
  editingTaskId: number | null;

  // Sync State
  lastSyncTimestamp: number | null;
  pendingSaves: Map<string, any>;
  conflicts: Set<string>;

  // Actions
  setData: (data: ProjectData) => void;
  setSelectedFeature: (id: number | null) => void;
  setSelectedTask: (id: number | null) => void;
  setSelectedExpert: (id: string | null) => void;
  setEditingFeature: (id: number | null) => void;
  setEditingTask: (id: number | null) => void;
  addPendingSave: (key: string, data: any) => void;
  removePendingSave: (key: string) => void;
  addConflict: (key: string) => void;
  removeConflict: (key: string) => void;

  // Getters
  getFeature: (id: number) => Feature | undefined;
  getTask: (id: number) => Task | undefined;
  getExpert: (id: string) => Expert | undefined;
  getTasksForFeature: (featureId: number) => Task[];
  getTaskDependencies: (taskId: number) => number[];
  getTaskDependents: (taskId: number) => number[];
  getAssignedExperts: () => Expert[];
}

export const useStore = create<AppState>((set, get) => ({
  // Initial Data
  project: null,
  features: [],
  tasks: [],
  taskEdges: [],
  featureEdges: [],
  experts: [],
  projectExperts: [],
  context: null,
  tech: null,
  design: null,
  state: {},

  // Initial UI State
  selectedFeatureId: null,
  selectedTaskId: null,
  selectedExpertId: null,
  editingFeatureId: null,
  editingTaskId: null,

  // Initial Sync State
  lastSyncTimestamp: null,
  pendingSaves: new Map(),
  conflicts: new Set(),

  // Actions
  setData: (data) =>
    set({
      project: data.project,
      features: data.features,
      tasks: data.tasks,
      taskEdges: data.taskEdges,
      featureEdges: data.featureEdges,
      experts: data.experts || [],
      projectExperts: data.projectExperts || [],
      context: data.context,
      tech: data.tech,
      design: data.design,
      state: data.state,
      lastSyncTimestamp: Date.now(),
    }),

  setSelectedFeature: (id) => set({ selectedFeatureId: id }),
  setSelectedTask: (id) => set({ selectedTaskId: id }),
  setSelectedExpert: (id) => set({ selectedExpertId: id }),
  setEditingFeature: (id) => set({ editingFeatureId: id }),
  setEditingTask: (id) => set({ editingTaskId: id }),

  addPendingSave: (key, data) =>
    set((state) => {
      const newMap = new Map(state.pendingSaves);
      newMap.set(key, data);
      return { pendingSaves: newMap };
    }),

  removePendingSave: (key) =>
    set((state) => {
      const newMap = new Map(state.pendingSaves);
      newMap.delete(key);
      return { pendingSaves: newMap };
    }),

  addConflict: (key) =>
    set((state) => {
      const newSet = new Set(state.conflicts);
      newSet.add(key);
      return { conflicts: newSet };
    }),

  removeConflict: (key) =>
    set((state) => {
      const newSet = new Set(state.conflicts);
      newSet.delete(key);
      return { conflicts: newSet };
    }),

  // Getters
  getFeature: (id) => get().features.find((f) => f.id === id),
  getTask: (id) => get().tasks.find((t) => t.id === id),
  getExpert: (id) => get().experts.find((e) => e.id === id),
  getTasksForFeature: (featureId) => get().tasks.filter((t) => t.feature_id === featureId),
  getTaskDependencies: (taskId) =>
    get()
      .taskEdges.filter((e) => e.to_id === taskId)
      .map((e) => e.from_id),
  getTaskDependents: (taskId) =>
    get()
      .taskEdges.filter((e) => e.from_id === taskId)
      .map((e) => e.to_id),
  getAssignedExperts: () => {
    const { experts, projectExperts } = get();
    return experts.filter((e) => projectExperts.includes(e.id));
  },
}));
