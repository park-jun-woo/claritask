# VSCode Extension CLI 호환성

> **버전**: v0.0.4

## 확장자 변경 마이그레이션

```bash
# 기존 프로젝트 마이그레이션
mv .claritask/db .claritask/db.clt
```

---

## clari CLI 수정 사항

1. DB 경로 변경: `.claritask/db` → `.claritask/db.clt`
2. WAL 모드 기본 활성화
3. version 컬럼 마이그레이션 추가

---

## Context/Tech/Design 편집

- JSON 에디터 또는 폼 기반 UI
- 스키마 검증

---

*Claritask VSCode Extension Spec v0.0.4 - 2026-02-03*
