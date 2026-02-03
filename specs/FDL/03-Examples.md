# ClariSpec: 작성 예시

> **버전**: v0.0.1

## 예시: 댓글 작성 (Post Comment)

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

## 예시 분석

### 1. DATA LAYER

```yaml
models:
  - name: Comment
    table: comments
    fields:
      - id: uuid (pk)
      - content: text (required)
      - post_id: uuid (fk: posts.id)
      - user_id: uuid (fk: users.id)
      - created_at: datetime (default: now)
```

- `Comment` 모델은 `comments` 테이블에 매핑
- `post_id`, `user_id`는 외래키로 다른 테이블 참조
- `created_at`은 기본값으로 현재 시간

### 2. LOGIC LAYER

```yaml
service:
  - name: createComment
    desc: 댓글 생성 및 알림 발송
    input: { userId: uuid, postId: uuid, content: string }
    steps:
      - validate: "content 길이가 1자 이상 1000자 이하인지 검증"
      - db: "INSERT INTO comments (user_id, post_id, content)"
      - event: "게시글 작성자에게 알림 발송 (optional)"
      - return: "생성된 Comment 객체"
```

- 입력값 검증 → DB 저장 → 이벤트 발행 순서
- 자연어로 로직 설명 (LLM이 해석)

### 3. INTERFACE LAYER

```yaml
api:
  - path: /posts/{postId}/comments
    method: POST
    summary: 댓글 작성
    use: service.createComment # [Wiring]
```

- RESTful 경로 설계
- `use:` 필드로 서비스 함수와 연결 (Wiring)

### 4. PRESENTATION LAYER

```yaml
ui:
  - component: CommentForm
    type: Molecule
    view:
      - Button: "등록"
        action: API.POST /posts/{postId}/comments
        onSuccess: "새로고침 없이 comments 목록에 추가"
```

- 버튼 클릭 시 API 호출
- 성공 시 동작을 자연어로 설명

---

*ClariSpec FDL Specification v0.0.1 - 2026-02-03*
