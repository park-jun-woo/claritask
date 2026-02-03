# ClariSpec: LOGIC LAYER (service)

> **버전**: v0.0.1

## 개요

LOGIC LAYER는 비즈니스 로직과 규칙을 정의합니다. 이 계층은 데이터 검증, 트랜잭션 처리, 이벤트 발행 등 애플리케이션의 핵심 로직을 담당합니다.

---

## 기본 구조

```yaml
service:
  - name: <FunctionName>      # 함수 이름 (camelCase)
    desc: <description>       # 함수 설명
    input: { <args> }         # 입력 파라미터
    output: <return_type>     # 반환 타입
    throws: [<error_codes>]   # 발생 가능한 에러 (선택)
    transaction: true         # 트랜잭션 필요 여부 (선택)
    steps:
      - <step_description>
```

---

## 입출력 정의

### Input 정의

```yaml
service:
  - name: createUser
    input:
      email: string (required)
      password: string (required, minLength: 8)
      name: string (required)
      role: enum(admin,user) (default: "user")
      metadata: json (optional)
```

### Input 검증 규칙

| 규칙 | 설명 | 예시 |
|------|------|------|
| `required` | 필수 값 | `email: string (required)` |
| `optional` | 선택 값 | `bio: string (optional)` |
| `default: <value>` | 기본값 | `role: string (default: "user")` |
| `minLength: <n>` | 최소 길이 | `password: string (minLength: 8)` |
| `maxLength: <n>` | 최대 길이 | `name: string (maxLength: 100)` |
| `min: <n>` | 최소값 | `age: int (min: 0)` |
| `max: <n>` | 최대값 | `quantity: int (max: 100)` |
| `pattern: <regex>` | 정규식 패턴 | `email: string (pattern: email)` |
| `enum(...)` | 허용 값 목록 | `status: enum(active,inactive)` |

### 내장 패턴

| 패턴명 | 설명 |
|--------|------|
| `email` | 이메일 형식 |
| `url` | URL 형식 |
| `uuid` | UUID 형식 |
| `phone` | 전화번호 형식 |
| `slug` | URL 슬러그 (영문, 숫자, 하이픈) |

### Output 정의

```yaml
service:
  - name: getUser
    output: User                    # 단일 모델 반환

  - name: listUsers
    output: Array<User>             # 배열 반환

  - name: listUsersWithPagination
    output:
      items: Array<User>
      total: int
      page: int
      pageSize: int

  - name: deleteUser
    output: void                    # 반환값 없음
```

---

## 스텝 타입

### 기본 스텝

| 스텝 | 설명 | 용도 |
|------|------|------|
| `validate:` | 검증 로직 | 입력값 검증, 비즈니스 규칙 검증 |
| `db:` | 데이터베이스 작업 | CRUD 연산 |
| `event:` | 이벤트 발행 | 비동기 처리, 알림 |
| `call:` | 외부 서비스 호출 | API 호출, 서비스 간 통신 |
| `cache:` | 캐시 작업 | 캐시 조회/저장/삭제 |
| `log:` | 로깅 | 감사 로그, 디버그 로그 |
| `transform:` | 데이터 변환 | 형식 변환, 가공 |
| `condition:` | 조건 분기 | if-else 로직 |
| `loop:` | 반복 처리 | 배열 순회 |
| `return:` | 반환값 설명 | 함수 결과 반환 |

### validate 스텝

```yaml
steps:
  # 기본 검증 (자연어)
  - validate: "이메일 형식이 올바른지 검증"

  # 조건부 검증
  - validate:
      condition: "email already exists"
      error: EMAIL_ALREADY_EXISTS
      message: "이미 등록된 이메일입니다"

  # 비즈니스 규칙 검증
  - validate:
      rule: "본인 글에만 수정 가능"
      check: "post.author_id == currentUser.id"
      error: PERMISSION_DENIED
```

### db 스텝

```yaml
steps:
  # 단순 쿼리 (자연어)
  - db: "INSERT INTO users (email, name) VALUES (...)"

  # 상세 쿼리
  - db:
      operation: insert
      table: users
      data: { email, name, role }
      returning: id

  # 조회 쿼리
  - db:
      operation: select
      table: posts
      where: "author_id = ? AND status = 'published'"
      orderBy: created_at DESC
      limit: 10

  # 트랜잭션 내 여러 쿼리
  - db:
      transaction: true
      queries:
        - insert: posts (title, content, author_id)
        - update: users SET post_count = post_count + 1 WHERE id = ?
```

