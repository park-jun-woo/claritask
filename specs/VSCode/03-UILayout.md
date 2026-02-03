# VSCode Extension UI 레이아웃

> **현재 버전**: v0.0.5 ([변경이력](../HISTORY.md))

---

## 탭 구조

```
┌─────────────────────────────────────────────────────────┐
│  Claritask: my-project                            [⟳]  │
├─────────────────────────────────────────────────────────┤
│  [Project]  [Messages]  [Features]  [Tasks]  [Experts]  │
├─────────────────────────────────────────────────────────┤
```

5개의 메인 탭으로 구성:
- **Project**: 프로젝트 정보, Context, Tech, Design 조회/편집
- **Messages**: 수정 요청 메시지 목록 및 관리 (CLI 연동)
- **Features**: Feature 목록 및 관리
- **Tasks**: Task 목록 및 Canvas 뷰
- **Experts**: Expert 목록, 마크다운 렌더링 뷰, 파일 편집

---

*Claritask VSCode Extension Spec v0.0.5*
