# Claribot Web UI 기획안

## 1. 개요

### 1.1 목적
Claribot의 모든 기능을 브라우저에서 시각적으로 관리하는 웹 대시보드. 텔레그램과 CLI로만 가능했던 조작을 직관적인 UI로 확장한다.

### 1.2 핵심 가치
- **시각화**: Task 트리, 프로젝트 현황 보드, 순회 진행률을 한눈에 파악
- **실시간**: 자동 폴링으로 Claude 실행 상태와 순회 진행 추적
- **편의성**: 인라인 편집, 원클릭 실행, 채팅 스타일 메시징으로 Task 관리

### 1.3 기술 스택
| 구분 | 선택 | 이유 |
|------|------|------|
| **프레임워크** | React + TypeScript | 컴포넌트 기반, 타입 안전성 |
| **빌드** | Vite | 빠른 HMR, 간결한 설정 |
| **UI 라이브러리** | shadcn/ui + Tailwind CSS | 커스터마이징 자유도, 경량 |
| **상태 관리** | TanStack Query | 서버 상태 캐싱, 자동 갱신 |
| **라우팅** | React Router v7 | SPA 라우팅 |
| **아이콘** | Lucide React | shadcn/ui 기본 아이콘셋 |
| **마크다운** | react-markdown + remark-gfm | Spec/Plan/Report HTML 렌더링 (MarkdownRenderer 컴포넌트) |
| **QR 코드** | qrcode.react (QRCodeSVG) | TOTP 설정 QR 생성 |
| **YAML** | yaml (npm) | Settings 페이지 config YAML 파싱/직렬화 |
| **배포** | Go embed | 빌드 결과물을 claribot 바이너리에 내장 |

### 1.4 디렉토리 구조
```
claribot/
├── gui/                          # Web UI 소스코드
│   ├── src/
│   │   ├── components/           # 공통 UI 컴포넌트
│   │   │   ├── layout/           # Header, Sidebar, Layout
│   │   │   ├── ui/               # shadcn/ui 컴포넌트
│   │   │   ├── ProjectSelector.tsx  # 프로젝트 드롭다운 선택기
│   │   │   ├── ChatBubble.tsx       # 채팅 메시지 버블 컴포넌트
│   │   │   └── MarkdownRenderer.tsx # 마크다운→HTML 렌더러
│   │   ├── pages/                # 페이지별 컴포넌트
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Projects.tsx
│   │   │   ├── ProjectEdit.tsx
│   │   │   ├── Tasks.tsx
│   │   │   ├── Messages.tsx
│   │   │   ├── Schedules.tsx
│   │   │   ├── Specs.tsx
│   │   │   ├── Settings.tsx
│   │   │   ├── Login.tsx
│   │   │   └── Setup.tsx
│   │   ├── hooks/                # 커스텀 훅
│   │   │   ├── useClaribot.ts    # TanStack Query 훅 (전체 API)
│   │   │   └── useAuth.ts        # 인증 훅 (로그인, 로그아웃, 설정)
│   │   ├── api/                  # API 클라이언트
│   │   │   └── client.ts         # RESTful API 클라이언트 (apiGet/apiPost/apiPatch/apiPut/apiDelete)
│   │   ├── types/                # TypeScript 타입
│   │   │   └── index.ts          # 전체 타입 정의
│   │   └── App.tsx               # 라우팅 + 인증 가드
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
├── bot/
│   └── internal/
│       ├── handler/
│       │   └── restful.go        # RESTful API 라우터 + 핸들러
│       └── webui/                # Go embed + HTTP 핸들러
│           ├── webui.go          # embed.FS, 정적 파일 서빙
│           └── dist/             # 빌드 결과물 (gitignore)
```

---

## 2. API 연동 설계

### 2.1 RESTful API

웹 UI는 RESTful API 엔드포인트를 통해 claribot 백엔드와 통신한다. 모든 엔드포인트는 `/api/` 접두사를 사용한다.

```typescript
// api/client.ts - 인증 쿠키 지원 HTTP 헬퍼 함수들
async function apiGet<T>(path: string): Promise<T> {
  const res = await fetch(`/api${path}`, { credentials: 'include' });
  if (!res.ok) throw new Error(`API error: ${res.status} ${res.statusText}`);
  return res.json();
}

async function apiPost<T>(path: string, body?: unknown): Promise<T> { ... }
async function apiPatch<T>(path: string, body: unknown): Promise<T> { ... }
async function apiPut<T>(path: string, body: unknown): Promise<T> { ... }
async function apiDelete<T>(path: string): Promise<T> { ... }
```

### 2.2 API 엔드포인트

