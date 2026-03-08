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
* **Capabilities:** Assigned **Skills** and **MCPs** (Model Context Protocol).

### 1.2 The Session

An active instance of an ID Card.

* **Lifecycle:** Sessions are **sub-processes** of KANKURO. If the main process dies, the sessions die.
* **Persistence:** Sessions run in a specific directory (`$PWD` or absolute path).
* **Detachment:** Users can "detach" from a session terminal using a global hotkey (e.g., `Ctrl+K`) to return to the dashboard without stopping the agent.

### 1.3 The Inbox (Ping Stack)

A FIFO (First-In-First-Out) notification system. When an agent stops to wait for user input, it "pings" KANKURO. These pings accumulate in a persistent notification stack for the user to address.

---

## 2. Technical Specification (Golang)

### 2.1 Project Structure

```text
/cmd/kankuro/main.go   # Entry point
/internal/
  /driver/             # Interface implementations for different CLIs
  /ui/                 # Bubble Tea TUI components
  /config/             # YAML/JSON persistence for ID Cards
  /session/            # Sub-process management

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

    // I/O
    SendInput(data string) error
    CaptureOutput() (chan []byte, error) // Stream raw PTY output chunks
    
    GetStatus() SessionStatus
}

```

---

## 3. Feature Roadmap

### 📦 Phase 0: The MVP (v0)

* **Dashboard TUI:** Grid layout showing **ID Cards**; Sidebar showing **Active Sessions**.
* **ID Card CRUD:** Create, edit, and save agent configurations.
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

* **Main Screen:** * **Center:** A collection of "Agent Identity Cards."
* **Right Sidebar:** A scrollable list of active sessions with status indicators (Running, Waiting/Ping, Terminated).


* **Interactions:**
* `Enter` on a Card: Launch/Edit.
* `Click/Select` on Session: Attach to the session PTY for direct interaction.
* `Ctrl+K`: Detach from the attached PTY and return to Kankuro Dashboard without killing the session.
* `Tab`: Switch dashboard focus between ID cards and sessions.
* `x`: Kill the selected running session from the sessions list.
* `Delete`: Remove the selected terminated session from the sessions list.


* **Visual Cues:** Agents awaiting input will "pulse" or show a notification badge in the sidebar.

---

## 5. Implementation Notes

* **Terminal Handling:** Use `creack/pty` or similar Go libraries to handle pseudo-terminal allocation, attach replay, and live terminal resizing for interactive CLIs.
* **TUI Framework:** Built using [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a modern, responsive terminal experience.
* **Attach Mode:** Bubble Tea should temporarily release terminal control during session attach, then restore the dashboard on detach.
* **Piping Logic:** Implement a `SanitizedWriter` that implements `io.Writer` to filter out escape sequences and UI artifacts during session-to-session transfers.

---
