# ClariSpec: 작성 가이드라인

> **버전**: v0.0.1

## 기본 규칙

### 1. 타입의 유연성

`uuid`, `string`, `text`, `int` 등 일반적인 프로그래밍 용어를 사용하십시오. 특정 DB 문법(예: `VARCHAR(255)`)에 얽매이지 않아도 됩니다.

**Good:**
```yaml
fields:
  - name: string
  - email: string (unique)
  - bio: text
```

**Avoid:**
```yaml
fields:
  - name: VARCHAR(100) NOT NULL
  - email: VARCHAR(255) UNIQUE
  - bio: TEXT
```

---

### 2. 연결(Wiring) 명시

* `api` 섹션에서는 반드시 `service`의 어떤 함수를 쓰는지 명시해야 합니다.
* `ui` 섹션에서는 버튼이나 폼이 `api`의 어떤 엔드포인트와 통신하는지 명시해야 합니다.

**Good:**
```yaml
api:
  - path: /users
    method: POST
    use: service.createUser  # Wiring 명시

ui:
  - component: SignUpForm
    view:
      - Button: "가입"
        action: API.POST /users  # Wiring 명시
```

**Bad:**
```yaml
api:
  - path: /users
    method: POST
    # use 누락 - 어떤 서비스 함수를 호출하는지 불명확

ui:
  - component: SignUpForm
    view:
      - Button: "가입"
        # action 누락 - 어떤 API를 호출하는지 불명확
```

---

### 3. 자연어 혼용

로직(steps)이나 검증 조건이 복잡할 경우, 억지로 코드로 표현하지 말고 자연어(한국어/영어)로 명확히 서술하십시오.

**Good:**
```yaml
service:
  - name: likePost
    steps:
      - validate: "본인 글에는 좋아요 불가"
      - validate: "이미 좋아요한 경우 중복 불가"
      - db: "좋아요 레코드 생성"
      - event: "작성자에게 알림 발송"
```

**Avoid:**
```yaml
service:
  - name: likePost
    steps:
      - if: "post.user_id == current_user.id"
        throw: "SELF_LIKE_ERROR"
      - if: "SELECT COUNT(*) FROM likes WHERE ..."
        throw: "DUPLICATE_LIKE_ERROR"
```

---

### 4. 생략 가능

단순 CRUD의 경우 `service` 스텝을 `steps: [CRUD Standard]`로 간소화할 수 있습니다.

**Standard CRUD:**
```yaml
service:
  - name: createPost
    steps: [CRUD Standard]

  - name: getPost
    steps: [CRUD Standard]

  - name: updatePost
    steps: [CRUD Standard]

  - name: deletePost
    steps: [CRUD Standard]
```

**Custom Logic이 필요한 경우:**
```yaml
service:
  - name: createPost
    steps:
      - validate: "제목은 필수"
      - validate: "본문 최소 10자"
      - db: "INSERT INTO posts"
      - event: "팔로워에게 알림"
      - return: "생성된 Post"
```

---

## 네이밍 컨벤션

| 항목 | 컨벤션 | 예시 |
|------|--------|------|
| feature | snake_case | `comment_system` |
| model name | PascalCase | `Comment`, `UserProfile` |
| table name | snake_case | `comments`, `user_profiles` |
| field name | snake_case | `created_at`, `user_id` |
| service function | camelCase | `createComment`, `listComments` |
| component name | PascalCase | `CommentSection`, `CommentForm` |
| API path | kebab-case | `/api/user-profiles` |

---

## 검증 체크리스트

FDL 작성 후 다음 항목을 확인하십시오:

- [ ] 모든 `api`에 `use:` 필드가 있는가?
- [ ] 모든 UI `action`에 API 경로가 명시되어 있는가?
- [ ] `service`의 `input`과 `api`의 `request`가 일치하는가?
- [ ] 외래키(`fk`)가 참조하는 테이블이 존재하는가?
- [ ] 컴포넌트 계층 구조가 명확한가? (`parent` 지정)

---

*ClariSpec FDL Specification v0.0.1 - 2026-02-03*
