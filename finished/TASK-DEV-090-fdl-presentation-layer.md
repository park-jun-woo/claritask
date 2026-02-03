# TASK-DEV-090: FDL Presentation Layer 구조화

## 개요

FDL 스펙의 Presentation Layer (UI 정의) 완전 구현

## 스펙 요구사항 (FDL/02-D-PresentationLayer.md)

### Component 타입
```
Page, Template, Organism, Molecule, Atom (Atomic Design)
```

### Props
```yaml
props:
  userId:
    type: string
    required: true
  onClose:
    type: function
    optional: true
```

### State
```yaml
state:
  - user: User | null
  - isLoading: boolean = false
  - error: string = ""
```

### Computed
```yaml
computed:
  - fullName: user.firstName + " " + user.lastName
  - canEdit: user.role === "admin" || user.id === currentUser.id
```

### Init
```yaml
init:
  - call: api.getUser($userId)
    set: user
    onError: set error = $error.message
  - parallel:
    - call: api.getProfile($userId)
      set: profile
    - call: api.getStats($userId)
      set: stats
```

### Methods
```yaml
methods:
  handleSubmit:
    - validate: form
    - call: api.updateUser($userId, $form)
    - navigate: /users/$userId
    - show: toast "Profile updated"
```

### View
```yaml
view:
  - Flex:
      direction: column
      children:
        - if: isLoading
          Text: "Loading..."
        - else:
          UserCard:
            user: $user
            onEdit: handleEdit
```

## 현재 상태

```go
type FDLUI struct {
    Component string
    Type      string
    Props     map[string]interface{}
    State     []string
    Init      []string
    View      []map[string]interface{}
    Parent    string
}
```

## 작업 내용

### 1. 구조체 확장

```go
type FDLUI struct {
    Component   string
    Type        string  // Page, Template, Organism, Molecule, Atom
    Description string
    Parent      string
    Props       map[string]FDLUIProp
    State       []FDLUIState
    Computed    []FDLUIComputed
    Init        []FDLUIAction
    Methods     map[string][]FDLUIAction
    View        []FDLUIElement
    Styles      map[string]interface{}
}

type FDLUIProp struct {
    Type     string  // string, number, boolean, function, object, array
    Required bool
    Optional bool
    Default  interface{}
}

type FDLUIState struct {
    Name    string
    Type    string
    Default interface{}
}

type FDLUIComputed struct {
    Name       string
    Expression string
}

type FDLUIAction struct {
    Type      string  // call, set, navigate, show, validate, confirm, emit, parallel, redirect
    Target    string
    Params    map[string]interface{}
    OnSuccess []FDLUIAction
    OnError   []FDLUIAction
}

type FDLUIElement struct {
    Type       string  // Text, Input, Button, Image, Flex, Grid, Stack, etc.
    Props      map[string]interface{}
    Children   []FDLUIElement
    Condition  *FDLUICondition  // if/else
}

type FDLUICondition struct {
    If   string
    Then []FDLUIElement
    Else []FDLUIElement
}
```

### 2. Props 파싱

```go
func parseUIProps(raw map[string]interface{}) map[string]FDLUIProp {
    result := make(map[string]FDLUIProp)
    for name, spec := range raw {
        prop := FDLUIProp{}
        if specMap, ok := spec.(map[string]interface{}); ok {
            if t, ok := specMap["type"].(string); ok {
                prop.Type = t
            }
            if r, ok := specMap["required"].(bool); ok {
                prop.Required = r
            }
            if o, ok := specMap["optional"].(bool); ok {
                prop.Optional = o
            }
            if d, ok := specMap["default"]; ok {
                prop.Default = d
            }
        }
        result[name] = prop
    }
    return result
}
```

### 3. View 파싱

```go
func parseUIView(raw []interface{}) []FDLUIElement {
    elements := []FDLUIElement{}
    for _, item := range raw {
        if itemMap, ok := item.(map[string]interface{}); ok {
            element := parseUIElement(itemMap)
            elements = append(elements, element)
        }
    }
    return elements
}

func parseUIElement(raw map[string]interface{}) FDLUIElement {
    element := FDLUIElement{
        Props: make(map[string]interface{}),
    }

    for key, value := range raw {
        if key == "if" {
            element.Condition = parseUICondition(raw)
            break
        }

        element.Type = key
        if props, ok := value.(map[string]interface{}); ok {
            for propKey, propValue := range props {
                if propKey == "children" {
                    if children, ok := propValue.([]interface{}); ok {
                        element.Children = parseUIView(children)
                    }
                } else {
                    element.Props[propKey] = propValue
                }
            }
        }
        break
    }

    return element
}
```

### 4. 검증 함수

```go
func validateUI(ui *FDLUI, allUIs []*FDLUI) []error {
    errors := []error{}

    // Type 검증
    validTypes := []string{"Page", "Template", "Organism", "Molecule", "Atom"}
    if !contains(validTypes, ui.Type) {
        errors = append(errors, fmt.Errorf("invalid UI type: %s", ui.Type))
    }

    // Parent 존재 확인
    if ui.Parent != "" {
        if !uiExists(ui.Parent, allUIs) {
            errors = append(errors, fmt.Errorf("parent not found: %s", ui.Parent))
        }
    }

    return errors
}
```

## 완료 조건

- [ ] FDLUI 구조체 확장
- [ ] FDLUIProp, FDLUIState, FDLUIComputed 구조체 추가
- [ ] FDLUIAction 구조체 추가
- [ ] FDLUIElement, FDLUICondition 구조체 추가
- [ ] Props 파싱 함수
- [ ] State, Computed 파싱 함수
- [ ] Init, Methods 파싱 함수
- [ ] View 파싱 함수 (재귀)
- [ ] UI 검증 함수
- [ ] 테스트 작성