| 엔드포인트 | 메서드 | 용도 |
|-----------|--------|------|
| `/api/health` | GET | 서비스 헬스체크 (버전, 가동시간) |
| `/api/status` | GET | Claude 상태 + 순회 상태 + Task 통계 |
| `/api/usage` | GET | Claude Code 사용량 통계 (stats-cache.json 기반) |
| `/api/usage/refresh` | POST | PTY에서 실시간 사용량 데이터 갱신 |
| `/api/auth/setup` | POST | 초기 비밀번호 설정 + TOTP 검증 (통합) |
| `/api/auth/totp-setup` | GET | QR 코드 생성을 위한 TOTP URI 조회 |
| `/api/auth/login` | POST | 비밀번호 + TOTP 로그인 |
| `/api/auth/logout` | POST | 로그아웃 (세션 삭제) |
| `/api/auth/status` | GET | 인증 상태 확인 |
| `/api/projects` | GET | 프로젝트 목록 조회 |
| `/api/projects` | POST | 프로젝트 추가/생성 |
| `/api/projects/stats` | GET | 전체 프로젝트 Task 통계 |
| `/api/projects/{id}` | GET | 프로젝트 상세 조회 |
| `/api/projects/{id}` | PATCH | 프로젝트 설정 변경 |
| `/api/projects/{id}` | DELETE | 프로젝트 삭제 |
| `/api/projects/{id}/switch` | POST | 활성 프로젝트 전환 |
| `/api/tasks` | GET | Task 목록 조회 (`?tree=true`, `?parent_id=`, `?all=true` 지원) |
| `/api/tasks` | POST | 새 Task 추가 |
| `/api/tasks/plan-all` | POST | 전체 todo Task 계획 수립 |
| `/api/tasks/run-all` | POST | 전체 planned leaf Task 실행 |
| `/api/tasks/cycle` | POST | 전체 Task 순회 (Plan + Run) |
| `/api/tasks/stop` | POST | 활성 순회 중단 |
| `/api/tasks/{id}` | GET | Task 상세 조회 |
| `/api/tasks/{id}` | PATCH | Task 필드 수정 |
| `/api/tasks/{id}` | DELETE | Task 삭제 |
| `/api/tasks/{id}/plan` | POST | 단일 Task 계획 수립 |
| `/api/tasks/{id}/run` | POST | 단일 Task 실행 |
| `/api/messages` | GET | 메시지 목록 조회 |
| `/api/messages` | POST | 새 메시지 전송 |
| `/api/messages/status` | GET | 메시지 상태 요약 |
| `/api/messages/processing` | GET | 현재 처리 중인 메시지 |
| `/api/messages/{id}` | GET | 메시지 상세 조회 |
| `/api/schedules` | GET | 스케줄 목록 조회 |
| `/api/schedules` | POST | 스케줄 추가 |
| `/api/schedules/{id}` | GET | 스케줄 상세 조회 |
| `/api/schedules/{id}` | PATCH | 스케줄 필드 수정 |
| `/api/schedules/{id}` | DELETE | 스케줄 삭제 |
| `/api/schedules/{id}/enable` | POST | 스케줄 활성화 |
| `/api/schedules/{id}/disable` | POST | 스케줄 비활성화 |
| `/api/schedules/{id}/runs` | GET | 스케줄 실행 이력 조회 |
| `/api/schedule-runs/{runId}` | GET | 개별 실행 상세 조회 |
| `/api/specs` | GET | Spec 목록 조회 |
| `/api/specs` | POST | Spec 추가 |
| `/api/specs/{id}` | GET | Spec 상세 조회 |
| `/api/specs/{id}` | PATCH | Spec 수정 |
| `/api/specs/{id}` | DELETE | Spec 삭제 |
| `/api/configs` | GET | 전체 config 키-값 쌍 조회 |
| `/api/configs/{key}` | GET | 키별 config 값 조회 |
| `/api/configs/{key}` | PUT | config 값 설정 |
| `/api/configs/{key}` | DELETE | config 키 삭제 |
| `/api/config-yaml` | GET | config YAML 내용 조회 |
| `/api/config-yaml` | PUT | config YAML 내용 저장 |

### 2.3 데이터 갱신 전략

TanStack Query를 통한 자동 폴링 (상황별 간격 조정):

| 데이터 | 갱신 간격 | 조건 |
|--------|-----------|------|
| Status (Claude/Cycle) | 5초 / 15초 | 순회 실행 중 5초, 유휴 시 15초 |
| Tasks | 15초 | 항상 |
| Messages | 10초 | 항상 |
| 단일 Message | 5초 | 상세 조회 시 |
| Message Status | 5초 | 항상 |
| Project Stats | 30초 | 항상 |
| Health | 30초 | 항상 |
| Auth Status | 30초 | 항상 |

---

## 3. 인증

### 3.1 Setup 페이지 (`/setup`)

첫 접근 시 표시되는 멀티스텝 초기 설정 위저드.

```
┌─────────────────────────────────┐
│  Claribot Setup                 │
│                                 │
│  Step [1] ─ [2] ─ [3]          │
│  (진행률 바 인디케이터)           │
│                                 │
│  ── Step 1: 비밀번호 설정 ────── │
│  비밀번호:     [••••••••]       │
│  비밀번호 확인: [••••••••]       │
│                         [다음]  │
│                                 │
│  ── Step 2: TOTP 설정 ──────── │
│  인증 앱으로 QR 코드를           │
│  스캔하세요:                     │
│  ┌─────────┐                   │
│  │ [QR Code]│  (QRCodeSVG)     │
│  └─────────┘                   │
│  Google Authenticator 또는       │
│  다른 TOTP 앱                   │
│                         [다음]  │
│                                 │
│  ── Step 3: TOTP 검증 ──────── │
│  6자리 코드 입력:                │
│  [______]  (숫자만)             │
│                       [완료]    │
│  [QR 코드 다시 보기]             │
└─────────────────────────────────┘
```

