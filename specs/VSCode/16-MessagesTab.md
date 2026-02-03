# VSCode Extension Messages Tab

> **현재 버전**: v0.0.1 ([변경이력](../HISTORY.md))

---

## 개요

Messages 탭은 사용자가 프로젝트 수정 요청을 전송하고 관리하는 UI를 제공합니다.
CLI의 `clari message send` 명령어를 호출하여 Claude Code와 연동됩니다.

---

## UI 레이아웃

```
┌─────────────────────────────────────────────────────────────────┐
│  [Project] [Messages] [Features] [Tasks] [Experts]              │
├─────────────────────────────────────────────────────────────────┤
│ Messages (3)        [+ New]  │  Message Detail                  │
├─────────────────────────────┤                                   │
│ ┌─────────────────────────┐ │  Status: pending                  │
│ │ ● 로그인 기능 수정 요청   │ │  Feature: user-auth               │
│ │   pending | 2분 전        │ │                                   │
│ └─────────────────────────┘ │  Content:                          │
│ ┌─────────────────────────┐ │  로그인 실패 시 에러 메시지를      │
│ │ ○ API 응답 형식 변경      │ │  더 구체적으로 표시해주세요.       │
│ │   completed | 1시간 전    │ │                                   │
│ └─────────────────────────┘ │  Response:                         │
│ ┌─────────────────────────┐ │  (처리 완료 후 응답 표시)          │
│ │ ○ 테스트 코드 추가        │ │                                   │
│ │   processing | 5분 전     │ │  Created: 2024-02-04 10:30:00     │
│ └─────────────────────────┘ │                                   │
│                             │  [Delete]                          │
└─────────────────────────────┴───────────────────────────────────┘
```

### 레이아웃 구조
- **왼쪽 패널 (1/3)**: 메시지 목록
- **오른쪽 패널 (2/3)**: 선택된 메시지 상세 정보

---

## 메시지 목록 (왼쪽)

### 헤더
- 제목: "Messages (N)" - N은 전체 메시지 수
- "+ New" 버튼: 새 메시지 작성 폼 토글

### 메시지 아이템
```
┌─────────────────────────────┐
│ ● 메시지 내용 (truncate)     │
│   status | 시간              │
│   Feature: feature-name     │  (optional)
└─────────────────────────────┘
```

- 상태 아이콘:
  - `●` (채워진 원): pending, processing
  - `○` (빈 원): completed, failed
- 상태별 배경색:
  - pending: yellow
  - processing: blue
  - completed: green
  - failed: red
- 선택된 아이템: 강조 배경색

---

## 메시지 상세 (오른쪽)

### 표시 정보
1. **Status**: 상태 배지
2. **Feature**: 연결된 Feature 이름 (있는 경우)
3. **Content**: 요청 내용 (전체 표시)
4. **Response**: 처리 응답 (completed 상태일 때)
5. **Error**: 에러 메시지 (failed 상태일 때)
6. **Created**: 생성 시간
7. **Completed**: 완료 시간 (있는 경우)

### 액션 버튼
- **Delete**: 메시지 삭제 (확인 모달)

---

## 새 메시지 작성 폼

"+ New" 버튼 클릭 시 목록 상단에 폼 표시:

```
┌─────────────────────────────┐
│ Feature (Optional)          │
│ [Select feature ▼]          │
│                             │
│ Message Content             │
│ ┌─────────────────────────┐ │
│ │ 수정 요청 내용 입력...    │ │
│ └─────────────────────────┘ │
│                             │
│ [Cancel] [Send]             │
└─────────────────────────────┘
```

### Send 동작
1. CLI 명령어 실행: `clari message send '<content>' [--feature <id>]`
2. TTY 세션 매니저를 통해 터미널에서 실행
3. 폼 닫기 및 목록 새로고침

---

## CLI 연동

### Send Message
```bash
# Feature 없이
clari message send '로그인 기능을 수정해주세요'

# Feature 지정
clari message send '로그인 기능을 수정해주세요' --feature 1
```

### 메시지 프로토콜

**WebView → Extension:**
```typescript
{ type: 'sendMessageCLI'; content: string; featureId?: number }
```

**Extension → WebView:**
```typescript
{ type: 'cliStarted'; command: 'message.send'; message: string }
```

---

## 상태별 스타일

| Status | 배지 색상 | 목록 아이콘 |
|--------|----------|-----------|
| pending | yellow | ● |
| processing | blue | ● |
| completed | green | ○ |
| failed | red | ○ |

---

*Claritask VSCode Extension Spec v0.0.1*
