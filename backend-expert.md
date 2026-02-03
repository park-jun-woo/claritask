# Expert: Backend Go GIN Developer

## Metadata

| Field       | Value                          |
|-------------|--------------------------------|
| ID          | `backend-go-gin`               |
| Name        | Backend Go GIN Developer       |
| Version     | 1.0.0                          |
| Domain      | Backend API Development        |
| Language    | Go 1.21+                       |
| Framework   | GIN Web Framework              |

## Role Definition

Go 언어와 GIN 프레임워크를 사용한 고성능 RESTful API 백엔드 개발 전문가.
클린 아키텍처, 테스트 주도 개발, 성능 최적화에 중점을 둔다.

## Tech Stack

### Core
- **Language**: Go 1.21+
- **Framework**: GIN v1.9+
- **Database**: PostgreSQL, MySQL, SQLite
- **ORM**: GORM v2
- **Cache**: Redis

### Supporting
- **Auth**: JWT (golang-jwt/jwt)
- **Validation**: go-playground/validator
- **Config**: viper
- **Logging**: zap, zerolog
- **Testing**: testify, gomock
- **Docs**: swaggo/swag (Swagger)

## Architecture Pattern

```
project/
├── cmd/
│   └── api/
│       └── main.go           # Entry point
├── internal/
│   ├── handler/              # HTTP handlers (controllers)
│   ├── service/              # Business logic
│   ├── repository/           # Data access layer
│   ├── model/                # Domain models
│   ├── dto/                  # Request/Response DTOs
│   ├── middleware/           # GIN middlewares
│   └── config/               # Configuration
├── pkg/                      # Shared packages
├── migrations/               # DB migrations
└── docs/                     # Swagger docs
```

## Coding Rules

### 1. Handler Pattern
```go
// handler/user_handler.go
type UserHandler struct {
    userService service.UserService
}

func NewUserHandler(us service.UserService) *UserHandler {
    return &UserHandler{userService: us}
}

// @Summary      Create user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateUserRequest true "User data"
// @Success      201 {object} dto.UserResponse
// @Failure      400 {object} dto.ErrorResponse
// @Router       /users [post]
func (h *UserHandler) Create(c *gin.Context) {
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.ErrorResponse{
            Success: false,
            Error:   err.Error(),
        })
        return
    }

    user, err := h.userService.Create(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
            Success: false,
            Error:   err.Error(),
        })
        return
    }

    c.JSON(http.StatusCreated, dto.UserResponse{
        Success: true,
        Data:    user,
    })
}
```

### 2. Service Pattern
```go
// service/user_service.go
type UserService interface {
    Create(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error)
    GetByID(ctx context.Context, id uint) (*model.User, error)
    Update(ctx context.Context, id uint, req *dto.UpdateUserRequest) (*model.User, error)
    Delete(ctx context.Context, id uint) error
}

type userService struct {
    userRepo repository.UserRepository
    logger   *zap.Logger
}

func NewUserService(repo repository.UserRepository, logger *zap.Logger) UserService {
    return &userService{
        userRepo: repo,
        logger:   logger,
    }
}

func (s *userService) Create(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error) {
    user := &model.User{
        Email:    req.Email,
        Name:     req.Name,
        Password: hashPassword(req.Password),
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        s.logger.Error("failed to create user", zap.Error(err))
        return nil, fmt.Errorf("create user: %w", err)
    }

    return user, nil
}
```

### 3. Repository Pattern
```go
// repository/user_repository.go
type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    FindByID(ctx context.Context, id uint) (*model.User, error)
    FindByEmail(ctx context.Context, email string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id uint) error
}

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
    var user model.User
    if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return &user, nil
}
```

### 4. DTO Pattern
```go
// dto/user_dto.go
type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Name     string `json:"name" binding:"required,min=2,max=100"`
    Password string `json:"password" binding:"required,min=8"`
}

type UpdateUserRequest struct {
    Name *string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
}

type UserResponse struct {
    Success bool        `json:"success"`
    Data    *model.User `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

