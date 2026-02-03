# TASK-DEV-101: DB Feature 테이블 필드 추가

## 목표
features 테이블에 파일 동기화를 위한 필드 추가

## 변경 파일
- `cli/internal/db/db.go`

## 작업 내용

### 1. 마이그레이션 추가
```go
// migrateFeaturesFileFields 마이그레이션 함수 추가
{
    "features_file_fields",
    `ALTER TABLE features ADD COLUMN file_path TEXT DEFAULT '';
     ALTER TABLE features ADD COLUMN content TEXT DEFAULT '';
     ALTER TABLE features ADD COLUMN content_hash TEXT DEFAULT '';`,
}
```

### 2. Feature 구조체 필드 확인
`cli/internal/model/models.go`에 이미 필드가 있는지 확인하고 없으면 추가:
```go
type Feature struct {
    // ... 기존 필드
    FilePath    string `json:"file_path"`
    Content     string `json:"content"`
    ContentHash string `json:"content_hash"`
}
```

## 테스트
- 마이그레이션 실행 후 features 테이블에 새 필드 존재 확인
- 기존 데이터에 영향 없음 확인

## 관련 스펙
- specs/DB/02-A-Core.md (v0.0.5)
