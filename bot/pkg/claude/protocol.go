package claude

// Go → Bridge messages (sent to bridge stdin)

// BridgeStartMsg tells the bridge to initialize for a project
type BridgeStartMsg struct {
	Type           string `json:"type"`            // "start"
	ProjectPath    string `json:"project_path"`
	PermissionMode string `json:"permission_mode"` // default, bypassPermissions, acceptEdits, plan
	SessionID      string `json:"session_id,omitempty"`
}

// BridgeUserMsg sends a user message to the bridge
type BridgeUserMsg struct {
	Type    string `json:"type"` // "user_message"
	Content string `json:"content"`
}

// BridgeToolResponseMsg responds to a tool_request/ask_user/plan_review from the bridge
type BridgeToolResponseMsg struct {
	Type      string          `json:"type"`       // "tool_response"
	RequestID string          `json:"request_id"`
	Result    BridgeToolResult `json:"result"`
}

// BridgeToolResult is the allow/deny response for a tool request
type BridgeToolResult struct {
	Behavior     string                 `json:"behavior"` // "allow" or "deny"
	UpdatedInput map[string]interface{} `json:"updated_input,omitempty"`
	Message      string                 `json:"message,omitempty"`
}

// BridgeInterruptMsg tells the bridge to stop the current execution
type BridgeInterruptMsg struct {
	Type string `json:"type"` // "interrupt"
}

// Bridge → Go messages (received from bridge stdout)

// BridgeMessage is the union type for all messages from the bridge
type BridgeMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id,omitempty"`

	// assistant_text
	Content string `json:"content,omitempty"`

	// tool_request
	RequestID string                 `json:"request_id,omitempty"`
	ToolName  string                 `json:"tool_name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`

	// ask_user
	Questions []BridgeQuestion `json:"questions,omitempty"`

	// plan_review
	Plan string `json:"plan,omitempty"`

	// result
	Status     string  `json:"status,omitempty"`
	Result     string  `json:"result,omitempty"`
	CostUSD    float64 `json:"cost_usd,omitempty"`
	DurationMs int64   `json:"duration_ms,omitempty"`

	// error
	Message string `json:"message,omitempty"`
}

// BridgeQuestion represents a question from AskUserQuestion
type BridgeQuestion struct {
	Question    string                `json:"question"`
	Header      string                `json:"header"`
	Options     []BridgeQuestionOption `json:"options"`
	MultiSelect bool                  `json:"multi_select"`
}

// BridgeQuestionOption represents one option in an AskUserQuestion
type BridgeQuestionOption struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}
