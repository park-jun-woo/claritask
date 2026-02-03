# TASK-EXT-033: Message Handler CLI 호출 연동

## 목표
Extension Host에서 CLI 호출 메시지 처리

## 변경 파일
- `vscode-extension/src/CltEditorProvider.ts`

## 작업 내용

### 1. CLI 서비스 import
```typescript
import { createFeature, validateFDL, generateTasks, generateSkeleton } from './cliService';
```

### 2. 메시지 핸들러 확장
```typescript
private async handleMessage(message: any, webview: vscode.Webview) {
    switch (message.type) {
        // ... 기존 케이스들

        case 'createFeature':
            await this.handleCreateFeature(message.data, webview);
            break;

        case 'validateFDL':
            await this.handleValidateFDL(message.featureId, webview);
            break;

        case 'generateTasks':
            await this.handleGenerateTasks(message.featureId, webview);
            break;

        case 'generateSkeleton':
            await this.handleGenerateSkeleton(message.featureId, message.dryRun, webview);
            break;
    }
}
```

### 3. CLI 호출 핸들러 구현
```typescript
private async handleCreateFeature(data: any, webview: vscode.Webview) {
    // 진행 상태 알림
    webview.postMessage({
        type: 'cliProgress',
        command: 'feature.create',
        step: 'creating',
        message: 'Creating feature...'
    });

    const result = await createFeature(data);

    // 결과 전송
    webview.postMessage({
        type: 'cliResult',
        command: 'feature.create',
        ...result
    });

    // 성공 시 데이터 새로고침
    if (result.success) {
        await this.syncData(webview);
    }
}

private async handleValidateFDL(featureId: number, webview: vscode.Webview) {
    const result = await validateFDL(featureId);
    webview.postMessage({
        type: 'cliResult',
        command: 'fdl.validate',
        ...result
    });
}

private async handleGenerateTasks(featureId: number, webview: vscode.Webview) {
    const result = await generateTasks(featureId);
    webview.postMessage({
        type: 'cliResult',
        command: 'fdl.tasks',
        ...result
    });

    if (result.success) {
        await this.syncData(webview);
    }
}

private async handleGenerateSkeleton(
    featureId: number,
    dryRun: boolean,
    webview: vscode.Webview
) {
    const result = await generateSkeleton(featureId);
    webview.postMessage({
        type: 'cliResult',
        command: 'fdl.skeleton',
        ...result
    });
}
```

## 테스트
- createFeature 메시지 전송 시 CLI 호출 확인
- 결과 메시지가 Webview에 전달되는지 확인
- 에러 시 에러 메시지 전달 확인

## 관련 스펙
- specs/VSCode/11-MessageProtocol.md (v0.0.6)
- specs/VSCode/14-CLICompatibility.md (v0.0.6)
