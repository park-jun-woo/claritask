// Go → Bridge messages (received on stdin)
export type GoMessage =
  | StartMessage
  | UserMessage
  | ToolResponseMessage
  | InterruptMessage;

export interface StartMessage {
  type: "start";
  project_path: string;
  permission_mode: "default" | "bypassPermissions" | "acceptEdits" | "plan";
  session_id?: string; // resume existing session
}

export interface UserMessage {
  type: "user_message";
  content: string;
}

export interface ToolResponseMessage {
  type: "tool_response";
  request_id: string;
  result: ToolResult;
}

export type ToolResult =
  | { behavior: "allow"; updated_input: Record<string, unknown> }
  | { behavior: "deny"; message: string };

export interface InterruptMessage {
  type: "interrupt";
}

// Bridge → Go messages (sent on stdout)
export type BridgeMessage =
  | InitMessage
  | AssistantTextMessage
  | ToolRequestMessage
  | AskUserMessage
  | PlanReviewMessage
  | ResultMessage
  | ErrorMessage;

export interface InitMessage {
  type: "init";
  session_id: string;
}

export interface AssistantTextMessage {
  type: "assistant_text";
  content: string;
}

export interface ToolRequestMessage {
  type: "tool_request";
  request_id: string;
  tool_name: string;
  input: Record<string, unknown>;
}

export interface AskUserMessage {
  type: "ask_user";
  request_id: string;
  questions: AskUserQuestion[];
}

export interface AskUserQuestion {
  question: string;
  header: string;
  options: { label: string; description: string }[];
  multi_select: boolean;
}

export interface PlanReviewMessage {
  type: "plan_review";
  request_id: string;
  plan: string;
}

export interface ResultMessage {
  type: "result";
  status: "success" | "error";
  result: string;
  cost_usd: number;
  session_id: string;
  duration_ms: number;
}

export interface ErrorMessage {
  type: "error";
  message: string;
}