**구현:**
- Step 1: `POST /api/auth/setup` (`{ password }`) → `{ totp_uri }` 반환
- Step 2: `qrcode.react`의 `QRCodeSVG`로 QR 코드 표시
- Step 3: `POST /api/auth/setup` (`{ password, totp_code }`) → 설정 완료
- 비밀번호 최소 4자
- TOTP 입력: 숫자만, 비숫자 자동 제거, 최대 6자
- 스텝 인디케이터: 3개 진행률 바 세그먼트

### 3.2 Login 페이지 (`/login`)

```
┌─────────────────────────────────┐
│  Claribot 로그인                 │
│                                 │
│  비밀번호:                       │
│  [••••••••]                     │
│                                 │
│  TOTP 코드:                     │
│  [123456]  (중앙 정렬, 넓은 자간) │
│                                 │
│  [에러 메시지]                   │
│                                 │
│                       [로그인]   │
└─────────────────────────────────┘
```

**기능:**
- `POST /api/auth/login`으로 비밀번호 + TOTP 6자리 코드 로그인
- 숫자만 입력 가능한 TOTP 입력 (비숫자 필터링, 최대 6자리)
- Enter 키 네비게이션: 비밀번호 → TOTP 포커스, TOTP → 제출
- 로그인 실패 시 에러 표시
- TOTP 입력 스타일: `text-center text-lg tracking-widest`

### 3.3 인증 라우팅 가드

`App.tsx`의 `AuthGuard` 컴포넌트로 구현:

```
App 시작
  │
  ├─ GET /api/auth/status
  │    ├─ 로딩 중 ──▶ 스피너 표시 (Loader2 animate-spin)
  │    ├─ 에러 ──▶ "서버에 연결할 수 없습니다" 메시지
  │    ├─ setup_completed = false ──▶ /setup으로 리다이렉트
  │    ├─ is_authenticated = false ──▶ /login으로 리다이렉트
  │    └─ is_authenticated = true ──▶ 메인 앱 렌더링 (Layout)
  │
  └─ 라우트:
       /setup ──▶ Setup (가드 없음)
       /login ──▶ Login (가드 없음)
       /      ──▶ AuthGuard → Layout → 자식 라우트
       /*     ──▶ /로 리다이렉트
```

- 보호된 라우트는 React Router outlet을 통해 `<Layout>` 내부에 네스팅
- Header의 로그아웃 버튼: `POST /api/auth/logout` 후 auth 쿼리 무효화
- 인증 훅: `useAuthStatus`, `useLogin`, `useLogout`, `useSetup`, `useSetupVerify` (`useAuth.ts`)

---

## 4. 페이지 구성

### 4.1 전체 레이아웃

```
데스크톱:
┌──────────────────────────────────────────────────┐
│  Header: [≡]모바일 로고 / ProjectSelector /       │
│          GlobalNav(데스크톱) / Claude 뱃지 /       │
│          연결 상태 / 로그아웃                       │
├──────────┬───────────────────────────────────────┤
│          │                                       │
│ Sidebar  │           Main Content                │
│ (220px)  │           (Outlet)                    │
│          │                                       │
│ [프로젝트│                                       │
│  카드]   │                                       │
│ Edit     │                                       │
│ Specs    │                                       │
│ Tasks    │                                       │
│          │                                       │
└──────────┴───────────────────────────────────────┘

모바일:
┌──────────────────────────────┐
│  Header: [≡] 로고 [뱃지]     │
├──────────────────────────────┤
│                              │
│        Main Content          │
│                              │
└──────────────────────────────┘
   ↓ 햄버거 드로어 열기
┌──────────┐
│ Sidebar  │
│ (오버레이)│
│ Global:  │
│ Dashboard│
│ Messages │
│ Projects │
│ Schedules│
│ Settings │
│ ──────── │
│ Project: │
│ Specs    │
│ Tasks    │
└──────────┘
```

**Header 구성 (Header.tsx):**
- 좌측: 햄버거 메뉴 버튼 (모바일 전용, 최소 44x44px) + Claribot 로고 (모바일: 아이콘, sm+: 텍스트)
- 중앙좌: ProjectSelector 드롭다운 (모바일: 축소, xs: 텍스트 숨김)
- 중앙: 전역 네비게이션 링크 (데스크톱만, md: 아이콘, lg: 아이콘+텍스트)
- 우측: Claude 상태 뱃지 (`X/Y`) + 연결 상태 뱃지 (모바일 숨김), 로그아웃 버튼
- 네비게이션: Dashboard, Messages, Projects, Schedules, Settings (전역); Specs, Tasks (프로젝트별, 모바일 드로어)

**Sidebar 구성 (Sidebar.tsx):**
- "Project" 섹션 헤더
- 현재 프로젝트 카드 (프로젝트 선택 시, GLOBAL 아닐 때):
  - 프로젝트명 + 폴더/회전 아이콘 (순회 실행 중)
  - 카테고리 뱃지
  - 상태별 카운트 뱃지 (todo, planned, done, failed)
  - 스택형 컬러 바 (녹/노/회/적)
  - 진행률 퍼센트 텍스트
