# 로컬 개발환경 가이드라인

공통 기술 스택(PostgreSQL, Redis, Go-GIN, React) 기반 프로젝트의 로컬 개발환경 표준.

---

## 기술 스택

| 레이어 | 기술 | 비고 |
|--------|------|------|
| DB | PostgreSQL 15+ (Alpine) | Docker 컨테이너 |
| Cache | Redis 7+ (Alpine) | Docker 컨테이너 |
| Backend | Go 1.24+ / Gin | 로컬 빌드 실행 |
| Frontend | React 18+ / Vite / TypeScript | 로컬 dev server |

---

## 1. 프로젝트 구조

```
project/
├── docker-compose.yml          # DB, Redis
├── .env                        # 환경변수 (gitignore)
├── .env.example                # 환경변수 템플릿 (git 관리)
├── Makefile                    # 개발 명령어 통합
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/             # 환경변수 로드
│   │   ├── db/                 # DB 연결, 마이그레이션
│   │   ├── middleware/         # Auth, CORS, Logger
│   │   ├── api/                # HTTP 핸들러
│   │   ├── service/            # 비즈니스 로직
│   │   ├── repository/         # 데이터 접근
│   │   └── model/              # 데이터 모델
│   ├── .env                    # 백엔드 환경변수 (gitignore)
│   ├── .env.example
│   └── go.mod
└── frontend/
    ├── src/
    │   ├── api/                # API 클라이언트
    │   ├── components/
    │   ├── pages/
    │   ├── stores/             # 상태 관리
    │   ├── hooks/
    │   └── types/
    ├── vite.config.ts
    ├── package.json
    └── tsconfig.json
```

---

## 2. 인프라 (Docker Compose)

### docker-compose.yml

```yaml
services:
  postgres:
    image: postgres:15-alpine
    ports:
      - "127.0.0.1:${DB_PORT:-5432}:5432"
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - pgdata:/var/lib/postgresql/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER}"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "127.0.0.1:${REDIS_PORT:-6379}:6379"
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redisdata:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  pgdata:
  redisdata:
```

### 규칙

- **포트 바인딩**: 항상 `127.0.0.1`로 제한. `0.0.0.0` 금지.
- **볼륨**: named volume 사용으로 데이터 영속성 보장.
- **healthcheck**: 서비스 준비 상태 확인.
- **비밀번호**: `.env` 파일에서 주입. docker-compose.yml에 하드코딩 금지.

---

## 3. 포트 할당

IANA 동적/사설 포트 범위(49152~65535)를 사용한다.
프로젝트별 16개(0x10) 블록 단위로 할당하여 충돌을 방지한다.

### 블록 구조 (0xCXX0 ~ 0xCXXF, 16개)

```
+0x0  DB (PostgreSQL)
+0x1  Backend API
+0x2  Cache (Redis)
+0x3  Frontend Dev
+0x4~0xF  예비 (워커, gRPC, WebSocket 등)
```

### 할당 현황

| 블록 | 프로젝트 | DB | Backend | Redis | Frontend Dev |
|--------|---------|-------|---------|-------|-------------|
| 0xC000 | claribot | 49152 | 49153 | 49154 | 49155 |
| 0xC010 | gozip | 49168 | 49169 | 49170 | 49171 |
| 0xC020 | (예비) | 49184 | 49185 | 49186 | 49187 |
| 0xC030 | (예비) | 49200 | 49201 | 49202 | 49203 |

### 블록 계산

```
base = 0xC000 + (프로젝트 순번 × 0x10)
DB       = base + 0
Backend  = base + 1
Redis    = base + 2
Frontend = base + 3
```

최대 **256개 프로젝트** (0xC000 ~ 0xCFF0) 할당 가능.

---

## 4. 환경변수

### .env.example (프로젝트 루트, git 관리)

```bash
# 포트 블록 (프로젝트별 할당표 참고)
DB_PORT=49168
BACKEND_PORT=49169
REDIS_PORT=49170
FRONTEND_PORT=49171

# Database
DB_USER=appuser
DB_PASSWORD=change-me
DB_NAME=appdb

# Redis
REDIS_PASSWORD=change-me
```

### backend/.env.example (git 관리)

```bash
# Server
APP_PORT=49169
GIN_MODE=debug    # debug | release

# Database
DB_URL=postgres://appuser:change-me@localhost:49168/appdb?sslmode=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFE_MIN=30

# Redis
REDIS_URL=redis://:change-me@localhost:49170/0

# Auth
JWT_SECRET=change-me-min-32-chars-random
```

### 규칙

- `.env` → `.gitignore`에 등록. 절대 커밋하지 않는다.
- `.env.example` → git 관리. 키 이름과 placeholder만 기록.
- 환경변수 prefix는 프로젝트명 대문자 (예: `GOZIP_`, `APP_`).
- 시크릿(JWT, API 키)은 최소 32자 랜덤 문자열.

