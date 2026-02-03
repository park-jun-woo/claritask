# TASK-EXT-027: ExpertCard 컴포넌트

## 개요
Expert 상세 정보를 표시하는 카드 컴포넌트 (마크다운 렌더링 포함)

## 배경
- **스펙**: specs/VSCode/07-ExpertsTab.md
- **현재 상태**: ExpertCard 컴포넌트 없음

## 작업 내용

### 1. 의존성 추가
**파일**: `vscode-extension/webview-ui/package.json`

```json
{
  "dependencies": {
    "react-markdown": "^9.0.0",
    "remark-gfm": "^4.0.0"
  }
}
```

### 2. ExpertCard 컴포넌트 생성
**파일**: `vscode-extension/webview-ui/src/components/ExpertCard.tsx`

```tsx
import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useStore } from '../store';
import { assignExpert, unassignExpert, openExpertFile } from '../vscode';

interface ExpertCardProps {
  expertId: string;
}

const ExpertCard: React.FC<ExpertCardProps> = ({ expertId }) => {
  const { experts } = useStore();
  const expert = experts.find(e => e.id === expertId);

  if (!expert) return null;

  const handleAssign = () => {
    if (expert.assigned) {
      unassignExpert(expert.id);
    } else {
      assignExpert(expert.id);
    }
  };

  const handleOpenFile = () => {
    openExpertFile(expert.id);
  };

  return (
    <div className="space-y-4">
      {/* 헤더 */}
      <div className="flex justify-between items-start">
        <div>
          <h2 className="text-xl font-bold">{expert.name}</h2>
          <p className="text-gray-500">{expert.domain}</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleAssign}
            className={`px-3 py-1 rounded ${
              expert.assigned
                ? 'bg-red-100 text-red-700 hover:bg-red-200'
                : 'bg-green-100 text-green-700 hover:bg-green-200'
            }`}
          >
            {expert.assigned ? 'Unassign' : 'Assign'}
          </button>
          <button
            onClick={handleOpenFile}
            className="px-3 py-1 bg-gray-100 rounded hover:bg-gray-200"
          >
            Open File
          </button>
        </div>
      </div>

      {/* 메타 정보 */}
      <div className="grid grid-cols-2 gap-2 text-sm">
        <div><span className="text-gray-500">Language:</span> {expert.language}</div>
        <div><span className="text-gray-500">Framework:</span> {expert.framework}</div>
        <div><span className="text-gray-500">Version:</span> {expert.version}</div>
        <div><span className="text-gray-500">Path:</span> {expert.path}</div>
      </div>

      {/* 마크다운 내용 */}
      <div className="border-t pt-4">
        <h3 className="font-bold mb-2">Content</h3>
        <div className="prose prose-sm max-w-none">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>
            {expert.content}
          </ReactMarkdown>
        </div>
      </div>
    </div>
  );
};

export default ExpertCard;
```

### 3. CSS 스타일 (prose 클래스용)
**파일**: `vscode-extension/webview-ui/src/index.css`

```css
/* Markdown 스타일 */
.prose h1 { font-size: 1.5em; font-weight: bold; margin-top: 1em; }
.prose h2 { font-size: 1.25em; font-weight: bold; margin-top: 0.75em; }
.prose h3 { font-size: 1.1em; font-weight: bold; margin-top: 0.5em; }
.prose p { margin-top: 0.5em; }
.prose ul { list-style-type: disc; margin-left: 1.5em; }
.prose ol { list-style-type: decimal; margin-left: 1.5em; }
.prose code { background: var(--vscode-textCodeBlock-background); padding: 0.1em 0.3em; border-radius: 3px; }
.prose pre { background: var(--vscode-textCodeBlock-background); padding: 1em; border-radius: 5px; overflow-x: auto; }
```

## 완료 기준
- [ ] react-markdown 의존성 설치
- [ ] ExpertCard 컴포넌트 생성
- [ ] Assign/Unassign 버튼 동작
- [ ] Open File 버튼 동작
- [ ] 마크다운 렌더링 스타일 적용
