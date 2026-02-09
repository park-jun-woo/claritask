import { query } from "@anthropic-ai/claude-agent-sdk";
import type { SDKUserMessage } from "@anthropic-ai/claude-agent-sdk";
import type { GoMessage, ToolResult, BridgeMessage, AskUserQuestion } from "./protocol.js";

// Pending tool response resolvers keyed by request_id
const pendingResponses = new Map<
  string,
  (result: ToolResult) => void
>();

let requestCounter = 0;

function nextRequestId(): string {
  return `r${++requestCounter}`;
}

// Emit a message to Go via stdout
function emit(msg: BridgeMessage): void {
  process.stdout.write(JSON.stringify(msg) + "\n");
}

// Wait for a tool response from Go
function waitForResponse(
  requestId: string,
  signal?: AbortSignal
): Promise<ToolResult> {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) {
      reject(new Error("Aborted"));
      return;
    }
    pendingResponses.set(requestId, resolve);
    signal?.addEventListener("abort", () => {
      pendingResponses.delete(requestId);
      reject(new Error("Aborted"));
    });
  });
}

// Resolve a pending tool response
export function resolveToolResponse(
  requestId: string,
  result: ToolResult
): void {
  const resolver = pendingResponses.get(requestId);
  if (resolver) {
    pendingResponses.delete(requestId);
    resolver(result);
  }
}

// Dangerous patterns that should be routed to user for approval
const dangerousPatterns = [
  /\bgit\s+push\b/,
  /\bgit\s+push\s+.*--force/,
  /\brm\s+-rf?\s/,
  /\bsudo\b/,
  /\bchmod\s+777\b/,
  /\bdrop\s+table\b/i,
  /\bdrop\s+database\b/i,
];

function isDangerousCommand(input: Record<string, unknown>): boolean {
  const command = (input.command as string) || "";
  return dangerousPatterns.some((p) => p.test(command));
}

// Extract text content from SDK assistant message
function extractText(message: { message?: { content?: unknown[] } }): string {
  if (!message.message?.content) return "";
  const parts = message.message.content as Array<{
    type: string;
    text?: string;
  }>;
  return parts
    .filter((c) => c.type === "text" && c.text)
    .map((c) => c.text!)
    .join("");
}

// Run an agent query with streaming user input
export async function runAgent(
  projectPath: string,
  permissionMode: string,
  resumeSessionId?: string,
  stdinMessages?: AsyncIterable<GoMessage>
): Promise<void> {
  // Channel for user messages
  let resolveNext: ((value: SDKUserMessage | null) => void) | null = null;
  const messageQueue: (SDKUserMessage | null)[] = [];

  function pushUserMessage(content: string): void {
    const msg: SDKUserMessage = {
      type: "user",
      message: { role: "user", content },
      parent_tool_use_id: null,
      session_id: "",
    };
    if (resolveNext) {
      const r = resolveNext;
      resolveNext = null;
      r(msg);
    } else {
      messageQueue.push(msg);
    }
  }

  function endStream(): void {
    if (resolveNext) {
      const r = resolveNext;
      resolveNext = null;
      r(null);
    } else {
      messageQueue.push(null);
    }
  }

  async function* promptStream(): AsyncGenerator<SDKUserMessage> {
    while (true) {
      const msg =
        messageQueue.length > 0
          ? messageQueue.shift()!
          : await new Promise<SDKUserMessage | null>((resolve) => {
              resolveNext = resolve;
            });
      if (msg === null) return;
      yield msg;
    }
  }

  // Handle incoming messages from Go (stdin)
  // This runs concurrently with the query
  const handleStdin = async () => {
    if (!stdinMessages) return;
    for await (const msg of stdinMessages) {
      switch (msg.type) {
        case "user_message":
          pushUserMessage(msg.content);
          break;
        case "tool_response":
          resolveToolResponse(msg.request_id, msg.result);
          break;
        case "interrupt":
          // End the prompt stream to stop the agent
          endStream();
          break;
      }
    }
    // stdin closed
    endStream();
  };

  // Start stdin handling concurrently
  const stdinPromise = handleStdin();

  try {
    // Build options
    const options: Record<string, unknown> = {
      cwd: projectPath,
      permissionMode: permissionMode,
      settingSources: ["project"], // Load CLAUDE.md
      systemPrompt: { type: "preset", preset: "claude_code" },
      canUseTool: async (
        toolName: string,
        input: Record<string, unknown>,
        opts: { signal: AbortSignal }
      ) => {
        const requestId = nextRequestId();

        if (toolName === "AskUserQuestion") {
          // Route clarifying questions to Go/Telegram
          const questions = (
            (input.questions as Array<Record<string, unknown>>) || []
          ).map(
            (q): AskUserQuestion => ({
              question: (q.question as string) || "",
              header: (q.header as string) || "",
              options: (
                (q.options as Array<{ label: string; description: string }>) ||
                []
              ).map((o) => ({
                label: o.label || "",
                description: o.description || "",
              })),
              multi_select: (q.multiSelect as boolean) || false,
            })
          );
          emit({ type: "ask_user", request_id: requestId, questions });
        } else if (toolName === "ExitPlanMode") {
          // Route plan review to Go/Telegram
          emit({
            type: "plan_review",
            request_id: requestId,
            plan: (input.plan as string) || "",
          });
        } else if (
          toolName === "Bash" &&
          isDangerousCommand(input)
        ) {
          // Dangerous commands need explicit user approval
          emit({
            type: "tool_request",
            request_id: requestId,
            tool_name: toolName,
            input,
          });
        } else {
          // Auto-approve safe tools
          return { behavior: "allow" as const, updatedInput: input };
        }

        // Wait for Go to send back the response
        const result = await waitForResponse(requestId, opts.signal);

        if (result.behavior === "allow") {
          return {
            behavior: "allow" as const,
            updatedInput: result.updated_input,
          };
        } else {
          return {
            behavior: "deny" as const,
            message: result.message,
          };
        }
      },
    };

    if (resumeSessionId) {
      options.resume = resumeSessionId;
    }

    const q = query({
      prompt: promptStream(),
      options: options as Parameters<typeof query>[0]["options"],
    });

    for await (const message of q) {
      if (message.type === "system" && (message as { subtype?: string }).subtype === "init") {
        const sessionId = (message as { session_id: string }).session_id;
        emit({ type: "init", session_id: sessionId });
      } else if (message.type === "assistant") {
        const text = extractText(message as { message?: { content?: unknown[] } });
        if (text) {
          emit({ type: "assistant_text", content: text });
        }
      } else if (message.type === "result") {
        const result = message as {
          subtype: string;
          result?: string;
          total_cost_usd?: number;
          session_id: string;
          duration_ms?: number;
          errors?: string[];
        };
        if (result.subtype === "success") {
          emit({
            type: "result",
            status: "success",
            result: result.result || "",
            cost_usd: result.total_cost_usd || 0,
            session_id: result.session_id,
            duration_ms: result.duration_ms || 0,
          });
        } else {
          emit({
            type: "result",
            status: "error",
            result: (result.errors || []).join("\n"),
            cost_usd: result.total_cost_usd || 0,
            session_id: result.session_id,
            duration_ms: result.duration_ms || 0,
          });
        }
      }
    }
  } catch (err) {
    emit({
      type: "error",
      message: err instanceof Error ? err.message : String(err),
    });
  }

  await stdinPromise;
}
