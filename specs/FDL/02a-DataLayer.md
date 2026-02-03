# ClariSpec: DATA LAYER (models)

> **버전**: v0.0.1

## 개요

DATA LAYER는 데이터베이스 스키마와 ORM 모델을 정의합니다. 이 계층은 애플리케이션의 데이터 구조를 결정하며, 다른 모든 계층의 기반이 됩니다.

---

## 기본 구조

```yaml
models:
  - name: <ModelName>       # 모델/엔티티 이름 (PascalCase)
    table: <table_name>     # DB 테이블 이름 (snake_case)
    description: <desc>     # 모델 설명 (선택)
    fields:
      - <name>: <type> [constraints...]
    indexes:                # 인덱스 정의 (선택)
      - fields: [field1, field2]
        unique: true
    relations:              # 관계 정의 (선택)
      - type: hasMany
        model: Comment
        foreignKey: post_id
```

---

## 필드 타입

### 기본 타입

| 타입 | 설명 | SQL 매핑 예시 | 사용 예 |
|------|------|--------------|---------|
| `uuid` | UUID 타입 | `CHAR(36)` / `UUID` | `id: uuid (pk)` |
| `string` | 짧은 문자열 (255자 이내) | `VARCHAR(255)` | `name: string` |
| `text` | 긴 문자열 (무제한) | `TEXT` | `content: text` |
| `int` | 정수 | `INTEGER` / `INT` | `count: int` |
| `bigint` | 큰 정수 | `BIGINT` | `views: bigint` |
| `float` | 부동소수점 | `FLOAT` / `REAL` | `rating: float` |
| `decimal` | 정밀 소수점 | `DECIMAL(10,2)` | `price: decimal` |
| `boolean` | 불리언 | `BOOLEAN` / `TINYINT(1)` | `is_active: boolean` |
| `datetime` | 날짜/시간 | `DATETIME` / `TIMESTAMP` | `created_at: datetime` |
| `date` | 날짜만 | `DATE` | `birth_date: date` |
| `time` | 시간만 | `TIME` | `start_time: time` |
| `json` | JSON 객체 | `JSON` / `TEXT` | `metadata: json` |
| `blob` | 바이너리 데이터 | `BLOB` | `avatar: blob` |
| `enum` | 열거형 | `ENUM(...)` | `status: enum(draft,published,archived)` |

### 타입 옵션

```yaml
# 문자열 길이 제한
username: string(50)          # VARCHAR(50)

# 소수점 정밀도
price: decimal(10,2)          # DECIMAL(10,2) - 총 10자리, 소수점 2자리

# 열거형 값 목록
status: enum(pending,active,inactive)
```

---

## 제약조건

### 기본 제약조건

| 제약조건 | 설명 | 예시 |
|----------|------|------|
| `pk` | Primary Key (자동 NOT NULL) | `id: uuid (pk)` |
| `fk: <table.field>` | Foreign Key | `user_id: uuid (fk: users.id)` |
| `required` | NOT NULL | `email: string (required)` |
| `unique` | 고유값 | `email: string (unique)` |
| `default: <value>` | 기본값 | `status: string (default: "active")` |
| `auto` | Auto Increment | `id: int (pk, auto)` |

### 고급 제약조건

| 제약조건 | 설명 | 예시 |
|----------|------|------|
| `index` | 일반 인덱스 | `email: string (index)` |
| `nullable` | NULL 허용 (명시적) | `deleted_at: datetime (nullable)` |
| `check: <condition>` | 체크 제약조건 | `age: int (check: age >= 0)` |
| `onDelete: <action>` | FK 삭제 시 동작 | `fk: users.id (onDelete: cascade)` |
| `onUpdate: <action>` | FK 수정 시 동작 | `fk: users.id (onUpdate: cascade)` |

### FK 동작 옵션

| 옵션 | 설명 |
|------|------|
| `cascade` | 부모 삭제/수정 시 자식도 삭제/수정 |
| `restrict` | 자식이 있으면 부모 삭제/수정 금지 |
| `setNull` | 부모 삭제/수정 시 자식 FK를 NULL로 설정 |
| `noAction` | 아무 동작 안함 (기본값) |

---

## 인덱스 정의

```yaml
models:
  - name: Comment
    table: comments
    fields:
      - id: uuid (pk)
      - post_id: uuid (fk: posts.id)
      - user_id: uuid (fk: users.id)
      - created_at: datetime
    indexes:
      # 단일 필드 인덱스
      - fields: [created_at]

      # 복합 인덱스
      - fields: [post_id, created_at]
        name: idx_post_comments    # 인덱스 이름 (선택)

      # 고유 인덱스
      - fields: [post_id, user_id]
        unique: true
        name: uniq_post_user_comment

      # 부분 인덱스 (조건부)
      - fields: [status]
        where: "status = 'active'"
```

