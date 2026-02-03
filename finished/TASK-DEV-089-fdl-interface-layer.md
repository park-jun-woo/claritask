# TASK-DEV-089: FDL Interface Layer 구조화

## 개요

FDL 스펙의 Interface Layer (API 정의) 완전 구현

## 스펙 요구사항 (FDL/02-C-InterfaceLayer.md)

### Request 구조
```yaml
request:
  params:
    id: int (required)
  query:
    page: int (default: 1)
    limit: int (default: 20, max: 100)
  headers:
    X-API-Key: string (required)
  body:
    email: string (required)
    password: string (required, minLength: 8)
```

### Response 상태코드
```yaml
response:
  200: User
  201: { id: int, message: string }
  400: { error: string, details: Array<string> }
  401: { error: "Unauthorized" }
  404: { error: "Not Found" }
  422: { error: string, fields: object }
```

### 기타 필드
```yaml
auth: required
roles: [admin, manager]
tags: [users, auth]
rateLimit:
  limit: 100
  window: 60
  by: ip
```

## 현재 상태

```go
type FDLAPI struct {
    Path     string
    Method   string
    Summary  string
    Use      string
    Request  map[string]interface{}
    Response map[string]interface{}
}
```

## 작업 내용

### 1. 구조체 확장

```go
type FDLAPI struct {
    Path      string
    Method    string
    Summary   string
    Use       string
    Request   FDLAPIRequest
    Response  map[int]interface{}  // 상태코드 → 스키마
    Auth      string               // required, optional, none, apiKey
    Roles     []string
    Tags      []string
    RateLimit *FDLRateLimit
    Mapping   map[string]string    // 파라미터 매핑
    Transform *FDLTransform        // 응답 변환
}

type FDLAPIRequest struct {
    Params  map[string]FDLRequestParam  // 경로 파라미터
    Query   map[string]FDLRequestParam  // 쿼리스트링
    Headers map[string]FDLRequestParam
    Body    map[string]FDLRequestParam
}

type FDLRequestParam struct {
    Type      string
    Required  bool
    Default   interface{}
    Min       float64
    Max       float64
    MinLength int
    MaxLength int
    Pattern   string
    Enum      []string
}

type FDLRateLimit struct {
    Limit  int     // 요청 수
    Window int     // 시간 (초)
    By     string  // ip, user, apiKey
}

type FDLTransform struct {
    Exclude []string
    Rename  map[string]string
}
```

### 2. Request 파싱

```go
func parseAPIRequest(raw map[string]interface{}) FDLAPIRequest {
    req := FDLAPIRequest{
        Params:  make(map[string]FDLRequestParam),
        Query:   make(map[string]FDLRequestParam),
        Headers: make(map[string]FDLRequestParam),
        Body:    make(map[string]FDLRequestParam),
    }

    if params, ok := raw["params"].(map[string]interface{}); ok {
        for name, spec := range params {
            req.Params[name] = parseRequestParam(spec)
        }
    }
    // ... query, headers, body 동일

    return req
}
```

### 3. Response 파싱

```go
func parseAPIResponse(raw map[string]interface{}) map[int]interface{} {
    result := make(map[int]interface{})
    for code, schema := range raw {
        codeInt, _ := strconv.Atoi(code)
        result[codeInt] = schema
    }
    return result
}
```

### 4. 검증 함수

```go
func validateAPI(api *FDLAPI, services []*FDLService) []error {
    errors := []error{}

    // Method 검증
    validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
    if !contains(validMethods, strings.ToUpper(api.Method)) {
        errors = append(errors, fmt.Errorf("invalid method: %s", api.Method))
    }

    // Path 형식 검증
    if !strings.HasPrefix(api.Path, "/") {
        errors = append(errors, fmt.Errorf("path must start with /"))
    }

    // Use 서비스 존재 확인
    if !serviceExists(api.Use, services) {
        errors = append(errors, fmt.Errorf("service not found: %s", api.Use))
    }

    // Response 상태코드 검증
    validCodes := []int{200, 201, 204, 400, 401, 403, 404, 409, 422, 500}
    for code := range api.Response {
        if !containsInt(validCodes, code) {
            errors = append(errors, fmt.Errorf("unusual status code: %d", code))
        }
    }

    return errors
}
```

## 완료 조건

- [ ] FDLAPI 구조체 확장
- [ ] FDLAPIRequest 구조체 추가
- [ ] FDLRequestParam 구조체 추가
- [ ] FDLRateLimit, FDLTransform 구조체 추가
- [ ] Request 파싱 함수
- [ ] Response 파싱 함수
- [ ] API 검증 함수 (method, path, use, 상태코드)
- [ ] 테스트 작성
