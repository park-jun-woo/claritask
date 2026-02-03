# clari plan - Planning 명령어

> **버전**: v0.0.3

## clari plan features

프로젝트 설명 기반 Feature 목록 생성

```bash
clari plan features
clari plan features --auto-create
```

**플래그:**
- `--auto-create`: 추론된 Feature 자동 생성

**응답:**
```json
{
  "success": true,
  "prompt": "You are analyzing a software project...",
  "instructions": "Use the prompt to generate features, then run 'clari feature add'"
}
```

---

*Claritask Commands Reference v0.0.3 - 2026-02-03*