---

## 5. Backend (Go-GIN)

### Config 로드 패턴

```go
// internal/config/config.go
package config

type Config struct {
    Port           string
    Mode           string
    DBURL          string
    RedisURL       string
    JWTSecret      string
    CORSOrigins    []string
}

func Load() *Config {
    godotenv.Load() // .env 파일 로드
    return &Config{
        Port:        getEnv("APP_PORT", "49169"),
        Mode:        getEnv("GIN_MODE", "debug"),
        DBURL:       getEnv("DB_URL", ""),
        RedisURL:    getEnv("REDIS_URL", ""),
        JWTSecret:   getEnv("JWT_SECRET", ""),
        CORSOrigins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:49171"), ","),
    }
}
```

### 서버 바인딩

```go
// 항상 127.0.0.1에 바인딩
r.Run("127.0.0.1:" + cfg.Port)
```

`0.0.0.0`이나 `:`만 사용하지 않는다.

### DB 마이그레이션

- `golang-migrate/migrate` 사용.
- 마이그레이션 파일은 `internal/db/migrations/`에 `go:embed`로 포함.
- 서버 시작 시 자동 마이그레이션 실행.

```
internal/db/migrations/
├── 000001_initial_schema.up.sql
├── 000001_initial_schema.down.sql
├── 000002_add_feature.up.sql
└── 000002_add_feature.down.sql
```

### CORS

개발 환경에서 프론트엔드 dev server 주소만 허용한다.

```go
cors.Config{
    AllowOrigins:     cfg.CORSOrigins,  // ["http://localhost:9911"]
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}
```

---

## 6. Frontend (React + Vite)

### vite.config.ts

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 49171,  // 프로젝트별 할당 포트
    proxy: {
      '/api': {
        target: 'http://localhost:49169',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
  },
})
```

### API 호출 규칙

Vite proxy를 사용하므로 API 호출은 **상대 경로**로 통일한다.

```typescript
// api/client.ts
import axios from 'axios'

const client = axios.create({
  baseURL: '/api',  // 절대 URL 사용 금지
})
```

- 개발: Vite proxy가 `/api` → `localhost:9901`로 전달.
- 프로덕션: Go 서버가 정적 파일과 API를 같은 포트에서 서빙.
- `VITE_API_BASE_URL` 같은 환경변수 불필요.

### 프로덕션 빌드 서빙

Go 서버가 프론트엔드 빌드 결과물을 직접 서빙한다.

```go
// SPA fallback
r.NoRoute(func(c *gin.Context) {
    c.File("frontend/dist/index.html")
})
r.Static("/assets", "frontend/dist/assets")
```

---

## 7. Makefile

```makefile
.PHONY: dev dev-backend dev-frontend db-up db-down db-reset setup

# 최초 설정
setup:
	cp -n .env.example .env
	cp -n backend/.env.example backend/.env
	docker compose up -d
	cd frontend && npm install

# 인프라
db-up:
	docker compose up -d

db-down:
	docker compose down

db-reset:
	docker compose down -v
	docker compose up -d

# 개발 실행
dev:
	@make -j2 dev-backend dev-frontend

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

# 빌드
build-frontend:
	cd frontend && npm run build

build-backend:
	cd backend && go build -o bin/server ./cmd/server

build: build-frontend build-backend
```

### 개발 시작 절차

```bash
# 최초 1회
make setup        # .env 복사, DB 시작, npm install
vi .env           # 비밀번호 설정
vi backend/.env   # 백엔드 환경변수 설정

# 이후 매일
make dev          # 백엔드 + 프론트엔드 동시 실행 (Ctrl+C로 종료)
```

---

## 8. .gitignore 필수 항목

```gitignore
# 환경변수
.env
!.env.example

# 빌드 산출물
backend/bin/
frontend/dist/
frontend/node_modules/

# 시크릿
*.pem
*.key

# IDE
.vscode/
.idea/

# OS
.DS_Store
```

---

## 9. 체크리스트

새 프로젝트 시작 시:

- [ ] 포트 블록 할당 (기존 프로젝트와 충돌 확인)
- [ ] docker-compose.yml 작성 (127.0.0.1 바인딩, named volume)
- [ ] .env.example 작성 (비밀번호 placeholder)
- [ ] .env → .gitignore 등록
- [ ] 백엔드 서버 바인딩 `127.0.0.1` 확인
- [ ] 프론트엔드 Vite proxy 설정
- [ ] API 호출 상대 경로(`/api`) 통일
- [ ] Makefile 작성 (setup, dev, build)
- [ ] DB 마이그레이션 구조 생성
