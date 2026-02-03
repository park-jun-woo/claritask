# ClariSpec: DSL 구조 개요

> **현재 버전**: v0.0.4 ([변경이력](../HISTORY.md))

---

## 전체 구조

ClariSpec은 4개의 계층으로 구성된 YAML 기반 DSL입니다.

```yaml
feature: <feature_name> (string)
description: <description> (string)

# 1. DATA LAYER (Schema & Models)
models:
  - name: <ModelName>
    table: <table_name>
    fields:
      - <name>: <type> [constraints...]

# 2. LOGIC LAYER (Service & Business Rules)
service:
  - name: <FunctionName>
    input: <args>
    output: <return_type>
    steps:
      - <step_description_or_pseudocode>

# 3. INTERFACE LAYER (API Contract)
api:
  - path: <http_path>
    method: <HTTP_METHOD>
    use: service.<function>  # Wiring Point
    request: <json_schema>
    response: <json_schema>

# 4. PRESENTATION LAYER (UI Components)
ui:
  - component: <ComponentName>
    type: <Page|Organism|Molecule|Atom>
    state:
      - <state_variable>
    view:
      - <element>: <label_or_content>
        action: <API.path>  # Wiring Point
```

---

## 계층별 상세 문서

| 계층 | 문서 | 역할 |
|------|------|------|
| DATA | [02-A-DataLayer.md](02-A-DataLayer.md) | 데이터베이스 스키마 및 모델 정의 |
| LOGIC | [02-B-LogicLayer.md](02-B-LogicLayer.md) | 비즈니스 로직 및 규칙 |
| INTERFACE | [02-C-InterfaceLayer.md](02-C-InterfaceLayer.md) | HTTP API 계약 |
| PRESENTATION | [02-D-PresentationLayer.md](02-D-PresentationLayer.md) | UI 컴포넌트 및 상태 관리 |

---

## 계층간 연결 (Wiring)

```
┌─────────────────────────────────────────────────────────────┐
│  PRESENTATION LAYER (ui)                                    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Button.action: "API.POST /posts/{postId}/comments" │───┼──┐
│  └─────────────────────────────────────────────────────┘   │  │
└─────────────────────────────────────────────────────────────┘  │
                                                                 │
┌─────────────────────────────────────────────────────────────┐  │
│  INTERFACE LAYER (api)                                      │  │
│  ┌─────────────────────────────────────────────────────┐   │  │
│  │  POST /posts/{postId}/comments                      │◄──┼──┘
│  │  use: service.createComment                         │───┼──┐
│  └─────────────────────────────────────────────────────┘   │  │
└─────────────────────────────────────────────────────────────┘  │
                                                                 │
┌─────────────────────────────────────────────────────────────┐  │
│  LOGIC LAYER (service)                                      │  │
│  ┌─────────────────────────────────────────────────────┐   │  │
│  │  createComment                                      │◄──┼──┘
│  │  db: "INSERT INTO comments"                         │───┼──┐
│  └─────────────────────────────────────────────────────┘   │  │
└─────────────────────────────────────────────────────────────┘  │
                                                                 │
┌─────────────────────────────────────────────────────────────┐  │
│  DATA LAYER (models)                                        │  │
│  ┌─────────────────────────────────────────────────────┐   │  │
│  │  Comment -> comments table                          │◄──┼──┘
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

*ClariSpec FDL Specification v0.0.4*
