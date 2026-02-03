import { useState } from 'react';
import { useStore } from './store';
import { ProjectPanel } from './components/ProjectPanel';
import { FeatureList } from './components/FeatureList';
import { TaskPanel } from './components/TaskPanel';
import ExpertsPanel from './components/ExpertsPanel';
import { MessagesPanel } from './components/MessagesPanel';
import { StatusBar } from './components/StatusBar';
import { useSync } from './hooks/useSync';

function App() {
  const { project, selectedFeatureId } = useStore();
  const { isConnected, lastSync, error } = useSync();
  const [view, setView] = useState<'project' | 'features' | 'tasks' | 'experts' | 'messages'>('project');

  return (
    <div className="flex flex-col h-screen">
      {/* Header */}
      <header className="flex items-center justify-between px-4 py-2 border-b border-vscode-border">
        <div className="flex items-center gap-4">
          <h1 className="text-lg font-semibold">
            {project?.name || 'No Project'}
          </h1>
          {project && (
            <span className="text-sm opacity-70">{project.id}</span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setView('project')}
            className={`px-3 py-1 rounded ${
              view === 'project'
                ? 'bg-vscode-button-bg text-vscode-button-fg'
                : 'hover:bg-vscode-list-hover'
            }`}
          >
            Project
          </button>
          <button
            onClick={() => setView('messages')}
            className={`px-3 py-1 rounded ${
              view === 'messages'
                ? 'bg-vscode-button-bg text-vscode-button-fg'
                : 'hover:bg-vscode-list-hover'
            }`}
          >
            Messages
          </button>
          <button
            onClick={() => setView('features')}
            className={`px-3 py-1 rounded ${
              view === 'features'
                ? 'bg-vscode-button-bg text-vscode-button-fg'
                : 'hover:bg-vscode-list-hover'
            }`}
          >
            Features
          </button>
          <button
            onClick={() => setView('tasks')}
            className={`px-3 py-1 rounded ${
              view === 'tasks'
                ? 'bg-vscode-button-bg text-vscode-button-fg'
                : 'hover:bg-vscode-list-hover'
            }`}
          >
            Tasks
          </button>
          <button
            onClick={() => setView('experts')}
            className={`px-3 py-1 rounded ${
              view === 'experts'
                ? 'bg-vscode-button-bg text-vscode-button-fg'
                : 'hover:bg-vscode-list-hover'
            }`}
          >
            Experts
          </button>
        </div>
      </header>

      {/* Error Banner */}
      {error && (
        <div className="px-4 py-2 bg-red-900 text-red-100">
          {error}
        </div>
      )}

      {/* Main Content */}
      <main className="flex-1 overflow-hidden">
        {view === 'project' && <ProjectPanel />}
        {view === 'features' && <FeatureList />}
        {view === 'tasks' && <TaskPanel featureId={selectedFeatureId} />}
        {view === 'experts' && <ExpertsPanel />}
        {view === 'messages' && <MessagesPanel />}
      </main>

      {/* Status Bar */}
      <StatusBar
        isConnected={isConnected}
        lastSync={lastSync}
        projectStatus={project?.status}
      />
    </div>
  );
}

export default App;
