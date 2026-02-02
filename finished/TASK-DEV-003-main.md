# TASK-DEV-003: 메인 진입점

## 파일
`cmd/talos/main.go`

## 목표
Cobra CLI 애플리케이션 메인 진입점 구현

## 작업 내용

### 1. main 함수
```go
package main

import (
    "os"
    "github.com/your/talos/internal/cmd"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

## 참조
- `internal/cmd/root.go` - Root 명령어