---

## 관계 정의

### 관계 타입

| 타입 | 설명 | 예시 |
|------|------|------|
| `hasOne` | 1:1 관계 | User hasOne Profile |
| `hasMany` | 1:N 관계 | Post hasMany Comment |
| `belongsTo` | N:1 역관계 | Comment belongsTo Post |
| `belongsToMany` | N:M 관계 | Post belongsToMany Tag |

### 관계 정의 예시

```yaml
models:
  - name: User
    table: users
    fields:
      - id: uuid (pk)
      - name: string
    relations:
      - type: hasOne
        model: Profile
        foreignKey: user_id

      - type: hasMany
        model: Post
        foreignKey: author_id

  - name: Post
    table: posts
    fields:
      - id: uuid (pk)
      - author_id: uuid (fk: users.id)
      - title: string
    relations:
      - type: belongsTo
        model: User
        foreignKey: author_id
        as: author              # 별칭 지정

      - type: belongsToMany
        model: Tag
        through: post_tags      # 중간 테이블
        foreignKey: post_id
        otherKey: tag_id
```

---

## Soft Delete 패턴

```yaml
models:
  - name: Post
    table: posts
    softDelete: true            # deleted_at 필드 자동 추가
    fields:
      - id: uuid (pk)
      - title: string
      # deleted_at: datetime (nullable) - 자동 생성됨
```

또는 명시적으로:

```yaml
models:
  - name: Post
    table: posts
    fields:
      - id: uuid (pk)
      - title: string
      - deleted_at: datetime (nullable)  # Soft Delete용
```

---

## Timestamps 패턴

```yaml
models:
  - name: Post
    table: posts
    timestamps: true            # created_at, updated_at 자동 추가
    fields:
      - id: uuid (pk)
      - title: string
      # created_at: datetime (default: now) - 자동 생성
      # updated_at: datetime (default: now, onUpdate: now) - 자동 생성
```

---

## 전체 예시

```yaml
models:
  - name: User
    table: users
    description: 시스템 사용자
    timestamps: true
    fields:
      - id: uuid (pk)
      - email: string (required, unique, index)
      - password_hash: string (required)
      - name: string (required)
      - role: enum(admin,user,guest) (default: "user")
      - is_active: boolean (default: true)
      - last_login_at: datetime (nullable)
      - metadata: json (nullable)
    indexes:
      - fields: [role, is_active]
        name: idx_user_role_status

  - name: Post
    table: posts
    description: 블로그 게시글
    timestamps: true
    softDelete: true
    fields:
      - id: uuid (pk)
      - author_id: uuid (fk: users.id, onDelete: cascade)
      - title: string (required)
      - slug: string (required, unique)
      - content: text (required)
      - excerpt: string(500) (nullable)
      - status: enum(draft,published,archived) (default: "draft")
      - view_count: bigint (default: 0)
      - published_at: datetime (nullable)
    indexes:
      - fields: [author_id, status]
      - fields: [published_at]
        where: "status = 'published'"
    relations:
      - type: belongsTo
        model: User
        foreignKey: author_id
        as: author
      - type: hasMany
        model: Comment
        foreignKey: post_id

  - name: Comment
    table: comments
    timestamps: true
    fields:
      - id: uuid (pk)
      - post_id: uuid (fk: posts.id, onDelete: cascade)
      - user_id: uuid (fk: users.id, onDelete: cascade)
      - parent_id: uuid (fk: comments.id, nullable, onDelete: cascade)
      - content: text (required)
      - is_approved: boolean (default: false)
    indexes:
      - fields: [post_id, created_at]
    relations:
      - type: belongsTo
        model: Post
        foreignKey: post_id
      - type: belongsTo
        model: User
        foreignKey: user_id
        as: author
      - type: hasMany
        model: Comment
        foreignKey: parent_id
        as: replies
```

---

## 네이밍 컨벤션

| 항목 | 규칙 | 예시 |
|------|------|------|
| 모델명 | PascalCase, 단수형 | `User`, `PostComment` |
| 테이블명 | snake_case, 복수형 | `users`, `post_comments` |
| 필드명 | snake_case | `created_at`, `user_id` |
| 외래키 | `<참조테이블_단수>_id` | `user_id`, `post_id` |
| 중간 테이블 | `<테이블1>_<테이블2>` (알파벳순) | `post_tags` |
| 인덱스명 | `idx_<테이블>_<필드>` | `idx_posts_author` |
| 유니크 인덱스명 | `uniq_<테이블>_<필드>` | `uniq_users_email` |

---

*ClariSpec FDL Specification v0.0.1 - 2026-02-03*
