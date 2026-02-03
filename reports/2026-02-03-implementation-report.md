# Implementation Report: specs 구현 완료

> **작성일**: 2026-02-03
> **현재 버전**: v0.0.4

---

## 작업 요약

| Task ID | 내용 | 상태 |
|---------|------|------|
| TASK-DEV-074 | experts 테이블 스키마 업데이트 | 완료 |
| TASK-DEV-075 | _migrations 테이블 및 버전 관리 | 완료 |
| TASK-DEV-076 | 인덱스 추가 | 완료 |
| TASK-DEV-077 | clari db 명령어 구현 | 완료 |
| TASK-DEV-078 | Expert 동기화 로직 | 완료 |

---

## 변경 파일

### 수정된 파일

1. **cli/internal/db/db.go**
   - experts 테이블에 content, content_hash, updated_at 컬럼 추가
   - LatestVersion 상수 추가 (버전 5)
   - GetVersion(), setVersion(), AutoMigrate() 함수 추가
   - migrateV5() - 인덱스 생성
   - Rollback(), Backup(), Path() 함수 추가

2. **cli/internal/model/models.go**
   - Expert 구조체에 Content, ContentHash, UpdatedAt 필드 추가

3. **cli/internal/service/expert_service.go**
   - crypto/sha256, encoding/hex import 추가
   - AddExpert()에서 content, content_hash 저장
   - SyncExpert(), restoreExpertFromDB(), SyncAllExperts() 함수 추가

4. **cli/internal/cmd/root.go**
   - dbCmd 등록 추가

### 신규 파일

1. **cli/internal/cmd/db.go**
   - clari db version
   - clari db migrate
   - clari db rollback --version <n>
   - clari db backup

---

## 인덱스 목록 (v5 마이그레이션)

```sql
idx_features_project     - features(project_id)
idx_features_status      - features(status)
idx_tasks_feature        - tasks(feature_id)
idx_tasks_status         - tasks(status)
idx_task_edges_to        - task_edges(to_task_id)
idx_feature_edges_to     - feature_edges(to_feature_id)
idx_memos_scope          - memos(scope, scope_id)
idx_memos_priority       - memos(priority)
idx_skeletons_feature    - skeletons(feature_id)
idx_skeletons_layer      - skeletons(layer)
idx_project_experts_project - project_experts(project_id)
```

---

## 검증 결과

- **빌드**: 성공
- **테스트**: 모두 통과 (0.890s)

---

## 새 명령어 사용법

```bash
# DB 버전 확인
clari db version

# 마이그레이션 실행
clari db migrate

# 롤백
clari db rollback --version 4

# 백업 생성
clari db backup
```

---

*Implementation Report v0.0.4*