- 네비게이션: Edit (`/projects/{id}/edit`로 동적 링크), Specs, Tasks
- 접기/펼치기 토글 버튼 (데스크톱)
- 모바일: 숨김 (햄버거 드로어 모드)

**ProjectSelector 컴포넌트 (ProjectSelector.tsx):**
- 드롭다운 트리거: 폴더 아이콘 + 현재 프로젝트명 (말줄임, xs에서 숨김) + 쉐브론
- 드롭다운 패널 (320px, 절대 위치):
  - 검색 입력 + 아이콘
  - 정렬: last_accessed/created_at/task_count 순환, asc/desc 토글
  - 카테고리 필터 버튼 (전체 + 동적 카테고리)
  - 최상단 GLOBAL 옵션
  - 프로젝트 목록: 핀 토글, 프로젝트 ID, 카테고리 뱃지, 설명, 인라인 카테고리 선택기 (호버 시)
  - 외부 클릭 감지로 닫기
  - ScrollArea (최대 높이 300px)

---

### 4.2 대시보드 (Dashboard)

**경로**: `/`

```
┌─────────────────────────────────────────────────────┐
│  Dashboard                                           │
├────────────┬────────────┬────────────┬──────────────┤
│ Claude     │ Cycle      │ Messages   │ Schedules    │
│ ● 2/10    │ ▶ Running   │ 3 처리중   │ 5 활성       │
│ Running   │ PlanAll    │ 47 완료    │ 8 전체       │
│           │ Task #12   │            │              │
│           │ 3m 24s     │            │              │
└────────────┴────────────┴────────────┴──────────────┘
│                                                      │
│  ── Recent Messages ──────────────────────────── [→] │
│  [done]  [cli]  로그인 버그 수정                       │
│  [processing] [telegram] 코드 리뷰                    │
│  [pending] [gui] 테스트 실행                          │
│                                                      │
├──────────────────────────────────────────────────────┤
│  ── Projects ────────────────────────────────────── │
│                                                      │
│  ┌─────────────────┐  ┌─────────────────┐          │
│  │ ↻ claribot      │  │ blog             │          │
│  │ [backend]       │  │ 개인 블로그       │          │
│  │ 12 todo 5 plan  │  │ 3 todo 2 done    │          │
│  │ ████████░░ 75%  │  │ ██████░░░░ 50%  │          │
│  │ Done/Task: 80/106│  │ Done/Task: 4/8   │          │
│  │ [Edit][Tasks]   │  │ [Edit][Tasks]    │          │
│  │      [Stop]     │  │      [Cycle]     │          │
│  └─────────────────┘  └─────────────────┘          │
│                                                      │
└──────────────────────────────────────────────────────┘
```

**구성 요소:**
1. **요약 카드 4개** (반응형 그리드: md 2열, lg 4열):
   - Claude: 사용/최대 수, Running/Idle 상태
   - Cycle: 상태 (idle/running/interrupted), 타입, 단계, 현재 Task ID, 경과 시간. 실행 중 아이콘 회전.
   - Messages: 처리 중 수, 완료 수
   - Schedules: 활성 수, 전체 수
2. **Recent Messages**: 최근 5개 메시지 (상태 뱃지, 소스 라벨, 말줄임 내용). Messages 페이지 이동 화살표.
3. **프로젝트 현황 보드**: 프로젝트별 카드 (반응형 1/2/3열):
   - 프로젝트명 + 순회 실행 중 회전 아이콘, 카테고리 뱃지
   - 설명 (말줄임)
   - 상태별 카운트 뱃지 (todo, split, planned, done, failed)
   - 스택형 상태 컬러 바 (녹/노/회/적)
   - 진행률 바 (done/leaf 비율 + 퍼센트)
   - 액션 버튼: Edit, Tasks, Cycle (실행 중이면 Stop)

**데이터 갱신**: TanStack Query 자동 폴링 (status: 5-15초, project stats: 30초)

---

### 4.3 프로젝트 관리 (Projects)

**경로**: `/projects`

**기능:**
- 프로젝트 카드 그리드 (반응형: 1/2/3열)
- 프로젝트 ID, 설명, 카테고리로 검색
- 정렬: last_accessed, created_at, task_count (순환 버튼 + asc/desc 토글)
- 동적 카테고리 필터 버튼
- 프로젝트 핀/언핀 (핀 우선 정렬, 호버 시 핀 아이콘 표시)
- 프로젝트별: 상태 카운트 뱃지, 스택형 컬러 바, 진행률 바 (done/leaf 비율)
- 액션 버튼: Edit (편집 페이지 이동), Tasks (전환 + 이동), Cycle/Stop
- 실행 중 프로젝트: 회전 아이콘 + Cycle 대신 Stop 버튼
- 추가 폼: 경로(`/` 포함) 또는 ID 입력. 설명 입력. 카테고리 선택(동적 생성).

**프로젝트 편집 페이지** (`/projects/:id/edit`):
- 프로젝트 목록으로 돌아가기 버튼
- 읽기 전용: 프로젝트 ID, 경로
- 편집 가능: 설명(textarea), 카테고리(버튼 그룹 + 동적 생성), 병렬 Claude 수 (1-10)
- 저장 버튼
- 위험 영역: 프로젝트 ID 입력하여 삭제 확인

