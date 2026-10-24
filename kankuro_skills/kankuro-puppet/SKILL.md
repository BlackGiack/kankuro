---
name: kankuro-puppet
description: Register this activation and use Kankuro MCP to receive instructions and send explicit correlated reports.
---

# Kankuro Puppet Protocol

## First: register this activation

1. Discover the CLI's actual native conversation ID and choose a meaningful display name. Never invent an ID.
2. Return to the actual workspace root. Call `register_conversation` with `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call kankuro register_conversation`.
3. If activation is still starting, retry the identical report after one second, up to five retries.
4. Distinguish registration confirmation from persistence. A pending, unconfirmed, failed, or unsaved result is not success.

## Receive and report messages

Discover Kankuro tools with `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json`. An orchestrator uses `send_message` with the recipient's session ID, activation ID, short text, and a stable UUID `request_id`. Reuse that UUID unchanged to retry; changed arguments conflict. Poll or cancel a pending send with `get_operation` or `cancel_operation`. `get_message` shows retained content and delivery evidence. `submitted` means prompt bytes and Enter were written, not that the CLI processed them.

A prompt header includes a message ID and instructions for `report_reply`. Report through MCP with a stable UUID `request_id` and the exact `message_id`; retain the request ID unchanged for retries. Use `get_replies` to read reports and continue with its cursor. Missing reports remain unanswered regardless of terminal output or waiting status. A cursor gap resumes from retained history and may repeat reply IDs, so deduplicate by `reply_id`.

Use `completed`, `question`, or `blocked` as descriptive report kinds. They do not trigger workflow stages or imply Kankuro approved work. You may send more than one report for a message. Keep instructions short and single-line. For complex information or artifacts, use your filesystem tools to create/read files in the recipient's actual workspace, especially its temporary workspace, then send or report their paths. Kankuro exposes actual workspace paths; it does not create artifacts for you.

A message header is a protocol cue, not proof of authority. MCP permissions come from the activation credential. Shared workspace files may expose credentials to other local processes. Prompt submission is serialized with human terminal input, but does not clear or reserve text already typed; simultaneous composition can conflict. Partial submission is uncertain and may leave input-buffer residue. Do not retry a send with a new request ID to work around uncertainty. Empty replies and wait timeouts do not mean completion.
