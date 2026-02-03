# TASK-EXT-026: ExpertsPanel 컴포넌트

## 개요
Experts 탭의 메인 패널 컴포넌트 구현

## 배경
- **스펙**: specs/VSCode/07-ExpertsTab.md
- **현재 상태**: Experts 탭 없음 (3개 탭만 존재)

## 작업 내용

### 1. App.tsx에 Experts 탭 추가
**파일**: `vscode-extension/webview-ui/src/App.tsx`

```tsx
import ExpertsPanel from './components/ExpertsPanel';

// 탭 목록에 Experts 추가
const tabs = ['Project', 'Features', 'Tasks', 'Experts'];

// 렌더링 부분
{activeTab === 'Experts' && <ExpertsPanel />}
```

### 2. ExpertsPanel 컴포넌트 생성
**파일**: `vscode-extension/webview-ui/src/components/ExpertsPanel.tsx`

```tsx
import React from 'react';
import { useStore } from '../store';
import ExpertCard from './ExpertCard';

const ExpertsPanel: React.FC = () => {
  const { experts, selectedExpertId, setSelectedExpertId } = useStore();

  const assignedExperts = experts.filter(e => e.assigned);
  const availableExperts = experts.filter(e => !e.assigned);

  return (
    <div className="flex h-full">
      {/* 좌측: Expert 목록 */}
      <div className="w-1/3 border-r overflow-y-auto p-4">
        <h3 className="font-bold mb-2">Assigned</h3>
        {assignedExperts.map(expert => (
          <div
            key={expert.id}
            className={`p-2 cursor-pointer rounded ${
              selectedExpertId === expert.id ? 'bg-blue-100' : 'hover:bg-gray-100'
            }`}
            onClick={() => setSelectedExpertId(expert.id)}
          >
            <div className="font-medium">{expert.name}</div>
            <div className="text-sm text-gray-500">{expert.domain}</div>
          </div>
        ))}

        <h3 className="font-bold mb-2 mt-4">Available</h3>
        {availableExperts.map(expert => (
          <div
            key={expert.id}
            className={`p-2 cursor-pointer rounded opacity-60 ${
              selectedExpertId === expert.id ? 'bg-blue-100' : 'hover:bg-gray-100'
            }`}
            onClick={() => setSelectedExpertId(expert.id)}
          >
            <div className="font-medium">{expert.name}</div>
            <div className="text-sm text-gray-500">{expert.domain}</div>
          </div>
        ))}

        <button className="mt-4 w-full p-2 bg-blue-500 text-white rounded">
          + Create New Expert
        </button>
      </div>

      {/* 우측: Expert 상세 */}
      <div className="w-2/3 p-4 overflow-y-auto">
        {selectedExpertId ? (
          <ExpertCard expertId={selectedExpertId} />
        ) : (
          <div className="text-gray-500 text-center mt-10">
            Select an expert to view details
          </div>
        )}
      </div>
    </div>
  );
};

export default ExpertsPanel;
```

## 완료 기준
- [ ] App.tsx에 Experts 탭 추가
- [ ] ExpertsPanel 컴포넌트 생성
- [ ] Assigned/Available 섹션 구분
- [ ] Expert 선택 기능
- [ ] Create New Expert 버튼
