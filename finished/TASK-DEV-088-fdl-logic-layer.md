# TASK-DEV-088: FDL Logic Layer 구조화

## 개요

FDL 스펙의 Logic Layer (서비스 정의) 완전 구현

## 스펙 요구사항 (FDL/02-B-LogicLayer.md)

### Input 검증
```yaml
input:
  email:
    type: string
    required: true
    minLength: 5
    maxLength: 255
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
```

### Output 형식
```yaml
output: User                  # 단순
output: Array<Comment>        # 배열
output: { user: User, token: string }  # 복합
output: void                  # 없음
```

### 기타 필드
```yaml
throws: [USER_NOT_FOUND, INVALID_EMAIL]
transaction: true
auth: required
roles: [admin]
ownership: user_id
```

### Steps 타입
```yaml
steps:
  - validate: email format
  - db: select User where email = $email
  - condition: if user exists throw DUPLICATE
  - db: insert User
  - event: emit UserCreated
  - return: user
```

## 현재 상태

```go
type FDLService struct {
    Name   string
    Desc   string
    Input  map[string]interface{}  // 자유형식
    Output string
    Steps  []string                // 자유형식
}
```

## 작업 내용

### 1. 구조체 확장

```go
type FDLService struct {
    Name        string
    Desc        string
    Input       map[string]FDLInputParam
    Output      FDLOutput
    Throws      []string
    Transaction bool
    Auth        string  // required, optional, none
    Roles       []string
    Ownership   string
    Steps       []FDLStep
}

type FDLInputParam struct {
    Type      string
    Required  bool
    Optional  bool
    Default   interface{}
    MinLength int
    MaxLength int
    Min       float64
    Max       float64
    Pattern   string
    Enum      []string
}

type FDLOutput struct {
    Type      string      // User, void
    IsArray   bool        // Array<User>
    IsComplex bool        // { user: User, token: string }
    Fields    map[string]string  // 복합 타입의 필드들
}

type FDLStep struct {
    Type      string  // validate, db, event, call, cache, log, transform, condition, loop, return
    Operation string  // 세부 동작
    Params    map[string]interface{}
}
```

### 2. Input 파싱

```go
func parseInputParams(raw map[string]interface{}) map[string]FDLInputParam {
    result := make(map[string]FDLInputParam)
    for name, spec := range raw {
        param := FDLInputParam{}
        if specMap, ok := spec.(map[string]interface{}); ok {
            if t, ok := specMap["type"].(string); ok {
                param.Type = t
            }
            if r, ok := specMap["required"].(bool); ok {
                param.Required = r
            }
            // ... 나머지 필드들
        }
        result[name] = param
    }
    return result
}
```

### 3. Output 파싱

```go
func parseOutput(raw string) FDLOutput {
    output := FDLOutput{}

    if raw == "void" {
        output.Type = "void"
        return output
    }

    if strings.HasPrefix(raw, "Array<") {
        output.IsArray = true
        output.Type = strings.TrimSuffix(strings.TrimPrefix(raw, "Array<"), ">")
        return output
    }

    if strings.HasPrefix(raw, "{") {
        output.IsComplex = true
        // 파싱 로직
        return output
    }

    output.Type = raw
    return output
}
```

### 4. Steps 파싱

```go
func parseSteps(raw []interface{}) []FDLStep {
    steps := []FDLStep{}
    for _, s := range raw {
        if stepStr, ok := s.(string); ok {
            steps = append(steps, parseStepString(stepStr))
        } else if stepMap, ok := s.(map[string]interface{}); ok {
            steps = append(steps, parseStepMap(stepMap))
        }
    }
    return steps
}

func parseStepString(s string) FDLStep {
    // "validate: email format" -> Step{Type: "validate", Operation: "email format"}
    // "db: select User where..." -> Step{Type: "db", Operation: "select User..."}
}
```

## 완료 조건

- [ ] FDLService 구조체 확장
- [ ] FDLInputParam 구조체 추가
- [ ] FDLOutput 구조체 추가
- [ ] FDLStep 구조체 추가
- [ ] Input 파싱 함수
- [ ] Output 파싱 함수
- [ ] Steps 파싱 함수
- [ ] 검증 함수 (throws, auth, roles)
- [ ] 테스트 작성