### event 스텝

```yaml
steps:
  # 비동기 이벤트
  - event:
      name: UserCreated
      payload: { userId, email }
      async: true

  # 동기 이벤트 (웹훅 등)
  - event:
      name: PaymentCompleted
      payload: { orderId, amount }
      async: false

  # 알림 발송
  - event:
      type: notification
      channel: email
      template: welcome_email
      to: user.email
```

### call 스텝

```yaml
steps:
  # 내부 서비스 호출
  - call:
      service: notificationService.sendEmail
      args: { to: user.email, template: "welcome" }

  # 외부 API 호출
  - call:
      external: true
      url: "https://api.stripe.com/v1/charges"
      method: POST
      headers:
        Authorization: "Bearer ${STRIPE_KEY}"
      body: { amount, currency, source }
      timeout: 30s
      retry: 3
```

### cache 스텝

```yaml
steps:
  # 캐시 조회
  - cache:
      action: get
      key: "user:${userId}"
      onMiss: "DB에서 조회 후 캐시 저장"

  # 캐시 저장
  - cache:
      action: set
      key: "user:${userId}"
      value: user
      ttl: 1h

  # 캐시 무효화
  - cache:
      action: invalidate
      pattern: "user:${userId}:*"
```

### condition 스텝

```yaml
steps:
  - condition:
      if: "user.role == 'admin'"
      then:
        - db: "관리자용 전체 데이터 조회"
      else:
        - db: "일반 사용자용 제한된 데이터 조회"
```

### loop 스텝

```yaml
steps:
  - loop:
      over: users
      as: user
      do:
        - call: notificationService.send({ userId: user.id })
```

---

## 에러 처리

### throws 정의

```yaml
service:
  - name: createUser
    throws:
      - EMAIL_ALREADY_EXISTS
      - INVALID_PASSWORD_FORMAT
      - RATE_LIMIT_EXCEEDED
```

### 에러 정의 섹션

```yaml
errors:
  - code: EMAIL_ALREADY_EXISTS
    status: 409
    message: "이미 등록된 이메일입니다"

  - code: INVALID_PASSWORD_FORMAT
    status: 400
    message: "비밀번호는 8자 이상, 대소문자와 숫자를 포함해야 합니다"

  - code: PERMISSION_DENIED
    status: 403
    message: "접근 권한이 없습니다"
```

---

## 트랜잭션

```yaml
service:
  - name: transferMoney
    transaction: true           # 전체 스텝을 트랜잭션으로 묶음
    steps:
      - validate: "잔액 충분한지 확인"
      - db: "출금 계좌에서 금액 차감"
      - db: "입금 계좌에 금액 추가"
      - event: "송금 완료 알림"

  - name: complexOperation
    steps:
      - db: "일반 작업 (트랜잭션 없음)"
      - db:
          transaction: true     # 특정 스텝만 트랜잭션
          queries:
            - "쿼리 1"
            - "쿼리 2"
```

---

## CRUD 단축 표현

단순 CRUD의 경우 `steps: [CRUD Standard]`로 간소화할 수 있습니다.

```yaml
service:
  # 표준 CRUD
  - name: createPost
    steps: [CRUD Standard]

  - name: getPost
    steps: [CRUD Standard]

  - name: updatePost
    steps: [CRUD Standard]

  - name: deletePost
    steps: [CRUD Standard]

  # 또는 그룹으로 정의
  - name: PostCRUD
    type: crud
    model: Post
    operations: [create, read, update, delete, list]
```

---

## 권한 검사

```yaml
service:
  - name: updatePost
    auth: required              # 인증 필수
    roles: [admin, author]      # 허용 역할
    ownership: post.author_id   # 소유권 검사 필드
    steps:
      - validate: "게시글 존재 여부 확인"
      - db: "UPDATE posts SET ..."
```

### 권한 옵션

