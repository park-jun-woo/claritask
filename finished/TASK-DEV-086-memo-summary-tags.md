# TASK-DEV-086: Memo summary, tags 지원

## 개요

Memo의 data 필드에서 summary, tags 지원 확인 및 구현

## 스펙 요구사항 (05-Memo.md)

```json
{
  "value": "Use httpOnly cookies for refresh tokens",
  "priority": 1,
  "summary": "JWT 보안 정책",
  "tags": ["security", "jwt"]
}
```

## 현재 상태

- Memo.Data는 JSON 문자열로 저장
- 구조화된 검증 없음
- summary, tags 필드 사용 가능하지만 명시적 지원 없음

## 작업 내용

### 1. MemoData 구조체 확장 (models.go)

```go
type MemoData struct {
    Scope    string                 `json:"scope"`
    ScopeID  string                 `json:"scope_id"`
    Key      string                 `json:"key"`
    Data     map[string]interface{} `json:"data"`
    Priority int                    `json:"priority"`
}

// Data 내부 구조 (문서화용)
type MemoContent struct {
    Value   string   `json:"value"`
    Summary string   `json:"summary,omitempty"`
    Tags    []string `json:"tags,omitempty"`
}
```

### 2. memo list 출력 개선 (cmd/memo.go)

```go
// runMemoList에서 summary 표시
for _, memo := range memos {
    var content map[string]interface{}
    json.Unmarshal([]byte(memo.Data), &content)

    summary := ""
    if s, ok := content["summary"].(string); ok {
        summary = s
    }

    memoList = append(memoList, map[string]interface{}{
        "key":      memo.Key,
        "priority": memo.Priority,
        "summary":  summary,
    })
}
```

### 3. memo get 출력 개선

created_at, updated_at 포함:

```go
outputJSON(map[string]interface{}{
    "success":    true,
    "scope":      memo.Scope,
    "scope_id":   memo.ScopeID,
    "key":        memo.Key,
    "data":       data,
    "priority":   memo.Priority,
    "created_at": memo.CreatedAt.Format(time.RFC3339),
    "updated_at": memo.UpdatedAt.Format(time.RFC3339),
})
```

## 완료 조건

- [ ] MemoContent 구조체 문서화
- [ ] memo list에서 summary 표시
- [ ] memo get에서 created_at, updated_at 포함
- [ ] 테스트 작성
