# TASK-EXT-046: Messages Split Layout UI

## 목표
MessagesPanel을 FeatureList와 동일한 레이아웃으로 변경 (왼쪽 목록 1/3, 오른쪽 상세 2/3)

## 작업 내용

### 1. MessagesPanel.tsx 전면 수정

#### 기본 구조
```tsx
<div className="flex h-full">
  {/* Message List - 1/3 */}
  <div className="w-1/3 border-r border-vscode-border overflow-y-auto">
    {/* Header with + New button */}
    {/* Create Form (toggle) */}
    {/* Message List Items */}
  </div>

  {/* Message Detail - 2/3 */}
  <div className="flex-1 overflow-y-auto">
    {selectedMessageId ? <MessageDetail /> : <EmptyState />}
  </div>
</div>
```

#### MessageListItem 컴포넌트
- 상태 아이콘 (●/○)
- 내용 truncate
- 상태 + 시간
- Feature 이름 (있으면)

#### MessageDetail 컴포넌트
- Status 배지
- Feature 연결 (있으면)
- Content 전체
- Response (completed)
- Error (failed)
- Created/Completed 시간
- Delete 버튼

## 완료 조건
- [ ] 왼쪽/오른쪽 분할 레이아웃
- [ ] 메시지 목록 아이템 스타일링
- [ ] 메시지 상세 표시
- [ ] 선택 상태 하이라이트
