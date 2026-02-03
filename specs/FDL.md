# ClariSpec: Feature Definition DSL

> **버전**: v0.0.1

## 변경이력
| 버전 | 날짜 | 내용 |
|------|------|------|
| v0.0.1 | 2026-02-03 | 최초 작성 |

## 개요
ClariSpec은 Claritask 시스템에서 하나의 기능(Feature)을 "Vertical Slice(수직적 격리)" 형태로 정의하기 위한 YAML 기반의 DSL(Domain Specific Language)입니다. 이 명세는 SQL, 백엔드 로직, API 인터페이스, 프론트엔드 UI를 통합적으로 기술합니다.

## 핵심 철학
1.  **Unified Source**: 데이터(DB)부터 화면(UI)까지 하나의 파일에 정의하여 정합성을 보장합니다.
2.  **Loose Syntax**: 엄격한 문법보다 가독성과 의미 전달을 우선합니다. (LLM 해석용)
3.  **Explicit Wiring**: UI가 어떤 API를 호출하고, API가 어떤 Service를 쓰는지 명시합니다.

---

## DSL 구조 (Schema)

```yaml
feature: <feature_name> (string)
description: <description> (string)

# 1. DATA LAYER (Schema & Models)
models:
  - name: <ModelName>
    table: <table_name>
    fields:
      - <name>: <type> [constraints...]

# 2. LOGIC LAYER (Service & Business Rules)
service:
  - name: <FunctionName>
    input: <args>
    output: <return_type>
    steps:
      - <step_description_or_pseudocode>
      - db: <db_operation>

# 3. INTERFACE LAYER (API Contract)
api:
  - path: <http_path>
    method: <GET|POST|PUT|DELETE>
    handler: <Service.FunctionName>  # Wiring Point
    request: <json_schema>
    response: <json_schema>

# 4. PRESENTATION LAYER (UI Components)
ui:
  - component: <ComponentName>
    type: <Page|Organism|Molecule>
    state:
      - <state_variable>
    view:
      - <element>: <label_or_content>
        bind: <state_variable>
        action: <API.path>  # Wiring Point

```

---

## 작성 예시: "댓글 작성 (Post Comment)"

아래는 실제 블로그 시스템의 댓글 작성 기능을 ClariSpec으로 정의한 예시입니다.

```yaml
feature: comment_system
description: 사용자가 게시글에 댓글을 작성하고 목록을 조회하는 기능

# [1] 데이터 모델 정의 (SQL/ORM 기준)
models:
  - name: Comment
    table: comments
    fields:
      - id: uuid (pk)
      - content: text (required)
      - post_id: uuid (fk: posts.id)
      - user_id: uuid (fk: users.id)
      - created_at: datetime (default: now)

# [2] 비즈니스 로직 정의 (Service 계층)
service:
  - name: createComment
    desc: 댓글 생성 및 알림 발송
    input: { userId: uuid, postId: uuid, content: string }
    steps:
      - validate: "content 길이가 1자 이상 1000자 이하인지 검증"
      - db: "INSERT INTO comments (user_id, post_id, content)"
      - event: "게시글 작성자에게 알림 발송 (optional)"
      - return: "생성된 Comment 객체"

  - name: listComments
    desc: 특정 게시글의 댓글 조회
    input: { postId: uuid }
    steps:
      - db: "SELECT * FROM comments WHERE post_id = ? ORDER BY created_at ASC"
      - enrich: "User 테이블 조인하여 작성자 닉네임 포함"

# [3] API 명세 (Controller/Router 계층)
api:
  - path: /posts/{postId}/comments
    method: POST
    summary: 댓글 작성
    use: service.createComment # [Wiring] 서비스 연결
    request:
      body: { content: string }
    response:
      201: { id: uuid, content: string, created_at: iso8601 }

  - path: /posts/{postId}/comments
    method: GET
    summary: 댓글 목록 조회
    use: service.listComments
    response:
      200: [ { id: uuid, content: string, user: { name: string } } ]

# [4] UI 명세 (React/Vue 컴포넌트 계층)
ui:
  - component: CommentSection
    type: Organism
    props: { postId: uuid }
    state:
      - comments: Array
      - newComment: string
    init:
      - call: API.GET /posts/{postId}/comments -> set comments

  - component: CommentForm
    type: Molecule
    parent: CommentSection
    view:
      - Textarea: "댓글을 입력하세요"
        bind: newComment
      - Button: "등록"
        action: API.POST /posts/{postId}/comments (body: { content: newComment })
        onSuccess: "새로고침 없이 comments 목록에 추가"

```

---

## 작성 가이드라인 (Rules)

1. **타입의 유연성**: `uuid`, `string`, `text`, `int` 등 일반적인 프로그래밍 용어를 사용하십시오. 특정 DB 문법(예: `VARCHAR(255)`)에 얽매이지 않아도 됩니다.
2. **연결(Wiring) 명시**:
* `api` 섹션에서는 반드시 `service`의 어떤 함수를 쓰는지 명시해야 합니다.
* `ui` 섹션에서는 버튼이나 폼이 `api`의 어떤 엔드포인트와 통신하는지 명시해야 합니다.


3. **자연어 혼용**: 로직(steps)이나 검증 조건이 복잡할 경우, 억지로 코드로 표현하지 말고 자연어(한국어/영어)로 명확히 서술하십시오. (예: `validate: "본인 글에는 좋아요 불가"`)
4. **생략 가능**: 단순 CRUD의 경우 `service` 스텝을 `steps: [CRUD Standard]`로 간소화할 수 있습니다.