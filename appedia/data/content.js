/* Hand-written content for the KANKURO appedia: the one file to edit.
 * The model (types, methods, code, relations) is generated from the Go code by
 * kankuro_skills/create-appedia/scripts/extract-go; this file adds what code
 * cannot say: what each part is for, how the pieces move together (flows), and
 * the rules people must know. Every reference written as "pkg.Type",
 * "pkg.Type.Method" or "pkg.Func" is checked against the model at build time.
 * See appedia/README.md to rebuild. */
window.APPEDIA_CONTENT = {
  app: "KANKURO",
  title: "Kankuro Appedia",
  tagline: "A terminal stage where AI agents are the puppets",
  language: "go",
  rulesSource: "project.md §6",
  tasksSource: "project.md §10",
  /* participants of flows that are not code: [label, subtitle] */
  externals: {
    user: ["You", "keyboard"], tea: ["Bubble Tea", "UI loop"], pty: ["Agent process", "in a PTY"],
    disk: ["Disk", "~/.config/kankuro"], cmd: ["Command", "any tea.Cmd"]
  },
  /* how the generic flow kinds read in this app */
  flowKinds: {
    cmd: ["⟳ Command", "Bubble Tea runs it on its own goroutine"],
    go: ["⟳ Goroutine", "starts goroutines"],
    msg: ["Message", "comes back through Bubble Tea's Update"]
  },
  snapshot: "Working tree after Act 2 and the bonus layout plan (on top of commit 4055764), 2026-10-06.",
  pitch: "KANKURO is a terminal UI written in Go that orchestrates AI coding-agent CLIs: Claude Code today, any CLI through a generic driver. You describe an agent once as an ID card, launch as many sessions of it as you like, and switch between them from one dashboard.",
  pitch2: "It is a Bubble Tea program. One goroutine runs Update and View; anything slow (disk, processes) runs in a command on another goroutine and comes back as a message. Each agent runs in its own pseudo-terminal and its record is saved as JSON, so the list survives a restart.",
  concepts: [
    ["ID card", "A saved agent template: driver, model, workspace, skills, personas, MCP servers, AGENTS.md, env, face and puppet.", "config.IDCard"],
    ["Session", "A running or stopped instance of a card: a process in a PTY, its scrollback, a JSON record.", "session.Session"],
    ["Driver", "How to talk to one CLI: which files to write, how to spawn and kill it.", "driver.AgentDriver"],
    ["Registries", "Workspaces, skills, personas and puppets discovered from source directories; MCP servers entered by hand.", "config.Repository"],
    ["Coordinator", "The one object that routes every message between the screens and the rest.", "app.Coordinator"],
    ["Status", "running, waiting (looks like it wants input) or terminated.", "session.Status"]
  ],
  path: [
    ["The package map: the coordinator sits between the screens and the processes.", "pkg-app"],
    ["Zoom into the coordinator type to see what it owns.", "type-app.Coordinator"],
    ["Watch a launch go from keypress to running process.", "flow-launch"],
    ["Watch the 2-second tick keep the screen fresh without blocking.", "flow-tick"],
    ["Read rules 1 and 2: goroutines and navigation, the easiest to break.", "rules"]
  ],

  /* ---------- level 0: the package map ---------- */
  map: {
    viewBox: [0, 0, 1000, 690],
    bands: [
      { label: "ENTRY", y: 16, h: 70 },
      { label: "UI LOOP", y: 100, h: 84 },
      { label: "SCREENS · PROCESSES", y: 198, h: 200 },
      { label: "DATA · ART", y: 412, h: 178 },
      { label: "TEST-ONLY", y: 604, h: 76 }
    ],
    nodes: {
      kankuro: [500, 51], tea: [190, 142], app: [500, 142],
      views: [150, 250], messages: [360, 250], session: [660, 250],
      components: [150, 350], component: [330, 350], styles: [480, 350], driver: [660, 350], pty: [870, 350],
      ascii: [130, 470], characters: [300, 470], config: [480, 470], fsutil: [670, 470], disk: [870, 470],
      root: [480, 552],
      uitest: [330, 642], testenv: [640, 642]
    },
    external: {
      tea: { label: "Bubble Tea", sub: "UI loop library" },
      pty: { label: "Agent CLI", sub: "in a PTY" },
      disk: { label: "~/.config/kankuro", sub: "YAML + JSON" }
    },
    extraEdges: [["kankuro", "tea"], ["tea", "app"], ["driver", "pty"], ["fsutil", "disk"], ["testenv", "session"]]
  },

  packages: {
    kankuro: {
      label: "cmd/kankuro", sub: "main.go",
      role: "Program entry. Wires the repository, the driver registry, the session manager and the UI together, and owns the exit path.",
      files: { "main.go": "flags, debug logger, wiring, SIGHUP → quit, stopSessions after Run" },
      notes: ["Every exit path (q, Ctrl+C, SIGTERM, closing the terminal) ends in Manager.Shutdown, so no agent outlives KANKURO.", "stderr may be written only here, after Bubble Tea has released the terminal."],
      rules: [], key: []
    },
    app: {
      label: "internal/app", sub: "the coordinator",
      role: "The glue. Routes every message, owns navigation, the error banner, the preview subscription and the refresh cadence. Blocking work is delegated to helpers that hold no pointer back to it.",
      files: {
        "coordinator.go": "the Coordinator struct, Update, message routing",
        "navigation.go": "navigateTo (rule 2), initViews, resizeAll, broadcasts",
        "handlers.go": "UI-goroutine halves of refresh and preview",
        "refresh.go": "refresher: read commands, fingerprint check, tick",
        "preview.go": "previewSub: live output of the selected session",
        "session_actions.go": "sessionActions: launch, relaunch, switch, deletes, IDE",
        "callbacks.go": "every view's callbacks",
        "attach.go": "raw-PTY attach/detach, Ctrl+K",
        "size_guard.go": "below 80×24: notice, keys dropped",
        "app.go": "tea.Model wrapper"
      },
      notes: ["refresher and sessionActions hold no *Coordinator: a closure they build cannot touch coordinator state (rule 1 by construction).", "The error banner clears on the next keypress, never on navigation (rule 4)."],
      rules: ["1", "2", "3", "4", "8"],
      key: ["Coordinator", "refresher", "sessionActions", "previewSub", "attachExec"]
    },
    views: {
      label: "ui/views", sub: "full screens",
      role: "Full screens. Each is a component.Model; views never call each other: they return a Cmd and the coordinator reacts to the message it produces.",
      files: {
        "dashboard.go": "three panes, keys, search, footer",
        "card_editor.go": "ID-card form; scrolls by blocks",
        "mcp_form.go": "MCP form driven by mcpFieldTable",
        "source_manager.go": "source directories per registry",
        "settings.go": "IDE command, editor sections",
        "session_launcher.go": "progress while a session starts",
        "agent_switch.go": "pick a card for a terminated session",
        "workspace_picker_modal.go": "modal: choose a workspace",
        "puppet_selector_modal.go": "modal: choose a puppet",
        "agentsmd_editor_modal.go": "modal: AGENTS.md content or path",
        "constants.go": "MinTerminalWidth/Height, CoordinatorRows",
        "source_type.go": "per-registry labels"
      },
      notes: ["A view that receives data implements messages.DataReceiver and guards every field with if msg.X != nil (rule 3)."],
      rules: ["2", "3", "5", "9", "10"],
      key: ["Dashboard", "CardEditor", "MCPForm", "SourceManager", "SettingsView", "AgentSwitchView", "SessionLauncher"]
    },
    messages: {
      label: "ui/messages", sub: "every tea.Msg",
      role: "Every message the app exchanges, in one file. Read it before adding a message.",
      files: { "messages.go": "all message types, DataReceiver, PreviewReceiver" },
      notes: ["DataLoadedMsg is always a full snapshot; partial updates get their own type."],
      rules: ["3"], key: ["DataLoadedMsg", "PreviewDataMsg", "SessionsRefreshedMsg", "DataReceiver"]
    },
    session: {
      label: "internal/session", sub: "process lifecycle",
      role: "Owns every Session: spawns them through a driver, tracks their output and status, saves their records, and stops everything on Shutdown.",
      files: {
        "manager.go": "Manager, Session, Snapshot",
        "lifecycle.go": "Launch, Relaunch, SwitchAgent, Kill, Remove, Shutdown",
        "watch.go": "status polling every 2 s",
        "io.go": "output tracking and scrollback",
        "attach.go": "AttachHandle and the locked input writer",
        "storage.go": "session JSON, atomic",
        "workspace.go": "temporary workspace copies",
        "ide_launcher.go": "open a workspace in the IDE"
      },
      notes: ["Every activation bumps a generation counter so goroutines of an older run stop touching the session.", "Relaunch and SwitchAgent reserve the session first (activating); Remove refuses while it is set."],
      rules: ["8"], key: ["Manager", "Session", "Snapshot", "AttachHandle"]
    },
    components: {
      label: "ui/components", sub: "reusable pieces",
      role: "Building blocks the views compose: lists, the card renderer, the preview, modals, text fitting.",
      files: {
        "card_list.go": "ID-card rendering, pills, ClipBlock",
        "session_list.go": "sessions pane list",
        "session_preview.go": "live preview of agent output",
        "capability_list.go": "skills/personas/MCPs checklist",
        "source_list.go": "source directories list",
        "modal.go": "Modal and ModalState",
        "text.go": "DisplayText, TruncateHead, TruncateTail",
        "renderer.go": "puppet rendering", "filter.go": "generic search",
        "source_item.go": "source list items", "constants.go": "layout constants"
      },
      notes: ["Fit by display width, never bytes. Clip art, never wrap it (rule 9)."],
      rules: ["9", "10"], key: ["CardList", "SessionList", "SessionPreview", "CapabilityList", "ModalState"]
    },
    component: { label: "ui/component", sub: "base type", role: "The Base type (Width, Height, SetSize) and the Model interface every view and component implements.", files: { "component.go": "Base, Model" }, notes: [], rules: [], key: ["Base", "Model"] },
    styles: { label: "ui/styles", sub: "palette", role: "The colour palette from crow_puppet_theme.md and shared lipgloss styles. Variables only, no types.", files: { "styles.go": "colours and styles" }, notes: [], rules: [], key: [] },
    driver: {
      label: "internal/driver", sub: "one per CLI",
      role: "One AgentDriver per CLI: which files to write into a workspace, how to spawn the CLI in a PTY, how to send it keys and how to kill it.",
      files: { "driver.go": "GenericCLIDriver, ClaudeCodeDriver, Spawn, Kill, file sync, IsPromptLike", "registry.go": "name → constructor", "factory.go": "construction helpers" },
      notes: ["Kill signals the whole process group: SIGTERM, up to 2 s grace, then SIGKILL for anything left. A descendant that starts its own session escapes."],
      rules: [], key: ["AgentDriver", "GenericCLIDriver", "ClaudeCodeDriver", "MapRegistry"]
    },
    ascii: { label: "internal/ascii", sub: "text-art engine", role: "Canvas and shapes for drawing ASCII art.", files: { "canvas.go": "drawing surface", "shape.go": "Shape interface", "options.go": "config options" }, notes: [], rules: [], key: ["Canvas", "Shape"] },
    characters: { label: "ascii/characters", sub: "puppets & faces", role: "Built-in puppets (cat, darth, default, kankuro, strawhat, totoro), faces, and user puppets loaded from .puppet files.", files: { "puppet.go": "built-in puppets", "face.go": "agent faces", "text_puppet.go": "user puppets" }, notes: [], rules: [], key: ["Face", "TextPuppet"] },
    config: {
      label: "internal/config", sub: "YAML repository",
      role: "The repository for everything the user configures: ID cards, registries and their source directories, MCP servers, app settings.",
      files: {
        "repository.go": "FileRepository, interfaces, settings", "storage.go": "WriteAtomic, Fingerprint",
        "types.go": "IDCard, Workspace, Skill, Persona, MCPServer, Puppet, AppSettings",
        "id_cards.go": "cards", "skills.go": "skills", "personas.go": "personas", "mcps.go": "MCP servers", "workspaces.go": "workspaces",
        "skill_sources.go": "skill sources", "persona_sources.go": "persona sources", "puppet_sources.go": "puppet sources",
        "sync.go": "scanning source dirs", "builtins.go": "built-in mcp-cli skill",
        "validation.go": "card validation", "validation_helpers.go": "validation helpers", "errors.go": "SyncFailedError"
      },
      notes: ["Cards reference workspaces and capabilities by ID, never by path."],
      rules: ["6", "7"], key: ["FileRepository", "Repository", "IDCard", "AppSettings"]
    },
    fsutil: { label: "internal/fsutil", sub: "atomic writes", role: "WriteFileAtomic: temp file in the same directory, fsync, rename, cleanup on error. Used by config and session.", files: { "atomic.go": "WriteFileAtomic" }, notes: ["A crash leaves at worst a .<name>.<rand>.tmp file, which loaders ignore."], rules: [], key: [] },
    root: { label: "kankuro (module root)", sub: "embeds mcp-cli", role: "The module root package; it only embeds kankuro_skills/mcp-cli/SKILL.md so config can ship the built-in skill.", files: { "embed_skills.go": "//go:embed of the mcp-cli skill" }, notes: [], rules: [], key: [] },
    uitest: { label: "ui/uitest", sub: "shape assertions", test: true, role: "Test-only helpers for layout tests: exact/bounded shape assertions, regions, adversarial content.", files: { "shape.go": "AssertExactRect, AssertWithin, Region, AdversarialNames" }, notes: ["Shape tests check fit; position tests check that elements stay put. A list-like component needs both."], rules: ["9", "10"], key: [] },
    testenv: { label: "internal/testenv", sub: "TestMain sandbox", test: true, role: "Runs a package's tests in a private TMPDIR and KANKURO_TMP_DIR and fails the package if anything is left behind.", files: { "testenv.go": "Run(m, name)" }, notes: ["Session-launching tests build managers with newTestManager, which registers a Shutdown cleanup."], rules: [], key: [] },
    tea: { label: "Bubble Tea", sub: "UI loop library", ext: true, role: "charm.land/bubbletea/v2. Runs Update(msg) and View() on one goroutine, and every tea.Cmd on its own goroutine, feeding the result back as a message.", files: {}, notes: ["That last fact is what rule 1 is about."], rules: ["1"], key: [] },
    pty: { label: "Agent CLI", sub: "in a PTY", ext: true, role: "The agent process (claude or any CLI), started with creack/pty in its own session and process group.", files: {}, notes: [], rules: [], key: [] },
    disk: { label: "~/.config/kankuro", sub: "YAML + JSON", ext: true, role: "The config directory. See Reference → Storage for its layout.", files: {}, notes: [], rules: [], key: [] }
  },

  /* ---------- flows: sequence diagrams ----------
   participants: [key, ref-or-external, label?]
   steps: [from, to, kind, label, note, codeRef]
   kinds: key (user input), call, return, cmd (Bubble Tea runs it on its own
   goroutine), msg (a message back through Bubble Tea), io (process or disk),
   chan (channel send), self, go (starts a goroutine) */
  flowGroups: [
    { name: "Start & stop", flows: ["startup", "quit", "resize"] },
    { name: "Sessions", flows: ["launch", "attach", "relaunch", "switch", "kill", "remove", "watch", "ide"] },
    { name: "Keeping the screen fresh", flows: ["tick", "preview", "mutation"] },
    { name: "Configuration", flows: ["editcard", "savecard", "deletecard", "source", "mcp", "settings"] }
  ],
  flows: {
    startup: {
      title: "Startup", summary: "From main() to the first frame: open the config, load the sessions, build the coordinator, hand the terminal to Bubble Tea.",
      participants: [["main", "kankuro.run", "main.go"], ["repo", "config.FileRepository"], ["mgr", "session.Manager"], ["coord", "app.Coordinator"], ["ref", "app.refresher"], ["tea", "tea"], ["disk", "disk"]],
      steps: [
        ["main", "repo", "call", "NewFileRepository(dir)", "Opens the YAML repository and creates missing files.", "config.NewFileRepository"],
        ["main", "repo", "call", "SyncSkillsFromSources()", "Rescans the skill source directories into skills.yaml.", "config.FileRepository.SyncSkillsFromSources"],
        ["main", "mgr", "call", "NewManager(dir, repo, opts…)", "Builds the session registry.", "session.NewManager"],
        ["mgr", "disk", "io", "loadSessions: sessions/*.json", "Every loaded session is marked terminated: a process cannot survive a restart. Corrupt files are skipped and reported once.", "session.loadSessions"],
        ["main", "coord", "call", "app.New → NewCoordinator", "Builds every view with its callbacks.", "app.NewCoordinator"],
        ["main", "tea", "call", "NewProgram(app).Run()", "Bubble Tea takes the terminal. From here on, everything goes through Update and View.", "kankuro.run"],
        ["tea", "coord", "call", "Init()", "Returns a batch of commands.", "app.Coordinator.Init"],
        ["tea", "ref", "cmd", "fullSeeded · syncPuppets · sessionLoadErrors", "Each runs on its own goroutine: read every registry, rescan puppet sources, report session-load errors.", "app.refresher.fullSeeded"],
        ["ref", "coord", "msg", "configFingerprintMsg (+ data)", "Applied on the UI goroutine: the baseline is stored and the data broadcast to every view.", "app.Coordinator.handleConfigFingerprint"],
        ["tea", "coord", "msg", "tickMsg, every 2 s", "Starts the refresh cadence (see the tick flow).", "app.tick"]
      ]
    },
    quit: {
      title: "Quit", summary: "Every exit path ends in Shutdown, which kills running agents before KANKURO exits.",
      participants: [["user", "user"], ["dash", "views.Dashboard"], ["tea", "tea"], ["main", "kankuro.run", "main.go"], ["mgr", "session.Manager"], ["drv", "driver.GenericCLIDriver"], ["pty", "pty"], ["disk", "disk"]],
      steps: [
        ["user", "dash", "key", "q or Ctrl+C", "Closing the terminal window sends SIGHUP, which main turns into p.Quit(). SIGTERM becomes a quit inside Bubble Tea.", "views.Dashboard.handleKey"],
        ["dash", "tea", "return", "tea.Quit", "", "views.Dashboard.handleKey"],
        ["tea", "main", "return", "Run() returns", "Whatever it returned, the shutdown below runs.", "kankuro.run"],
        ["main", "mgr", "call", "RunningCount() → one stderr line", "Only when sessions are running. stderr is safe now: the TUI is gone.", "kankuro.stopSessions"],
        ["main", "mgr", "call", "Shutdown()", "Sets closed: Launch, Relaunch and SwitchAgent now return ErrManagerClosed.", "session.Manager.Shutdown"],
        ["mgr", "drv", "go", "Kill(id) for each live session, in parallel", "The wait is about one grace period in total, not one per session.", "session.Manager.Kill"],
        ["drv", "pty", "io", "SIGTERM to the process group", "The agent gets a chance to save its state.", "driver.GenericCLIDriver.Kill"],
        ["drv", "pty", "io", "after ≤ 2 s: SIGKILL if the group is alive", "Also reaches children that ignore SIGTERM and SIGHUP.", "driver.groupAlive"],
        ["mgr", "disk", "io", "records saved as terminated", "", "session.Manager.persistSessionLockedErr"],
        ["mgr", "mgr", "self", "wait activations · close stop · wait watchers", "A launch caught mid-spawn kills its own process; watchers exit without a final write.", "session.Manager.Shutdown"]
      ]
    },
    resize: {
      title: "Resize & too-small terminal", summary: "How views get their size, and what happens below 80×24.",
      participants: [["tea", "tea"], ["coord", "app.Coordinator"], ["dash", "views.Dashboard"]],
      steps: [
        ["tea", "coord", "msg", "tea.WindowSizeMsg", "", "app.Coordinator.handleSpecialMessages"],
        ["coord", "dash", "call", "resizeAll → SetSize(w, h − CoordinatorRows)", "Every view gets the terminal minus the rows the coordinator keeps back; the dashboard adds them back to fill the screen exactly.", "app.Coordinator.resizeAll"],
        ["tea", "coord", "call", "View()", "", "app.Coordinator.View"],
        ["coord", "coord", "self", "tooSmall() → tooSmallView()", "Below 80×24 every screen is replaced by a centred notice sized to the terminal.", "app.Coordinator.tooSmallView"],
        ["tea", "coord", "msg", "a keypress while too small", "", "app.Coordinator.Update"],
        ["coord", "coord", "self", "guardKey: dropped (q / Ctrl+C still quit)", "A blind keypress must not launch or delete anything. Non-key messages keep flowing, so the screen is current when the terminal grows.", "app.Coordinator.guardKey"]
      ]
    },
    launch: {
      title: "Launch a session", summary: "Enter on a card: prepare the launcher screen on the UI goroutine, then resolve, copy, write files, spawn and save on a command goroutine.",
      participants: [["user", "user"], ["dash", "views.Dashboard"], ["coord", "app.Coordinator"], ["tea", "tea"], ["act", "app.sessionActions"], ["mgr", "session.Manager"], ["drv", "driver.GenericCLIDriver"], ["pty", "pty"], ["disk", "disk"]],
      steps: [
        ["user", "dash", "key", "Enter on a card", "", "views.Dashboard.handleEnter"],
        ["dash", "coord", "call", "OnLaunchCard(card) → launchCard", "Runs inside the dashboard's Update, on the UI goroutine. It prepares the SessionLauncher view and returns commands; it must not navigate itself (rule 2).", "app.Coordinator.launchCard"],
        ["coord", "tea", "return", "Batch(launch(card), NavigateMsg)", "", "app.Coordinator.launchCard"],
        ["tea", "coord", "msg", "NavigateMsg → SessionLauncher", "Navigation happens in a message handler.", "app.Coordinator.navigateTo"],
        ["tea", "act", "cmd", "launch(card)", "Blocking work from here on, on a command goroutine. sessionActions has no coordinator pointer (rule 1).", "app.sessionActions.launch"],
        ["act", "act", "self", "resolveCapabilities(card)", "Reads the skill, persona and MCP registries once; a failed read aborts the launch instead of starting an agent with nothing.", "app.sessionActions.resolveCapabilities"],
        ["act", "mgr", "call", "Launch(card, skills, personas, mcps)", "", "session.Manager.Launch"],
        ["mgr", "disk", "io", "prepareWorkspace: copy to tmp/<id>/", "Only when the card asks for a temporary workspace.", "session.Manager.prepareWorkspace"],
        ["mgr", "drv", "call", "syncDriverFiles", "Writes AGENTS.md or CLAUDE.md, skills, personas and .kankuro/mcp-servers.json into the workspace.", "session.Manager.syncDriverFiles"],
        ["drv", "pty", "io", "Spawn: pty.Start, own session", "The card's env is merged over the parent's.", "driver.GenericCLIDriver.Spawn"],
        ["mgr", "mgr", "self", "activateSessionLocked + trackOutput + watch", "Bumps the generation counter and starts the per-activation goroutines.", "session.Manager.activateSessionLocked"],
        ["mgr", "disk", "io", "saveSession (atomic)", "A failure here is ErrPersistFailed: started, just not saved (rule 8).", "session.saveSession"],
        ["act", "coord", "msg", "SessionLaunchCompletedMsg", "The launcher shows the result; from there you can name the session or attach. A refresh follows.", "app.Coordinator.handleSessionMessages"]
      ]
    },
    attach: {
      title: "Attach & detach", summary: "Bubble Tea hands the raw terminal to the agent until Ctrl+K.",
      participants: [["user", "user"], ["dash", "views.Dashboard"], ["coord", "app.Coordinator"], ["tea", "tea"], ["exec", "app.attachExec"], ["mgr", "session.Manager"], ["pty", "pty"]],
      steps: [
        ["user", "dash", "key", "a, or Enter on a running session", "", "views.Dashboard.handleAttach"],
        ["dash", "coord", "call", "OnAttach(id) → attachSession", "", "app.Coordinator.attachSession"],
        ["coord", "tea", "return", "tea.Exec(attachExec)", "Bubble Tea releases the terminal to the exec command.", "app.Coordinator.attachSession"],
        ["tea", "exec", "call", "Run()", "", "app.attachExec.Run"],
        ["exec", "mgr", "call", "Attach(id) → AttachHandle", "Scrollback, a one-shot header, the output channel and an input writer.", "session.Manager.Attach"],
        ["exec", "exec", "self", "replay scrollback, then pumpOutput", "", "app.attachExec.pumpOutput"],
        ["user", "exec", "key", "keystrokes", "", "app.attachExec.pumpInput"],
        ["exec", "pty", "io", "forwardInput → SendInputBytes", "Through the driver's locked writer, never the raw fd: that would race Kill.", "app.attachExec.forwardInput"],
        ["user", "exec", "key", "Ctrl+K", "Recognised as the legacy byte or as a Kitty keyboard-protocol sequence.", "app.detectKittyDetach"],
        ["exec", "coord", "msg", "AttachFinishedMsg", "Bubble Tea takes the terminal back; a refresh follows.", "app.Coordinator.handleSessionMessages"]
      ]
    },
    relaunch: {
      title: "Relaunch a session", summary: "Enter on a terminated session restarts it in the same workspace, then attaches.",
      participants: [["user", "user"], ["dash", "views.Dashboard"], ["tea", "tea"], ["act", "app.sessionActions"], ["mgr", "session.Manager"], ["drv", "driver.GenericCLIDriver"], ["coord", "app.Coordinator"]],
      steps: [
        ["user", "dash", "key", "Enter on a terminated session", "", "views.Dashboard.handleSessionEnter"],
        ["dash", "tea", "return", "OnRelaunch(id) command", "The callback is the sessionActions method itself: no coordinator state involved.", "app.sessionActions.relaunch"],
        ["tea", "act", "cmd", "relaunch(id)", "", "app.sessionActions.relaunch"],
        ["act", "act", "self", "resolveCard(session.CardID)", "A failed read of id_cards.yaml is reported as such, not as \"card not found\".", "app.sessionActions.resolveCard"],
        ["act", "mgr", "call", "Relaunch(id, card, …)", "", "session.Manager.Relaunch"],
        ["mgr", "mgr", "self", "reserveForActivation", "Must be terminated and not already starting; marks it activating so a second relaunch, a switch or a remove fails fast.", "session.Manager.reserveForActivation"],
        ["mgr", "drv", "call", "startSession in the same workspace", "No new copy. Files are synced and the driver spawned as in a launch.", "session.Manager.startSession"],
        ["act", "coord", "msg", "SessionAttachRequestedMsg", "Carries a Warning when only the record failed to save.", "app.Coordinator.handleSessionMessages"],
        ["coord", "coord", "self", "navigate to dashboard, attachSession", "", "app.Coordinator.attachSession"]
      ]
    },
    switch: {
      title: "Switch agent", summary: "Restart a terminated session with a different card, in the same workspace.",
      participants: [["user", "user"], ["dash", "views.Dashboard"], ["coord", "app.Coordinator"], ["sw", "views.AgentSwitchView"], ["tea", "tea"], ["act", "app.sessionActions"], ["mgr", "session.Manager"]],
      steps: [
        ["user", "dash", "key", "s on a terminated session", "", "views.Dashboard.handleSwitch"],
        ["dash", "coord", "msg", "AgentSwitchRequestedMsg", "", "app.Coordinator.handleSessionMessages"],
        ["coord", "sw", "call", "SetSession, navigateTo(AgentSwitch)", "", "views.AgentSwitchView.SetSession"],
        ["user", "sw", "key", "pick a card, Enter", "", "views.AgentSwitchView.handleKey"],
        ["sw", "tea", "return", "OnSwitch → switchAgent command", "The switch must not run inside Update: it writes files and spawns a process.", "app.sessionActions.switchAgent"],
        ["tea", "act", "cmd", "switchAgent(id, cardID)", "", "app.sessionActions.switchAgent"],
        ["act", "mgr", "call", "SwitchAgent(id, card, …)", "Reserves the session, replaces AGENTS.md (no append), spawns the new driver in the same workspace.", "session.Manager.SwitchAgent"],
        ["act", "coord", "msg", "AgentSwitchedMsg", "", "app.Coordinator.handleAgentSwitched"],
        ["coord", "coord", "self", "refresh + SessionAttachRequestedMsg", "The attach handler navigates; a persist warning travels with it (rules 4 and 8).", "app.Coordinator.handleAgentSwitched"]
      ]
    },
    kill: {
      title: "Kill a session", summary: "x on a running session: SIGTERM, grace, SIGKILL, then mark it terminated.",
      participants: [["user", "user"], ["dash", "views.Dashboard"], ["tea", "tea"], ["mgr", "session.Manager"], ["drv", "driver.GenericCLIDriver"], ["pty", "pty"], ["coord", "app.Coordinator"]],
      steps: [
        ["user", "dash", "key", "x on a running session", "Only offered for running or waiting sessions.", "views.Dashboard.handleKill"],
        ["dash", "tea", "return", "OnKill(id) command", "", "app.Coordinator.dashboardCallbacks"],
        ["tea", "mgr", "cmd", "Kill(id)", "Captures the driver and generation under the lock, then kills without holding it.", "session.Manager.Kill"],
        ["mgr", "drv", "call", "Kill()", "", "driver.GenericCLIDriver.Kill"],
        ["drv", "pty", "io", "SIGTERM group → ≤ 2 s → SIGKILL group", "", "driver.signalGroup"],
        ["mgr", "mgr", "self", "same activation? → terminated + persist", "If the session was removed or relaunched meanwhile, it is left alone, so a removed record is never re-created.", "session.Manager.Kill"],
        ["mgr", "coord", "msg", "DataRefreshRequestedMsg (or ErrorMsg)", "", "app.Coordinator.handleSpecialMessages"]
      ]
    },
    remove: {
      title: "Remove a session", summary: "Delete a terminated session's record and its temporary workspace.",
      participants: [["user", "user"], ["dash", "views.Dashboard"], ["modal", "components.Modal"], ["tea", "tea"], ["mgr", "session.Manager"], ["disk", "disk"], ["coord", "app.Coordinator"]],
      steps: [
        ["user", "dash", "key", "Backspace or Delete on a terminated session", "", "views.Dashboard.handleDeleteSession"],
        ["dash", "modal", "call", "open the confirm modal", "", "components.ModalState.Open"],
        ["user", "dash", "key", "Enter or y", "", "views.Dashboard.confirmDelete"],
        ["dash", "tea", "return", "OnRemove(id) command", "", "app.Coordinator.dashboardCallbacks"],
        ["tea", "mgr", "cmd", "Remove(id)", "Refuses a running session, or one being started.", "session.Manager.Remove"],
        ["mgr", "disk", "io", "delete sessions/<id>.json and tmp/<id>/", "A record that is already gone is not an error.", "session.deleteSession"],
        ["mgr", "coord", "msg", "DataRefreshRequestedMsg", "", "app.Coordinator.handleSpecialMessages"]
      ]
    },
    watch: {
      title: "Status & output tracking", summary: "Two goroutines per activation: one copies output into scrollback, one polls the status every 2 s.",
      participants: [["pty", "pty"], ["drv", "driver.GenericCLIDriver"], ["track", "session.Manager", "Manager · trackOutput"], ["watch", "session.Manager", "Manager · watch"], ["sub", "app.previewSub"], ["disk", "disk"]],
      steps: [
        ["pty", "drv", "io", "output bytes", "", "driver.GenericCLIDriver.streamOutput"],
        ["drv", "track", "chan", "Output channel", "", "session.Manager.trackOutput"],
        ["track", "track", "self", "append scrollback (≤ 256 KiB), last line", "Stops at once if the generation changed: it belongs to an older run.", "session.appendScrollback"],
        ["track", "sub", "chan", "broadcast to subscribers", "Non-blocking: a slow subscriber drops chunks instead of stalling the agent.", "session.broadcast"],
        ["watch", "drv", "call", "every 2 s: GetStatus()", "", "session.Manager.watch"],
        ["watch", "watch", "self", "IsPromptLike → waiting or running", "Waiting means the last line looks like a prompt and no word arrived recently.", "driver.IsPromptLike"],
        ["watch", "disk", "io", "terminated → persist, close subscribers", "Also stops when Shutdown closes the stop channel.", "session.Manager.watch"]
      ]
    },
    ide: {
      title: "Open in IDE", summary: "o on a session opens its workspace with the configured IDE command.",
      participants: [["user", "user"], ["dash", "views.Dashboard"], ["coord", "app.Coordinator"], ["tea", "tea"], ["act", "app.sessionActions"], ["ide", "session.IDELauncher"]],
      steps: [
        ["user", "dash", "key", "o on a session", "", "views.Dashboard.handleLaunchIDE"],
        ["dash", "coord", "call", "OnLaunchIDE(id) → launchIDE", "Reads c.settings.IDECommand here, on the UI goroutine, and passes the value on.", "app.Coordinator.launchIDE"],
        ["tea", "act", "cmd", "launchIDE(id, command)", "", "app.sessionActions.launchIDE"],
        ["act", "ide", "call", "NewIDELauncher(cmd).Launch(workspace)", "", "session.IDELauncher.Launch"],
        ["act", "coord", "msg", "IDELaunchedMsg", "An error shows in the banner.", "app.Coordinator.handleSpecialMessages"]
      ]
    },
    tick: {
      title: "The 2-second tick", summary: "Sessions refresh from memory every tick; config re-reads only when a file changed.",
      participants: [["tea", "tea"], ["coord", "app.Coordinator"], ["ref", "app.refresher"], ["mgr", "session.Manager"], ["repo", "config.FileRepository"], ["dash", "views.Dashboard"]],
      steps: [
        ["tea", "coord", "msg", "tickMsg", "", "app.Coordinator.handleSpecialMessages"],
        ["coord", "coord", "self", "snapshot preview ID + fingerprint", "Plain values: the commands below never see the coordinator (rule 1).", "app.Coordinator.refreshSessions"],
        ["tea", "ref", "cmd", "sessionsOnly(previewID)", "", "app.refresher.sessionsOnly"],
        ["ref", "mgr", "call", "List(), GetScrollback(id)", "Memory only, no disk.", "session.Manager.List"],
        ["ref", "coord", "msg", "SessionsRefreshedMsg", "Applied to the dashboard only.", "app.Coordinator.handleSessionsRefreshed"],
        ["tea", "ref", "cmd", "configIfChanged(baseline)", "", "app.refresher.configIfChanged"],
        ["ref", "repo", "call", "Fingerprint(): stat each registry", "No parsing. Only if a stamp changed are the registries re-read.", "config.FileRepository.Fingerprint"],
        ["ref", "coord", "msg", "configFingerprintMsg", "A new baseline, plus a full DataLoadedMsg when something changed.", "app.Coordinator.handleConfigFingerprint"],
        ["coord", "dash", "call", "broadcastData → SetData", "Every DataReceiver gets it; a nil field means keep what you have (rule 3).", "app.Coordinator.broadcastData"]
      ]
    },
    preview: {
      title: "Live preview", summary: "The dashboard's centre pane streams the selected session's output.",
      participants: [["user", "user"], ["dash", "views.Dashboard"], ["coord", "app.Coordinator"], ["sub", "app.previewSub"], ["mgr", "session.Manager"], ["tea", "tea"], ["pv", "components.SessionPreview"]],
      steps: [
        ["user", "dash", "key", "select another session", "", "views.Dashboard.syncPreviewSelection"],
        ["dash", "coord", "msg", "SessionPreviewRequestedMsg", "", "app.Coordinator.handlePreviewRequest"],
        ["coord", "sub", "call", "subscribe(id)", "Drops the previous subscription first.", "app.previewSub.subscribe"],
        ["sub", "mgr", "call", "GetScrollback · SubscribeOutput", "", "session.Manager.SubscribeOutput"],
        ["coord", "pv", "call", "broadcastPreview(PreviewDataMsg)", "A partial update: never through SetData (rule 3).", "app.Coordinator.broadcastPreview"],
        ["tea", "sub", "cmd", "next(): wait on the channel", "The closure captures only the channel and the ID.", "app.previewSub.next"],
        ["sub", "coord", "msg", "SessionOutputMsg", "", "app.Coordinator.handleSessionOutput"],
        ["coord", "pv", "call", "accepts? → append, re-arm next()", "Output for a session we have left is ignored.", "app.previewSub.accepts"],
        ["coord", "sub", "call", "leaving the dashboard: unsubscribe()", "Clears the ID with the channel, so queued output cannot re-arm the wait.", "app.previewSub.unsubscribe"]
      ]
    },
    mutation: {
      title: "Refresh after a change, and errors", summary: "How a command asks for fresh data, and how the error banner lives and dies.",
      participants: [["user", "user"], ["cmd", "cmd", "a command"], ["tea", "tea"], ["coord", "app.Coordinator"], ["ref", "app.refresher"], ["views", "views.Dashboard", "every view"]],
      steps: [
        ["cmd", "coord", "msg", "DataRefreshRequestedMsg", "A closure never calls refreshData() itself (rule 1); it asks.", "app.refreshRequestedCmd"],
        ["tea", "ref", "cmd", "fullSeeded(previewID)", "Reads everything, plus the new fingerprint baseline so the next tick does not reload it again.", "app.refresher.fullSeeded"],
        ["ref", "coord", "msg", "configFingerprintMsg{stamps, data}", "", "app.Coordinator.handleConfigFingerprint"],
        ["coord", "views", "call", "handleDataLoaded → broadcastData", "", "app.Coordinator.handleDataLoaded"],
        ["cmd", "coord", "msg", "ErrorMsg", "Sets the banner (c.err).", "app.Coordinator.handleSpecialMessages"],
        ["user", "coord", "key", "the next keypress", "Clears the banner. Navigation alone never does, so an error batched with a navigation survives (rule 4).", "app.Coordinator.Update"]
      ]
    },
    editcard: {
      title: "Open the card editor", summary: "The safe way to change screens: prepare the target view in the callback, navigate from a message.",
      participants: [["user", "user"], ["dash", "views.Dashboard"], ["coord", "app.Coordinator"], ["ed", "views.CardEditor"], ["tea", "tea"]],
      steps: [
        ["user", "dash", "key", "e on a card", "", "views.Dashboard.handleEdit"],
        ["dash", "coord", "call", "OnEditCard(card)", "Still inside the dashboard's Update.", "app.Coordinator.dashboardCallbacks"],
        ["coord", "ed", "call", "SetCard(card)", "Allowed: same goroutine, and it does not change the current view.", "views.CardEditor.SetCard"],
        ["coord", "tea", "return", "command → NavigateMsg", "Navigating here would store the dashboard under the editor's key (rule 2).", "app.Coordinator.dashboardCallbacks"],
        ["tea", "coord", "msg", "NavigateMsg{CardEditor}", "", "app.Coordinator.navigateTo"],
        ["coord", "ed", "call", "Init()", "", "views.CardEditor.Init"]
      ]
    },
    savecard: {
      title: "Save an ID card", summary: "Validate and write id_cards.yaml off the UI goroutine.",
      participants: [["user", "user"], ["ed", "views.CardEditor"], ["tea", "tea"], ["repo", "config.FileRepository"], ["disk", "disk"], ["coord", "app.Coordinator"]],
      steps: [
        ["user", "ed", "key", "Enter", "", "views.CardEditor.handleSubmit"],
        ["ed", "ed", "self", "Card(): overlay the form on the loaded card", "Fields the form does not show survive the save (rule 5).", "views.CardEditor.Card"],
        ["ed", "tea", "return", "OnSave(card) command", "", "app.Coordinator.cardEditorCallbacks"],
        ["tea", "repo", "cmd", "CreateIDCard / UpdateIDCard", "", "config.FileRepository.UpdateIDCard"],
        ["repo", "repo", "self", "ValidateCard", "The workspace, skills, personas and MCPs must exist.", "config.FileRepository.ValidateCard"],
        ["repo", "disk", "io", "WriteAtomic(id_cards.yaml)", "Under the repository mutex.", "config.FileRepository.WriteAtomic"],
        ["repo", "coord", "msg", "CardSavedMsg", "Error: banner, stay. Success: refresh and back to the dashboard.", "app.Coordinator.handleSpecialMessages"]
      ]
    },
    deletecard: {
      title: "Delete an ID card", summary: "A card a session still points at cannot be deleted.",
      participants: [["user", "user"], ["dash", "views.Dashboard"], ["tea", "tea"], ["act", "app.sessionActions"], ["mgr", "session.Manager"], ["repo", "config.FileRepository"], ["coord", "app.Coordinator"]],
      steps: [
        ["user", "dash", "key", "Backspace on a card, confirm", "", "views.Dashboard.handleDeleteCard"],
        ["dash", "tea", "return", "OnDeleteCard → deleteCard command", "", "app.sessionActions.deleteCard"],
        ["tea", "act", "cmd", "deleteCard(cardID)", "", "app.sessionActions.deleteCard"],
        ["act", "mgr", "call", "List(): who references it?", "config cannot see sessions, so the guard lives here.", "session.Manager.List"],
        ["act", "repo", "call", "DeleteIDCard(id)", "Only if no session references the card.", "config.FileRepository.DeleteIDCard"],
        ["act", "coord", "msg", "DataRefreshRequestedMsg or ErrorMsg", "", "app.Coordinator.handleSpecialMessages"]
      ]
    },
    source: {
      title: "Add a source directory", summary: "Read, change, write and re-sync a registry's sources under one lock.",
      participants: [["user", "user"], ["sm", "views.SourceManager"], ["coord", "app.Coordinator"], ["tea", "tea"], ["repo", "config.FileRepository"], ["disk", "disk"]],
      steps: [
        ["user", "sm", "key", "add a skills directory", "", "views.SourceManager.Update"],
        ["sm", "coord", "call", "SaveSimpleSource(path) → saveSimpleSource", "Chooses UpdateSkillSources, UpdatePersonaSources or UpdatePuppetSources for this screen's type.", "app.Coordinator.saveSimpleSource"],
        ["tea", "repo", "cmd", "UpdateSkillSources(mutate)", "Read, apply, write and re-sync under one lock. Never list, change and save in separate calls (rule 6).", "config.FileRepository.UpdateSkillSources"],
        ["repo", "disk", "io", "skills_sources.yaml, rescan → skills.yaml", "", "config.FileRepository.syncSkillsFromSourcesLocked"],
        ["repo", "coord", "msg", "refresh, or refresh + warning, or ErrorMsg", "A save that worked but whose re-sync failed is *config.SyncFailedError: refresh and warn.", "app.Coordinator.sourceUpdateMsg"],
        ["coord", "sm", "call", "SetData → reload the list", "", "views.SourceManager.SetData"]
      ]
    },
    mcp: {
      title: "Save an MCP server", summary: "Parse the form, upsert into mcps.yaml, go back.",
      participants: [["user", "user"], ["form", "views.MCPForm"], ["tea", "tea"], ["repo", "config.FileRepository"], ["coord", "app.Coordinator"]],
      steps: [
        ["user", "form", "key", "Enter", "", "views.MCPForm.handleSubmit"],
        ["form", "form", "self", "parse args, env, headers", "Comma-separated with \\ escaping; a bad key=value list shows an error in the form.", "views.MCPForm.handleSubmit"],
        ["form", "tea", "return", "OnSave(mcp) command", "", "app.Coordinator.mcpCallbacks"],
        ["tea", "repo", "cmd", "UpsertMCP(mcp)", "", "config.FileRepository.UpsertMCP"],
        ["repo", "coord", "msg", "MCPSavedMsg", "Refresh and go back to the previous view. The card editor auto-selects the mcp-cli skill when an MCP is added.", "app.Coordinator.handleSpecialMessages"]
      ]
    },
    settings: {
      title: "Save settings", summary: "Write app_settings.yaml and push the new settings to the card editor.",
      participants: [["user", "user"], ["set", "views.SettingsView"], ["tea", "tea"], ["repo", "config.FileRepository"], ["coord", "app.Coordinator"], ["ed", "views.CardEditor"]],
      steps: [
        ["user", "set", "key", "save", "", "views.SettingsView.Update"],
        ["set", "tea", "return", "OnSave(settings) command", "", "app.Coordinator.settingsCallbacks"],
        ["tea", "repo", "cmd", "SaveAppSettings(settings)", "", "config.FileRepository.SaveAppSettings"],
        ["repo", "coord", "msg", "SettingsSavedMsg", "", "app.Coordinator.handleSpecialMessages"],
        ["coord", "ed", "call", "c.settings = …; SetSettings", "Decides whether the persona and MCP sections are shown.", "views.CardEditor.SetSettings"],
        ["coord", "coord", "self", "refresh + navigate to dashboard", "", "app.Coordinator.navigateTo"]
      ]
    }
  },

  rules: [
    { n: "1", title: "No tea.Cmd closure touches coordinator state", ref: "app.Coordinator", why: "A command's closure runs on its own goroutine. Reading or writing c.views, c.preview, c.currentView, c.err or c.settings there is a concurrent access that can crash the app. Snapshot what you need before returning the closure.", check: "grep -n \"c.refreshData()\" internal/app/callbacks.go must be empty; TestCommandClosuresDoNotRaceCoordinator under -race (a backstop, not a proof).", flows: ["tick", "launch"] },
    { n: "2", title: "Never navigate from inside a view's Update", ref: "app.Coordinator.navigateTo", why: "Update writes the view back into c.views[c.currentView] after the view returns. Navigating during that call stores the old view under the new key; it once replaced the dashboard with the agent-switch view.", check: "Navigate from handleSpecialMessages / handleSessionMessages. See navigateTo's doc comment.", flows: ["editcard", "launch"] },
    { n: "3", title: "DataLoadedMsg is always a full snapshot", ref: "messages.DataLoadedMsg", why: "A nil field means \"that registry could not be read this round, keep what you have\". Wiping a list makes CapabilityList.SelectedIDs() drop items, and the next save silently loses the card's skills.", check: "Every SetData guards each field with if msg.X != nil; partial updates use PreviewDataMsg or SessionsRefreshedMsg.", flows: ["tick", "preview"] },
    { n: "4", title: "The error banner clears on the next keypress", ref: "app.Coordinator.Update", why: "Navigation leaves it alone, so an ErrorMsg batched beside a NavigateMsg survives in either order.", check: "TestErrorSurvivesNavigationInEitherOrder, TestKeyPressClearsBanner.", flows: ["mutation"] },
    { n: "5", title: "The card editor overlays, it does not rebuild", ref: "views.CardEditor.Card", why: "Card() starts from the loaded card and assigns only the fields the form edits, so SystemInstruction, Env and timestamps survive a save.", check: "TestCardEditorPreservesUnmanagedFields.", flows: ["savecard"] },
    { n: "6", title: "Config read-modify-write goes through Update*Sources(fn)", ref: "config.FileRepository.UpdateSkillSources", why: "Listing, mutating and saving in separate calls means a failed read persists a truncated list. A save whose re-sync failed returns *config.SyncFailedError: refresh and warn.", check: "Use the Update*Sources functions; never compose List + Save yourself.", flows: ["source"] },
    { n: "7", title: "Exported repository methods lock; *Locked helpers assume the lock", ref: "config.FileRepository", why: "Go mutexes are not reentrant. A locked method calling another locked method deadlocks the whole TUI.", check: "Internal calls use the *Locked variants.", flows: [] },
    { n: "8", title: "\"Started, but the record didn't save\" is not a failure", ref: "session.Manager.Launch", why: "Launch, Relaunch and SwitchAgent wrap a post-spawn persist error in ErrPersistFailed. The agent is running; the UI still offers attach and shows a warning.", check: "Callers use errors.Is(err, session.ErrPersistFailed).", flows: ["launch", "relaunch"] },
    { n: "9", title: "Rendered output keeps its shape", ref: "components.TruncateHead", why: "Fit content by measured width, never bytes, and never let lipgloss word-wrap a row that must stay on one line: one extra column makes an outer border wrap the whole row.", check: "Layout sweeps in *_layout_test.go with uitest.AssertExactRect / AssertWithin.", flows: [] },
    { n: "10", title: "A sized component keeps its elements in place", ref: "components.CapabilityList.viewSized", why: "Where an element sits depends on the size the component was given, never on what it shows (page, scroll, search). Pad lists to their width, join left-aligned, reserve rows for optional elements.", check: "TestWorkspacePickerKeepsColumnsAcrossPages, TestCapabilityListItemsKeepTheirRows, TestCardEditorListsKeepTheirPlace.", flows: [] }
  ],

  tasks: [
    { q: "Add a new agent CLI", a: "Embed GenericCLIDriver and override only what differs (agent folder, instruction file, model flag); register it in driver/registry.go.", refs: ["driver.GenericCLIDriver", "driver.ClaudeCodeDriver", "driver.NewBuiltinRegistry"] },
    { q: "Change what gets written into a session's workspace", a: "The driver's SetAgentsMD, SetSkills, SetPersonas, SelectMCPs, and syncDriverFiles, which also clears generated files in temp workspaces.", refs: ["driver.GenericCLIDriver.SetAgentsMD", "session.Manager.syncDriverFiles"] },
    { q: "Change session lifecycle or status detection", a: "lifecycle.go, watch.go, io.go and IsPromptLike. Mind the generation counter and the activating reservation; goroutines that write to disk must be stoppable by Shutdown.", refs: ["session.Manager", "session.Manager.watch", "driver.IsPromptLike"], flow: "watch" },
    { q: "Change attach / detach / keys during attach", a: "app/attach.go.", refs: ["app.attachExec"], flow: "attach" },
    { q: "Add a field to ID cards", a: "types.go and validation.go; if the form edits it, CardEditor's SetCard, Card and focusTargetTable. Rule 5.", refs: ["config.IDCard", "views.CardEditor.Card"] },
    { q: "Add a config registry or source type", a: "repository.go interfaces and filename, a new config/*.go, Update*Sources, and Fingerprint's file list.", refs: ["config.Repository", "config.registryFiles"] },
    { q: "Add a screen", a: "A new type in ui/views; register it in initViews, add a ViewType, wire its callbacks in callbacks.go.", refs: ["app.Coordinator.initViews", "app.ViewType"], flow: "editcard" },
    { q: "Add a message", a: "Declare it in messages.go, handle it in handleSpecialMessages or handleSessionMessages. Partial data gets its own type (rule 3).", refs: ["app.Coordinator.handleSpecialMessages", "app.Coordinator.handleSessionMessages"] },
    { q: "Change how cards look", a: "RenderCardItem, the puppet renderer and the palette. Rule 9 and the card sweep test.", refs: ["components.RenderCardItem"] },
    { q: "Add or change a puppet", a: "Built-in art in ascii/characters; user puppets are .puppet files found in puppet source dirs.", refs: ["characters.TextPuppet", "config.FileRepository.SyncPuppetsFromSources"] },
    { q: "Change dashboard layout or keys", a: "dashboard.go (footerText for the footer), then the key tables in project.md §4 and kankuro_specs.md.", refs: ["views.Dashboard.handleKey", "views.Dashboard.footerText"] },
    { q: "Put user text into a fixed-width slot", a: "DisplayText first, then TruncateHead (keep the start) or TruncateTail (keep the end: paths, text being typed).", refs: ["components.DisplayText", "components.TruncateHead", "components.TruncateTail"] },
    { q: "Change the minimum terminal size", a: "views/constants.go and app/size_guard.go.", refs: ["app.Coordinator.tooSmall"], flow: "resize" },
    { q: "Change refresh behaviour", a: "refresh.go for the commands, handlers.go for applying results, and config.Fingerprint.", refs: ["app.refresher", "config.FileRepository.Fingerprint"], flow: "tick" },
    { q: "Add a session-level action (Phase 1 linking)", a: "A method on sessionActions returning a tea.Cmd that takes coordinator state by value; a message and a handler in coordinator.go; the trigger in callbacks.go.", refs: ["app.sessionActions"], flow: "relaunch" },
    { q: "Change the preview subscription", a: "previewSub and the preview handlers.", refs: ["app.previewSub", "app.Coordinator.handleSessionOutput"], flow: "preview" },
    { q: "Change how the IDE is opened", a: "ide_launcher.go; the command comes from app settings and is passed by value into launchIDE.", refs: ["session.IDELauncher"], flow: "ide" }
  ],

  storage: {
    root: "~/.config/kankuro/",
    tree: [["id_cards.yaml", ""], ["skills.yaml", "registry"], ["skills_sources.yaml", "dirs it is scanned from"], ["personas.yaml", ""], ["persona_sources.yaml", ""], ["puppets.yaml", ""], ["puppet_sources.yaml", ""], ["workspaces.yaml", ""], ["workspace_sources.yaml", ""], ["mcps.yaml", "entered by hand"], ["app_settings.yaml", "IDE command, editor sections"], ["sessions/<id>.json", "one record per session"], ["tmp/<id>/", "temp workspace copies"]],
    facts: [
      "Override the root with KANKURO_CONFIG_DIR or --config-dir; temp copies with KANKURO_TMP_DIR.",
      "All YAML and session JSON go through fsutil.WriteFileAtomic, under the repository or manager mutex.",
      "Loading a session always marks it terminated.",
      "The debug log (--debug) goes to <user cache dir>/kankuro/kankuro.log. Without --debug nothing may write to stdout or stderr: it corrupts the TUI.",
      "When running the app by hand, point KANKURO_CONFIG_DIR and KANKURO_TMP_DIR at a throwaway directory. Reading the real config is fine; writing to it is not."
    ]
  },
  gates: ["go build ./...", "go vet ./...", "go test ./...", "go test -race ./internal/...", "golangci-lint run          # must exit 0", "", "# run the app isolated from your real config", "KANKURO_TMP_DIR=/tmp/kankuro-test/tmp \\", "KANKURO_CONFIG_DIR=/tmp/kankuro-test/config \\", "  go run ./cmd/kankuro --debug"],
  testfacts: [
    "Test the coordinator by driving it: NewCoordinator, Update(msg), then inspect View(), CurrentView(), Views(). Go through Update, not a callback, or you skip the write-back where rule 2 breaks.",
    "Make sure each new test can fail: remove the fix, watch it fail, put it back.",
    "Session-launching tests use newTestManager; testenv.Run fails the package if anything is left in its private temp root (about 3 s per package).",
    "Shape tests check that a render fits its box; position tests check that elements stay put as the content changes."
  ],
  glossary: [
    ["Coordinator goroutine", "The goroutine Bubble Tea runs Update and View on. Coordinator state may only be touched there."],
    ["Command (tea.Cmd)", "A function Bubble Tea runs on its own goroutine; its return value comes back as a message."],
    ["generation", "Per-session counter bumped on every activation, so goroutines of an old run stop touching the session."],
    ["activating", "Reservation set by Relaunch and SwitchAgent between the guard and activation."],
    ["Fingerprint", "mtime + size of each registry file; the tick re-reads config only when one changed."],
    ["Seeded refresh", "A full refresh that also stores the new fingerprint baseline."],
    ["waiting", "Status for an agent whose last output looks like a prompt (driver.IsPromptLike)."],
    ["ModalState", "Embeddable open / confirmed / cancelled bookkeeping shared by every modal."],
    ["CoordinatorRows", "Rows the coordinator keeps back before sizing a view (4); one definition in views/constants.go."],
    ["Driver files", "What a driver writes into a workspace: AGENTS.md or CLAUDE.md, skills, personas, .kankuro/mcp-servers.json."],
    ["mcp-cli", "The one built-in skill, embedded in the binary; it teaches agents to use .kankuro/mcp-servers.json."],
    ["Shape test / position test", "Shape: the render fits its box. Position: elements stay put as content changes."]
  ]
};
