# TASK-EXT-029: Expert 파일 동기화 (FileSystemWatcher)

## 개요
EXPERT.md 파일 변경을 감지하고 DB와 자동 동기화

## 배경
- **스펙**: specs/VSCode/08-ExpertSync.md
- **현재 상태**: 파일 감시 없음

## 작업 내용

### 1. FileSystemWatcher 설정
**파일**: `vscode-extension/src/CltEditorProvider.ts`

```typescript
import * as vscode from 'vscode';
import * as crypto from 'crypto';

private expertWatcher: vscode.FileSystemWatcher | undefined;

private setupExpertWatcher() {
  const expertsPattern = new vscode.RelativePattern(
    this.dbDir,
    'experts/*/EXPERT.md'
  );

  this.expertWatcher = vscode.workspace.createFileSystemWatcher(expertsPattern);

  // 파일 변경 감지
  this.expertWatcher.onDidChange(async (uri) => {
    await this.syncExpertFromFile(uri);
  });

  // 파일 생성 감지
  this.expertWatcher.onDidCreate(async (uri) => {
    await this.syncExpertFromFile(uri);
  });

  // 파일 삭제 감지
  this.expertWatcher.onDidDelete(async (uri) => {
    await this.handleExpertFileDeleted(uri);
  });
}

private async syncExpertFromFile(uri: vscode.Uri) {
  const filePath = uri.fsPath;
  const expertId = this.extractExpertId(filePath);

  try {
    const content = await fs.promises.readFile(filePath, 'utf-8');
    const contentHash = crypto.createHash('sha256').update(content).digest('hex');

    // DB에서 현재 해시 확인
    const stmt = this.db.prepare('SELECT content_hash FROM experts WHERE id = ?');
    stmt.bind([expertId]);

    if (stmt.step()) {
      const row = stmt.getAsObject();
      if (row.content_hash !== contentHash) {
        // 내용이 변경됨 - DB 업데이트
        const now = new Date().toISOString();
        this.db.run(`
          UPDATE experts SET content = ?, content_hash = ?, updated_at = ?
          WHERE id = ?
        `, [content, contentHash, now, expertId]);

        // 메타데이터 파싱 및 업데이트
        const metadata = this.parseExpertMetadata(content);
        if (metadata) {
          this.db.run(`
            UPDATE experts SET name = ?, domain = ?, language = ?, framework = ?
            WHERE id = ?
          `, [metadata.name, metadata.domain, metadata.language, metadata.framework, expertId]);
        }

        this.saveDatabase();
        this.syncToWebview();
      }
    }
    stmt.free();
  } catch (error) {
    console.error('Error syncing expert file:', error);
  }
}

private async handleExpertFileDeleted(uri: vscode.Uri) {
  const filePath = uri.fsPath;
  const expertId = this.extractExpertId(filePath);

  // DB 백업에서 복구 시도
  const stmt = this.db.prepare('SELECT content FROM experts WHERE id = ?');
  stmt.bind([expertId]);

  if (stmt.step()) {
    const row = stmt.getAsObject();
    const content = row.content as string;

    if (content) {
      // 파일 복구
      const expertsDir = path.dirname(filePath);
      await fs.promises.mkdir(expertsDir, { recursive: true });
      await fs.promises.writeFile(filePath, content);

      vscode.window.showInformationMessage(
        `Expert file '${expertId}' was restored from backup.`
      );
    }
  }
  stmt.free();
}

private extractExpertId(filePath: string): string {
  // .claritask/experts/<expert-id>/EXPERT.md
  const parts = filePath.split(path.sep);
  const expertIndex = parts.indexOf('experts');
  if (expertIndex >= 0 && parts.length > expertIndex + 1) {
    return parts[expertIndex + 1];
  }
  return '';
}

private parseExpertMetadata(content: string): ExpertMetadata | null {
  // # <name> 추출
  const nameMatch = content.match(/^#\s+(.+)$/m);
  const name = nameMatch ? nameMatch[1] : '';

  // ## Role 섹션에서 domain 추출
  const roleMatch = content.match(/##\s+Role\s*\n([^\n#]+)/i);
  const domain = roleMatch ? roleMatch[1].trim() : '';

  // Language, Framework 추출
  const langMatch = content.match(/Language:\s*(.+)/i);
  const language = langMatch ? langMatch[1].trim() : '';

  const fwMatch = content.match(/Framework:\s*(.+)/i);
  const framework = fwMatch ? fwMatch[1].trim() : '';

  return { name, domain, language, framework };
}
```

### 2. Dispose 시 Watcher 정리
```typescript
dispose() {
  this.expertWatcher?.dispose();
  // 기존 dispose 로직...
}
```

### 3. 초기화 시 Watcher 시작
```typescript
// resolveCustomEditor에서 호출
this.setupExpertWatcher();
```

## 완료 기준
- [ ] FileSystemWatcher 설정
- [ ] 파일 변경 시 DB 동기화
- [ ] 파일 삭제 시 백업에서 복구
- [ ] 메타데이터 파싱
- [ ] Webview 자동 업데이트
