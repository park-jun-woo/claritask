import { useState } from 'react';
import { useStore } from '../store';
import { sendMessageCLI, deleteMessage } from '../hooks/useSync';
import type { Message } from '../types';

export function MessagesPanel() {
  const { messages, features, selectedMessageId, setSelectedMessage } = useStore();
  const [showCreateForm, setShowCreateForm] = useState(false);

  return (
    <div className="flex h-full">
      {/* Message List - 1/3 */}
      <div className="w-1/3 border-r border-vscode-border overflow-y-auto">
        <div className="p-2 border-b border-vscode-border flex justify-between items-center">
          <span className="font-semibold">Messages ({messages.length})</span>
          <button
            onClick={() => setShowCreateForm(true)}
            className="px-2 py-1 text-xs bg-vscode-button-bg text-vscode-button-fg rounded hover:bg-vscode-button-hover"
          >
            + New
          </button>
        </div>

        {showCreateForm && (
          <CreateMessageForm
            features={features}
            onClose={() => setShowCreateForm(false)}
          />
        )}

        <ul>
          {messages.map((msg) => (
            <MessageListItem
              key={msg.id}
              message={msg}
              isSelected={selectedMessageId === msg.id}
              featureName={getFeatureName(features, msg.feature_id)}
              onClick={() => setSelectedMessage(msg.id)}
            />
          ))}
        </ul>

        {messages.length === 0 && !showCreateForm && (
          <div className="p-4 text-center opacity-50 text-sm">
            No messages yet
          </div>
        )}
      </div>

      {/* Message Detail - 2/3 */}
      <div className="flex-1 overflow-y-auto">
        {selectedMessageId ? (
          <MessageDetail
            messageId={selectedMessageId}
            features={features}
          />
        ) : (
          <div className="flex items-center justify-center h-full opacity-50">
            Select a message to view details
          </div>
        )}
      </div>
    </div>
  );
}

function getFeatureName(features: { id: number; name: string }[], featureId: number | null): string | null {
  if (!featureId) return null;
  const feature = features.find(f => f.id === featureId);
  return feature?.name || null;
}

interface CreateMessageFormProps {
  features: { id: number; name: string }[];
  onClose: () => void;
}

function CreateMessageForm({ features, onClose }: CreateMessageFormProps) {
  const [content, setContent] = useState('');
  const [featureId, setFeatureId] = useState<number | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (content.trim()) {
      sendMessageCLI(content.trim(), featureId ?? undefined);
      onClose();
    }
  };

  return (
    <form onSubmit={handleSubmit} className="p-3 border-b border-vscode-border bg-vscode-input-bg">
      <div className="mb-2">
        <label className="block text-xs mb-1 opacity-70">Feature (Optional)</label>
        <select
          value={featureId ?? ''}
          onChange={(e) => setFeatureId(e.target.value ? Number(e.target.value) : null)}
          className="w-full px-2 py-1 text-sm bg-vscode-input-bg border border-vscode-border rounded"
        >
          <option value="">-- None --</option>
          {features.map((f) => (
            <option key={f.id} value={f.id}>{f.name}</option>
          ))}
        </select>
      </div>
      <div className="mb-2">
        <label className="block text-xs mb-1 opacity-70">Message Content</label>
        <textarea
          placeholder="Enter your modification request..."
          value={content}
          onChange={(e) => setContent(e.target.value)}
          className="w-full px-2 py-1 text-sm bg-vscode-input-bg border border-vscode-border rounded resize-none"
          rows={3}
          autoFocus
        />
      </div>
      <div className="flex gap-2">
        <button
          type="submit"
          disabled={!content.trim()}
          className="px-2 py-1 text-xs bg-vscode-button-bg text-vscode-button-fg rounded hover:bg-vscode-button-hover disabled:opacity-50"
        >
          Send
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

interface MessageListItemProps {
  message: Message;
  isSelected: boolean;
  featureName: string | null;
  onClick: () => void;
}

function MessageListItem({ message, isSelected, featureName, onClick }: MessageListItemProps) {
  const statusIcon = message.status === 'pending' || message.status === 'processing' ? '●' : '○';

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    return `${diffDays}d ago`;
  };

  return (
    <li
      onClick={onClick}
      className={`p-3 cursor-pointer border-b border-vscode-border ${
        isSelected
          ? 'bg-vscode-list-active'
          : 'hover:bg-vscode-list-hover'
      }`}
    >
      <div className="flex items-start gap-2">
        <span className={`text-xs ${getStatusColor(message.status)}`}>{statusIcon}</span>
        <div className="flex-1 min-w-0">
          <div className="text-sm truncate">{message.content}</div>
          <div className="flex items-center gap-2 mt-1 text-xs opacity-60">
            <StatusBadge status={message.status} />
            <span>{formatTime(message.created_at)}</span>
          </div>
          {featureName && (
            <div className="text-xs opacity-50 mt-1">
              Feature: {featureName}
            </div>
          )}
        </div>
      </div>
    </li>
  );
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'pending': return 'text-yellow-400';
    case 'processing': return 'text-blue-400';
    case 'completed': return 'text-green-400';
    case 'failed': return 'text-red-400';
    default: return 'text-gray-400';
  }
}

