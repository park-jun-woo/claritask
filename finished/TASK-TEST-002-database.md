# TASK-TEST-002: 데이터베이스 레이어 테스트

## 테스트 대상
`internal/db/db.go`

## 테스트 코드 위치
`test/db_test.go`

## 테스트 시나리오

### 1. DB 연결 테스트
- Open 함수로 임시 DB 생성
- 디렉토리 자동 생성 확인
- Close 함수 정상 동작

### 2. 마이그레이션 테스트
- Migrate 함수 실행
- 모든 테이블 생성 확인 (projects, phases, tasks, context, tech, design, state, memos)
- 멱등성 확인 (2번 실행해도 에러 없음)

### 3. 유틸리티 함수 테스트
- TimeNow 함수 ISO 8601 포맷 확인
- ParseTime 함수 정상 파싱 확인
- ParseTime 잘못된 포맷 에러 확인

### 4. Foreign Key 테스트
- PRAGMA foreign_keys = ON 확인

## 완료 기준
- [ ] DB 연결/종료 테스트
- [ ] 마이그레이션 테스트
- [ ] 유틸리티 함수 테스트
- [ ] 테스트 통과
