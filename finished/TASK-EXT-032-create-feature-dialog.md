# TASK-EXT-032: Feature 생성 다이얼로그

## 목표
VSCode Webview에서 Feature 생성 UI 구현 (FDL + Task 통합)

## 변경 파일
- `vscode-extension/webview-ui/src/components/CreateFeatureDialog.tsx` (신규)
- `vscode-extension/webview-ui/src/components/FeatureList.tsx`

## 작업 내용

### 1. CreateFeatureDialog.tsx 생성
```typescript
import React, { useState } from 'react';

interface CreateFeatureDialogProps {
    open: boolean;
    onClose: () => void;
    onCreated: (featureId: number) => void;
}

export const CreateFeatureDialog: React.FC<CreateFeatureDialogProps> = ({
    open, onClose, onCreated
}) => {
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [fdl, setFdl] = useState('');
    const [generateTasks, setGenerateTasks] = useState(true);
    const [generateSkeleton, setGenerateSkeleton] = useState(false);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleCreate = async () => {
        setLoading(true);
        setError(null);

        vscode.postMessage({
            type: 'createFeature',
            data: {
                name,
                description,
                fdl: fdl || undefined,
                generateTasks,
                generateSkeleton
            }
        });
    };

    const handleValidateFDL = () => {
        // FDL 검증 로직
    };

    // ... 렌더링 코드
};
```

### 2. 다이얼로그 UI 구성
- Name 입력 필드 (snake_case 검증)
- Description 입력 필드
- FDL 에디터 (Monaco Editor 또는 textarea)
- Task 자동 생성 체크박스
- 스켈레톤 생성 체크박스
- FDL 검증 버튼
- Cancel / Create 버튼

### 3. FeatureList.tsx 수정
- [+ Add...] 버튼 클릭 시 다이얼로그 열기
- 생성 완료 시 목록 새로고침

### 4. 메시지 핸들링 (App.tsx 또는 hooks)
```typescript
useEffect(() => {
    window.addEventListener('message', (event) => {
        const message = event.data;
        if (message.type === 'cliResult' && message.command === 'feature.create') {
            if (message.success) {
                // 성공 처리
            } else {
                // 에러 표시
            }
        }
    });
}, []);
```

## 테스트
- 다이얼로그 열기/닫기
- 필수 필드 검증
- CLI 호출 및 응답 처리
- 에러 표시

## 관련 스펙
- specs/VSCode/05-FeaturesTab.md (v0.0.6)