| 옵션 | 설명 |
|------|------|
| `auth: required` | 인증된 사용자만 접근 |
| `auth: optional` | 인증 선택 (미인증도 가능) |
| `roles: [...]` | 특정 역할만 접근 |
| `ownership: <field>` | 소유자만 접근 가능 |
| `permissions: [...]` | 특정 권한 필요 |

---

## 전체 예시

```yaml
service:
  # 사용자 생성
  - name: createUser
    desc: 새 사용자 계정 생성
    auth: optional
    input:
      email: string (required, pattern: email)
      password: string (required, minLength: 8, maxLength: 100)
      name: string (required, maxLength: 50)
    output: User
    throws: [EMAIL_ALREADY_EXISTS, INVALID_PASSWORD]
    transaction: true
    steps:
      - validate:
          condition: "이메일 중복 검사"
          error: EMAIL_ALREADY_EXISTS
      - validate:
          condition: "비밀번호 복잡성 검사"
          error: INVALID_PASSWORD
      - transform: "비밀번호 해시 생성"
      - db: "INSERT INTO users (email, password_hash, name)"
      - event:
          name: UserCreated
          async: true
      - return: "생성된 User 객체 (비밀번호 제외)"

  # 게시글 목록 조회 (페이지네이션)
  - name: listPosts
    desc: 게시글 목록 페이지네이션 조회
    input:
      page: int (default: 1, min: 1)
      pageSize: int (default: 20, min: 1, max: 100)
      status: enum(all,published,draft) (default: "published")
      authorId: uuid (optional)
    output:
      items: Array<Post>
      total: int
      page: int
      pageSize: int
    steps:
      - cache:
          action: get
          key: "posts:list:${page}:${pageSize}:${status}"
          onMiss: continue
      - db:
          operation: select
          table: posts
          where: "동적 필터 적용"
          orderBy: published_at DESC
          limit: pageSize
          offset: (page - 1) * pageSize
      - db:
          operation: count
          table: posts
          where: "동일 필터"
      - cache:
          action: set
          ttl: 5m
      - return: "페이지네이션된 결과"

  # 게시글 좋아요
  - name: likePost
    desc: 게시글에 좋아요 추가
    auth: required
    input:
      postId: uuid (required)
    output: { liked: boolean, likeCount: int }
    throws: [POST_NOT_FOUND, SELF_LIKE_NOT_ALLOWED, ALREADY_LIKED]
    transaction: true
    steps:
      - validate:
          condition: "게시글 존재 확인"
          error: POST_NOT_FOUND
      - validate:
          condition: "본인 글 좋아요 불가"
          check: "post.author_id != currentUser.id"
          error: SELF_LIKE_NOT_ALLOWED
      - validate:
          condition: "이미 좋아요한 경우"
          error: ALREADY_LIKED
      - db: "INSERT INTO likes (post_id, user_id)"
      - db: "UPDATE posts SET like_count = like_count + 1"
      - event:
          name: PostLiked
          payload: { postId, userId: currentUser.id }
          async: true
      - cache:
          action: invalidate
          key: "post:${postId}"
      - return: "{ liked: true, likeCount: 업데이트된 카운트 }"

errors:
  - code: EMAIL_ALREADY_EXISTS
    status: 409
    message: "이미 등록된 이메일입니다"

  - code: POST_NOT_FOUND
    status: 404
    message: "게시글을 찾을 수 없습니다"

  - code: SELF_LIKE_NOT_ALLOWED
    status: 400
    message: "본인 글에는 좋아요할 수 없습니다"

  - code: ALREADY_LIKED
    status: 409
    message: "이미 좋아요한 게시글입니다"
```

---

## 네이밍 컨벤션

| 항목 | 규칙 | 예시 |
|------|------|------|
| 함수명 | camelCase | `createUser`, `listPosts` |
| CRUD 함수 | 동사 + 모델명 | `create`, `get`, `update`, `delete`, `list` |
| 조회 함수 | `get` (단일), `list` (복수) | `getUser`, `listUsers` |
| 에러 코드 | UPPER_SNAKE_CASE | `EMAIL_ALREADY_EXISTS` |
| 이벤트명 | PascalCase | `UserCreated`, `PostLiked` |

---

*ClariSpec FDL Specification v0.0.1 - 2026-02-03*
