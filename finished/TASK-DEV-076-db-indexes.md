# TASK-DEV-076: 인덱스 추가

## 개요

specs에 정의된 인덱스를 DB에 추가하여 쿼리 성능 최적화

## 대상 파일

- `cli/internal/db/db.go`

## 작업 내용

### 1. 인덱스 정의

Migrate() 함수에 인덱스 생성 쿼리 추가:

```sql
-- Core Tables 인덱스
CREATE INDEX IF NOT EXISTS idx_features_project ON features(project_id);
CREATE INDEX IF NOT EXISTS idx_features_status ON features(status);
CREATE INDEX IF NOT EXISTS idx_tasks_feature ON tasks(feature_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_task_edges_to ON task_edges(to_task_id);
CREATE INDEX IF NOT EXISTS idx_feature_edges_to ON feature_edges(to_feature_id);

-- Content Tables 인덱스
CREATE INDEX IF NOT EXISTS idx_memos_scope ON memos(scope, scope_id);
CREATE INDEX IF NOT EXISTS idx_memos_priority ON memos(priority);
CREATE INDEX IF NOT EXISTS idx_skeletons_feature ON skeletons(feature_id);
CREATE INDEX IF NOT EXISTS idx_skeletons_layer ON skeletons(layer);

-- Expert 인덱스
CREATE INDEX IF NOT EXISTS idx_experts_status ON experts(status);
CREATE INDEX IF NOT EXISTS idx_project_experts_project ON project_experts(project_id);
```

### 2. 마이그레이션으로 인덱스 추가

마이그레이션 v5로 인덱스 추가 (TASK-DEV-075 완료 후):

```go
func migrateV5(db *DB) error {
    indexes := []string{
        "CREATE INDEX IF NOT EXISTS idx_features_project ON features(project_id)",
        "CREATE INDEX IF NOT EXISTS idx_features_status ON features(status)",
        "CREATE INDEX IF NOT EXISTS idx_tasks_feature ON tasks(feature_id)",
        "CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)",
        "CREATE INDEX IF NOT EXISTS idx_task_edges_to ON task_edges(to_task_id)",
        "CREATE INDEX IF NOT EXISTS idx_feature_edges_to ON feature_edges(to_feature_id)",
        "CREATE INDEX IF NOT EXISTS idx_memos_scope ON memos(scope, scope_id)",
        "CREATE INDEX IF NOT EXISTS idx_memos_priority ON memos(priority)",
        "CREATE INDEX IF NOT EXISTS idx_skeletons_feature ON skeletons(feature_id)",
        "CREATE INDEX IF NOT EXISTS idx_skeletons_layer ON skeletons(layer)",
        "CREATE INDEX IF NOT EXISTS idx_experts_status ON experts(status)",
        "CREATE INDEX IF NOT EXISTS idx_project_experts_project ON project_experts(project_id)",
    }

    for _, idx := range indexes {
        if _, err := db.Exec(idx); err != nil {
            return err
        }
    }
    return nil
}
```

## 의존성

- TASK-DEV-075 완료 필요 (마이그레이션 시스템)

## 완료 조건

- [ ] 모든 인덱스 생성 쿼리 추가
- [ ] 마이그레이션 v5로 등록
- [ ] 기존 테스트 통과
