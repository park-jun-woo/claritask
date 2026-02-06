# clari CLI 사용법

## 공통 옵션

모든 list 명령어에서 사용 가능:
- `-p <page>` - 페이지 번호
- `-n <size>` - 페이지 크기
- `--all` - 전체 조회 (페이징 없음)

## project (프로젝트 관리)

| 명령어 | 설명 |
|--------|------|
| `project list [--all]` | 프로젝트 목록 조회 |
| `project add <path> [desc]` | 기존 경로를 프로젝트로 등록 |
| `project create <id> [desc]` | 새 프로젝트 생성 |
| `project get [id]` | 프로젝트 상세 조회 |
| `project delete <id>` | 프로젝트 삭제 |
| `project switch <id>` | 프로젝트 전환 |
| `project switch none` | 글로벌 모드로 전환 |

## task (작업 관리)

| 명령어 | 설명 |
|--------|------|
| `task list [parent_id] [--all]` | 작업 목록 조회 |
| `task add <title> [--parent <id>] [--spec <spec>]` | 작업 추가 |
| `task get <id>` | 작업 상세 조회 |
| `task set <id> <field> <value>` | 작업 필드 수정 |
| `task delete <id>` | 작업 삭제 |
| `task plan [id]` | 단일 Task Plan 생성 |
| `task plan --all` | 전체 todo Task Plan 생성 |
| `task run [id]` | 단일 Task 실행 |
| `task run --all` | 전체 planned Task 실행 |
| `task cycle` | Plan + 실행 자동 순회 |

## message (메시지)

| 명령어 | 설명 |
|--------|------|
| `message list [--all]` | 메시지 목록 조회 |
| `message send <content>` | 메시지 전송 (Claude 실행) |
| `message get <id>` | 메시지 상세 조회 |
| `message status` | 메시지 처리 상태 |
| `message processing` | 처리 중인 메시지 조회 |
| `send <content>` | message send 단축 명령어 |

## spec (요구사항 명세서)

| 명령어 | 설명 |
|--------|------|
| `spec list [--all]` | 스펙 목록 조회 |
| `spec add <title>` | 스펙 추가 |
| `spec get <id>` | 스펙 상세 조회 |
| `spec set <id> <field> <value>` | 스펙 필드 수정 (title, content, status, priority) |
| `spec delete <id>` | 스펙 삭제 |

**status 값**: draft, review, approved, deprecated

## schedule (스케줄링)

| 명령어 | 설명 |
|--------|------|
| `schedule list [--all]` | 스케줄 목록 조회 |
| `schedule add <cron> <msg> [--project <id>] [--once]` | 스케줄 추가 |
| `schedule get <id>` | 스케줄 상세 조회 |
| `schedule set <id> project <id\|none>` | 프로젝트 변경 |
| `schedule delete <id>` | 스케줄 삭제 |
| `schedule enable <id>` | 스케줄 활성화 |
| `schedule disable <id>` | 스케줄 비활성화 |
| `schedule runs <schedule_id>` | 실행 기록 목록 |
| `schedule run <run_id>` | 실행 기록 상세 |

## 기타

| 명령어 | 설명 |
|--------|------|
| `status` | 현재 선택된 프로젝트 상태 |
