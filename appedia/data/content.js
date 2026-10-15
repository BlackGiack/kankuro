window.APPEDIA_CONTENT = {
  "app": "KANKURO",
  "title": "Kankuro Appedia",
  "tagline": "A terminal stage where AI agents are the puppets",
  "language": "go",
  "rulesSource": "project.md §6",
  "tasksSource": "project.md §10",
  "externals": {
    "user": [
      "You",
      "keyboard"
    ],
    "tea": [
      "Bubble Tea",
      "UI loop"
    ],
    "pty": [
      "Agent process",
      "in a PTY"
    ],
    "disk": [
      "Disk",
      "~/.config/kankuro"
    ],
    "cmd": [
      "Command",
      "any tea.Cmd"
    ],
    "files": [
      "Workspace files",
      "source trees, copies and manifest"
    ],
    "process": [
      "External application",
      "Configured IDE or terminal"
    ]
  },
  "flowKinds": {
    "cmd": [
      "⟳ Command",
      "Bubble Tea runs it on its own goroutine"
    ],
    "go": [
      "⟳ Goroutine",
      "starts goroutines"
    ],
    "msg": [
      "Message",
      "comes back through Bubble Tea's Update"
    ]
  },
  "snapshot": "Working tree on top of 3bdb260, including independent terminal commands and shared workspace launching, 2026-10-15.",
  "pitch": "KANKURO is a terminal UI written in Go that orchestrates AI coding-agent CLIs: Claude Code today, any CLI through a generic driver. You describe an agent once as an ID card, launch as many sessions of it as you like, and switch between them from one dashboard.",
  "pitch2": "It is a Bubble Tea program. One goroutine runs Update and View; anything slow (disk, processes) runs in a command on another goroutine and comes back as a message. Each agent runs in its own pseudo-terminal and its record is saved as JSON, so the list survives a restart.",
  "concepts": [
    [
      "ID card",
      "A saved agent template: driver, model, main and optional sub workspaces, capabilities, instructions, environment and appearance.",
      "config.IDCard"
    ],
    [
      "Session",
      "A running or stopped instance of a card: a process in a PTY, its scrollback, a JSON record.",
      "session.Session"
    ],
    [
      "Driver",
      "How to talk to one CLI: which files to write, how to spawn and kill it.",
      "driver.AgentDriver"
    ],
    [
      "Registries",
      "Workspaces, skills, personas and puppets discovered from source directories; MCP servers entered by hand.",
      "config.Repository"
    ],
    [
      "Coordinator",
      "The one object that routes every message between the screens and the rest.",
      "app.Coordinator"
    ],
    [
      "Status",
      "running, waiting (looks like it wants input) or terminated.",
      "session.Status"
    ],
    [
      "Sub workspace",
      "A selected registry entry copied into the session workspace’s tmp folder; managed edits persist across activations.",
      "session.subWorkspaceManifest"
    ],
    [
      "Workspace command",
      "An independently configured IDE or terminal command started in the selected session’s actual working directory.",
      "session.LaunchWorkspaceCommand"
    ]
  ],
  "path": [
    [
      "The package map: the coordinator sits between the screens and the processes.",
      "pkg-app"
    ],
    [
      "Zoom into the coordinator type to see what it owns.",
      "type-app.Coordinator"
    ],
    [
      "Watch a launch go from keypress to running process.",
      "flow-launch"
    ],
    [
      "Trace optional workspace copies and their ownership manifest.",
      "flow-sub-copy"
    ],
    [
      "Watch the 2-second tick keep the screen fresh without blocking.",
      "flow-tick"
    ],
    [
      "Read rules 1 and 2: goroutines and navigation, the easiest to break.",
      "rules"
    ]
  ],
  "map": {
    "viewBox": [
      0,
      0,
      1000,
      690
    ],
    "bands": [
      {
        "label": "ENTRY",
        "y": 16,
        "h": 70
      },
      {
        "label": "UI LOOP",
        "y": 100,
        "h": 84
      },
      {
        "label": "SCREENS · PROCESSES",
        "y": 198,
        "h": 200
      },
      {
        "label": "DATA · ART",
        "y": 412,
        "h": 148
      },
      {
        "label": "TEST-ONLY",
        "y": 604,
        "h": 76
      }
    ],
    "nodes": {
      "tea": [
        190,
        142
      ],
      "app": [
        500,
        142
      ],
      "views": [
        150,
        250
      ],
      "messages": [
        360,
        250
      ],
      "session": [
        660,
        250
      ],
      "components": [
        150,
        350
      ],
      "component": [
        330,
        350
      ],
      "styles": [
        480,
        350
      ],
      "driver": [
        660,
        350
      ],
      "pty": [
        870,
        350
      ],
      "ascii": [
        130,
        470
      ],
      "characters": [
        300,
        470
      ],
      "config": [
        480,
        470
      ],
      "fsutil": [
        670,
        470
      ],
      "disk": [
        870,
        470
      ],
      "root": [
        500,
        51
      ],
      "uitest": [
        330,
        642
      ],
      "testenv": [
        640,
        642
      ]
    },
    "external": {
      "tea": {
        "label": "Bubble Tea",
        "sub": "UI loop library"
      },
      "pty": {
        "label": "Agent CLI",
        "sub": "in a PTY"
      },
      "disk": {
        "label": "~/.config/kankuro",
        "sub": "YAML + JSON"
      }
    },
    "extraEdges": [
      [
        "root",
        "tea"
      ],
      [
        "tea",
        "app"
      ],
      [
        "driver",
        "pty"
      ],
      [
        "fsutil",
        "disk"
      ],
      [
        "testenv",
        "session"
      ],
      [
        "root",
        "config",
        "inject skill bytes"
      ]
    ]
  },
  "packages": {
    "app": {
      "label": "internal/app",
      "sub": "the coordinator",
      "role": "Coordinates screens, navigation, settings and error banners. Captures IDE and terminal command settings before handing blocking launches to sessionActions, which has no coordinator pointer.",
      "files": {
        "coordinator.go": "the Coordinator struct, Update, message routing",
        "navigation.go": "navigateTo (rule 2), initViews, resizeAll, broadcasts",
        "handlers.go": "UI-goroutine halves of refresh and preview",
        "refresh.go": "refresher: read commands, fingerprint check, tick",
        "preview.go": "previewSub: live output of the selected session",
        "session_actions.go": "Session actions, shared workspace-command lookup, IDE and terminal launch commands",
        "callbacks.go": "every view's callbacks",
        "attach.go": "raw-PTY attach/detach, Ctrl+K",
        "size_guard.go": "below 80×24: notice, keys dropped",
        "app.go": "tea.Model wrapper"
      },
      "notes": [
        "refresher and sessionActions hold no *Coordinator: a closure they build cannot touch coordinator state (rule 1 by construction).",
        "The error banner clears on the next keypress, never on navigation (rule 4)."
      ],
      "rules": [
        "1",
        "2",
        "3",
        "4",
        "8"
      ],
      "key": [
        "Coordinator",
        "refresher",
        "sessionActions",
        "previewSub",
        "attachExec"
      ],
      "sources": {
        "refs": [
          "app.Coordinator.launchTerminal",
          "app.Coordinator.handleWorkspaceCommandResult",
          "app.sessionActions.launchWorkspaceCommand"
        ],
        "files": [
          "internal/app/terminal_launch_test.go",
          "internal/app/command_race_internal_test.go"
        ]
      }
    },
    "views": {
      "label": "ui/views",
      "sub": "full screens",
      "role": "Owns full-screen interaction and layout. The dashboard gates IDE/terminal actions on Sessions focus and selection; Settings edits independent command inputs and fits their saved text.",
      "files": {
        "dashboard.go": "three panes, keys, search, footer",
        "card_editor.go": "card fields, one key switch, focused sub workspace clearing, source shortcuts, skills filter",
        "mcp_form.go": "MCP form driven by mcpFieldTable",
        "source_manager.go": "source directories per registry",
        "settings.go": "Independent IDE/terminal command inputs, focus, save/cancel and bounded rendering",
        "session_launcher.go": "progress while a session starts",
        "agent_switch.go": "pick a card for a terminated session",
        "workspace_picker_modal.go": "single/multiple selection, provisional checks, missing IDs",
        "puppet_selector_modal.go": "modal: choose a puppet",
        "agentsmd_editor_modal.go": "modal: AGENTS.md content or path",
        "constants.go": "MinTerminalWidth/Height, CoordinatorRows",
        "source_type.go": "per-registry labels"
      },
      "notes": [
        "A view that receives data implements messages.DataReceiver and guards every field with if msg.X != nil (rule 3).",
        "Workspace and sub workspace fields open the existing source manager with Ctrl+E. Sub selections are provisional until picker confirmation; registry refreshes retain checked missing IDs.",
        "Ctrl+X clears all selected IDs only while Sub Workspaces is focused in the form. Open modals own their keys. The focused row reserves space for ctrl+x=clear; Enter saves and Esc cancels. Existing workspace copies are retained."
      ],
      "rules": [
        "2",
        "3",
        "5",
        "9",
        "10"
      ],
      "key": [
        "Dashboard",
        "CardEditor",
        "MCPForm",
        "SourceManager",
        "SettingsView",
        "AgentSwitchView",
        "SessionLauncher"
      ],
      "sources": {
        "refs": [
          "views.editorKeyIdentity",
          "views.focusTargetTable",
          "views.NewCardEditor",
          "views.CardEditor.SetCard",
          "views.CardEditor.Reset",
          "views.CardEditor.fieldBlocks",
          "views.workspaceItem",
          "views.workspaceItem.Title",
          "views.NewSettingsView",
          "views.DashboardCallbacks",
          "views.Dashboard.handleLaunchSession",
          "views.Dashboard.footerText",
          "views.SettingsView.TerminalInput"
        ],
        "files": [
          "internal/ui/views/card_editor.go",
          "internal/ui/views/workspace_picker_modal.go",
          "internal/ui/views/settings_test.go",
          "internal/ui/views/dashboard_test.go",
          "internal/ui/views/dashboard_layout_test.go"
        ]
      }
    },
    "messages": {
      "label": "ui/messages",
      "sub": "every tea.Msg",
      "role": "Defines messages crossing the UI loop. TerminalLaunchedMsg carries external terminal launch completion or startup failure alongside the existing IDE completion message.",
      "files": {
        "messages.go": "all message types, DataReceiver, PreviewReceiver"
      },
      "notes": [
        "DataLoadedMsg is always a full snapshot; partial updates get their own type."
      ],
      "rules": [
        "3"
      ],
      "key": [
        "DataLoadedMsg",
        "PreviewDataMsg",
        "SessionsRefreshedMsg",
        "DataReceiver"
      ],
      "sources": {
        "refs": [
          "messages.TerminalLaunchedMsg"
        ],
        "files": [
          "internal/ui/messages/messages.go"
        ]
      }
    },
    "session": {
      "label": "internal/session",
      "sub": "process lifecycle",
      "role": "Owns agent session lifecycle, snapshots and workspace paths. Also supplies shared external workspace-command execution, with the existing IDELauncher retained as a compatibility wrapper.",
      "files": {
        "manager.go": "Manager, Session, Snapshot",
        "lifecycle.go": "Launch, Relaunch, SwitchAgent, Kill, Remove, Shutdown",
        "watch.go": "status polling every 2 s",
        "io.go": "output tracking and scrollback",
        "attach.go": "AttachHandle and the locked input writer",
        "storage.go": "session JSON, atomic",
        "workspace.go": "temporary workspace copies",
        "ide_launcher.go": "Compatible IDE API and command parser",
        "sub_workspaces.go": "source resolution, ownership manifest, copy staging/publication, rollback",
        "workspace_launcher.go": "Shared path resolution, placeholder replacement, direct process spawn and asynchronous reaping"
      },
      "notes": [
        "Every activation bumps a generation counter so goroutines of an older run stop touching the session.",
        "Relaunch and SwitchAgent reserve the session first (activating); Remove refuses while it is set.",
        "Sub workspace preparation holds one manager mutex across decisions and publication. Existing managed copies and deselected folders persist; independent app instances are outside this guarantee."
      ],
      "rules": [
        "8",
        "11"
      ],
      "key": [
        "Manager",
        "Session",
        "Snapshot",
        "AttachHandle",
        "IDELauncher"
      ],
      "sources": {
        "refs": [
          "session.subWorkspaceSource",
          "session.subWorkspaceCopy",
          "session.subWorkspaceManifest",
          "session.subWorkspaceTransaction",
          "session.IDELauncher",
          "session.LaunchWorkspaceCommand"
        ],
        "files": [
          "internal/session/sub_workspaces.go",
          "internal/session/workspace_test.go",
          "internal/session/workspace_launcher.go",
          "internal/session/ide_launcher_test.go"
        ]
      }
    },
    "components": {
      "label": "ui/components",
      "sub": "reusable pieces",
      "role": "Building blocks the views compose: lists, the card renderer, the preview, modals, text fitting.",
      "files": {
        "card_list.go": "ID-card rendering, pills, ClipBlock",
        "session_list.go": "sessions pane list",
        "session_preview.go": "live preview of agent output",
        "capability_list.go": "checklists, search and selected-only filtering, cursor normalization",
        "source_list.go": "source directories list",
        "modal.go": "Modal and ModalState",
        "text.go": "DisplayText, TruncateHead, TruncateTail",
        "renderer.go": "puppet rendering",
        "filter.go": "generic search",
        "source_item.go": "source list items",
        "constants.go": "layout constants"
      },
      "notes": [
        "Fit by display width, never bytes. Clip art, never wrap it (rule 9).",
        "Selected-only mode changes visible items, not the full item collection or selection map. Normalize the cursor after registry and selection changes."
      ],
      "rules": [
        "9",
        "10"
      ],
      "key": [
        "CardList",
        "SessionList",
        "SessionPreview",
        "CapabilityList",
        "ModalState"
      ]
    },
    "component": {
      "label": "ui/component",
      "sub": "base type",
      "role": "The Base type (Width, Height, SetSize) and the Model interface every view and component implements.",
      "files": {
        "component.go": "Base, Model"
      },
      "notes": [],
      "rules": [],
      "key": [
        "Base",
        "Model"
      ]
    },
    "styles": {
      "label": "ui/styles",
      "sub": "palette",
      "role": "The colour palette from crow_puppet_theme.md and shared lipgloss styles. Variables only, no types.",
      "files": {
        "styles.go": "colours and styles"
      },
      "notes": [],
      "rules": [],
      "key": []
    },
    "driver": {
      "label": "internal/driver",
      "sub": "one per CLI",
      "role": "One AgentDriver per CLI: which files to write into a workspace, how to spawn the CLI in a PTY, how to send it keys and how to kill it.",
      "files": {
        "driver.go": "GenericCLIDriver, ClaudeCodeDriver, Spawn, Kill, file sync, IsPromptLike",
        "registry.go": "name → constructor",
        "factory.go": "construction helpers"
      },
      "notes": [
        "Kill signals the whole process group: SIGTERM, up to 2 s grace, then SIGKILL for anything left. A descendant that starts its own session escapes.",
        "Built-in skill deployment reads injected bytes from config. Missing content is reported as an error; no source file is loaded at runtime."
      ],
      "rules": [],
      "key": [
        "AgentDriver",
        "GenericCLIDriver",
        "ClaudeCodeDriver",
        "MapRegistry"
      ],
      "sources": {
        "refs": [
          "driver.deploySkill",
          "driver.writeBuiltinSkill"
        ],
        "files": [
          "internal/driver/driver.go",
          "internal/driver/driver_test.go"
        ]
      }
    },
    "ascii": {
      "label": "internal/ascii",
      "sub": "text-art engine",
      "role": "Canvas and shapes for drawing ASCII art.",
      "files": {
        "canvas.go": "drawing surface",
        "shape.go": "Shape interface",
        "options.go": "config options"
      },
      "notes": [],
      "rules": [],
      "key": [
        "Canvas",
        "Shape"
      ]
    },
    "characters": {
      "label": "ascii/characters",
      "sub": "puppets & faces",
      "role": "Built-in puppets (cat, darth, default, kankuro, strawhat, totoro), faces, and user puppets loaded from .puppet files.",
      "files": {
        "puppet.go": "built-in puppets",
        "face.go": "agent faces",
        "text_puppet.go": "user puppets"
      },
      "notes": [],
      "rules": [],
      "key": [
        "Face",
        "TextPuppet"
      ]
    },
    "config": {
      "label": "internal/config",
      "sub": "YAML repository",
      "role": "Owns YAML-backed registries and settings with locked atomic writes. AppSettings gains an optional terminal_command string; absent fields retain defaults without a migration.",
      "files": {
        "repository.go": "FileRepository, interfaces, settings",
        "storage.go": "WriteAtomic, Fingerprint",
        "types.go": "IDCard, Workspace, Skill, Persona, MCPServer, Puppet, AppSettings",
        "id_cards.go": "cards",
        "skills.go": "skills",
        "personas.go": "personas",
        "mcps.go": "MCP servers",
        "workspaces.go": "workspaces",
        "skill_sources.go": "skill sources",
        "persona_sources.go": "persona sources",
        "puppet_sources.go": "puppet sources",
        "sync.go": "scanning source dirs",
        "builtins.go": "built-in metadata and startup-injected skill content",
        "validation.go": "card validation",
        "validation_helpers.go": "validation helpers",
        "errors.go": "SyncFailedError"
      },
      "notes": [
        "Cards reference workspaces and capabilities by ID, never by path.",
        "MCP-CLI bytes are supplied by main, protected by a separate RWMutex, and cloned on injection and lookup. Metadata listing does not require content injection."
      ],
      "rules": [
        "6",
        "7",
        "12"
      ],
      "key": [
        "FileRepository",
        "Repository",
        "IDCard",
        "AppSettings"
      ],
      "sources": {
        "refs": [
          "config.AppSettings"
        ],
        "files": [
          "internal/config/app_settings_test.go"
        ]
      }
    },
    "fsutil": {
      "label": "internal/fsutil",
      "sub": "atomic writes",
      "role": "WriteFileAtomic: temp file in the same directory, fsync, rename, cleanup on error. Used by config and session.",
      "files": {
        "atomic.go": "WriteFileAtomic"
      },
      "notes": [
        "A crash leaves at worst a .<name>.<rand>.tmp file, which loaders ignore."
      ],
      "rules": [],
      "key": []
    },
    "root": {
      "label": "kankuro",
      "sub": "root main.go",
      "role": "The root executable embeds the canonical MCP-CLI skill, injects its bytes into config, wires the repository/drivers/session manager/UI, and owns application shutdown.",
      "files": {
        "main.go": "embed, config injection, flags, wiring, signals, shutdown"
      },
      "notes": [
        "Main injects built-in content before creating the driver registry; config never imports the executable package.",
        "Every normal exit after the TUI run calls Manager.Shutdown. stdout/stderr logging must respect whether the TUI owns the terminal."
      ],
      "rules": [
        "12"
      ],
      "key": [],
      "sources": {
        "refs": [
          "root.main",
          "root.run",
          "root.logDir",
          "root.stopSessions"
        ],
        "files": [
          "main.go",
          "main_test.go",
          "go.mod",
          "README.md",
          "kankuro_skills/mcp-cli/SKILL.md"
        ]
      }
    },
    "uitest": {
      "label": "ui/uitest",
      "sub": "shape assertions",
      "test": true,
      "role": "Test-only helpers for layout tests: exact/bounded shape assertions, regions, adversarial content.",
      "files": {
        "shape.go": "AssertExactRect, AssertWithin, Region, AdversarialNames"
      },
      "notes": [
        "Shape tests check fit; position tests check that elements stay put. A list-like component needs both."
      ],
      "rules": [
        "9",
        "10"
      ],
      "key": []
    },
    "testenv": {
      "label": "internal/testenv",
      "sub": "TestMain sandbox",
      "test": true,
      "role": "Runs a package's tests in a private TMPDIR and KANKURO_TMP_DIR and fails the package if anything is left behind.",
      "files": {
        "testenv.go": "Run(m, name)"
      },
      "notes": [
        "Session-launching tests build managers with newTestManager, which registers a Shutdown cleanup."
      ],
      "rules": [],
      "key": []
    },
    "tea": {
      "label": "Bubble Tea",
      "sub": "UI loop library",
      "ext": true,
      "role": "charm.land/bubbletea/v2. Runs Update(msg) and View() on one goroutine, and every tea.Cmd on its own goroutine, feeding the result back as a message.",
      "files": {},
      "notes": [
        "That last fact is what rule 1 is about."
      ],
      "rules": [
        "1"
      ],
      "key": []
    },
    "pty": {
      "label": "Agent CLI",
      "sub": "in a PTY",
      "ext": true,
      "role": "The agent process (claude or any CLI), started with creack/pty in its own session and process group.",
      "files": {},
      "notes": [],
      "rules": [],
      "key": []
    },
    "disk": {
      "label": "~/.config/kankuro",
      "sub": "YAML + JSON",
      "ext": true,
      "role": "Stores configuration YAML and session JSON. Session working directories also hold tmp/<sub workspace basename> copies and .kankuro/sub-workspaces.json ownership metadata.",
      "files": {},
      "notes": [],
      "rules": [],
      "key": []
    }
  },
  "flowGroups": [
    {
      "name": "Start & stop",
      "flows": [
        "startup",
        "quit",
        "resize"
      ]
    },
    {
      "name": "Sessions",
      "flows": [
        "launch",
        "attach",
        "relaunch",
        "switch",
        "kill",
        "remove",
        "watch",
        "ide",
        "terminal",
        "workspace-command-errors",
        "failed-activation"
      ]
    },
    {
      "name": "Workspace preparation",
      "flows": [
        "sub-copy",
        "copy-rollback",
        "builtin-skill"
      ]
    },
    {
      "name": "Keeping the screen fresh",
      "flows": [
        "tick",
        "preview",
        "mutation"
      ]
    },
    {
      "name": "Card editor",
      "flows": [
        "editcard",
        "temporary-toggle",
        "skill-filter",
        "sub-picker",
        "sub-clear",
        "picker-refresh",
        "workspace-sources",
        "savecard",
        "deletecard"
      ]
    },
    {
      "name": "Registries & settings",
      "flows": [
        "source",
        "mcp",
        "settings",
        "settings-edit"
      ]
    }
  ],
  "flows": {
    "startup": {
      "title": "Startup",
      "summary": "The root command injects the embedded skill before opening config, loading sessions and starting Bubble Tea.",
      "participants": [
        [
          "main",
          "root.main",
          "main.go"
        ],
        [
          "content",
          "config.SetBuiltinMCPCLISkillContent",
          "Skill bytes"
        ],
        [
          "repo",
          "config.FileRepository"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "ref",
          "app.refresher"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "main",
          "content",
          "call",
          "Inject skill",
          "Main owns the go:embed bytes; config stores a clone under its content mutex before any agent can deploy a skill.",
          "config.SetBuiltinMCPCLISkillContent"
        ],
        [
          "main",
          "main",
          "self",
          "Register drivers",
          "Register the Claude driver; unknown command names use the generic driver.",
          "driver.NewBuiltinRegistry"
        ],
        [
          "main",
          "main",
          "call",
          "run(registry)",
          "Parse flags, configure optional debug logging and initialize the application.",
          "root.run"
        ],
        [
          "main",
          "repo",
          "call",
          "Open config",
          "Opens the YAML repository and creates missing files.",
          "config.NewFileRepository"
        ],
        [
          "main",
          "repo",
          "call",
          "Sync skills",
          "Rescans the skill source directories into skills.yaml.",
          "config.FileRepository.SyncSkillsFromSources"
        ],
        [
          "main",
          "mgr",
          "call",
          "NewManager(dir, repo)",
          "Builds the session registry.",
          "session.NewManager"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Load session records",
          "Every loaded session is marked terminated: a process cannot survive a restart. Corrupt files are skipped and reported once.",
          "session.loadSessions"
        ],
        [
          "main",
          "coord",
          "call",
          "app.New → coordinator",
          "Builds every view with its callbacks.",
          "app.NewCoordinator"
        ],
        [
          "main",
          "tea",
          "call",
          "NewProgram.Run()",
          "Bubble Tea takes the terminal. From here on, everything goes through Update and View.",
          "root.run"
        ],
        [
          "tea",
          "coord",
          "call",
          "Init()",
          "Returns a batch of commands.",
          "app.Coordinator.Init"
        ],
        [
          "tea",
          "ref",
          "cmd",
          "Read registries",
          "Each runs on its own goroutine: read every registry, rescan puppet sources, report session-load errors.",
          "app.refresher.fullSeeded"
        ],
        [
          "ref",
          "coord",
          "msg",
          "Config snapshot",
          "Applied on the UI goroutine: the baseline is stored and the data broadcast to every view.",
          "app.Coordinator.handleConfigFingerprint"
        ],
        [
          "tea",
          "coord",
          "msg",
          "Tick every 2s",
          "Starts the refresh cadence (see the tick flow).",
          "app.tick"
        ]
      ],
      "sources": {
        "refs": [
          "root.main",
          "root.logDir",
          "config.GetBuiltinSkillContent"
        ],
        "files": [
          "main.go",
          "main_test.go",
          "go.mod",
          "README.md",
          "kankuro_skills/mcp-cli/SKILL.md"
        ]
      }
    },
    "quit": {
      "title": "Quit",
      "summary": "Every exit path ends in Shutdown, which kills running agents before KANKURO exits.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "main",
          "root.run",
          "main.go"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "drv",
          "driver.GenericCLIDriver"
        ],
        [
          "pty",
          "pty"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "q or Ctrl+C",
          "Closing the terminal window sends SIGHUP, which main turns into p.Quit(). SIGTERM becomes a quit inside Bubble Tea.",
          "views.Dashboard.handleKey"
        ],
        [
          "dash",
          "tea",
          "return",
          "tea.Quit",
          "",
          "views.Dashboard.handleKey"
        ],
        [
          "tea",
          "main",
          "return",
          "Run() returns",
          "Whatever it returned, the shutdown below runs.",
          "root.run"
        ],
        [
          "main",
          "mgr",
          "call",
          "RunningCount() → one stderr line",
          "Only when sessions are running. stderr is safe now: the TUI is gone.",
          "root.stopSessions"
        ],
        [
          "main",
          "mgr",
          "call",
          "Shutdown()",
          "Sets closed: Launch, Relaunch and SwitchAgent now return ErrManagerClosed.",
          "session.Manager.Shutdown"
        ],
        [
          "mgr",
          "drv",
          "go",
          "Kill(id) for each live session, in parallel",
          "The wait is about one grace period in total, not one per session.",
          "session.Manager.Kill"
        ],
        [
          "drv",
          "pty",
          "io",
          "SIGTERM to the process group",
          "The agent gets a chance to save its state.",
          "driver.GenericCLIDriver.Kill"
        ],
        [
          "drv",
          "pty",
          "io",
          "after ≤ 2 s: SIGKILL if the group is alive",
          "Also reaches children that ignore SIGTERM and SIGHUP.",
          "driver.groupAlive"
        ],
        [
          "mgr",
          "disk",
          "io",
          "records saved as terminated",
          "",
          "session.Manager.persistSessionLockedErr"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "wait activations · close stop · wait watchers",
          "A launch caught mid-spawn kills its own process; watchers exit without a final write.",
          "session.Manager.Shutdown"
        ]
      ]
    },
    "resize": {
      "title": "Resize & too-small terminal",
      "summary": "How views get their size, and what happens below 80×24.",
      "participants": [
        [
          "tea",
          "tea"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "settings",
          "views.SettingsView"
        ]
      ],
      "steps": [
        [
          "tea",
          "coord",
          "msg",
          "tea.WindowSizeMsg",
          "",
          "app.Coordinator.handleSpecialMessages"
        ],
        [
          "coord",
          "dash",
          "call",
          "resizeAll → SetSize(w, h − CoordinatorRows)",
          "Every view gets the terminal minus the rows the coordinator keeps back; the dashboard adds them back to fill the screen exactly.",
          "app.Coordinator.resizeAll"
        ],
        [
          "coord",
          "settings",
          "call",
          "Size both command inputs",
          "Account for border and padding; no negative input widths.",
          "views.SettingsView.SetSize"
        ],
        [
          "tea",
          "coord",
          "call",
          "View()",
          "",
          "app.Coordinator.View"
        ],
        [
          "coord",
          "coord",
          "self",
          "tooSmall() → tooSmallView()",
          "Below 80×24 every screen is replaced by a centred notice sized to the terminal.",
          "app.Coordinator.tooSmallView"
        ],
        [
          "dash",
          "dash",
          "self",
          "footerText: compact session hints",
          "At width 80 keep t=Terminal and q=Quit alongside useful session actions.",
          "views.Dashboard.footerText"
        ],
        [
          "tea",
          "coord",
          "msg",
          "a keypress while too small",
          "",
          "app.Coordinator.Update"
        ],
        [
          "coord",
          "coord",
          "self",
          "guardKey: dropped (q / Ctrl+C still quit)",
          "A blind keypress must not launch or delete anything. Non-key messages keep flowing, so the screen is current when the terminal grows.",
          "app.Coordinator.guardKey"
        ]
      ],
      "sources": {
        "refs": [
          "views.SettingsView.renderEdit",
          "views.SettingsView.renderDisplay"
        ],
        "files": [
          "internal/ui/views/settings_test.go",
          "internal/ui/views/dashboard_layout_test.go"
        ]
      }
    },
    "launch": {
      "title": "Launch a session",
      "summary": "Enter on a card: prepare the launcher screen on the UI goroutine, then resolve, copy, write files, spawn and save on a command goroutine.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "act",
          "app.sessionActions"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "drv",
          "driver.GenericCLIDriver"
        ],
        [
          "pty",
          "pty"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "Enter on a card",
          "",
          "views.Dashboard.handleEnter"
        ],
        [
          "dash",
          "coord",
          "call",
          "launchCard: prepare view and commands",
          "On the UI goroutine, prepare SessionLauncher and return launch/navigation commands; navigation itself happens in a message handler.",
          "app.Coordinator.launchCard"
        ],
        [
          "tea",
          "coord",
          "msg",
          "NavigateMsg → SessionLauncher",
          "Navigation happens in a message handler.",
          "app.Coordinator.navigateTo"
        ],
        [
          "tea",
          "act",
          "cmd",
          "launch(card)",
          "Blocking work from here on, on a command goroutine. sessionActions has no coordinator pointer (rule 1).",
          "app.sessionActions.launch"
        ],
        [
          "act",
          "act",
          "self",
          "resolveCapabilities(card)",
          "Reads the skill, persona and MCP registries once; a failed read aborts the launch instead of starting an agent with nothing.",
          "app.sessionActions.resolveCapabilities"
        ],
        [
          "act",
          "mgr",
          "call",
          "Launch(card, skills, personas, mcps)",
          "",
          "session.Manager.Launch"
        ],
        [
          "mgr",
          "disk",
          "io",
          "prepareWorkspace: copy to tmp/<id>/",
          "Only when the card asks for a temporary workspace.",
          "session.Manager.prepareWorkspace"
        ],
        [
          "mgr",
          "disk",
          "io",
          "prepareSubWorkspaces()",
          "Runs before driver files and spawn in both workspace modes. Existing managed copies retain edits; preparation failures prevent activation.",
          "session.Manager.prepareSubWorkspaces"
        ],
        [
          "mgr",
          "drv",
          "call",
          "syncDriverFiles",
          "Writes AGENTS.md or CLAUDE.md, skills, personas and .kankuro/mcp-servers.json into the workspace.",
          "session.Manager.syncDriverFiles"
        ],
        [
          "drv",
          "pty",
          "io",
          "Spawn: pty.Start, own session",
          "The card's env is merged over the parent's.",
          "driver.GenericCLIDriver.Spawn"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "activateSessionLocked + trackOutput + watch",
          "Bumps the generation counter and starts the per-activation goroutines.",
          "session.Manager.activateSessionLocked"
        ],
        [
          "mgr",
          "disk",
          "io",
          "saveSession (atomic)",
          "A failure here is ErrPersistFailed: started, just not saved (rule 8).",
          "session.saveSession"
        ],
        [
          "act",
          "coord",
          "msg",
          "SessionLaunchCompletedMsg",
          "The launcher shows the result; from there you can name the session or attach. A refresh follows.",
          "app.Coordinator.handleSessionMessages"
        ]
      ],
      "sources": {
        "refs": [
          "session.Manager.startSession",
          "session.Manager.cleanupNewTemporaryWorkspace"
        ],
        "files": [
          "internal/session/lifecycle.go"
        ]
      }
    },
    "attach": {
      "title": "Attach & detach",
      "summary": "Bubble Tea hands the raw terminal to the agent until Ctrl+K.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "exec",
          "app.attachExec"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "pty",
          "pty"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "a, or Enter on a running session",
          "",
          "views.Dashboard.handleAttach"
        ],
        [
          "dash",
          "coord",
          "call",
          "OnAttach(id) → attachSession",
          "",
          "app.Coordinator.attachSession"
        ],
        [
          "coord",
          "tea",
          "return",
          "tea.Exec(attachExec)",
          "Bubble Tea releases the terminal to the exec command.",
          "app.Coordinator.attachSession"
        ],
        [
          "tea",
          "exec",
          "call",
          "Run()",
          "",
          "app.attachExec.Run"
        ],
        [
          "exec",
          "mgr",
          "call",
          "Attach(id) → AttachHandle",
          "Scrollback, a one-shot header, the output channel and an input writer.",
          "session.Manager.Attach"
        ],
        [
          "exec",
          "exec",
          "self",
          "replay scrollback, then pumpOutput",
          "",
          "app.attachExec.pumpOutput"
        ],
        [
          "user",
          "exec",
          "key",
          "keystrokes",
          "",
          "app.attachExec.pumpInput"
        ],
        [
          "exec",
          "pty",
          "io",
          "forwardInput → SendInputBytes",
          "Through the driver's locked writer, never the raw fd: that would race Kill.",
          "app.attachExec.forwardInput"
        ],
        [
          "user",
          "exec",
          "key",
          "Ctrl+K",
          "Recognised as the legacy byte or as a Kitty keyboard-protocol sequence.",
          "app.detectKittyDetach"
        ],
        [
          "exec",
          "coord",
          "msg",
          "AttachFinishedMsg",
          "Bubble Tea takes the terminal back; a refresh follows.",
          "app.Coordinator.handleSessionMessages"
        ]
      ]
    },
    "relaunch": {
      "title": "Relaunch a session",
      "summary": "Enter on a terminated session restarts it in the same workspace, then attaches.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "act",
          "app.sessionActions"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "drv",
          "driver.GenericCLIDriver"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "disk",
          "files"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "Enter on a terminated session",
          "",
          "views.Dashboard.handleSessionEnter"
        ],
        [
          "dash",
          "tea",
          "return",
          "OnRelaunch(id) command",
          "The callback is the sessionActions method itself: no coordinator state involved.",
          "app.sessionActions.relaunch"
        ],
        [
          "tea",
          "act",
          "cmd",
          "relaunch(id)",
          "",
          "app.sessionActions.relaunch"
        ],
        [
          "act",
          "act",
          "self",
          "resolveCard(session.CardID)",
          "A failed read of id_cards.yaml is reported as such, not as \"card not found\".",
          "app.sessionActions.resolveCard"
        ],
        [
          "act",
          "mgr",
          "call",
          "Relaunch(id, card, …)",
          "",
          "session.Manager.Relaunch"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "reserveForActivation",
          "Must be terminated and not already starting; marks it activating so a second relaunch, a switch or a remove fails fast.",
          "session.Manager.reserveForActivation"
        ],
        [
          "mgr",
          "mgr",
          "call",
          "startSession in the same workspace",
          "Reuse the main session workspace; preparation, driver sync and spawn happen before activation.",
          "session.Manager.startSession"
        ],
        [
          "mgr",
          "disk",
          "io",
          "prepareSubWorkspaces()",
          "Preserve existing managed edits, add absent selected copies, and retain deselected folders.",
          "session.Manager.prepareSubWorkspaces"
        ],
        [
          "act",
          "coord",
          "msg",
          "SessionAttachRequestedMsg",
          "Carries a Warning when only the record failed to save.",
          "app.Coordinator.handleSessionMessages"
        ],
        [
          "coord",
          "coord",
          "self",
          "navigate to dashboard, attachSession",
          "",
          "app.Coordinator.attachSession"
        ]
      ]
    },
    "switch": {
      "title": "Switch agent",
      "summary": "Restart a terminated session with a different card, in the same workspace.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "sw",
          "views.AgentSwitchView"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "act",
          "app.sessionActions"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "disk",
          "files"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "s on a terminated session",
          "",
          "views.Dashboard.handleSwitch"
        ],
        [
          "dash",
          "coord",
          "msg",
          "AgentSwitchRequestedMsg",
          "",
          "app.Coordinator.handleSessionMessages"
        ],
        [
          "coord",
          "sw",
          "call",
          "SetSession, navigateTo(AgentSwitch)",
          "",
          "views.AgentSwitchView.SetSession"
        ],
        [
          "user",
          "sw",
          "key",
          "pick a card, Enter",
          "",
          "views.AgentSwitchView.handleKey"
        ],
        [
          "sw",
          "tea",
          "return",
          "OnSwitch → switchAgent command",
          "The switch must not run inside Update: it writes files and spawns a process.",
          "app.sessionActions.switchAgent"
        ],
        [
          "tea",
          "act",
          "cmd",
          "switchAgent(id, cardID)",
          "",
          "app.sessionActions.switchAgent"
        ],
        [
          "act",
          "mgr",
          "call",
          "SwitchAgent(id, card, …)",
          "Reserve the terminated session, prepare sub copies, replace instructions and spawn in the existing workspace.",
          "session.Manager.SwitchAgent"
        ],
        [
          "mgr",
          "disk",
          "io",
          "prepareSubWorkspaces()",
          "Preserve existing managed edits, add absent selected copies, and retain deselected folders.",
          "session.Manager.prepareSubWorkspaces"
        ],
        [
          "act",
          "coord",
          "msg",
          "AgentSwitchedMsg",
          "",
          "app.Coordinator.handleAgentSwitched"
        ],
        [
          "coord",
          "coord",
          "self",
          "refresh + SessionAttachRequestedMsg",
          "The attach handler navigates; a persist warning travels with it (rules 4 and 8).",
          "app.Coordinator.handleAgentSwitched"
        ]
      ]
    },
    "kill": {
      "title": "Kill a session",
      "summary": "x on a running session: SIGTERM, grace, SIGKILL, then mark it terminated.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "drv",
          "driver.GenericCLIDriver"
        ],
        [
          "pty",
          "pty"
        ],
        [
          "coord",
          "app.Coordinator"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "x on a running session",
          "Only offered for running or waiting sessions.",
          "views.Dashboard.handleKill"
        ],
        [
          "dash",
          "tea",
          "return",
          "OnKill(id) command",
          "",
          "app.Coordinator.dashboardCallbacks"
        ],
        [
          "tea",
          "mgr",
          "cmd",
          "Kill(id)",
          "Captures the driver and generation under the lock, then kills without holding it.",
          "session.Manager.Kill"
        ],
        [
          "mgr",
          "drv",
          "call",
          "Kill()",
          "",
          "driver.GenericCLIDriver.Kill"
        ],
        [
          "drv",
          "pty",
          "io",
          "SIGTERM group → ≤ 2 s → SIGKILL group",
          "",
          "driver.signalGroup"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "same activation? → terminated + persist",
          "If the session was removed or relaunched meanwhile, it is left alone, so a removed record is never re-created.",
          "session.Manager.Kill"
        ],
        [
          "mgr",
          "coord",
          "msg",
          "DataRefreshRequestedMsg (or ErrorMsg)",
          "",
          "app.Coordinator.handleSpecialMessages"
        ]
      ]
    },
    "remove": {
      "title": "Remove a session",
      "summary": "Remove a terminated session; temporary workspace copies are removed, original-workspace files remain.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "modal",
          "components.Modal"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "disk",
          "disk"
        ],
        [
          "coord",
          "app.Coordinator"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "Backspace or Delete on a terminated session",
          "",
          "views.Dashboard.handleDeleteSession"
        ],
        [
          "dash",
          "modal",
          "call",
          "open the confirm modal",
          "",
          "components.ModalState.Open"
        ],
        [
          "user",
          "dash",
          "key",
          "Enter or y",
          "",
          "views.Dashboard.confirmDelete"
        ],
        [
          "dash",
          "tea",
          "return",
          "OnRemove(id) command",
          "",
          "app.Coordinator.dashboardCallbacks"
        ],
        [
          "tea",
          "mgr",
          "cmd",
          "Remove(id)",
          "Refuses a running session, or one being started.",
          "session.Manager.Remove"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Delete record and temporary root",
          "A temporary session loses its enclosing workspace, including sub copies. Original-workspace sub copies and manifest remain.",
          "session.Manager.Remove"
        ],
        [
          "mgr",
          "coord",
          "msg",
          "DataRefreshRequestedMsg",
          "",
          "app.Coordinator.handleSpecialMessages"
        ]
      ]
    },
    "watch": {
      "title": "Status & output tracking",
      "summary": "Two goroutines per activation: one copies output into scrollback, one polls the status every 2 s.",
      "participants": [
        [
          "pty",
          "pty"
        ],
        [
          "drv",
          "driver.GenericCLIDriver"
        ],
        [
          "track",
          "session.Manager",
          "Manager · trackOutput"
        ],
        [
          "watch",
          "session.Manager",
          "Manager · watch"
        ],
        [
          "sub",
          "app.previewSub"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "pty",
          "drv",
          "io",
          "output bytes",
          "",
          "driver.GenericCLIDriver.streamOutput"
        ],
        [
          "drv",
          "track",
          "chan",
          "Output channel",
          "",
          "session.Manager.trackOutput"
        ],
        [
          "track",
          "track",
          "self",
          "append scrollback (≤ 256 KiB), last line",
          "Stops at once if the generation changed: it belongs to an older run.",
          "session.appendScrollback"
        ],
        [
          "track",
          "sub",
          "chan",
          "broadcast to subscribers",
          "Non-blocking: a slow subscriber drops chunks instead of stalling the agent.",
          "session.broadcast"
        ],
        [
          "watch",
          "drv",
          "call",
          "every 2 s: GetStatus()",
          "",
          "session.Manager.watch"
        ],
        [
          "watch",
          "watch",
          "self",
          "IsPromptLike → waiting or running",
          "Waiting means the last line looks like a prompt and no word arrived recently.",
          "driver.IsPromptLike"
        ],
        [
          "watch",
          "disk",
          "io",
          "terminated → persist, close subscribers",
          "Also stops when Shutdown closes the stop channel.",
          "session.Manager.watch"
        ]
      ]
    },
    "ide": {
      "title": "Open in IDE",
      "summary": "Press o in Sessions to start the configured IDE in the selected session’s actual workspace through shared command execution.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "act",
          "app.sessionActions"
        ],
        [
          "ide",
          "session.IDELauncher"
        ],
        [
          "launch",
          "session.LaunchWorkspaceCommand",
          "Workspace launcher"
        ],
        [
          "process",
          "process"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "o in Sessions",
          "Search consumes typing first; no action for card focus or an empty session selection.",
          "views.Dashboard.handleKey"
        ],
        [
          "dash",
          "dash",
          "self",
          "handleLaunchSession(OnLaunchIDE)",
          "Selection and nil-callback guards are shared with t.",
          "views.Dashboard.handleLaunchSession"
        ],
        [
          "dash",
          "coord",
          "call",
          "launchIDE(selected ID)",
          "Capture IDECommand on the UI goroutine before constructing a command.",
          "app.Coordinator.launchIDE"
        ],
        [
          "tea",
          "act",
          "cmd",
          "launchIDE(id, command)",
          "Blocking work executes off the coordinator goroutine.",
          "app.sessionActions.launchIDE"
        ],
        [
          "act",
          "act",
          "self",
          "launchWorkspaceCommand",
          "Look up the session at execution time; validate WorkingDir and configuration. Temporary sessions use their copies.",
          "app.sessionActions.launchWorkspaceCommand"
        ],
        [
          "act",
          "ide",
          "call",
          "IDELauncher.Launch(WorkingDir)",
          "Compatibility wrapper preserves ErrIDENotConfigured.",
          "session.IDELauncher.Launch"
        ],
        [
          "ide",
          "launch",
          "call",
          "LaunchWorkspaceCommand(command, path)",
          "Resolve an absolute existing workspace, replace every $PATH and parse command quoting.",
          "session.LaunchWorkspaceCommand"
        ],
        [
          "launch",
          "process",
          "io",
          "Start process, reap asynchronously",
          "Set the workspace as Dir even without a placeholder. No implicit shell or TUI standard streams.",
          "session.LaunchWorkspaceCommand"
        ],
        [
          "act",
          "coord",
          "msg",
          "IDELaunchedMsg{Err}",
          "Completion reports startup; later external process exit is not monitored by this action.",
          "app.Coordinator.handleSpecialMessages"
        ],
        [
          "coord",
          "coord",
          "self",
          "Apply result without navigation",
          "Errors become the existing banner.",
          "app.Coordinator.handleWorkspaceCommandResult"
        ]
      ],
      "sources": {
        "refs": [
          "session.SplitCommand"
        ],
        "files": [
          "internal/session/workspace_launcher.go",
          "internal/session/ide_launcher_test.go"
        ]
      }
    },
    "tick": {
      "title": "The 2-second tick",
      "summary": "Sessions refresh from memory every tick; config re-reads only when a file changed.",
      "participants": [
        [
          "tea",
          "tea"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "ref",
          "app.refresher"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "repo",
          "config.FileRepository"
        ],
        [
          "dash",
          "views.Dashboard"
        ]
      ],
      "steps": [
        [
          "tea",
          "coord",
          "msg",
          "tickMsg",
          "",
          "app.Coordinator.handleSpecialMessages"
        ],
        [
          "coord",
          "coord",
          "self",
          "snapshot preview ID + fingerprint",
          "Plain values: the commands below never see the coordinator (rule 1).",
          "app.Coordinator.refreshSessions"
        ],
        [
          "tea",
          "ref",
          "cmd",
          "sessionsOnly(previewID)",
          "",
          "app.refresher.sessionsOnly"
        ],
        [
          "ref",
          "mgr",
          "call",
          "List(), GetScrollback(id)",
          "Memory only, no disk.",
          "session.Manager.List"
        ],
        [
          "ref",
          "coord",
          "msg",
          "SessionsRefreshedMsg",
          "Applied to the dashboard only.",
          "app.Coordinator.handleSessionsRefreshed"
        ],
        [
          "tea",
          "ref",
          "cmd",
          "configIfChanged(baseline)",
          "",
          "app.refresher.configIfChanged"
        ],
        [
          "ref",
          "repo",
          "call",
          "Fingerprint(): stat each registry",
          "No parsing. Only if a stamp changed are the registries re-read.",
          "config.FileRepository.Fingerprint"
        ],
        [
          "ref",
          "coord",
          "msg",
          "configFingerprintMsg",
          "A new baseline, plus a full DataLoadedMsg when something changed.",
          "app.Coordinator.handleConfigFingerprint"
        ],
        [
          "coord",
          "dash",
          "call",
          "broadcastData → SetData",
          "Every DataReceiver gets it; a nil field means keep what you have (rule 3).",
          "app.Coordinator.broadcastData"
        ]
      ]
    },
    "preview": {
      "title": "Live preview",
      "summary": "The dashboard's centre pane streams the selected session's output.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "sub",
          "app.previewSub"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "pv",
          "components.SessionPreview"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "select another session",
          "",
          "views.Dashboard.syncPreviewSelection"
        ],
        [
          "dash",
          "coord",
          "msg",
          "SessionPreviewRequestedMsg",
          "",
          "app.Coordinator.handlePreviewRequest"
        ],
        [
          "coord",
          "sub",
          "call",
          "subscribe(id)",
          "Drops the previous subscription first.",
          "app.previewSub.subscribe"
        ],
        [
          "sub",
          "mgr",
          "call",
          "GetScrollback · SubscribeOutput",
          "",
          "session.Manager.SubscribeOutput"
        ],
        [
          "coord",
          "pv",
          "call",
          "broadcastPreview(PreviewDataMsg)",
          "A partial update: never through SetData (rule 3).",
          "app.Coordinator.broadcastPreview"
        ],
        [
          "tea",
          "sub",
          "cmd",
          "next(): wait on the channel",
          "The closure captures only the channel and the ID.",
          "app.previewSub.next"
        ],
        [
          "sub",
          "coord",
          "msg",
          "SessionOutputMsg",
          "",
          "app.Coordinator.handleSessionOutput"
        ],
        [
          "coord",
          "pv",
          "call",
          "accepts? → append, re-arm next()",
          "Output for a session we have left is ignored.",
          "app.previewSub.accepts"
        ],
        [
          "coord",
          "sub",
          "call",
          "leaving the dashboard: unsubscribe()",
          "Clears the ID with the channel, so queued output cannot re-arm the wait.",
          "app.previewSub.unsubscribe"
        ]
      ]
    },
    "mutation": {
      "title": "Refresh after a change, and errors",
      "summary": "How a command asks for fresh data, and how the error banner lives and dies.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "cmd",
          "cmd",
          "a command"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "ref",
          "app.refresher"
        ],
        [
          "views",
          "views.Dashboard",
          "every view"
        ]
      ],
      "steps": [
        [
          "cmd",
          "coord",
          "msg",
          "DataRefreshRequestedMsg",
          "A closure never calls refreshData() itself (rule 1); it asks.",
          "app.refreshRequestedCmd"
        ],
        [
          "tea",
          "ref",
          "cmd",
          "fullSeeded(previewID)",
          "Reads everything, plus the new fingerprint baseline so the next tick does not reload it again.",
          "app.refresher.fullSeeded"
        ],
        [
          "ref",
          "coord",
          "msg",
          "configFingerprintMsg{stamps, data}",
          "",
          "app.Coordinator.handleConfigFingerprint"
        ],
        [
          "coord",
          "views",
          "call",
          "handleDataLoaded → broadcastData",
          "",
          "app.Coordinator.handleDataLoaded"
        ],
        [
          "cmd",
          "coord",
          "msg",
          "ErrorMsg",
          "Sets the banner (c.err).",
          "app.Coordinator.handleSpecialMessages"
        ],
        [
          "user",
          "coord",
          "key",
          "the next keypress",
          "Clears the banner. Navigation alone never does, so an error batched with a navigation survives (rule 4).",
          "app.Coordinator.Update"
        ]
      ],
      "sources": {
        "refs": [
          "app.Coordinator.handleWorkspaceCommandResult"
        ],
        "files": [
          "internal/app/terminal_launch_test.go"
        ]
      }
    },
    "editcard": {
      "title": "Open the card editor",
      "summary": "The safe way to change screens: prepare the target view in the callback, navigate from a message.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "ed",
          "views.CardEditor"
        ],
        [
          "tea",
          "tea"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "e on a card",
          "",
          "views.Dashboard.handleEdit"
        ],
        [
          "dash",
          "coord",
          "call",
          "OnEditCard(card)",
          "Still inside the dashboard's Update.",
          "app.Coordinator.dashboardCallbacks"
        ],
        [
          "coord",
          "ed",
          "call",
          "SetCard(card)",
          "Load the main/sub workspace IDs, temporary boolean and instructions; reset selected-only Skills mode and keep unmanaged card fields.",
          "views.CardEditor.SetCard"
        ],
        [
          "coord",
          "tea",
          "return",
          "command → NavigateMsg",
          "Navigating here would store the dashboard under the editor's key (rule 2).",
          "app.Coordinator.dashboardCallbacks"
        ],
        [
          "tea",
          "coord",
          "msg",
          "NavigateMsg{CardEditor}",
          "",
          "app.Coordinator.navigateTo"
        ],
        [
          "coord",
          "ed",
          "call",
          "Init()",
          "",
          "views.CardEditor.Init"
        ]
      ],
      "sources": {
        "refs": [
          "views.CardEditor.SetCard",
          "views.CardEditor.Reset",
          "views.CardEditor.fieldBlocks",
          "views.editorKeyIdentity"
        ],
        "files": [
          "internal/ui/views/card_editor.go"
        ]
      }
    },
    "savecard": {
      "title": "Save an ID card",
      "summary": "Validate and write id_cards.yaml off the UI goroutine.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "ed",
          "views.CardEditor"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "repo",
          "config.FileRepository"
        ],
        [
          "disk",
          "disk"
        ],
        [
          "coord",
          "app.Coordinator"
        ]
      ],
      "steps": [
        [
          "user",
          "ed",
          "key",
          "Enter",
          "",
          "views.CardEditor.handleSubmit"
        ],
        [
          "ed",
          "ed",
          "self",
          "Card(): overlay the form on the loaded card",
          "Overlay the boolean temporary toggle and cloned sub workspace IDs; preserve unmanaged fields and selected capabilities.",
          "views.CardEditor.Card"
        ],
        [
          "ed",
          "tea",
          "return",
          "OnSave(card) command",
          "",
          "app.Coordinator.cardEditorCallbacks"
        ],
        [
          "tea",
          "repo",
          "cmd",
          "CreateIDCard / UpdateIDCard",
          "",
          "config.FileRepository.UpdateIDCard"
        ],
        [
          "repo",
          "repo",
          "self",
          "ValidateCard",
          "Read main/sub workspace references from one registry snapshot; reject unknown, duplicate and main-as-sub IDs. Skills, personas and MCP references must also exist.",
          "config.FileRepository.ValidateCard"
        ],
        [
          "repo",
          "disk",
          "io",
          "WriteAtomic(id_cards.yaml)",
          "Under the repository mutex.",
          "config.FileRepository.WriteAtomic"
        ],
        [
          "repo",
          "coord",
          "msg",
          "CardSavedMsg",
          "Error: banner, stay. Success: refresh and back to the dashboard.",
          "app.Coordinator.handleSpecialMessages"
        ]
      ],
      "sources": {
        "refs": [
          "config.IDCard",
          "views.CardEditor.Card",
          "config.FileRepository.ValidateCard"
        ],
        "files": [
          "internal/config/types.go",
          "internal/config/validation.go"
        ]
      }
    },
    "deletecard": {
      "title": "Delete an ID card",
      "summary": "A card a session still points at cannot be deleted.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "act",
          "app.sessionActions"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "repo",
          "config.FileRepository"
        ],
        [
          "coord",
          "app.Coordinator"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "Backspace on a card, confirm",
          "",
          "views.Dashboard.handleDeleteCard"
        ],
        [
          "dash",
          "tea",
          "return",
          "OnDeleteCard → deleteCard command",
          "",
          "app.sessionActions.deleteCard"
        ],
        [
          "tea",
          "act",
          "cmd",
          "deleteCard(cardID)",
          "",
          "app.sessionActions.deleteCard"
        ],
        [
          "act",
          "mgr",
          "call",
          "List(): who references it?",
          "config cannot see sessions, so the guard lives here.",
          "session.Manager.List"
        ],
        [
          "act",
          "repo",
          "call",
          "DeleteIDCard(id)",
          "Only if no session references the card.",
          "config.FileRepository.DeleteIDCard"
        ],
        [
          "act",
          "coord",
          "msg",
          "DataRefreshRequestedMsg or ErrorMsg",
          "",
          "app.Coordinator.handleSpecialMessages"
        ]
      ]
    },
    "source": {
      "title": "Add a source directory",
      "summary": "Read, change, write and re-sync a registry's sources under one lock.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "sm",
          "views.SourceManager"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "repo",
          "config.FileRepository"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "user",
          "sm",
          "key",
          "add a skills directory",
          "",
          "views.SourceManager.Update"
        ],
        [
          "sm",
          "coord",
          "call",
          "SaveSimpleSource(path) → saveSimpleSource",
          "Chooses UpdateSkillSources, UpdatePersonaSources or UpdatePuppetSources for this screen's type.",
          "app.Coordinator.saveSimpleSource"
        ],
        [
          "tea",
          "repo",
          "cmd",
          "UpdateSkillSources(mutate)",
          "Read, apply, write and re-sync under one lock. Never list, change and save in separate calls (rule 6).",
          "config.FileRepository.UpdateSkillSources"
        ],
        [
          "repo",
          "disk",
          "io",
          "skills_sources.yaml, rescan → skills.yaml",
          "",
          "config.FileRepository.syncSkillsFromSourcesLocked"
        ],
        [
          "repo",
          "coord",
          "msg",
          "refresh, or refresh + warning, or ErrorMsg",
          "A save that worked but whose re-sync failed is *config.SyncFailedError: refresh and warn.",
          "app.Coordinator.sourceUpdateMsg"
        ],
        [
          "coord",
          "sm",
          "call",
          "SetData → reload the list",
          "",
          "views.SourceManager.SetData"
        ]
      ]
    },
    "mcp": {
      "title": "Save an MCP server",
      "summary": "Parse the form, upsert into mcps.yaml, go back.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "form",
          "views.MCPForm"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "repo",
          "config.FileRepository"
        ],
        [
          "coord",
          "app.Coordinator"
        ]
      ],
      "steps": [
        [
          "user",
          "form",
          "key",
          "Enter",
          "",
          "views.MCPForm.handleSubmit"
        ],
        [
          "form",
          "form",
          "self",
          "parse args, env, headers",
          "Comma-separated with \\ escaping; a bad key=value list shows an error in the form.",
          "views.MCPForm.handleSubmit"
        ],
        [
          "form",
          "tea",
          "return",
          "OnSave(mcp) command",
          "",
          "app.Coordinator.mcpCallbacks"
        ],
        [
          "tea",
          "repo",
          "cmd",
          "UpsertMCP(mcp)",
          "",
          "config.FileRepository.UpsertMCP"
        ],
        [
          "repo",
          "coord",
          "msg",
          "MCPSavedMsg",
          "Refresh and go back to the previous view. The card editor auto-selects the mcp-cli skill when an MCP is added.",
          "app.Coordinator.handleSpecialMessages"
        ]
      ]
    },
    "settings": {
      "title": "Save settings",
      "summary": "Persist independent IDE and terminal commands plus card-editor visibility options, then refresh and return to the dashboard.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "set",
          "views.SettingsView"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "repo",
          "config.FileRepository"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "ed",
          "views.CardEditor"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "user",
          "set",
          "key",
          "Enter saves both commands",
          "Trim both buffers and preserve toggle values. Enter on Manage Puppet Sources opens that action instead.",
          "views.SettingsView.handleEditEnter"
        ],
        [
          "set",
          "tea",
          "return",
          "OnSave(settings) command",
          "Return a command; do not navigate from inside view.Update.",
          "app.Coordinator.settingsCallbacks"
        ],
        [
          "tea",
          "repo",
          "cmd",
          "SaveAppSettings(settings)",
          "The repository performs its existing locked atomic write.",
          "config.FileRepository.SaveAppSettings"
        ],
        [
          "repo",
          "disk",
          "io",
          "Write app_settings.yaml",
          "IDECommand and TerminalCommand are independent; terminal_command defaults empty in older files.",
          "config.AppSettings"
        ],
        [
          "repo",
          "coord",
          "msg",
          "SettingsSavedMsg",
          "A write failure becomes an error banner; success replaces c.settings on the UI goroutine.",
          "app.Coordinator.handleSpecialMessages"
        ],
        [
          "coord",
          "ed",
          "call",
          "SetSettings(new visibility options)",
          "Personas and MCP section preferences retain their existing behavior.",
          "views.CardEditor.SetSettings"
        ],
        [
          "coord",
          "coord",
          "self",
          "Refresh and navigate to dashboard",
          "Settings changes affect subsequent commands; commands already built retain their captured string.",
          "app.Coordinator.navigateTo"
        ]
      ],
      "sources": {
        "refs": [
          "config.FileRepository.GetAppSettings"
        ],
        "files": [
          "internal/config/app_settings_test.go",
          "internal/ui/views/settings_test.go"
        ]
      }
    },
    "workspace-sources": {
      "title": "Add workspace sources from the editor",
      "summary": "Ctrl+E on either workspace field visits the existing source manager and returns to the unsaved card.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "ed",
          "views.CardEditor"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "sm",
          "views.SourceManager"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "repo",
          "config.FileRepository"
        ]
      ],
      "steps": [
        [
          "user",
          "ed",
          "key",
          "Ctrl+E on Workspace / Sub Workspaces",
          "The shortcut is available on both fields.",
          "views.CardEditor.handleCtrlE"
        ],
        [
          "ed",
          "coord",
          "call",
          "OnOpenWorkspaces()",
          "Already wired callback; the editor now invokes it.",
          "app.Coordinator.cardEditorCallbacks"
        ],
        [
          "coord",
          "tea",
          "return",
          "NavigateMsg{SourceManagerWorkspaces}",
          "Navigation occurs later in a message handler.",
          "app.Coordinator.cardEditorCallbacks"
        ],
        [
          "tea",
          "coord",
          "msg",
          "navigateTo(workspace sources)",
          "Retain the editor instance and unsaved values.",
          "app.Coordinator.navigateTo"
        ],
        [
          "user",
          "sm",
          "key",
          "Add or edit a workspace source",
          "Reuses the existing source form.",
          "views.SourceManager.Update"
        ],
        [
          "tea",
          "repo",
          "cmd",
          "UpdateWorkspaceSources(mutate)",
          "Read, update, save and sync under one lock.",
          "config.FileRepository.UpdateWorkspaceSources"
        ],
        [
          "repo",
          "coord",
          "msg",
          "Refresh data or refresh + sync warning",
          "A successful save with failed sync still refreshes.",
          "app.Coordinator.sourceUpdateMsg"
        ],
        [
          "user",
          "sm",
          "key",
          "Esc returns to the editor",
          "NavigateBackMsg preserves the editor state.",
          "views.SourceManager.Update"
        ]
      ],
      "sources": {
        "refs": [
          "views.CardEditor.renderWorkspaceLine",
          "views.CardEditor.handleCtrlE"
        ],
        "files": []
      }
    },
    "skill-filter": {
      "title": "Show only selected skills",
      "summary": "Ctrl+S in Skills changes the visible subset; search and saved selections remain independent.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "ed",
          "views.CardEditor"
        ],
        [
          "list",
          "components.CapabilityList"
        ],
        [
          "coord",
          "app.Coordinator"
        ]
      ],
      "steps": [
        [
          "user",
          "ed",
          "key",
          "Ctrl+S while Skills is focused",
          "Intercept the shortcut before routing text to search.",
          "views.CardEditor.handleKey"
        ],
        [
          "ed",
          "list",
          "call",
          "SetSelectedOnly(!SelectedOnly())",
          "Changing focus or visiting sources retains the mode.",
          "components.CapabilityList.SetSelectedOnly"
        ],
        [
          "list",
          "list",
          "self",
          "Filter skills",
          "Before: search only. After: selected-only mode intersects checked IDs with search without replacing the full items collection.",
          "components.CapabilityList.FilteredItems"
        ],
        [
          "list",
          "list",
          "self",
          "renderHeader: selected indicator",
          "The indicator and shortcut are fitted to measured width.",
          "components.CapabilityList.renderHeader"
        ],
        [
          "user",
          "ed",
          "key",
          "Space toggles the visible skill",
          "Selection still uses workspace-independent skill IDs.",
          "views.CardEditor.handleSpace"
        ],
        [
          "ed",
          "list",
          "call",
          "ToggleSelected(), normalizeCursor()",
          "Deselecting the last result leaves a safe empty cursor.",
          "components.CapabilityList.ToggleSelected"
        ],
        [
          "coord",
          "ed",
          "call",
          "SetData after registry refresh",
          "Nil registry fields retain existing items.",
          "views.CardEditor.SetData"
        ],
        [
          "ed",
          "list",
          "call",
          "SetItems(), normalizeCursor()",
          "Clamp the cursor against the filtered subset.",
          "components.CapabilityList.SetItems"
        ],
        [
          "ed",
          "list",
          "call",
          "SelectedIDs() when saving",
          "Hidden-by-filter selections remain in the saved card.",
          "components.CapabilityList.SelectedIDs"
        ]
      ],
      "sources": {
        "refs": [
          "components.CapabilityList.Select",
          "components.CapabilityList.Deselect",
          "components.CapabilityList.Toggle",
          "components.CapabilityList.SetSelected",
          "components.CapabilityList.SelectedOnly",
          "components.CapabilityList.SetCursor",
          "components.CapabilityList.normalizeCursor"
        ],
        "files": [
          "internal/ui/views/card_editor_test.go"
        ]
      }
    },
    "temporary-toggle": {
      "title": "Toggle temporary workspace mode",
      "summary": "Space changes a boolean instead of requiring a typed y or n.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "ed",
          "views.CardEditor"
        ]
      ],
      "steps": [
        [
          "user",
          "ed",
          "key",
          "Space on Use Tmp Workspace",
          "Toggle the boolean directly.",
          "views.CardEditor.handleSpace"
        ],
        [
          "ed",
          "ed",
          "self",
          "renderTmpWorkspaceLine: [x] or [ ]",
          "Focus highlights the checkbox row.",
          "views.CardEditor.renderTmpWorkspaceLine"
        ],
        [
          "ed",
          "ed",
          "self",
          "Card(): assign boolean",
          "The persisted YAML field remains a boolean.",
          "views.CardEditor.Card"
        ]
      ]
    },
    "sub-picker": {
      "title": "Choose multiple sub workspaces",
      "summary": "The existing picker supports checkboxes, search, confirmation and cancellation; Ctrl+X on the focused form field clears all selections before reopening it.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "ed",
          "views.CardEditor"
        ],
        [
          "picker",
          "views.WorkspacePickerModal"
        ]
      ],
      "steps": [
        [
          "user",
          "ed",
          "key",
          "Space: sub list",
          "The summary field is directly below Workspace.",
          "views.CardEditor.handleSpace"
        ],
        [
          "ed",
          "picker",
          "call",
          "SetMultiple(true)",
          "Exclude the main workspace unless it is already checked and needs repair.",
          "views.WorkspacePickerModal.SetMultiple"
        ],
        [
          "ed",
          "picker",
          "call",
          "SetSelectedIDs()",
          "Replace provisional checks from the form, including an empty selection after Ctrl+X; rebuilding preserves the selected mode.",
          "views.WorkspacePickerModal.SetSelectedIDs"
        ],
        [
          "picker",
          "picker",
          "self",
          "Rebuild checks",
          "Single mode restores its selected cursor; multiple mode keeps all checks.",
          "views.WorkspacePickerModal.rebuildList"
        ],
        [
          "user",
          "picker",
          "key",
          "Search / toggle",
          "Search changes visible rows without discarding checked IDs.",
          "views.WorkspacePickerModal.Update"
        ],
        [
          "picker",
          "picker",
          "self",
          "toggleChecked()",
          "Stable provisional ordering includes newly checked IDs.",
          "views.WorkspacePickerModal.toggleChecked"
        ],
        [
          "user",
          "picker",
          "key",
          "Enter / Esc",
          "Cancelled edits never replace the editor selection.",
          "views.WorkspacePickerModal.Update"
        ],
        [
          "ed",
          "picker",
          "call",
          "SelectedIDs()",
          "Registered IDs first, then unresolved IDs in selection order.",
          "views.WorkspacePickerModal.SelectedIDs"
        ],
        [
          "ed",
          "ed",
          "self",
          "Commit sub IDs",
          "Switch the shared picker back to single mode.",
          "views.CardEditor.updateWorkspacePicker"
        ],
        [
          "ed",
          "ed",
          "self",
          "Render summary",
          "Show names or unresolved IDs. The focused row reserves width for ctrl+e=sources and ctrl+x=clear; an empty selection displays none.",
          "views.CardEditor.renderSubWorkspacesLine"
        ]
      ],
      "sources": {
        "refs": [
          "views.WorkspacePickerModal.SetExcludedID",
          "views.WorkspacePickerModal.SetSelectedID",
          "views.WorkspacePickerModal.restoreSingleSelection",
          "views.WorkspacePickerModal.title",
          "views.WorkspacePickerModal.View",
          "views.workspaceItem.Title"
        ],
        "files": [
          "internal/ui/views/workspace_picker_test.go",
          "internal/ui/views/card_editor_test.go",
          "internal/ui/views/card_editor_layout_test.go"
        ]
      }
    },
    "picker-refresh": {
      "title": "Keep checked workspaces through refresh",
      "summary": "Registry updates can remove a checked workspace while its picker is open.",
      "participants": [
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "ed",
          "views.CardEditor"
        ],
        [
          "picker",
          "views.WorkspacePickerModal"
        ],
        [
          "user",
          "user"
        ]
      ],
      "steps": [
        [
          "coord",
          "ed",
          "call",
          "SetData(DataLoadedMsg)",
          "Only a non-nil workspace snapshot replaces options.",
          "views.CardEditor.SetData"
        ],
        [
          "ed",
          "picker",
          "call",
          "SetWorkspaces(new registry)",
          "Remember checked workspace labels and paths before replacement.",
          "views.WorkspacePickerModal.SetWorkspaces"
        ],
        [
          "picker",
          "picker",
          "self",
          "rememberWorkspace for checked IDs",
          "Keep enough information to search a now-missing selected row.",
          "views.WorkspacePickerModal.rememberWorkspace"
        ],
        [
          "picker",
          "picker",
          "self",
          "filterWorkspaces includes missing checks",
          "Use all provisional IDs, including those checked during this opening.",
          "views.WorkspacePickerModal.filterWorkspaces"
        ],
        [
          "picker",
          "picker",
          "self",
          "refreshItems and normalizeSelection",
          "Checkbox states are rebuilt without losing the selected set.",
          "views.WorkspacePickerModal.refreshItems"
        ],
        [
          "user",
          "picker",
          "key",
          "Space explicitly deselects missing row",
          "Remove it from the checked map and ordered IDs.",
          "views.WorkspacePickerModal.toggleChecked"
        ],
        [
          "picker",
          "picker",
          "self",
          "removeID from provisional order",
          "Preserve order of the remaining IDs.",
          "views.removeID"
        ],
        [
          "user",
          "picker",
          "key",
          "Enter commits; Esc cancels",
          "An absent checked ID remains selected unless explicitly removed.",
          "views.WorkspacePickerModal.Update"
        ]
      ]
    },
    "sub-copy": {
      "title": "Copy or reuse sub workspaces",
      "summary": "One mutex protects source resolution, ownership decisions, copy publication and the manifest commit; private helpers own each stage.",
      "participants": [
        [
          "mgr",
          "session.Manager"
        ],
        [
          "repo",
          "config.FileRepository"
        ],
        [
          "disk",
          "files"
        ],
        [
          "atomic",
          "fsutil.WriteFileAtomic"
        ]
      ],
      "steps": [
        [
          "mgr",
          "mgr",
          "self",
          "Guard + lock",
          "An empty selection is a no-op; the manager mutex spans all helper calls.",
          "session.Manager.prepareSubWorkspaces"
        ],
        [
          "mgr",
          "mgr",
          "call",
          "canonicalDirectory()",
          "Resolve the main working directory once; source aliases are resolved separately.",
          "session.canonicalDirectory"
        ],
        [
          "mgr",
          "repo",
          "call",
          "ListWorkspaces()",
          "resolveSubWorkspaceSources reads the registry, resolves each selected ID and preserves registered basenames.",
          "config.FileRepository.ListWorkspaces"
        ],
        [
          "mgr",
          "mgr",
          "call",
          "Check containment",
          "Reject a source containing the staging/destination parent.",
          "session.ensureDestinationOutsideSource"
        ],
        [
          "mgr",
          "disk",
          "io",
          "loadSubWorkspaceManifest()",
          "Check metadata links, schema version and safe destination keys.",
          "session.loadSubWorkspaceManifest"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Prepare tmp directory",
          "Reject a symlink and ensure a real directory.",
          "session.prepareSubWorkspaceDirectory"
        ],
        [
          "mgr",
          "mgr",
          "call",
          "Clone manifest entries",
          "Retain ownership records for deselected copies.",
          "session.copySubWorkspaceManifest"
        ],
        [
          "mgr",
          "mgr",
          "call",
          "shouldCopySubWorkspace()",
          "Reuse only a matching managed directory; an occupied unrelated destination is an error.",
          "session.shouldCopySubWorkspace"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Copy into stage, publish",
          "Copy source contents into a real staging directory, recheck destination absence and rename.",
          "session.copyAndPublishSubWorkspace"
        ],
        [
          "mgr",
          "atomic",
          "call",
          "Persist ownership manifest",
          "Write updated entries atomically with mode 0600; preparation still holds the mutex.",
          "session.persistSubWorkspaceManifest"
        ]
      ],
      "sources": {
        "refs": [
          "session.safeSubWorkspaceName",
          "session.rejectSymlink",
          "session.resolveSubWorkspaceSources"
        ],
        "files": []
      }
    },
    "copy-rollback": {
      "title": "Roll back a failed copy batch",
      "summary": "The transaction object tracks only staging paths and new destinations owned by this preparation.",
      "participants": [
        [
          "mgr",
          "session.Manager"
        ],
        [
          "tx",
          "session.subWorkspaceTransaction"
        ],
        [
          "disk",
          "files"
        ]
      ],
      "steps": [
        [
          "mgr",
          "disk",
          "io",
          "Copy/publish missing source",
          "Successful publication moves its path from staged to published bookkeeping.",
          "session.copyAndPublishSubWorkspace"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Persist may fail",
          "A preparation error occurs before driver synchronization and spawn.",
          "session.persistSubWorkspaceManifest"
        ],
        [
          "mgr",
          "tx",
          "call",
          "rollback(cause)",
          "The orchestrator owns error routing and invokes transaction cleanup.",
          "session.subWorkspaceTransaction.rollback"
        ],
        [
          "tx",
          "disk",
          "io",
          "Remove tracked stages",
          "Earlier managed copies and unrelated siblings are outside this transaction.",
          "session.subWorkspaceTransaction.rollback"
        ],
        [
          "tx",
          "disk",
          "io",
          "Remove new destinations",
          "Reclaim only destinations published by this preparation.",
          "session.subWorkspaceTransaction.rollback"
        ],
        [
          "tx",
          "mgr",
          "return",
          "Join cause + cleanup errors",
          "Cleanup failures remain visible alongside the triggering error.",
          "session.subWorkspaceTransaction.rollback"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Clean a new temporary root",
          "startSession removes a failed new temporary workspace; retained and original workspaces survive.",
          "session.Manager.cleanupNewTemporaryWorkspace"
        ]
      ]
    },
    "failed-activation": {
      "title": "Clean a failed temporary activation",
      "summary": "A failed new activation has no record that could later reclaim its copy.",
      "participants": [
        [
          "act",
          "app.sessionActions"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "drv",
          "driver.GenericCLIDriver"
        ],
        [
          "disk",
          "files"
        ],
        [
          "coord",
          "app.Coordinator"
        ]
      ],
      "steps": [
        [
          "act",
          "mgr",
          "cmd",
          "startSession: prepare then sync and spawn",
          "Blocking filesystem and process work runs outside the UI loop.",
          "session.Manager.startSession"
        ],
        [
          "mgr",
          "drv",
          "call",
          "syncDriverFiles / Spawn may fail",
          "Error before activation; no live session was registered.",
          "session.Manager.syncDriverFiles"
        ],
        [
          "mgr",
          "mgr",
          "call",
          "cleanupNewTemporaryWorkspace(error)",
          "Only isNew && UseTmpWorkspace triggers root removal.",
          "session.Manager.cleanupNewTemporaryWorkspace"
        ],
        [
          "mgr",
          "disk",
          "io",
          "RemoveAll(session root); join errors",
          "Do not remove existing relaunch or agent-switch workspaces.",
          "session.Manager.cleanupNewTemporaryWorkspace"
        ],
        [
          "act",
          "coord",
          "msg",
          "Report launch error",
          "Existing launcher error handling displays the failure.",
          "app.Coordinator.handleSessionMessages"
        ]
      ]
    },
    "builtin-skill": {
      "title": "Deploy the bundled MCP-CLI skill",
      "summary": "Main embeds and injects the content once; a driver writes the selected built-in skill into a session workspace.",
      "participants": [
        [
          "mgr",
          "session.Manager"
        ],
        [
          "drv",
          "driver.GenericCLIDriver"
        ],
        [
          "content",
          "config.GetBuiltinSkillContent",
          "Config content"
        ],
        [
          "disk",
          "files"
        ]
      ],
      "steps": [
        [
          "mgr",
          "drv",
          "call",
          "SetSkills(selected skills)",
          "Driver synchronization happens after workspace preparation.",
          "driver.GenericCLIDriver.SetSkills"
        ],
        [
          "drv",
          "drv",
          "self",
          "deploySkill: built-in entry",
          "Metadata marks this skill Builtin; a normal source directory is not required.",
          "driver.deploySkill"
        ],
        [
          "drv",
          "drv",
          "self",
          "writeBuiltinSkill()",
          "Deployment uses the injected content lookup.",
          "driver.writeBuiltinSkill"
        ],
        [
          "drv",
          "content",
          "call",
          "GetBuiltinSkillContent(name)",
          "Return a clone under RLock; missing or unknown content returns false and is reported as an error.",
          "config.GetBuiltinSkillContent"
        ],
        [
          "drv",
          "disk",
          "io",
          "Write skills/mcp-cli/SKILL.md",
          "Use the driver’s agent folder: .agents or .claude. Existing deployment directories retain their existing content.",
          "driver.writeBuiltinSkill"
        ]
      ],
      "sources": {
        "refs": [
          "config.SetBuiltinMCPCLISkillContent",
          "root.main"
        ],
        "files": [
          "main.go",
          "main_test.go",
          "kankuro_skills/mcp-cli/SKILL.md"
        ]
      }
    },
    "sub-clear": {
      "title": "Clear all sub workspace selections",
      "summary": "Ctrl+X clears the unsaved selection only on the focused Sub Workspaces field; Enter saves and Esc cancels without deleting existing copies.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "ed",
          "views.CardEditor"
        ]
      ],
      "steps": [
        [
          "user",
          "ed",
          "key",
          "Ctrl+X in the form",
          "Sub Workspaces must be focused. Open modals receive keys first, so the shortcut does not clear selections behind a picker.",
          "views.CardEditor.Update"
        ],
        [
          "ed",
          "ed",
          "self",
          "Clear selected IDs",
          "The Ctrl+X switch case sets subWorkspaceIDs to nil, including unresolved IDs. Other fields retain their normal handling; focus stays put and no save callback runs.",
          "views.CardEditor.handleKey"
        ],
        [
          "ed",
          "ed",
          "self",
          "Show none + hint",
          "The empty field displays none. The focused shortcut hint stays visible by reserving its width before clipping the names.",
          "views.CardEditor.renderSubWorkspacesLine"
        ],
        [
          "user",
          "ed",
          "key",
          "Enter to save",
          "Use the normal save callback. Esc instead cancels the form through handleKey; the loaded card has not been mutated.",
          "views.CardEditor.handleSubmit"
        ],
        [
          "ed",
          "ed",
          "self",
          "Card(): empty sub IDs",
          "Clone the empty form selection into SubWorkspaceIDs, preserving other fields. The save-card flow persists this value; retained managed copies and manifests are not deleted.",
          "views.CardEditor.Card"
        ]
      ],
      "sources": {
        "refs": [
          "views.editorKeyIdentity",
          "views.CardEditor.SetCard",
          "views.CardEditor.handleSpace",
          "views.WorkspacePickerModal.SetSelectedIDs"
        ],
        "files": [
          "internal/ui/views/card_editor.go",
          "internal/ui/views/constants.go",
          "internal/ui/views/card_editor_test.go",
          "internal/ui/views/card_editor_layout_test.go",
          "kankuro_specs.md",
          "project.md"
        ]
      }
    },
    "terminal": {
      "title": "Open the selected session in a terminal",
      "summary": "Press t in the Sessions pane to launch an independently configured terminal in the actual session workspace.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "act",
          "app.sessionActions"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "launch",
          "session.LaunchWorkspaceCommand",
          "Workspace launcher"
        ],
        [
          "process",
          "process"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "t in Sessions",
          "Search consumes text first; both live and terminated sessions qualify.",
          "views.Dashboard.handleKey"
        ],
        [
          "dash",
          "dash",
          "self",
          "Check focus and selected session",
          "No command for card focus, empty selection or an absent callback.",
          "views.Dashboard.handleLaunchSession"
        ],
        [
          "dash",
          "coord",
          "call",
          "OnLaunchTerminal(selected ID)",
          "The callback is wired with the IDE callback.",
          "app.Coordinator.dashboardCallbacks"
        ],
        [
          "coord",
          "act",
          "call",
          "Snapshot TerminalCommand by value",
          "Only the coordinator goroutine reads c.settings.",
          "app.Coordinator.launchTerminal"
        ],
        [
          "coord",
          "act",
          "cmd",
          "Run terminal launch command",
          "Bubble Tea executes this closure off the UI goroutine.",
          "app.sessionActions.launchTerminal"
        ],
        [
          "act",
          "mgr",
          "call",
          "List current session snapshots",
          "Find the selected ID when the command executes, not the card registry path.",
          "session.Manager.List"
        ],
        [
          "act",
          "launch",
          "call",
          "Launch command in session.WorkingDir",
          "Temporary sessions use their copy; OriginalDir must not be substituted.",
          "app.sessionActions.launchWorkspaceCommand"
        ],
        [
          "launch",
          "process",
          "io",
          "Resolve path, replace $PATH, Start",
          "Parse quotes and execute directly. Set Dir; disconnect standard streams; reap asynchronously.",
          "session.LaunchWorkspaceCommand"
        ],
        [
          "act",
          "coord",
          "msg",
          "TerminalLaunchedMsg{Err}",
          "Reports process startup, not the external application’s eventual exit.",
          "app.Coordinator.handleSpecialMessages"
        ],
        [
          "coord",
          "coord",
          "self",
          "Apply completion result",
          "An error becomes the existing banner; success stays on the dashboard.",
          "app.Coordinator.handleWorkspaceCommandResult"
        ]
      ],
      "sources": {
        "refs": [
          "messages.TerminalLaunchedMsg",
          "views.DashboardCallbacks"
        ],
        "files": [
          "internal/app/terminal_launch_test.go",
          "internal/ui/views/dashboard_test.go",
          "internal/session/workspace_launcher.go"
        ]
      }
    },
    "workspace-command-errors": {
      "title": "Report a workspace-command launch error",
      "summary": "A missing session, workspace, terminal setting or executable returns a completion error; the IDE action uses the same banner handler.",
      "participants": [
        [
          "act",
          "app.sessionActions"
        ],
        [
          "launch",
          "session.LaunchWorkspaceCommand",
          "Workspace launcher"
        ],
        [
          "coord",
          "app.Coordinator"
        ],
        [
          "user",
          "user"
        ]
      ],
      "steps": [
        [
          "act",
          "act",
          "self",
          "Validate selected session and workspace",
          "A session removed before command execution becomes session-not-found.",
          "app.sessionActions.launchWorkspaceCommand"
        ],
        [
          "act",
          "act",
          "self",
          "Reject empty terminal setting",
          "Return Terminal command not configured; do not infer an emulator.",
          "app.sessionActions.launchTerminal"
        ],
        [
          "act",
          "launch",
          "call",
          "Resolve and start configured command",
          "Missing directories and executable/start errors are returned.",
          "session.LaunchWorkspaceCommand"
        ],
        [
          "act",
          "coord",
          "msg",
          "TerminalLaunchedMsg{Err}",
          "The event carries the startup error back into Update.",
          "messages.TerminalLaunchedMsg"
        ],
        [
          "coord",
          "coord",
          "self",
          "handleWorkspaceCommandResult(error)",
          "Stores the error in coordinator-owned state without navigation.",
          "app.Coordinator.handleWorkspaceCommandResult"
        ],
        [
          "user",
          "coord",
          "key",
          "Next key clears the banner",
          "Existing banner lifecycle is unchanged.",
          "app.Coordinator.Update"
        ]
      ],
      "sources": {
        "refs": [
          "messages.TerminalLaunchedMsg"
        ],
        "files": [
          "internal/app/terminal_launch_test.go",
          "internal/app/command_race_internal_test.go"
        ]
      }
    },
    "settings-edit": {
      "title": "Edit IDE and terminal commands",
      "summary": "Two independent text inputs retain edits while focus moves. Command cancellation restores both saved values; rendering fits the supported terminal size.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "set",
          "views.SettingsView"
        ],
        [
          "coord",
          "app.Coordinator"
        ]
      ],
      "steps": [
        [
          "set",
          "set",
          "self",
          "NewSettingsView()",
          "Create separate IDE and terminal input models; the terminal placeholder illustrates a separate window.",
          "views.NewSettingsView"
        ],
        [
          "coord",
          "set",
          "call",
          "SetSettings(snapshot)",
          "Populate both buffers from AppSettings.",
          "views.SettingsView.SetSettings"
        ],
        [
          "user",
          "set",
          "key",
          "Enter/e starts editing",
          "Start on IDE Command with both saved command values.",
          "views.SettingsView.startEdit"
        ],
        [
          "user",
          "set",
          "key",
          "Tab / Shift+Tab / arrows",
          "Focus order: IDE, terminal, Personas, MCPs, Puppet Sources.",
          "views.SettingsView.updateEdit"
        ],
        [
          "set",
          "set",
          "self",
          "syncFocus()",
          "Focus only one command input; toggles and actions blur both. Focus changes retain text.",
          "views.SettingsView.syncFocus"
        ],
        [
          "user",
          "set",
          "key",
          "Type or press Esc",
          "Only the focused input receives editing events. Esc restores both command buffers without invoking OnSave; existing toggle cancellation behavior is unchanged.",
          "views.SettingsView.updateEdit"
        ],
        [
          "coord",
          "set",
          "call",
          "SetSize(width, height minus 4)",
          "Both input widths account for box padding and borders.",
          "views.SettingsView.SetSize"
        ],
        [
          "set",
          "set",
          "self",
          "renderEdit()",
          "Compact bordered inputs keep controls/help in the allocated area at 80×24.",
          "views.SettingsView.renderEdit"
        ],
        [
          "set",
          "set",
          "self",
          "renderDisplay()",
          "Sanitize and truncate saved command text with display-width helpers; stored commands remain intact.",
          "views.SettingsView.renderDisplay"
        ]
      ],
      "sources": {
        "refs": [
          "views.SettingsView.TerminalInput",
          "components.DisplayText",
          "components.TruncateHead"
        ],
        "files": [
          "internal/ui/views/settings_test.go"
        ]
      }
    }
  },
  "rules": [
    {
      "n": "1",
      "title": "No tea.Cmd closure touches coordinator state",
      "ref": "app.Coordinator",
      "why": "A command's closure runs on its own goroutine. Reading or writing c.views, c.preview, c.currentView, c.err or c.settings there is a concurrent access that can crash the app. Snapshot what you need before returning the closure.",
      "check": "grep -n \"c.refreshData()\" internal/app/callbacks.go must be empty; TestCommandClosuresDoNotRaceCoordinator under -race (a backstop, not a proof). Terminal and IDE launch wrappers snapshot their command strings before building tea.Cmd closures.",
      "flows": [
        "tick",
        "launch",
        "terminal",
        "ide"
      ],
      "sources": {
        "refs": [
          "app.Coordinator.launchIDE",
          "app.Coordinator.launchTerminal",
          "app.sessionActions.launchWorkspaceCommand"
        ],
        "files": [
          "internal/app/command_race_internal_test.go"
        ]
      }
    },
    {
      "n": "2",
      "title": "Never navigate from inside a view's Update",
      "ref": "app.Coordinator.navigateTo",
      "why": "Update writes the view back into c.views[c.currentView] after the view returns. Navigating during that call stores the old view under the new key; it once replaced the dashboard with the agent-switch view.",
      "check": "Navigate from handleSpecialMessages / handleSessionMessages. See navigateTo's doc comment.",
      "flows": [
        "editcard",
        "launch"
      ]
    },
    {
      "n": "3",
      "title": "DataLoadedMsg is always a full snapshot",
      "ref": "messages.DataLoadedMsg",
      "why": "A nil field means \"that registry could not be read this round, keep what you have\". Wiping a list makes CapabilityList.SelectedIDs() drop items, and the next save silently loses the card's skills.",
      "check": "Every SetData guards each field with if msg.X != nil; partial updates use PreviewDataMsg or SessionsRefreshedMsg.",
      "flows": [
        "tick",
        "preview"
      ]
    },
    {
      "n": "4",
      "title": "The error banner clears on the next keypress",
      "ref": "app.Coordinator.Update",
      "why": "Navigation leaves it alone, so an ErrorMsg batched beside a NavigateMsg survives in either order.",
      "check": "TestErrorSurvivesNavigationInEitherOrder, TestKeyPressClearsBanner.",
      "flows": [
        "mutation",
        "workspace-command-errors"
      ],
      "sources": {
        "refs": [
          "app.Coordinator.handleWorkspaceCommandResult"
        ],
        "files": [
          "internal/app/error_banner_test.go",
          "internal/app/terminal_launch_test.go"
        ]
      }
    },
    {
      "n": "5",
      "title": "The card editor overlays, it does not rebuild",
      "ref": "views.CardEditor.Card",
      "why": "Card() overlays only edited fields onto the loaded card. New boolean/sub workspace controls must preserve environment, system instructions, timestamps and hidden sections.",
      "check": "TestCardEditorPreservesUnmanagedFields and TestCtrlXClearsOnlySubWorkspaceSelection.",
      "flows": [
        "savecard",
        "sub-clear"
      ]
    },
    {
      "n": "6",
      "title": "Config read-modify-write goes through Update*Sources(fn)",
      "ref": "config.FileRepository.UpdateSkillSources",
      "why": "Listing, mutating and saving in separate calls means a failed read persists a truncated list. A save whose re-sync failed returns *config.SyncFailedError: refresh and warn.",
      "check": "Use the Update*Sources functions; never compose List + Save yourself.",
      "flows": [
        "source"
      ]
    },
    {
      "n": "7",
      "title": "Exported repository methods lock; *Locked helpers assume the lock",
      "ref": "config.FileRepository",
      "why": "Go mutexes are not reentrant. A locked method calling another locked method deadlocks the whole TUI.",
      "check": "Internal calls use the *Locked variants.",
      "flows": []
    },
    {
      "n": "8",
      "title": "\"Started, but the record didn't save\" is not a failure",
      "ref": "session.Manager.Launch",
      "why": "Launch, Relaunch and SwitchAgent wrap a post-spawn persist error in ErrPersistFailed. The agent is running; the UI still offers attach and shows a warning.",
      "check": "Callers use errors.Is(err, session.ErrPersistFailed).",
      "flows": [
        "launch",
        "relaunch"
      ]
    },
    {
      "n": "9",
      "title": "Rendered output keeps its shape",
      "ref": "components.TruncateHead",
      "why": "Fit content by measured width, never bytes, and never let lipgloss word-wrap a row that must stay on one line: one extra column makes an outer border wrap the whole row.",
      "check": "Layout sweeps in *_layout_test.go with uitest.AssertExactRect / AssertWithin.",
      "flows": []
    },
    {
      "n": "10",
      "title": "A sized component keeps its elements in place",
      "ref": "components.CapabilityList.viewSized",
      "why": "Where an element sits depends on the size the component was given, never on what it shows (page, scroll, search). Pad lists to their width, join left-aligned, reserve rows for optional elements.",
      "check": "TestWorkspacePickerKeepsColumnsAcrossPages, TestCapabilityListItemsKeepTheirRows, TestCardEditorListsKeepTheirPlace.",
      "flows": []
    },
    {
      "n": "11",
      "title": "Keep sub workspace ownership and rollback under one preparation mutex",
      "ref": "session.Manager.prepareSubWorkspaces",
      "why": "Two activations can share an original workspace. Splitting helpers must preserve serialization and rollback only paths created by the current preparation; existing edits and unrelated files must survive.",
      "check": "Review the mutex scope and transaction bookkeeping; run workspace/collision/failure tests and go test -race ./....",
      "flows": [
        "sub-copy",
        "copy-rollback",
        "failed-activation"
      ],
      "sources": {
        "refs": [
          "session.subWorkspaceTransaction.rollback",
          "session.shouldCopySubWorkspace"
        ],
        "files": [
          "internal/session/sub_workspaces.go"
        ]
      }
    },
    {
      "n": "12",
      "title": "Main supplies built-in bytes; config owns a protected copy",
      "ref": "config.SetBuiltinMCPCLISkillContent",
      "why": "Config cannot import the root executable package. Startup injection breaks that dependency; cloning input/output prevents callers from modifying shared skill data. Tests that replace process-wide content must restore it and run sequentially.",
      "check": "TestBuiltinSkillContentOwnershipAndMissingValues, TestBuiltinSkillContentConcurrentReads and TestEmbeddedMCPCLISkillDeploysThroughConfig.",
      "flows": [
        "startup",
        "builtin-skill"
      ],
      "sources": {
        "refs": [
          "config.GetBuiltinSkillContent",
          "root.main"
        ],
        "files": [
          "internal/config/builtins_test.go",
          "main_test.go"
        ]
      }
    }
  ],
  "tasks": [
    {
      "q": "Add a new agent CLI",
      "a": "Embed GenericCLIDriver and override only what differs (agent folder, instruction file, model flag); register it in driver/registry.go.",
      "refs": [
        "driver.GenericCLIDriver",
        "driver.ClaudeCodeDriver",
        "driver.NewBuiltinRegistry"
      ]
    },
    {
      "q": "Change what gets written into a session's workspace",
      "a": "Driver methods own instructions, skills, personas and MCP files. Manager.prepareSubWorkspaces owns additional copied workspaces and their manifest; keep these responsibilities separate.",
      "refs": [
        "driver.GenericCLIDriver.SetAgentsMD",
        "session.Manager.syncDriverFiles",
        "session.Manager.prepareSubWorkspaces"
      ]
    },
    {
      "q": "Change session lifecycle or status detection",
      "a": "Start in Manager and lifecycle/watch/output helpers. Activation now includes sub workspace preparation before driver synchronization. Preserve reservations, generation checks and cleanup of failed new temporary roots.",
      "refs": [
        "session.Manager",
        "session.Manager.watch",
        "driver.IsPromptLike"
      ],
      "flow": "watch"
    },
    {
      "q": "Change attach / detach / keys during attach",
      "a": "app/attach.go.",
      "refs": [
        "app.attachExec"
      ],
      "flow": "attach"
    },
    {
      "q": "Add a field to ID cards",
      "a": "Update config.IDCard and validation, then SetCard, Reset, Card and the focus table in the editor. Preserve unmanaged fields and add coverage for loading/saving the new field.",
      "refs": [
        "config.IDCard",
        "views.CardEditor.Card"
      ]
    },
    {
      "q": "Add a config registry or source type",
      "a": "repository.go interfaces and filename, a new config/*.go, Update*Sources, and Fingerprint's file list.",
      "refs": [
        "config.Repository",
        "config.registryFiles"
      ]
    },
    {
      "q": "Add a screen",
      "a": "A new type in ui/views; register it in initViews, add a ViewType, wire its callbacks in callbacks.go.",
      "refs": [
        "app.Coordinator.initViews",
        "app.ViewType"
      ],
      "flow": "editcard"
    },
    {
      "q": "Add a message",
      "a": "Declare it in messages.go, handle it in handleSpecialMessages or handleSessionMessages. Partial data gets its own type (rule 3). TerminalLaunchedMsg is an example of a completion result carrying an error.",
      "refs": [
        "app.Coordinator.handleSpecialMessages",
        "app.Coordinator.handleSessionMessages",
        "messages.TerminalLaunchedMsg"
      ]
    },
    {
      "q": "Change how cards look",
      "a": "RenderCardItem, the puppet renderer and the palette. Rule 9 and the card sweep test.",
      "refs": [
        "components.RenderCardItem"
      ]
    },
    {
      "q": "Add or change a puppet",
      "a": "Built-in art in ascii/characters; user puppets are .puppet files found in puppet source dirs.",
      "refs": [
        "characters.TextPuppet",
        "config.FileRepository.SyncPuppetsFromSources"
      ]
    },
    {
      "q": "Change dashboard layout or keys",
      "a": "dashboard.go (footerText for the footer), then the key tables in project.md §4 and kankuro_specs.md. Preserve the compact Sessions legend at the 80-column minimum when adding shortcuts.",
      "refs": [
        "views.Dashboard.handleKey",
        "views.Dashboard.footerText"
      ],
      "sources": {
        "files": [
          "internal/ui/views/dashboard_layout_test.go"
        ]
      }
    },
    {
      "q": "Put user text into a fixed-width slot",
      "a": "DisplayText first, then TruncateHead (keep the start) or TruncateTail (keep the end: paths, text being typed).",
      "refs": [
        "components.DisplayText",
        "components.TruncateHead",
        "components.TruncateTail"
      ]
    },
    {
      "q": "Change the minimum terminal size",
      "a": "views/constants.go and app/size_guard.go.",
      "refs": [
        "app.Coordinator.tooSmall"
      ],
      "flow": "resize"
    },
    {
      "q": "Change refresh behaviour",
      "a": "refresh.go for the commands, handlers.go for applying results, and config.Fingerprint.",
      "refs": [
        "app.refresher",
        "config.FileRepository.Fingerprint"
      ],
      "flow": "tick"
    },
    {
      "q": "Add a session-level action (Phase 1 linking)",
      "a": "A method on sessionActions returning a tea.Cmd that takes coordinator state by value; a message and a handler in coordinator.go; the trigger in callbacks.go. The terminal action demonstrates by-value settings and shared completion-error handling.",
      "refs": [
        "app.sessionActions",
        "app.Coordinator.launchTerminal",
        "app.Coordinator.handleWorkspaceCommandResult"
      ],
      "flow": "terminal"
    },
    {
      "q": "Change the preview subscription",
      "a": "previewSub and the preview handlers.",
      "refs": [
        "app.previewSub",
        "app.Coordinator.handleSessionOutput"
      ],
      "flow": "preview"
    },
    {
      "q": "Change how IDE or terminal windows open",
      "a": "Settings provides independent commands. Coordinator wrappers capture strings; sessionActions chooses the selected session’s WorkingDir, and LaunchWorkspaceCommand handles path replacement, parsing, Dir, spawn and reaping. Keep IDELauncher compatibility and temporary-workspace regression coverage.",
      "refs": [
        "config.AppSettings",
        "app.Coordinator.launchTerminal",
        "app.sessionActions.launchWorkspaceCommand",
        "session.LaunchWorkspaceCommand",
        "session.IDELauncher"
      ],
      "flow": "terminal",
      "sources": {
        "files": [
          "internal/app/terminal_launch_test.go",
          "internal/session/ide_launcher_test.go"
        ]
      }
    },
    {
      "q": "Change sub workspace copying or collision handling",
      "a": "Read sub_workspaces.go: the orchestrator holds one mutex while helpers resolve sources, load ownership metadata, decide reuse, stage/publish and persist. Rollback only the current transaction’s paths.",
      "refs": [
        "session.Manager.prepareSubWorkspaces",
        "session.shouldCopySubWorkspace",
        "session.subWorkspaceTransaction.rollback"
      ],
      "flow": "sub-copy"
    },
    {
      "q": "Change workspace picking or the selected Skills view",
      "a": "The editor owns unsaved selections; Ctrl+X clears sub workspace IDs only on the focused form field. WorkspacePickerModal owns provisional checks and stable ID order; opening it reseeds all checks, including an empty selection. CapabilityList filters its full collection. Test search, refresh, cancel, clearing, modal scope and cursor behavior.",
      "refs": [
        "views.WorkspacePickerModal",
        "views.CardEditor.handleKey",
        "components.CapabilityList"
      ],
      "flow": "sub-picker",
      "sources": {
        "refs": [
          "views.CardEditor.renderSubWorkspacesLine",
          "views.WorkspacePickerModal.SetSelectedIDs"
        ],
        "files": [
          "internal/ui/views/card_editor_test.go",
          "internal/ui/views/card_editor_layout_test.go"
        ]
      }
    },
    {
      "q": "Change the embedded skill or startup installation",
      "a": "Edit the canonical kankuro_skills/mcp-cli/SKILL.md asset. Root main.go embeds and injects it; config owns metadata/content lookup. Root tests exercise deployment, and go install . installs the root command.",
      "refs": [
        "root.main",
        "config.SetBuiltinMCPCLISkillContent",
        "config.GetBuiltinSkillContent"
      ],
      "flow": "builtin-skill"
    }
  ],
  "storage": {
    "root": "~/.config/kankuro/",
    "tree": [
      [
        "id_cards.yaml",
        ""
      ],
      [
        "skills.yaml",
        "registry"
      ],
      [
        "skills_sources.yaml",
        "dirs it is scanned from"
      ],
      [
        "personas.yaml",
        ""
      ],
      [
        "persona_sources.yaml",
        ""
      ],
      [
        "puppets.yaml",
        ""
      ],
      [
        "puppet_sources.yaml",
        ""
      ],
      [
        "workspaces.yaml",
        ""
      ],
      [
        "workspace_sources.yaml",
        ""
      ],
      [
        "mcps.yaml",
        "entered by hand"
      ],
      [
        "app_settings.yaml",
        "Independent IDE/terminal commands and editor sections"
      ],
      [
        "sessions/<id>.json",
        "one record per session"
      ],
      [
        "tmp/<id>/",
        "temp workspace copies"
      ]
    ],
    "facts": [
      "Override the root with KANKURO_CONFIG_DIR or --config-dir; temp copies with KANKURO_TMP_DIR.",
      "All YAML and session JSON go through fsutil.WriteFileAtomic, under the repository or manager mutex.",
      "Loading a session always marks it terminated.",
      "While Bubble Tea owns the terminal, app/session code must not print to stdout/stderr. Main may report flags, startup errors and shutdown after terminal release; --debug logging goes to the user cache.",
      "When running the app by hand, point KANKURO_CONFIG_DIR and KANKURO_TMP_DIR at a throwaway directory. Reading the real config is fine; writing to it is not.",
      "Sub workspace copies live at <session working directory>/tmp/<registered source basename>, separate from the configuration directory’s tmp/<session-id>/ main-workspace copies.",
      "Ownership is recorded at <session working directory>/.kankuro/sub-workspaces.json. Matching managed copies retain edits on relaunch/switch/restart; deselection retains folders and metadata.",
      "With temporary mode off, sub copies and metadata are written to the original workspace. Destination conflicts abort activation; separate running application instances are not serialized by one manager mutex.",
      "o and t work only with Sessions focused and a selected session. Both use session.WorkingDir, including temporary copies; live and terminated sessions qualify.",
      "TerminalCommand (terminal_command) defaults empty: users choose an emulator command that opens a separate window. Both commands substitute $PATH; quote \"$PATH\" for spaces. Commands execute directly with the workspace as Dir and no implicit shell expansion.",
      "External application completion means process startup, not eventual exit. No command is inferred; launch failures use the normal error banner without navigation."
    ]
  },
  "gates": [
    "go build ./...",
    "go vet ./...",
    "go test ./...",
    "go test -race ./...",
    "golangci-lint run          # must exit 0",
    "",
    "# run the app isolated from your real config",
    "KANKURO_TMP_DIR=/tmp/kankuro-test/tmp \\",
    "KANKURO_CONFIG_DIR=/tmp/kankuro-test/config \\",
    "  go run . --debug",
    "",
    "# local installation from the root",
    "go install .",
    "# public installation after publishing the root-command change",
    "go install github.com/BlackGiack/kankuro@latest"
  ],
  "testfacts": [
    "Test the coordinator by driving it: NewCoordinator, Update(msg), then inspect View(), CurrentView(), Views(). Go through Update, not a callback, or you skip the write-back where rule 2 breaks.",
    "Make sure each new test can fail: remove the fix, watch it fail, put it back.",
    "Session-launching tests use newTestManager; testenv.Run fails the package if anything is left in its private temp root (about 3 s per package).",
    "Shape tests check that a render fits its box; position tests check that elements stay put as the content changes.",
    "Built-in content tests inject explicit fixtures and restore the previous value with t.Cleanup; do not run content-replacing tests in parallel. The root integration test deploys the real embedded bytes through config and the generic driver.",
    "Terminal launch integration records a helper process’s real working directory and arguments. Substituting OriginalDir for WorkingDir must make TestTerminalLaunchThroughDashboard fail; shared-launcher tests cover repeated quoted placeholders and commands without one.",
    "Terminal action routing covers running, waiting and terminated selections, search, empty selections and nil callbacks. Settings tests cover independent buffers, save/cancel, display sanitization and bounded layout; the coordinator race harness includes terminal launch."
  ],
  "glossary": [
    [
      "Coordinator goroutine",
      "The goroutine Bubble Tea runs Update and View on. Coordinator state may only be touched there."
    ],
    [
      "Command (tea.Cmd)",
      "A function Bubble Tea runs on its own goroutine; its return value comes back as a message."
    ],
    [
      "generation",
      "Per-session counter bumped on every activation, so goroutines of an old run stop touching the session."
    ],
    [
      "activating",
      "Reservation set by Relaunch and SwitchAgent between the guard and activation."
    ],
    [
      "Fingerprint",
      "mtime + size of each registry file; the tick re-reads config only when one changed."
    ],
    [
      "Seeded refresh",
      "A full refresh that also stores the new fingerprint baseline."
    ],
    [
      "waiting",
      "Status for an agent whose last output looks like a prompt (driver.IsPromptLike)."
    ],
    [
      "ModalState",
      "Embeddable open / confirmed / cancelled bookkeeping shared by every modal."
    ],
    [
      "CoordinatorRows",
      "Rows the coordinator keeps back before sizing a view (4); one definition in views/constants.go."
    ],
    [
      "Driver files",
      "What a driver writes into a workspace: AGENTS.md or CLAUDE.md, skills, personas, .kankuro/mcp-servers.json."
    ],
    [
      "mcp-cli",
      "The one built-in skill, embedded in the binary; it teaches agents to use .kankuro/mcp-servers.json."
    ],
    [
      "Shape test / position test",
      "Shape: the render fits its box. Position: elements stay put as content changes."
    ],
    [
      "Sub workspace",
      "An optional workspace registry entry copied under the active session workspace’s tmp directory."
    ],
    [
      "Sub workspace manifest",
      "Versioned workspace-local metadata mapping destination names to workspace IDs and canonical source paths."
    ],
    [
      "Copy transaction",
      "Bookkeeping for staging and newly published directories; rollback removes only those paths."
    ],
    [
      "Content injection",
      "Root main supplies embedded MCP-CLI bytes to config before drivers can deploy the skill."
    ],
    [
      "Workspace command",
      "IDECommand or TerminalCommand executed directly in the session’s WorkingDir, with $PATH substituted before quote-aware parsing."
    ]
  ]
};
