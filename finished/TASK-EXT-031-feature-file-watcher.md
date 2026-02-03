# TASK-EXT-031: Feature 파일 Watcher

## 목표
`features/*.md` 파일 변경 감지 및 DB 동기화

## 변경 파일
- `vscode-extension/src/extension.ts`
- `vscode-extension/src/sync.ts`

## 작업 내용

### 1. FileSystemWatcher 설정 (extension.ts)
```typescript
export function activate(context: vscode.ExtensionContext) {
    // ... 기존 코드

    // Feature 파일 감시
    const featureWatcher = vscode.workspace.createFileSystemWatcher(
        '**/features/*.md'
    );

    featureWatcher.onDidChange(uri => syncFeatureToDB(uri));
    featureWatcher.onDidCreate(uri => syncFeatureToDB(uri));
    featureWatcher.onDidDelete(uri => clearFeatureFilePath(uri));

    context.subscriptions.push(featureWatcher);
}
```

### 2. sync.ts에 동기화 함수 추가
```typescript
import * as crypto from 'crypto';

export async function syncFeatureToDB(uri: vscode.Uri): Promise<void> {
    const filePath = uri.fsPath;
    const fileName = path.basename(filePath, '.md');

    // 파일 내용 읽기
    const content = await fs.readFile(filePath, 'utf-8');
    const contentHash = crypto.createHash('sha256').update(content).digest('hex');

    // DB에서 해당 feature 찾기 (name으로)
    const db = getDatabase();
    const feature = db.prepare('SELECT * FROM features WHERE name = ?').get(fileName);

    if (feature && feature.content_hash !== contentHash) {
        // 해시가 다르면 업데이트
        db.prepare(`
            UPDATE features
            SET content = ?, content_hash = ?, file_path = ?
            WHERE id = ?
        `).run(content, contentHash, filePath, feature.id);

        // Webview에 알림
        notifyWebview('featuresUpdated');
    }
}

export async function clearFeatureFilePath(uri: vscode.Uri): Promise<void> {
    const fileName = path.basename(uri.fsPath, '.md');
    const db = getDatabase();

    db.prepare(`
        UPDATE features
        SET file_path = ''
        WHERE name = ?
    `).run(fileName);
}
```

## 테스트
- `features/test.md` 파일 수정 시 DB 업데이트 확인
- 파일 삭제 시 file_path 클리어 확인
- Webview에 변경 알림 전달 확인

## 관련 스펙
- specs/VSCode/05-FeaturesTab.md (v0.0.6)
