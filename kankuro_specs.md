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
* **Workspace:** A selected workspace from the workspace registry.
* **System Instruction:** Default behavior overrides.
* **Temporary Workspace:** Option to run sessions in an isolated directory.
* **Capabilities:** Assigned **Skills**, **MCPs** (Model Context Protocol through cli), **Personas**.
* **Face Style:** Visual identity used in the dashboard.

### 1.1.1 The Workspace Registry

Workspaces are managed as a registry, similar to skills and personas.

* **Workspace Sources:** Users can register either:
  * A root directory containing multiple workspaces, where each direct child directory becomes a selectable workspace.
  * A direct path to a single workspace.
* **Workspace Selection:** ID cards reference a workspace entry from this registry instead of storing a raw working directory path directly.

### 1.2 The Session

An active or persisted instance of an ID Card.

* **Lifecycle:** Sessions are tracked as **sub-processes**. Metadata is persisted to disk, allowing terminated sessions to be **relaunched** later.
* **Persistence:** 
    * **Standard:** Runs in a specific directory (`$PWD` or absolute path).
    * **Temporary:** Clones the original workspace into a disposable folder (via `KANKURO_TMP_DIR` or `~/.config/kankuro/tmp/`). Changes made by the agent are isolated from the original source but persist across session relaunches.
* **Detachment:** Users can "detach" from a session terminal using a global hotkey (e.g., `Ctrl+K`) to return to the dashboard without stopping the agent.
* **Switch:** Users can switch ID cards keeping the same session.
* **Quitting:** Quitting KANKURO (`q`, `Ctrl+C`, SIGTERM, or closing the terminal) kills every running session: its process group gets SIGTERM, then SIGKILL after 2 seconds. The session records stay on disk, so the sessions can be relaunched next time.

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

Every CLI tool integrated into KANKURO must implement this interface to ensure the orchestrator remains tool-agnostic. A Generic Driver is available with functional methods, specific drivers should only override the needed features.

