# Feature FDL TTY Handover 구현 (2026-02-03)

## 배경

이전 대화에서 사용자가 다음 문제점들을 지적:
1. Feature 추가 시 description을 Claude Code로 전달하여 FDL YAML을 생성해야 하는데 구현 안됨
2. Feature md 파일은 불필요하므로 생성 기능 삭제 필요
3. FDL yaml만 생성하는 것으로 변경
4. VSCode Extension에서 Feature 생성 시 CLI 호출이 처리 안됨

사용자 요청: "specs/에 반영하고, 적용하는 계획 수립해"

## 완료된 작업

### TASK-DEV-104: Feature MD 파일 생성 기능 제거

**변경 파일:**
- `cli/internal/service/feature_service.go`
- `cli/internal/cmd/feature.go`
- `cli/internal/cmd/fdl.go`
- `cli/internal/cmd/plan.go`

**주요 변경:**
```go
// Before: CreateFeature가 md 파일 생성 및 *CreateFeatureResult 반환
func CreateFeature(database *db.DB, projectID, name, description string) (*CreateFeatureResult, error)

// After: DB 레코드만 생성, int64 반환
func CreateFeature(database *db.DB, projectID, name, description string) (int64, error)
```

- `featureMarkdownTemplate` 상수 제거
- `CalculateContentHash` 함수 제거
- 모든 호출부 업데이트 (`feature.go`, `fdl.go`, `plan.go`)

---

### TASK-DEV-105: Feature Add TTY Handover 구현

**변경 파일:**
- `cli/internal/service/tty_service.go`
- `cli/internal/cmd/feature.go`

**추가된 함수 (tty_service.go):**
```go
// FDL 생성 시스템 프롬프트
func FDLGenerationSystemPrompt() string

// TTY Handover로 FDL 생성 실행
func RunFDLGenerationWithTTY(database *db.DB, featureID int64, featureName, description string) error

// FDL 생성 프롬프트 빌드
func BuildFDLPrompt(featureID int64, name, description, projectContext, techContext, designContext string) string
```

**feature.go 변경:**
```go
// feature add 명령어에 --no-tty 플래그 추가
featureAddCmd.Flags().Bool("no-tty", false, "Skip TTY handover (just create DB record)")

// runFeatureAdd: DB 생성 후 TTY Handover 호출
func runFeatureAdd(cmd *cobra.Command, args []string) error {
    // ... DB에 Feature 생성 ...

    if !noTTY {
        // Claude Code로 FDL 생성
        service.RunFDLGenerationWithTTY(database, featureID, input.Name, input.Description)
    }
}
```

---

### TASK-DEV-106: Feature FDL 재생성 명령어

**변경 파일:**
- `cli/internal/cmd/feature.go`

**추가된 명령어:**
```go
var featureFdlCmd = &cobra.Command{
    Use:   "fdl <id>",
    Short: "Regenerate FDL for a feature using Claude Code",
    Args:  cobra.ExactArgs(1),
    RunE:  runFeatureFdl,
}
```

사용법: `clari feature fdl 1` - Feature ID 1의 FDL 재생성

---

### TASK-EXT-034: VSCode Terminal CLI 호출

**변경 파일:**
- `vscode-extension/src/CltEditorProvider.ts`
- `vscode-extension/src/extension.ts`
- `vscode-extension/src/database.ts`

**CltEditorProvider.ts - handleCreateFeature 변경:**
```typescript
// Before: 직접 DB에 삽입
const id = database.createFeature(message.name, message.description);

// After: Terminal로 CLI 실행 (TTY Handover 지원)
const terminal = vscode.window.createTerminal({
    name: 'Claritask - Create Feature',
    cwd: vscode.workspace.workspaceFolders?.[0].uri.fsPath
});
terminal.show();
terminal.sendText(`clari feature add '${escapedInput}'`);
```

**extension.ts - FDL 파일 워처:**
```typescript
// md 파일 대신 fdl.yaml 파일 감시
fdlWatcher = vscode.workspace.createFileSystemWatcher(
    new vscode.RelativePattern(workspaceFolder, 'features/*.fdl.yaml')
);

fdlWatcher.onDidChange(uri => syncFDLFile(uri, database!));
fdlWatcher.onDidCreate(uri => syncFDLFile(uri, database!));
fdlWatcher.onDidDelete(uri => handleFDLFileDeleted(uri, database!));
```

**database.ts - FDL sync 메서드 추가:**
```typescript
updateFeatureFDL(featureId: number, fdl: string, fdlHash: string): void
clearFeatureFDL(featureId: number): void
```

---

## 동작 흐름

```
[VSCode Feature 생성]
        │
        ▼
[Terminal 열기 + clari feature add 실행]
        │
        ▼
[CLI: DB 레코드 생성]
        │
        ▼
[Claude Code TTY Handover]
        │
        ▼
[Claude Code: FDL YAML 생성]
        │
        ▼
[features/<name>.fdl.yaml 저장]
        │
        ▼
[FDL 파일 워처: DB 자동 동기화]
```

## specs 변경 (이전 대화에서 완료)

- `specs/CLI/07-Feature.md` → v0.0.7
- `specs/VSCode/05-FeaturesTab.md` → v0.0.7
- `specs/HISTORY.md` → v0.0.7

---

*2026-02-03 작업 완료*
