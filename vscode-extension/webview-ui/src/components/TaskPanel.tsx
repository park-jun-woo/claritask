import { useState } from 'react';
import { useStore } from '../store';
import { createTask, saveTask } from '../hooks/useSync';
import type { Task } from '../types';

interface TaskPanelProps {
  featureId: number | null;
}

export function TaskPanel({ featureId }: TaskPanelProps) {
  const { tasks, getTasksForFeature, selectedTaskId, setSelectedTask, conflicts } = useStore();
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [filter, setFilter] = useState<string>('all');

  const displayTasks = featureId ? getTasksForFeature(featureId) : tasks;
  const filteredTasks = filter === 'all'
    ? displayTasks
    : displayTasks.filter((t) => t.status === filter);

  return (
    <div className="flex h-full">
      {/* Task List */}
      <div className="w-1/3 border-r border-vscode-border overflow-y-auto">
        <div className="p-2 border-b border-vscode-border">
          <div className="flex justify-between items-center mb-2">
            <span className="font-semibold">Tasks ({filteredTasks.length})</span>
            {featureId && (
              <button
                onClick={() => setShowCreateForm(true)}
                className="px-2 py-1 text-xs bg-vscode-button-bg text-vscode-button-fg rounded hover:bg-vscode-button-hover"
              >
                + New
              </button>
            )}
          </div>
          <div className="flex gap-1">
            {['all', 'pending', 'doing', 'done', 'failed'].map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={`px-2 py-0.5 text-xs rounded ${
                  filter === f
                    ? 'bg-vscode-button-bg text-vscode-button-fg'
                    : 'hover:bg-vscode-list-hover'
                }`}
              >
                {f}
              </button>
            ))}
          </div>
        </div>

        {showCreateForm && featureId && (
          <CreateTaskForm featureId={featureId} onClose={() => setShowCreateForm(false)} />
        )}

        <ul>
          {filteredTasks.map((task) => (
            <li
              key={task.id}
              onClick={() => setSelectedTask(task.id)}
              className={`p-3 cursor-pointer border-b border-vscode-border ${
                selectedTaskId === task.id
                  ? 'bg-vscode-list-active'
                  : 'hover:bg-vscode-list-hover'
              } ${conflicts.has(`tasks:${task.id}`) ? 'border-l-4 border-l-yellow-500' : ''}`}
            >
              <div className="flex items-center justify-between">
                <span className="font-medium truncate">{task.title}</span>
                <TaskStatusBadge status={task.status} />
              </div>
              {task.target_file && (
                <div className="text-xs opacity-70 mt-1 truncate">
                  {task.target_file}
                  {task.target_line && `:${task.target_line}`}
                </div>
              )}
            </li>
          ))}
        </ul>
      </div>

      {/* Task Detail */}
      <div className="flex-1 overflow-y-auto">
        {selectedTaskId ? (
          <TaskDetail
            taskId={selectedTaskId}
            isEditing={editingId === selectedTaskId}
            onEdit={() => setEditingId(selectedTaskId)}
            onCancelEdit={() => setEditingId(null)}
          />
        ) : (
          <div className="flex items-center justify-center h-full opacity-50">
            Select a task to view details
          </div>
        )}
      </div>
    </div>
  );
}

function CreateTaskForm({ featureId, onClose }: { featureId: number; onClose: () => void }) {
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (title.trim()) {
      createTask(featureId, title.trim(), content.trim());
      onClose();
    }
  };

  return (
    <form onSubmit={handleSubmit} className="p-3 border-b border-vscode-border bg-vscode-input-bg">
      <input
        type="text"
        placeholder="Task title"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        className="w-full px-2 py-1 mb-2 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded"
        autoFocus
      />
      <textarea
        placeholder="Content (optional)"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        className="w-full px-2 py-1 mb-2 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded resize-none"
        rows={3}
      />
      <div className="flex gap-2">
        <button
          type="submit"
          className="px-2 py-1 text-xs bg-vscode-button-bg text-vscode-button-fg rounded hover:bg-vscode-button-hover"
        >
          Create
        </button>
        <button
          type="button"
          onClick={onClose}
          className="px-2 py-1 text-xs border border-vscode-border rounded hover:bg-vscode-list-hover"
        >
          Cancel
        </button>
      </div>
    </form>
  );
}

interface TaskDetailProps {
  taskId: number;
  isEditing: boolean;
  onEdit: () => void;
  onCancelEdit: () => void;
}