---

### 4.4 Task 관리 (Tasks) - 핵심 페이지

**경로**: `/tasks`

이 페이지가 Claribot 웹 UI의 핵심이다. Task 트리 구조를 1:1 분할 패널 레이아웃으로 시각적으로 관리한다.

#### 4.4.1 상태 바

**기능:**
- **순회 상태 행** (유휴가 아닐 때 표시): Running/Interrupted 인디케이터, 타입, 단계 뱃지, 현재 Task ID, 완료/대상 수, 경과 시간
- **상태 카운트 행**: 클릭 가능한 상태 필터 버튼 (컬러 점 + 카운트). 클릭으로 필터, 재클릭으로 해제. 우측에 `done/leaf` 비율과 퍼센트.
- 각 상태 버튼: 컬러 점 (회/파/노/녹/적) + 상태명 + 카운트

#### 4.4.2 트리 뷰

**레이아웃:** 데스크톱 1:1 분할 (w-1/2씩), 모바일 전체 너비.

**툴바:**
- 뷰 토글: Tree / List (아이콘 버튼)
- 추가 버튼 (+)
- 일괄 액션 버튼: Plan, Run, Cycle, Stop (모바일: 텍스트 숨김, 아이콘만)
- 액션 상태 인디케이터 (대기 중 노란 바 + 스피너)

**트리 인터랙션:**
- 행 클릭: 우측 패널에 Task 상세 열기
- 노드 접기/펼치기 (쉐브론 버튼)
- 트리 들여쓰기: `depth * 12 + 8px` 패딩
- 상태 점 + `#id` + 제목
- 선택된 행 `bg-accent` 하이라이트

**정렬:** 최신순 (ID 내림차순).

#### 4.4.3 리스트 뷰

**데스크톱:** 테이블 (열: ID, Title, Status 점, Depth, Parent)
**모바일:** 카드 뷰 (ID + 상태 점 헤더, 제목, depth + parent 정보)
**정렬:** 최신순 (ID 내림차순).

#### 4.4.4 Task 상세 패널 (우측 패널, 1:1 분할)

**기능:**
- 제목 + 상태 점 + depth 정보 + leaf 뱃지
- 액션 버튼: Plan, Run, Delete (confirm 다이얼로그)
- 탭 기반 Spec/Plan/Report 전환 (3열 그리드 TabsList)
- MarkdownRenderer로 마크다운→HTML 렌더링
- 인라인 편집: Edit 클릭 → textarea + Save/Cancel 버튼
- ScrollArea로 콘텐츠 오버플로 처리

**모바일:** 상세 패널이 전체 화면 오버레이(`fixed inset-0 z-50`)로 열림.

#### 4.4.5 Task 액션 버튼

| 버튼 | 동작 | API 호출 |
|------|------|---------|
| `+ Task 추가` | 새 Task 생성 폼 (제목, parent_id, spec) | `POST /api/tasks` |
| `Plan` (일괄) | 전체 todo Task 계획 수립 | `POST /api/tasks/plan-all` |
| `Run` (일괄) | 전체 planned leaf Task 실행 | `POST /api/tasks/run-all` |
| `Cycle` (일괄) | Plan + Run 자동 순회 | `POST /api/tasks/cycle` |
| `Stop` (일괄) | 활성 순회 중단 | `POST /api/tasks/stop` |
| `▶ Plan` (개별) | 단일 Task 계획 수립 | `POST /api/tasks/{id}/plan` |
| `▶ Run` (개별) | 단일 Task 실행 | `POST /api/tasks/{id}/run` |

---

### 4.5 메시지 (Messages) - 채팅 UI

**경로**: `/messages`

채팅 인터페이스로 리디자인, 1:1 분할 패널 레이아웃.

**기능:**
- **채팅 버블 UI** (ChatBubble 컴포넌트): 사용자 메시지 (우측 정렬, primary 배경, rounded-br-md) + 봇 응답 (좌측 정렬, muted 배경, rounded-bl-md)
- **날짜 그룹**: 날짜별 구분선으로 메시지 그룹화
- **소스 라벨**: 사용자 버블 위에 Telegram, CLI, Schedule 표시
- **봇 버블**: 상태 인디케이터 뱃지 (pending/processing/done/failed), 결과 요약 (1-2줄), "자세히보기" 링크
- **메시지 입력**: 하단 Textarea + Send 버튼, Ctrl+Enter/Cmd+Enter 단축키, 2행
- **낙관적 업데이트**: 임시 ID로 즉시 표시, 서버 확인 후 제거
- **상세 패널**: 우측 패널 (1:1 분할) - 전체 메시지 내용 + MarkdownRenderer 결과, 에러는 빨간 pre 블록
- **자동 스크롤**: 진입 시 최하단 즉시 이동 (애니메이션 없음), 이후 smooth
- **메시지 정렬**: created_at 오름차순 (채팅 흐름용)

**모바일:** 단일 패널 뷰, 상태 토글로 채팅/상세 전환. 상세에 Back 버튼.

---

### 4.6 Specs

