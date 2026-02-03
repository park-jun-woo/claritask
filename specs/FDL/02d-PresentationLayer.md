# ClariSpec: PRESENTATION LAYER (ui)

> **버전**: v0.0.1

## 개요

PRESENTATION LAYER는 UI 컴포넌트와 상태 관리를 정의합니다. 이 계층은 사용자 인터페이스의 구조, 데이터 바인딩, 이벤트 처리를 명세하며, INTERFACE LAYER의 API와 연결(Wiring)됩니다.

---

## 기본 구조

```yaml
ui:
  - component: <ComponentName>    # 컴포넌트 이름 (PascalCase)
    type: <Page|Organism|Molecule|Atom>
    description: <desc>           # 컴포넌트 설명
    parent: <ParentComponent>     # 부모 컴포넌트 (선택)
    props: { ... }                # 외부에서 받는 props
    state: [ ... ]                # 내부 상태
    computed: { ... }             # 계산된 값
    init: [ ... ]                 # 초기화 로직
    view: [ ... ]                 # UI 요소 정의
    methods: { ... }              # 이벤트 핸들러
```

---

## 컴포넌트 타입 (Atomic Design)

| 타입 | 설명 | 특징 | 예시 |
|------|------|------|------|
| `Page` | 전체 페이지 | 라우팅 대상, 레이아웃 포함 | `HomePage`, `UserProfilePage` |
| `Template` | 페이지 레이아웃 | 여러 Organism 배치 | `DashboardTemplate` |
| `Organism` | 독립적 기능 블록 | 자체 상태, API 호출 가능 | `CommentSection`, `UserCard` |
| `Molecule` | 여러 Atom 조합 | 재사용 가능한 UI 그룹 | `SearchForm`, `NavItem` |
| `Atom` | 최소 단위 요소 | 단일 목적, 무상태 | `Button`, `Input`, `Avatar` |

---

## Props 정의

```yaml
ui:
  - component: UserCard
    type: Molecule
    props:
      # 필수 props
      user: User (required)

      # 선택 props (기본값)
      showAvatar: boolean (default: true)
      size: enum(small,medium,large) (default: "medium")

      # 이벤트 핸들러 props
      onClick: function (optional)
      onFollow: function (optional)
```

### Props 타입

| 타입 | 설명 | 예시 |
|------|------|------|
| `string` | 문자열 | `title: string` |
| `number` | 숫자 | `count: number` |
| `boolean` | 불리언 | `isActive: boolean` |
| `Array<T>` | 배열 | `items: Array<Post>` |
| `<ModelName>` | 모델 참조 | `user: User` |
| `function` | 콜백 함수 | `onClick: function` |
| `ReactNode` | 자식 컴포넌트 | `children: ReactNode` |

---

## State 정의

```yaml
ui:
  - component: PostList
    type: Organism
    state:
      # 기본 상태
      - posts: Array<Post> (default: [])
      - loading: boolean (default: true)
      - error: string (default: null)

      # 페이지네이션 상태
      - page: int (default: 1)
      - hasMore: boolean (default: true)

      # 폼 상태
      - searchQuery: string (default: "")
      - filters:
          status: enum(all,published,draft) (default: "all")
          sortBy: string (default: "created_at")
```

---

## Computed (계산된 값)

```yaml
ui:
  - component: CartSummary
    type: Molecule
    state:
      - items: Array<CartItem>
      - discountCode: string
    computed:
      # 단순 계산
      itemCount: "items.length"

      # 복잡한 계산
      subtotal: "items.reduce((sum, item) => sum + item.price * item.quantity, 0)"

      # 조건부 계산
      discount: "discountCode ? calculateDiscount(subtotal, discountCode) : 0"

      # 최종 계산
      total: "subtotal - discount"

      # 불리언 계산
      isEmpty: "items.length === 0"
      canCheckout: "!isEmpty && total > 0"
```

---

## Init (초기화)

