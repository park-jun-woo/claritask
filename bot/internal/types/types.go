package types

// Result represents a command execution result
type Result struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	NeedsInput bool        `json:"needs_input,omitempty"`
	Prompt     string      `json:"prompt,omitempty"`
	Context    string      `json:"context,omitempty"`    // 대화 컨텍스트 유지용
	ErrorType  string      `json:"error_type,omitempty"` // 에러 유형 (auth_error 등)
}