**경로**: `/specs`

**기능:**
- **듀얼 패널 레이아웃**: 목록 (1/3) + 상세 (2/3) 데스크톱
- **목록 패널**: 카드 기반 레이아웃 (ID, 상태 뱃지, 제목, 우선순위, 날짜, 내용 미리보기 2줄)
- **Spec 추가 폼**: 제목 입력 + 마크다운 내용 textarea
- **검색**: 제목과 내용으로 필터링
- **상태 필터**: all, draft, review, approved, deprecated 버튼 (카운트 표시)
- **상세 패널**: 제목 인라인 편집, 상태 변경 버튼 그룹, 우선순위 [1-5] 버튼, Preview/Edit 토글
- **삭제**: `confirm()` 다이얼로그
- ID 오름차순 정렬

**상태 뱃지:** Draft(회색), Review(노랑), Approved(녹색), Deprecated(빨강)

**모바일:** 전체 화면 오버레이(`fixed inset-0 z-50`).

---

### 4.7 스케줄 (Schedules)

**경로**: `/schedules`

**스케줄 추가 폼 (카드):**
- Cron 표현식 입력 + 읽기 쉬운 미리보기
- 타입 선택: `<select>` (Claude AI / Bash Command)
- 메시지/명령어 textarea (타입에 따라 placeholder 변경)
- 프로젝트 선택 드롭다운 (전역 + 모든 프로젝트)
- 1회 실행 체크박스

**기능:**
- 스케줄 카드 목록 + 뱃지: ON/OFF, 타입 (Claude/Bash), run_once, 프로젝트
- 세로 액션 버튼 (우측): 활성화/비활성화 토글, 이력, 삭제
- 실행 이력 뷰어 (스케줄별 History 버튼으로 확장)
- 이력: 실행 ID, 타입 아이콘, 상태 뱃지, 타임스탬프, 결과/에러 (pre 블록)
- Cron 표현식 `<code>` 스타일 + 읽기 쉬운 설명
- Last run / Next run 타임스탬프
- 모바일: 버튼 세로 정렬, cron 스크롤 가능

---

### 4.8 설정 (Settings)

**경로**: `/settings`

**기능:**
- 시스템 정보: 버전 (`/api/health`), 가동시간 (일/시/분), DB 경로, 연결 상태 뱃지
- Config YAML 에디터 (섹션별 아이콘):
  - Service (호스트, 포트)
  - Telegram (토큰 password, admin_chat_id, allowed_users 콤마 구분)
  - Claude (최대 동시 실행, 유휴 타임아웃, 최대 타임아웃)
  - Project (기본 경로)
  - Pagination (페이지 크기)
  - Log (레벨 드롭다운, 파일 경로)
- `GET/PUT /api/config-yaml`로 실시간 설정 저장/로드
- `yaml` npm 패키지로 YAML 파싱
- 스마트 저장: 기본값이 아닌 값만 저장된 YAML에 포함
- 저장 후 성공/에러 피드백 메시지

---

## 5. 모바일 반응형

모든 페이지와 컴포넌트에 걸쳐 포괄적인 모바일 최적화가 구현되었다.

### 5.1 브레이크포인트
| 브레이크포인트 | 너비 | 용도 |
|------------|-------|------|
| `sm` | 640px | 소규모 조정 |
| `md` | 768px | 태블릿 레이아웃 전환 |
| `lg` | 1024px | 전체 데스크톱 레이아웃 |

### 5.2 컴포넌트별 모바일 최적화

| 컴포넌트 | 모바일 동작 |
|---------|------------|
| **Header** | 햄버거 메뉴 (최소 44x44px), 네비 링크 대체, Claude/연결 뱃지 숨김, ProjectSelector 아이콘만 |
| **Sidebar** | 기본 숨김, Sheet 오버레이 드로어, 44px 터치 타겟 네비 |
| **Layout** | 패딩 축소 (`p-3 sm:p-4 md:p-6`) |
| **Dashboard** | 그리드 2열(md) → 1열 축소 |
| **Tasks** | 상세 패널 전체 화면 오버레이; 트리 들여쓰기 축소(`depth*12`); 테이블→카드 뷰; 툴바 텍스트 숨김(아이콘만); `flex-1 min-w-0` 버튼 |
| **Messages** | 단일 패널 모드, 채팅/상세 상태 토글, Back 버튼 |
| **Specs** | 상세 뷰 전체 화면 오버레이, 카드 기반 목록 |
| **Schedules** | 버튼 세로 정렬, cron 스크롤 |
| **Projects** | 그리드 1열 축소 |
| **페이지 제목** | 반응형 폰트 (`text-2xl md:text-3xl`) |

### 5.3 터치 타겟
- 모든 인터랙티브 요소 최소 44x44px 보장
- 버튼: `min-h-[44px]`, `min-w-[44px]`
- 입력, 뱃지, 네비게이션 링크 터치용 크기
- Sidebar 네비: `py-3` 적정 높이

---

## 6. 공통 UI 컴포넌트

### 6.1 MarkdownRenderer

`react-markdown` + `remark-gfm` 플러그인으로 마크다운→HTML 렌더링. `markdown-body` 클래스 div로 래핑.