```yaml
ui:
  - component: UserProfile
    type: Page
    props:
      userId: uuid (required)
    state:
      - user: User (default: null)
      - posts: Array<Post> (default: [])
      - loading: boolean (default: true)
    init:
      # API 호출 후 상태 설정
      - call: API.GET /users/{userId} -> set user
      - call: API.GET /users/{userId}/posts -> set posts
      - set: loading = false

      # 조건부 초기화
      - if: "!userId"
        redirect: /login

      # 병렬 로딩
      - parallel:
          - call: API.GET /users/{userId} -> set user
          - call: API.GET /users/{userId}/posts -> set posts
```

---

## View 정의

### 기본 요소

```yaml
view:
  # 텍스트 요소
  - Text: "제목"
    variant: h1

  - Text: "{user.name}"             # 동적 텍스트
    variant: body1

  # 입력 요소
  - Input: "이메일"
    type: email
    bind: email                      # 양방향 바인딩
    placeholder: "이메일을 입력하세요"

  - Textarea: "내용"
    bind: content
    rows: 5

  - Select: "상태"
    bind: status
    options:
      - { value: "draft", label: "임시저장" }
      - { value: "published", label: "발행" }

  # 버튼
  - Button: "저장"
    variant: primary
    action: methods.handleSubmit
    disabled: "{!isValid}"

  # 이미지
  - Image: "{user.avatarUrl}"
    alt: "{user.name}의 프로필"
    fallback: "/default-avatar.png"
```

### 조건부 렌더링

```yaml
view:
  # if 조건
  - if: "{loading}"
    then:
      - Spinner: null
    else:
      - UserCard: null
        props: { user }

  # show/hide (DOM 유지)
  - Button: "삭제"
    show: "{user.role === 'admin'}"
```

### 반복 렌더링

```yaml
view:
  # 리스트 렌더링
  - for: post in posts
    key: post.id
    render:
      - PostCard: null
        props: { post }
        onClick: "methods.handlePostClick(post.id)"

  # 빈 상태 처리
  - if: "{posts.length === 0}"
    then:
      - EmptyState: "게시글이 없습니다"
        action:
          label: "첫 게시글 작성하기"
          onClick: methods.navigateToCreate
```

### 레이아웃

```yaml
view:
  # Flexbox
  - Flex:
      direction: row
      justify: space-between
      align: center
      gap: 16
      children:
        - Text: "{title}"
        - Button: "편집"

  # Grid
  - Grid:
      columns: 3
      gap: 24
      children:
        - for: item in items
          render:
            - Card: null
              props: { item }

  # Stack (수직 Flex)
  - Stack:
      spacing: 16
      children:
        - Input: "이름"
        - Input: "이메일"
        - Button: "제출"
```

---

## Methods (이벤트 핸들러)

```yaml
ui:
  - component: CommentForm
    type: Molecule
    state:
      - content: string
      - submitting: boolean
    methods:
      # 기본 핸들러
      handleSubmit:
        - validate: "content 비어있지 않음"
        - set: submitting = true
        - call: API.POST /comments (body: { content })
        - onSuccess:
            - set: content = ""
            - call: props.onCommentAdded
        - onError:
            - show: toast("댓글 작성에 실패했습니다")
        - finally:
            - set: submitting = false

      # 간단한 핸들러
      handleClear:
        - set: content = ""

      # 파라미터 받는 핸들러
      handleDelete(commentId):
        - confirm: "정말 삭제하시겠습니까?"
        - call: API.DELETE /comments/{commentId}
        - onSuccess:
            - call: props.onCommentDeleted(commentId)
```

### Action 타입

| 액션 | 설명 | 예시 |
|------|------|------|
| `set:` | 상태 변경 | `set: loading = true` |
| `call:` | API/메서드 호출 | `call: API.POST /users` |
| `navigate:` | 페이지 이동 | `navigate: /users/{id}` |
| `show:` | UI 피드백 | `show: toast("성공")` |
| `validate:` | 검증 | `validate: "폼 유효성"` |
| `confirm:` | 확인 대화상자 | `confirm: "삭제하시겠습니까?"` |
| `emit:` | 이벤트 발생 | `emit: change(value)` |

---

## API Wiring