function TaskDetail({ taskId, isEditing, onEdit, onCancelEdit }: TaskDetailProps) {
  const { getTask, getFeature, getTaskDependencies, getTaskDependents } = useStore();
  const task = getTask(taskId);

  if (!task) {
    return <div className="p-4">Task not found</div>;
  }

  const feature = getFeature(task.feature_id);
  const dependencies = getTaskDependencies(taskId);
  const dependents = getTaskDependents(taskId);

  if (isEditing) {
    return <TaskEditForm task={task} onCancel={onCancelEdit} />;
  }

  return (
    <div className="p-4">
      <div className="flex items-start justify-between mb-4">
        <div>
          <h2 className="text-xl font-semibold">{task.title}</h2>
          <div className="text-sm opacity-70">
            ID: {task.id} | Feature: {feature?.name || task.feature_id}
          </div>
        </div>
        <button
          onClick={onEdit}
          className="px-3 py-1 text-sm border border-vscode-border rounded hover:bg-vscode-list-hover"
        >
          Edit
        </button>
      </div>

      <div className="space-y-4">
        <div>
          <h3 className="text-sm font-semibold mb-1">Status</h3>
          <TaskStatusBadge status={task.status} />
        </div>

        {task.target_file && (
          <div>
            <h3 className="text-sm font-semibold mb-1">Target</h3>
            <code className="text-sm bg-vscode-input-bg px-2 py-1 rounded">
              {task.target_file}
              {task.target_line && `:${task.target_line}`}
              {task.target_function && ` (${task.target_function})`}
            </code>
          </div>
        )}

        {task.content && (
          <div>
            <h3 className="text-sm font-semibold mb-1">Content</h3>
            <pre className="text-xs p-2 bg-vscode-input-bg rounded overflow-x-auto whitespace-pre-wrap">
              {task.content}
            </pre>
          </div>
        )}

        {task.result && (
          <div>
            <h3 className="text-sm font-semibold mb-1">Result</h3>
            <pre className="text-xs p-2 bg-green-900 rounded overflow-x-auto whitespace-pre-wrap">
              {task.result}
            </pre>
          </div>
        )}

        {task.error && (
          <div>
            <h3 className="text-sm font-semibold mb-1">Error</h3>
            <pre className="text-xs p-2 bg-red-900 rounded overflow-x-auto whitespace-pre-wrap">
              {task.error}
            </pre>
          </div>
        )}

        {(dependencies.length > 0 || dependents.length > 0) && (
          <div>
            <h3 className="text-sm font-semibold mb-1">Dependencies</h3>
            <div className="text-xs space-y-1">
              {dependencies.length > 0 && (
                <div>
                  <span className="opacity-70">Depends on:</span> {dependencies.join(', ')}
                </div>
              )}
              {dependents.length > 0 && (
                <div>
                  <span className="opacity-70">Blocks:</span> {dependents.join(', ')}
                </div>
              )}
            </div>
          </div>
        )}

        <div className="text-xs opacity-50 space-y-1">
          <div>Created: {new Date(task.created_at).toLocaleString()}</div>
          {task.started_at && <div>Started: {new Date(task.started_at).toLocaleString()}</div>}
          {task.completed_at && <div>Completed: {new Date(task.completed_at).toLocaleString()}</div>}
          {task.failed_at && <div>Failed: {new Date(task.failed_at).toLocaleString()}</div>}
          <div>Version: {task.version}</div>
        </div>
      </div>
    </div>
  );
}

function TaskEditForm({ task, onCancel }: { task: Task; onCancel: () => void }) {
  const [title, setTitle] = useState(task.title);
  const [content, setContent] = useState(task.content);
  const [status, setStatus] = useState(task.status);
  const [targetFile, setTargetFile] = useState(task.target_file);
  const [targetLine, setTargetLine] = useState(task.target_line?.toString() || '');
  const [targetFunction, setTargetFunction] = useState(task.target_function);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    saveTask(
      task.id,
      {
        title,
        content,
        status,
        target_file: targetFile,
        target_line: targetLine ? parseInt(targetLine, 10) : null,
        target_function: targetFunction,
      },
      task.version
    );
    onCancel();
  };

  return (
    <form onSubmit={handleSubmit} className="p-4">
      <h2 className="text-xl font-semibold mb-4">Edit Task</h2>

      <div className="space-y-4">
        <div>
          <label className="block text-sm font-semibold mb-1">Title</label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="w-full px-2 py-1 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded"
          />
        </div>

        <div>
          <label className="block text-sm font-semibold mb-1">Status</label>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value as Task['status'])}
            className="px-2 py-1 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded"
          >
            <option value="pending">Pending</option>
            <option value="doing">Doing</option>
            <option value="done">Done</option>
            <option value="failed">Failed</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-semibold mb-1">Target File</label>
          <input
            type="text"
            value={targetFile}
            onChange={(e) => setTargetFile(e.target.value)}
            className="w-full px-2 py-1 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded"
            placeholder="e.g., src/main.go"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-semibold mb-1">Target Line</label>
            <input
              type="number"
              value={targetLine}
              onChange={(e) => setTargetLine(e.target.value)}
              className="w-full px-2 py-1 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded"
            />
          </div>
          <div>
            <label className="block text-sm font-semibold mb-1">Target Function</label>
            <input
              type="text"
              value={targetFunction}
              onChange={(e) => setTargetFunction(e.target.value)}
              className="w-full px-2 py-1 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-semibold mb-1">Content</label>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            className="w-full px-2 py-1 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded resize-none"
            rows={6}
          />
        </div>

        <div className="flex gap-2">
          <button
            type="submit"
            className="px-3 py-1 bg-vscode-button-bg text-vscode-button-fg rounded hover:bg-vscode-button-hover"
          >
            Save
          </button>
          <button
            type="button"
            onClick={onCancel}
            className="px-3 py-1 border border-vscode-border rounded hover:bg-vscode-list-hover"
          >
            Cancel
          </button>
        </div>
      </div>
    </form>
  );
}

function TaskStatusBadge({ status }: { status: Task['status'] }) {
  const colors: Record<string, string> = {
    pending: 'bg-yellow-800 text-yellow-200',
    doing: 'bg-blue-800 text-blue-200',
    done: 'bg-green-800 text-green-200',
    failed: 'bg-red-800 text-red-200',
  };

  return (
    <span className={`px-2 py-0.5 text-xs rounded ${colors[status] || 'bg-gray-700'}`}>
      {status}
    </span>
  );
}
