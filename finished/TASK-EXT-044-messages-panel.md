# TASK-EXT-044: MessagesPanel 컴포넌트

## 목표
Messages 탭에 표시할 메시지 목록 패널 컴포넌트 구현

## 작업 내용

### 1. webview-ui/src/components/MessagesPanel.tsx 생성

```tsx
import { useState } from 'react';
import { useStore } from '../store';
import { vscode } from '../vscode';
import { SectionCard } from './SectionCard';

export function MessagesPanel() {
  const { messages, features, selectedMessageId, setSelectedMessage } = useStore();
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [newContent, setNewContent] = useState('');
  const [selectedFeatureId, setSelectedFeatureId] = useState<number | null>(null);

  const handleSend = () => {
    if (!newContent.trim()) return;

    vscode.postMessage({
      type: 'sendMessage',
      content: newContent,
      featureId: selectedFeatureId ?? undefined,
    });

    setNewContent('');
    setShowCreateDialog(false);
  };

  const handleDelete = (messageId: number) => {
    if (confirm('이 메시지를 삭제하시겠습니까?')) {
      vscode.postMessage({
        type: 'deleteMessage',
        messageId,
      });
    }
  };

  const getStatusBadge = (status: string) => {
    const colors: Record<string, string> = {
      pending: 'bg-yellow-600',
      processing: 'bg-blue-600',
      completed: 'bg-green-600',
      failed: 'bg-red-600',
    };
    return (
      <span className={`px-2 py-0.5 text-xs rounded ${colors[status] || 'bg-gray-600'}`}>
        {status}
      </span>
    );
  };

  const getFeatureName = (featureId: number | null) => {
    if (!featureId) return null;
    const feature = features.find(f => f.id === featureId);
    return feature?.name;
  };

  return (
    <div className="p-4 h-full overflow-y-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold">Messages ({messages.length})</h2>
        <button
          onClick={() => setShowCreateDialog(true)}
          className="px-3 py-1 bg-vscode-button-bg text-vscode-button-fg rounded hover:opacity-80"
        >
          + New Message
        </button>
      </div>

      {/* Create Dialog */}
      {showCreateDialog && (
        <SectionCard title="New Message" className="mb-4">
          <div className="space-y-3">
            <div>
              <label className="block text-sm mb-1">Feature (Optional)</label>
              <select
                value={selectedFeatureId ?? ''}
                onChange={(e) => setSelectedFeatureId(e.target.value ? Number(e.target.value) : null)}
                className="w-full px-2 py-1 bg-vscode-input-bg border border-vscode-border rounded"
              >
                <option value="">-- None --</option>
                {features.map((f) => (
                  <option key={f.id} value={f.id}>{f.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm mb-1">Message Content</label>
              <textarea
                value={newContent}
                onChange={(e) => setNewContent(e.target.value)}
                placeholder="Enter your modification request..."
                rows={4}
                className="w-full px-2 py-1 bg-vscode-input-bg border border-vscode-border rounded resize-none"
              />
            </div>
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => setShowCreateDialog(false)}
                className="px-3 py-1 border border-vscode-border rounded hover:bg-vscode-list-hover"
              >
                Cancel
              </button>
              <button
                onClick={handleSend}
                disabled={!newContent.trim()}
                className="px-3 py-1 bg-vscode-button-bg text-vscode-button-fg rounded hover:opacity-80 disabled:opacity-50"
              >
                Send
              </button>
            </div>
          </div>
        </SectionCard>
      )}

      {/* Message List */}
      <div className="space-y-2">
        {messages.length === 0 ? (
          <div className="text-center py-8 opacity-60">
            No messages yet. Click "New Message" to create one.
          </div>
        ) : (
          messages.map((msg) => (
            <div
              key={msg.id}
              onClick={() => setSelectedMessage(msg.id)}
              className={`p-3 border rounded cursor-pointer transition-colors ${
                selectedMessageId === msg.id
                  ? 'border-vscode-focusBorder bg-vscode-list-activeSelectionBg'
                  : 'border-vscode-border hover:bg-vscode-list-hover'
              }`}
            >
              <div className="flex items-start justify-between gap-2">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    {getStatusBadge(msg.status)}
                    {getFeatureName(msg.feature_id) && (
                      <span className="text-xs opacity-70">
                        Feature: {getFeatureName(msg.feature_id)}
                      </span>
                    )}
                  </div>
                  <p className="text-sm truncate">{msg.content}</p>
                  <p className="text-xs opacity-50 mt-1">
                    {new Date(msg.created_at).toLocaleString()}
                  </p>
                </div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    handleDelete(msg.id);
                  }}
                  className="p-1 hover:bg-red-600 rounded opacity-60 hover:opacity-100"
                  title="Delete"
                >
                  ✕
                </button>
              </div>

              {/* Response (if completed) */}
              {msg.status === 'completed' && msg.response && (
                <div className="mt-2 pt-2 border-t border-vscode-border">
                  <p className="text-xs opacity-70 mb-1">Response:</p>
                  <p className="text-sm whitespace-pre-wrap">{msg.response}</p>
                </div>
              )}

              {/* Error (if failed) */}
              {msg.status === 'failed' && msg.error && (
                <div className="mt-2 pt-2 border-t border-red-600">
                  <p className="text-xs text-red-400 mb-1">Error:</p>
                  <p className="text-sm text-red-300">{msg.error}</p>
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
```

## 완료 조건
- [ ] MessagesPanel.tsx 파일 생성
- [ ] 메시지 목록 표시
- [ ] 메시지 생성 다이얼로그
- [ ] 메시지 삭제 기능
- [ ] 상태별 배지 색상
- [ ] Feature 연결 표시
- [ ] Response/Error 표시