사용 위치:
- Task Spec/Plan/Report 탭
- Message 결과
- Spec 내용 미리보기

### 6.2 ChatBubble

재사용 채팅 버블 컴포넌트. Props: type (user/bot), content, status, source, result, time, onDetailClick, isSelected.

- 사용자 버블: 우측 정렬, primary 배경, rounded-br-md
- 봇 버블: 좌측 정렬, muted 배경, 상태 뱃지, 결과 요약 (2줄), 상세 링크
- 사용자 버블 위에 소스 라벨 (Telegram/CLI/Schedule)
- 버블 아래 타임스탬프

### 6.3 확인 다이얼로그

위험한 작업 (Task 삭제, 프로젝트 삭제, 스케줄 삭제) 시 표시. 간단한 작업은 `confirm()`, 프로젝트 삭제는 ID 입력 확인 UI.

### 6.4 상태 뱃지

전체 페이지에서 일관된 컬러 코드 상태 인디케이터:

| 상태 | 색상 | 사용처 |
|------|------|-------|
| todo | 회색 | Task |
| planned | 노랑 | Task |
| split | 파랑 | Task |
| done | 녹색 | Task, Message, Schedule Run |
| failed | 빨강 | Task, Message, Schedule Run |
| pending | 회색 | Message |
| processing | 노랑 | Message |
| draft | 회색/Secondary | Spec |
| review | 노랑/Warning | Spec |
| approved | 녹색/Success | Spec |
| deprecated | 빨강/Destructive | Spec |

---

## 7. 배포 및 통합

### 7.1 Go embed 통합

웹 UI 빌드 결과물을 Go 바이너리에 내장하여 별도 웹서버 없이 동작한다.

```go
// bot/internal/webui/webui.go
package webui

import "embed"

//go:embed dist/*
var staticFiles embed.FS
```

**빌드 흐름:**
```
gui/ → npm run build → gui/dist/ → cp → bot/internal/webui/dist/ → go build (embed) → claribot 바이너리
```

### 7.2 HTTP 라우팅

```go
// RESTful 라우터 (bot/internal/handler/restful.go)
// 인증 엔드포인트 (미들웨어 없음)
POST /api/auth/setup
GET  /api/auth/totp-setup
POST /api/auth/login
POST /api/auth/logout
GET  /api/auth/status

// 보호된 엔드포인트 (인증 미들웨어)
GET  /api/health
GET  /api/status
GET  /api/usage
POST /api/usage/refresh
GET/PUT /api/config-yaml
GET/PUT/DELETE /api/configs/{key}
// ... 모든 리소스 엔드포인트 (projects, tasks, messages, schedules, specs)

// 정적 파일 서빙 (SPA 폴백)
GET  /*  → webui.Handler() → index.html
```

**SPA Fallback**: `/api`로 시작하지 않는 모든 요청은 `index.html`로 리다이렉트 (React Router 지원).

### 7.3 Makefile 추가

```makefile
build-gui:
	cd gui && npm install && npm run build
	rm -rf bot/internal/webui/dist
	cp -r gui/dist bot/internal/webui/dist

build: build-gui build-cli build-bot

dev-gui:
	cd gui && npm run dev
```

---

## 8. 구현 현황

### Phase 1: 기반 (MVP) ✅
1. ~~프로젝트 scaffolding (Vite + React + TypeScript + shadcn/ui)~~ ✅
2. ~~API 클라이언트 모듈 (RESTful API 통신)~~ ✅
3. ~~Layout 컴포넌트 (Header + Sidebar + Main)~~ ✅
4. ~~ProjectSelector (검색, 정렬, 카테고리 필터, 핀)~~ ✅
5. ~~Dashboard 페이지 (요약 카드 + 프로젝트 현황 보드)~~ ✅
6. ~~Go embed 통합 및 정적 파일 서빙~~ ✅

### Phase 2: 핵심 기능 ✅
7. ~~Projects 페이지 (CRUD + 검색 + 정렬 + 카테고리 + 핀)~~ ✅
8. ~~ProjectEdit 페이지 (설명, parallel, 카테고리 편집, 삭제)~~ ✅
9. ~~Tasks 페이지 - 리스트 뷰 (상태 필터, 최신순 정렬)~~ ✅
10. ~~Tasks 페이지 - 트리 뷰 (접기/펼치기, 상태 점)~~ ✅
11. ~~Task 상세 패널 (Spec/Plan/Report 탭, 마크다운 렌더링, 인라인 편집)~~ ✅
12. ~~Task 실행 버튼 (Plan/Run/Cycle/Stop)~~ ✅
13. ~~Task 상태 바 (상태 카운트 + 순회 진행률)~~ ✅
14. ~~Messages 페이지 (채팅 UI, 버블, 날짜 그룹, 낙관적 업데이트)~~ ✅

### Phase 3: 시각화 ✅
15. ~~Dashboard 프로젝트 현황 보드 (프로젝트별 진행률 바)~~ ✅
16. ~~순회 상태 표시 (단계, 진행률, 대상 수)~~ ✅