```go
type AgentDriver interface {
	Spawn(config AgentConfig) (*SessionHandle, error)
	Kill() error

	SetAgentsMD(card config.IDCard, personas []config.Persona, mcps []config.MCPServer, workspaceRoot string, appendIfExist bool) error
	SetSkills(skills []config.Skill, workspaceRoot ...string) error
	SetPersonas(personas []config.Persona, workspaceRoot ...string) error
	SelectMCPs(mcps []config.MCPServer, workspaceRoot ...string) error
	SetModel(model string) []string

	SendInput(data string) error

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
* **Capability Registries:** User-editable registries for Skills, Personas, MCP servers, and Workspaces.

### 🔗 Phase 1: Connectivity (v1)

* **A2A:** 
* **Inter-Session Communication:** UI triggers to link agents in a chain.

### 🤖 Phase 2: Hybrid Automation (v2)

* **Deterministic Engine:** Regex and time-based triggers (e.g., *"If output contains 'Done' OR no activity for 5m, then [Action]"*).
* **Intelligent Controller:** Integration of a local mini-LLM (via Ollama or similar) to process pings and generate automated responses based on natural language instructions.
* **Notification Stack:** Persistent UI element for managing pings from backgrounded agents.

---

## 4. UI/UX Design

* **Main Screen (Dashboard):** 
    * **Left:** A collection of "Agent Identity Cards" with assigned ASCII puppets.
    * **Center:** A a preview of the selected session.
    * **Right Sidebar:** A scrollable list of sessions showing status (Running, Waiting, Terminated) and the agent's identity face.

* **Dashboard Interactions:**
    * `Enter` on a Card: Launch a new session.
    * `Enter` on a Session: Attach if running, **Relaunch** if terminated.
    * `Ctrl+K`: Detach from the attached PTY and return to Kankuro Dashboard.
    * `Tab`: Switch dashboard focus between ID cards and sessions.
    * `e`: Edit the selected ID Card.
    * `n`: Create a new ID Card.
    * `x`: Kill the selected running session.
    * `s`: Switch agent on the selected terminated session.
    * `m`: Open the MCP registry form to add a new server.
    * `w`: Open the workspace source manager.
    * `c`: Open the settings view.
    * `Backspace/Delete` (on Session): Trigger a **confirmation dialog** to remove a terminated session (and its temporary workspace).

* **Minimum terminal size:** 80×24. Below it every screen is replaced by a
  "Terminal too small: need 80×24, have W×H" notice and keys other than quit (`q`, `Ctrl+C`)
  are ignored, so nothing can be launched or deleted blind; the screen returns as soon as the terminal grows.

* **Dashboard footer:** shows only the keys that apply: navigation, the focused pane's keys
  (`[Card]` or `[Session]`) and the global keys. On a narrow terminal it drops whole groups
  (`q=Quit` is the last to go) rather than being cut off.

* **Settings View:**
    * Lets the user configure the IDE launch command.
    * Lets the user choose whether the Personas and MCPs capability lists are shown in the ID card editor.
    * `Enter`: Save settings.
    * `Space` (on Toggles): Toggle the highlighted option.
    * `Esc`: Return to the previous view.

* **Agent Card Editor Interactions:**
    * `Tab`: Cycle focus between fields (Name, Driver, Model, Workspace, Face, Tmp, AgentsMD, Skills, Personas, MCPs).
      Persona and MCP sections are optional and disappear from focus order when disabled in Settings.
    * `Space` (on Workspace): Open the searchable workspace picker.
    * `Space` (on Lists): Toggle selection of the highlighted Skill, Persona, or MCP.
    * `ctrl+e` (on Lists): Open the specialized form to add a new Skill Source, Persona Source, or MCP Server.
    * `Delete/Del` (on MCP List): Remove the highlighted MCP from the global registry (only if not referenced by any ID Card).
    * `Enter`: Save the ID Card and return to the Dashboard.
    * `Esc`: Cancel changes and return to the Dashboard.
    * When the form is taller than the terminal it scrolls to keep the focused field visible;
      `↑ more` / `↓ more` mark a cut edge.

* **Workspace Manager Interactions:**
    * `n`: Add a new workspace source.
    * `e` / `Space`: Edit the selected workspace source.
    * `d`: Delete the selected workspace source.
    * `r`: Refresh workspace discovery from configured sources.
    * `Esc`: Return to the dashboard.

* **Visual Cues:** 
    * Agents awaiting input will show a **neutral face** and use a **hot-pink pulse** style in the sidebar.

---

## 5. Implementation Notes

* **Terminal Handling:** Use `creack/pty` or similar Go libraries to handle pseudo-terminal allocation, attach replay, and live terminal resizing for interactive CLIs.
* **TUI Framework:** Built using [Bubble Tea](https://github.com/charmbracelet/bubbletea) for a modern, responsive terminal experience.
* **Attach Mode:** Bubble Tea should temporarily release terminal control during session attach, then restore the dashboard on detach.
* **Storage:** Configs and application settings are stored in YAML; Sessions are stored in JSON within the user's config directory.
* **Config Data:** ID cards store `workspace_id`; workspace paths are resolved through the workspace registry.
* **Workspace Isolation:** Temporary workspaces are created using standard system copy tools to ensure attribute preservation.

### 5.1 Running the App for Testing

Never run a test/dev instance against the real user directories — a session
launched from a manual test writes a session record, and a temporary workspace
copies an entire workspace tree. Both land under the user's own config directory
by default and are easy to forget about.

Point both overrides at a throwaway location before launching:

```bash
KANKURO_TMP_DIR=/tmp/kankuro-test/tmp \
KANKURO_CONFIG_DIR=/tmp/kankuro-test/config \
  go run ./cmd/kankuro
```

* **`KANKURO_TMP_DIR`** — root for temporary workspaces. Defaults to
  `~/.config/kankuro/tmp/`, so leaving it unset means every tmp-workspace test
  leaves a full copy of the source tree there (`internal/session/workspace.go`,
  `lifecycle.go`).
* **`KANKURO_CONFIG_DIR`** — root for ID cards, sessions and the capability
  registries (also settable with `--config-dir`). Leaving it unset means a test
  run adds real ID cards and session records to the user's own config.

Both are read at startup, so they must be set on the launching process — exporting
them after the app starts has no effect. Tests in `internal/session` already use
`t.TempDir()` and are unaffected; this note is about running the binary by hand.

---
