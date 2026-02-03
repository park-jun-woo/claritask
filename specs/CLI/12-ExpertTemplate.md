# Expert 템플릿 및 Manifest 연동

> **버전**: v0.0.3

## Expert 템플릿

`clari expert add` 실행 시 생성되는 기본 템플릿:

```markdown
# Expert: [Expert Name]

## Metadata

| Field       | Value                          |
|-------------|--------------------------------|
| ID          | `expert-id`                    |
| Name        | Expert Name                    |
| Version     | 1.0.0                          |
| Domain      | Domain Description             |
| Language    | Language Version               |
| Framework   | Framework Name                 |

## Role Definition

[전문가 역할 설명 - 한 문장]

## Tech Stack

### Core
- **Language**:
- **Framework**:
- **Database**:

### Supporting
- **Auth**:
- **Validation**:
- **Logging**:
- **Testing**:

## Architecture Pattern

[디렉토리 구조]

## Coding Rules

[패턴별 코드 템플릿]

## Error Handling

[에러 처리 규칙]

## Testing Rules

[테스트 코드 규칙]

## Security Checklist

- [ ] 보안 항목들

## References

- [문서 링크]
```

---

## Expert와 Task Manifest 연동

프로젝트에 Expert가 할당되면, `clari task pop` 응답의 manifest에 포함됩니다:

```json
{
  "success": true,
  "task": {...},
  "manifest": {
    "context": {...},
    "tech": {...},
    "design": {...},
    "experts": [
      {
        "id": "backend-go-gin",
        "name": "Backend Go GIN Developer",
        "content": "# Expert: Backend Go GIN Developer\n..."
      }
    ],
    "state": {...},
    "memos": [...]
  }
}
```

---

*Claritask Commands Reference v0.0.3 - 2026-02-03*
