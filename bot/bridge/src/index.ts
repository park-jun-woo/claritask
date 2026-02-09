import * as readline from "readline";
import type { GoMessage, StartMessage } from "./protocol.js";
import { runAgent } from "./agent.js";

// Read NDJSON lines from stdin
function createStdinReader(): AsyncIterable<GoMessage> {
  const rl = readline.createInterface({
    input: process.stdin,
    crlfDelay: Infinity,
  });

  const queue: GoMessage[] = [];
  let resolve: ((value: IteratorResult<GoMessage>) => void) | null = null;
  let done = false;

  rl.on("line", (line) => {
    const trimmed = line.trim();
    if (!trimmed) return;
    try {
      const msg = JSON.parse(trimmed) as GoMessage;
      if (resolve) {
        const r = resolve;
        resolve = null;
        r({ value: msg, done: false });
      } else {
        queue.push(msg);
      }
    } catch {
      process.stderr.write(`[bridge] invalid JSON: ${trimmed}\n`);
    }
  });

  rl.on("close", () => {
    done = true;
    if (resolve) {
      const r = resolve;
      resolve = null;
      r({ value: undefined as unknown as GoMessage, done: true });
    }
  });

  return {
    [Symbol.asyncIterator]() {
      return {
        next(): Promise<IteratorResult<GoMessage>> {
          if (queue.length > 0) {
            return Promise.resolve({ value: queue.shift()!, done: false });
          }
          if (done) {
            return Promise.resolve({
              value: undefined as unknown as GoMessage,
              done: true,
            });
          }
          return new Promise((r) => {
            resolve = r;
          });
        },
      };
    },
  };
}

async function main(): Promise<void> {
  process.stderr.write("[bridge] starting...\n");

  const stdinReader = createStdinReader();
  const iterator = stdinReader[Symbol.asyncIterator]();

  // Wait for the "start" message
  const first = await iterator.next();
  if (first.done) {
    process.stderr.write("[bridge] stdin closed before start message\n");
    process.exit(1);
  }

  const startMsg = first.value;
  if (startMsg.type !== "start") {
    process.stderr.write(
      `[bridge] expected start message, got: ${startMsg.type}\n`
    );
    process.exit(1);
  }

  const config = startMsg as StartMessage;
  process.stderr.write(
    `[bridge] started for project: ${config.project_path}\n`
  );

  // Create an async iterable from remaining stdin messages
  const remainingMessages: AsyncIterable<GoMessage> = {
    [Symbol.asyncIterator]() {
      return {
        next() {
          return iterator.next();
        },
      };
    },
  };

  // Run the agent loop
  await runAgent(
    config.project_path,
    config.permission_mode,
    config.session_id,
    remainingMessages
  );

  process.stderr.write("[bridge] agent loop ended\n");
  process.exit(0);
}

main().catch((err) => {
  process.stderr.write(`[bridge] fatal: ${err}\n`);
  process.exit(1);
});
