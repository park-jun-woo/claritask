import { useState } from 'react';
import { useStore } from '../store';
import { createFeature, saveFeature } from '../hooks/useSync';
import type { Feature } from '../types';

export function FeatureList() {
  const { features, selectedFeatureId, setSelectedFeature, conflicts } = useStore();
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);

  return (
    <div className="flex h-full">
      {/* Feature List */}
      <div className="w-1/3 border-r border-vscode-border overflow-y-auto">
        <div className="p-2 border-b border-vscode-border flex justify-between items-center">
          <span className="font-semibold">Features ({features.length})</span>
          <button
            onClick={() => setShowCreateForm(true)}
            className="px-2 py-1 text-xs bg-vscode-button-bg text-vscode-button-fg rounded hover:bg-vscode-button-hover"
          >
            + New
          </button>
        </div>

        {showCreateForm && (
          <CreateFeatureForm onClose={() => setShowCreateForm(false)} />
        )}

        <ul>
          {features.map((feature) => (
            <li
              key={feature.id}
              onClick={() => setSelectedFeature(feature.id)}
              className={`p-3 cursor-pointer border-b border-vscode-border ${
                selectedFeatureId === feature.id
                  ? 'bg-vscode-list-active'
                  : 'hover:bg-vscode-list-hover'
              } ${conflicts.has(`features:${feature.id}`) ? 'border-l-4 border-l-yellow-500' : ''}`}
            >
              <div className="flex items-center justify-between">
                <span className="font-medium">{feature.name}</span>
                <StatusBadge status={feature.status} />
              </div>
              <div className="text-xs opacity-70 mt-1 truncate">
                {feature.description || 'No description'}
              </div>
            </li>
          ))}
        </ul>
      </div>

      {/* Feature Detail */}
      <div className="flex-1 overflow-y-auto">
        {selectedFeatureId ? (
          <FeatureDetail
            featureId={selectedFeatureId}
            isEditing={editingId === selectedFeatureId}
            onEdit={() => setEditingId(selectedFeatureId)}
            onCancelEdit={() => setEditingId(null)}
          />
        ) : (
          <div className="flex items-center justify-center h-full opacity-50">
            Select a feature to view details
          </div>
        )}
      </div>
    </div>
  );
}

function CreateFeatureForm({ onClose }: { onClose: () => void }) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (name.trim()) {
      createFeature(name.trim(), description.trim());
      onClose();
    }
  };

  return (
    <form onSubmit={handleSubmit} className="p-3 border-b border-vscode-border bg-vscode-input-bg">
      <input
        type="text"
        placeholder="Feature name"
        value={name}
        onChange={(e) => setName(e.target.value)}
        className="w-full px-2 py-1 mb-2 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded"
        autoFocus
      />
      <textarea
        placeholder="Description"
        value={description}
        onChange={(e) => setDescription(e.target.value)}
        className="w-full px-2 py-1 mb-2 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded resize-none"
        rows={2}
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

interface FeatureDetailProps {
  featureId: number;
  isEditing: boolean;
  onEdit: () => void;
  onCancelEdit: () => void;
}

function FeatureDetail({ featureId, isEditing, onEdit, onCancelEdit }: FeatureDetailProps) {
  const { getFeature, getTasksForFeature } = useStore();
  const feature = getFeature(featureId);
  const tasks = getTasksForFeature(featureId);

  if (!feature) {
    return <div className="p-4">Feature not found</div>;
  }

  if (isEditing) {
    return <FeatureEditForm feature={feature} onCancel={onCancelEdit} />;
  }

  const taskStats = {
    total: tasks.length,
    pending: tasks.filter((t) => t.status === 'pending').length,
    doing: tasks.filter((t) => t.status === 'doing').length,
    done: tasks.filter((t) => t.status === 'done').length,
    failed: tasks.filter((t) => t.status === 'failed').length,
  };

  return (
    <div className="p-4">
      <div className="flex items-start justify-between mb-4">
        <div>
          <h2 className="text-xl font-semibold">{feature.name}</h2>
          <div className="text-sm opacity-70">ID: {feature.id}</div>
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
          <StatusBadge status={feature.status} />
        </div>

        <div>
          <h3 className="text-sm font-semibold mb-1">Description</h3>
          <p className="text-sm opacity-80">{feature.description || 'No description'}</p>
        </div>

        <div>
          <h3 className="text-sm font-semibold mb-1">Tasks</h3>
          <div className="flex gap-2 text-xs">
            <span className="px-2 py-1 bg-gray-700 rounded">Total: {taskStats.total}</span>
            <span className="px-2 py-1 bg-yellow-800 rounded">Pending: {taskStats.pending}</span>
            <span className="px-2 py-1 bg-blue-800 rounded">Doing: {taskStats.doing}</span>
            <span className="px-2 py-1 bg-green-800 rounded">Done: {taskStats.done}</span>
            {taskStats.failed > 0 && (
              <span className="px-2 py-1 bg-red-800 rounded">Failed: {taskStats.failed}</span>
            )}
          </div>
        </div>

        {feature.spec && (
          <div>
            <h3 className="text-sm font-semibold mb-1">Spec</h3>
            <pre className="text-xs p-2 bg-vscode-input-bg rounded overflow-x-auto">
              {feature.spec}
            </pre>
          </div>
        )}

        {feature.fdl && (
          <div>
            <h3 className="text-sm font-semibold mb-1">FDL</h3>
            <pre className="text-xs p-2 bg-vscode-input-bg rounded overflow-x-auto">
              {feature.fdl}
            </pre>
          </div>
        )}

        <div className="text-xs opacity-50">
          Created: {new Date(feature.created_at).toLocaleString()}
          {' | '}
          Version: {feature.version}
        </div>
      </div>
    </div>
  );
}

function FeatureEditForm({ feature, onCancel }: { feature: Feature; onCancel: () => void }) {
  const [name, setName] = useState(feature.name);
  const [description, setDescription] = useState(feature.description);
  const [status, setStatus] = useState(feature.status);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    saveFeature(
      feature.id,
      { name, description, status },
      feature.version
    );
    onCancel();
  };

  return (
    <form onSubmit={handleSubmit} className="p-4">
      <h2 className="text-xl font-semibold mb-4">Edit Feature</h2>

      <div className="space-y-4">
        <div>
          <label className="block text-sm font-semibold mb-1">Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-2 py-1 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded"
          />
        </div>

        <div>
          <label className="block text-sm font-semibold mb-1">Description</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="w-full px-2 py-1 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded resize-none"
            rows={3}
          />
        </div>

        <div>
          <label className="block text-sm font-semibold mb-1">Status</label>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className="px-2 py-1 bg-vscode-input-bg text-vscode-input-fg border border-vscode-border rounded"
          >
            <option value="pending">Pending</option>
            <option value="active">Active</option>
            <option value="completed">Completed</option>
            <option value="archived">Archived</option>
          </select>
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

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    pending: 'bg-yellow-800 text-yellow-200',
    active: 'bg-green-800 text-green-200',
    completed: 'bg-blue-800 text-blue-200',
    archived: 'bg-gray-700 text-gray-300',
  };

  return (
    <span className={`px-2 py-0.5 text-xs rounded ${colors[status] || 'bg-gray-700'}`}>
      {status}
    </span>
  );
}
