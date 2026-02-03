# TASK-DEV-022: Feature 커맨드

## 개요
- **파일**: `internal/cmd/feature.go`
- **유형**: 신규
- **우선순위**: High
- **Phase**: 1 (핵심 기초 구조)
- **예상 LOC**: ~200

## 목적
`clari feature` 명령어 그룹 구현

## 작업 내용

### 1. 커맨드 구조

```go
var featureCmd = &cobra.Command{
    Use:   "feature",
    Short: "Feature management commands",
}

// 서브커맨드
var featureListCmd    // clari feature list
var featureAddCmd     // clari feature add '<json>'
var featureGetCmd     // clari feature get <id>
var featureSpecCmd    // clari feature <id> spec '<spec>'
var featureStartCmd   // clari feature <id> start
```

### 2. clari feature list

Feature 목록 조회

```bash
clari feature list
```

**응답**:
```json
{
  "success": true,
  "features": [
    {
      "id": 1,
      "name": "로그인",
      "description": "사용자 인증 기능",
      "status": "done",
      "tasks_total": 4,
      "tasks_done": 4,
      "depends_on": []
    }
  ],
  "total": 1
}
```

### 3. clari feature add

Feature 추가

```bash
clari feature add '{"name": "로그인", "description": "사용자 인증 기능"}'
```

**JSON 포맷**:
```go
type featureAddInput struct {
    Name        string `json:"name"`        // 필수
    Description string `json:"description"` // 필수
}
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "name": "로그인",
  "message": "Feature created successfully"
}
```

### 4. clari feature get

Feature 상세 조회

```bash
clari feature get 1
```

**응답**:
```json
{
  "success": true,
  "feature": {
    "id": 1,
    "name": "로그인",
    "description": "사용자 인증 기능",
    "spec": "JWT 기반 인증...",
    "status": "active",
    "created_at": "2026-02-03T10:00:00Z"
  }
}
```

### 5. clari feature spec

Feature Spec 설정

```bash
clari feature 1 spec 'JWT 기반 인증. Access token 1시간, Refresh token 7일.'
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "message": "Feature spec updated successfully"
}
```

### 6. clari feature start

Feature 실행 시작

```bash
clari feature 1 start
```

**응답**:
```json
{
  "success": true,
  "feature_id": 1,
  "name": "로그인",
  "mode": "execution",
  "pending_tasks": 4,
  "message": "Feature execution started"
}
```

### 7. root.go 수정
`internal/cmd/root.go`의 `init()`에 featureCmd 등록

## 의존성
- TASK-DEV-021 (Feature 서비스)

## 완료 기준
- [ ] 모든 서브커맨드 구현됨
- [ ] JSON 출력 형식 준수
- [ ] 에러 처리 구현됨
- [ ] root.go에 등록됨
- [ ] go build 성공
