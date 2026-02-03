# TASK-EXT-028: Expert 생성 다이얼로그

## 개요
새로운 Expert를 생성하는 다이얼로그 컴포넌트

## 배경
- **스펙**: specs/VSCode/07-ExpertsTab.md
- **현재 상태**: Expert 생성 UI 없음

## 작업 내용

### 1. CreateExpertDialog 컴포넌트
**파일**: `vscode-extension/webview-ui/src/components/CreateExpertDialog.tsx`

```tsx
import React, { useState } from 'react';
import { createExpert } from '../vscode';

interface CreateExpertDialogProps {
  isOpen: boolean;
  onClose: () => void;
}

const CreateExpertDialog: React.FC<CreateExpertDialogProps> = ({ isOpen, onClose }) => {
  const [expertId, setExpertId] = useState('');
  const [error, setError] = useState('');

  if (!isOpen) return null;

  const validateId = (id: string): boolean => {
    // 영문 소문자, 숫자, 하이픈만 허용
    const pattern = /^[a-z0-9-]+$/;
    return pattern.test(id) && id.length > 0;
  };

  const handleSubmit = () => {
    if (!validateId(expertId)) {
      setError('ID must contain only lowercase letters, numbers, and hyphens');
      return;
    }

    createExpert(expertId);
    setExpertId('');
    setError('');
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 w-96 shadow-xl">
        <h2 className="text-lg font-bold mb-4">Create New Expert</h2>

        <div className="mb-4">
          <label className="block text-sm font-medium mb-1">Expert ID</label>
          <input
            type="text"
            value={expertId}
            onChange={(e) => {
              setExpertId(e.target.value.toLowerCase());
              setError('');
            }}
            placeholder="e.g., backend-go-gin"
            className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          {error && <p className="text-red-500 text-sm mt-1">{error}</p>}
          <p className="text-gray-500 text-xs mt-1">
            Lowercase letters, numbers, and hyphens only
          </p>
        </div>

        <div className="flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            Create
          </button>
        </div>
      </div>
    </div>
  );
};

export default CreateExpertDialog;
```

### 2. ExpertsPanel에 다이얼로그 연결
```tsx
// ExpertsPanel.tsx에 추가
const [showCreateDialog, setShowCreateDialog] = useState(false);

// 버튼 수정
<button
  onClick={() => setShowCreateDialog(true)}
  className="mt-4 w-full p-2 bg-blue-500 text-white rounded"
>
  + Create New Expert
</button>

<CreateExpertDialog
  isOpen={showCreateDialog}
  onClose={() => setShowCreateDialog(false)}
/>
```

### 3. Extension에서 Expert 생성 처리
**파일**: `vscode-extension/src/CltEditorProvider.ts`

```typescript
private async handleCreateExpert(expertId: string) {
  const expertsDir = path.join(this.dbDir, 'experts', expertId);
  const expertFile = path.join(expertsDir, 'EXPERT.md');

  // 폴더 및 파일 생성
  await fs.promises.mkdir(expertsDir, { recursive: true });

  const template = `# ${expertId}

## Role
TODO: Define the expert's role

## Tech Stack
- Language:
- Framework:

## Coding Rules
TODO: Define coding conventions

## Best Practices
TODO: Define best practices
`;

  await fs.promises.writeFile(expertFile, template);

  // DB에 등록
  const now = new Date().toISOString();
  this.db.run(`
    INSERT INTO experts (id, name, version, domain, language, framework, path, description, content, content_hash, status, created_at, updated_at)
    VALUES (?, ?, '1.0.0', '', '', '', ?, '', ?, '', 'active', ?, ?)
  `, [expertId, expertId, expertFile, template, now, now]);

  // 파일 열기
  const uri = vscode.Uri.file(expertFile);
  await vscode.window.showTextDocument(uri);

  // 동기화
  this.syncToWebview();
}
```

## 완료 기준
- [ ] CreateExpertDialog 컴포넌트 생성
- [ ] Expert ID 유효성 검증
- [ ] Extension에서 폴더/파일 생성
- [ ] DB에 Expert 등록
- [ ] 생성 후 파일 열기
