# TASK-DEV-087: FDL Data Layer 구조화

## 개요

FDL 스펙의 Data Layer (모델 정의) 완전 구현

## 스펙 요구사항 (FDL/02-A-DataLayer.md)

### 필드 타입
```
uuid, string, text, int, bigint, float, decimal, boolean,
datetime, date, time, json, blob, enum(values)
```

### 타입 옵션
```
string(50), decimal(10,2), enum(admin,user,guest)
```

### 제약조건
```
pk, fk, required, unique, default, auto, index,
nullable, check, onDelete, onUpdate
```

### 인덱스
```yaml
indexes:
  - fields: [email]
    unique: true
  - fields: [created_at, status]
    name: idx_created_status
```

### 관계
```yaml
relations:
  - hasMany: Post
    foreignKey: author_id
  - belongsTo: Department
    as: department
```

### 패턴
```yaml
patterns:
  - timestamps    # created_at, updated_at 자동
  - softDelete    # deleted_at 자동
```

## 현재 상태

```go
type FDLModel struct {
    Name   string
    Table  string
    Fields []FDLField
}

type FDLField struct {
    Name        string
    Type        string
    Constraints string  // 파싱 안 됨
}
```

## 작업 내용

### 1. 구조체 확장 (fdl_service.go)

```go
type FDLModel struct {
    Name        string
    Table       string
    Description string
    Fields      []FDLField
    Indexes     []FDLIndex
    Relations   []FDLRelation
    Patterns    []string
}

type FDLField struct {
    Name        string
    Type        string            // 기본 타입
    TypeOption  string            // 50, (10,2), (admin,user)
    Constraints FDLFieldConstraint
    Description string
}

type FDLFieldConstraint struct {
    IsPK       bool
    IsFK       bool
    FKRef      string  // 참조 테이블.컬럼
    IsRequired bool
    IsUnique   bool
    IsAuto     bool
    IsIndex    bool
    IsNullable bool
    Default    string
    Check      string
    OnDelete   string
    OnUpdate   string
}

type FDLIndex struct {
    Fields []string
    Unique bool
    Name   string
    Where  string  // 부분 인덱스
}

type FDLRelation struct {
    Type       string  // hasOne, hasMany, belongsTo, belongsToMany
    Target     string  // 대상 모델
    ForeignKey string
    As         string  // 별칭
    Through    string  // 중간 테이블
}
```

### 2. 타입 파싱 함수

```go
func parseFieldType(typeStr string) (baseType, option string) {
    // "string(50)" -> "string", "50"
    // "decimal(10,2)" -> "decimal", "10,2"
    // "enum(admin,user)" -> "enum", "admin,user"
}

func parseConstraints(constraintStr string) FDLFieldConstraint {
    // "pk, required, unique" -> FDLFieldConstraint{IsPK: true, IsRequired: true, IsUnique: true}
    // "fk(users.id), onDelete: cascade" -> FDLFieldConstraint{IsFK: true, FKRef: "users.id", OnDelete: "cascade"}
}
```

### 3. 검증 함수 확장

```go
func validateModel(model *FDLModel, allModels []*FDLModel) []error {
    errors := []error{}

    // 타입 검증
    validTypes := []string{"uuid", "string", "text", "int", "bigint", ...}
    for _, field := range model.Fields {
        if !contains(validTypes, field.Type) {
            errors = append(errors, fmt.Errorf("invalid type: %s", field.Type))
        }
    }

    // FK 참조 검증
    for _, field := range model.Fields {
        if field.Constraints.IsFK {
            if !modelExists(field.Constraints.FKRef, allModels) {
                errors = append(errors, fmt.Errorf("FK reference not found: %s", field.Constraints.FKRef))
            }
        }
    }

    // 관계 검증
    for _, rel := range model.Relations {
        if !modelExists(rel.Target, allModels) {
            errors = append(errors, fmt.Errorf("relation target not found: %s", rel.Target))
        }
    }

    return errors
}
```

## 완료 조건

- [ ] FDLModel 구조체 확장
- [ ] FDLField 구조체 확장 (FDLFieldConstraint)
- [ ] FDLIndex, FDLRelation 구조체 추가
- [ ] 타입 파싱 함수 구현
- [ ] 제약조건 파싱 함수 구현
- [ ] 타입 검증 구현
- [ ] FK 참조 검증 구현
- [ ] 관계 검증 구현
- [ ] 테스트 작성
