# TASK-DEV-077: clari db 명령어 구현

## 개요

데이터베이스 관리를 위한 clari db 명령어 구현

## 대상 파일

- `cli/internal/cmd/db.go` (신규)
- `cli/internal/service/db_service.go` (신규)

## 작업 내용

### 1. cmd/db.go 생성

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
    Use:   "db",
    Short: "Database management commands",
}

var dbVersionCmd = &cobra.Command{
    Use:   "version",
    Short: "Show current database version",
    RunE:  runDBVersion,
}

var dbMigrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Run database migrations",
    RunE:  runDBMigrate,
}

var dbRollbackCmd = &cobra.Command{
    Use:   "rollback",
    Short: "Rollback to specific version",
    RunE:  runDBRollback,
}

var dbBackupCmd = &cobra.Command{
    Use:   "backup",
    Short: "Create database backup",
    RunE:  runDBBackup,
}

var rollbackVersion int

func init() {
    dbCmd.AddCommand(dbVersionCmd)
    dbCmd.AddCommand(dbMigrateCmd)
    dbCmd.AddCommand(dbRollbackCmd)
    dbCmd.AddCommand(dbBackupCmd)

    dbRollbackCmd.Flags().IntVar(&rollbackVersion, "version", 0, "Target version to rollback")
    dbRollbackCmd.MarkFlagRequired("version")
}
```

### 2. 명령어 구현

#### clari db version
```go
func runDBVersion(cmd *cobra.Command, args []string) error {
    database, err := getDB()
    if err != nil {
        outputError(err)
        return nil
    }
    defer database.Close()

    version := database.GetVersion()
    outputJSON(map[string]interface{}{
        "success": true,
        "current_version": version,
        "latest_version": db.LatestVersion,
        "up_to_date": version >= db.LatestVersion,
    })
    return nil
}
```

#### clari db migrate
```go
func runDBMigrate(cmd *cobra.Command, args []string) error {
    database, err := getDB()
    if err != nil {
        outputError(err)
        return nil
    }
    defer database.Close()

    beforeVersion := database.GetVersion()
    if err := database.AutoMigrate(); err != nil {
        outputError(err)
        return nil
    }
    afterVersion := database.GetVersion()

    outputJSON(map[string]interface{}{
        "success": true,
        "before_version": beforeVersion,
        "after_version": afterVersion,
        "migrations_applied": afterVersion - beforeVersion,
        "message": "Migration completed",
    })
    return nil
}
```

#### clari db rollback
```go
func runDBRollback(cmd *cobra.Command, args []string) error {
    database, err := getDB()
    if err != nil {
        outputError(err)
        return nil
    }
    defer database.Close()

    beforeVersion := database.GetVersion()
    if err := database.Rollback(rollbackVersion); err != nil {
        outputError(err)
        return nil
    }

    outputJSON(map[string]interface{}{
        "success": true,
        "before_version": beforeVersion,
        "after_version": rollbackVersion,
        "message": "Rollback completed",
    })
    return nil
}
```

#### clari db backup
```go
func runDBBackup(cmd *cobra.Command, args []string) error {
    database, err := getDB()
    if err != nil {
        outputError(err)
        return nil
    }
    defer database.Close()

    backupPath, err := database.Backup()
    if err != nil {
        outputError(err)
        return nil
    }

    outputJSON(map[string]interface{}{
        "success": true,
        "backup_path": backupPath,
        "message": "Backup created successfully",
    })
    return nil
}
```

### 3. root.go에 db 명령어 등록

```go
rootCmd.AddCommand(dbCmd)
```

### 4. db.go에 Backup 함수 추가

```go
func (db *DB) Backup() (string, error) {
    backupPath := fmt.Sprintf("%s.backup.%s", db.path, time.Now().Format("20060102150405"))

    src, err := os.Open(db.path)
    if err != nil {
        return "", err
    }
    defer src.Close()

    dst, err := os.Create(backupPath)
    if err != nil {
        return "", err
    }
    defer dst.Close()

    if _, err := io.Copy(dst, src); err != nil {
        return "", err
    }

    return backupPath, nil
}
```

## 의존성

- TASK-DEV-075 완료 필요 (마이그레이션 시스템)

## 완료 조건

- [ ] cmd/db.go 파일 생성
- [ ] clari db version 구현
- [ ] clari db migrate 구현
- [ ] clari db rollback 구현
- [ ] clari db backup 구현
- [ ] root.go에 db 명령어 등록
- [ ] 테스트 작성 및 통과
