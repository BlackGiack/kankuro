---
name: mcp-cli
description: Interface for MCP (Model Context Protocol) servers via CLI. Use when you need to interact with external tools, APIs, or data sources through MCP servers.
---

# MCP-CLI

Access MCP servers through the command line. MCP enables interaction with external systems like GitHub, filesystems, databases, and APIs.

## Commands

| Command | Output |
|---------|--------|
| `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json` | List all servers and tools |
| `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json info <server>` | Show server tools and parameters |
| `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json info <server> <tool>` | Get tool JSON schema |
| `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json grep "<pattern>"` | Search tools by name |
| `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call <server> <tool>` | Call tool (reads JSON from stdin if no args) |
| `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call <server> <tool> '<json>'` | Call tool with arguments |

**Both formats work:** `<server> <tool>` or `<server>/<tool>`

## Workflow

1. **Discover**: `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json` → see available servers
2. **Explore**: `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json info <server>` → see tools with parameters
3. **Inspect**: `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json info <server> <tool>` → get full JSON schema
4. **Execute**: `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call <server> <tool> '<json>'` → run with arguments

## Examples

```bash
# List all selected servers
MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json

# With descriptions
MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json -d

# See server tools
MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json info filesystem

# Get tool schema (both formats work)
MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json info filesystem read_file
MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json info filesystem/read_file

# Call tool
MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call filesystem read_file '{"path": "./README.md"}'

# Pipe from stdin (no '-' needed!)
cat args.json | MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call filesystem read_file

# Search for tools
MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json grep "*file*"

# Output is raw text (pipe-friendly)
MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call filesystem read_file '{"path": "./file"}' | head -10
```

## Advanced Chaining

```bash
# Chain: search files → read first match
MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call filesystem search_files '{"path": ".", "pattern": "*.md"}' \
  | head -1 \
  | while read -r f; do MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call filesystem read_file "{\"path\": \"$f\"}"; done

# Loop: process multiple files
MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call filesystem list_directory '{"path": "./src"}' \
  | while read f; do MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call filesystem read_file "{\"path\": \"$f\"}"; done

# Conditional: check before reading
MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call filesystem list_directory '{"path": "."}' \
  | grep -q "README" \
  && MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call filesystem read_file '{"path": "./README.md"}'

# Multi-server aggregation
{
  MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call github search_repositories '{"query": "mcp", "per_page": 3}'
  MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call filesystem list_directory '{"path": "."}'
}

# Save to file
MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json call github get_file_contents '{"owner": "x", "repo": "y", "path": "z"}' > output.txt
```

**Note:** `call` outputs raw text content directly (no jq needed for text extraction)

## Options

| Flag | Purpose |
|------|---------|
| `-d` | Include descriptions |
| `-c <path>` | Specify config file |

## Common Errors

| Wrong Command | Error | Fix |
|---------------|-------|-----|
| `mcp-cli server tool` | AMBIGUOUS_COMMAND | Use `call server tool` or `info server tool` |
| `mcp-cli run server tool` | UNKNOWN_SUBCOMMAND | Use `call` instead of `run` |
| `mcp-cli list` | UNKNOWN_SUBCOMMAND | Use `info` instead of `list` |
| `mcp-cli call server` | MISSING_ARGUMENT | Add tool name |
| `mcp-cli call server tool {bad}` | INVALID_JSON | Use valid JSON with quotes |

## Exit Codes

- `0`: Success
- `1`: Client error (bad args, missing config)
- `2`: Server error (tool failed)
- `3`: Network error

# IMPORTANT
- Always use the canonical command prefix `MCP_NO_DAEMON=1 mcp-cli -c .kankuro/mcp-servers.json`. Run it from the session workspace root; the file contains exactly the servers selected for the current activation, including Kankuro when selected.
- If you changed into a subdirectory, return to the session workspace root before invoking the prefix. Do not rely on inherited environment variables or daemon caches; `MCP_NO_DAEMON=1` makes each call reread the current workspace file.
