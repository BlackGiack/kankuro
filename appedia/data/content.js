window.APPEDIA_CONTENT = {
  "app": "KANKURO",
  "title": "Kankuro Appedia",
  "tagline": "A human dashboard and local MCP control plane for agent sessions",
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
    ],
    "orchestrator": [
      "Orchestrator",
      "an ordinary CLI session"
    ],
    "agent": [
      "Recipient agent",
      "reports through MCP"
    ],
    "trust": [
      "CLI user config",
      "Codex TOML / Claude JSON"
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
  "snapshot": "Working tree on feat/claude-resume at 5c6c0f7, including staged removal of legacy registration-block cleanup, 2026-10-29.",
  "pitch": "Kankuro runs coding-agent CLIs in separate PTYs and lets humans and authenticated orchestrator sessions operate the same sessions. Cards select projects, models, skills, personas and MCP servers; orchestrators can create named sessions, switch cards, and exchange explicit reports.",
  "pitch2": "Bubble Tea owns human interaction; a loopback MCP endpoint adapts tools to a shared application service. One session manager owns processes, activation identity, native conversation bindings, startup timers and message history. Factory workflow policy stays in user instructions.",
  "concepts": [
    [
      "ID card",
      "Saved project-scoped agent template and optional Dashboard Control permission.",
      "config.IDCard"
    ],
    [
      "Session",
      "Persistent workspace identity and user name across successive CLI activations.",
      "session.Session"
    ],
    [
      "Activation",
      "One process generation, credential and startup gate.",
      "session.Snapshot"
    ],
    [
      "Shared service",
      "Application operations used by human UI and local MCP.",
      "api.Service"
    ],
    [
      "Conversation binding",
      "Saved native conversation for one card in one session.",
      "session.ConversationBinding"
    ],
    [
      "Message and reply",
      "Marked process input plus explicit correlated MCP reports.",
      "session.MessageRecord"
    ]
  ],
  "path": [
    [
      "See the shared service between UI, local MCP and process manager.",
      "map"
    ],
    [
      "Trace a fresh session from preparation through publication.",
      "flow-launch"
    ],
    [
      "Separate immediate human attachment from the message startup gate.",
      "flow-readiness"
    ],
    [
      "Follow registration and native conversation resume.",
      "flow-registration"
    ],
    [
      "Trace explicit messages and progress/completion reports.",
      "flow-send"
    ],
    [
      "Read invariants and entry points for future changes.",
      "rules"
    ]
  ],
  "map": {
    "extraEdges": [
      [
        "mcp",
        "api",
        "MCP tools"
      ],
      [
        "app",
        "api",
        "TUI commands"
      ],
      [
        "session",
        "pty",
        "spawn / input"
      ],
      [
        "driver",
        "disk",
        "MCP config / trust"
      ]
    ],
    "flowEdges": [
      [
        "root",
        "mcp",
        "start endpoint"
      ],
      [
        "mcp",
        "api",
        "authenticated operations"
      ],
      [
        "app",
        "api",
        "human actions"
      ],
      [
        "api",
        "session",
        "shared lifecycle"
      ],
      [
        "session",
        "driver",
        "spawn / submit"
      ],
      [
        "driver",
        "pty",
        "PTY input"
      ],
      [
        "session",
        "disk",
        "records / reports"
      ]
    ],
    "viewBox": [
      0,
      0,
      1200,
      880
    ],
    "nodes": {
      "root": [
        580,
        80
      ],
      "testenv": [
        1040,
        80
      ],
      "app": [
        350,
        220
      ],
      "mcp": [
        810,
        220
      ],
      "api": [
        810,
        360
      ],
      "views": [
        120,
        360
      ],
      "messages": [
        580,
        500
      ],
      "session": [
        810,
        500
      ],
      "components": [
        120,
        500
      ],
      "driver": [
        1040,
        640
      ],
      "config": [
        580,
        640
      ],
      "fsutil": [
        810,
        780
      ],
      "ascii": [
        120,
        780
      ],
      "characters": [
        350,
        780
      ],
      "component": [
        120,
        640
      ],
      "styles": [
        350,
        640
      ],
      "uitest": [
        1040,
        780
      ],
      "tea": [
        120,
        220
      ],
      "pty": [
        1040,
        500
      ],
      "disk": [
        580,
        780
      ]
    }
  },
  "packages": {
    "app": {
      "label": "internal/app",
      "sub": "the coordinator",
      "role": "Coordinates Bubble Tea views and captures immutable command inputs. Session actions delegate domain operations to api.Service; conversation-choice callbacks open an explicit new/resume picker.",
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
        "Bubble Tea command closures run off the UI goroutine. Native-conversation choices are validated again by the session manager."
      ],
      "rules": [],
      "key": [
        "Coordinator",
        "refresher",
        "sessionActions",
        "previewSub",
        "attachExec"
      ],
      "sources": {
        "refs": [
          "app.sessionActions.launch",
          "app.sessionActions.relaunchWithChoice",
          "app.sessionActions.switchAgentWithChoice"
        ]
      }
    },
    "views": {
      "label": "ui/views",
      "sub": "full screens",
      "role": "Renders dashboard, card forms, switch picker and the new native-conversation picker. Humans explicitly choose a saved conversation or a new one when bindings exist.",
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
        "Dashboard Control is a per-card opt-in; the selected MCP list determines whether it can take effect.",
        "Existing session workspace identity survives card switches."
      ],
      "rules": [],
      "key": [
        "Dashboard",
        "CardEditor",
        "ConversationChoiceView",
        "AgentSwitchView"
      ]
    },
    "messages": {
      "label": "ui/messages",
      "sub": "every tea.Msg",
      "role": "Defines typed messages crossing the Bubble Tea loop, including conversation-choice requests/results and MCP endpoint failures.",
      "files": {
        "messages.go": "all message types, DataReceiver, PreviewReceiver"
      },
      "notes": [
        "DataLoadedMsg is always a full snapshot; partial updates get their own type."
      ],
      "rules": [],
      "key": [
        "DataLoadedMsg",
        "PreviewDataMsg",
        "SessionsRefreshedMsg",
        "DataReceiver"
      ]
    },
    "session": {
      "label": "internal/session",
      "sub": "activations and messaging",
      "role": "Owns PTYs, activation generations, saved workspace identity, conversation bindings, startup timers and explicit message history. It is shared by both frontends and enforces lifecycle reservations.",
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
        "workspace_launcher.go": "Shared path resolution, placeholder replacement, direct process spawn and asynchronous reaping",
        "conversation.go": "activation credentials and native conversation reports",
        "messaging.go": "intent, reports, retention and delivery admission",
        "card_references.go": "pending card reference and deletion guard",
        "startup.go": "startup policy and unpublished-process teardown"
      },
      "notes": [
        "A Kankuro session persists across card changes; each process activation has a fresh credential and generation.",
        "Human attachment is immediate after spawn; only MCP message admission waits ten seconds.",
        "Temporary-workspace deletion removes exact Codex/Claude trust grants before deleting the record; cleanup failure retains the session for retry."
      ],
      "rules": [
        "8",
        "11",
        "13",
        "16",
        "17",
        "18"
      ],
      "key": [
        "Manager",
        "Session",
        "Snapshot",
        "MessageStore",
        "MessageAdmission"
      ],
      "sources": {
        "refs": [
          "session.Manager.startSession",
          "session.Manager.SwitchAgentWithChoiceContext",
          "session.Manager.deliveryAvailableLocked",
          "session.MessageStore"
        ]
      }
    },
    "components": {
      "label": "ui/components",
      "sub": "reusable pieces",
      "role": "Renders reusable TUI lists, previews, art and modals. Session phase/delivery hints and responsive preview metadata make activation state visible; narrow previews omit art when necessary.",
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
      "rules": [],
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
      "sub": "CLI adapters and trust",
      "role": "Constructs a fresh driver per activation and owns process input serialization. Codex and Claude specialize resume/trust behavior; generic drivers support arbitrary executable commands.",
      "files": {
        "driver.go": "GenericCLIDriver, ClaudeCodeDriver, Spawn, Kill, file sync, IsPromptLike",
        "registry.go": "name → constructor",
        "factory.go": "construction helpers",
        "trust.go": "exact workspace trust preparation and removal",
        "mcp_config.go": "atomic selected-server config and credential-matched cleanup"
      },
      "notes": [
        "Codex sends text, waits 500 ms, then writes Enter. Submitted bytes do not prove processing.",
        "Trust edits preserve configuration structurally, with atomic writes and a process-local update lock."
      ],
      "rules": [
        "12",
        "13",
        "16"
      ],
      "key": [
        "AgentDriver",
        "GenericCLIDriver",
        "CodexCLIDriver",
        "ClaudeCodeDriver"
      ],
      "sources": {
        "refs": [
          "driver.CodexCLIDriver.SetSession",
          "driver.ClaudeCodeDriver.SetSession",
          "driver.GenericCLIDriver.SubmitPrompt",
          "driver.RemoveWorkspaceTrust"
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
      "role": "Persists cards, registries and application settings as YAML. Adds per-card Dashboard Control permission and built-in Kankuro MCP / puppet-skill metadata; activation credentials stay runtime-only.",
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
        "Cards select Kankuro MCP and opt into Dashboard Control separately. The permission is captured at activation."
      ],
      "rules": [],
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
      "role": "Embeds both built-in skills, opens the repository, creates one session manager and operation runner, starts authenticated loopback MCP, then runs Bubble Tea. Owns ordered shutdown and endpoint-failure reporting.",
      "files": {
        "main.go": "embed, config injection, flags, wiring, signals, shutdown"
      },
      "notes": [
        "The endpoint starts before agents can launch. On exit the operation runner stops, sessions terminate, then the MCP endpoint closes."
      ],
      "rules": [],
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
      "rules": [],
      "key": []
    },
    "testenv": {
      "label": "internal/testenv",
      "sub": "TestMain sandbox",
      "test": true,
      "role": "Isolates tests in private temporary/workspace/home directories so lifecycle and trust tests avoid real user configuration; checks temporary workspace leaks.",
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
      "rules": [],
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
    },
    "api": {
      "label": "internal/api",
      "sub": "shared application service",
      "role": "Resolves cards and capabilities for both TUI and MCP. Owns operation admission, request deduplication and safe response DTOs; the session manager still owns all processes.",
      "key": [
        "Service",
        "OperationRunner",
        "OperationResult",
        "SessionInspection"
      ],
      "files": {
        "service.go": "shared lifecycle, discovery and inspection",
        "operations.go": "durable journal and bounded workers",
        "messaging.go": "send, report and bounded reply reads"
      },
      "rules": [
        "14",
        "15",
        "17"
      ]
    },
    "mcp": {
      "label": "internal/mcp",
      "sub": "authenticated local endpoint",
      "role": "Adapts Streamable HTTP MCP tools to the shared service. Validates loopback host, origin, request size and bearer identity; owns the endpoint lifetime, not agent processes.",
      "key": [
        "Runtime"
      ],
      "files": {
        "server.go": "registration, discovery, lifecycle and message tools"
      },
      "rules": [
        "14",
        "15"
      ]
    }
  },
  "flowGroups": [
    {
      "name": "Architecture and entry points",
      "flows": [
        "startup",
        "launch",
        "control",
        "skills"
      ]
    },
    {
      "name": "Conversation and activation identity",
      "flows": [
        "registration",
        "resume",
        "readiness",
        "trust",
        "workspace",
        "mcp-config"
      ]
    },
    {
      "name": "Orchestrator communication",
      "flows": [
        "send",
        "reports",
        "operations"
      ]
    },
    {
      "name": "Safety, recovery and human use",
      "flows": [
        "card-delete",
        "stop-delete",
        "recovery",
        "quit",
        "dashboard",
        "watch"
      ]
    },
    {
      "name": "Start & stop",
      "flows": [
        "resize"
      ]
    },
    {
      "name": "Sessions",
      "flows": [
        "attach",
        "ide",
        "terminal",
        "workspace-command-errors"
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
        "savecard"
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
      "title": "Boot the local control plane",
      "summary": "The same Kankuro process now serves the TUI and authenticated MCP tools.",
      "participants": [
        [
          "main",
          "root.run"
        ],
        [
          "cfg",
          "config.FileRepository"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "runner",
          "api.OperationRunner"
        ],
        [
          "http",
          "mcp.Runtime"
        ],
        [
          "ui",
          "app.Coordinator"
        ]
      ],
      "steps": [
        [
          "main",
          "main",
          "call",
          "Embed and inject both skills",
          "mcp-cli explains access; kankuro-puppet explains registration and explicit reports.",
          "root.main"
        ],
        [
          "main",
          "cfg",
          "call",
          "Open YAML repository",
          "",
          "config.NewFileRepository"
        ],
        [
          "main",
          "mgr",
          "call",
          "Load sessions and message history",
          "Processes restore as terminated; unfinished message delivery becomes uncertain.",
          "session.NewManager"
        ],
        [
          "main",
          "svc",
          "call",
          "Create shared API service",
          "",
          "api.New"
        ],
        [
          "main",
          "runner",
          "call",
          "Load journal and start bounded workers",
          "Four workers; interrupted operations are never replayed automatically.",
          "api.NewOperationRunner"
        ],
        [
          "main",
          "http",
          "io",
          "Bind authenticated loopback MCP",
          "Host/origin checks and request limits wrap Streamable HTTP.",
          "mcp.Start"
        ],
        [
          "main",
          "mgr",
          "call",
          "Set endpoint before any agent launches",
          "",
          "session.Manager.SetMCPEndpoint"
        ],
        [
          "main",
          "ui",
          "call",
          "Build coordinator and run Bubble Tea",
          "TUI actions use a service backed by the same manager.",
          "root.run"
        ]
      ],
      "sources": {
        "refs": []
      }
    },
    "quit": {
      "title": "Quit and stop the control plane",
      "summary": "All managed agent processes stop when Kankuro exits.",
      "participants": [
        [
          "ui",
          "app.Coordinator"
        ],
        [
          "main",
          "root.run"
        ],
        [
          "r",
          "api.OperationRunner"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "http",
          "mcp.Runtime"
        ]
      ],
      "steps": [
        [
          "ui",
          "main",
          "return",
          "Bubble Tea exits",
          "Signals and ordinary quit converge on the same cleanup.",
          "root.run"
        ],
        [
          "main",
          "r",
          "call",
          "Stop admission, cancel and join workers",
          "",
          "api.OperationRunner.Shutdown"
        ],
        [
          "main",
          "mgr",
          "call",
          "Cancel startup timers and stop all processes",
          "Wait activations and watcher/teardown work; no persistent daemon is added.",
          "session.Manager.Shutdown"
        ],
        [
          "main",
          "http",
          "call",
          "Gracefully close endpoint after manager shutdown",
          "The endpoint exists while agents finish teardown.",
          "root.shutdownMCP"
        ],
        [
          "http",
          "main",
          "return",
          "Propagate unexpected endpoint failures",
          "A watcher also sends MCPServerFailedMsg to the UI.",
          "mcp.Runtime.Failures"
        ]
      ],
      "sources": {
        "refs": []
      }
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
      "title": "Launch: one workspace, one fresh PTY",
      "summary": "The TUI and MCP both prepare the chosen workspace, synchronize capabilities and spawn a fresh process through the shared manager.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "act",
          "app.sessionActions"
        ],
        [
          "svc",
          "api.Service"
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
        ]
      ],
      "steps": [
        [
          "user",
          "act",
          "cmd",
          "Launch from card",
          "Blocking work runs inside a tea.Cmd.",
          "app.sessionActions.launch"
        ],
        [
          "act",
          "svc",
          "call",
          "Launch(card)",
          "Shared capability resolution and pending card reservation.",
          "api.Service.Launch"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Allocate new session / activation",
          "No driver or PTY is reused by command name.",
          "session.Manager.LaunchWithIDContext"
        ],
        [
          "mgr",
          "mgr",
          "io",
          "Prepare main and sub workspaces",
          "New sessions use the selected card; temporary copies are session-local.",
          "session.Manager.prepareWorkspace"
        ],
        [
          "mgr",
          "drv",
          "io",
          "Sync files, skills and selected MCP config",
          "No registration block is injected; the puppet skill supplies guidance.",
          "session.Manager.syncDriverFiles"
        ],
        [
          "mgr",
          "mgr",
          "io",
          "Prepare rotating activation credential",
          "",
          "session.Manager.prepareActivationMCPConfig"
        ],
        [
          "mgr",
          "drv",
          "call",
          "Prepare exact workspace trust",
          "Generic hook is a no-op; Codex/Claude update their own configuration.",
          "session.Manager.startSession"
        ],
        [
          "drv",
          "pty",
          "io",
          "Spawn fresh process and PTY",
          "Model and native resume arguments remain driver-owned.",
          "driver.GenericCLIDriver.Spawn"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Publish immediately; start timer and watchers",
          "Humans can attach now; incoming MCP prompts remain gated.",
          "session.Manager.startSession"
        ]
      ],
      "sources": {
        "refs": []
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
    "watch": {
      "title": "Observe process exit and invalidate message waits",
      "summary": "Output remains human-facing; explicit MCP reports remain separate.",
      "participants": [
        [
          "drv",
          "driver.GenericCLIDriver"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "store",
          "session.MessageStore"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "drv",
          "mgr",
          "chan",
          "Track PTY output into bounded scrollback",
          "Output heuristics only describe running/waiting; they do not complete a task.",
          "session.Manager.trackOutput"
        ],
        [
          "mgr",
          "mgr",
          "go",
          "Watch process generation and registration timeout",
          "Old-generation watchers stop affecting replacement activations.",
          "session.Manager.watch"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Invalidate queued message reservations on exit",
          "",
          "session.Manager.cancelPendingMessageAdmissionsLocked"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Revoke token and cancel startup timer",
          "",
          "session.Manager.revokeSessionActivationLocked"
        ],
        [
          "mgr",
          "store",
          "chan",
          "Wake reply waiters",
          "Readers see terminated/replaced recipient rather than waiting for a completion inference.",
          "session.MessageStore.Wake"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Persist terminated record and bindings",
          "",
          "session.Manager.persistSessionLockedErr"
        ]
      ],
      "sources": {
        "refs": [
          "session.updateSessionOutputState",
          "session.Session.snapshot",
          "session.Manager.RegisterConversation"
        ]
      }
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
    },
    "readiness": {
      "title": "Immediate attachment, delayed MCP delivery",
      "summary": "Successful spawn is visible immediately; ten seconds gate only incoming orchestrator prompts.",
      "participants": [
        [
          "mgr",
          "session.Manager"
        ],
        [
          "ui",
          "app.Coordinator"
        ],
        [
          "drv",
          "driver.GenericCLIDriver"
        ],
        [
          "svc",
          "api.Service"
        ]
      ],
      "steps": [
        [
          "mgr",
          "mgr",
          "self",
          "Install process, identity and generation",
          "",
          "session.Manager.activateSessionLocked"
        ],
        [
          "mgr",
          "mgr",
          "go",
          "Start activation-owned timer",
          "Deadline is anchored to spawn and tied to generation, activation and process.",
          "session.Manager.startStartupLocked"
        ],
        [
          "ui",
          "mgr",
          "call",
          "Attach while phase is starting",
          "Human observation does not pause orchestration.",
          "session.Manager.Attach"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Validate generation and finish startup",
          "The timer itself validates identity; lifecycle teardown cancels it separately.",
          "session.Manager.startStartupLocked"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Inspect a consistent snapshot",
          "ProcessStarted, phase and delivery availability are separate facts.",
          "session.Manager.snapshotLocked"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Compute message eligibility",
          "Live process, current authority, MCP selected, no conflicting lifecycle, startup finished.",
          "session.Manager.deliveryAvailableLocked"
        ],
        [
          "svc",
          "svc",
          "return",
          "Use one inspection conversion",
          "Operation outcomes are historical; get_session is current.",
          "api.inspectionFromSnapshot"
        ]
      ],
      "sources": {
        "refs": []
      }
    },
    "registration": {
      "title": "Register a native CLI conversation",
      "summary": "The agent reports its real native ID and meaningful name through its deployed MCP credential.",
      "participants": [
        [
          "agent",
          "agent"
        ],
        [
          "http",
          "mcp.Runtime"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "agent",
          "http",
          "io",
          "register_conversation(id, name)",
          "The built-in puppet skill explains how to discover the actual native ID and report it.",
          "mcp.Start"
        ],
        [
          "http",
          "http",
          "self",
          "Validate bearer, origin and host",
          "",
          "mcp.authorize"
        ],
        [
          "http",
          "svc",
          "call",
          "RegisterConversation(token, id, name)",
          "",
          "api.Service.RegisterConversation"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Authenticate current activation",
          "Derive card/session/driver from the credential, never caller-supplied IDs.",
          "session.Manager.RegisterConversation"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Stage early report or commit binding",
          "Conflicting and stale reports cannot replace another activation.",
          "session.Manager.registerConversationLocked"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Persist native binding in session JSON",
          "Registered/saved and registered/unsaved are distinct outcomes.",
          "session.Manager.persistSessionLockedErr"
        ]
      ],
      "sources": {
        "refs": [
          "session.Manager.activationAuthorityLocked",
          "session.Manager.validateReportedConversationLocked",
          "session.addConversationBinding",
          "session.Manager.publishActivationLocked"
        ]
      }
    },
    "resume": {
      "title": "Choose native conversation on relaunch/switch",
      "summary": "A session retains its workspace while a human or orchestrator chooses a saved native conversation for the selected card, or starts a new one.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "ui",
          "app.Coordinator"
        ],
        [
          "pick",
          "views.ConversationChoiceView"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "drv",
          "driver.AgentDriver"
        ]
      ],
      "steps": [
        [
          "user",
          "ui",
          "key",
          "Request relaunch or switch",
          "",
          "app.Coordinator.requestRelaunch"
        ],
        [
          "ui",
          "svc",
          "call",
          "Load bindings for destination card",
          "",
          "api.Service.Conversations"
        ],
        [
          "ui",
          "pick",
          "msg",
          "Pick saved conversation or New",
          "Even one binding requires an explicit choice.",
          "views.ConversationChoiceView.SetContext"
        ],
        [
          "pick",
          "ui",
          "return",
          "Return explicit choice",
          "",
          "app.Coordinator.conversationChoiceCallbacks"
        ],
        [
          "ui",
          "svc",
          "cmd",
          "RelaunchWithChoice / SwitchAgentWithChoice",
          "Do not substitute the destination card workspace for the saved one.",
          "api.Service.SwitchAgentWithChoice"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Reserve activation and validate binding",
          "Choice belongs to this Kankuro session, destination card and driver.",
          "session.Manager.reserveForActivation"
        ],
        [
          "mgr",
          "mgr",
          "io",
          "Validate saved directory exists",
          "Missing workspaces fail with restoration guidance; never recreate an empty replacement.",
          "session.validateSessionWorkspace"
        ],
        [
          "mgr",
          "drv",
          "call",
          "Build native resume args",
          "Generic --session; Codex resume; Claude --resume.",
          "driver.GenericCLIDriver.SetSession"
        ],
        [
          "mgr",
          "drv",
          "call",
          "Codex resume <id>",
          "",
          "driver.CodexCLIDriver.SetSession"
        ],
        [
          "mgr",
          "drv",
          "call",
          "Claude --resume <id>",
          "",
          "driver.ClaudeCodeDriver.SetSession"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Publish fresh activation in same workspace",
          "Relaunch preserves deployment; explicit switch reapplies selected card configuration.",
          "session.Manager.SwitchAgentWithChoiceContext"
        ]
      ],
      "sources": {
        "refs": [
          "app.Coordinator.requestSwitch",
          "app.Coordinator.handleConversationChoiceRequested",
          "views.ConversationChoiceView.RefreshContext",
          "views.ConversationChoiceView.SelectedBinding",
          "views.ConversationChoiceView.Update",
          "session.validateConversationChoice"
        ]
      }
    },
    "control": {
      "title": "Authenticated discovery and session control",
      "summary": "An ordinary orchestrator session opts into dashboard control and manages other sessions.",
      "participants": [
        [
          "o",
          "orchestrator"
        ],
        [
          "http",
          "mcp.Runtime"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "o",
          "http",
          "io",
          "List cards and sessions",
          "",
          "mcp.addDiscoveryTools"
        ],
        [
          "http",
          "svc",
          "call",
          "ResolveCaller from deployed credential",
          "Permission is captured per activation and preserved on relaunch.",
          "api.Service.ResolveCaller"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Validate published live Dashboard Control",
          "",
          "session.Manager.ResolveDashboardCaller"
        ],
        [
          "svc",
          "svc",
          "return",
          "Return safe paths and current activation data",
          "Listings exclude the caller; no output, environment or credentials are returned.",
          "api.Service.GetSession"
        ],
        [
          "o",
          "http",
          "io",
          "create_session(card, optional name, UUID)",
          "",
          "mcp.addLifecycleTools"
        ],
        [
          "http",
          "svc",
          "call",
          "Create named session",
          "Name enters the first record and request fingerprint; omission keeps card default.",
          "api.Service.CreateSession"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Launch selected existing project card",
          "Orchestrator writes feature/plan files itself using exposed paths.",
          "api.Service.launchReservedSession"
        ]
      ],
      "sources": {
        "refs": []
      }
    },
    "operations": {
      "title": "Retries, cancellation and crash recovery",
      "summary": "Lifecycle mutations are durable requests executed by a bounded worker pool.",
      "participants": [
        [
          "o",
          "orchestrator"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "r",
          "api.OperationRunner"
        ],
        [
          "disk",
          "disk"
        ],
        [
          "mgr",
          "session.Manager"
        ]
      ],
      "steps": [
        [
          "o",
          "svc",
          "call",
          "Submit lifecycle action with stable UUID",
          "",
          "api.Service.CreateSession"
        ],
        [
          "svc",
          "r",
          "call",
          "Replay or admit matching request",
          "Changed arguments conflict; prior outcomes remain historical.",
          "api.OperationRunner.Replay"
        ],
        [
          "r",
          "disk",
          "io",
          "Persist intent before queueing",
          "Four workers, 32 queued requests, five-minute operation deadline.",
          "api.OperationRunner.accept"
        ],
        [
          "r",
          "mgr",
          "cmd",
          "Revalidate authority, execute effect",
          "HTTP disconnect ends waiting, not the admitted task.",
          "api.OperationRunner.execute"
        ],
        [
          "r",
          "disk",
          "io",
          "Persist outcome and safe partial effects",
          "A final-write failure returns observed target IDs with a durability warning.",
          "api.OperationRunner.finishLocked"
        ],
        [
          "o",
          "r",
          "call",
          "Poll / cancel operation",
          "A cancellation after an irreversible effect may arrive too late.",
          "api.OperationRunner.Cancel"
        ],
        [
          "r",
          "r",
          "self",
          "On restart mark unfinished interrupted",
          "Inspect current sessions; do not repeat unknown effects.",
          "api.OperationRunner.loadRecord"
        ]
      ],
      "sources": {
        "refs": []
      }
    },
    "send": {
      "title": "Deliver one explicit orchestrator message",
      "summary": "The sender chooses a recipient activation; Kankuro injects a marked single-line prompt.",
      "participants": [
        [
          "o",
          "orchestrator"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "r",
          "api.OperationRunner"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "store",
          "session.MessageStore"
        ],
        [
          "drv",
          "driver.CodexCLIDriver"
        ],
        [
          "agent",
          "agent"
        ]
      ],
      "steps": [
        [
          "o",
          "svc",
          "io",
          "Send pinned prompt",
          "Single-line UTF-8, at most 2 KiB; complex instructions live in agent-created files.",
          "api.Service.SendMessage"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Reserve activations",
          "Reject starting/stale/non-MCP targets before any prompt admission.",
          "session.Manager.ReserveMessageAdmission"
        ],
        [
          "svc",
          "r",
          "call",
          "Admit operation",
          "A full queue cannot leave an accepted message intent behind.",
          "api.OperationRunner.SubmitWithAdmission"
        ],
        [
          "r",
          "store",
          "io",
          "Persist intent",
          "",
          "session.MessageStore.CreateIntentWithAdmission"
        ],
        [
          "mgr",
          "store",
          "call",
          "Wait FIFO turn",
          "",
          "session.MessageStore.WaitTurn"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Validate activation",
          "",
          "session.Manager.beginMessageSubmission"
        ],
        [
          "mgr",
          "drv",
          "io",
          "Text, pause, Enter",
          "Codex uses a cancellable 500 ms pause; input shares the human write gate.",
          "driver.CodexCLIDriver.SubmitPrompt"
        ],
        [
          "drv",
          "agent",
          "io",
          "Recipient PTY input",
          "Already-typed human text is not cleared; no simultaneous-composition guarantee.",
          "driver.GenericCLIDriver.submitPrompt"
        ],
        [
          "mgr",
          "store",
          "io",
          "Save delivery state",
          "Submission is byte evidence, not proof of processing or completion.",
          "session.MessageStore.UpdateDelivery"
        ]
      ],
      "sources": {
        "refs": [
          "session.Manager.PublishMessageAdmission",
          "session.Manager.SubmitMessage",
          "driver.ValidatePrompt",
          "driver.GenericCLIDriver.submitPromptBytes",
          "driver.writePromptPhases",
          "driver.classifySubmission"
        ]
      }
    },
    "reports": {
      "title": "Explicit progress, questions and completion",
      "summary": "The recipient reports over MCP; the orchestrator reads bounded correlated replies.",
      "participants": [
        [
          "agent",
          "agent"
        ],
        [
          "http",
          "mcp.Runtime"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "store",
          "session.MessageStore"
        ],
        [
          "o",
          "orchestrator"
        ]
      ],
      "steps": [
        [
          "agent",
          "http",
          "io",
          "report_reply(message, kind, text, UUID)",
          "Kinds: progress, completed, question, blocked.",
          "mcp.addMessagingTools"
        ],
        [
          "http",
          "svc",
          "call",
          "ReportReply with stable request identity",
          "",
          "api.Service.ReportReply"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Derive actual recipient identity",
          "Recipient needs Kankuro MCP, but not Dashboard Control.",
          "session.Manager.ReportMessage"
        ],
        [
          "mgr",
          "store",
          "io",
          "Persist idempotent correlated report",
          "Exact recipient activation must match; maximum 16 KiB and 128 reports/message.",
          "session.MessageStore.RecordReply"
        ],
        [
          "o",
          "svc",
          "io",
          "get_replies(cursor, filter, bounded wait)",
          "At most 10 seconds; empty replies do not mean completion.",
          "api.Service.GetReplies"
        ],
        [
          "svc",
          "store",
          "call",
          "Return bounded page and next cursor",
          "Eviction/deletion gaps may repeat IDs; deduplicate by reply ID.",
          "session.MessageStore.GetReplies"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Revalidate caller and recipient during waits",
          "Termination, revocation and deletion wake waiters.",
          "session.Manager.MessageRecipientLive"
        ]
      ],
      "sources": {
        "refs": [
          "session.MessageStore.createIntent",
          "session.MessageStore.evictLocked",
          "session.MessageStore.Recover",
          "api.Service.GetMessage"
        ]
      }
    },
    "trust": {
      "title": "Prepare and remove exact workspace trust",
      "summary": "CLI trust configuration is prepared before spawn and cleaned when a temporary workspace is deleted.",
      "participants": [
        [
          "mgr",
          "session.Manager"
        ],
        [
          "drv",
          "driver.AgentDriver"
        ],
        [
          "trust",
          "trust"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "mgr",
          "drv",
          "call",
          "PrepareWorkspace(actual WorkingDir)",
          "Use the temporary copy, never substitute the source workspace.",
          "session.Manager.startSession"
        ],
        [
          "drv",
          "trust",
          "io",
          "Codex projects[path].trust_level = trusted",
          "~/.codex/config.toml; preserve other table values.",
          "driver.CodexCLIDriver.PrepareWorkspace"
        ],
        [
          "drv",
          "trust",
          "io",
          "Claude projects[path].hasTrustDialogAccepted",
          "Final implementation uses ~/.claude.json, not the earlier settings-array proposal.",
          "driver.ClaudeCodeDriver.PrepareWorkspace"
        ],
        [
          "drv",
          "trust",
          "io",
          "Atomic config replacement",
          "Process-local lock; unrelated JSON values survive through RawMessage.",
          "driver.atomicTrustWrite"
        ],
        [
          "mgr",
          "drv",
          "call",
          "Remove exact grant from both configs",
          "Needed after card switches and after restoring a terminated session.",
          "driver.RemoveWorkspaceTrust"
        ],
        [
          "drv",
          "trust",
          "io",
          "Remove only own trust key/path",
          "Missing files/entries are no-ops; malformed config is preserved and reported.",
          "driver.updateTrust"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Delete temporary workspace after trust cleanup",
          "On cleanup failure retain session for user repair and retry.",
          "session.Manager.RemoveDetailed"
        ]
      ],
      "sources": {
        "refs": [
          "driver.updateCodexTrust",
          "driver.updateClaudeTrust",
          "driver.trustJSONObject",
          "driver.GenericCLIDriver.PrepareWorkspace"
        ]
      }
    },
    "card-delete": {
      "title": "Protect cards during concurrent activation",
      "summary": "Card deletion now shares the activation admission lock across TUI and MCP services.",
      "participants": [
        [
          "ui",
          "app.sessionActions"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "cfg",
          "config.FileRepository"
        ]
      ],
      "steps": [
        [
          "ui",
          "svc",
          "call",
          "DeleteCard(cardID)",
          "The shared manager serializes card deletion with activation admission.",
          "api.Service.DeleteCard"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Reserve current card before preparation",
          "Existence validation and pending reference admission share the manager lock.",
          "session.Manager.ReserveCardReference"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Reject pending and durable references",
          "Includes stopped sessions and saved native bindings.",
          "session.Manager.DeleteCardIfUnreferenced"
        ],
        [
          "mgr",
          "cfg",
          "io",
          "Delete unreferenced card under admission lock",
          "",
          "config.FileRepository.DeleteIDCard"
        ],
        [
          "svc",
          "mgr",
          "return",
          "Release pending reference after publish/rollback",
          "Separate api.Service objects still coordinate through one manager.",
          "api.Service.reserveCardReference"
        ]
      ],
      "sources": {
        "refs": []
      }
    },
    "stop-delete": {
      "title": "Stop and delete with truthful effects",
      "summary": "Waiting still means a live process; stop before switching or deleting.",
      "participants": [
        [
          "o",
          "orchestrator"
        ],
        [
          "svc",
          "api.Service"
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
          "store",
          "session.MessageStore"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "o",
          "svc",
          "call",
          "StopSession(UUID, target)",
          "",
          "api.Service.StopSession"
        ],
        [
          "svc",
          "mgr",
          "call",
          "KillDetailed reserves teardown",
          "Pending messages are invalidated; in-flight input drains before teardown.",
          "session.Manager.killDetailed"
        ],
        [
          "mgr",
          "drv",
          "io",
          "Terminate process group",
          "",
          "driver.GenericCLIDriver.Kill"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Revoke token and persist terminated record",
          "",
          "session.Manager.revokeSessionActivationLocked"
        ],
        [
          "o",
          "svc",
          "call",
          "DeleteSession(UUID, target)",
          "",
          "api.Service.DeleteSession"
        ],
        [
          "svc",
          "mgr",
          "call",
          "RemoveDetailed only for idle terminated session",
          "Trust cleanup failure retains record/workspace for retry.",
          "session.Manager.RemoveDetailed"
        ],
        [
          "mgr",
          "store",
          "io",
          "Remove recipient message content",
          "Compact send receipts prevent old requests from resubmitting evicted/deleted content.",
          "session.MessageStore.DeleteRecipient"
        ]
      ],
      "sources": {
        "refs": []
      }
    },
    "recovery": {
      "title": "Recover bindings and reports after restart",
      "summary": "Saved sessions survive, but live processes and activation credentials do not.",
      "participants": [
        [
          "mgr",
          "session.Manager"
        ],
        [
          "disk",
          "disk"
        ],
        [
          "store",
          "session.MessageStore"
        ],
        [
          "runner",
          "api.OperationRunner"
        ]
      ],
      "steps": [
        [
          "mgr",
          "disk",
          "io",
          "Load session JSON",
          "",
          "session.loadSessions"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Restore terminated session and saved bindings",
          "Saved workspace, grant and conversation maps survive; runtime identity does not.",
          "session.Data.restore"
        ],
        [
          "mgr",
          "store",
          "call",
          "Load bounded message history",
          "",
          "session.NewMessageStore"
        ],
        [
          "store",
          "store",
          "self",
          "Recover unfinished sends as uncertain",
          "Drop history for deleted recipients; never replay input automatically.",
          "session.MessageStore.Recover"
        ],
        [
          "runner",
          "disk",
          "io",
          "Load durable mutation journal",
          "",
          "api.OperationRunner.load"
        ],
        [
          "runner",
          "runner",
          "self",
          "Mark nonterminal entries interrupted",
          "",
          "api.OperationRunner.loadRecord"
        ]
      ],
      "sources": {
        "refs": []
      }
    },
    "dashboard": {
      "title": "Understand sessions on the existing dashboard",
      "summary": "The dashboard gains conversation choice, control opt-in and clearer bounded metadata.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "editor",
          "views.CardEditor"
        ],
        [
          "list",
          "components.SessionListItem"
        ],
        [
          "preview",
          "components.SessionPreview"
        ],
        [
          "dash",
          "views.Dashboard"
        ]
      ],
      "steps": [
        [
          "user",
          "editor",
          "key",
          "Toggle Dashboard Control",
          "Requires Kankuro MCP selection to become effective.",
          "views.CardEditor.renderDashboardControlLine"
        ],
        [
          "list",
          "list",
          "self",
          "Show starting / delivery state in session list",
          "",
          "components.SessionListItem.Description"
        ],
        [
          "preview",
          "preview",
          "self",
          "Reserve text width beside complete puppet art",
          "Long paths fit by cell width; omit art when too narrow.",
          "components.SessionPreview.renderPreview"
        ],
        [
          "preview",
          "preview",
          "self",
          "Wrap centered empty preview text",
          "Avoid double border-width deduction and broken words.",
          "components.SessionPreview.renderPlaceholder"
        ],
        [
          "dash",
          "dash",
          "self",
          "Keep pane alignment across sizes and selection",
          "",
          "views.Dashboard.View"
        ]
      ],
      "sources": {
        "refs": [
          "components.renderPreviewField",
          "components.renderedBlockWidth",
          "components.wrapPreviewWords",
          "components.SessionList.refreshItems",
          "components.cardDelegate.Render",
          "views.Dashboard.renderPreviewPane"
        ]
      }
    },
    "workspace": {
      "title": "Keep main and sub-workspace identity",
      "summary": "Temporary copies are prepared for new sessions; relaunch never replaces a saved working directory.",
      "participants": [
        [
          "svc",
          "api.Service"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "svc",
          "mgr",
          "call",
          "Resolve the selected project card",
          "",
          "api.Service.resolveCard"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Copy main workspace with cancellation",
          "Only a new temporary session gets a new main copy.",
          "session.Manager.copyWorkspace"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Serialize sub-workspace preparation",
          "The manifest owns copied destinations; existing managed work is retained.",
          "session.Manager.prepareSubWorkspaces"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Stage and publish selected sub-workspace",
          "Cancel or fail preparation with rollback of this transaction.",
          "session.copyAndPublishSubWorkspace"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Capture deployed source/destination descriptors",
          "Descriptors survive relaunch; legacy inspection reads existing manifest without preparation.",
          "session.ReadSubWorkspaceDescriptors"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Return actual paths to orchestrator",
          "Extra feature/plan/artifact files are written by agents themselves.",
          "api.Service.GetSession"
        ]
      ],
      "sources": {
        "refs": [
          "session.preparedSubWorkspaceDescriptors",
          "session.Manager.cleanupNewTemporaryWorkspace",
          "session.Manager.resolveWorkspacePath",
          "session.Manager.RelaunchWithChoiceContext"
        ]
      }
    },
    "mcp-config": {
      "title": "Deploy MCP credentials in the workspace",
      "summary": "Each activation rotates its token and writes the selected servers to one workspace-local config.",
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
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "mgr",
          "mgr",
          "self",
          "Create pending activation credential",
          "",
          "session.Manager.prepareActivationMCPConfig"
        ],
        [
          "mgr",
          "drv",
          "io",
          "Serialize complete selected MCP configuration",
          "Built-in Kankuro receives current endpoint and bearer; credentials are not copied into operation journals.",
          "driver.GenericCLIDriver.SelectMCPs"
        ],
        [
          "drv",
          "disk",
          "io",
          "Atomically replace .kankuro/mcp-servers.json",
          "Separate temporary workspaces isolate credentials.",
          "driver.writeWorkspaceMCPConfig"
        ],
        [
          "mgr",
          "drv",
          "io",
          "On relaunch refresh only Kankuro entry",
          "Preserve ordinary MCP entries and workspace metadata.",
          "driver.RefreshKankuroMCP"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Revoke credential when activation ends",
          "",
          "session.Manager.revokeAuthorityLocked"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Remove entry only when credential matches",
          "Stopping an older shared-workspace activation must not remove the newer token.",
          "driver.RemoveKankuroMCPIfAuthorizationMatches"
        ]
      ],
      "sources": {
        "refs": [
          "driver.MCPConfigPath",
          "session.syncActivationMCPConfig",
          "session.activationEnv"
        ]
      }
    },
    "skills": {
      "title": "Use the built-in puppet skill for registration",
      "summary": "Selecting Kankuro MCP deploys access and puppet-protocol skills; registration guidance lives in the skill.",
      "participants": [
        [
          "main",
          "root.main"
        ],
        [
          "cfg",
          "config.FileRepository"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "drv",
          "driver.GenericCLIDriver"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "main",
          "cfg",
          "call",
          "Inject both embedded skill byte arrays",
          "",
          "config.SetBuiltinKankuroPuppetSkillContent"
        ],
        [
          "svc",
          "svc",
          "self",
          "Ensure access and puppet skills are selected",
          "Automatically append only the missing built-in skills.",
          "api.appendKankuroAccessSkill"
        ],
        [
          "svc",
          "drv",
          "io",
          "Deploy selected skills in driver folder",
          "",
          "driver.GenericCLIDriver.SetSkills"
        ],
        [
          "drv",
          "disk",
          "io",
          "Resolve embedded content and write skill",
          "The root owns go:embed; config owns synchronized cloned bytes.",
          "driver.writeBuiltinSkill"
        ],
        [
          "drv",
          "disk",
          "io",
          "Synchronize configured instructions",
          "Preserve configured card/persona behavior; no registration block injection or legacy stripping.",
          "driver.syncInstructionFile"
        ]
      ],
      "sources": {
        "refs": [
          "config.BuiltinSkills",
          "config.BuiltinMCPs",
          "config.GetBuiltinSkillContent",
          "driver.writeInstructionFile"
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
        "resume"
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
      "title": "Keep sub-workspace ownership and publication serialized",
      "ref": "session.Manager.prepareSubWorkspaces",
      "why": "Sub-workspace manifest decisions and copy publication share the manager preparation gate. Existing managed copies retain work; concurrent preparations must not publish conflicting destinations.",
      "check": "Review the mutex scope and transaction bookkeeping; run workspace/collision/failure tests and go test -race ./....",
      "flows": [
        "workspace",
        "launch"
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
      "title": "Main supplies built-in skill bytes; config owns protected copies",
      "ref": "config.SetBuiltinMCPCLISkillContent",
      "why": "main embeds mcp-cli and kankuro-puppet content; config clones bytes on injection and lookup under separate mutexes. Drivers deploy built-in content by stable identity.",
      "check": "TestBuiltinSkillContentOwnershipAndMissingValues, TestBuiltinSkillContentConcurrentReads and TestEmbeddedMCPCLISkillDeploysThroughConfig.",
      "flows": [
        "startup",
        "skills"
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
    },
    {
      "n": "13",
      "title": "Bindings belong to a session/card; credentials belong to an activation",
      "ref": "session.Manager.RegisterConversation",
      "why": "A report authenticated with an old activation must not change the native conversation for a replacement. Bindings persist; runtime credentials do not. Registration confirmation and persistence are separate outcomes.",
      "check": "conversation_test.go and MCP registration tests exercise early reports, stale tokens, conflicts and persistence failures.",
      "flows": [
        "registration",
        "resume",
        "recovery"
      ]
    },
    {
      "n": "14",
      "title": "Dashboard access comes from deployed activation permission",
      "ref": "session.Manager.ResolveDashboardCaller",
      "why": "Caller-supplied session IDs are not identity. Dashboard Control requires a live published credential and selected Kankuro MCP; recipients can report without controller permission.",
      "check": "service and MCP tests cover omitted permission, revoked credentials and scoped discovery/control.",
      "flows": [
        "control",
        "registration",
        "reports"
      ]
    },
    {
      "n": "15",
      "title": "Retries must not repeat process or input effects",
      "ref": "api.OperationRunner",
      "why": "Persist intent before execution, deduplicate by caller/request/fingerprint, and retain compact send receipts. Restart marks uncertain effects interrupted rather than executing them again.",
      "check": "OperationRunner tests plus messaging admission/persistence/ordering regressions.",
      "flows": [
        "operations",
        "send",
        "recovery"
      ]
    },
    {
      "n": "16",
      "title": "Published process is distinct from message availability",
      "ref": "session.Manager.deliveryAvailableLocked",
      "why": "Humans attach immediately after spawn. A ten-second activation-owned timer gates MCP messages; old timers must not enable a replacement activation. Waiting and silence never mean task completion.",
      "check": "startup lifecycle tests use an injected clock and cover early registration, replacement, exit and shutdown.",
      "flows": [
        "readiness",
        "watch",
        "send"
      ]
    },
    {
      "n": "17",
      "title": "Delete cards only under the shared reference guard",
      "ref": "session.Manager.DeleteCardIfUnreferenced",
      "why": "A card can be referenced by a pending launch/switch before it appears in session listings. Serialize existence checks, pending references and deletion across TUI/MCP service instances.",
      "check": "Pending-card cross-service create/switch and deletion-first regression tests.",
      "flows": [
        "card-delete",
        "launch"
      ]
    },
    {
      "n": "18",
      "title": "Explicit reports carry user-defined workflow meaning",
      "ref": "session.MessageStore.RecordReply",
      "why": "Progress, completed, question and blocked are reports correlated to a message and recipient activation. They do not cause automatic stages, and submitted PTY bytes do not prove processing.",
      "check": "MessageStore and API/MCP messaging tests cover identity, deduplication, bounds and cursor gaps.",
      "flows": [
        "send",
        "reports"
      ]
    }
  ],
  "tasks": [
    {
      "q": "Add a new agent CLI",
      "a": "Embed GenericCLIDriver and override workspace trust preparation, native resume arguments or coordinated prompt submission where the CLI differs. Register a constructor; each activation gets a new driver.",
      "refs": [
        "driver.AgentDriver",
        "driver.GenericCLIDriver",
        "driver.NewBuiltinRegistry"
      ],
      "flow": "trust"
    },
    {
      "q": "Change what gets written into a session's workspace",
      "a": "Driver synchronization owns instructions, skills, personas and workspace MCP config; registration guidance is deployed through the puppet skill. No legacy registration-text rewriting remains. The manager owns main/sub-workspace preparation.",
      "refs": [
        "driver.syncInstructionFile",
        "driver.GenericCLIDriver.SetSkills",
        "session.Manager.prepareSubWorkspaces"
      ],
      "flow": "skills"
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
      "q": "Add a session operation for TUI and MCP",
      "a": "Put shared rules in api.Service and the session manager. Add an MCP adapter and a TUI command adapter; reuse permission, operation journaling and partial-success semantics.",
      "refs": [
        "api.Service",
        "api.OperationRunner",
        "mcp.addLifecycleTools",
        "app.sessionActions"
      ],
      "flow": "control"
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
      "flow": "workspace"
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
      "flow": "skills"
    },
    {
      "q": "Change session names or activation snapshots",
      "a": "CreateSession accepts an optional name before first persistence. Keep one inspectionFromSnapshot conversion for operation outcomes and current get_session; outcomes are historical snapshots.",
      "refs": [
        "api.Service.CreateSession",
        "api.inspectionFromSnapshot",
        "session.Manager.snapshotLocked"
      ],
      "flow": "control"
    },
    {
      "q": "Change native conversation registration or resume",
      "a": "Registration identity is derived from an activation credential; native bindings remain session/card-local. Driver SetSession methods build native arguments; the conversation picker handles explicit new/resume choices.",
      "refs": [
        "session.Manager.RegisterConversation",
        "session.validateConversationChoice",
        "views.ConversationChoiceView",
        "driver.ClaudeCodeDriver.SetSession"
      ],
      "flow": "resume"
    },
    {
      "q": "Change prompt delivery or explicit report kinds",
      "a": "Follow API admission, durable message intent, manager final activation checks and driver input serialization. Keep report kinds and puppet skill documentation consistent; progress never implies completion.",
      "refs": [
        "api.Service.SendMessage",
        "session.Manager.SubmitMessage",
        "driver.GenericCLIDriver.SubmitPrompt",
        "session.MessageStore.RecordReply"
      ],
      "flow": "send"
    },
    {
      "q": "Change startup delay or workspace trust",
      "a": "Trust configuration runs before spawn; process publication and human access are immediate. Only MCP delivery waits on the activation timer. Remove exact temporary trust grants from both CLI configs on deletion.",
      "refs": [
        "session.Manager.startStartupLocked",
        "session.Manager.deliveryAvailableLocked",
        "driver.updateTrust",
        "session.Manager.RemoveDetailed"
      ],
      "flow": "readiness"
    },
    {
      "q": "Change retry, cancellation or restart behavior",
      "a": "Use the existing bounded OperationRunner and durable journal. Keep accepted intent, observed effects, warnings and interrupted recovery distinct; never replay an uncertain send automatically.",
      "refs": [
        "api.OperationRunner",
        "session.MessageStore.Recover",
        "api.SafeRejection"
      ],
      "flow": "operations"
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
      ],
      [
        "operations/<caller>-<request>.json",
        "durable lifecycle request and outcome"
      ],
      [
        "messaging/history.json",
        "bounded messages, reports and compact dedup receipts"
      ],
      [
        "<WorkingDir>/.kankuro/mcp-servers.json",
        "selected MCP servers and current activation credential"
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
      "External application completion means process startup, not eventual exit. No command is inferred; launch failures use the normal error banner without navigation.",
      "Codex trust is stored under ~/.codex/config.toml projects[path].trust_level; Claude trust uses ~/.claude.json projects[path].hasTrustDialogAccepted. Exact temporary-path trust is removed on deletion; a config cleanup failure retains the session for repair/retry.",
      "Temporary workspaces isolate MCP configs. Sessions sharing an original workspace intentionally accept last-writer server/credential identity. Older cleanup removes only its matching credential.",
      "Message history is bounded by 64 MiB or 10,000 whole messages per dashboard. Reports are capped at 16 KiB and 128/message; reply pages cap at 100 reports/256 KiB, with at most ten-second waits. Compact send receipts outlive content eviction.",
      "Startup phase is runtime-only. Sessions restore terminated; bindings and reports persist; unfinished operation effects are interrupted and message submissions become uncertain."
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
    "Session-launching tests use newTestManager; testenv.Run fails the package if anything is left in its private temp/home root (about 3 s per package).",
    "Shape tests check that a render fits its box; position tests check that elements stay put as the content changes.",
    "Built-in content tests inject explicit fixtures and restore the previous value with t.Cleanup; do not run content-replacing tests in parallel. The root integration test deploys the real embedded bytes through config and the generic driver.",
    "Terminal launch integration records a helper process’s real working directory and arguments. Substituting OriginalDir for WorkingDir must make TestTerminalLaunchThroughDashboard fail; shared-launcher tests cover repeated quoted placeholders and commands without one.",
    "Terminal action routing covers running, waiting and terminated selections, search, empty selections and nil callbacks. Settings tests cover independent buffers, save/cancel, display sanitization and bounded layout; the coordinator race harness includes terminal launch.",
    "Trust and lifecycle tests isolate HOME; do not run native agent experiments against live factory workspaces. Claude runtime validation is taking place in the user’s separate environment.",
    "The latest cleanup removal passed focused driver regressions, build, vet, full normal/race suites and lint (0 issues). Evidence: .kankuro/remove-legacy-cleanup-verification/. Those source checks are not native CLI compatibility tests."
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
      "Built-in skill explaining access to the selected workspace MCP configuration; kankuro-puppet separately explains native registration and explicit reports."
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
    ],
    [
      "Activation",
      "One live process generation in a Kankuro session, with runtime identity, credential, deployed permission and startup timer."
    ],
    [
      "Conversation binding",
      "Native CLI conversation ID, name and driver saved per card within one Kankuro session."
    ],
    [
      "Dashboard Control",
      "Opt-in deployed permission for an activation to discover and control other sessions."
    ],
    [
      "Message delivery availability",
      "Kankuro can admit a prompt for this activation; this does not prove CLI readiness or task completion."
    ],
    [
      "Progress reply",
      "Explicit correlated information about ongoing work; it has no completion or transition effect."
    ],
    [
      "Operation journal",
      "Durable request identity, arguments fingerprint and observed outcome for retry/recovery."
    ]
  ]
};
