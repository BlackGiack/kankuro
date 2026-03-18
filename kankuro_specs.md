---

# 🎭 KANKURO: The AI-CLI Puppet Master

**KANKURO** is a terminal-based orchestrator written in **Go** that allows you to manage, configure, and link multiple AI agent CLIs (Claude Code, Gemini CLI, OpenCode, etc.) through a unified interface.

By treating individual AI CLIs as "puppets," KANKURO enables complex multi-agent workflows with a focus on persistence, interoperability, and automation.

---

## 1. Core Architecture & Concepts

### 1.1 The "ID Card" (Config)

A saved configuration template for an agent. It encapsulates:

* **CLI Tool:** The underlying binary (e.g., `claude`, `gemini`).
* **Model:** Specific LLM selection.
* **System Instruction:** Default behavior overrides.
* **Temporary Workspace:** Option to run sessions in an isolated directory.
* **Capabilities:** Assigned **Skills** and **MCPs** (Model Context Protocol).
* **Face Style:** Visual identity used in the dashboard.

### 1.2 The Session

An active or persisted instance of an ID Card.

* **Lifecycle:** Sessions are tracked as **sub-processes**. Metadata is persisted to disk, allowing terminated sessions to be **relaunched** later.
* **Persistence:** 
    * **Standard:** Runs in a specific directory (`$PWD` or absolute path).
    * **Temporary:** Clones the original workspace into a disposable folder (via `KANKURO_TMP_DIR` or `~/.config/kankuro/tmp/`). Changes made by the agent are isolated from the original source but persist across session relaunches.
* **Detachment:** Users can "detach" from a session terminal using a global hotkey (e.g., `Ctrl+K`) to return to the dashboard without stopping the agent.
* **Switch:** Users can switch ID cards keeping the same session.

### 1.3 The Waiting status

KANKURO detects when the agents are stopped either because they are done or because they are waiting for an input. When this happens it lets the user know by highlighting the session and replacing the base face with the neutral one.

---

## 2. Technical Specification (Golang)

### 2.1 Project Structure

```text
/cmd/kankuro/main.go   # Entry point
/internal/
  /driver/             # Interface implementations for different CLIs
  /ui/                 # Bubble Tea TUI components
  /config/             # YAML persistence for ID Cards
  /session/            # Sub-process management and JSON persistence
  /ascii/              # Ascii art package
```

### 2.2 The `AgentDriver` Interface

Every CLI tool integrated into KANKURO must implement this interface to ensure the orchestrator remains tool-agnostic.

```go
type AgentDriver interface {
    // Lifecycle
    Spawn(config AgentConfig) (*SessionHandle, error)
    Kill() error

    // Configuration
    SetSkills(skills []Skill) error // Anthropic Open Standard
    SetMcps(mcps []McpServer) error // Model Context Protocol
    SetModel(model string) []string

    // I/O
    SendInput(data string) error // Not implemented, in prevision of automations
    
    GetStatus() SessionStatus
}
```

---

## 3. Feature Roadmap

### 📦 Phase 0: The MVP (v0)

* **Dashboard TUI:** Grid layout showing **ID Cards**; Sidebar showing **Active/Persisted Sessions**.
* **ID Card CRUD:** Create, edit, and save agent configurations with ASCII face styles.
* **Session Persistence:** Save session state to disk to allow relaunching after app restarts.
* **Temporary Workspaces:** Isolated "sandbox" execution for agents to prevent source code corruption.
* **Session Wrapping:** Launching and detaching from interactive CLI sub-processes with raw PTY attach.
* **Skill/MCP Registry:** User-editable list of agentic capabilities following the Anthropic open standard.

### 🔗 Phase 1: Connectivity (v1)

* **Direct Piping:** Route `STDOUT` from Session A directly into `STDIN` of Session B.
* **Sanitized Output:** A middleware layer to strip non-textual noise (ANSI codes, progress bars, spinners) from pipes to ensure the receiving agent gets clean context.
* **Inter-Session Communication:** UI triggers to link agents in a chain.

### 🤖 Phase 2: Hybrid Automation (v2)

* **Deterministic Engine:** Regex and time-based triggers (e.g., *"If output contains 'Done' OR no activity for 5m, then [Action]"*).
* **Intelligent Controller:** Integration of a local mini-LLM (via Ollama or similar) to process pings and generate automated responses based on natural language instructions.
* **Notification Stack:** Persistent UI element for managing pings from backgrounded agents.

---

## 4. UI/UX Design

* **Main Screen:** 
    * **Center:** A collection of "Agent Identity Cards" with assigned ASCII faces.
    * **Right Sidebar:** A scrollable list of sessions showing status (Running, Waiting, Terminated) and the agent's identity face.

* **Interactions:**
    * `Enter` on a Card: Launch a new session.
    * `Enter` on a Session: Attach if running, **Relaunch** if terminated.
    * `Ctrl+K`: Detach from the attached PTY and return to Kankuro Dashboard.
    * `Tab`: Switch dashboard focus between ID cards and sessions.
    * `x`: Kill the selected running session.
    * `Backspace/Delete`: Trigger a **confirmation dialog** to remove a terminated session (and its temporary workspace).

* **Visual Cues:** 
    * Agents awaiting input will show a **neutral face** and use a **hot-pink pulse** style in the sidebar.

---

## 5. Implementation Notes

* **Terminal Handling:** Use `creack/pty` or similar Go libraries to handle pseudo-terminal allocation, attach replay, and live terminal resizing for interactive CLIs.
* **TUI Framework:** Built using [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a modern, responsive terminal experience.
* **Attach Mode:** Bubble Tea should temporarily release terminal control during session attach, then restore the dashboard on detach.
* **Storage:** ID Cards are stored in YAML; Sessions are stored in JSON within the user's config directory.
* **Workspace Isolation:** Temporary workspaces are created using standard system copy tools to ensure attribute preservation.

---