```yaml
view:
  # 버튼 클릭 시 API 호출
  - Button: "저장"
    action: API.POST /posts (body: { title, content })
    onSuccess: "navigate: /posts/{response.id}"
    onError: "show: toast(error.message)"

  # 폼 제출
  - Form:
      onSubmit: API.POST /users (body: formData)
      children:
        - Input: "이메일"
          bind: formData.email
        - Input: "비밀번호"
          type: password
          bind: formData.password
        - Button: "가입"
          type: submit
          loading: "{submitting}"
```

---

## 스타일링

```yaml
ui:
  - component: UserCard
    type: Molecule
    styles:
      # 기본 스타일
      container:
        padding: 16
        borderRadius: 8
        boxShadow: "0 2px 4px rgba(0,0,0,0.1)"

      # 조건부 스타일
      avatar:
        size: "{props.size === 'large' ? 64 : 40}"

      # 상태 기반 스타일
      name:
        color: "{isActive ? 'green' : 'gray'}"
    view:
      - Container:
          style: styles.container
          children:
            - Avatar: "{user.avatarUrl}"
              style: styles.avatar
            - Text: "{user.name}"
              style: styles.name
```

---

## 전체 예시

```yaml
ui:
  # 페이지 컴포넌트
  - component: PostListPage
    type: Page
    description: 블로그 게시글 목록 페이지
    state:
      - posts: Array<Post> (default: [])
      - loading: boolean (default: true)
      - error: string (default: null)
      - page: int (default: 1)
      - hasMore: boolean (default: true)
      - searchQuery: string (default: "")
      - statusFilter: enum(all,published,draft) (default: "all")
    computed:
      filteredPosts: "posts.filter(post => statusFilter === 'all' || post.status === statusFilter)"
      isEmpty: "filteredPosts.length === 0 && !loading"
    init:
      - call: API.GET /posts?page=1 -> set posts
      - set: loading = false
    methods:
      handleSearch:
        - set: loading = true
        - call: API.GET /posts?search={searchQuery}&status={statusFilter}
        - onSuccess:
            - set: posts = response.data
            - set: hasMore = response.pagination.hasNext
        - finally:
            - set: loading = false

      handleLoadMore:
        - if: "!hasMore || loading"
          return: null
        - set: page = page + 1
        - call: API.GET /posts?page={page}
        - onSuccess:
            - set: posts = [...posts, ...response.data]
            - set: hasMore = response.pagination.hasNext

      handleCreatePost:
        - navigate: /posts/new
    view:
      # 헤더
      - PageHeader:
          title: "게시글"
          actions:
            - Button: "새 글 작성"
              variant: primary
              action: methods.handleCreatePost

      # 검색 및 필터
      - SearchFilterBar:
          children:
            - Input: "검색"
              bind: searchQuery
              onEnter: methods.handleSearch
            - Select: "상태"
              bind: statusFilter
              options:
                - { value: "all", label: "전체" }
                - { value: "published", label: "발행됨" }
                - { value: "draft", label: "임시저장" }
              onChange: methods.handleSearch

      # 로딩 상태
      - if: "{loading && posts.length === 0}"
        then:
          - LoadingSpinner: null

      # 에러 상태
      - if: "{error}"
        then:
          - ErrorMessage: "{error}"
            retry: methods.handleSearch

      # 빈 상태
      - if: "{isEmpty}"
        then:
          - EmptyState:
              title: "게시글이 없습니다"
              description: "첫 번째 게시글을 작성해보세요"
              action:
                label: "새 글 작성"
                onClick: methods.handleCreatePost

      # 게시글 목록
      - Grid:
          columns: 3
          gap: 24
          children:
            - for: post in filteredPosts
              key: post.id
              render:
                - PostCard: null
                  props: { post }

      # 더 보기 버튼
      - if: "{hasMore && !loading}"
        then:
          - Button: "더 보기"
            variant: secondary
            action: methods.handleLoadMore
            loading: "{loading}"

  # 게시글 카드 컴포넌트
  - component: PostCard
    type: Organism
    description: 게시글 카드 컴포넌트
    props:
      post: Post (required)
      onClick: function (optional)
    computed:
      formattedDate: "formatDate(post.publishedAt || post.createdAt)"
      excerpt: "post.excerpt || truncate(post.content, 150)"
    view:
      - Card:
          onClick: "{props.onClick ? props.onClick : () => navigate('/posts/' + post.id)}"
          hover: true
          children:
            # 썸네일
            - if: "{post.coverImage}"
              then:
                - Image: "{post.coverImage}"
                  aspectRatio: "16:9"

            # 내용
            - CardContent:
                children:
                  # 상태 뱃지
                  - if: "{post.status === 'draft'}"
                    then:
                      - Badge: "임시저장"
                        variant: warning

                  # 제목
                  - Text: "{post.title}"
                    variant: h3
                    lines: 2

                  # 발췌
                  - Text: "{excerpt}"
                    variant: body2
                    color: secondary
                    lines: 3

                  # 메타 정보
                  - Flex:
                      justify: space-between
                      align: center
                      children:
                        - Flex:
                            gap: 8
                            align: center
                            children:
                              - Avatar: "{post.author.avatarUrl}"
                                size: small
                              - Text: "{post.author.name}"
                                variant: caption
                        - Text: "{formattedDate}"
                          variant: caption
                          color: secondary

            # 푸터 (조회수, 좋아요)
            - CardFooter:
                children:
                  - Flex:
                      gap: 16
                      children:
                        - IconText:
                            icon: eye
                            text: "{post.viewCount}"
                        - IconText:
                            icon: heart
                            text: "{post.likeCount}"

  # 댓글 섹션 컴포넌트
  - component: CommentSection
    type: Organism
    description: 게시글 댓글 섹션
    props:
      postId: uuid (required)
    state:
      - comments: Array<Comment> (default: [])
      - loading: boolean (default: true)
      - newComment: string (default: "")
      - submitting: boolean (default: false)
    computed:
      commentCount: "comments.length"
      canSubmit: "newComment.trim().length > 0 && !submitting"
    init:
      - call: API.GET /posts/{postId}/comments -> set comments
      - set: loading = false
    methods:
      handleSubmit:
        - if: "!canSubmit"
          return: null
        - set: submitting = true
        - call: API.POST /posts/{postId}/comments (body: { content: newComment })
        - onSuccess:
            - set: comments = [response, ...comments]
            - set: newComment = ""
            - show: toast("댓글이 작성되었습니다")
        - onError:
            - show: toast(error.message, "error")
        - finally:
            - set: submitting = false

      handleDelete(commentId):
        - confirm: "댓글을 삭제하시겠습니까?"
        - call: API.DELETE /comments/{commentId}
        - onSuccess:
            - set: comments = comments.filter(c => c.id !== commentId)
            - show: toast("댓글이 삭제되었습니다")
    view:
      - Section:
          children:
            # 헤더
            - Text: "댓글 ({commentCount})"
              variant: h4

            # 댓글 작성 폼
            - CommentForm:
                children:
                  - Textarea: "댓글을 입력하세요"
                    bind: newComment
                    rows: 3
                    maxLength: 1000
                  - Flex:
                      justify: flex-end
                      children:
                        - Button: "댓글 작성"
                          action: methods.handleSubmit
                          loading: "{submitting}"
                          disabled: "{!canSubmit}"

            # 로딩 상태
            - if: "{loading}"
              then:
                - CommentSkeleton: null
                  count: 3

            # 댓글 목록
            - for: comment in comments
              key: comment.id
              render:
                - CommentItem:
                    props: { comment }
                    onDelete: "methods.handleDelete(comment.id)"

            # 빈 상태
            - if: "{!loading && comments.length === 0}"
              then:
                - Text: "아직 댓글이 없습니다. 첫 댓글을 작성해보세요!"
                  variant: body2
                  color: secondary
                  align: center
```

---

## 네이밍 컨벤션

| 항목 | 규칙 | 예시 |
|------|------|------|
| 컴포넌트명 | PascalCase | `UserCard`, `PostListPage` |
| Props | camelCase | `onClick`, `isActive` |
| State | camelCase | `isLoading`, `userData` |
| Methods | camelCase, handle 접두사 | `handleSubmit`, `handleClick` |
| 이벤트 Props | on 접두사 | `onClick`, `onChange`, `onSubmit` |

---

*ClariSpec FDL Specification v0.0.1 - 2026-02-03*
