# TTY Handover: Usage Scenarios

> **버전**: v0.0.1

---

## 시나리오 1: 정상 Task 실행

```
[Claritask] Starting Task 42: createComment
   Target: services/comment_service.py

╭─────────────────────────────────────────────╮
│  Claude Code                                │
│                                             │
│  > Reading services/comment_service.py...   │
│  > Implementing createComment function...   │
│  > Running pytest test_comment.py::test_... │
│  > ✓ Test passed!                           │
│                                             │
│  Implemented:                               │
│  - Content validation (1-1000 chars)        │
│  - DB insert with user_id, post_id          │
│  - Return created Comment object            │
│                                             │
│  /exit                                      │
╰─────────────────────────────────────────────╯

✅ [Claritask] Task Session Ended.
🔍 [Claritask] Verifying...
🎉 Verification Passed!
[Claritask] Task 42 completed. Moving to Task 43...
```

---

## 시나리오 2: 테스트 실패 후 자동 수정

```
[Claritask] Starting Task 42: createComment
   Target: services/comment_service.py

╭─────────────────────────────────────────────╮
│  Claude Code                                │
│                                             │
│  > Reading services/comment_service.py...   │
│  > Implementing createComment function...   │
│  > Running pytest test_comment.py::test_... │
│  > ❌ Test failed: ValidationError          │
│                                             │
│  > Analyzing error...                       │
│  > Error: content length not validated      │
│  > Editing line 23-25...                    │
│  > Running pytest again...                  │
│  > ✓ Test passed!                           │
│                                             │
│  Fixed: Added content length validation.    │
│  /exit                                      │
╰─────────────────────────────────────────────╯

✅ [Claritask] Task Session Ended.
🔍 [Claritask] Verifying...
🎉 Verification Passed!
[Claritask] Task 42 completed. Moving to Task 43...
```

---

## 시나리오 3: 사용자 개입 필요

```
[Claritask] Starting Task 55: paymentProcess
   Target: services/payment_service.py

╭─────────────────────────────────────────────╮
│  Claude Code                                │
│                                             │
│  > Reading services/payment_service.py...   │
│  > Running pytest test_payment.py::test_... │
│  > ❌ Error: Missing environment variable   │
│  > STRIPE_API_KEY not set                   │
│                                             │
│  I need the Stripe API key to proceed.      │
│  Please set the environment variable:       │
│  export STRIPE_API_KEY=sk_test_...          │
│                                             │
│  User: export STRIPE_API_KEY=sk_test_xxx    │ ← 사용자 개입
│                                             │
│  > Retrying...                              │
│  > ✓ Test passed!                           │
│  /exit                                      │
╰─────────────────────────────────────────────╯

✅ [Claritask] Task Session Ended.
```

---

## 시나리오 4: 해결 불가 (최대 시도 초과)

```
[Claritask] Starting Task 60: complexAlgorithm
   Target: services/algo_service.py

╭─────────────────────────────────────────────╮
│  Claude Code                                │
│                                             │
│  > Running test...                          │
│  > ❌ Test failed: Timeout                  │
│  > Attempt 1/3: Optimizing algorithm...     │
│  > Running test...                          │
│  > ❌ Test failed: Timeout                  │
│  > Attempt 2/3: Trying different approach...│
│  > Running test...                          │
│  > ❌ Test failed: Timeout                  │
│  > Attempt 3/3: Unable to resolve           │
│                                             │
│  BLOCKER: The algorithm requires O(n²)      │
│  complexity but the test expects O(n log n).│
│  This may need architectural changes.       │
│                                             │
│  /exit                                      │
╰─────────────────────────────────────────────╯

⚠️ [Claritask] Task Session Ended.
🔍 [Claritask] Verifying...
❌ Verification Failed.
[Claritask] Task 60 marked as FAILED.
[Claritask] Stopping execution. Please review Task 60.
```

---

## 시나리오 5: Phase 1 요구사항 수립

```
$ clari init my-shop "온라인 쇼핑몰"

[Claude Code 세션 시작]

╭─────────────────────────────────────────────╮
│  Claude Code                                │
│                                             │
│  안녕하세요! 온라인 쇼핑몰 프로젝트를 시작합니다. │
│                                             │
│  다음 기능들을 제안합니다:                    │
│  1. 회원 관리 (가입/로그인/프로필)            │
│  2. 상품 관리 (등록/수정/삭제)               │
│  3. 장바구니                                 │
│  4. 주문/결제                                │
│  5. 배송 추적                                │
│  6. 리뷰/평점                                │
│                                             │
│  추가하거나 수정할 기능이 있나요?              │
│                                             │
│  User: 쿠폰 기능 추가해줘                     │
│                                             │
│  좋습니다! 쿠폰 기능을 추가했습니다.           │
│                                             │
│  $ clari feature add '{"name": "coupon"...}' │
│                                             │
│  User: 좋아. 개발 시작해                      │
│                                             │
│  Phase 2를 시작합니다.                        │
│  $ clari project start                       │
│                                             │
╰─────────────────────────────────────────────╯

[Claude Code 종료]
[clari project start 실행됨]

[Claritask] Starting Phase 2: Auto Execution
[Claritask] Planning...
[Claritask] Generated 45 tasks
[Claritask] Starting Task 1...
```

---

## 시나리오 요약

| 시나리오 | 결과 | 사용자 개입 |
|----------|------|------------|
| 정상 실행 | 성공 | 없음 |
| 테스트 실패 → 자동 수정 | 성공 | 없음 |
| 환경 변수 누락 | 성공 | 필요 |
| 해결 불가 | 실패 | 검토 필요 |
| 요구사항 수립 | 성공 | 대화 참여 |

---

*TTY Handover Specification v0.0.1 - 2026-02-03*
