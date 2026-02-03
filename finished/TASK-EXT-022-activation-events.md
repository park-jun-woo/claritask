# TASK-EXT-022: Activation Events 설정

## 개요
VSCode extension이 `.claritask/db.clt` 파일이 있는 워크스페이스에서 자동 활성화되도록 설정

## 배경
- **스펙**: specs/VSCode/10-ExtensionStructure.md
- **현재 상태**: activationEvents가 빈 배열 `[]`
- **문제**: 사용자가 수동으로 extension을 활성화해야 함

## 작업 내용

### 1. package.json 수정
**파일**: `vscode-extension/package.json`

현재:
```json
"activationEvents": []
```

수정:
```json
"activationEvents": [
  "workspaceContains:.claritask/db.clt"
]
```

## 검증
1. VSCode에서 `.claritask/db.clt` 파일이 있는 폴더 열기
2. Extension이 자동으로 활성화되는지 확인

## 완료 기준
- [ ] activationEvents 설정 완료
- [ ] 워크스페이스 자동 활성화 동작 확인
