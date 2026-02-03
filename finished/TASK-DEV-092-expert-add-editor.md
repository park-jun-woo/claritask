# TASK-DEV-092: expert add 에디터 자동열기

## 개요

specs에 따르면 expert add 후 에디터를 자동으로 열어야 함

## 스펙 요구사항 (11-Expert.md)

```
**동작:**
1. `.claritask/experts/<expert-id>/` 폴더 생성
2. `EXPERT.md` 템플릿 파일 생성
3. (옵션) 에디터로 파일 열기
```

## 현재 상태

1, 2단계만 구현됨. 에디터 자동열기 없음.

## 작업 내용

### 1. 플래그 추가 (cmd/expert.go)

```go
var expertAddOpenEditor bool

func init() {
    expertAddCmd.Flags().BoolVar(&expertAddOpenEditor, "edit", false, "Open editor after creation")
}
```

### 2. runExpertAdd 수정

```go
func runExpertAdd(cmd *cobra.Command, args []string) error {
    // ... 기존 로직 ...

    expert, err := service.AddExpert(database, expertID)
    if err != nil {
        outputError(err)
        return nil
    }

    // 에디터 열기 (--edit 플래그 또는 기본값)
    if expertAddOpenEditor {
        openEditor(expert.Path)
    }

    outputJSON(map[string]interface{}{
        "success":   true,
        "expert_id": expert.ID,
        "path":      expert.Path,
        "message":   "Expert created. Edit the file to define the expert.",
    })

    return nil
}

func openEditor(filePath string) {
    editor := os.Getenv("EDITOR")
    if editor == "" {
        if runtime.GOOS == "windows" {
            editor = "notepad"
        } else {
            editor = "vi"
        }
    }

    execCmd := exec.Command(editor, filePath)
    execCmd.Stdin = os.Stdin
    execCmd.Stdout = os.Stdout
    execCmd.Stderr = os.Stderr
    execCmd.Run()
}
```

## 완료 조건

- [ ] --edit 플래그 추가
- [ ] 에디터 자동열기 구현
- [ ] 테스트