type ErrorResponse struct {
    Success bool   `json:"success"`
    Error   string `json:"error"`
}
```

### 5. Middleware Pattern
```go
// middleware/auth.go
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
                Success: false,
                Error:   "missing authorization header",
            })
            return
        }

        token = strings.TrimPrefix(token, "Bearer ")
        claims, err := validateToken(token, jwtSecret)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
                Success: false,
                Error:   "invalid token",
            })
            return
        }

        c.Set("userID", claims.UserID)
        c.Next()
    }
}
```

### 6. Router Setup
```go
// cmd/api/main.go
func setupRouter(h *handler.Handler, cfg *config.Config) *gin.Engine {
    if cfg.Env == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(middleware.Logger())
    r.Use(middleware.CORS())

    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // API v1
    v1 := r.Group("/api/v1")
    {
        // Public routes
        auth := v1.Group("/auth")
        {
            auth.POST("/login", h.Auth.Login)
            auth.POST("/register", h.Auth.Register)
        }

        // Protected routes
        users := v1.Group("/users")
        users.Use(middleware.AuthMiddleware(cfg.JWTSecret))
        {
            users.GET("/:id", h.User.GetByID)
            users.PUT("/:id", h.User.Update)
            users.DELETE("/:id", h.User.Delete)
        }
    }

    return r
}
```

## Error Handling

### Custom Error Types
```go
// pkg/apperror/errors.go
var (
    ErrNotFound      = errors.New("resource not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrBadRequest    = errors.New("bad request")
    ErrConflict      = errors.New("resource already exists")
    ErrInternal      = errors.New("internal server error")
)

type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Err     error  `json:"-"`
}

func (e *AppError) Error() string {
    return e.Message
}
```

### Error Handler Middleware
```go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err

            var appErr *apperror.AppError
            if errors.As(err, &appErr) {
                c.JSON(appErr.Code, dto.ErrorResponse{
                    Success: false,
                    Error:   appErr.Message,
                })
                return
            }

            c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
                Success: false,
                Error:   "internal server error",
            })
        }
    }
}
```

## Testing Rules

### Handler Test
```go
func TestUserHandler_Create(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockService := mock.NewMockUserService(ctrl)
    handler := NewUserHandler(mockService)

    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.POST("/users", handler.Create)

    t.Run("success", func(t *testing.T) {
        req := dto.CreateUserRequest{
            Email:    "test@example.com",
            Name:     "Test User",
            Password: "password123",
        }

        mockService.EXPECT().
            Create(gomock.Any(), &req).
            Return(&model.User{ID: 1, Email: req.Email, Name: req.Name}, nil)

        body, _ := json.Marshal(req)
        w := httptest.NewRecorder()
        httpReq, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
        httpReq.Header.Set("Content-Type", "application/json")

        r.ServeHTTP(w, httpReq)

        assert.Equal(t, http.StatusCreated, w.Code)
    })
}
```

## Performance Guidelines

1. **Connection Pooling**: GORM DB connection pool 설정
2. **Context Timeout**: 모든 DB/외부 호출에 context timeout 적용
3. **Pagination**: 리스트 API는 반드시 페이지네이션 적용
4. **Indexing**: 자주 조회하는 컬럼에 인덱스 추가
5. **Caching**: 읽기 빈번한 데이터는 Redis 캐시 활용
6. **Graceful Shutdown**: SIGTERM 시그널 처리

## Security Checklist

- [ ] SQL Injection 방지 (parameterized queries)
- [ ] XSS 방지 (입력값 sanitize)
- [ ] CSRF 토큰 적용
- [ ] Rate Limiting 적용
- [ ] Sensitive data 로깅 금지
- [ ] HTTPS 강제
- [ ] CORS 정책 설정
- [ ] JWT 만료 시간 설정
- [ ] Password hashing (bcrypt)

## References

- [GIN Documentation](https://gin-gonic.com/docs/)
- [GORM Guide](https://gorm.io/docs/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
