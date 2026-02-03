# 에러 처리 및 미구현 명령어

> **버전**: v0.0.3

## 에러 처리

모든 명령어는 다음 형식으로 에러 반환:

```json
{
  "success": false,
  "error": "에러 메시지"
}
```

**일반적인 에러:**
- `open database: ...` - DB 연결 실패
- `parse JSON: ...` - JSON 파싱 실패
- `missing required field: ...` - 필수 필드 누락
- `task status must be '...' to ...` - 잘못된 상태 전이
- `memo not found` - 메모 없음

---

## 미구현 명령어 (향후 계획)

다음 명령어들은 향후 구현 예정입니다:

```bash
# FDL 심층 검증
clari fdl verify <feature_id> --strict    # 엄격한 검증 모드

# 자동 Edge 추가
clari edge infer --feature <id> --auto-add
```

---

*Claritask Commands Reference v0.0.3 - 2026-02-03*
