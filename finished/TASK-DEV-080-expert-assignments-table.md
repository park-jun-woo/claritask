# TASK-DEV-080: expert_assignments 테이블 추가

## 개요

specs/DB/02-C-Content.md에 정의된 expert_assignments 테이블 구현

## 스펙 정의

```sql
CREATE TABLE expert_assignments (
    expert_id INTEGER NOT NULL,
    feature_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (expert_id, feature_id),
    FOREIGN KEY (expert_id) REFERENCES experts(id),
    FOREIGN KEY (feature_id) REFERENCES features(id)
);

CREATE INDEX idx_expert_assignments_feature ON expert_assignments(feature_id);
```

## 현재 상태

- project_experts 테이블만 존재 (project-expert 연결)
- feature-expert 연결 없음

## 작업 내용

### 1. DB 스키마 추가 (db.go)

```sql
CREATE TABLE IF NOT EXISTS expert_assignments (
    expert_id TEXT NOT NULL,  -- 현재 experts.id가 TEXT이므로
    feature_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (expert_id, feature_id),
    FOREIGN KEY (expert_id) REFERENCES experts(id),
    FOREIGN KEY (feature_id) REFERENCES features(id)
);

CREATE INDEX IF NOT EXISTS idx_expert_assignments_feature ON expert_assignments(feature_id);
```

### 2. Model 추가 (models.go)

```go
type ExpertAssignment struct {
    ExpertID  string
    FeatureID int64
    CreatedAt time.Time
}
```

### 3. Service 추가 (expert_service.go)

```go
func AssignExpertToFeature(db *db.DB, expertID string, featureID int64) error
func UnassignExpertFromFeature(db *db.DB, expertID string, featureID int64) error
func GetAssignedExpertsByFeature(db *db.DB, featureID int64) ([]model.ExpertInfo, error)
```

### 4. Cmd 추가 (expert.go)

기존 assign/unassign에 --feature 플래그 추가:
```bash
clari expert assign <expert-id> --feature <feature-id>
clari expert unassign <expert-id> --feature <feature-id>
```

## 설계 결정

- project_experts 테이블 유지 (project 레벨 할당)
- expert_assignments 테이블 추가 (feature 레벨 할당)
- task pop 시 feature의 할당된 expert들도 manifest에 포함

## 완료 조건

- [ ] expert_assignments 테이블 생성
- [ ] ExpertAssignment 모델 추가
- [ ] feature 레벨 expert 할당/해제 함수
- [ ] --feature 플래그 추가
- [ ] 테스트 작성
