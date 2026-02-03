# TASK-DEV-040: Python Skeleton Generator 연동

## 개요
- **파일**: `internal/cmd/fdl.go`, `internal/service/skeleton_service.go`, `scripts/skeleton_generator.py`
- **유형**: 수정/완성
- **스펙 참조**: Claritask.md "스켈레톤 생성" 섹션

## 배경
현재 `clari fdl skeleton` 명령어에 TODO 주석이 있음:
```go
// TODO: Call Python skeleton generator (TASK-DEV-028)
// For now, just mark as generated and return the file list
```

실제 Python skeleton generator 호출 및 파일 생성이 필요함.

## 구현 내용

### 1. scripts/skeleton_generator.py 완성
```python
#!/usr/bin/env python3
"""
FDL to Skeleton Code Generator
Generates code skeletons from FDL YAML files.
"""

def generate_model(model_spec, tech_stack):
    """Generate model/entity code"""
    pass

def generate_service(service_spec, tech_stack):
    """Generate service functions with TODO markers"""
    pass

def generate_api(api_spec, tech_stack):
    """Generate API handlers/routes"""
    pass

def generate_ui(ui_spec, tech_stack):
    """Generate UI component skeletons"""
    pass
```

### 2. skeleton_service.go 수정
```go
// RunSkeletonGenerator executes Python skeleton generator
func RunSkeletonGenerator(database *db.DB, featureID int64, force bool) (*SkeletonResult, error) {
    // 1. Get FDL from database
    // 2. Get tech stack
    // 3. Write FDL to temp file
    // 4. Call Python script: python scripts/skeleton_generator.py <fdl_file> <output_dir> <tech_json>
    // 5. Parse output
    // 6. Update skeletons table
    // 7. Return result
}
```

### 3. fdl.go runFDLSkeleton 수정
- TODO 주석 제거
- RunSkeletonGenerator 함수 호출
- 생성된 파일 목록 반환

### 4. 지원 언어/프레임워크
- Backend: Go, Python (FastAPI), Node.js (Express)
- Frontend: React, Vue
- 언어별 템플릿 분리

## 완료 기준
- [ ] skeleton_generator.py 완성
- [ ] RunSkeletonGenerator Go 함수 완성
- [ ] clari fdl skeleton 실제 파일 생성
- [ ] 다양한 tech stack 지원 테스트