interface MessageDetailProps {
  messageId: number;
  features: { id: number; name: string }[];
}

function MessageDetail({ messageId, features }: MessageDetailProps) {
  const { getMessage, setSelectedMessage } = useStore();
  const message = getMessage(messageId);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  if (!message) {
    return <div className="p-4">Message not found</div>;
  }

  const featureName = getFeatureName(features, message.feature_id);

  const handleDelete = () => {
    deleteMessage(message.id);
    setSelectedMessage(null);
    setShowDeleteConfirm(false);
  };

  return (
    <div className="p-4">
      {showDeleteConfirm && (
        <ConfirmModal
          title="Delete Message"
          message="Are you sure you want to delete this message?"
          onConfirm={handleDelete}
          onCancel={() => setShowDeleteConfirm(false)}
        />
      )}

      <div className="flex items-start justify-between mb-4">
        <div>
          <h2 className="text-lg font-semibold">Message #{message.id}</h2>
        </div>
        <button
          onClick={() => setShowDeleteConfirm(true)}
          className="px-3 py-1 text-sm border border-red-600 text-red-400 rounded hover:bg-red-900"
        >
          Delete
        </button>
      </div>

      <div className="space-y-4">
        <div>
          <h3 className="text-sm font-semibold mb-1">Status</h3>
          <StatusBadge status={message.status} large />
        </div>

        {featureName && (
          <div>
            <h3 className="text-sm font-semibold mb-1">Feature</h3>
            <span className="text-sm px-2 py-1 bg-vscode-input-bg rounded">{featureName}</span>
          </div>
        )}

        <div>
          <h3 className="text-sm font-semibold mb-1">Content</h3>
          <div className="text-sm p-3 bg-vscode-input-bg rounded whitespace-pre-wrap">
            {message.content}
          </div>
        </div>

        {message.status === 'completed' && message.response && (
          <div>
            <h3 className="text-sm font-semibold mb-1 text-green-400">Response</h3>
            <div className="text-sm p-3 bg-green-900/30 border border-green-800 rounded whitespace-pre-wrap">
              {message.response}
            </div>
          </div>
        )}

        {message.status === 'failed' && message.error && (
          <div>
            <h3 className="text-sm font-semibold mb-1 text-red-400">Error</h3>
            <div className="text-sm p-3 bg-red-900/30 border border-red-800 rounded whitespace-pre-wrap">
              {message.error}
            </div>
          </div>
        )}

        <div className="text-xs opacity-50 space-y-1">
          <div>Created: {new Date(message.created_at).toLocaleString()}</div>
          {message.completed_at && (
            <div>Completed: {new Date(message.completed_at).toLocaleString()}</div>
          )}
        </div>
      </div>
    </div>
  );
}

function StatusBadge({ status, large = false }: { status: string; large?: boolean }) {
  const colors: Record<string, string> = {
    pending: 'bg-yellow-800 text-yellow-200',
    processing: 'bg-blue-800 text-blue-200',
    completed: 'bg-green-800 text-green-200',
    failed: 'bg-red-800 text-red-200',
  };

  const sizeClass = large ? 'px-3 py-1 text-sm' : 'px-2 py-0.5 text-xs';

  return (
    <span className={`${sizeClass} rounded ${colors[status] || 'bg-gray-700'}`}>
      {status}
    </span>
  );
}

interface ConfirmModalProps {
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
}

function ConfirmModal({ title, message, onConfirm, onCancel }: ConfirmModalProps) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
      <div className="bg-vscode-editor-bg border border-vscode-border rounded-lg shadow-xl max-w-md w-full mx-4">
        <div className="p-4 border-b border-vscode-border">
          <h3 className="text-lg font-semibold">{title}</h3>
        </div>
        <div className="p-4">
          <p className="text-sm opacity-80">{message}</p>
        </div>
        <div className="p-4 border-t border-vscode-border flex justify-end gap-2">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-sm border border-vscode-border rounded hover:bg-vscode-list-hover"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className="px-4 py-2 text-sm bg-red-600 text-white rounded hover:bg-red-700"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  );
}
