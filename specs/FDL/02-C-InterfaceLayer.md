# ClariSpec: INTERFACE LAYER (api)

> **버전**: v0.0.1

## 개요

INTERFACE LAYER는 HTTP API 계약을 정의합니다. 이 계층은 클라이언트(웹, 모바일, 외부 시스템)와의 통신 인터페이스를 명세하며, LOGIC LAYER의 서비스 함수와 연결(Wiring)됩니다.

---

## 기본 구조

```yaml
api:
  - path: <http_path>           # 엔드포인트 경로
    method: <HTTP_METHOD>       # HTTP 메서드
    summary: <description>      # API 설명
    use: service.<function>     # [Wiring] 서비스 연결
    auth: <auth_type>           # 인증 요구사항
    tags: [<tags>]              # API 그룹 태그
    request:
      params: { ... }           # URL 경로 파라미터
      query: { ... }            # 쿼리 스트링 파라미터
      headers: { ... }          # 요청 헤더
      body: { ... }             # 요청 본문
    response:
      <status_code>: <schema>   # 응답 스키마
```

---

## HTTP 메서드

| 메서드 | 용도 | 멱등성 | 안전성 | 예시 |
|--------|------|--------|--------|------|
| `GET` | 리소스 조회 | O | O | `GET /users/123` |
| `POST` | 리소스 생성 | X | X | `POST /users` |
| `PUT` | 리소스 전체 수정 | O | X | `PUT /users/123` |
| `PATCH` | 리소스 부분 수정 | O | X | `PATCH /users/123` |
| `DELETE` | 리소스 삭제 | O | X | `DELETE /users/123` |
| `HEAD` | 헤더만 조회 | O | O | `HEAD /users/123` |
| `OPTIONS` | 허용 메서드 조회 | O | O | `OPTIONS /users` |

---

## 경로 설계

### 경로 파라미터

```yaml
api:
  # 기본 경로 파라미터
  - path: /users/{userId}
    method: GET
    request:
      params:
        userId: uuid (required)

  # 중첩 리소스
  - path: /users/{userId}/posts/{postId}
    method: GET
    request:
      params:
        userId: uuid (required)
        postId: uuid (required)

  # 버전 접두사
  - path: /v1/users
    method: GET
```

### 경로 컨벤션

| 패턴 | 설명 | 예시 |
|------|------|------|
| 복수형 명사 | 컬렉션 리소스 | `/users`, `/posts` |
| 단수 ID | 개별 리소스 | `/users/{userId}` |
| 중첩 리소스 | 관계 표현 | `/users/{userId}/posts` |
| 액션 | 비 CRUD 작업 | `/users/{userId}/activate` |
| kebab-case | 경로명 규칙 | `/user-profiles` |

---

## 요청 정의

### Query Parameters

```yaml
api:
  - path: /posts
    method: GET
    request:
      query:
        # 필터링
        status: enum(all,published,draft) (default: "all")
        authorId: uuid (optional)

        # 페이지네이션
        page: int (default: 1, min: 1)
        pageSize: int (default: 20, min: 1, max: 100)

        # 정렬
        sortBy: enum(created_at,updated_at,title) (default: "created_at")
        sortOrder: enum(asc,desc) (default: "desc")

        # 검색
        search: string (optional, maxLength: 100)

        # 날짜 필터
        createdAfter: datetime (optional)
        createdBefore: datetime (optional)
```

### Request Body

```yaml
api:
  - path: /users
    method: POST
    request:
      body:
        # 필수 필드
        email: string (required, format: email)
        password: string (required, minLength: 8)
        name: string (required)

        # 선택 필드
        bio: string (optional, maxLength: 500)
        avatarUrl: string (optional, format: url)

        # 중첩 객체
        profile:
          timezone: string (optional)
          language: enum(ko,en,ja) (default: "ko")

        # 배열
        tags: Array<string> (optional, maxItems: 10)
```

### Request Headers

```yaml
api:
  - path: /files/upload
    method: POST
    request:
      headers:
        Content-Type: "multipart/form-data"
        X-Request-ID: string (optional)
        X-Idempotency-Key: string (optional)  # 멱등성 키
```

---

## 응답 정의

### 상태 코드

