# 2026-02-03 버전 표기 정리

## 작업 내용

### 1. 버전 표기 규칙 변경
버전 표기를 `vX.X.0` 또는 `v0.X.0` 형식에서 `v0.0.X` 형식으로 통일.

**변경된 파일:**
| 파일 | 변경 전 | 변경 후 |
|------|---------|---------|
| `vscode-extension/package.json` | 0.3.0 | 0.0.3 |
| `vscode-extension/webview-ui/package.json` | 0.1.0 | 0.0.1 |

**참고:** 외부 라이브러리 버전(Cobra v1.8.0, sqlite3 v1.14.22 등)은 실제 패키지 버전이므로 변경하지 않음.

### 2. specs/ 문서 버전 추가
specs/ 폴더의 모든 md 파일에 버전(v0.0.1)과 변경이력 섹션 추가.

**대상 파일 (6개):**
- `specs/Claritask.md`
- `specs/Commands.md`
- `specs/ClariInit.md`
- `specs/FDL.md`
- `specs/TTY-Handover.md`
- `specs/VscodeGUI.md`

**추가된 형식:**
```markdown
> **버전**: v0.0.1

## 변경이력
| 버전 | 날짜 | 내용 |
|------|------|------|
| v0.0.1 | 2026-02-03 | 최초 작성 |
```

## 버전 표기 규칙 (CLAUDE.md)
- `vX.X.N` 형식이며 테스트하며 수정할때 N 숫자만 올린다
- 10이 넘어도 `vX.X.11`로 표기
