# TASK-DEV-028: Python 스켈레톤 생성기

## 개요
- **파일**: `scripts/skeleton_generator.py`
- **유형**: 신규
- **우선순위**: High
- **Phase**: 2 (FDL 시스템)
- **예상 LOC**: ~600

## 목적
FDL YAML을 파싱하여 스켈레톤 코드 자동 생성

## 작업 내용

### 1. CLI 인터페이스

```bash
python3 scripts/skeleton_generator.py \
    --fdl features/comment_system.fdl.yaml \
    --output-dir . \
    --backend python \
    --frontend react \
    [--force] \
    [--dry-run]
```

**인자**:
- `--fdl`: FDL 파일 경로 (필수)
- `--output-dir`: 출력 디렉토리 (기본: `.`)
- `--backend`: 백엔드 타입 (python, go, node)
- `--frontend`: 프론트엔드 타입 (react, vue, none)
- `--force`: 기존 파일 덮어쓰기
- `--dry-run`: 생성될 파일 목록만 출력

### 2. FDL 파서

```python
import yaml
from dataclasses import dataclass
from typing import List, Dict, Optional

@dataclass
class FDLModel:
    name: str
    table: str
    fields: List[Dict]

@dataclass
class FDLService:
    name: str
    desc: str
    input: Dict
    output: Optional[str]
    steps: List[str]

@dataclass
class FDLAPI:
    path: str
    method: str
    use: str
    request: Optional[Dict]
    response: Dict

@dataclass
class FDLSpec:
    feature: str
    description: str
    models: List[FDLModel]
    service: List[FDLService]
    api: List[FDLAPI]
    ui: List[Dict]

def parse_fdl(file_path: str) -> FDLSpec:
    """FDL 파일 파싱"""
    pass
```

### 3. Model 스켈레톤 생성 (Python)

```python
def generate_model_python(model: FDLModel, output_dir: str) -> str:
    """
    생성 예시:
    # models/comment.py
    from sqlalchemy import Column, String, DateTime, ForeignKey
    from app.db import Base

    class Comment(Base):
        __tablename__ = "comments"

        id = Column(String, primary_key=True)
        content = Column(String, nullable=False)
        post_id = Column(String, ForeignKey("posts.id"))
        user_id = Column(String, ForeignKey("users.id"))
        created_at = Column(DateTime, server_default=func.now())

        # TODO: Add relationships if needed
    """
    pass
```

### 4. Service 스켈레톤 생성 (Python)

```python
def generate_service_python(feature: str, services: List[FDLService], output_dir: str) -> str:
    """
    생성 예시:
    # services/comment_system_service.py
    \"\"\"
    comment_system Service
    Auto-generated from FDL. DO NOT modify function signatures.
    \"\"\"
    from typing import List
    from uuid import UUID
    from app.models.comment import Comment

    async def createComment(userId: UUID, postId: UUID, content: str) -> Comment:
        \"\"\"
        댓글 생성 및 알림 발송

        Steps (from FDL):
        - validate: "content 길이가 1자 이상 1000자 이하인지 검증"
        - db: "INSERT INTO comments (user_id, post_id, content)"
        - return: "생성된 Comment 객체"
        \"\"\"
        # TODO: 위 Steps를 구현하세요
        raise NotImplementedError("createComment not implemented")
    """
    pass
```

### 5. API 스켈레톤 생성 (Python/FastAPI)

```python
def generate_api_python(feature: str, apis: List[FDLAPI], output_dir: str) -> str:
    """
    생성 예시:
    # api/comment_system_api.py
    from fastapi import APIRouter, HTTPException
    from pydantic import BaseModel
    from app.services import comment_system_service

    router = APIRouter(prefix="/posts/{postId}/comments", tags=["comments"])

    class CreateCommentRequest(BaseModel):
        content: str

    class CommentResponse(BaseModel):
        id: str
        content: str
        created_at: str

    @router.post("", status_code=201, response_model=CommentResponse)
    async def create_comment(postId: str, request: CreateCommentRequest):
        # TODO: Call service.createComment
        raise NotImplementedError("API handler not implemented")
    """
    pass
```

### 6. UI 스켈레톤 생성 (React/TSX)

```python
def generate_ui_react(uis: List[Dict], output_dir: str) -> List[str]:
    """
    생성 예시:
    // components/CommentSection.tsx
    import React, { useState, useEffect } from 'react';

    interface Comment {
      id: string;
      content: string;
    }

    interface CommentSectionProps {
      postId: string;
    }

    export const CommentSection: React.FC<CommentSectionProps> = ({ postId }) => {
      const [comments, setComments] = useState<Comment[]>([]);
      const [newComment, setNewComment] = useState('');

      useEffect(() => {
        // TODO: API.GET /posts/{postId}/comments -> set comments
      }, [postId]);

      const handleSubmit = async () => {
        // TODO: API.POST /posts/{postId}/comments
      };

      return (
        <div>
          {/* TODO: Render comments and form */}
        </div>
      );
    };
    """
    pass
```

### 7. 출력 구조

```python
def generate_all(spec: FDLSpec, output_dir: str, backend: str, frontend: str, force: bool):
    """
    생성 구조:
    {output_dir}/
    ├── models/
    │   └── {feature}.py
    ├── services/
    │   └── {feature}_service.py
    ├── api/
    │   └── {feature}_api.py
    └── components/
        └── {Component}.tsx
    """
    pass
```

### 8. JSON 출력 (Go와 통신용)

```python
import json
import hashlib

def main():
    # ... 생성 로직

    result = {
        "generated_files": [
            {"path": "models/comment.py", "layer": "model", "checksum": "..."},
            {"path": "services/comment_system_service.py", "layer": "service", "checksum": "..."}
        ],
        "errors": []
    }
    print(json.dumps(result))
```

## 의존성
- Python 3.8+
- PyYAML

## 완료 기준
- [ ] FDL 파싱 구현됨
- [ ] Python Model 생성 구현됨
- [ ] Python Service 생성 구현됨
- [ ] Python API (FastAPI) 생성 구현됨
- [ ] React UI 생성 구현됨
- [ ] JSON 출력 형식 준수
- [ ] --dry-run 옵션 동작
- [ ] 단위 테스트 작성됨
