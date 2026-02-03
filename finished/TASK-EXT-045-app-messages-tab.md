# TASK-EXT-045: App.tsx Messages Tab 추가

## 목표
App.tsx에 Messages 탭 버튼 및 패널 추가

## 작업 내용

### 1. webview-ui/src/App.tsx 수정

#### Import 추가
```tsx
import { MessagesPanel } from './components/MessagesPanel';
```

#### view 상태 타입 수정
```tsx
const [view, setView] = useState<'project' | 'features' | 'tasks' | 'experts' | 'messages'>('project');
```

#### Messages 탭 버튼 추가 (Experts 버튼 다음)
```tsx
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
```

#### Main Content에 MessagesPanel 추가
```tsx
{view === 'messages' && <MessagesPanel />}
```

## 완료 조건
- [ ] MessagesPanel import
- [ ] view 타입에 'messages' 추가
- [ ] Messages 탭 버튼 추가
- [ ] MessagesPanel 렌더링