| 코드 | 의미 | 사용 상황 |
|------|------|----------|
| `200` | OK | 성공 (조회, 수정) |
| `201` | Created | 리소스 생성 성공 |
| `204` | No Content | 성공, 응답 본문 없음 (삭제) |
| `400` | Bad Request | 잘못된 요청 (검증 실패) |
| `401` | Unauthorized | 인증 필요 |
| `403` | Forbidden | 권한 없음 |
| `404` | Not Found | 리소스 없음 |
| `409` | Conflict | 충돌 (중복 등) |
| `422` | Unprocessable Entity | 처리 불가 (비즈니스 규칙 위반) |
| `429` | Too Many Requests | 요청 횟수 초과 |
| `500` | Internal Server Error | 서버 오류 |

### 응답 스키마

```yaml
api:
  - path: /users
    method: POST
    response:
      # 성공 응답
      201:
        id: uuid
        email: string
        name: string
        createdAt: datetime

      # 에러 응답
      400:
        error:
          code: string
          message: string
          details: Array<{ field: string, message: string }>

      409:
        error:
          code: "EMAIL_ALREADY_EXISTS"
          message: string
```

### 페이지네이션 응답

```yaml
api:
  - path: /posts
    method: GET
    response:
      200:
        data: Array<Post>
        pagination:
          total: int
          page: int
          pageSize: int
          totalPages: int
          hasNext: boolean
          hasPrev: boolean
```

### Envelope 패턴

```yaml
# 모든 응답을 일관된 형식으로 감싸기
response:
  200:
    success: true
    data: <actual_data>
    meta:
      requestId: string
      timestamp: datetime

  4xx/5xx:
    success: false
    error:
      code: string
      message: string
      details: object (optional)
```

---

## 인증 및 권한

### auth 옵션

```yaml
api:
  # 인증 필수
  - path: /users/me
    method: GET
    auth: required

  # 인증 선택 (미인증 시 제한된 데이터)
  - path: /posts
    method: GET
    auth: optional

  # 인증 불필요 (공개 API)
  - path: /health
    method: GET
    auth: none

  # 특정 역할 필요
  - path: /admin/users
    method: GET
    auth: required
    roles: [admin]

  # API 키 인증
  - path: /webhooks/receive
    method: POST
    auth: apiKey
```

---

## Rate Limiting

```yaml
api:
  - path: /auth/login
    method: POST
    rateLimit:
      limit: 5           # 최대 요청 수
      window: 1m         # 시간 창
      by: ip             # 기준 (ip, user, apiKey)

  - path: /posts
    method: GET
    rateLimit:
      limit: 100
      window: 1h
      by: user
```

---

## 파일 업로드

```yaml
api:
  - path: /files/upload
    method: POST
    summary: 파일 업로드
    use: service.uploadFile
    request:
      headers:
        Content-Type: "multipart/form-data"
      body:
        file: file (required)
        type: enum(image,document) (required)
    constraints:
      maxSize: 10MB
      allowedTypes: [image/jpeg, image/png, application/pdf]
    response:
      201:
        id: uuid
        url: string
        size: int
        mimeType: string
```

---

## Wiring (서비스 연결)

### 기본 Wiring

```yaml
api:
  - path: /users
    method: POST
    use: service.createUser    # 서비스 함수 연결
```

### 파라미터 매핑

```yaml
api:
  - path: /users/{userId}
    method: PUT
    use: service.updateUser
    mapping:
      # 경로 파라미터 → 서비스 입력
      params.userId -> input.id
      # 요청 본문 → 서비스 입력
      body.* -> input.*
      # 인증된 사용자 → 서비스 입력
      auth.userId -> input.updatedBy
```

### 응답 변환

```yaml
api:
  - path: /users/{userId}
    method: GET
    use: service.getUser
    transform:
      # 민감 정보 제외
      exclude: [password_hash, internal_notes]
      # 필드 이름 변환
      rename:
        created_at: createdAt
        updated_at: updatedAt
```

---

## API 그룹화

```yaml
api:
  # 태그로 그룹화
  - path: /users
    method: GET
    tags: [Users]

  - path: /users
    method: POST
    tags: [Users]

  - path: /posts
    method: GET
    tags: [Posts]

# 또는 그룹 블록 사용
apiGroups:
  - prefix: /admin
    auth: required
    roles: [admin]
    tags: [Admin]
    endpoints:
      - path: /users
        method: GET
        use: service.adminListUsers
      - path: /stats
        method: GET
        use: service.getAdminStats
```

---

## 전체 예시

