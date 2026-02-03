# 워크플로우 예시

> **버전**: v0.0.3

## 1. 프로젝트 초기화

```bash
# 프로젝트 생성
clari init blog-api "Developer blogging platform"
cd blog-api

# 필수 설정 확인
clari required

# 전체 설정
clari project set '{
  "name": "Blog Platform",
  "context": {"project_name": "Blog Platform", "description": "Developer blogging"},
  "tech": {"backend": "FastAPI", "frontend": "React", "database": "PostgreSQL"},
  "design": {"architecture": "Monolithic", "auth_method": "JWT", "api_style": "RESTful"}
}'
```

---

## 2. Planning

```bash
# 플래닝 모드 시작
clari project plan

# Feature 생성
clari feature add '{"name": "user_auth", "description": "사용자 인증 시스템"}'
clari feature add '{"name": "blog_posts", "description": "블로그 포스트 관리"}'

# Task 추가
clari task push '{"feature_id": 1, "title": "user_table_sql", "content": "CREATE TABLE users..."}'
clari task push '{"feature_id": 1, "title": "user_model", "content": "User 모델 구현"}'
```

---

## 3. Execution

```bash
# 실행 시작
clari project start

# Task 가져오기 (자동으로 doing 상태로 변경)
clari task pop

# 작업 수행 후 완료
clari task complete 1 '{"result": "users 테이블 생성 완료"}'

# 다음 Task
clari task pop
clari task complete 2 '{"result": "User 모델 구현 완료"}'

# 진행 상황 확인
clari task status
```

---

## 4. Memo 활용

```bash
# 중요한 발견 저장 (priority 1)
clari memo set jwt_security '{"value": "Use httpOnly cookies", "priority": 1}'

# 다음 task pop 시 manifest에 자동 포함됨
clari task pop
```

---

*Claritask Commands Reference v0.0.3 - 2026-02-03*
