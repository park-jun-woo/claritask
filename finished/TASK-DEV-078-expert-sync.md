# TASK-DEV-078: Expert 동기화 로직 구현

## 개요

Expert 파일과 DB 간의 동기화 로직 구현

## 대상 파일

- `cli/internal/service/expert_service.go`

## 작업 내용

### 1. 동기화 정책

specs/CLI/11-Expert.md 정의:
- 파일 수정 → DB content 컬럼에 자동 백업
- 파일 삭제 → DB 백업에서 자동 복구
- UI에서 삭제 → DB + 파일 모두 삭제

### 2. SyncExpert 함수 구현

```go
// SyncExpert syncs expert file with DB
func SyncExpert(database *db.DB, expertID string) error {
    expertPath := filepath.Join(ExpertsDir, expertID, ExpertFileName)

    // Check if file exists
    content, err := os.ReadFile(expertPath)
    if os.IsNotExist(err) {
        // File deleted - restore from DB backup
        return restoreExpertFromDB(database, expertID, expertPath)
    }
    if err != nil {
        return fmt.Errorf("read expert file: %w", err)
    }

    // Calculate hash
    hash := sha256.Sum256(content)
    contentHash := hex.EncodeToString(hash[:])

    // Check if content changed
    var dbHash string
    database.QueryRow("SELECT content_hash FROM experts WHERE id = ?", expertID).Scan(&dbHash)

    if dbHash != contentHash {
        // File modified - update DB backup
        now := db.TimeNow()
        _, err := database.Exec(
            `UPDATE experts SET content = ?, content_hash = ?, updated_at = ? WHERE id = ?`,
            string(content), contentHash, now, expertID,
        )
        if err != nil {
            return fmt.Errorf("update expert content: %w", err)
        }
    }

    return nil
}
```

### 3. restoreExpertFromDB 함수 구현

```go
// restoreExpertFromDB restores expert file from DB backup
func restoreExpertFromDB(database *db.DB, expertID, expertPath string) error {
    var content string
    err := database.QueryRow("SELECT content FROM experts WHERE id = ?", expertID).Scan(&content)
    if err != nil {
        return fmt.Errorf("expert '%s' not found in DB", expertID)
    }

    if content == "" {
        return fmt.Errorf("no backup content for expert '%s'", expertID)
    }

    // Create directory
    dir := filepath.Dir(expertPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("create directory: %w", err)
    }

    // Write file
    if err := os.WriteFile(expertPath, []byte(content), 0644); err != nil {
        return fmt.Errorf("write expert file: %w", err)
    }

    return nil
}
```

### 4. SyncAllExperts 함수 구현

```go
// SyncAllExperts syncs all experts
func SyncAllExperts(database *db.DB) error {
    // Get all expert IDs from DB
    rows, err := database.Query("SELECT id FROM experts")
    if err != nil {
        return err
    }
    defer rows.Close()

    var expertIDs []string
    for rows.Next() {
        var id string
        rows.Scan(&id)
        expertIDs = append(expertIDs, id)
    }

    // Sync each expert
    for _, id := range expertIDs {
        if err := SyncExpert(database, id); err != nil {
            // Log error but continue
            fmt.Fprintf(os.Stderr, "sync expert %s: %v\n", id, err)
        }
    }

    return nil
}
```

### 5. AddExpert 함수 수정

content와 content_hash를 저장하도록 수정:

```go
func AddExpert(database *db.DB, expertID string) (*model.Expert, error) {
    // ... existing code ...

    // Create EXPERT.md template
    template := getExpertTemplate(expertID)
    if err := os.WriteFile(expertPath, []byte(template), 0644); err != nil {
        return nil, fmt.Errorf("create expert file: %w", err)
    }

    // Calculate hash
    hash := sha256.Sum256([]byte(template))
    contentHash := hex.EncodeToString(hash[:])

    // Save to database with content
    now := db.TimeNow()
    _, err := database.Exec(
        `INSERT INTO experts (id, name, version, domain, language, framework, path, content, content_hash, created_at, updated_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        expertID, expertID, "1.0.0", "", "", "", expertPath, template, contentHash, now, now,
    )
    // ... rest of code ...
}
```

### 6. GetExpert 함수에서 동기화 호출

```go
func GetExpert(database *db.DB, expertID string) (*model.Expert, error) {
    // Sync before returning
    SyncExpert(database, expertID)

    // ... existing code ...
}
```

## 의존성

- TASK-DEV-074 완료 필요 (experts 테이블 스키마)

## 완료 조건

- [ ] SyncExpert 함수 구현
- [ ] restoreExpertFromDB 함수 구현
- [ ] SyncAllExperts 함수 구현
- [ ] AddExpert 함수에서 content 저장
- [ ] GetExpert 함수에서 동기화 호출
- [ ] 테스트 작성 및 통과