```yaml
api:
  # 사용자 생성
  - path: /users
    method: POST
    summary: 새 사용자 등록
    tags: [Users, Auth]
    auth: none
    use: service.createUser
    rateLimit:
      limit: 10
      window: 1h
      by: ip
    request:
      body:
        email: string (required, format: email)
        password: string (required, minLength: 8)
        name: string (required, maxLength: 50)
        agreeToTerms: boolean (required)
    response:
      201:
        id: uuid
        email: string
        name: string
        createdAt: datetime
      400:
        error:
          code: string
          message: string
          details: Array<{ field: string, message: string }>
      409:
        error:
          code: "EMAIL_ALREADY_EXISTS"
          message: "이미 등록된 이메일입니다"

  # 게시글 목록 조회
  - path: /posts
    method: GET
    summary: 게시글 목록 조회 (페이지네이션)
    tags: [Posts]
    auth: optional
    use: service.listPosts
    request:
      query:
        page: int (default: 1, min: 1)
        pageSize: int (default: 20, min: 1, max: 100)
        status: enum(all,published,draft) (default: "published")
        authorId: uuid (optional)
        search: string (optional, maxLength: 100)
        sortBy: enum(created_at,updated_at,view_count) (default: "created_at")
        sortOrder: enum(asc,desc) (default: "desc")
    response:
      200:
        data:
          - id: uuid
            title: string
            excerpt: string
            author:
              id: uuid
              name: string
            status: string
            viewCount: int
            createdAt: datetime
            publishedAt: datetime
        pagination:
          total: int
          page: int
          pageSize: int
          totalPages: int
          hasNext: boolean
          hasPrev: boolean

  # 게시글 상세 조회
  - path: /posts/{postId}
    method: GET
    summary: 게시글 상세 조회
    tags: [Posts]
    auth: optional
    use: service.getPost
    request:
      params:
        postId: uuid (required)
    response:
      200:
        id: uuid
        title: string
        content: text
        author:
          id: uuid
          name: string
          avatarUrl: string
        status: string
        viewCount: int
        likeCount: int
        tags: Array<string>
        createdAt: datetime
        updatedAt: datetime
        publishedAt: datetime
      404:
        error:
          code: "POST_NOT_FOUND"
          message: "게시글을 찾을 수 없습니다"

  # 게시글 생성
  - path: /posts
    method: POST
    summary: 새 게시글 작성
    tags: [Posts]
    auth: required
    use: service.createPost
    request:
      body:
        title: string (required, maxLength: 200)
        content: text (required, minLength: 10)
        status: enum(draft,published) (default: "draft")
        tags: Array<string> (optional, maxItems: 10)
    response:
      201:
        id: uuid
        title: string
        status: string
        createdAt: datetime
      400:
        error:
          code: string
          message: string

  # 게시글 좋아요
  - path: /posts/{postId}/like
    method: POST
    summary: 게시글 좋아요
    tags: [Posts, Interactions]
    auth: required
    use: service.likePost
    request:
      params:
        postId: uuid (required)
    response:
      200:
        liked: boolean
        likeCount: int
      400:
        error:
          code: "SELF_LIKE_NOT_ALLOWED"
          message: "본인 글에는 좋아요할 수 없습니다"
      409:
        error:
          code: "ALREADY_LIKED"
          message: "이미 좋아요한 게시글입니다"

  # 파일 업로드
  - path: /files/upload
    method: POST
    summary: 파일 업로드
    tags: [Files]
    auth: required
    use: service.uploadFile
    request:
      headers:
        Content-Type: "multipart/form-data"
      body:
        file: file (required)
        purpose: enum(avatar,attachment,cover) (required)
    constraints:
      maxSize: 10MB
      allowedTypes: [image/jpeg, image/png, image/gif, application/pdf]
    response:
      201:
        id: uuid
        url: string
        size: int
        mimeType: string
        createdAt: datetime
      400:
        error:
          code: "INVALID_FILE_TYPE"
          message: "허용되지 않는 파일 형식입니다"
      413:
        error:
          code: "FILE_TOO_LARGE"
          message: "파일 크기가 10MB를 초과합니다"
```

---

## 네이밍 컨벤션

| 항목 | 규칙 | 예시 |
|------|------|------|
| 경로 | kebab-case, 복수형 | `/user-profiles`, `/blog-posts` |
| 경로 파라미터 | camelCase | `{userId}`, `{postId}` |
| 쿼리 파라미터 | camelCase | `pageSize`, `sortBy` |
| 요청/응답 필드 | camelCase | `createdAt`, `firstName` |

---

*ClariSpec FDL Specification v0.0.1 - 2026-02-03*