### Phase 4: 인증 ✅
17. ~~Setup 페이지 (멀티스텝: 비밀번호 → TOTP QR → 검증)~~ ✅
18. ~~Login 페이지 (비밀번호 + TOTP 6자리)~~ ✅
19. ~~인증 라우팅 가드 (App.tsx: setup 체크 → login 체크 → 렌더)~~ ✅
20. ~~로그아웃 기능 (Header 버튼)~~ ✅

### Phase 5: 확장 기능 ✅
21. ~~Schedules 페이지 (CRUD + 타입 선택: Claude/Bash + 실행 이력)~~ ✅
22. ~~Settings 페이지 (시스템 정보 + config YAML 에디터)~~ ✅
23. ~~Specs 페이지 (CRUD + 검색 + 상태 필터 + 우선순위 + 마크다운 에디터)~~ ✅

### Phase 6: 모바일 반응형 ✅
24. ~~Sidebar 햄버거 드로어~~ ✅
25. ~~Header 반응형 (뱃지 숨기기, ProjectSelector 축소)~~ ✅
26. ~~Layout 패딩 반응형~~ ✅
27. ~~Tasks 상세 패널 모바일 오버레이~~ ✅
28. ~~Tasks 테이블 → 카드 뷰~~ ✅
29. ~~Tasks 트리 들여쓰기 축소~~ ✅
30. ~~Tasks 툴바 버튼 wrap~~ ✅
31. ~~Schedules 카드 반응형~~ ✅
32. ~~페이지 제목 폰트 반응형~~ ✅
33. ~~터치 타겟 최소 44x44px~~ ✅

### Phase 7: 미구현
34. WebSocket 연동 (`/api/stream`) 실시간 업데이트
35. Claude 실행 로그 실시간 스트리밍
36. 다크 모드 토글 (CSS 변수 준비완료)
37. 키보드 단축키 (Task 탐색, 실행)

---

## 9. 화면 흐름도

```
[App 시작]
  │
  ├─ needs_setup ──▶ [Setup] ── 완료 ──▶ [Login]
  │
  ├─ 미인증 ──▶ [Login] ── 성공 ──▶ [Dashboard]
  │
  └─ 인증됨 ──▶ [Dashboard]
                     │
                     ├──▶ [Projects] ── Edit 클릭 ──▶ [ProjectEdit]
                     │         └── Tasks 클릭 ──▶ 전환 + [Tasks]
                     │
                     ├──▶ [Tasks]
                     │      ├── Tree/List 뷰 ── Task 클릭 ──▶ [Task 상세 패널]
                     │      ├── 상태 필터 ── 상태 점 클릭
                     │      └── Plan/Run/Cycle/Stop 버튼 ──▶ API 호출
                     │
                     ├──▶ [Messages]
                     │      ├── 메시지 전송 ──▶ 낙관적 + API
                     │      └── 상세 클릭 ──▶ [Message 상세]
                     │
                     ├──▶ [Specs]
                     │      ├── Spec 추가/편집 ──▶ 마크다운 에디터
                     │      ├── 상태/검색 필터
                     │      └── Spec 클릭 ──▶ [Spec 상세]
                     │
                     ├──▶ [Schedules]
                     │      ├── 스케줄 추가 ──▶ [추가 폼]
                     │      ├── ON/OFF 토글
                     │      └── 이력 보기 ──▶ [실행 이력]
                     │
                     └──▶ [Settings]
                            └── 설정 편집 ──▶ YAML 저장

[Header ProjectSelector] ── 프로젝트 선택 ──▶ 전환 + 전체 쿼리 무효화
[Header 로그아웃] ──▶ auth 무효화 ──▶ [Login]
[Sidebar Edit] ──▶ [ProjectEdit]
```

---

## 10. 타입 정의

GUI 전반에서 사용되는 주요 TypeScript 타입 (`gui/src/types/index.ts`):

| 타입 | 필드 | 사용처 |
|------|------|-------|
| `Project` | id, name, path, type, description, status, category, pinned, last_accessed, created_at, updated_at | Projects, ProjectSelector |
| `Task` | id, parent_id, title, spec, plan, report, status, error, is_leaf, depth, created_at, updated_at | Tasks |
| `Message` | id, project_id, content, source, status, result, error, created_at, completed_at | Messages |
| `Schedule` | id, project_id, cron_expr, message, type, enabled, run_once, last_run, next_run, created_at, updated_at | Schedules |
| `ScheduleRun` | id, schedule_id, status, result, error, started_at, completed_at | Schedule 이력 |
| `Spec` | id, title, content, status, priority, created_at, updated_at | Specs |
| `ClaudeStatus` | used, max, available | Status 폴링 |
| `CycleStatus` | status, type, project_id, started_at, current_task_id, active_workers, phase, target_total, completed, elapsed_sec | Dashboard, Tasks |
| `TaskStats` | total, leaf, todo, planned, split, done, failed | Dashboard, Sidebar |
| `ProjectStats` | project_id, project_name, project_description, stats (TaskStats & { in_progress }) | Dashboard |
| `StatusResponse` | success, message, data (ClaudeStatus), cycle_status, cycle_statuses[], task_stats | Status 폴링 |
| `PaginatedList<T>` | items, total, page, page_size, total_pages | 목록 API 응답 |
| `UsageData` | success, message, live?, updated_at? | Usage API (client.ts) |
