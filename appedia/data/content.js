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
    ],
    "registry": [
      "Workspace registry",
      "Registered IDs, names and paths"
    ],
    "client": [
      "MCP caller",
      "Live Dashboard Control"
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
  "snapshot": "Workspace discovery and initial launch overrides · source e741572",
  "pitch": "Organize agent configurations and sessions into independent dashboards. Share selected cards across dashboards while keeping each session, workspace identity and conversation binding in its creation dashboard.",
  "pitch2": "Bubble Tea uses independent application contracts through focused capabilities. The application service owns membership and operations. The center keeps the selected-session header above authored reports or bounded workspace information. Raw terminal output is available through the human attachment.",
  "concepts": [
    [
      "Dashboard",
      "Persistent ID, name and description owning sessions and card creation.",
      "config.Dashboard"
    ],
    [
      "ID card",
      "Saved agent template for the CLI, model, optional reasoning effort, workspace and capabilities.",
      "config.IDCard"
    ],
    [
      "Session",
      "Immutable dashboard membership, workspace identity and per-card native bindings.",
      "session.Session"
    ],
    [
      "Activation",
      "One process generation, deployed credential and message startup gate.",
      "session.Snapshot"
    ],
    [
      "Scoped service",
      "Common membership policy for human UI and authenticated MCP.",
      "api.Service"
    ],
    [
      "Conversation binding",
      "Saved native CLI conversation within one session and card.",
      "session.ConversationBinding"
    ],
    [
      "Message and reply",
      "Pinned task delivery, correlated reports and durable wake-ups to the original sender.",
      "session.MessageRecord"
    ],
    [
      "Deletion manifest",
      "Durable cascade effects and cleanup checkpoints for retry.",
      "config.DashboardDeletion"
    ]
  ],
  "path": [
    [
      "See the package boundaries and common UI/MCP service.",
      "map"
    ],
    [
      "Open a dashboard and return safely to the list.",
      "flow-navigation"
    ],
    [
      "Trace scoped launch through publication, confirmed naming and session-focus restoration.",
      "flow-launch"
    ],
    [
      "See why MCP scope follows the caller session.",
      "flow-caller-scope"
    ],
    [
      "Follow reports from persistence to an automatic orchestrator wake-up.",
      "flow-reports"
    ],
    [
      "Read the invariants before changing lifecycle or recovery.",
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
        "session",
        "pty",
        "spawn / input"
      ],
      [
        "driver",
        "disk",
        "MCP config / trust"
      ],
      [
        "ui",
        "contracts",
        "focused contracts"
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
      ],
      [
        "ui",
        "api",
        "capability calls"
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
      ],
      "configstartup": [
        580,
        220
      ],
      "contracts": [
        1040,
        360
      ],
      "models": [
        350,
        500
      ],
      "ui": [
        350,
        220
      ],
      "terminal": [
        350,
        360
      ]
    }
  },
  "packages": {
    "views": {
      "label": "Terminal views",
      "sub": "forms, lists and modals",
      "role": "Owns terminal screens and their interaction state. The dashboard preserves launch focus through refreshed naming, filters and session-list rebuilds, and hosts the existing-session rename dialog. The launch form and card editor share workspace rows and pickers.",
      "files": {
        "dashboard.go": "Three panes, keys, search, launch-focus reconciliation and existing-session rename modal with pending-save/error handling",
        "card_editor.go": "card fields, one key switch, focused sub workspace clearing, source shortcuts, skills filter",
        "mcp_form.go": "Create/edit form; overlays visible fields to preserve loaded metadata",
        "source_manager.go": "source directories per registry",
        "settings.go": "Independent IDE/terminal command inputs, focus, save/cancel and bounded rendering",
        "session_launcher.go": "Workspace confirmation, scoped cancellation, naming, and attach handoff",
        "agent_switch.go": "pick a card for a terminated session",
        "workspace_picker_modal.go": "single/multiple selection, provisional checks, missing IDs",
        "puppet_selector_modal.go": "modal: choose a puppet",
        "agentsmd_editor_modal.go": "modal: AGENTS.md content or path",
        "constants.go": "MinTerminalWidth/Height, CoordinatorRows",
        "source_type.go": "per-registry labels",
        "dashboard_list.go": "list-first navigation, forms and delete/retry controls",
        "mcp_manager.go": "Global MCP list, stable selection, protected built-in handling and scoped delete confirmation",
        "workspace_fields.go": "Shared workspace rows, selection summaries, and picker state for editing and launch"
      },
      "notes": [
        "The entry screen is the dashboard list. Escape cancels a dialog or stays on the list; q quits there.",
        "Delete/Backspace resumes an incomplete dashboard deletion; Enter inspects its status.",
        "Select an existing session and press e to rename; Enter saves and Esc cancels before submission. Pending saves block repeated submit/dismissal; errors retain the draft for retry.",
        "After launching a card, the sessions pane selects the created session. A confirmed name is reconciled against fresh session data; only a filter hiding the target is cleared.",
        "MCPManager reuses SourceList and Modal; the coordinator owns scoped navigation and off-loop registry mutations."
      ],
      "rules": [],
      "key": [
        "DashboardList",
        "Dashboard",
        "CardEditor"
      ],
      "sources": {
        "refs": [
          "views.CardEditor.Draft",
          "views.CardEditor.Reset",
          "views.CardEditor.LoadEditor",
          "views.CardEditor.SetDashboardContext",
          "views.CardEditor.SetChoices",
          "views.CardEditor.SetSettings",
          "views.CardEditor.Update",
          "views.CardEditor.fieldBlocks",
          "views.CardEditor.handleCtrlE",
          "views.CardEditor.handleSpace",
          "views.CardEditor.renderDashboardControlLine",
          "views.CardEditor.renderSharedLine",
          "views.CardEditor.renderTmpWorkspaceLine",
          "views.CardEditor.renderWorkspaceLine",
          "views.CardEditor.visibleFocusOrder",
          "views.ConversationChoiceView.RefreshContext",
          "views.ConversationChoiceView.SelectedOption",
          "views.ConversationChoiceView.Update",
          "views.Dashboard.SetDashboardName",
          "views.Dashboard.SetDashboardScope",
          "views.Dashboard.SetInventory",
          "views.Dashboard.SetSize",
          "views.Dashboard.ShowCardDeletionImpact",
          "views.Dashboard.Update",
          "views.Dashboard.View",
          "views.Dashboard.footerText",
          "views.Dashboard.handleDeleteCard",
          "views.Dashboard.handleEdit",
          "views.Dashboard.handleEnter",
          "views.Dashboard.handleKey",
          "views.Dashboard.handleSessionEnter",
          "views.Dashboard.handleSwitch",
          "views.Dashboard.renderBase",
          "views.Dashboard.renderCenterPane",
          "views.Dashboard.updateCardImpact",
          "views.DashboardCallbacks",
          "views.DashboardList.Init",
          "views.DashboardList.SetDashboards",
          "views.DashboardList.SetSize",
          "views.DashboardList.Update",
          "views.DashboardList.View",
          "views.DashboardList.confirmDeletion",
          "views.DashboardList.openForm",
          "views.DashboardList.renderPage",
          "views.DashboardList.retryDeletion",
          "views.DashboardList.selected",
          "views.DashboardList.updateConfirmation",
          "views.DashboardList.updateForm",
          "views.NewDashboardList",
          "views.SessionLauncher.Reset",
          "views.SessionLauncher.SetNameResult",
          "views.SessionLauncher.handleNameSubmit",
          "views.SettingsView.TerminalInput",
          "views.SettingsView.renderDisplay",
          "views.SettingsView.renderEdit",
          "views.WorkspacePickerModal.SetExcludedID",
          "views.WorkspacePickerModal.SetSelectedID",
          "views.WorkspacePickerModal.SetSelectedIDs",
          "views.WorkspacePickerModal.View",
          "views.WorkspacePickerModal.restoreSingleSelection",
          "views.WorkspacePickerModal.title",
          "views.dashboardItem.Description",
          "views.dashboardItem.FilterValue",
          "views.dashboardItem.Title",
          "views.editorKeyIdentity",
          "views.focusTargetTable",
          "views.workspaceItem.Title",
          "views.NewDashboard",
          "views.dashboardModalInputWidth",
          "views.Dashboard.CompleteRename",
          "views.Dashboard.updateRename",
          "views.Dashboard.closeRename",
          "views.SessionLauncher",
          "views.workspaceFields"
        ]
      }
    },
    "messages": {
      "label": "UI messages",
      "sub": "scoped async results",
      "role": "Defines scoped launch, naming and refresh events; refresh messages carry a session epoch so pre-mutation snapshots can be rejected.",
      "files": {
        "messages.go": "UI intents and scoped contract results"
      },
      "notes": [
        "DataLoadedMsg is always a full snapshot; partial updates get their own type. Session snapshots carry the mutation epoch captured when their read started.",
        "DashboardSessionRenamedMsg is separate from launcher naming: successful rename refreshes without attaching, and includes captured dashboard, generation and target identity."
      ],
      "rules": [],
      "key": [],
      "sources": {
        "refs": [
          "messages.ScopedActionFinishedMsg",
          "messages.DashboardSessionRenamedMsg"
        ]
      }
    },
    "session": {
      "label": "Sessions",
      "sub": "process and workspace owner",
      "role": "Owns processes, immutable membership and activation lifecycle; stores reports and notification evidence, reserves exact destinations and submits through one checked message path.",
      "files": {
        "manager.go": "Manager, Session, Snapshot",
        "lifecycle.go": "Launch, Relaunch, SwitchAgent, Kill, Remove, Shutdown",
        "watch.go": "status polling every 2 s",
        "io.go": "output tracking and scrollback",
        "attach.go": "AttachHandle and the locked input writer",
        "storage.go": "session JSON, atomic",
        "workspace.go": "temporary workspace copies",
        "ide_launcher.go": "shared workspace-command parsing, startup and reaping",
        "sub_workspaces.go": "source resolution, ownership manifest, copy staging/publication, rollback",
        "workspace_launcher.go": "Shared path resolution, placeholder replacement, direct process spawn and asynchronous reaping",
        "conversation.go": "activation credentials and native conversation reports",
        "messaging.go": "atomic reports, child intents and delivery evidence",
        "card_references.go": "pending card reference and deletion guard",
        "startup.go": "startup policy and unpublished-process teardown",
        "message_notifications.go": "pending snapshots, attempt updates and exact-target reservation"
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
        "18",
        "25",
        "26",
        "27"
      ],
      "key": [
        "Manager",
        "Session",
        "Data",
        "CardReferenceImpact",
        "MessageReply",
        "ReplyNotification"
      ],
      "sources": {
        "refs": [
          "session.Data.restore",
          "session.Manager.AttachChecked",
          "session.Manager.CancelDashboardAdmissions",
          "session.Manager.CancelPendingCardReferences",
          "session.Manager.CardReferences",
          "session.Manager.List",
          "session.Manager.PublishMessageAdmission",
          "session.Manager.RegisterConversation",
          "session.Manager.RelaunchWithChoiceCheckedContext",
          "session.Manager.RemoveResidual",
          "session.Manager.ReserveCardDeletion",
          "session.Manager.ReserveCardReferenceForDashboardWithCancel",
          "session.Manager.ReserveDashboardDeletion",
          "session.Manager.SetNameDetailedChecked",
          "session.Manager.ShutdownStarted",
          "session.Manager.SubmitMessage",
          "session.Manager.TemporaryWorkspaceRoot",
          "session.Manager.WaitCardReferences",
          "session.Manager.WaitDashboardAdmissions",
          "session.Manager.WithAdmissionSnapshot",
          "session.Manager.WithCardReferencesAndSnapshot",
          "session.Manager.activationAuthorityLocked",
          "session.Manager.bindPendingWorkspace",
          "session.Manager.cardReferencesLocked",
          "session.Manager.cleanupNewTemporaryWorkspace",
          "session.Manager.listLocked",
          "session.Manager.notifyCardReferencesLocked",
          "session.Manager.publishActivationLocked",
          "session.Manager.reserveCardReference",
          "session.Manager.resolveWorkspacePath",
          "session.Manager.validateReportedConversationLocked",
          "session.MessageStore.Recover",
          "session.MessageStore.createIntent",
          "session.MessageStore.evictLocked",
          "session.NewManager",
          "session.Session.data",
          "session.Session.snapshot",
          "session.SplitCommand",
          "session.activationEnv",
          "session.addConversationBinding",
          "session.loadSessions",
          "session.preparedSubWorkspaceDescriptors",
          "session.saveSession",
          "session.syncActivationMCPConfig",
          "session.updateSessionOutputState",
          "session.validateConversationChoice",
          "session.activationArgs",
          "session.Manager.prepareSubWorkspacesSnapshot"
        ]
      }
    },
    "components": {
      "label": "UI components",
      "sub": "lists, text and center",
      "role": "Reusable terminal widgets; deleted-card sessions remain visible with an explanatory description.",
      "files": {
        "card_list.go": "ID-card rendering, pills, ClipBlock",
        "session_list.go": "sessions pane list",
        "capability_list.go": "checklists, search and selected-only filtering, cursor normalization",
        "source_list.go": "source directories list",
        "modal.go": "Modal and ModalState",
        "text.go": "DisplayText, TruncateHead, TruncateTail",
        "renderer.go": "puppet rendering",
        "filter.go": "generic search",
        "source_item.go": "source list items",
        "constants.go": "layout constants",
        "session_center.go": "Bounded report and workspace presentation"
      },
      "notes": [
        "Fit by display width, never bytes. Clip art, never wrap it (rule 9).",
        "Selected-only mode changes visible items, not the full item collection or selection map. Normalize the cursor after registry and selection changes.",
        "SourceList retains source-manager defaults while providing configurable labels and bounded MCP-list rendering.",
        "Extractor caveat: implements arrows use method-name matching. SessionCenter.SetData(SessionCenterData) does not satisfy messages.DataReceiver.SetData(DataLoadedMsg); that inferred arrow is false. Open source signatures for authority."
      ],
      "rules": [],
      "key": [],
      "sources": {
        "refs": [
          "components.CapabilityList.Deselect",
          "components.CapabilityList.Select",
          "components.CapabilityList.SelectedOnly",
          "components.CapabilityList.SetCursor",
          "components.CapabilityList.SetSelected",
          "components.CapabilityList.Toggle",
          "components.CapabilityList.normalizeCursor",
          "components.DisplayText",
          "components.Filter",
          "components.SessionList.refreshItems",
          "components.SessionListItem.Description",
          "components.centerSafeText",
          "components.TruncateHead",
          "components.cardDelegate.Render",
          "components.SessionCenter.View",
          "components.ClampToSize",
          "components.SessionCenter.View"
        ]
      }
    },
    "component": {
      "label": "View interface",
      "sub": "layout contract",
      "role": "Provides the common terminal component contract and sizing state.",
      "files": {
        "component.go": "Base, Model"
      },
      "notes": [],
      "rules": [],
      "key": [],
      "sources": {
        "refs": []
      }
    },
    "styles": {
      "label": "Terminal theme",
      "sub": "colours and spacing",
      "role": "Defines shared terminal styling without domain state.",
      "files": {
        "styles.go": "colours and styles"
      },
      "notes": [],
      "rules": [],
      "key": [],
      "sources": {
        "refs": []
      }
    },
    "driver": {
      "label": "CLI drivers",
      "sub": "trust, PTY and deployment",
      "role": "Owns serialized PTY input, prompt formatting and driver-specific Enter timing. A typed envelope distinguishes a task from a report notification; neither header grants authority.",
      "files": {
        "driver.go": "typed prompt formatting and serialized PTY input",
        "registry.go": "name → constructor",
        "trust.go": "exact workspace trust preparation and removal",
        "mcp_config.go": "atomic selected-server config and credential-matched cleanup"
      },
      "notes": [
        "Codex and Claude send text, wait 500 ms, then write Enter. Submitted bytes do not prove processing.",
        "Trust edits preserve configuration structurally, with atomic writes and a process-local update lock."
      ],
      "rules": [
        "12",
        "13",
        "16",
        "27",
        "28"
      ],
      "key": [
        "PromptEnvelope"
      ],
      "sources": {
        "refs": [
          "driver.ClaudeCodeDriver.SubmitPrompt",
          "driver.CodexCLIDriver.SubmitPrompt",
          "driver.GenericCLIDriver.PrepareWorkspace",
          "driver.GenericCLIDriver.submitPromptBytes",
          "driver.MCPConfigPath",
          "driver.ValidatePrompt",
          "driver.classifySubmission",
          "driver.trustJSONObject",
          "driver.updateClaudeTrust",
          "driver.updateCodexTrust",
          "driver.writeInstructionFile",
          "driver.writePromptPhases",
          "driver.AgentDriver",
          "driver.AntigravityCLIDriver",
          "driver.NewAntigravityCLIDriver"
        ]
      }
    },
    "ascii": {
      "label": "ASCII rendering",
      "sub": "terminal canvas",
      "role": "Supplies the canvas and drawing primitives used by puppet rendering.",
      "files": {
        "canvas.go": "bounded cell/text drawing surface",
        "shape.go": "Shape interface",
        "options.go": "style configuration"
      },
      "notes": [],
      "rules": [],
      "key": [],
      "sources": {
        "refs": []
      }
    },
    "characters": {
      "label": "Puppet artwork",
      "sub": "faces and characters",
      "role": "Defines built-in faces and puppet art; unchanged by dashboard membership.",
      "files": {
        "puppet.go": "built-in puppets",
        "face.go": "agent faces",
        "text_puppet.go": "user puppets"
      },
      "notes": [
        "Faces have fixed 5×1 bounds; fixed-text puppets derive bounds from their first line and line count."
      ],
      "rules": [],
      "key": [],
      "sources": {
        "refs": []
      }
    },
    "config": {
      "label": "Configuration",
      "sub": "registries and recovery",
      "role": "Persists dashboards, owned/shared cards, tombstones and deletion manifests atomically; capability and workspace registries remain global.",
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
        "Cards select Kankuro MCP and opt into Dashboard Control separately. The permission is captured at activation.",
        "Cards retain creation ownership and an explicit shared flag. Tombstones retain deleted identities without executable configuration."
      ],
      "rules": [],
      "key": [
        "Dashboard",
        "IDCard",
        "CardTombstone",
        "DashboardDeletion"
      ],
      "sources": {
        "refs": [
          "config.BuiltinMCPs",
          "config.BuiltinSkills",
          "config.FileRepository.BeginCardDeletion",
          "config.FileRepository.BeginCardDeletionWithPending",
          "config.FileRepository.BeginDashboardDeletion",
          "config.FileRepository.CreateDashboard",
          "config.FileRepository.CreateIDCard",
          "config.FileRepository.FinalizeDashboardDeletion",
          "config.FileRepository.FinishCardDeletion",
          "config.FileRepository.GetAppSettings",
          "config.FileRepository.ListCardTombstones",
          "config.FileRepository.ListDashboardDeletions",
          "config.FileRepository.ListDashboardTombstones",
          "config.FileRepository.ListDashboards",
          "config.FileRepository.LoadIDCardsFile",
          "config.FileRepository.SaveCardDeletionPending",
          "config.FileRepository.SaveDashboardDeletion",
          "config.FileRepository.UpdateDashboardDetails",
          "config.FileRepository.UpdateIDCard",
          "config.FileRepository.WriteIDCardsFile",
          "config.FileRepository.checkCardNotTombstonedLocked",
          "config.FileRepository.listDashboardDeletionsLocked",
          "config.FileRepository.loadCardTombstonesLocked",
          "config.FileRepository.loadDashboardTombstonesLocked",
          "config.FileRepository.loadDashboardsLocked",
          "config.FileRepository.retainDashboardTombstoneLocked",
          "config.FileRepository.writeDashboardDeletionLocked",
          "config.GetBuiltinSkillContent",
          "config.NewFileRepository",
          "config.OpenExistingFileRepository",
          "config.ResolveDirectory",
          "config.dashboardDeletionFilename",
          "config.registryFiles",
          "config.validTemporaryIdentity",
          "config.validateCardMembershipFormat",
          "config.validateDashboard",
          "config.validateDashboardDeletion",
          "config.validateNewCardPending",
          "config.IDCard"
        ]
      }
    },
    "fsutil": {
      "label": "Atomic file writes",
      "sub": "shared storage primitive",
      "role": "Writes restrictive temporary files, syncs and renames them for configuration, session and migration durability.",
      "files": {
        "atomic.go": "WriteFileAtomic"
      },
      "notes": [
        "A crash leaves at worst a .<name>.<rand>.tmp file, which loaders ignore."
      ],
      "rules": [],
      "key": [],
      "sources": {
        "refs": []
      }
    },
    "root": {
      "label": "Kankuro",
      "sub": "application startup",
      "role": "Acquires the configuration-root lock, validates startup, shares one scoped service between MCP and the dashboard-list TUI, and coordinates shutdown.",
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
          "root.run"
        ]
      }
    },
    "uitest": {
      "label": "UI assertions",
      "sub": "test helpers",
      "test": true,
      "role": "Checks rendered terminal bounds and visual structure in view tests.",
      "files": {
        "shape.go": "AssertExactRect, AssertWithin, Region, AdversarialNames"
      },
      "notes": [
        "Shape tests check fit; position tests check that elements stay put. A list-like component needs both."
      ],
      "rules": [],
      "key": [],
      "sources": {
        "refs": []
      }
    },
    "testenv": {
      "label": "Test isolation",
      "sub": "temporary-root guard",
      "test": true,
      "role": "Isolates test files and checks that session/app/API suites leave no temporary data behind.",
      "files": {
        "testenv.go": "Run(m, name)"
      },
      "notes": [
        "Session-launching tests build managers with newTestManager, which registers a Shutdown cleanup."
      ],
      "rules": [],
      "key": [],
      "sources": {
        "refs": []
      }
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
      "label": "Application service",
      "sub": "membership and operations",
      "role": "Shared dashboard policy and operation execution. Ordinary sends and report wake-ups share prepared delivery hooks and the bounded runner; narrowly scoped report authority grants no general worker control. Workspace discovery checks deployed Dashboard Control before reading the global registry. Initial launch validates the selected IDs under dashboard admission.",
      "key": [
        "Service",
        "membershipInventory",
        "persistedDashboardScope",
        "OperationRunner",
        "preparedDelivery"
      ],
      "files": {
        "service.go": "shared lifecycle, discovery and inspection",
        "operations.go": "bounded execution, authorization and cancellation",
        "messaging.go": "prepared delivery hooks and message/report APIs",
        "membership.go": "scoped human operations and admission",
        "persisted_scope.go": "authenticated caller membership",
        "card_confirmation.go": "owner confirmation and foreign impact",
        "dashboard_deletion.go": "checkpointed cascading cleanup",
        "notifications.go": "owned dispatcher, retry scheduling and report authority",
        "center_process_unix.go": "Owned Git process-group cancellation",
        "center_process_other.go": "Non-Unix direct process fallback"
      },
      "rules": [
        "14",
        "15",
        "17",
        "25",
        "26",
        "27"
      ],
      "sources": {
        "refs": [
          "api.New",
          "api.Service.AttachInDashboard",
          "api.Service.CheckSessionInDashboard",
          "api.Service.ConfirmCardDeletion",
          "api.Service.ConfirmDashboardDeletion",
          "api.Service.CreateEditorCard",
          "api.Service.CreateDashboard",
          "api.Service.GetDashboardDeletion",
          "api.Service.GetMessage",
          "api.Service.GetSession",
          "api.Service.LaunchSession",
          "api.Service.ListDashboards",
          "api.Service.ListSessions",
          "api.Service.PreflightCardDeletion",
          "api.Service.PreflightDashboardDeletion",
          "api.Service.ReadDashboardOverview",
          "api.Service.RelaunchInDashboard",
          "api.Service.RemoveDashboardSession",
          "api.Service.RenameDashboardSession",
          "api.Service.ResolveCaller",
          "api.Service.RetryCardDeletion",
          "api.Service.RetryDashboardDeletion",
          "api.Service.StopDashboardSession",
          "api.Service.SwitchInDashboard",
          "api.Service.UpdateEditorCard",
          "api.Service.UpdateDashboardDetails",
          "api.Service.authorizeExistingOperation",
          "api.Service.availableDashboardCard",
          "api.Service.cardOwnedBy",
          "api.Service.cleanupCardPending",
          "api.Service.dashboardAlreadyDeleted",
          "api.Service.dashboardDeletion",
          "api.Service.deleteDashboardCard",
          "api.Service.deletedCardIDs",
          "api.Service.inspectionFromScopedSnapshot",
          "api.Service.launchDashboardWithID",
          "api.Service.listCardTombstones",
          "api.Service.listCardsForCaller",
          "api.Service.loadMembershipInventory",
          "api.Service.prepareDashboardDeletion",
          "api.Service.removeDashboardSession",
          "api.Service.runDashboardDeletion",
          "api.Service.validateCardConfirmationLocked",
          "api.Service.validateDeletionInventory",
          "api.Service.validateDeletionOwner",
          "api.cardImpactDigest",
          "api.cardPendingCleanup",
          "api.dashboardImpactDigest",
          "api.hasForeignCardReferences",
          "api.inspectionFromSnapshot",
          "api.membershipInventory.activeDashboard",
          "api.membershipInventory.allowsCard",
          "api.membershipInventory.allowsSession",
          "api.membershipInventory.hasTombstone",
          "api.newMembershipInventory",
          "api.persistedDashboardScope.AllowsCard",
          "api.persistedDashboardScope.AllowsSession",
          "api.persistedDashboardScope.DashboardForSession",
          "api.persistedDashboardScope.SessionDashboard",
          "api.persistedDashboardScope.inventory",
          "api.persistedDashboardScope.validateScopeInventory",
          "api.sessionInDashboard",
          "api.sessionUsesCard",
          "api.cardValues",
          "api.applyCardChanges",
          "api.Service.ListWorkspaces",
          "api.Service.ReadLaunchPreparation",
          "api.Service.CreateSessionWithOptions"
        ]
      }
    },
    "mcp": {
      "label": "MCP transport",
      "sub": "loopback tools",
      "role": "Authenticated loopback tools. Report input and reply wrapper stay compatible; tool descriptions explain wake-ups and additive delivery status. Agents use list_workspaces to discover IDs for create_session.",
      "key": [],
      "files": {
        "server.go": "registration, discovery, lifecycle and message tools"
      },
      "rules": [
        "14",
        "15",
        "25",
        "28"
      ],
      "sources": {
        "refs": [
          "mcp.addDiscoveryTools",
          "mcp.addLifecycleTools"
        ]
      }
    },
    "configstartup": {
      "label": "Config startup",
      "sub": "validation and root locking",
      "role": "Prepares configuration for safe startup: exclusive root locking, direct initialization of empty directories and validation of complete current-format records.",
      "key": [
        "Lock"
      ],
      "files": {
        "startup.go": "fresh initialization and current-format entry gate",
        "validate.go": "strict cross-record membership validation",
        "lock.go": "exclusive config-root process lock",
        "json.go": "reject malformed durable JSON"
      },
      "sources": {
        "refs": [
          "configstartup.AcquireLock",
          "configstartup.Lock.Close",
          "configstartup.PrepareStartup",
          "configstartup.ValidateLayout",
          "configstartup.jsonValue",
          "configstartup.validateJSON",
          "configstartup.validateLayoutHistory",
          "configstartup.validateLayoutSessions"
        ]
      }
    },
    "contracts": {
      "label": "api/contracts",
      "sub": "independent facts",
      "role": "Define independent application inputs, outcomes and resource contracts consumed by the frontend. No backend storage records or terminal presentation types.",
      "files": {
        "center.go": "Reports, workspace entries and subscription contract",
        "inventory.go": "Dashboard/card/session focused summaries",
        "card_editor.go": "Editable values and revision changes"
      },
      "key": [
        "SessionReport",
        "ReportSubscription",
        "SessionWorkspaceResult",
        "DashboardInventory"
      ],
      "sources": {
        "refs": [
          "contracts.CardValues",
          "contracts.CardChanges"
        ]
      }
    },
    "models": {
      "label": "ui/models",
      "sub": "presentation data",
      "role": "Own display rows, local editor drafts and detached center presentation state. Application policy and persistence remain behind contracts.",
      "files": {
        "session_center.go": "Center state and deep-copy boundary"
      },
      "key": [
        "SessionCenterData",
        "SessionRow",
        "CardEditorDraft"
      ]
    },
    "ui": {
      "label": "UI coordinator",
      "sub": "Scoped command routing",
      "role": "Owns navigation, captured commands, request ordering, and selected-session resource ownership through consumer capability interfaces.",
      "files": {
        "coordinator.go": "the Coordinator struct, Update, message routing",
        "navigation.go": "navigateTo (rule 2), initViews, resizeAll, broadcasts",
        "handlers.go": "Apply scoped async results",
        "refresh.go": "refresher: read commands, fingerprint check, tick",
        "session_actions.go": "Session actions, shared workspace-command lookup, IDE and terminal launch commands",
        "callbacks.go": "every view's callbacks",
        "attach.go": "Hands an owned attachment to the terminal command and processes the result",
        "size_guard.go": "below 80×24: notice, keys dropped",
        "app.go": "tea.Model wrapper",
        "dashboard_list.go": "dashboard inventory, scope-aware callbacks and cascade confirmation",
        "scoped_refresh.go": "captured scope, generation and fingerprint-aware commands",
        "center.go": "Center reads, detached navigation cache and recovery"
      },
      "notes": [
        "Commands capture dashboard ID, view generation, settings and session-data epoch before running off the coordinator goroutine.",
        "Successful launch and rename advance the epoch so delayed session snapshots cannot displace the selected target.",
        "Test-only adapters in export_test.go exercise the current scoped handlers.",
        "MCP forms have explicit return destinations. Manager commands carry dashboard/generation/visit identity; discarded unadmitted confirmations release pending state."
      ],
      "rules": [],
      "key": [
        "Coordinator",
        "sessionActions",
        "refresher"
      ],
      "sources": {
        "refs": [
          "ui.Coordinator.Init",
          "ui.Coordinator.agentSwitchCallbacks",
          "ui.Coordinator.cardEditorCallbacks",
          "ui.Coordinator.conversationChoiceCallbacks",
          "ui.Coordinator.dashboardCallbacks",
          "ui.Coordinator.dashboardListCallbacks",
          "ui.Coordinator.guardKey",
          "ui.Coordinator.handleSessionMessages",
          "ui.Coordinator.handleConfigFingerprint",
          "ui.Coordinator.prepareConversation",
          "ui.Coordinator.handleDataLoaded",
          "ui.Coordinator.reconcileCenter",
          "ui.Coordinator.handleSessionMessages",
          "ui.Coordinator.handleSessionsRefreshed",
          "ui.Coordinator.handleSpecialMessages",
          "ui.Coordinator.launchCard",
          "ui.Coordinator.launchIDE",
          "ui.Coordinator.launchTerminal",
          "ui.Coordinator.maybeRefreshConfig",
          "ui.Coordinator.reconcileCenter",
          "ui.Coordinator.mcpCallbacks",
          "ui.Coordinator.navigateTo",
          "ui.Coordinator.navigationCommand",
          "ui.Coordinator.refreshDashboardList",
          "ui.Coordinator.refreshDataSeeded",
          "ui.Coordinator.refreshSessions",
          "ui.Coordinator.requestAttach",
          "ui.Coordinator.requestRelaunch",
          "ui.Coordinator.requestSwitch",
          "ui.Coordinator.resizeAll",
          "ui.Coordinator.sessionLauncherCallbacks",
          "ui.Coordinator.settingsCallbacks",
          "ui.New",
          "ui.NewCoordinator",
          "ui.cardDeletionPreflightCommand",
          "ui.confirmCardDeletionCommand",
          "ui.convertViewType",
          "ui.dashboardDeletionStatus",
          "ui.dashboardDeletionSummary",
          "ui.dashboardListCommand",
          "ui.Coordinator.clearCenter",
          "ui.refresher.scopedConfigIfChanged",
          "ui.refresher.scopedFull",
          "ui.refresher.scopedSessions",
          "ui.retryCardDeletionCommand",
          "ui.scopedAttachCommand",
          "ui.sessionActions.launch",
          "ui.sessionActions.launchWorkspace",
          "ui.sessionActions.launchWorkspace",
          "ui.sessionActions.launchWorkspace",
          "ui.sessionActions.activate",
          "ui.sessionActions.activate",
          "ui.Coordinator.renameDashboardSession"
        ]
      }
    },
    "terminal": {
      "label": "Human terminal",
      "sub": "Raw mode and detach",
      "role": "Owns raw terminal input, keyboard protocol restoration, resize observation and detach for an abstract attachment.",
      "files": {
        "attach.go": "Manages the human terminal, attachment input/output, resize and detach"
      }
    }
  },
  "flowGroups": [
    {
      "name": "Dashboard ownership and recovery",
      "flows": [
        "dashboard-details",
        "navigation",
        "caller-scope",
        "manual-session",
        "card-recovery",
        "deleted-session",
        "delete-dashboard",
        "cascade-recovery",
        "list-errors",
        "deletion-outcomes"
      ]
    },
    {
      "name": "Architecture and entry points",
      "flows": [
        "startup",
        "launch",
        "control",
        "mcp-workspace-discovery",
        "mcp-workspace-create",
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
        "report-progress",
        "notification-dispatch",
        "notification-write",
        "report-inspection",
        "notification-authority",
        "notification-retry",
        "notification-cancel",
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
        "resize",
        "reasoning-effort-activation"
      ]
    },
    {
      "name": "Sessions",
      "flows": [
        "launch-workspace-selection",
        "launch-workspace-navigation",
        "launch-focus",
        "attach",
        "ide",
        "terminal",
        "workspace-command-errors",
        "rename-edit",
        "rename-save",
        "rename-layout",
        "attach-disposal"
      ]
    },
    {
      "name": "Keeping the screen fresh",
      "flows": [
        "tick",
        "preview",
        "center-workspace",
        "center-recovery",
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
        "mcp-open",
        "mcp",
        "mcp-delete",
        "mcp-stale",
        "mcp-card",
        "settings",
        "settings-edit",
        "workspace-choices-refresh"
      ]
    }
  ],
  "flows": {
    "startup": {
      "title": "Prepare configuration and start the control plane",
      "summary": "Lock and validate configuration, wire one scoped service, then start loopback MCP and the dashboard-list TUI.",
      "participants": [
        [
          "boot",
          "root.run"
        ],
        [
          "setup",
          "configstartup.PrepareStartup"
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
          "ui.Coordinator"
        ]
      ],
      "steps": [
        [
          "boot",
          "boot",
          "call",
          "Inject built-in skills",
          "Root main supplies the MCP CLI and puppet skill bytes.",
          "root.main"
        ],
        [
          "boot",
          "setup",
          "call",
          "Lock configuration root",
          "Held throughout startup and runtime.",
          "configstartup.AcquireLock"
        ],
        [
          "boot",
          "setup",
          "call",
          "Initialize or validate",
          "Empty roots initialize directly; existing roots require complete current-format records.",
          "configstartup.PrepareStartup"
        ],
        [
          "boot",
          "repo",
          "call",
          "Open repository",
          "",
          "config.NewFileRepository"
        ],
        [
          "boot",
          "repo",
          "call",
          "Sync skill registry",
          "",
          "config.FileRepository.SyncSkillsFromSources"
        ],
        [
          "boot",
          "repo",
          "call",
          "Sync puppet registry before the UI opens",
          "Runs while holding the configuration lock. Missing or invalid sources are reported as startup warnings and the application continues.",
          "config.FileRepository.SyncPuppetsFromSources"
        ],
        [
          "boot",
          "mgr",
          "call",
          "Restore sessions and history",
          "Processes restore as terminated; unfinished submissions become uncertain.",
          "session.NewManager"
        ],
        [
          "boot",
          "svc",
          "call",
          "Create shared scoped service",
          "The concrete persisted policy applies to both UI and MCP.",
          "api.New"
        ],
        [
          "boot",
          "runner",
          "call",
          "Load durable operation journal",
          "Bounded workers retain interrupted outcomes rather than replaying uncertain effects.",
          "api.NewOperationRunner"
        ],
        [
          "boot",
          "svc",
          "call",
          "Wire runner and start report dispatcher",
          "SetOperationRunner starts the owned dispatcher before MCP requests can arrive.",
          "api.Service.SetOperationRunner"
        ],
        [
          "boot",
          "ui",
          "call",
          "Build list-first UI",
          "Pass independent capabilities and any session-load or puppet-sync warnings to the list-first UI.",
          "ui.New"
        ],
        [
          "boot",
          "http",
          "io",
          "Start loopback MCP",
          "Start authenticated loopback MCP and set its endpoint on the session manager before any activation.",
          "mcp.Start"
        ],
        [
          "boot",
          "ui",
          "cmd",
          "Initialize entry view and refresh",
          "Read the registry baseline after the startup writes have finished. Begin the periodic refresh from the coordinator.",
          "ui.Coordinator.Init"
        ]
      ],
      "sources": {
        "refs": [
          "ui.New",
          "config.CardRepository",
          "config.DashboardRepository",
          "config.NewFileRepository",
          "config.OpenExistingFileRepository",
          "config.ResolveDirectory",
          "configstartup.AcquireLock",
          "configstartup.Lock",
          "configstartup.Lock.Close",
          "configstartup.PrepareStartup",
          "configstartup.ValidateLayout",
          "configstartup.jsonValue",
          "configstartup.validateJSON",
          "configstartup.validateLayoutHistory",
          "configstartup.validateLayoutSessions",
          "root.run",
          "config.FileRepository.SyncPuppetsFromSources",
          "session.Manager.SetMCPEndpoint",
          "ui.startupErrorCommand",
          "ui.refresher.scopedFull"
        ]
      }
    },
    "quit": {
      "title": "Quit and stop the control plane",
      "summary": "All managed agent processes stop when Kankuro exits.",
      "participants": [
        [
          "ui",
          "ui.Coordinator"
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
          "Join dispatcher, cancel and join workers",
          "Stop the report dispatcher before closing the shared operation queue; then stop sessions and MCP in main.",
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
      "summary": "How views get their size, and what happens below 80×24. Rename input follows modal width; long save errors are sanitized and truncated to keep save/cancel controls visible at 80×24.",
      "participants": [
        [
          "tea",
          "tea"
        ],
        [
          "coord",
          "ui.Coordinator"
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
          "ui.Coordinator.handleSpecialMessages"
        ],
        [
          "coord",
          "dash",
          "call",
          "resizeAll → SetSize(w, h − CoordinatorRows)",
          "Every view gets the terminal minus the rows the coordinator keeps back; the dashboard adds them back to fill the screen exactly.",
          "ui.Coordinator.resizeAll"
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
          "ui.Coordinator.View"
        ],
        [
          "coord",
          "coord",
          "self",
          "tooSmall() → tooSmallView()",
          "Below 80×24 every screen is replaced by a centred notice sized to the terminal.",
          "ui.Coordinator.tooSmallView"
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
          "ui.Coordinator.Update"
        ],
        [
          "coord",
          "coord",
          "self",
          "guardKey: dropped (q / Ctrl+C still quit)",
          "A blind keypress must not launch or delete anything. Non-key messages keep flowing, so the screen is current when the terminal grows.",
          "ui.Coordinator.guardKey"
        ]
      ],
      "sources": {
        "refs": [
          "views.SettingsView.renderDisplay",
          "views.SettingsView.renderEdit"
        ],
        "files": [
          "internal/ui/views/settings_test.go",
          "internal/ui/views/dashboard_layout_test.go"
        ]
      }
    },
    "launch": {
      "title": "Launch with membership before publication",
      "summary": "Confirmed dashboard requests and MCP creates share scoped admission. Initial workspace selection is applied before publication.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "c",
          "ui.Coordinator"
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
          "driver",
          "driver.GenericCLIDriver"
        ],
        [
          "disk",
          "disk"
        ]
      ],
      "steps": [
        [
          "u",
          "c",
          "key",
          "Confirm workspace selection",
          "The coordinator captures the displayed IDs when the form submits.",
          "views.SessionLauncher.updateWorkspaceForm"
        ],
        [
          "c",
          "svc",
          "cmd",
          "LaunchSession with workspace IDs",
          "MCP create uses the same backend launch path.",
          "api.Service.LaunchSession"
        ],
        [
          "svc",
          "mgr",
          "call",
          "ReserveCardReferenceForDashboardWithCancel",
          "",
          "session.Manager.ReserveCardReferenceForDashboardWithCancel"
        ],
        [
          "svc",
          "svc",
          "self",
          "Apply workspace IDs",
          "Validate and apply the selection to a launch-only copy of the card. Compare the saved card during publication.",
          "api.Service.cardWithWorkspaceOverrides"
        ],
        [
          "svc",
          "mgr",
          "call",
          "LaunchInDashboardWithIDResult",
          "",
          "session.Manager.LaunchInDashboardWithIDResult"
        ],
        [
          "mgr",
          "mgr",
          "call",
          "bindPendingWorkspace before any copy",
          "Newly added during review: exact paths survive later card edits.",
          "session.Manager.bindPendingWorkspace"
        ],
        [
          "mgr",
          "driver",
          "call",
          "startSession: prepare workspace and spawn",
          "The shared argument builder adds model, optional reasoning effort and native conversation arguments before spawning.",
          "session.Manager.startSession"
        ],
        [
          "mgr",
          "disk",
          "io",
          "publish + save with immutable dashboard_id",
          "Revalidates card/dashboard under admission; rejects deletion races.",
          "session.Session.data"
        ],
        [
          "mgr",
          "c",
          "msg",
          "SessionLaunchCompletedMsg(session ID)",
          "A usable launch advances the session epoch and requests focus for the exact ID.",
          "ui.Coordinator.handleSessionMessages"
        ]
      ],
      "sources": {
        "refs": [
          "session.CardPendingReference",
          "session.CardReferenceImpact",
          "session.CardSessionReference",
          "session.Manager.CancelPendingCardReferences",
          "session.Manager.CardReferences",
          "session.Manager.List",
          "session.Manager.ReserveCardReferenceForDashboardWithCancel",
          "session.Manager.SetNameDetailedChecked",
          "ui.Coordinator.handleDataLoaded",
          "ui.Coordinator.handleSessionMessages",
          "ui.Coordinator.handleSessionsRefreshed",
          "session.Manager.ShutdownStarted",
          "session.Manager.TemporaryWorkspaceRoot",
          "session.Manager.WaitCardReferences",
          "session.Manager.WithCardReferencesAndSnapshot",
          "session.Manager.bindPendingWorkspace",
          "session.Manager.cardReferencesLocked",
          "session.Manager.listLocked",
          "session.Manager.notifyCardReferencesLocked",
          "session.Manager.reserveCardReference",
          "session.NewManager",
          "session.Session.data",
          "session.Session.snapshot",
          "session.cardReservation",
          "views.SessionLauncher",
          "views.SessionLauncher.Reset",
          "views.SessionLauncher.SetNameResult",
          "views.SessionLauncher.handleNameSubmit",
          "views.SessionLauncherCallbacks",
          "views.Dashboard.FocusSession",
          "views.Dashboard.FocusSessionAfterRename",
          "views.Dashboard.applyPendingSessionFocus",
          "views.matchesSessionQuery",
          "components.SessionList.refreshItems",
          "ui.Coordinator.hasStaleSessionSnapshot",
          "messages.SessionsRefreshedMsg",
          "session.activationArgs"
        ],
        "files": [
          "internal/ui/launch_focus_internal_test.go",
          "internal/ui/inline_rename_refresh_internal_test.go",
          "internal/ui/components/session_list_test.go",
          "tests/e2e/launch_focus_test.go"
        ]
      }
    },
    "attach": {
      "title": "Attach and detach within the opened dashboard",
      "summary": "Membership is checked before attachment; obsolete async handles are discarded after navigation.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "ui",
          "ui.Coordinator"
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
          "exec",
          "terminal.Command"
        ]
      ],
      "steps": [
        [
          "user",
          "ui",
          "key",
          "Attach selected session",
          "",
          "ui.Coordinator.requestAttach"
        ],
        [
          "ui",
          "svc",
          "cmd",
          "AttachInDashboard",
          "Capture dashboard and view generation before asynchronous work.",
          "api.Service.AttachInDashboard"
        ],
        [
          "svc",
          "mgr",
          "call",
          "AttachChecked",
          "Validate ownership while creating the subscription.",
          "session.Manager.AttachChecked"
        ],
        [
          "mgr",
          "ui",
          "msg",
          "ScopedAttachReadyMsg",
          "Discard handles from an obsolete view generation.",
          "ui.Coordinator.handleSessionMessages"
        ],
        [
          "ui",
          "exec",
          "cmd",
          "Release UI terminal; run attach",
          "",
          "ui.scopedAttachCommand"
        ],
        [
          "exec",
          "exec",
          "io",
          "Prepare raw terminal and output",
          "",
          "terminal.Command.Run"
        ],
        [
          "exec",
          "exec",
          "go",
          "Pump process output",
          "",
          "terminal.Command.pumpOutput"
        ],
        [
          "exec",
          "exec",
          "call",
          "Pump terminal input",
          "",
          "terminal.Command.pumpInput"
        ],
        [
          "exec",
          "exec",
          "call",
          "Forward through locked driver input",
          "",
          "terminal.Command.forwardInput"
        ],
        [
          "exec",
          "exec",
          "call",
          "Recognize detach key",
          "Restore terminal state on return.",
          "terminal.inputDecoder.forward"
        ],
        [
          "exec",
          "ui",
          "msg",
          "AttachFinishedMsg",
          "Only matching scope/generation may update the view.",
          "ui.Coordinator.handleSessionMessages"
        ]
      ],
      "sources": {
        "refs": [
          "ui.scopedAttachCommand",
          "session.driverInputWriter"
        ],
        "files": []
      }
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
          "session.Manager.RegisterConversation",
          "session.Session.snapshot",
          "session.updateSessionOutputState"
        ]
      }
    },
    "ide": {
      "title": "Open in IDE",
      "summary": "A captured command calls the application workspace operation, which resolves saved workspace and settings and revalidates scope.",
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
          "ui.Coordinator"
        ],
        [
          "tea",
          "tea"
        ],
        [
          "act",
          "ui.sessionActions"
        ],
        [
          "ide",
          "session.LaunchWorkspaceCommand"
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
          "ui.Coordinator.launchIDE"
        ],
        [
          "tea",
          "act",
          "cmd",
          "launchWorkspace with captured scope/kind",
          "Frontend returns a command using independent contracts.",
          "ui.sessionActions.launchWorkspace"
        ],
        [
          "act",
          "act",
          "call",
          "Check captured membership",
          "Validate the session against the dashboard captured by the command.",
          "api.Service.CheckSessionInDashboard"
        ],
        [
          "act",
          "ide",
          "call",
          "OpenSessionWorkspace(scope, kind)",
          "Application owns settings lookup and saved-directory validation.",
          "api.Service.OpenSessionWorkspace"
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
          "launchWorkspace with captured scope/kind",
          "Frontend returns a command using independent contracts.",
          "ui.sessionActions.launchWorkspace"
        ],
        [
          "coord",
          "coord",
          "self",
          "Apply scoped completion",
          "Reject stale dashboard/generation results; display current errors without navigation.",
          "ui.Coordinator.handleSessionMessages"
        ]
      ],
      "sources": {
        "refs": [
          "session.SplitCommand"
        ],
        "files": [
          "internal/session/ide_launcher.go",
          "internal/session/ide_launcher_test.go"
        ]
      }
    },
    "tick": {
      "title": "Refresh only the opened dashboard",
      "summary": "Inventory and registry refresh use focused contract reads. Center updates have their own subscription and 30-second fallback cadence.",
      "participants": [
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "r",
          "ui.refresher"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "dash",
          "views.Dashboard"
        ]
      ],
      "steps": [
        [
          "c",
          "c",
          "msg",
          "Periodic inventory tick",
          "Capture scope/generation/epoch and request identities.",
          "ui.Coordinator.refreshSessions"
        ],
        [
          "c",
          "r",
          "cmd",
          "scopedSessions",
          "No coordinator state is accessed by the command.",
          "ui.refresher.scopedSessions"
        ],
        [
          "r",
          "svc",
          "call",
          "ReadDashboardOverview",
          "Persisted membership scopes available cards and owned sessions.",
          "api.Service.ReadDashboardOverview"
        ],
        [
          "r",
          "c",
          "msg",
          "SessionsRefreshedMsg",
          "Reject stale snapshots; keep successful data on read failure.",
          "ui.Coordinator.handleSessionsRefreshed"
        ],
        [
          "c",
          "dash",
          "call",
          "Apply focused inventory",
          "",
          "views.Dashboard.SetInventory"
        ],
        [
          "c",
          "c",
          "call",
          "Reconcile center selection without polling workspace",
          "Only changed selection or increased report limit starts new center work.",
          "ui.Coordinator.reconcileCenter"
        ]
      ],
      "sources": {
        "files": []
      }
    },
    "preview": {
      "title": "Selected-session reports in the center",
      "summary": "Authored reports span activations. Subscribe before reading; coalesce changes during reads. Healthy report discovery remains immediate, while failed reads wait for timed recovery.",
      "participants": [
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "store",
          "session.MessageStore"
        ],
        [
          "sub",
          "api.reportSubscription"
        ],
        [
          "panel",
          "components.SessionCenter"
        ]
      ],
      "steps": [
        [
          "dash",
          "dash",
          "self",
          "Reconcile center selection by stable session ID",
          "Stop or activation changes do not change the selected session identity.",
          "views.Dashboard.syncCenterSelection"
        ],
        [
          "dash",
          "panel",
          "call",
          "Refresh the selected-session header",
          "Keep current name, status, ID, workspace, start time, card and puppet artwork visible. A refresh of the same session preserves its body.",
          "components.SessionCenter.SetSession"
        ],
        [
          "dash",
          "c",
          "msg",
          "Reconcile selection and size",
          "Restore matching detached content on return to the same dashboard/session, while canceling old operations. A different selection cannot inherit it.",
          "ui.Coordinator.reconcileCenter"
        ],
        [
          "c",
          "svc",
          "cmd",
          "SubscribeSessionReports(scope)",
          "Capture scope/context/generation before the command runs.",
          "ui.Coordinator.centerSubscribe"
        ],
        [
          "svc",
          "store",
          "call",
          "Capture report and scope signals before reading",
          "Check dashboard membership before and after setup. The application draws identity from the human dashboard selection.",
          "api.Service.SubscribeSessionReports"
        ],
        [
          "sub",
          "store",
          "go",
          "Rearm retained-report changes and revalidate scope",
          "Committed report additions and removals coalesce into one buffered notice. Lifecycle notices recheck scope separately. Delivery changes wake readers only when they evict retained reports.",
          "api.watchReportChanges"
        ],
        [
          "c",
          "svc",
          "cmd",
          "ReadSessionReports(scope, row-derived limit)",
          "Setup completion starts the initial read; notices during a read set one pending follow-up.",
          "ui.Coordinator.centerRead"
        ],
        [
          "svc",
          "store",
          "call",
          "RecentAuthoredReports",
          "Dashboard plus reply author filter, timestamp/sequence descending, stable ID tie. Up to 256 reports and 4 MiB; existing durable retention is unchanged.",
          "session.MessageStore.RecentAuthoredReports"
        ],
        [
          "svc",
          "c",
          "msg",
          "Handle report result",
          "Successful reports cancel obsolete workspace work; mere store notices leave it running. Failures preserve same-session stale reports and wait for the recovery cadence.",
          "ui.Coordinator.handleCenter"
        ],
        [
          "c",
          "panel",
          "call",
          "Publish detached center data",
          "Warnings and structured content are copied into presentation state.",
          "ui.Coordinator.publishCenter"
        ],
        [
          "panel",
          "panel",
          "self",
          "Render reports beneath the session header",
          "Keep timestamps on visible reports. Omit older entries that cannot fit a timestamp and body, and compact the newest report into one row when needed.",
          "components.SessionCenter.View"
        ]
      ],
      "sources": {
        "files": [
          "internal/api/center.go",
          "internal/session/authored_reports.go",
          "internal/ui/center.go",
          "internal/ui/components/session_center.go",
          "internal/session/messaging.go"
        ],
        "refs": [
          "components.SessionCenter.headerRows",
          "components.SessionCenter.ReportLimit",
          "components.RenderPuppetArtwork",
          "models.CloneSessionRow",
          "session.MessageStore.ReportsChanged",
          "session.MessageStore.ReportScopeChanged",
          "session.MessageStore.signalRecordedReportLocked",
          "components.centerReportHeader"
        ]
      }
    },
    "mutation": {
      "title": "Refresh after a change, and errors",
      "summary": "How a command asks for fresh data, and how the error banner lives and dies. Existing-session rename refreshes after either save outcome; its error stays in the dialog rather than the coordinator banner.",
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
          "ui.Coordinator"
        ],
        [
          "ref",
          "ui.refresher"
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
          "ui.Coordinator.refreshDataSeeded"
        ],
        [
          "tea",
          "ref",
          "cmd",
          "scopedFull(scope, generation, epoch, request IDs)",
          "Load global capabilities and scoped cards/sessions, then seed the fingerprint baseline.",
          "ui.refresher.scopedFull"
        ],
        [
          "ref",
          "coord",
          "msg",
          "configFingerprintMsg{token, data}",
          "",
          "ui.Coordinator.handleConfigFingerprint"
        ],
        [
          "coord",
          "views",
          "call",
          "handleDataLoaded → broadcastData",
          "Reject obsolete dashboard/generation and session-epoch snapshots before applying any data; nil failed fields preserve prior values.",
          "ui.Coordinator.handleDataLoaded"
        ],
        [
          "cmd",
          "coord",
          "msg",
          "ErrorMsg",
          "Sets the banner (c.err).",
          "ui.Coordinator.handleSpecialMessages"
        ],
        [
          "user",
          "coord",
          "key",
          "the next keypress",
          "Clears the banner. Navigation alone never does, so an error batched with a navigation survives (rule 4).",
          "ui.Coordinator.Update"
        ]
      ],
      "sources": {
        "refs": [
          "ui.Coordinator.handleSessionMessages"
        ],
        "files": [
          "internal/ui/terminal_launch_test.go"
        ]
      }
    },
    "editcard": {
      "title": "Read and open the card editor",
      "summary": "Scoped reads supply editable values and opaque revision. Async results are checked before navigation.",
      "participants": [
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "ed",
          "views.CardEditor"
        ]
      ],
      "steps": [
        [
          "dash",
          "c",
          "key",
          "e on selected card",
          "",
          "ui.Coordinator.requestCardEditor"
        ],
        [
          "c",
          "svc",
          "cmd",
          "LoadCardEditor or NewCardEditor",
          "Capture scope, epoch and request before blocking work.",
          "ui.cardEditorLoadCommand"
        ],
        [
          "svc",
          "svc",
          "call",
          "Check ownership and project editable values",
          "Full-record revision belongs to application policy.",
          "api.Service.LoadCardEditor"
        ],
        [
          "svc",
          "c",
          "msg",
          "CardEditorLoadedMsg",
          "Discard stale results before changing views.",
          "ui.Coordinator.handleSpecialMessages"
        ],
        [
          "c",
          "ed",
          "call",
          "LoadEditor(data) and choices",
          "The editable draft loads saved reasoning effort into the row immediately after Model.",
          "views.CardEditor.LoadEditor"
        ],
        [
          "c",
          "c",
          "call",
          "Navigate outside view Update",
          "",
          "ui.Coordinator.navigateTo"
        ]
      ],
      "sources": {
        "files": [],
        "refs": [
          "api.cardValues",
          "views.focusTargetTable",
          "views.CardEditor.fieldBlocks"
        ]
      }
    },
    "savecard": {
      "title": "Save explicit card edits against a revision",
      "summary": "UI submits editable draft and presence-aware changes; application reloads current data, checks owner/revision and preserves unedited fields.",
      "participants": [
        [
          "ed",
          "views.CardEditor"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "repo",
          "config.FileRepository"
        ]
      ],
      "steps": [
        [
          "ed",
          "ed",
          "call",
          "Build editable draft and changes",
          "Trim the effort value and record presence even when clearing it.",
          "views.CardEditor.Draft"
        ],
        [
          "ed",
          "c",
          "cmd",
          "Capture draft, scope and expected revision",
          "Clone draft before creating command.",
          "ui.Coordinator.cardEditorCallbacks"
        ],
        [
          "c",
          "svc",
          "call",
          "CreateEditorCard for new records",
          "",
          "api.Service.CreateEditorCard"
        ],
        [
          "c",
          "svc",
          "call",
          "UpdateEditorCard for existing records",
          "Recheck ownership and full-record revision under admission.",
          "api.Service.UpdateEditorCard"
        ],
        [
          "svc",
          "svc",
          "call",
          "Apply only explicit changes",
          "Apply the effort field only when present. A changed saved effort also invalidates stale revisions.",
          "api.applyCardChanges"
        ],
        [
          "svc",
          "repo",
          "io",
          "Persist checked current record",
          "",
          "config.FileRepository.UpdateIDCard"
        ]
      ],
      "sources": {
        "files": [],
        "refs": [
          "contracts.CardChanges",
          "contracts.CardValues",
          "config.IDCard"
        ]
      }
    },
    "source": {
      "title": "Mutate and resync a source through contracts",
      "summary": "Source manager commands capture expected identity. Application policy chooses the atomic repository operation and returns saved-with-warning outcomes.",
      "participants": [
        [
          "view",
          "views.SourceManager"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "repo",
          "config.FileRepository"
        ]
      ],
      "steps": [
        [
          "view",
          "view",
          "key",
          "Capture SourceMutation and expected row identity",
          "",
          "views.SourceManager.mutate"
        ],
        [
          "view",
          "svc",
          "cmd",
          "MutateSource(context, input)",
          "",
          "api.Service.MutateSource"
        ],
        [
          "svc",
          "repo",
          "call",
          "Update sources under repository lock",
          "Stale identity rejected; no split read/save mutation.",
          "api.updatePathSource"
        ],
        [
          "repo",
          "repo",
          "io",
          "Persist and resync registry",
          "",
          "config.FileRepository.UpdateSkillSources"
        ],
        [
          "svc",
          "view",
          "msg",
          "Structured saved/refresh/warning result",
          "Failure after saving retains the refresh requirement.",
          "views.SourceManager.Update"
        ],
        [
          "view",
          "svc",
          "cmd",
          "ReadSources for refreshed rows",
          "",
          "api.Service.ReadSources"
        ]
      ],
      "sources": {
        "files": []
      }
    },
    "mcp": {
      "title": "Create or edit and save",
      "summary": "New or a selected custom MCP opens an initialized form; saving preserves identity/metadata and returns to the explicit caller.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "m",
          "views.MCPManager"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "f",
          "views.MCPForm"
        ],
        [
          "r",
          "config.FileRepository"
        ],
        [
          "svc",
          "api.Service"
        ]
      ],
      "steps": [
        [
          "u",
          "m",
          "key",
          "Enter custom MCP or New",
          "Protected built-in generates an error instead.",
          "views.MCPManager.Update"
        ],
        [
          "m",
          "c",
          "cmd",
          "Scoped edit request",
          "Reject requests from an earlier visit.",
          "ui.Coordinator.matchesMCPManagerScope"
        ],
        [
          "c",
          "f",
          "call",
          "Reset first, then populate selected MCP",
          "Init runs synchronously before SetMCP.",
          "views.MCPForm.SetMCP"
        ],
        [
          "u",
          "f",
          "key",
          "Enter to save",
          "Form also remains the card editor create entry point.",
          "views.MCPForm.Update"
        ],
        [
          "f",
          "c",
          "call",
          "Overlay edited fields on loaded MCP",
          "Preserve description; new form starts from zero value.",
          "views.MCPForm.handleSubmit"
        ],
        [
          "c",
          "r",
          "cmd",
          "Upsert MCP; return scoped save result",
          "Command captures the configuration capability, cloned editable fields and scope/request.",
          "ui.Coordinator.mcpCallbacks"
        ],
        [
          "c",
          "svc",
          "call",
          "SaveMCP(editable contract)",
          "Application preserves noneditable metadata and enforces built-in protection before repository save.",
          "api.Service.SaveMCP"
        ],
        [
          "c",
          "m",
          "msg",
          "Return to caller and refresh data",
          "Manager return path avoids previous-view bouncing.",
          "ui.Coordinator.handleSpecialMessages"
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
          "ui.Coordinator"
        ],
        [
          "ed",
          "views.CardEditor"
        ],
        [
          "disk",
          "disk"
        ],
        [
          "svc",
          "api.Service"
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
          "ui.Coordinator.settingsCallbacks"
        ],
        [
          "coord",
          "svc",
          "cmd",
          "SaveSettings(contract)",
          "Captured configuration capability delegates locked persistence.",
          "api.Service.SaveSettings"
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
          "ui.Coordinator.handleSpecialMessages"
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
          "Current-visit successful save applies UI state; later workspace commands resolve settings inside the application operation.",
          "ui.Coordinator.navigateTo"
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
          "ui.Coordinator"
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
        ],
        [
          "svc",
          "api.Service"
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
          "ui.Coordinator.cardEditorCallbacks"
        ],
        [
          "coord",
          "tea",
          "return",
          "NavigateMsg{SourceManagerWorkspaces}",
          "Navigation occurs later in a message handler.",
          "ui.Coordinator.cardEditorCallbacks"
        ],
        [
          "tea",
          "coord",
          "msg",
          "navigateTo(workspace sources)",
          "Retain the editor instance and unsaved values.",
          "ui.Coordinator.navigateTo"
        ],
        [
          "user",
          "sm",
          "key",
          "Handle structured source mutation result",
          "Refresh saved changes and retain synchronization warnings.",
          "views.SourceManager.Update"
        ],
        [
          "tea",
          "svc",
          "cmd",
          "MutateSource(workspaces, expected identity)",
          "Application validates stable workspace-source identity and owns the locked update/sync.",
          "api.Service.MutateSource"
        ],
        [
          "repo",
          "coord",
          "msg",
          "Handle structured source mutation result",
          "Refresh saved changes and retain synchronization warnings.",
          "views.SourceManager.Update"
        ],
        [
          "user",
          "sm",
          "key",
          "Handle structured source mutation result",
          "Refresh saved changes and retain synchronization warnings.",
          "views.SourceManager.Update"
        ]
      ],
      "sources": {
        "refs": [
          "views.CardEditor.handleCtrlE",
          "views.CardEditor.renderWorkspaceLine"
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
          "ui.Coordinator"
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
          "Apply refreshed choices without clearing unavailable selected IDs",
          "Failed inventories retain prior valid choices.",
          "views.CardEditor.SetChoices"
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
          "components.CapabilityList.Deselect",
          "components.CapabilityList.Select",
          "components.CapabilityList.SelectedOnly",
          "components.CapabilityList.SetCursor",
          "components.CapabilityList.SetSelected",
          "components.CapabilityList.Toggle",
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
          "user",
          "ed",
          "key",
          "Enter submits the form",
          "The toggle is persisted only when the card is saved.",
          "views.CardEditor.handleSubmit"
        ],
        [
          "ed",
          "ed",
          "self",
          "Draft explicit editable values and changes",
          "Application preserves hidden and unedited fields during save.",
          "views.CardEditor.Draft"
        ]
      ]
    },
    "sub-picker": {
      "title": "Choose multiple sub workspaces",
      "summary": "The shared picker supports checkboxes, search, confirmation and cancellation in the card editor and initial launch form.",
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
          "fields",
          "views.workspaceFields"
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
          "fields",
          "call",
          "Open shared workspace fields",
          "Seed the shared picker from the currently selected IDs and exclude the main workspace.",
          "views.workspaceFields.open"
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
          "fields",
          "self",
          "Commit draft sub-workspace IDs",
          "Switch the shared picker back to single mode.",
          "views.workspaceFields.updatePicker"
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
          "views.WorkspacePickerModal.View",
          "views.WorkspacePickerModal.restoreSingleSelection",
          "views.WorkspacePickerModal.title",
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
      "summary": "Registry updates can remove a checked workspace while a picker is open. Shared field state keeps unavailable selections visible and a failed read retains cached choices.",
      "participants": [
        [
          "coord",
          "ui.Coordinator"
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
          "Apply refreshed choices without clearing unavailable selected IDs",
          "Failed inventories retain prior valid choices.",
          "views.CardEditor.SetChoices"
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
      "summary": "Ctrl+X clears the draft’s Sub Workspaces row. The card editor saves changes and the launch form proceeds with that selection.",
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
          "Draft explicit editable values and changes",
          "Application preserves hidden and unedited fields during save.",
          "views.CardEditor.Draft"
        ]
      ],
      "sources": {
        "refs": [
          "views.CardEditor.LoadEditor",
          "views.CardEditor.handleSpace",
          "views.WorkspacePickerModal.SetSelectedIDs",
          "views.editorKeyIdentity"
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
      "summary": "A captured command calls the application workspace operation, which resolves saved workspace and settings and revalidates scope.",
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
          "ui.Coordinator"
        ],
        [
          "act",
          "ui.sessionActions"
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
          "ui.Coordinator.dashboardCallbacks"
        ],
        [
          "coord",
          "act",
          "call",
          "Snapshot TerminalCommand by value",
          "Only the coordinator goroutine reads c.settings.",
          "ui.Coordinator.launchTerminal"
        ],
        [
          "coord",
          "act",
          "cmd",
          "launchWorkspace with captured scope/kind",
          "Frontend returns a command using independent contracts.",
          "ui.sessionActions.launchWorkspace"
        ],
        [
          "act",
          "act",
          "call",
          "Check captured dashboard/session",
          "The selected session must still belong to the captured dashboard.",
          "api.Service.CheckSessionInDashboard"
        ],
        [
          "act",
          "launch",
          "call",
          "OpenSessionWorkspace(scope, kind)",
          "Application owns settings lookup and saved-directory validation.",
          "api.Service.OpenSessionWorkspace"
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
          "ScopedActionFinishedMsg{Err}",
          "Reports process startup, not the external application’s eventual exit.",
          "ui.Coordinator.handleSpecialMessages"
        ],
        [
          "coord",
          "coord",
          "self",
          "Apply scoped completion",
          "Reject stale dashboard/generation results; display current errors without navigation.",
          "ui.Coordinator.handleSessionMessages"
        ]
      ],
      "sources": {
        "refs": [
          "messages.ScopedActionFinishedMsg",
          "views.DashboardCallbacks"
        ],
        "files": [
          "internal/ui/terminal_launch_test.go",
          "internal/ui/views/dashboard_test.go",
          "internal/session/ide_launcher.go"
        ]
      }
    },
    "workspace-command-errors": {
      "title": "Report a workspace-command launch error",
      "summary": "A captured command calls the application workspace operation, which resolves saved workspace and settings and revalidates scope.",
      "participants": [
        [
          "act",
          "ui.sessionActions"
        ],
        [
          "launch",
          "session.LaunchWorkspaceCommand",
          "Workspace launcher"
        ],
        [
          "coord",
          "ui.Coordinator"
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
          "Check scope and session",
          "Membership must still be valid when the command executes.",
          "api.Service.CheckSessionInDashboard"
        ],
        [
          "act",
          "act",
          "self",
          "launchWorkspace with captured scope/kind",
          "Frontend returns a command using independent contracts.",
          "ui.sessionActions.launchWorkspace"
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
          "ScopedActionFinishedMsg{Err}",
          "The event carries the startup error back into Update.",
          "messages.ScopedActionFinishedMsg"
        ],
        [
          "coord",
          "coord",
          "self",
          "Apply matching scoped error",
          "Only a result for the current dashboard/generation changes the banner.",
          "ui.Coordinator.handleSessionMessages"
        ],
        [
          "user",
          "coord",
          "key",
          "Next key clears the banner",
          "Existing banner lifecycle is unchanged.",
          "ui.Coordinator.Update"
        ]
      ],
      "sources": {
        "refs": [
          "messages.ScopedActionFinishedMsg"
        ],
        "files": [
          "internal/ui/terminal_launch_test.go",
          "internal/ui/command_race_internal_test.go"
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
          "ui.Coordinator"
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
          "components.DisplayText",
          "components.TruncateHead",
          "views.SettingsView.TerminalInput",
          "views.SettingsCallbacks"
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
          "ui.Coordinator"
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
          "Checked immediate attachment",
          "Membership and published process are required; the MCP timer does not delay human attachment.",
          "session.Manager.AttachChecked"
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
          "session.CallerIdentity",
          "session.Manager.activationAuthorityLocked",
          "session.Manager.publishActivationLocked",
          "session.Manager.validateReportedConversationLocked",
          "session.RegistrationResult",
          "session.activationAuthority",
          "session.addConversationBinding"
        ]
      }
    },
    "resume": {
      "title": "Choose native conversation on relaunch/switch",
      "summary": "A session retains its workspace while a human or orchestrator chooses a saved native conversation for the selected card, or starts a new one. The saved card’s driver supplies any effort arguments before the conversation arguments.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "ui",
          "ui.Coordinator"
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
        ],
        [
          "snap",
          "session.Manager"
        ]
      ],
      "steps": [
        [
          "user",
          "ui",
          "key",
          "Request relaunch or switch",
          "",
          "ui.Coordinator.requestRelaunch"
        ],
        [
          "ui",
          "svc",
          "call",
          "ReadConversationPreparationResult",
          "Application returns scoped, ordered compatibility options; no backend session snapshot enters the UI.",
          "api.Service.ReadConversationPreparationResult"
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
          "ui.Coordinator.conversationChoiceCallbacks"
        ],
        [
          "ui",
          "svc",
          "cmd",
          "RelaunchDashboardSession / SwitchSession",
          "Captured choice and scope; application revalidates bindings and publishes explicit lifecycle outcome.",
          "api.Service.SwitchSession"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Reserve and validate activation",
          "Recheck session/card availability and native binding before publication.",
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
          "Publish checked activation",
          "Preserve dashboard membership and saved workspace identity.",
          "session.Manager.SwitchAgentWithChoiceCheckedContext"
        ]
      ],
      "sources": {
        "refs": [
          "ui.Coordinator.prepareConversation",
          "ui.Coordinator.requestSwitch",
          "session.savedTemporaryRoot",
          "session.validateConversationChoice",
          "views.AgentSwitchView.SetRows",
          "views.AgentSwitchView.handleKey",
          "views.ConversationChoiceView.RefreshContext",
          "views.ConversationChoiceView.SelectedOption",
          "views.ConversationChoiceView.Update",
          "session.activationArgs"
        ]
      }
    },
    "control": {
      "title": "Authenticated discovery and session control",
      "summary": "An authenticated activation with deployed Dashboard Control manages cards and sessions in its linked dashboard. Workspace discovery reads the global registry.",
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
          "Call list_workspaces",
          "The discovery flow shows how callers resolve initial launch workspace IDs.",
          "mcp.addDiscoveryTools"
        ],
        [
          "http",
          "svc",
          "call",
          "Return registered workspaces",
          "Use the discovery IDs to build an ordered create_session request.",
          "api.Service.ListWorkspaces"
        ],
        [
          "http",
          "svc",
          "call",
          "CreateSessionWithOptions()",
          "The create request carries optional name and workspace IDs, plus a stable UUID.",
          "api.Service.CreateSessionWithOptions"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Launch in caller dashboard",
          "The saved card is rechecked before publication.",
          "api.Service.launchDashboardWithIDEvidence"
        ]
      ],
      "sources": {
        "refs": [
          "api.CardDiscovery",
          "api.Repository",
          "api.SessionDiscovery"
        ]
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
        ],
        [
          "authority",
          "api.controllerAuthority"
        ]
      ],
      "steps": [
        [
          "o",
          "svc",
          "call",
          "Submit lifecycle action with stable UUID",
          "For create_session, include the optional name and ordered workspace IDs in retry identity.",
          "api.Service.CreateSessionWithOptions"
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
          "Controller or narrowly scoped notification authority is checked in the shared runner. HTTP disconnect ends waiting, not admitted work.",
          "api.OperationRunner.execute"
        ],
        [
          "r",
          "authority",
          "call",
          "Validate controller task scope",
          "Controller mutations retain their live credential and dashboard checks. Notification tasks use their own restricted policy.",
          "api.controllerAuthority.validate"
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
        "refs": [
          "api.operationResult",
          "api.taskAuthority",
          "api.controllerAuthority.validate",
          "api.controllerAuthority.monitor"
        ]
      }
    },
    "send": {
      "title": "Send an orchestrator task through shared delivery",
      "summary": "Authenticate a controller, pin the target activation and admit one idempotent prompt.",
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
          "d",
          "api.preparedDelivery"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "r",
          "api.OperationRunner"
        ],
        [
          "s",
          "session.MessageStore"
        ]
      ],
      "steps": [
        [
          "o",
          "svc",
          "io",
          "SendMessage(UUID, target, activation)",
          "Single-line UTF-8, at most 2 KiB. Replay checks run before admission; complex instructions live in files.",
          "api.Service.SendMessage"
        ],
        [
          "svc",
          "mgr",
          "call",
          "ReserveMessageAdmission",
          "",
          "session.Manager.ReserveMessageAdmission"
        ],
        [
          "svc",
          "r",
          "call",
          "SubmitWithAdmission",
          "",
          "api.OperationRunner.SubmitWithAdmission"
        ],
        [
          "r",
          "d",
          "call",
          "publish(persistAcceptance)",
          "",
          "api.preparedDelivery.publish"
        ],
        [
          "d",
          "mgr",
          "call",
          "PublishMessageAdmission",
          "",
          "session.Manager.PublishMessageAdmission"
        ],
        [
          "mgr",
          "s",
          "io",
          "Create intent + journal acceptance",
          "",
          "session.MessageStore.CreateIntentWithAdmission"
        ],
        [
          "r",
          "d",
          "go",
          "submission() callback",
          "",
          "api.preparedDelivery.submission"
        ],
        [
          "d",
          "mgr",
          "call",
          "SubmitMessage",
          "",
          "session.Manager.SubmitMessage"
        ],
        [
          "svc",
          "o",
          "return",
          "Bounded operation result",
          "Dispatcher admission deliberately does not use this wait.",
          "api.OperationRunner.submit"
        ]
      ],
      "sources": {
        "refs": [
          "api.preparedDelivery.arguments"
        ],
        "files": [
          "internal/api/messaging_shared_delivery_test.go"
        ]
      }
    },
    "reports": {
      "title": "Save a report and wake the original sender",
      "summary": "Completed, blocked and question reports save a pending notification atomically; the worker returns without waiting for delivery.",
      "participants": [
        [
          "w",
          "agent"
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
          "s",
          "session.MessageStore"
        ],
        [
          "r",
          "api.OperationRunner"
        ]
      ],
      "steps": [
        [
          "w",
          "svc",
          "io",
          "report_reply + stable UUID",
          "",
          "mcp.addMessagingTools"
        ],
        [
          "svc",
          "mgr",
          "call",
          "ReportMessage",
          "The exact recipient activation authenticates; Dashboard Control is not required.",
          "session.Manager.ReportMessage"
        ],
        [
          "mgr",
          "s",
          "io",
          "Save report + pending obligation",
          "Completed, blocked and question start pending; identical retries keep the same reply.",
          "session.MessageStore.RecordReply"
        ],
        [
          "s",
          "s",
          "self",
          "signalLocked wakes dispatcher",
          "The same change channel also wakes existing reply waiters.",
          "session.MessageStore.signalLocked"
        ],
        [
          "svc",
          "w",
          "return",
          "Return saved reply",
          "Does not wait for prompt delivery.",
          "api.Service.ReportReply"
        ],
        [
          "r",
          "s",
          "go",
          "Read pending report snapshots",
          "",
          "session.MessageStore.Notifications"
        ]
      ],
      "sources": {
        "refs": [
          "session.MessageReply",
          "session.ReplyNotification",
          "session.PendingNotification",
          "session.NotificationState"
        ],
        "files": [
          "internal/session/messaging_test.go",
          "internal/session/messaging_review_test.go",
          "kankuro_skills/kankuro-puppet/SKILL.md"
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
          "driver.GenericCLIDriver.PrepareWorkspace",
          "driver.trustJSONObject",
          "driver.updateClaudeTrust",
          "driver.updateCodexTrust"
        ]
      }
    },
    "card-delete": {
      "title": "Confirm global deletion from the card owner",
      "summary": "Foreign current, saved and pending references require explicit human consent.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "c",
          "ui.Coordinator"
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
          "repo",
          "config.FileRepository"
        ]
      ],
      "steps": [
        [
          "u",
          "dash",
          "key",
          "Delete owned card; review impact",
          "",
          "views.Dashboard.handleDeleteCard"
        ],
        [
          "c",
          "svc",
          "call",
          "Prepare card deletion summary",
          "",
          "api.Service.PrepareHumanCardDeletion"
        ],
        [
          "svc",
          "mgr",
          "call",
          "WithCardReferencesAndSnapshot",
          "",
          "session.Manager.WithCardReferencesAndSnapshot"
        ],
        [
          "svc",
          "dash",
          "return",
          "ShowCardDeletionImpact for foreign refs",
          "",
          "views.Dashboard.ShowCardDeletionImpact"
        ],
        [
          "u",
          "svc",
          "key",
          "Confirm token and consent through the deletion capability",
          "Rechecks owner, revision, expiry and reference-set digest before effects.",
          "api.Service.ConfirmHumanCardDeletion"
        ],
        [
          "svc",
          "repo",
          "call",
          "BeginCardDeletionWithPending atomically",
          "Tombstone closes admission and retains unpublished temporary cleanup identities.",
          "config.FileRepository.BeginCardDeletionWithPending"
        ],
        [
          "svc",
          "svc",
          "call",
          "RetryCardDeletion drains then cleans",
          "",
          "api.Service.RetryCardDeletion"
        ]
      ],
      "sources": {
        "refs": [
          "api.CardDeletionPreflight",
          "api.CardDeletionResult",
          "api.cardConfirmation",
          "api.cardImpactDigest",
          "api.cardPendingCleanup",
          "api.tombstoneWriter"
        ]
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
      "title": "Recover obligations without replaying input",
      "summary": "Saved reports persist, but processes restore terminated and replacement activations inherit no wake-ups.",
      "participants": [
        [
          "mgr",
          "session.Manager"
        ],
        [
          "s",
          "session.MessageStore"
        ],
        [
          "r",
          "api.OperationRunner"
        ]
      ],
      "steps": [
        [
          "mgr",
          "mgr",
          "self",
          "Restore sessions as terminated",
          "",
          "session.Data.restore"
        ],
        [
          "mgr",
          "s",
          "call",
          "Recover unfinished sends as uncertain",
          "Existing behavior applies to notification child messages as well.",
          "session.MessageStore.Recover"
        ],
        [
          "r",
          "r",
          "io",
          "Load journal; mark interrupted operations",
          "",
          "api.OperationRunner.loadRecord"
        ],
        [
          "r",
          "s",
          "call",
          "Read retained pending/accepted reports",
          "",
          "session.MessageStore.Notifications"
        ],
        [
          "r",
          "r",
          "self",
          "Accepted: reconcile delivery evidence",
          "Recovered accepted/submitting child messages are uncertain and are not replayed.",
          "api.OperationRunner.reconcileNotification"
        ],
        [
          "r",
          "s",
          "io",
          "Never admitted + stale target: unavailable",
          "A resumed activation is not the original destination.",
          "api.OperationRunner.dispatchNotification"
        ]
      ],
      "sources": {
        "refs": []
      }
    },
    "dashboard": {
      "title": "Render owned sessions, cards and session center",
      "summary": "The dashboard keeps its three-pane layout. Center reports or workspace facts replace terminal artwork/output; session actions remain in the session list.",
      "participants": [
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "list",
          "components.SessionListItem"
        ],
        [
          "panel",
          "components.SessionCenter"
        ]
      ],
      "steps": [
        [
          "dash",
          "dash",
          "call",
          "Apply contract inventory and retain identity selection",
          "Failed inventory reads preserve valid data; successful empty snapshots clear it.",
          "views.Dashboard.SetInventory"
        ],
        [
          "list",
          "list",
          "self",
          "Render session metadata and available action hints",
          "",
          "components.SessionListItem.Description"
        ],
        [
          "dash",
          "panel",
          "call",
          "Synchronize selected stable session",
          "Changing identity clears unrelated content; activation changes do not.",
          "views.Dashboard.syncCenterSelection"
        ],
        [
          "panel",
          "panel",
          "self",
          "Render reports or workspace facts in available area",
          "Warnings, timestamps and display-column wrapping are presentation concerns.",
          "components.SessionCenter.View"
        ],
        [
          "dash",
          "dash",
          "self",
          "Keep pane alignment across sizes",
          "No live terminal output subscription.",
          "views.Dashboard.renderCenterPane"
        ]
      ],
      "sources": {
        "files": []
      }
    },
    "workspace": {
      "title": "Keep main and sub-workspace identity",
      "summary": "Initial launch prepares the selected main and sub-workspaces. Relaunch uses the saved paths and deployed copies.",
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
          "Launch with an effective card",
          "The current initial launch selection supplies the main and sub-workspace registry IDs.",
          "api.Service.launchDashboardWithIDEvidence"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Prepare main directory",
          "The card controls whether to start in place or copy into a session directory.",
          "session.Manager.prepareWorkspace"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "Prepare selected copies",
          "Serialize copy ownership and collect this launch’s descriptors under the same gate.",
          "session.Manager.prepareSubWorkspacesSnapshot"
        ],
        [
          "mgr",
          "disk",
          "io",
          "Stage and publish copies",
          "Reuse owned copies and retain shared manifest entries.",
          "session.copyAndPublishSubWorkspace"
        ],
        [
          "mgr",
          "svc",
          "return",
          "Publish workspace identity",
          "Save the original and working paths and the launch’s own sub-workspace descriptors.",
          "session.Manager.startSessionCore"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Relaunch with saved paths",
          "Reuse the saved working directory and the deployed copy information.",
          "session.Manager.relaunchWithChoiceContext"
        ],
        [
          "svc",
          "svc",
          "self",
          "Inspect actual paths",
          "Session inspection reports recorded paths and descriptors.",
          "api.Service.GetSession"
        ]
      ],
      "sources": {
        "refs": [
          "session.Manager.RelaunchWithChoiceCheckedContext",
          "session.Manager.cleanupNewTemporaryWorkspace",
          "session.Manager.resolveWorkspacePath",
          "session.preparedSubWorkspaceDescriptors"
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
          "session.activationEnv",
          "session.syncActivationMCPConfig"
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
          "config.BuiltinMCPs",
          "config.BuiltinSkills",
          "config.GetBuiltinSkillContent",
          "driver.writeInstructionFile"
        ]
      }
    },
    "dashboard-details": {
      "title": "Create or edit a dashboard",
      "summary": "The initial list owns name/description forms and immutable selection IDs.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "list",
          "views.DashboardList"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "repo",
          "config.FileRepository"
        ]
      ],
      "steps": [
        [
          "u",
          "list",
          "key",
          "n / e; edit name and description",
          "",
          "views.DashboardList.Update"
        ],
        [
          "list",
          "list",
          "call",
          "updateForm: validate nonempty name",
          "",
          "views.DashboardList.updateForm"
        ],
        [
          "list",
          "c",
          "cmd",
          "Save callback captures view generation",
          "",
          "ui.Coordinator.dashboardListCallbacks"
        ],
        [
          "c",
          "svc",
          "call",
          "CreateDashboard / UpdateDashboardDetails",
          "",
          "api.Service.CreateDashboard"
        ],
        [
          "svc",
          "repo",
          "call",
          "UpdateDashboardDetails: active owner only",
          "",
          "config.FileRepository.UpdateDashboardDetails"
        ],
        [
          "c",
          "svc",
          "call",
          "ListDashboards includes recovery entries",
          "",
          "api.Service.ListDashboards"
        ],
        [
          "c",
          "list",
          "call",
          "SetDashboards preserves selected ID",
          "",
          "views.DashboardList.SetDashboards"
        ]
      ],
      "sources": {
        "refs": [
          "config.Dashboard",
          "config.FileRepository.CreateDashboard",
          "config.FileRepository.ListDashboards",
          "config.FileRepository.UpdateDashboardDetails",
          "config.FileRepository.loadDashboardsLocked",
          "config.dashboardsFile",
          "config.validateDashboard"
        ]
      }
    },
    "navigation": {
      "title": "Open a dashboard and return to the list",
      "summary": "Enter opens its card/session view; Esc returns without stopping sessions. Changing dashboard scope also closes any rename draft, so it cannot leak into a reopened dashboard.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "list",
          "views.DashboardList"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "svc",
          "api.Service"
        ]
      ],
      "steps": [
        [
          "u",
          "list",
          "key",
          "Enter on active dashboard",
          "",
          "views.DashboardList.Update"
        ],
        [
          "list",
          "c",
          "msg",
          "DashboardOpenRequestedMsg",
          "Captures dashboard identity and increments view generation.",
          "ui.Coordinator.handleSpecialMessages"
        ],
        [
          "c",
          "dash",
          "call",
          "SetDashboardScope and cached view",
          "",
          "views.Dashboard.SetDashboardScope"
        ],
        [
          "c",
          "svc",
          "call",
          "ReadDashboardOverview(openedID)",
          "",
          "api.Service.ReadDashboardOverview"
        ],
        [
          "u",
          "dash",
          "key",
          "Esc requests dashboard list",
          "",
          "views.Dashboard.handleKey"
        ],
        [
          "dash",
          "c",
          "msg",
          "DashboardListRequestedMsg",
          "Drops the center subscription and invalidates old scoped results.",
          "ui.Coordinator.handleSpecialMessages"
        ],
        [
          "u",
          "list",
          "key",
          "Esc stays; q quits",
          "Default Bubbles quit bindings are disabled. Forms and confirmations retain Escape cancellation.",
          "views.DashboardList.Update"
        ]
      ],
      "sources": {
        "refs": [
          "ui.Coordinator.Init",
          "ui.Coordinator.dashboardListCallbacks",
          "ui.Coordinator.guardKey",
          "ui.Coordinator.handleSessionMessages",
          "ui.Coordinator.prepareConversation",
          "ui.Coordinator.handleSessionMessages",
          "ui.Coordinator.handleSpecialMessages",
          "ui.Coordinator.navigateTo",
          "ui.Coordinator.refreshDashboardList",
          "ui.Coordinator.resizeAll",
          "ui.NewCoordinator",
          "ui.cardDeletionPreflightCommand",
          "ui.confirmCardDeletionCommand",
          "ui.convertViewType",
          "ui.dashboardDeletionStatus",
          "ui.dashboardDeletionSummary",
          "ui.dashboardListCommand",
          "ui.retryCardDeletionCommand",
          "views.Dashboard.SetDashboardName",
          "views.Dashboard.SetDashboardScope",
          "views.Dashboard.SetSize",
          "views.Dashboard.ShowCardDeletionImpact",
          "views.Dashboard.Update",
          "views.Dashboard.View",
          "views.Dashboard.footerText",
          "views.Dashboard.handleDeleteCard",
          "views.Dashboard.handleEdit",
          "views.Dashboard.handleEnter",
          "views.Dashboard.handleKey",
          "views.Dashboard.handleSessionEnter",
          "views.Dashboard.handleSwitch",
          "views.Dashboard.renderBase",
          "views.Dashboard.updateCardImpact",
          "views.DashboardList",
          "views.DashboardList.Init",
          "views.DashboardList.SetDashboards",
          "views.DashboardList.SetSize",
          "views.DashboardList.Update",
          "views.DashboardList.View",
          "views.DashboardList.confirmDeletion",
          "views.DashboardList.openForm",
          "views.DashboardList.renderPage",
          "views.DashboardList.retryDeletion",
          "views.DashboardList.selected",
          "views.DashboardList.updateConfirmation",
          "views.DashboardList.updateForm",
          "views.DashboardListCallbacks",
          "views.NewDashboardList",
          "views.dashboardItem",
          "views.dashboardItem.Description",
          "views.dashboardItem.FilterValue",
          "views.dashboardItem.Title",
          "views.dashboardListMode"
        ]
      }
    },
    "caller-scope": {
      "title": "Derive MCP scope from the live caller",
      "summary": "A shared card never shares the sessions launched from it.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "mcp",
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
          "scope",
          "api.persistedDashboardScope"
        ]
      ],
      "steps": [
        [
          "u",
          "mcp",
          "io",
          "list_cards / list_sessions / get_session",
          "",
          "mcp.addDiscoveryTools"
        ],
        [
          "mcp",
          "svc",
          "call",
          "ResolveCaller(token)",
          "",
          "api.Service.ResolveCaller"
        ],
        [
          "svc",
          "mgr",
          "call",
          "ResolveDashboardCaller(token)",
          "",
          "session.Manager.ResolveDashboardCaller"
        ],
        [
          "svc",
          "scope",
          "call",
          "DashboardForSession: persisted membership",
          "",
          "api.persistedDashboardScope.DashboardForSession"
        ],
        [
          "svc",
          "scope",
          "call",
          "AllowsCard: owner match or shared",
          "",
          "api.persistedDashboardScope.AllowsCard"
        ],
        [
          "svc",
          "scope",
          "call",
          "AllowsSession: exact session dashboard",
          "",
          "api.persistedDashboardScope.AllowsSession"
        ],
        [
          "svc",
          "mcp",
          "return",
          "Discovery includes dashboard/card state",
          "Human navigation is absent from caller authority.",
          "api.Service.ListSessions"
        ]
      ],
      "sources": {
        "refs": [
          "contracts.DashboardInventory",
          "api.New",
          "api.Service.GetSession",
          "api.Service.ListSessions",
          "api.Service.ResolveCaller",
          "api.Service.authorizeExistingOperation",
          "api.Service.deletedCardIDs",
          "api.Service.inspectionFromScopedSnapshot",
          "api.Service.listCardsForCaller",
          "api.dashboardWriter",
          "api.hasForeignCardReferences",
          "api.inspectionFromSnapshot",
          "api.membershipInventory",
          "api.membershipInventory.activeDashboard",
          "api.membershipInventory.allowsCard",
          "api.membershipInventory.allowsSession",
          "api.membershipInventory.hasTombstone",
          "api.newMembershipInventory",
          "api.persistedDashboardScope",
          "api.persistedDashboardScope.AllowsCard",
          "api.persistedDashboardScope.AllowsSession",
          "api.persistedDashboardScope.DashboardForSession",
          "api.persistedDashboardScope.SessionDashboard",
          "api.persistedDashboardScope.inventory",
          "api.persistedDashboardScope.validateScopeInventory",
          "api.scopedCardWriter",
          "api.sessionInDashboard",
          "api.sessionUsesCard",
          "session.Manager.List",
          "session.Manager.SetNameDetailedChecked",
          "session.Manager.ShutdownStarted",
          "session.Manager.listLocked",
          "session.NewManager",
          "session.Session.data",
          "session.Session.snapshot"
        ]
      }
    },
    "manual-session": {
      "title": "Attach, name a launch, stop or remove an owned session",
      "summary": "Manual session actions use shared dashboard policy. Launch-time naming can attach; existing-session rename has a separate scoped result and is traced in the rename flows.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "c",
          "ui.Coordinator"
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
          "launch",
          "views.SessionLauncher"
        ]
      ],
      "steps": [
        [
          "u",
          "c",
          "key",
          "Attach selected session",
          "",
          "ui.Coordinator.requestAttach"
        ],
        [
          "c",
          "svc",
          "call",
          "AttachInDashboard checks at subscription",
          "",
          "api.Service.AttachInDashboard"
        ],
        [
          "svc",
          "mgr",
          "call",
          "AttachChecked returns handle",
          "",
          "session.Manager.AttachChecked"
        ],
        [
          "c",
          "c",
          "call",
          "ScopedAttachReadyMsg rejects stale handles",
          "",
          "ui.Coordinator.handleSessionMessages"
        ],
        [
          "launch",
          "svc",
          "call",
          "RenameDashboardSession",
          "Launcher naming still attaches after success. Existing-session e rename uses DashboardSessionRenamedMsg and only refreshes.",
          "api.Service.RenameDashboardSession"
        ],
        [
          "c",
          "svc",
          "call",
          "StopDashboardSession validates owned target",
          "",
          "api.Service.StopDashboardSession"
        ],
        [
          "c",
          "svc",
          "call",
          "RemoveDashboardSession retains removal guards",
          "",
          "api.Service.RemoveDashboardSession"
        ]
      ],
      "sources": {
        "refs": [
          "ui.Coordinator.requestAttach",
          "ui.scopedAttachCommand",
          "ui.sessionActions.launch",
          "ui.sessionActions.launchWorkspace",
          "ui.sessionActions.launchWorkspace",
          "ui.sessionActions.launchWorkspace",
          "ui.sessionActions.activate",
          "ui.sessionActions.activate",
          "session.Manager.AttachChecked",
          "views.SessionLauncher.Reset",
          "views.SessionLauncher.SetNameResult",
          "views.SessionLauncher.handleNameSubmit"
        ]
      }
    },
    "card-recovery": {
      "title": "Retry failed unpublished-workspace cleanup",
      "summary": "A cancelled launch can fail to remove its temporary copy or trust entries.",
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
          "svc",
          "mgr",
          "call",
          "ReserveCardDeletion prevents parallel retry",
          "",
          "session.Manager.ReserveCardDeletion"
        ],
        [
          "svc",
          "mgr",
          "call",
          "CancelPendingCardReferences; wait",
          "",
          "session.Manager.CancelPendingCardReferences"
        ],
        [
          "svc",
          "mgr",
          "call",
          "WaitCardReferences drains launch cleanup",
          "",
          "session.Manager.WaitCardReferences"
        ],
        [
          "svc",
          "svc",
          "call",
          "cleanupCardPending reads durable paths",
          "",
          "api.Service.cleanupCardPending"
        ],
        [
          "svc",
          "mgr",
          "call",
          "RemoveResidual at the recorded path",
          "Does not touch workspaces belonging to already-published sessions.",
          "session.Manager.RemoveResidual"
        ],
        [
          "svc",
          "repo",
          "call",
          "SaveCardDeletionPending(error or cleaned)",
          "",
          "config.FileRepository.SaveCardDeletionPending"
        ],
        [
          "svc",
          "repo",
          "call",
          "FinishCardDeletion only after all cleaned",
          "",
          "config.FileRepository.FinishCardDeletion"
        ],
        [
          "repo",
          "disk",
          "io",
          "Keep non-executable tombstone for history",
          "",
          "config.CardTombstone"
        ]
      ],
      "sources": {
        "refs": [
          "api.Service.ConfirmCardDeletion",
          "api.Service.PreflightCardDeletion",
          "api.Service.RetryCardDeletion",
          "api.Service.cardOwnedBy",
          "api.Service.cleanupCardPending",
          "api.Service.validateCardConfirmationLocked",
          "api.Service.validateDeletionInventory",
          "api.Service.validateDeletionOwner",
          "api.cardImpactDigest",
          "api.cardPendingCleanup",
          "config.CardDeletionPending",
          "config.CardTombstone",
          "config.FileRepository.BeginCardDeletion",
          "config.FileRepository.BeginCardDeletionWithPending",
          "config.FileRepository.FinishCardDeletion",
          "config.FileRepository.ListCardTombstones",
          "config.FileRepository.SaveCardDeletionPending",
          "config.FileRepository.checkCardNotTombstonedLocked",
          "config.FileRepository.loadCardTombstonesLocked",
          "config.cardTombstonesFile",
          "config.validateNewCardPending",
          "session.Manager.RemoveResidual"
        ]
      }
    },
    "deleted-session": {
      "title": "Continue using a session after its card is deleted",
      "summary": "Existing foreign activations survive; a stopped session can switch to a different card.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "mgr",
          "session.Manager"
        ]
      ],
      "steps": [
        [
          "svc",
          "c",
          "return",
          "ReadDashboardData includes CardDeleted",
          "",
          "api.Service.ReadDashboardOverview"
        ],
        [
          "c",
          "dash",
          "call",
          "Session description names deleted card",
          "",
          "components.SessionListItem.Description"
        ],
        [
          "u",
          "dash",
          "key",
          "Enter blocks deleted-card relaunch",
          "",
          "views.Dashboard.handleSessionEnter"
        ],
        [
          "u",
          "svc",
          "key",
          "Stop and select a different available card",
          "",
          "api.Service.SwitchInDashboard"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Switch preserves session dashboard and files",
          "",
          "session.Manager.SwitchAgentWithChoiceCheckedContext"
        ],
        [
          "mgr",
          "mgr",
          "call",
          "Data.restore keeps saved bindings on restart",
          "",
          "session.Data.restore"
        ]
      ],
      "sources": {
        "refs": [
          "components.SessionListItem.Description"
        ]
      }
    },
    "delete-dashboard": {
      "title": "Delete owned sessions and cards as a recoverable cascade",
      "summary": "Deleting the owner also deletes its shared cards globally after foreign-impact confirmation.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "list",
          "views.DashboardList"
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
          "repo",
          "config.FileRepository"
        ]
      ],
      "steps": [
        [
          "u",
          "list",
          "key",
          "Delete dashboard; review cascade summary",
          "",
          "views.DashboardList.Update"
        ],
        [
          "list",
          "svc",
          "call",
          "PreflightDashboardDeletion",
          "",
          "api.Service.PreflightDashboardDeletion"
        ],
        [
          "svc",
          "svc",
          "call",
          "prepareDashboardDeletion captures all effects",
          "",
          "api.Service.prepareDashboardDeletion"
        ],
        [
          "u",
          "svc",
          "key",
          "ConfirmDashboardDeletion with consent",
          "",
          "api.Service.ConfirmDashboardDeletion"
        ],
        [
          "svc",
          "repo",
          "call",
          "BeginDashboardDeletion before teardown",
          "",
          "config.FileRepository.BeginDashboardDeletion"
        ],
        [
          "svc",
          "mgr",
          "call",
          "CancelDashboardAdmissions for owned scope",
          "",
          "session.Manager.CancelDashboardAdmissions"
        ],
        [
          "svc",
          "mgr",
          "call",
          "WaitDashboardAdmissions without locks",
          "",
          "session.Manager.WaitDashboardAdmissions"
        ],
        [
          "svc",
          "svc",
          "call",
          "runDashboardDeletion applies checkpoints",
          "",
          "api.Service.runDashboardDeletion"
        ]
      ],
      "sources": {
        "refs": [
          "api.DashboardDeletionPreflight",
          "api.Service.PreflightDashboardDeletion",
          "api.Service.prepareDashboardDeletion",
          "api.dashboardConfirmation",
          "api.dashboardImpactDigest",
          "ui.Coordinator.dashboardListCallbacks",
          "ui.Coordinator.refreshDashboardList",
          "ui.cardDeletionPreflightCommand",
          "ui.confirmCardDeletionCommand",
          "ui.dashboardDeletionStatus",
          "ui.dashboardDeletionSummary",
          "ui.dashboardListCommand",
          "ui.retryCardDeletionCommand",
          "config.DashboardTombstone",
          "config.dashboardTombstonesFile",
          "session.AdmissionSnapshot",
          "session.Manager.CancelDashboardAdmissions",
          "session.Manager.ReserveCardDeletion",
          "session.Manager.ReserveDashboardDeletion",
          "session.Manager.WaitDashboardAdmissions",
          "session.Manager.WithAdmissionSnapshot"
        ]
      }
    },
    "cascade-recovery": {
      "title": "Recover partial dashboard cleanup after restart",
      "summary": "The durable manifest survives removed memory rows and failed writes. Delete or Backspace resumes its checkpointed effects.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "list",
          "views.DashboardList"
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
          "repo",
          "config.FileRepository"
        ],
        [
          "store",
          "session.MessageStore"
        ]
      ],
      "steps": [
        [
          "u",
          "list",
          "key",
          "Delete / Backspace retries",
          "The dashboard is already deleting, so resume its durable manifest without repeating initial confirmation.",
          "views.DashboardList.Update"
        ],
        [
          "list",
          "list",
          "call",
          "Clear error; resume deletion",
          "",
          "views.DashboardList.retryDeletion"
        ],
        [
          "list",
          "svc",
          "call",
          "RetryDashboardDeletion loads manifest",
          "",
          "api.Service.RetryDashboardDeletion"
        ],
        [
          "svc",
          "svc",
          "call",
          "removeDashboardSession checks identity",
          "",
          "api.Service.removeDashboardSession"
        ],
        [
          "svc",
          "mgr",
          "call",
          "RemoveDetailed after process teardown",
          "",
          "session.Manager.RemoveDetailed"
        ],
        [
          "svc",
          "mgr",
          "call",
          "RemoveResidual if row already disappeared",
          "",
          "session.Manager.RemoveResidual"
        ],
        [
          "svc",
          "repo",
          "call",
          "SaveDashboardDeletion after each effect",
          "",
          "config.FileRepository.SaveDashboardDeletion"
        ],
        [
          "svc",
          "svc",
          "call",
          "deleteDashboardCard retains tombstone",
          "",
          "api.Service.deleteDashboardCard"
        ],
        [
          "svc",
          "store",
          "call",
          "DeleteDashboard removes scoped history",
          "",
          "session.MessageStore.DeleteDashboard"
        ],
        [
          "svc",
          "repo",
          "call",
          "FinalizeDashboardDeletion removes record last",
          "",
          "config.FileRepository.FinalizeDashboardDeletion"
        ]
      ],
      "sources": {
        "refs": [
          "api.DashboardDeletionResult",
          "api.Service.ConfirmDashboardDeletion",
          "api.Service.GetDashboardDeletion",
          "api.Service.RetryDashboardDeletion",
          "api.Service.dashboardAlreadyDeleted",
          "api.Service.dashboardDeletion",
          "api.Service.deleteDashboardCard",
          "api.Service.removeDashboardSession",
          "api.Service.runDashboardDeletion",
          "api.dashboardDeletionWriter",
          "config.DashboardDeletion",
          "config.DashboardDeletionCard",
          "config.DashboardDeletionPending",
          "config.DashboardDeletionReference",
          "config.DashboardDeletionSession",
          "config.FileRepository.BeginDashboardDeletion",
          "config.FileRepository.FinalizeDashboardDeletion",
          "config.FileRepository.ListDashboardDeletions",
          "config.FileRepository.ListDashboardTombstones",
          "config.FileRepository.SaveDashboardDeletion",
          "config.FileRepository.listDashboardDeletionsLocked",
          "config.FileRepository.loadDashboardTombstonesLocked",
          "config.FileRepository.retainDashboardTombstoneLocked",
          "config.FileRepository.writeDashboardDeletionLocked",
          "config.dashboardDeletionFilename",
          "config.validTemporaryIdentity",
          "config.validateDashboardDeletion"
        ]
      }
    },
    "list-errors": {
      "title": "Show deletion preparation and recovery errors",
      "summary": "Preflight or status-inspection errors must explain why a list action failed.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "list",
          "views.DashboardList"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "svc",
          "api.Service"
        ]
      ],
      "steps": [
        [
          "u",
          "list",
          "key",
          "Delete active / Enter deleting dashboard",
          "",
          "views.DashboardList.Update"
        ],
        [
          "list",
          "c",
          "cmd",
          "PreflightDelete / InspectDelete command",
          "",
          "ui.Coordinator.dashboardListCallbacks"
        ],
        [
          "c",
          "svc",
          "call",
          "GetDashboardDeletion or preflight",
          "",
          "api.Service.GetDashboardDeletion"
        ],
        [
          "c",
          "list",
          "call",
          "Result sets safe visible error in list mode",
          "",
          "views.DashboardList.Update"
        ],
        [
          "list",
          "list",
          "call",
          "View clips and sanitizes error text",
          "Errors are visible even when the refreshed list becomes empty.",
          "views.DashboardList.View"
        ],
        [
          "u",
          "list",
          "key",
          "Delete retries after an error",
          "Preparation failures restart preflight; partially completed deletions resume their durable manifest.",
          "views.DashboardList.Update"
        ]
      ],
      "sources": {
        "refs": [
          "views.DashboardList.Init",
          "views.DashboardList.SetDashboards",
          "views.DashboardList.SetSize",
          "views.DashboardList.Update",
          "views.DashboardList.View",
          "views.DashboardList.confirmDeletion",
          "views.DashboardList.openForm",
          "views.DashboardList.retryDeletion",
          "views.DashboardList.selected",
          "views.DashboardList.updateConfirmation",
          "views.DashboardList.updateForm",
          "views.NewDashboardList",
          "views.dashboardItem.Description",
          "views.dashboardItem.FilterValue",
          "views.dashboardItem.Title"
        ]
      }
    },
    "report-progress": {
      "title": "Progress remains silent",
      "summary": "Progress saves a correlated report with not_required notification status. The worker continues working; progress sends no wake-up.",
      "participants": [
        [
          "w",
          "agent"
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
          "s",
          "session.MessageStore"
        ]
      ],
      "steps": [
        [
          "w",
          "svc",
          "io",
          "report_reply(kind=progress)",
          "",
          "api.Service.ReportReply"
        ],
        [
          "svc",
          "mgr",
          "call",
          "Authenticate report",
          "",
          "session.Manager.ReportMessage"
        ],
        [
          "mgr",
          "s",
          "io",
          "Save progress + not_required",
          "",
          "session.MessageStore.RecordReply"
        ],
        [
          "s",
          "s",
          "self",
          "Exclude from dispatcher snapshots",
          "Reports with not_required status and older reports without notification metadata are not dispatched.",
          "session.MessageStore.Notifications"
        ]
      ],
      "sources": {
        "refs": []
      }
    },
    "notification-dispatch": {
      "title": "Dispatch a saved report without worker involvement",
      "summary": "One owned dispatcher schedules pending reports and feeds the existing bounded operation queue.",
      "participants": [
        [
          "svc",
          "api.Service"
        ],
        [
          "r",
          "api.OperationRunner"
        ],
        [
          "s",
          "session.MessageStore"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "d",
          "api.preparedDelivery"
        ]
      ],
      "steps": [
        [
          "svc",
          "r",
          "call",
          "SetOperationRunner starts dispatcher",
          "",
          "api.Service.SetOperationRunner"
        ],
        [
          "r",
          "r",
          "go",
          "startDispatcher",
          "",
          "api.OperationRunner.startDispatcher"
        ],
        [
          "r",
          "s",
          "call",
          "Notifications in report order",
          "",
          "session.MessageStore.Notifications"
        ],
        [
          "r",
          "r",
          "self",
          "Select due recipients fairly",
          "Earlier pending/accepted reports retain order per recipient; another recipient can proceed.",
          "api.dueNotifications"
        ],
        [
          "r",
          "s",
          "io",
          "Persist deterministic attempt UUID",
          "",
          "session.MessageStore.UpdateNotification"
        ],
        [
          "r",
          "mgr",
          "call",
          "ReserveReportNotification",
          "Destination comes from original message, never from reporter input.",
          "session.Manager.ReserveReportNotification"
        ],
        [
          "r",
          "d",
          "call",
          "Build delivery arguments",
          "",
          "api.preparedDelivery.arguments"
        ],
        [
          "r",
          "r",
          "call",
          "accept without waiting",
          "Reuse four workers and the existing bounded queue.",
          "api.OperationRunner.accept"
        ],
        [
          "r",
          "d",
          "call",
          "publish history + operation",
          "",
          "api.preparedDelivery.publish"
        ],
        [
          "d",
          "s",
          "io",
          "createIntent links parent to child",
          "Accepted state and message ID are saved with child intent; failed journal admission rolls history back.",
          "session.MessageStore.createIntent"
        ],
        [
          "r",
          "s",
          "io",
          "Record operation ID",
          "",
          "session.MessageStore.UpdateNotification"
        ]
      ],
      "sources": {
        "refs": [
          "api.OperationRunner.dispatchNotification",
          "api.notificationSchedule"
        ]
      }
    },
    "notification-authority": {
      "title": "Authorize one report notification",
      "summary": "A retained report permits delivery only to its original sender activation; worker Dashboard Control is not required.",
      "participants": [
        [
          "r",
          "api.OperationRunner"
        ],
        [
          "n",
          "api.notificationAuthority"
        ],
        [
          "s",
          "session.MessageStore"
        ],
        [
          "mgr",
          "session.Manager"
        ]
      ],
      "steps": [
        [
          "r",
          "n",
          "call",
          "Validate notification authority",
          "Only the report-derived destination is permitted.",
          "api.notificationAuthority.validate"
        ],
        [
          "n",
          "s",
          "call",
          "Find retained pending/accepted report",
          "",
          "session.MessageStore.Notifications"
        ],
        [
          "n",
          "mgr",
          "call",
          "Check exact target activation",
          "Dashboard membership and message availability must still match.",
          "session.Manager.GetSession"
        ],
        [
          "r",
          "r",
          "go",
          "Start authorization monitor",
          "The shared runner monitors authority concurrently with task execution.",
          "api.OperationRunner.monitorCaller"
        ],
        [
          "r",
          "n",
          "call",
          "Revalidate on monitor ticks",
          "Runs in the monitor goroutine while submission proceeds.",
          "api.notificationAuthority.monitor"
        ],
        [
          "r",
          "r",
          "call",
          "Apply the admitted task",
          "Execution and cancellation use the same worker loop as controller tasks.",
          "api.OperationRunner.execute"
        ]
      ],
      "sources": {
        "refs": [
          "api.taskAuthority",
          "api.operationTask",
          "session.Manager.reportDashboardAvailableLocked",
          "session.MessageStore.notification",
          "session.Manager.reportAdmissionValidLocked"
        ]
      }
    },
    "notification-write": {
      "title": "Submit a notification that asks for no reply",
      "summary": "A typed prompt carries correlation IDs, asks for get_replies and requires no report_reply; the report body stays in history.",
      "participants": [
        [
          "r",
          "api.OperationRunner"
        ],
        [
          "d",
          "api.preparedDelivery"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "s",
          "session.MessageStore"
        ],
        [
          "drv",
          "driver.GenericCLIDriver"
        ],
        [
          "o",
          "orchestrator"
        ]
      ],
      "steps": [
        [
          "r",
          "d",
          "go",
          "Execute shared submission",
          "",
          "api.preparedDelivery.submission"
        ],
        [
          "d",
          "mgr",
          "call",
          "SubmitMessage",
          "",
          "session.Manager.SubmitMessage"
        ],
        [
          "mgr",
          "s",
          "call",
          "Wait existing FIFO turn",
          "",
          "session.MessageStore.WaitTurn"
        ],
        [
          "mgr",
          "mgr",
          "self",
          "beginMessageSubmission",
          "",
          "session.Manager.beginMessageSubmission"
        ],
        [
          "mgr",
          "drv",
          "call",
          "SubmitPrompt(PromptEnvelope)",
          "",
          "driver.GenericCLIDriver.SubmitPrompt"
        ],
        [
          "drv",
          "drv",
          "self",
          "formatPrompt(report_notification)",
          "Original message ID, reply ID, reporter, kind; get_replies; no report_reply required.",
          "driver.formatPrompt"
        ],
        [
          "drv",
          "mgr",
          "call",
          "BeforeWrite revalidates after input gate",
          "",
          "session.Manager.SubmitMessage"
        ],
        [
          "drv",
          "o",
          "io",
          "Prompt then Enter",
          "Ordinary input serialization remains; Codex/Claude retain their cancellable 500 ms pause.",
          "driver.GenericCLIDriver.submitPrompt"
        ],
        [
          "mgr",
          "s",
          "io",
          "Persist delivery evidence",
          "",
          "session.MessageStore.UpdateDelivery"
        ]
      ],
      "sources": {
        "refs": [
          "driver.PromptEnvelope",
          "driver.AgentDriver",
          "driver.CodexCLIDriver.SubmitPrompt",
          "driver.ClaudeCodeDriver.SubmitPrompt"
        ],
        "files": [
          "internal/driver/driver_test.go",
          "internal/mcp/server_test.go"
        ]
      }
    },
    "notification-retry": {
      "title": "Retry admission and proven zero-byte failures",
      "summary": "Retry temporary admission failures automatically; create a new attempt only after conclusive zero-byte delivery.",
      "participants": [
        [
          "r",
          "api.OperationRunner"
        ],
        [
          "s",
          "session.MessageStore"
        ],
        [
          "d",
          "api.preparedDelivery"
        ]
      ],
      "steps": [
        [
          "r",
          "r",
          "self",
          "Queue full or temporary admission error",
          "",
          "api.OperationRunner.dispatchNotification"
        ],
        [
          "r",
          "d",
          "call",
          "Release unpublished reservation",
          "",
          "api.preparedDelivery.releaseUnpublished"
        ],
        [
          "r",
          "r",
          "self",
          "Backoff: 1,2,4,8,16,30 seconds",
          "",
          "api.nextNotificationDelay"
        ],
        [
          "r",
          "r",
          "go",
          "Retry the same attempt identity",
          "",
          "api.OperationRunner.dispatchNotifications"
        ],
        [
          "r",
          "r",
          "self",
          "Reconcile accepted attempt",
          "",
          "api.OperationRunner.reconcileNotification"
        ],
        [
          "r",
          "s",
          "call",
          "GetMessage includes runtime evidence",
          "",
          "session.MessageStore.GetMessage"
        ],
        [
          "r",
          "s",
          "io",
          "Not submitted: advance attempt atomically",
          "Only conclusive zero-byte outcome permits another delivery identity.",
          "session.MessageStore.UpdateNotification"
        ],
        [
          "r",
          "s",
          "io",
          "Submitted/uncertain: terminal status",
          "Never replay partial input or rely on a missing tool response as proof of zero bytes.",
          "session.MessageStore.UpdateNotification"
        ]
      ],
      "sources": {
        "refs": [],
        "files": [
          "internal/api/messaging_review_test.go"
        ]
      }
    },
    "report-inspection": {
      "title": "Read full reports and notification status",
      "summary": "Read the full report and copied notification status through existing bounded, authorized history APIs.",
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
          "s",
          "session.MessageStore"
        ]
      ],
      "steps": [
        [
          "o",
          "svc",
          "io",
          "get_replies(original message ID)",
          "",
          "mcp.addMessagingTools"
        ],
        [
          "svc",
          "s",
          "call",
          "GetReplies bounded page",
          "",
          "session.MessageStore.GetReplies"
        ],
        [
          "s",
          "s",
          "self",
          "Clone reply + notification metadata",
          "",
          "session.cloneReply"
        ],
        [
          "svc",
          "o",
          "return",
          "Report text, reply ID and delivery status",
          "Use reply IDs and pagination; submission is not proof of processing.",
          "api.Service.GetReplies"
        ],
        [
          "o",
          "svc",
          "io",
          "Optional get_message inspection",
          "",
          "api.Service.GetMessage"
        ],
        [
          "svc",
          "s",
          "call",
          "Copied message and replies",
          "",
          "session.MessageStore.GetMessage"
        ]
      ],
      "sources": {
        "refs": []
      }
    },
    "notification-cancel": {
      "title": "Stop notification dispatch and cancel queued delivery",
      "summary": "Shutdown joins the dispatcher before closing the shared queue; common cleanup releases canceled reservations.",
      "participants": [
        [
          "r",
          "api.OperationRunner"
        ],
        [
          "d",
          "api.preparedDelivery"
        ],
        [
          "s",
          "session.MessageStore"
        ],
        [
          "mgr",
          "session.Manager"
        ]
      ],
      "steps": [
        [
          "r",
          "r",
          "self",
          "Cancel and join dispatcher first",
          "",
          "api.OperationRunner.Shutdown"
        ],
        [
          "r",
          "r",
          "self",
          "Cancel queued/running operations",
          "",
          "api.OperationRunner.Shutdown"
        ],
        [
          "r",
          "d",
          "call",
          "Shared cancel callback",
          "",
          "api.preparedDelivery.cancel"
        ],
        [
          "d",
          "s",
          "io",
          "Mark not_submitted before write",
          "",
          "session.MessageStore.UpdateDelivery"
        ],
        [
          "d",
          "mgr",
          "call",
          "Release reservation",
          "",
          "session.MessageAdmission.Release"
        ],
        [
          "r",
          "r",
          "self",
          "Join operation workers",
          "",
          "api.OperationRunner.Shutdown"
        ]
      ],
      "sources": {
        "refs": []
      }
    },
    "rename-edit": {
      "title": "Open, edit or cancel a session name",
      "summary": "Press e with a session selected; the focused input owns typing until save or cancel.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "dash",
          "views.Dashboard"
        ]
      ],
      "steps": [
        [
          "user",
          "dash",
          "key",
          "Update(e)",
          "Search consumes e before shortcut handling.",
          "views.Dashboard.Update"
        ],
        [
          "dash",
          "dash",
          "call",
          "handleEdit: selected session",
          "Capture immutable ID; prefill the current name; reuse modal and text input.",
          "views.Dashboard.handleEdit"
        ],
        [
          "user",
          "dash",
          "key",
          "updateRename: type or Esc",
          "Typing stays inside the dialog. Esc cancels before submission.",
          "views.Dashboard.updateRename"
        ],
        [
          "dash",
          "dash",
          "call",
          "closeRename",
          "Cancellation clears the target, pending state and previous error.",
          "views.Dashboard.closeRename"
        ],
        [
          "dash",
          "dash",
          "call",
          "SetDashboardScope",
          "A changed dashboard or generation resets the form.",
          "views.Dashboard.SetDashboardScope"
        ]
      ]
    },
    "rename-save": {
      "title": "Save, refresh or retry a rename",
      "summary": "An asynchronous scoped command reuses existing policy and persistence, then completes only the matching pending dialog.",
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
          "ui.Coordinator"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "manager",
          "session.Manager"
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
          "updateRename(Enter)",
          "Trim input; blank becomes session ID. Block duplicate submit and Esc while pending.",
          "views.Dashboard.updateRename"
        ],
        [
          "dash",
          "coord",
          "call",
          "dashboardCallbacks → rename command",
          "The callback captures service, dashboard ID and generation on the UI goroutine.",
          "ui.Coordinator.dashboardCallbacks"
        ],
        [
          "coord",
          "svc",
          "cmd",
          "renameDashboardSession",
          "Execute RenameDashboardSession using captured values.",
          "ui.Coordinator.renameDashboardSession"
        ],
        [
          "svc",
          "manager",
          "call",
          "RenameDashboardSession → SetNameDetailedChecked",
          "Existing membership policy runs under manager admission.",
          "api.Service.RenameDashboardSession"
        ],
        [
          "manager",
          "disk",
          "io",
          "SetNameDetailedChecked: persist session",
          "Reject lifecycle conflicts; update memory before saving, so save failure is not rollback.",
          "session.Manager.SetNameDetailedChecked"
        ],
        [
          "coord",
          "coord",
          "msg",
          "handleSessionMessages(result)",
          "DashboardSessionRenamedMsg must match dashboard, generation and current view. A successful inline rename also advances the session epoch so older refreshes cannot restore the old name.",
          "ui.Coordinator.handleSessionMessages"
        ],
        [
          "coord",
          "dash",
          "call",
          "CompleteRename(sessionID, err)",
          "Only a matching pending target completes: success closes; error keeps the draft for retry.",
          "views.Dashboard.CompleteRename"
        ],
        [
          "coord",
          "coord",
          "call",
          "refreshDataSeeded",
          "Refresh after success or failure. No launch, switch, restart or attach request.",
          "ui.Coordinator.refreshDataSeeded"
        ]
      ],
      "sources": {
        "refs": [
          "messages.DashboardSessionRenamedMsg"
        ],
        "files": [
          "internal/ui/coordinator_test.go",
          "internal/ui/views/dashboard_test.go"
        ]
      }
    },
    "rename-layout": {
      "title": "Render rename controls and help",
      "summary": "The new dialog and compact help remain usable at the supported 80×24 minimum.",
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
        ]
      ],
      "steps": [
        [
          "dash",
          "modal",
          "call",
          "NewDashboard: reuse Modal",
          "Create rename input and modal alongside the existing delete modal.",
          "views.NewDashboard"
        ],
        [
          "user",
          "dash",
          "key",
          "SetSize(width, height)",
          "Resize the dialog together with dashboard panes.",
          "views.Dashboard.SetSize"
        ],
        [
          "dash",
          "dash",
          "call",
          "dashboardModalInputWidth",
          "Use the responsive modal width and cap the editable input width.",
          "views.dashboardModalInputWidth"
        ],
        [
          "dash",
          "modal",
          "call",
          "View: input, controls and error",
          "Sanitize and truncate a long save error rather than wrapping it beyond the terminal.",
          "views.Dashboard.View"
        ],
        [
          "dash",
          "dash",
          "call",
          "footerText",
          "Scoped compact footer includes e=Rename at 80 columns.",
          "views.Dashboard.footerText"
        ]
      ]
    },
    "mcp-open": {
      "title": "Open the MCP server list",
      "summary": "Press m on an open dashboard to choose a server or New.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "d",
          "views.Dashboard"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "m",
          "views.MCPManager"
        ]
      ],
      "steps": [
        [
          "u",
          "d",
          "key",
          "m",
          "Dashboard key invokes its captured navigation callback.",
          "views.Dashboard.Update"
        ],
        [
          "d",
          "c",
          "cmd",
          "Request MCP configuration",
          "Callback now names the manager.",
          "ui.Coordinator.dashboardCallbacks"
        ],
        [
          "c",
          "m",
          "call",
          "Open manager; assign visit scope",
          "Navigation remains on the coordinator goroutine.",
          "ui.Coordinator.navigateTo"
        ],
        [
          "c",
          "m",
          "call",
          "Initialize modal and list",
          "Registry data already arrives via full snapshots.",
          "views.MCPManager.Init"
        ],
        [
          "m",
          "m",
          "call",
          "Render available MCPs and New",
          "Built-in row is marked read-only.",
          "views.MCPManager.View"
        ]
      ],
      "sources": {
        "refs": [
          "components.SourceList",
          "components.SourceListCallbacks",
          "components.SourceListConfig",
          "components.SourceListMode",
          "components.SourceList.HandleDelete",
          "components.SourceList.HandleSave",
          "components.SourceList.Init",
          "components.SourceList.OpenForm",
          "components.SourceList.SetItems",
          "components.SourceList.SetSize",
          "components.SourceList.Update",
          "components.SourceList.View",
          "components.SourceList.cycleInputFocus",
          "components.SourceList.handleListRune",
          "components.SourceList.initEditForm",
          "components.SourceList.initNewForm",
          "components.SourceList.moveDown",
          "components.SourceList.moveUp",
          "components.SourceList.renderForm",
          "components.SourceList.renderList",
          "components.SourceList.updateForm",
          "components.SourceList.updateList",
          "messages.MCPManagerDeleteMsg",
          "messages.MCPManagerDeletedMsg",
          "messages.MCPManagerEditMsg",
          "messages.MCPManagerScope",
          "views.MCPFormCallbacks",
          "views.MCPManager",
          "views.mcpManagerItem",
          "views.MCPManager.AdmitDelete",
          "views.MCPManager.DeleteFinished",
          "views.MCPManager.DiscardDelete",
          "views.MCPManager.Init",
          "views.MCPManager.Leave",
          "views.MCPManager.SetData",
          "views.MCPManager.SetScope",
          "views.MCPManager.SetSize",
          "views.MCPManager.Update",
          "views.MCPManager.View",
          "views.mcpManagerItem.Display",
          "views.mcpManagerItem.ID",
          "views.NewMCPManager",
          "ui.Coordinator.initViews",
          "ui.convertViewType"
        ],
        "files": [
          "feature-mcpedit.md",
          "internal/ui/views/mcp_manager_test.go",
          "internal/ui/navigation_test.go"
        ]
      }
    },
    "mcp-delete": {
      "title": "Confirm deletion of a global MCP",
      "summary": "Delete opens a named modal; confirmation invokes application policy and refreshes consumers.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "m",
          "views.MCPManager"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "a",
          "ui.sessionActions"
        ],
        [
          "r",
          "config.FileRepository"
        ],
        [
          "svc",
          "api.Service"
        ]
      ],
      "steps": [
        [
          "u",
          "m",
          "key",
          "Delete on a custom server",
          "New is ignored; built-in returns a visible restriction.",
          "views.MCPManager.Update"
        ],
        [
          "m",
          "m",
          "call",
          "Open confirmation with captured ID",
          "Refresh/reorder cannot redirect the target.",
          "views.MCPManager.Update"
        ],
        [
          "u",
          "m",
          "key",
          "Enter/y confirms; Esc/n cancels",
          "No repository mutation before confirmation.",
          "views.MCPManager.Update"
        ],
        [
          "m",
          "c",
          "cmd",
          "Deliver scoped delete request",
          "Pending state blocks repeated Delete.",
          "ui.Coordinator.matchesMCPManagerScope"
        ],
        [
          "c",
          "m",
          "call",
          "Admit once for the matching visit",
          "Unadmitted requests may be canceled by leaving.",
          "views.MCPManager.AdmitDelete"
        ],
        [
          "c",
          "a",
          "cmd",
          "Run deletion outside UI goroutine",
          "Stable ID is passed rather than a row index.",
          "ui.deleteMCPFromManagerCommand"
        ],
        [
          "a",
          "svc",
          "call",
          "DeleteConfiguredMCP(id)",
          "Application revalidates built-in/reference rules and delegates persistence.",
          "api.Service.DeleteConfiguredMCP"
        ],
        [
          "a",
          "c",
          "msg",
          "Deletion result",
          "Success triggers global refresh; current errors stay visible.",
          "ui.Coordinator.handleSpecialMessages"
        ],
        [
          "c",
          "m",
          "call",
          "Release deletion pending state",
          "Allows retry after policy or persistence errors.",
          "views.MCPManager.DeleteFinished"
        ]
      ]
    },
    "mcp-stale": {
      "title": "Leave while a command is pending",
      "summary": "Delayed work cannot mutate the interaction state of a later manager visit.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "m",
          "views.MCPManager"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "t",
          "tea"
        ]
      ],
      "steps": [
        [
          "m",
          "t",
          "cmd",
          "Queue scoped edit/delete request",
          "Capture the origin before the command runs.",
          "views.MCPManager.Update"
        ],
        [
          "u",
          "m",
          "key",
          "Esc; return to dashboard",
          "An unadmitted confirmation is canceled.",
          "views.MCPManager.Update"
        ],
        [
          "c",
          "m",
          "call",
          "Leave clears unadmitted pending state",
          "Admitted repository work continues to its result.",
          "views.MCPManager.Leave"
        ],
        [
          "c",
          "m",
          "call",
          "Reopen with a new visit identity",
          "Later commands/results cannot reuse prior visit admission.",
          "views.MCPManager.SetScope"
        ],
        [
          "t",
          "c",
          "msg",
          "Deliver old command",
          "Scope comparison rejects it even if manager is current again.",
          "ui.Coordinator.matchesMCPManagerScope"
        ],
        [
          "c",
          "m",
          "call",
          "Discard only its unadmitted state",
          "Does not strand or clear a newer pending request.",
          "views.MCPManager.DiscardDelete"
        ],
        [
          "t",
          "c",
          "msg",
          "An admitted deletion eventually finishes",
          "Success still refreshes global registry; no result-driven navigation.",
          "ui.Coordinator.handleSpecialMessages"
        ]
      ],
      "sources": {
        "refs": [
          "messages.MCPManagerScope",
          "messages.MCPManagerEditMsg",
          "messages.MCPManagerDeleteMsg",
          "messages.MCPManagerDeletedMsg"
        ],
        "files": [
          "internal/ui/mcp_manager_gaps_test.go",
          "internal/ui/mcp_manager_review_test.go"
        ]
      }
    },
    "mcp-card": {
      "title": "Retain card editor MCP configuration",
      "summary": "The existing card editor create shortcut and deletion policy remain available.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "e",
          "views.CardEditor"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "a",
          "ui.sessionActions"
        ],
        [
          "r",
          "config.FileRepository"
        ],
        [
          "svc",
          "api.Service"
        ]
      ],
      "steps": [
        [
          "u",
          "e",
          "key",
          "Delete highlighted MCP",
          "This pre-existing route does not gain a manager modal.",
          "views.CardEditor.deleteSelectedMCP"
        ],
        [
          "e",
          "c",
          "cmd",
          "MCPDeleteRequestedMsg",
          "Coordinator delegates application work.",
          "ui.Coordinator.handleSpecialMessages"
        ],
        [
          "c",
          "a",
          "cmd",
          "deleteMCP",
          "Existing card reference policy is reused.",
          "ui.deleteMCPCommand"
        ],
        [
          "a",
          "svc",
          "call",
          "DeleteConfiguredMCP(id)",
          "Application revalidates built-in/reference rules and delegates persistence.",
          "api.Service.DeleteConfiguredMCP"
        ],
        [
          "a",
          "c",
          "msg",
          "Centralize completion and refresh",
          "Preserve editor deselection, refresh the registry globally.",
          "ui.Coordinator.handleSpecialMessages"
        ]
      ]
    },
    "launch-focus": {
      "title": "Keep the created session selected through naming and detach",
      "summary": "After workspace confirmation, launch and naming results maintain one selected session ID. Old refresh epochs cannot displace it.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "mgr",
          "session.Manager"
        ],
        [
          "launcher",
          "views.SessionLauncher"
        ],
        [
          "ref",
          "ui.refresher"
        ],
        [
          "dash",
          "views.Dashboard"
        ],
        [
          "list",
          "components.SessionList"
        ],
        [
          "preview",
          "components.SessionCenter"
        ]
      ],
      "steps": [
        [
          "mgr",
          "c",
          "msg",
          "SessionLaunchCompletedMsg(session ID)",
          "A usable launch advances the session epoch and requests focus for the exact ID.",
          "ui.Coordinator.handleSessionMessages"
        ],
        [
          "c",
          "dash",
          "call",
          "Focus the new session",
          "If its snapshot is not loaded yet, the one-shot focus waits for session data.",
          "ui.Coordinator.handleSessionMessages"
        ],
        [
          "u",
          "launcher",
          "key",
          "Confirm the session name",
          "The scoped naming result carries the confirmed name.",
          "views.SessionLauncher.handleNameSubmit"
        ],
        [
          "launcher",
          "c",
          "msg",
          "ScopedSessionNamedMsg(session ID, name)",
          "Advance the session epoch and keep focus pending until refreshed data contains that name.",
          "ui.Coordinator.handleSessionMessages"
        ],
        [
          "ref",
          "c",
          "msg",
          "Delayed old-epoch refresh",
          "Reject snapshots started before launch or rename, even within the same dashboard generation.",
          "ui.Coordinator.handleSessionsRefreshed"
        ],
        [
          "ref",
          "dash",
          "msg",
          "Current session snapshot",
          "Select by ID, clear search only when it hides the target, and sync center.",
          "ui.Coordinator.handleDataLoaded"
        ],
        [
          "u",
          "c",
          "key",
          "Ctrl+K to detach",
          "The detach refresh retains the selected ID; later refreshes do not steal focus again.",
          "ui.Coordinator.handleSessionMessages"
        ],
        [
          "dash",
          "list",
          "call",
          "Rebuild session list and center by selected ID",
          "Session-list ordering changes do not displace the created session.",
          "views.Dashboard.SetInventory"
        ],
        [
          "dash",
          "preview",
          "call",
          "sync center with focused session",
          "Session action hints and center agree with the selected session.",
          "views.Dashboard.syncCenterSelection"
        ]
      ],
      "sources": {
        "refs": [
          "session.CardPendingReference",
          "session.CardReferenceImpact",
          "session.CardSessionReference",
          "session.Manager.CancelPendingCardReferences",
          "session.Manager.CardReferences",
          "session.Manager.List",
          "session.Manager.ReserveCardReferenceForDashboardWithCancel",
          "session.Manager.SetNameDetailedChecked",
          "ui.Coordinator.handleDataLoaded",
          "ui.Coordinator.handleSessionMessages",
          "ui.Coordinator.handleSessionsRefreshed",
          "session.Manager.ShutdownStarted",
          "session.Manager.TemporaryWorkspaceRoot",
          "session.Manager.WaitCardReferences",
          "session.Manager.WithCardReferencesAndSnapshot",
          "session.Manager.bindPendingWorkspace",
          "session.Manager.cardReferencesLocked",
          "session.Manager.listLocked",
          "session.Manager.notifyCardReferencesLocked",
          "session.Manager.reserveCardReference",
          "session.NewManager",
          "session.Session.data",
          "session.Session.snapshot",
          "session.cardReservation",
          "views.SessionLauncher",
          "views.SessionLauncher.Reset",
          "views.SessionLauncher.SetNameResult",
          "views.SessionLauncher.handleNameSubmit",
          "views.SessionLauncherCallbacks",
          "views.Dashboard.FocusSession",
          "views.Dashboard.FocusSessionAfterRename",
          "views.Dashboard.applyPendingSessionFocus",
          "views.matchesSessionQuery",
          "components.SessionList.refreshItems",
          "ui.Coordinator.hasStaleSessionSnapshot",
          "messages.SessionsRefreshedMsg"
        ],
        "files": [
          "internal/ui/launch_focus_internal_test.go",
          "internal/ui/inline_rename_refresh_internal_test.go",
          "internal/ui/components/session_list_test.go",
          "tests/e2e/launch_focus_test.go"
        ]
      }
    },
    "center-workspace": {
      "title": "Workspace facts when reports are empty",
      "summary": "Inspect the saved actual workspace on initial empty/fallback loading or a scheduled refresh. Failed attempts without cache still wait for the 30-second retry; healthy report events do not restart workspace work.",
      "participants": [
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "panel",
          "components.SessionCenter"
        ]
      ],
      "steps": [
        [
          "c",
          "c",
          "msg",
          "Successful empty report read, or failed read without report cache",
          "Initial workspace loading is immediate; after an attempt, only a due refresh enables another inspection. A failed history read remains a warning, never proof of empty history.",
          "ui.Coordinator.handleCenter"
        ],
        [
          "c",
          "svc",
          "cmd",
          "ReadSessionWorkspace(scope)",
          "Cancelable read; independent report events can preempt it.",
          "ui.Coordinator.centerWorkspace"
        ],
        [
          "svc",
          "svc",
          "call",
          "Validate scope and resolve saved actual directory",
          "Use the session WorkingDir, including temporary copies; revalidate after I/O.",
          "api.Service.ReadSessionWorkspace"
        ],
        [
          "svc",
          "svc",
          "call",
          "Inspect directory and discover repository metadata",
          "Missing/inaccessible workspaces error; do not recreate them.",
          "api.inspectWorkspace"
        ],
        [
          "svc",
          "svc",
          "io",
          "Discover Git paths and inspect with isolated configuration",
          "Build a private Git directory from selected data, including the active split index when present. Disable fsmonitor and hooks. Configured clean filters and repositories with submodules use a labelled directory fallback.",
          "api.inspectGitWorkspace"
        ],
        [
          "svc",
          "svc",
          "call",
          "Parse Git records, including rename origin",
          "Parse NUL-delimited porcelain v2, preserve rename origins, and cap results at 1,000 entries. Git commands use argv, bounded output and a deadline. Unix cancellation and output overflow terminate the owned process group.",
          "api.parseGitStatus"
        ],
        [
          "svc",
          "svc",
          "io",
          "Non-Git or labelled Git-error fallback: list immediate children",
          "A metadata or status error is visible alongside the fallback when directory access succeeds. Include dotfiles and optional stat metadata without recursion.",
          "api.listWorkspace"
        ],
        [
          "svc",
          "c",
          "msg",
          "Structured workspace result and warnings",
          "Late results cannot overwrite reports. Failed workspace data stays marked stale until actual replacement workspace data or report content succeeds; history-only reads cannot certify its freshness.",
          "ui.Coordinator.handleCenter"
        ],
        [
          "c",
          "panel",
          "call",
          "Render Git or directory information beneath the header",
          "Render root-relative Git entries or directory metadata; sanitize each name, rename origin and symlink target independently from multiline reports. The header stays visible when a report takes over the body.",
          "components.SessionCenter.renderWorkspace"
        ]
      ],
      "sources": {
        "files": [
          "internal/api/center_workspace.go",
          "internal/api/center_workspace_linux.go",
          "internal/api/center_workspace_other.go",
          "internal/api/center_process_unix.go",
          "internal/api/center_process_other.go",
          "internal/api/center_git.go"
        ],
        "refs": [
          "api.configureGitProcess",
          "api.killGitProcess",
          "api.runGitWithEnv",
          "api.gitStatusSettings",
          "api.linkSharedGitIndex"
        ]
      }
    },
    "center-recovery": {
      "title": "Center recovery and resource release",
      "summary": "Cancel selection-scoped work on navigation, but preserve a detached successful-content cache for the same dashboard and stable session. Recovery cadence and workspace freshness remain independent of retained content.",
      "participants": [
        [
          "c",
          "ui.Coordinator"
        ],
        [
          "sub",
          "api.reportSubscription"
        ],
        [
          "clock",
          "ui.systemCenterClock"
        ],
        [
          "dash",
          "views.Dashboard"
        ]
      ],
      "steps": [
        [
          "c",
          "c",
          "msg",
          "Read failure or unexpected subscription closure",
          "Coalesce notices after report failure. Track workspaceAttempted/workspaceDue separately from cache presence; retain workspaceStale across successful-empty and failed report reads.",
          "ui.Coordinator.handleCenter"
        ],
        [
          "c",
          "clock",
          "cmd",
          "Schedule 30-second retry only when idle",
          "No duplicate timer while setup/read/workspace work is active.",
          "ui.Coordinator.centerSchedule"
        ],
        [
          "clock",
          "clock",
          "io",
          "Wait for timer or cancellation",
          "Context cancellation stops pending waits.",
          "ui.systemCenterClock.Wait"
        ],
        [
          "clock",
          "c",
          "msg",
          "Timer generation/request checked",
          "Resubscribe first if disconnected, otherwise reread reports.",
          "ui.Coordinator.handleCenter"
        ],
        [
          "dash",
          "c",
          "msg",
          "Resize or selected-session reconciliation",
          "Return restores matching successful content before subscribe/reread; failed history keeps cached reports stale, successful empty clears them, and another selection inherits nothing.",
          "ui.Coordinator.reconcileCenter"
        ],
        [
          "c",
          "sub",
          "call",
          "Suspend on navigation: detach cache, cancel operations",
          "suspendCenter saves detached same-dashboard/session report or workspace content, then clearCenter cancels reads, timer and subscription. Returning restores only a matching selection; scope loss discards it.",
          "ui.Coordinator.suspendCenter"
        ],
        [
          "c",
          "dash",
          "call",
          "Scope rejection clears cached selected-session content",
          "Block immediate retry for the same invalid selection.",
          "ui.Coordinator.invalidateCenterScope"
        ]
      ],
      "sources": {
        "files": [
          "internal/ui/center.go",
          "internal/ui/navigation.go",
          "internal/ui/ui.go"
        ]
      }
    },
    "attach-disposal": {
      "title": "Dispose of stale attachments",
      "summary": "A confirmed missing session cancels acquisition and closes queued handles.",
      "participants": [
        [
          "coord",
          "ui.Coordinator"
        ],
        [
          "owner",
          "ui.attachmentOwner"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "term",
          "terminal.Command"
        ]
      ],
      "steps": [
        [
          "coord",
          "owner",
          "call",
          "Capture session and dashboard ownership",
          "",
          "ui.attachmentOwner.begin"
        ],
        [
          "coord",
          "svc",
          "cmd",
          "Acquire scoped attachment",
          "",
          "api.Service.AttachInDashboard"
        ],
        [
          "coord",
          "owner",
          "call",
          "Invalidate missing target",
          "Uses a successful, newer inventory or explicit removal.",
          "ui.attachmentOwner.invalidateMissing"
        ],
        [
          "svc",
          "owner",
          "return",
          "Publish or close late resource",
          "",
          "ui.attachmentOwner.publish"
        ],
        [
          "coord",
          "term",
          "call",
          "Run a still-current attachment",
          "",
          "ui.scopedAttachCommand"
        ]
      ]
    },
    "deletion-outcomes": {
      "title": "Show deletion progress and retry",
      "summary": "Opaque tokens authorize a rechecked operation, and UI feedback follows its authoritative outcome.",
      "participants": [
        [
          "list",
          "views.DashboardList"
        ],
        [
          "ui",
          "ui.Coordinator"
        ],
        [
          "svc",
          "api.Service"
        ],
        [
          "disk",
          "config.FileRepository"
        ]
      ],
      "steps": [
        [
          "list",
          "svc",
          "cmd",
          "Prepare dashboard deletion summary",
          "",
          "api.Service.PrepareHumanDashboardDeletion"
        ],
        [
          "list",
          "svc",
          "cmd",
          "Confirm token and foreign impact",
          "",
          "api.Service.ConfirmHumanDashboardDeletion"
        ],
        [
          "svc",
          "disk",
          "io",
          "Persist intent before teardown",
          "",
          "config.FileRepository.BeginDashboardDeletion"
        ],
        [
          "svc",
          "svc",
          "call",
          "Run cascade and checkpoint work",
          "",
          "api.Service.runDashboardDeletion"
        ],
        [
          "svc",
          "ui",
          "msg",
          "Project completed or pending progress",
          "",
          "api.dashboardDeletionProgress"
        ],
        [
          "ui",
          "list",
          "call",
          "Show outcome and available retry",
          "",
          "views.DashboardList.ShowDeletionResult"
        ]
      ]
    },
    "reasoning-effort-activation": {
      "title": "Apply reasoning effort when starting a process",
      "summary": "Launch, relaunch and card switch build argv from the resolved card. The driver handles CLI-specific effort options.",
      "participants": [
        [
          "manager",
          "session.Manager"
        ],
        [
          "registry",
          "driver.MapRegistry"
        ],
        [
          "driver",
          "driver.AgentDriver"
        ]
      ],
      "steps": [
        [
          "manager",
          "registry",
          "call",
          "Get(card.Driver)",
          "The built-in registry registers Antigravity for agy. OpenCode uses the generic adapter.",
          "driver.MapRegistry.Get"
        ],
        [
          "manager",
          "manager",
          "call",
          "activationArgs",
          "Both spawn paths use this builder. A switch uses the destination card.",
          "session.activationArgs"
        ],
        [
          "manager",
          "driver",
          "call",
          "SetModel(model)",
          "Append model arguments first.",
          "driver.GenericCLIDriver.SetModel"
        ],
        [
          "manager",
          "driver",
          "call",
          "SetReasoningEffort(effort)",
          "Claude and Antigravity use --effort. Codex uses -c model_reasoning_effort=value. Generic drivers emit no effort arguments.",
          "driver.AgentDriver"
        ],
        [
          "manager",
          "driver",
          "call",
          "SetSession(conversationID)",
          "For Codex, resume follows the configuration override.",
          "driver.CodexCLIDriver.SetSession"
        ],
        [
          "manager",
          "driver",
          "io",
          "Spawn with argv",
          "The PTY process receives discrete argv elements.",
          "driver.GenericCLIDriver.Spawn"
        ]
      ],
      "sources": {
        "refs": [
          "session.Manager.startSessionCore",
          "session.Manager.switchAgentWithChoiceCore",
          "driver.NewBuiltinRegistry",
          "driver.NewAntigravityCLIDriver",
          "driver.AntigravityCLIDriver.SetReasoningEffort",
          "driver.ClaudeCodeDriver.SetReasoningEffort",
          "driver.CodexCLIDriver.SetReasoningEffort",
          "driver.GenericCLIDriver.SetReasoningEffort",
          "driver.effortFlagArgs"
        ],
        "files": [
          "internal/driver/driver_test.go",
          "internal/session/reasoning_effort_test.go",
          "tests/e2e/launch_focus_test.go"
        ]
      }
    },
    "launch-workspace-selection": {
      "title": "Confirm a new session launch",
      "summary": "Choose the displayed main and sub-workspaces before starting and naming a new session.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "ui",
          "ui.Coordinator"
        ],
        [
          "view",
          "views.SessionLauncher"
        ],
        [
          "api",
          "api.Service"
        ],
        [
          "mgr",
          "session.Manager"
        ]
      ],
      "steps": [
        [
          "user",
          "ui",
          "key",
          "Enter on a card",
          "The active dashboard and card scope travel with the read.",
          "ui.Coordinator.dashboardCallbacks"
        ],
        [
          "ui",
          "api",
          "cmd",
          "ReadLaunchPreparation()",
          "Read the eligible card’s default workspace IDs and workspace choices.",
          "api.Service.ReadLaunchPreparation"
        ],
        [
          "ui",
          "view",
          "msg",
          "Open the workspace form",
          "Initialize a detached draft from the scoped result.",
          "ui.Coordinator.launchCard"
        ],
        [
          "user",
          "view",
          "key",
          "Space opens the picker",
          "Tab moves between rows. The card editor and launcher use the same picker.",
          "views.workspaceFields.open"
        ],
        [
          "user",
          "view",
          "key",
          "Enter on the workspace form",
          "Submit all displayed IDs, even when the user has left the defaults as shown.",
          "views.SessionLauncher.updateWorkspaceForm"
        ],
        [
          "view",
          "api",
          "cmd",
          "LaunchSession(WorkspaceIDs)",
          "The command captures a detached copy of the selection.",
          "api.Service.LaunchSession"
        ],
        [
          "api",
          "ui",
          "return",
          "Apply launch result",
          "The existing naming and attach flow handles usable and failed activation outcomes.",
          "ui.Coordinator.handleSessionMessages"
        ]
      ],
      "sources": {
        "refs": [
          "ui.Coordinator.sessionLauncherCallbacks",
          "views.SessionLauncher.Reset",
          "views.SessionLauncher.SetWorkspaceDraft",
          "views.SessionLauncher.View",
          "views.NewSessionLauncher",
          "views.SessionLauncher.SetSize"
        ]
      }
    },
    "launch-workspace-navigation": {
      "title": "Cancel, quit or visit workspace sources",
      "summary": "Source visits are scoped to the current launch draft. Escape closes the draft to new input.",
      "participants": [
        [
          "user",
          "user"
        ],
        [
          "view",
          "views.SessionLauncher"
        ],
        [
          "ui",
          "ui.Coordinator"
        ],
        [
          "sources",
          "views.SourceManager"
        ]
      ],
      "steps": [
        [
          "user",
          "view",
          "key",
          "Ctrl+E from the workspace rows",
          "Open the source manager as a side trip.",
          "views.SessionLauncher.updateWorkspaceForm"
        ],
        [
          "view",
          "ui",
          "msg",
          "Check dashboard, card and request",
          "Only an active draft can open sources.",
          "ui.Coordinator.handleLaunchWorkspaceNavigation"
        ],
        [
          "ui",
          "sources",
          "call",
          "Open workspace sources",
          "The launch scope survives this side trip.",
          "ui.Coordinator.invalidateSessionRequestsOnNavigation"
        ],
        [
          "user",
          "sources",
          "key",
          "Return from workspace sources",
          "The draft keeps its selections and focus.",
          "views.SourceManager.Update"
        ],
        [
          "user",
          "view",
          "key",
          "Esc cancels the draft",
          "Disable submission immediately while cancellation navigation is pending.",
          "views.SessionLauncher.Update"
        ],
        [
          "view",
          "ui",
          "msg",
          "Return to the dashboard",
          "Late source commands cannot act on a cancelled or submitted draft.",
          "ui.Coordinator.launchWorkspaceNavigation"
        ],
        [
          "user",
          "view",
          "key",
          "Ctrl+C quits from the rows",
          "An open picker handles its own keys, as in the card editor.",
          "views.SessionLauncher.updateWorkspaceForm"
        ]
      ],
      "sources": {
        "refs": [
          "views.SessionLauncher.WorkspaceDraftActive",
          "views.SessionLauncher.WorkspaceDraftCancelling"
        ]
      }
    },
    "workspace-choices-refresh": {
      "title": "Refresh workspace choices",
      "summary": "Workspace registry reads update option lists while preserving the selected IDs and previous choices on a read failure.",
      "participants": [
        [
          "reg",
          "registry"
        ],
        [
          "api",
          "api.Service"
        ],
        [
          "ui",
          "ui.Coordinator"
        ],
        [
          "view",
          "views.SessionLauncher"
        ],
        [
          "editor",
          "views.CardEditor"
        ]
      ],
      "steps": [
        [
          "api",
          "reg",
          "io",
          "ReadWorkspaceChoices()",
          "The shared mapping returns options or a scoped read error.",
          "api.Service.ReadWorkspaceChoices"
        ],
        [
          "api",
          "ui",
          "return",
          "Return launch preparation",
          "The error remains visible alongside the editable draft.",
          "api.Service.ReadLaunchPreparation"
        ],
        [
          "ui",
          "editor",
          "call",
          "Apply editor choices",
          "Refresh the existing editor and also supply the launcher.",
          "ui.Coordinator.applyEditorChoices"
        ],
        [
          "ui",
          "view",
          "call",
          "SetWorkspaceChoices()",
          "A successful read refreshes choices without replacing selected IDs.",
          "views.SessionLauncher.SetWorkspaceChoices"
        ],
        [
          "view",
          "view",
          "self",
          "Retain options on a failed read",
          "The shared fields leave the previous options available.",
          "views.workspaceFields.setChoices"
        ]
      ],
      "sources": {
        "refs": [
          "ui.Coordinator.handleDataLoaded",
          "views.CardEditor.SetChoices"
        ]
      }
    },
    "mcp-workspace-discovery": {
      "title": "Discover registered workspaces",
      "summary": "A live Dashboard Control caller lists registered workspace IDs, including entries no card references.",
      "participants": [
        [
          "client",
          "client"
        ],
        [
          "mcp",
          "mcp.Runtime"
        ],
        [
          "api",
          "api.Service"
        ],
        [
          "reg",
          "registry"
        ]
      ],
      "steps": [
        [
          "client",
          "mcp",
          "io",
          "Choose a card",
          "Use list_cards to choose an available card. This optional lookup remains dashboard-scoped.",
          "api.Service.ListCards"
        ],
        [
          "client",
          "mcp",
          "io",
          "Call list_workspaces",
          "The tool accepts an empty input object and uses the request credential.",
          "mcp.addDiscoveryTools"
        ],
        [
          "mcp",
          "api",
          "call",
          "ListWorkspaces()",
          "The service owns authorization and discovery.",
          "api.Service.ListWorkspaces"
        ],
        [
          "api",
          "api",
          "self",
          "ResolveCaller()",
          "Require a live activation with Dashboard Control in an available dashboard before registry access.",
          "api.Service.ResolveCaller"
        ],
        [
          "api",
          "reg",
          "io",
          "ReadWorkspaceChoices()",
          "Read the registered entries without syncing sources or checking paths.",
          "api.Service.ReadWorkspaceChoices"
        ],
        [
          "api",
          "mcp",
          "return",
          "Return sorted items",
          "The result includes ID, name and registered path. An empty registry yields an empty array and a read failure propagates.",
          "api.Service.ListWorkspaces"
        ],
        [
          "mcp",
          "client",
          "return",
          "Send discovery result",
          "Both dashboard callers see the global registry. Card and session reads keep their dashboard scope.",
          "mcp.addDiscoveryTools"
        ]
      ],
      "sources": {
        "refs": [
          "api.WorkspaceDiscovery",
          "mcp.discoveryService"
        ]
      }
    },
    "mcp-workspace-create": {
      "title": "Create a session through MCP",
      "summary": "An authenticated caller passes the IDs returned by list_workspaces in an ordered create_session request.",
      "participants": [
        [
          "client",
          "client"
        ],
        [
          "mcp",
          "mcp.Runtime"
        ],
        [
          "api",
          "api.Service"
        ],
        [
          "runner",
          "api.OperationRunner"
        ],
        [
          "mgr",
          "session.Manager"
        ]
      ],
      "steps": [
        [
          "client",
          "mcp",
          "io",
          "Call create_session",
          "Pass the ordered IDs from list_workspaces. Omitting the field uses saved card defaults.",
          "mcp.addLifecycleTools"
        ],
        [
          "mcp",
          "mcp",
          "self",
          "Decode workspace_ids",
          "An explicit null is rejected and an empty list reaches request validation.",
          "mcp.workspaceIDs.UnmarshalJSON"
        ],
        [
          "mcp",
          "api",
          "call",
          "CreateSessionWithOptions()",
          "Capture the name and ordered workspace IDs before queuing work.",
          "api.Service.CreateSessionWithOptions"
        ],
        [
          "api",
          "runner",
          "call",
          "Replay or queue the operation",
          "A changed list with the same request ID conflicts. Omission keeps the prior fingerprint shape.",
          "api.OperationRunner.Submit"
        ],
        [
          "runner",
          "api",
          "cmd",
          "Admit and prepare the launch",
          "Existing dashboard scope and card checks still apply.",
          "api.Service.launchDashboardWithIDEvidence"
        ],
        [
          "api",
          "mgr",
          "call",
          "Start the requested session",
          "An invalid selection fails before workspace preparation.",
          "api.Service.cardWithWorkspaceOverrides"
        ],
        [
          "mgr",
          "api",
          "return",
          "Return operation outcome",
          "The runner returns the saved effect with actual paths.",
          "api.Service.CreateSessionWithOptions"
        ]
      ],
      "sources": {
        "refs": [
          "api.firstName",
          "api.Service.launchDashboardWithID"
        ]
      }
    }
  },
  "rules": [
    {
      "n": "1",
      "title": "No tea.Cmd closure touches coordinator state",
      "ref": "ui.Coordinator",
      "why": "A command's closure runs on its own goroutine. Reading or writing c.views, c.center, c.currentView, c.err or c.settings there is a concurrent access that can crash the app. Snapshot what you need before returning the closure.",
      "check": "Capture IDs, generation and settings by value. Run the coordinator command-closure race regression; external test adapters call the real scoped route.",
      "flows": [
        "tick",
        "launch",
        "terminal",
        "ide"
      ],
      "sources": {
        "refs": [
          "ui.Coordinator.launchIDE",
          "ui.Coordinator.launchTerminal",
          "api.Service.OpenSessionWorkspace"
        ],
        "files": [
          "internal/ui/command_race_internal_test.go"
        ]
      }
    },
    {
      "n": "2",
      "title": "Never navigate from inside a view's Update",
      "ref": "ui.Coordinator.navigateTo",
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
      "check": "Focused read results carry explicit attempted/error state. Failed inventories retain prior choices; center data is published separately through SetCenterData, not a partial DataLoadedMsg.",
      "flows": [
        "tick",
        "preview"
      ]
    },
    {
      "n": "4",
      "title": "The error banner clears on the next keypress",
      "ref": "ui.Coordinator.Update",
      "why": "Navigation leaves it alone, so an ErrorMsg batched beside a NavigateMsg survives in either order.",
      "check": "TestErrorSurvivesNavigationInEitherOrder, TestKeyPressClearsBanner.",
      "flows": [
        "mutation",
        "workspace-command-errors"
      ],
      "sources": {
        "refs": [
          "ui.Coordinator.handleSessionMessages"
        ],
        "files": [
          "internal/ui/error_banner_test.go",
          "internal/ui/terminal_launch_test.go"
        ]
      }
    },
    {
      "n": "5",
      "title": "Application saves preserve unedited card fields",
      "ref": "api.Service.UpdateEditorCard",
      "why": "The UI carries only editable values and explicit changes. Application policy reloads storage and checks ownership and opaque full-record revision before applying those changes.",
      "check": "Card editor and API membership/revision tests; missing choices must not erase selected IDs.",
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
      "ref": "session.Manager.LaunchInDashboardWithIDContext",
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
      "ref": "session.Manager.prepareSubWorkspacesSnapshot",
      "why": "Sub-workspace manifest decisions and copy publication share the manager preparation gate. Existing managed copies retain work; concurrent preparations must not publish conflicting destinations. Initial launch captures only its selected descriptors under the gate, while the manifest retains directory-wide ownership.",
      "check": "Review the mutex scope and transaction bookkeeping; run workspace/collision/failure tests and go test -race ./....",
      "flows": [
        "workspace",
        "launch"
      ],
      "sources": {
        "refs": [
          "session.subWorkspaceTransaction.rollback",
          "session.shouldCopySubWorkspace",
          "session.Manager.prepareSubWorkspacesSnapshot"
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
      "why": "Dashboard Control requires a live deployed credential and selected Kankuro MCP. Persisted session membership determines the caller dashboard; human navigation and caller-supplied IDs cannot widen access.",
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
        "recovery",
        "mcp-workspace-discovery"
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
      "title": "Card ownership and deletion share admission",
      "ref": "api.Service.ConfirmCardDeletion",
      "why": "Only the immutable owner edits or deletes a card. Unsharing refuses foreign references; confirmed global deletion preserves existing foreign sessions while blocking new use. Persist exact pending cleanup paths before cancelling activations.",
      "check": "Owner/foreign-reference tests, pending-activation cleanup failures, durable tombstone restart/retry tests.",
      "flows": [
        "card-delete",
        "card-recovery",
        "deleted-session"
      ]
    },
    {
      "n": "18",
      "title": "Explicit reports carry user-defined workflow meaning",
      "ref": "session.MessageStore.RecordReply",
      "why": "Progress, completed, question and blocked retain user-defined workflow meaning. Completed, blocked and question also wake the exact original sender; progress is silent and the worker must keep working. Wake-up does not mean Kankuro approved work or advanced a stage.",
      "check": "Store, API and MCP tests cover report identity, idempotency, kinds, notification metadata and correlated retrieval.",
      "flows": [
        "reports",
        "report-progress",
        "report-inspection"
      ]
    },
    {
      "n": "19",
      "title": "Prepare configuration before runtime",
      "ref": "configstartup.PrepareStartup",
      "why": "Initialize an empty root directly; existing roots require complete current-format records. Hold the root lock throughout runtime.",
      "check": "Fresh/restart/partial-root and cross-record validation tests in internal/configstartup.",
      "flows": [
        "startup"
      ]
    },
    {
      "n": "20",
      "title": "Session membership never moves",
      "ref": "api.persistedDashboardScope.DashboardForSession",
      "why": "Sharing a card never shares sessions. Resolve human and MCP operations against immutable session dashboard membership.",
      "check": "Scoped UI, service and MCP tests; checked manager launch/switch/relaunch tests.",
      "flows": [
        "caller-scope",
        "launch",
        "resume"
      ]
    },
    {
      "n": "21",
      "title": "Dashboard cascades are resumable",
      "ref": "api.Service.RetryDashboardDeletion",
      "why": "Save intent before teardown and checkpoint each effect. Preserve foreign sessions and global registries; do not finalize while cleanup remains.",
      "check": "Deletion failure, removed-row recovery, restart and foreign-impact regressions.",
      "flows": [
        "delete-dashboard",
        "cascade-recovery"
      ]
    },
    {
      "n": "22",
      "title": "Ignore obsolete asynchronous UI results",
      "ref": "ui.Coordinator.handleDataLoaded",
      "why": "Capture dashboard/generation before commands, plus the session epoch for refreshes and manager visit for MCP edits/deletes. Late work must not retarget another view or displace a newly launched/renamed selection.",
      "check": "Scoped command-race tests, cross-dashboard MCP tests, and full/session/fingerprint stale-refresh tests.",
      "flows": [
        "navigation",
        "tick",
        "attach",
        "launch",
        "mcp-stale"
      ]
    },
    {
      "n": "23",
      "title": "Escape never quits the dashboard list",
      "ref": "views.DashboardList.Update",
      "why": "Escape returns from a dashboard or cancels a dialog. Repeated Escape on the list stays there. Disable the list widget default quit bindings; q is the explicit quit key.",
      "check": "Dashboard-list keyboard regression at supported and small terminal sizes; real terminal check.",
      "flows": [
        "navigation"
      ]
    },
    {
      "n": "24",
      "title": "Retry deletion with its original key",
      "ref": "views.DashboardList.retryDeletion",
      "why": "Delete/Backspace starts deletion for an active dashboard and resumes its durable manifest after partial failure. The status screen uses the same retry keys.",
      "check": "Both key regressions, preparation failure and actual permission-failure/repaired-retry terminal check.",
      "flows": [
        "cascade-recovery",
        "list-errors"
      ]
    },
    {
      "n": "25",
      "title": "Save the notification with the report",
      "ref": "session.MessageStore.RecordReply",
      "why": "A qualifying report and its pending wake-up must share one atomic write. Once saved, the dispatcher owns retries and the worker can finish its turn without waiting for delivery. Progress creates no delivery obligation.",
      "check": "TestReportNotificationInitialStateIsDurableAndIdempotent and report persistence rollback tests.",
      "flows": [
        "reports",
        "report-progress",
        "notification-dispatch"
      ]
    },
    {
      "n": "26",
      "title": "Report authority permits one original destination",
      "ref": "session.Manager.ReserveReportNotification",
      "why": "A worker need not have Dashboard Control. Its retained report permits a notification only to the original sender session and activation; stopping the reporter does not revoke it, but history deletion and target replacement do.",
      "check": "Non-controller, reporter-stop, original-activation replacement, history-deletion and dashboard-deletion notification regressions.",
      "flows": [
        "notification-authority",
        "notification-dispatch"
      ]
    },
    {
      "n": "27",
      "title": "Only conclusive zero-byte evidence permits another attempt",
      "ref": "api.OperationRunner.reconcileNotification",
      "why": "Queue or temporary persistence failure retries the same pre-admission identity. After acceptance, reconcile the existing delivery; advance an attempt only when no bytes were written. Submitted or uncertain input is never replayed automatically.",
      "check": "Queue/journal recovery, reconciliation evidence, stable-attempt and concurrent-dispatch tests; restart tests.",
      "flows": [
        "notification-retry",
        "recovery",
        "notification-cancel"
      ]
    },
    {
      "n": "28",
      "title": "Notifications request no report back",
      "ref": "driver.formatPrompt",
      "why": "A wake-up carries original message/reply IDs and asks the orchestrator to retrieve retained text. Reusing task report-back instructions risks a reply loop; injecting report bodies violates one-line prompt constraints.",
      "check": "TestReportNotificationPromptDoesNotIncludeReportBody, real-MCP report wake-up and PTY-backed API tests.",
      "flows": [
        "notification-write",
        "report-inspection"
      ]
    },
    {
      "n": "29",
      "title": "MCP editing preserves hidden fields and repository policy",
      "ref": "views.MCPForm.handleSubmit",
      "why": "An edit overlays the loaded definition, so metadata absent from the form survives. Global deletion keeps built-in and saved-card reference guards; confirmation is tied to a stable ID.",
      "check": "Metadata/New-reset, protected/reference deletion, captured-ID, write-error/retry and tui-test configuration journeys.",
      "flows": [
        "mcp",
        "mcp-delete"
      ],
      "sources": {
        "refs": [
          "config.FileRepository.DeleteMCP",
          "views.MCPManager.Update"
        ],
        "files": [
          "internal/ui/mcp_manager_review_test.go",
          "tests/e2e/mcp_manager_test.go"
        ]
      }
    },
    {
      "n": "30",
      "title": "Reports belong to stable author sessions",
      "ref": "session.MessageStore.RecentAuthoredReports",
      "why": "Parent message sender and reply author differ. Filter author and dashboard across activations; never present received messages or wake-up records as authored reports.",
      "check": "Authored report projection, bounds/reload and dashboard scope tests.",
      "flows": [
        "preview"
      ]
    },
    {
      "n": "31",
      "title": "Report failure is not empty history",
      "ref": "ui.Coordinator.handleCenter",
      "why": "Report failure is not empty history. Preserve stale report content or labelled workspace fallback. A history-only read cannot clear workspace staleness; only successful workspace replacement or report takeover can.",
      "check": "Center failure-policy and workspace classification tests.",
      "flows": [
        "preview",
        "center-workspace",
        "center-recovery"
      ]
    },
    {
      "n": "32",
      "title": "Subscribe before reading; cancel obsolete center work",
      "ref": "api.Service.SubscribeSessionReports",
      "why": "Capture both signals before the initial read and rearm before forwarding. Retained content changes wake readers, while lifecycle events independently recheck scope. Failed reads and workspace inspection retain their recovery cadence.",
      "check": "First-read and rearm races, report eviction and failed persistence, scope loss without a report change, and fake-clock recovery tests.",
      "flows": [
        "preview",
        "center-recovery"
      ]
    },
    {
      "n": "33",
      "title": "Content cache outlives navigation operations",
      "ref": "ui.Coordinator.suspendCenter",
      "why": "Stopping reads/subscriptions on navigation must not erase usable same-session reports or workspace facts. Cache detached data keyed by dashboard and stable session, clear on scope loss/other selection, and let confirmed empty reads replace history.",
      "check": "Navigation report/workspace fallback, confirmed-empty clearing and selection/scope regressions.",
      "flows": [
        "preview",
        "center-recovery"
      ]
    }
  ],
  "tasks": [
    {
      "q": "Add a new agent CLI",
      "a": "Embed GenericCLIDriver and override workspace trust preparation, native resume arguments or coordinated prompt submission where the CLI differs. Register a constructor; each activation gets a new driver. Antigravity is an example of embedding the generic driver and overriding reasoning effort.",
      "refs": [
        "driver.AgentDriver",
        "driver.GenericCLIDriver",
        "driver.NewBuiltinRegistry",
        "driver.AntigravityCLIDriver"
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
      "a": "Start in internal/ui/terminal/attach.go for human terminal input and display. The coordinator hands it an owned attachment through the attachment capability.",
      "refs": [
        "terminal.Command",
        "ui.scopedAttachCommand",
        "api.Service.AttachInDashboard"
      ],
      "flow": "attach"
    },
    {
      "q": "Add a field to ID cards",
      "a": "Add the storage field, then explicitly choose whether it belongs in editable contracts/drafts. Application saves preserve unedited fields and validate full-record revisions; never round-trip full storage records through the UI.",
      "refs": [
        "config.IDCard",
        "views.CardEditor.Draft"
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
        "ui.Coordinator.initViews",
        "ui.ViewType"
      ],
      "flow": "editcard"
    },
    {
      "q": "Add a message",
      "a": "Declare it in messages.go and route it through the coordinator. Partial data gets its own type; async view results carry dashboard ID and generation. Separate dashboard rename completion from launch naming because their success actions differ.",
      "refs": [
        "ui.Coordinator.handleSpecialMessages",
        "ui.Coordinator.handleSessionMessages",
        "messages.ScopedActionFinishedMsg",
        "messages.DashboardSessionRenamedMsg"
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
      "a": "Dashboard panes use dashboard.go; the entry list, forms and deletion/status controls use dashboard_list.go. Keep the legend visible and Escape safe on return. Session rename reuses Modal and textinput; keep e=Rename visible in scoped 80-column help and bound errors inside the dialog.",
      "refs": [
        "views.Dashboard.handleKey",
        "views.Dashboard.footerText",
        "views.DashboardList.Update",
        "views.DashboardList.renderPage",
        "views.dashboardModalInputWidth",
        "views.Dashboard.updateRename"
      ],
      "sources": {
        "files": [
          "internal/ui/views/dashboard_layout_test.go",
          "internal/ui/views/dashboard_test.go"
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
      "a": "views/constants.go and ui/size_guard.go.",
      "refs": [
        "ui.Coordinator.tooSmall"
      ],
      "flow": "resize"
    },
    {
      "q": "Change refresh behaviour",
      "a": "refresh.go for the commands, handlers.go for applying results, and config.Fingerprint. Capture session epochs with snapshots and reject old epochs before data or fingerprints are applied.",
      "refs": [
        "ui.refresher",
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
        "ui.sessionActions"
      ],
      "flow": "control"
    },
    {
      "q": "Change center updates or fallback behavior",
      "a": "Start with the consumer-owned center port, application report/workspace operations and coordinator state machine. Preserve subscribe-before-read, same-session stale content, workspace cancellation and 30-second recovery. Report content changes and lifecycle scope checks use separate signals.",
      "refs": [
        "ui.centerPort",
        "api.Service.SubscribeSessionReports",
        "ui.Coordinator.handleCenter"
      ],
      "flow": "preview"
    },
    {
      "q": "Change how IDE or terminal windows open",
      "a": "Capture settings and scope, check session membership, then launch the configured command in the saved WorkingDir. Completion stays tied to its originating view.",
      "refs": [
        "ui.sessionActions.launchWorkspace",
        "session.LaunchWorkspaceCommand"
      ],
      "flow": "terminal",
      "sources": {
        "files": [
          "internal/ui/terminal_launch_test.go",
          "internal/session/ide_launcher_test.go"
        ]
      }
    },
    {
      "q": "Change sub workspace copying or collision handling",
      "a": "Read sub_workspaces.go: the orchestrator holds one mutex while helpers resolve sources, load ownership metadata, decide reuse, stage/publish and persist. Rollback only the current transaction’s paths. Initial launch captures selected descriptors under the same gate.",
      "refs": [
        "session.Manager.prepareSubWorkspaces",
        "session.shouldCopySubWorkspace",
        "session.subWorkspaceTransaction.rollback",
        "session.Manager.prepareSubWorkspacesSnapshot"
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
      "a": "For existing-session e rename, start at Dashboard.handleEdit/updateRename, the captured coordinator callback and its dedicated completion message. Reuse RenameDashboardSession and SetNameDetailedChecked; refresh on either result and preserve process identity. Launch naming keeps its separate attach flow. For operation snapshots, keep inspectionFromSnapshot as the single conversion. Launch naming retains focus by ID and confirmed name; inline rename invalidates older session refreshes.",
      "refs": [
        "views.Dashboard.handleEdit",
        "views.Dashboard.updateRename",
        "ui.Coordinator.renameDashboardSession",
        "messages.DashboardSessionRenamedMsg",
        "api.Service.RenameDashboardSession",
        "session.Manager.SetNameDetailedChecked",
        "api.inspectionFromSnapshot"
      ],
      "flow": "rename-save"
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
      "a": "Start with preparedDelivery for publication, cancellation, release and submission shared by ordinary sends and notifications. PromptEnvelope and formatPrompt distinguish task and wake-up wording. Keep report kinds and the deployed puppet skill consistent; preserve driver input serialization and Codex/Claude Enter timing.",
      "refs": [
        "api.preparedDelivery",
        "driver.PromptEnvelope",
        "driver.formatPrompt",
        "session.MessageStore.RecordReply",
        "session.Manager.SubmitMessage"
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
      "a": "Use the bounded OperationRunner and its explicit task authorities. Notification scheduling retries admission without involving the worker; accepted attempts are reconciled from byte evidence. Preserve no-replay recovery and dispatcher-before-queue shutdown.",
      "refs": [
        "api.OperationRunner",
        "api.OperationRunner.dispatchNotifications",
        "api.OperationRunner.reconcileNotification",
        "session.MessageStore.Recover"
      ],
      "flow": "notification-retry"
    },
    {
      "q": "Change dashboard ownership or shared-card availability",
      "a": "Start in the shared service membership inventory and persisted caller policy. Preserve immutable session membership, card creation ownership and common UI/MCP access.",
      "refs": [
        "api.membershipInventory",
        "api.persistedDashboardScope",
        "config.Dashboard",
        "config.IDCard"
      ],
      "flow": "caller-scope"
    },
    {
      "q": "Change card or dashboard deletion recovery",
      "a": "Keep consent, admission closure, exact pending cleanup paths and per-effect durable checkpoints together. Delete/Backspace resumes the existing cascade.",
      "refs": [
        "api.Service.ConfirmCardDeletion",
        "api.Service.runDashboardDeletion",
        "views.DashboardList.retryDeletion"
      ],
      "flow": "cascade-recovery"
    },
    {
      "q": "Change configuration startup",
      "a": "Use configstartup for the root lock, fresh initialization and strict current-record validation. Runtime does not convert old formats.",
      "refs": [
        "configstartup.AcquireLock",
        "configstartup.PrepareStartup",
        "configstartup.ValidateLayout"
      ],
      "flow": "startup"
    },
    {
      "q": "Change when a report wakes its orchestrator",
      "a": "RecordReply atomically selects pending or not_required. The original message fixes the destination; the dispatcher owns subsequent work. Update the MCP descriptions and deployed puppet instructions with any kind or protocol change.",
      "refs": [
        "session.MessageStore.RecordReply",
        "session.ReplyNotification",
        "mcp.addMessagingTools"
      ],
      "flow": "reports"
    },
    {
      "q": "Change report-notification retry timing or fairness",
      "a": "Keep one owned dispatcher and the existing operation queue. dueNotifications preserves per-recipient report order while allowing other recipients to proceed. nextNotificationDelay caps backoff at 30 seconds; accepted evidence is reconciled separately.",
      "refs": [
        "api.OperationRunner.dispatchNotifications",
        "api.dueNotifications",
        "api.nextNotificationDelay",
        "api.OperationRunner.reconcileNotification"
      ],
      "flow": "notification-retry"
    },
    {
      "q": "Change notification authorization or lifecycle handling",
      "a": "Retain the report-derived destination and exact original activation. Do not borrow controller credentials or add a general worker send permission. Recheck history, dashboard and target at reservation and the driver write boundary.",
      "refs": [
        "api.notificationAuthority",
        "session.Manager.ReserveReportNotification",
        "session.Manager.reportAdmissionValidLocked",
        "driver.PromptEnvelope"
      ],
      "flow": "notification-authority"
    },
    {
      "q": "Change MCP registry management",
      "a": "MCPManager owns list/modal state; MCPForm owns edited fields. Coordinator commands capture request scope and target IDs, then call the repository. Preserve built-in and saved-card reference guards, metadata overlays and stale-visit rejection.",
      "refs": [
        "views.MCPManager",
        "views.MCPForm",
        "ui.Coordinator.mcpCallbacks",
        "config.FileRepository.DeleteMCP"
      ],
      "flow": "mcp-open"
    },
    {
      "q": "Change launch selection or stale-refresh protection",
      "a": "The coordinator advances session epochs on launch and rename. Dashboard focus waits for the exact ID/name, clears only a hiding session search, and is consumed once. SessionList preserves selected ID across reordering.",
      "refs": [
        "ui.Coordinator.hasStaleSessionSnapshot",
        "views.Dashboard.FocusSession",
        "views.Dashboard.FocusSessionAfterRename",
        "components.SessionList.refreshItems"
      ],
      "flow": "launch"
    },
    {
      "q": "Change workspace inspection limits or classification",
      "a": "Read the saved workspace through the API. Git inspection assembles a private data directory, adds the referenced split index, and falls back visibly on Git errors or unsupported configurations.",
      "refs": [
        "api.Service.ReadSessionWorkspace",
        "api.inspectWorkspace",
        "api.listWorkspace",
        "api.inspectGitWorkspace",
        "api.linkSharedGitIndex"
      ],
      "flow": "center-workspace"
    },
    {
      "q": "Change workspace discovery or initial launch choices",
      "a": "For MCP start with ListWorkspaces and CreateSessionWithOptions. For the dashboard start with ReadLaunchPreparation and the shared workspace fields. Apply IDs only at new session launch and keep relaunch tied to stored workspace paths.",
      "refs": [
        "api.Service.ListWorkspaces",
        "api.Service.CreateSessionWithOptions",
        "api.Service.ReadLaunchPreparation",
        "views.workspaceFields"
      ],
      "flow": "mcp-workspace-discovery"
    }
  ],
  "storage": {
    "root": "~/.config/kankuro/",
    "tree": [
      [
        "layout-version",
        "completed current-format marker"
      ],
      [
        ".dashboard-layout.lock",
        "exclusive configuration-root writer lock"
      ],
      [
        "dashboards.yaml",
        "dashboard identity, description and state"
      ],
      [
        "card_tombstones.yaml",
        "deleted-card identities and pending cleanup"
      ],
      [
        "dashboard_tombstones.yaml",
        "deleted dashboard identities"
      ],
      [
        "dashboard_deletions/<id>.json",
        "resumable cascade manifest"
      ],
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
        "durable lifecycle and notification delivery operations"
      ],
      [
        "messaging/history.json",
        "bounded messages, reports, notification attempts and dedup receipts"
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
      "o and t work only with Sessions focused and a selected session. Both use session.WorkingDir, including temporary copies; live and terminated sessions qualify.",
      "TerminalCommand (terminal_command) defaults empty: users choose an emulator command that opens a separate window. Both commands substitute $PATH; quote \"$PATH\" for spaces. Commands execute directly with the workspace as Dir and no implicit shell expansion.",
      "External application completion means process startup, not eventual exit. No command is inferred; launch failures use the normal error banner without navigation.",
      "Codex trust is stored under ~/.codex/config.toml projects[path].trust_level; Claude trust uses ~/.claude.json projects[path].hasTrustDialogAccepted. Exact temporary-path trust is removed on deletion; a config cleanup failure retains the session for repair/retry.",
      "Temporary workspaces isolate MCP configs. Sessions sharing an original workspace intentionally accept last-writer server/credential identity. Older cleanup removes only its matching credential.",
      "Message history is bounded by 64 MiB or 10,000 whole messages per dashboard. Reports are capped at 16 KiB and 128/message; reply pages cap at 100 reports/256 KiB, with at most ten-second waits. Compact send receipts outlive content eviction.",
      "Startup phase is runtime-only. Sessions restore terminated; bindings and reports persist; unfinished operation effects are interrupted and message submissions become uncertain.",
      "Cards and sessions require schema 2 and explicit dashboard membership. Global capability/workspace registries remain shared.",
      "Use separate dedicated config and temporary-workspace directories. An empty config root initializes directly; unrelated artifacts in an uninitialized directory cause startup rejection.",
      "Configuration locking serializes writers for one config root; applications using different roots can still target a shared workspace.",
      "Confirmed shared-card deletion removes availability globally but preserves foreign sessions, workspace paths and saved bindings.",
      "Dashboard deletion removes owned sessions and owned cards after foreign-impact consent; each incomplete effect remains retryable after restart.",
      "Notification metadata lives with its reply, and child messages retain the correlation and byte evidence. Statuses are pending, accepted, submitted, uncertain, unavailable and not_required.",
      "Older reports without notification metadata remain readable and do not generate retrospective wake-ups. New/resumed activations never inherit the original activation’s notifications."
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
    "Claude now shares Codex’s 500 ms text/Enter separation. The registry-dispatched cancellation regression covers both overrides. The user reports this workaround succeeded on their Claude environment; no local real-Claude run is claimed for this refresh.",
    "Keyboard regression coverage checks Delete/Backspace retries and repeated Escape after navigating back. Actual terminal permission failure/recovery was verified.",
    "Messaging regression tests use isolated generic CLI/PTY fixtures and the real MCP service. Report-kind persistence tests cover all four kinds; integration tests observe the original controller prompt without granting workers Dashboard Control.",
    "The shared-delivery suite exercises queue rejection, journal rollback, cancellation and stable replay through ordinary sends and notification entry points. Notification reconciliation must never turn uncertain bytes into an automatic retry.",
    "Project-owned terminal tests use the tui-test Go binding behind //go:build e2e. Run explicitly for Appedia/Appediff; normal tests and evidence collection create no screenshots.",
    "The corrected OpenCode fixture verifies supported credential export/import, isolates service configuration and selects openrouter/openrouter/free. Its actual result is failed at register. The older Big Pickle quota screen came from an incomplete fixture and does not describe the configured user environment.",
    "Session rename is covered with ordinary Go tests: cancellation/scope changes discard drafts without saving; the coordinator path refreshes without attachment or process replacement and restores the saved name through a fresh manager. Native-agent E2E is not required for these assertions.",
    "Current ordinary-suite statement coverage is collected with atomic instrumentation and matching source/run provenance; it does not execute E2E-tagged tests. The 2026-11-09 profile is archived with its original metadata and is not a paired baseline for a coverage delta.",
    "Model-driven E2E runs require a provider/model choice and total token budget covering retries, per tests/e2e/AGENTS.md. Ordinary tests and model-free screenshot preflight need no model call.",
    "Current MCP and startup-only launch-focus E2E journeys pass with the upgraded UI dependencies. They capture actual terminal checkpoints, without model tasks or pixel-baseline comparisons.",
    "Step 6 center source assertions are linked with exact source hashes and explicit limits. This Appedia refresh does not rerun application tests or claim feature approval; changed historical assessments remain unassessed."
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
      "Four rows reserved for content views; the dashboard compensates to fill the terminal, while the dashboard list receives full terminal height."
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
    ],
    [
      "Dashboard membership",
      "Immutable session creation dashboard; the basis of both human checks and authenticated caller scope."
    ],
    [
      "Shared card",
      "A card usable across dashboards whose creation owner alone can edit or delete it."
    ],
    [
      "View generation",
      "Counter captured with dashboard identity to reject late asynchronous UI results."
    ],
    [
      "Tombstone",
      "Retained deleted identity and recovery state, without executable settings."
    ],
    [
      "Config startup",
      "Root locking, direct fresh initialization and current-format validation before the app serves UI or MCP."
    ],
    [
      "Report notification",
      "A durable, correlated wake-up to the task sender’s exact activation; it requests no report back."
    ],
    [
      "Delivery attempt",
      "One stable notification request/message identity; only conclusive zero-byte evidence allows advancement."
    ]
  ],
  "tests": {
    "scope": "Workspace feature source e741572. The ordinary Linux suite and model-free capture harness passed. New workspace evidence is bound to this source snapshot. Older terminal captures, model runs, and statement coverage retain their original provenance.",
    "features": [
      {
        "id": "workspace-feature-dashboard",
        "title": "Dashboard workspace selection",
        "area": "TUI",
        "summary": "Review the two workspace rows, picker and initial launch confirmation.",
        "scenarios": [
          {
            "id": "confirm",
            "title": "Confirm the displayed launch",
            "dimension": "behavior",
            "summary": "Confirm the form and use its displayed IDs even after a card edit.",
            "status": "tested",
            "flows": [
              "launch-workspace-selection"
            ],
            "evidence": [
              "workspace-feature-display",
              "workspace-feature-launch-flow",
              "workspace-feature-warning"
            ],
            "limit": "The real local process uses a generic driver."
          },
          {
            "id": "edit",
            "title": "Navigate and edit selections",
            "dimension": "interaction",
            "summary": "Change roles, clear sub-workspaces and confirm from the form.",
            "status": "tested",
            "flows": [
              "sub-picker"
            ],
            "evidence": [
              "workspace-feature-draft",
              "workspace-feature-promote",
              "workspace-feature-editor-promote",
              "workspace-feature-picker-search"
            ],
            "limit": "The tests exercise models and callbacks."
          },
          {
            "id": "escape",
            "title": "Cancel, quit or visit sources",
            "dimension": "interaction",
            "summary": "Cancel before submission and ignore scoped navigation that arrives late.",
            "status": "tested",
            "flows": [
              "launch-workspace-navigation"
            ],
            "evidence": [
              "workspace-feature-cancel",
              "workspace-feature-sources",
              "workspace-feature-sources-submit",
              "workspace-feature-quit",
              "workspace-feature-picker-quit"
            ],
            "limit": "The assertions drive coordinator and view messages."
          },
          {
            "id": "refresh",
            "title": "Recover from an option read failure",
            "dimension": "interaction",
            "summary": "Keep cached choices and show a scoped error.",
            "status": "tested",
            "flows": [
              "workspace-choices-refresh"
            ],
            "evidence": [
              "workspace-feature-cached"
            ],
            "limit": "The failed read is injected."
          },
          {
            "id": "minimum",
            "title": "Minimum terminal layout",
            "dimension": "layout",
            "summary": "The shared picker and launch draft are exercised at small sizes.",
            "status": "partial",
            "flows": [
              "sub-picker"
            ],
            "evidence": [
              "workspace-feature-draft",
              "workspace-feature-picker-size"
            ],
            "limit": "A native terminal capture of the entire workspace-selection interaction was not collected."
          }
        ]
      },
      {
        "id": "workspace-feature-workspace-discovery",
        "title": "Discover workspaces through MCP",
        "area": "MCP",
        "summary": "Resolve an authorized caller and return the global registry for initial launch selection.",
        "scenarios": [
          {
            "id": "global",
            "title": "List the same registry across dashboards",
            "dimension": "behavior",
            "summary": "Authorized callers see unreferenced entries while card and session reads remain scoped.",
            "status": "tested",
            "flows": [
              "mcp-workspace-discovery"
            ],
            "evidence": [
              "workspace-feature-workspace-scopes",
              "workspace-feature-transport"
            ],
            "limit": "The comparison uses two fixture dashboards and a local registry."
          },
          {
            "id": "authorization",
            "title": "Enforce deployed Dashboard Control",
            "dimension": "behavior",
            "summary": "Ordinary, revoked and invalid credentials cannot read the registry.",
            "status": "tested",
            "flows": [
              "mcp-workspace-discovery"
            ],
            "evidence": [
              "workspace-feature-workspace-service",
              "workspace-feature-transport"
            ],
            "limit": "Service probes count reads and the HTTP lifecycle test exercises the live caller without the grant."
          },
          {
            "id": "registry",
            "title": "Handle ordered, empty and failed reads",
            "dimension": "behavior",
            "summary": "Order by name and ID, project only the intended fields and distinguish empty results from read errors.",
            "status": "tested",
            "flows": [
              "mcp-workspace-discovery"
            ],
            "evidence": [
              "workspace-feature-workspace-service",
              "workspace-feature-workspace-registration",
              "workspace-feature-workspace-error"
            ],
            "limit": "The service probe injects read results and failures. The transport fixtures check the JSON empty array and error response."
          }
        ]
      },
      {
        "id": "workspace-feature-mcp-create",
        "title": "MCP create and replay",
        "area": "MCP",
        "summary": "Resolve the ordered workspace selection in an authorized create operation.",
        "scenarios": [
          {
            "id": "validate",
            "title": "Validate and replay a create request",
            "dimension": "behavior",
            "summary": "Reject invalid shapes and conflict on changed retry arguments.",
            "status": "tested",
            "flows": [
              "mcp-workspace-create"
            ],
            "evidence": [
              "workspace-feature-transport",
              "workspace-feature-retry",
              "workspace-feature-admission"
            ],
            "limit": "The live HTTP fixture exercises lifecycle and request validation with a local generic agent."
          },
          {
            "id": "multi",
            "title": "Create with multiple explicit IDs",
            "dimension": "behavior",
            "summary": "Discover unreferenced IDs through MCP, then launch with a different main workspace and two ordered sub-workspace copies.",
            "status": "tested",
            "flows": [
              "mcp-workspace-discovery",
              "mcp-workspace-create",
              "workspace"
            ],
            "evidence": [
              "workspace-feature-transport"
            ],
            "limit": "The real HTTP fixture uses a local test process and temporary files. It checks the copied source, default exclusion, exact replay and reordered-ID conflict."
          }
        ]
      },
      {
        "id": "workspace-feature-workspace-state",
        "title": "Workspace state across activations",
        "area": "Sessions",
        "summary": "Prepare selected copies and retain saved workspace identity for relaunch.",
        "scenarios": [
          {
            "id": "copies",
            "title": "Prepare copies and preserve shared ownership",
            "dimension": "behavior",
            "summary": "Record only the copies selected by this launch and keep shared manifest ownership.",
            "status": "tested",
            "flows": [
              "workspace"
            ],
            "evidence": [
              "workspace-feature-manifest",
              "workspace-feature-copy",
              "workspace-feature-cleanup"
            ],
            "limit": "Manager tests exercise filesystem behavior directly."
          },
          {
            "id": "restart",
            "title": "Relaunch in the saved directory",
            "dimension": "behavior",
            "summary": "Keep the overridden working path and deployed state across activation.",
            "status": "tested",
            "flows": [
              "resume"
            ],
            "evidence": [
              "workspace-feature-relaunch-override",
              "workspace-feature-relaunch-baseline",
              "workspace-feature-transport"
            ],
            "limit": "Evidence covers the saved overridden main directory and retained deployed copy behavior in separate tests."
          }
        ]
      },
      {
        "id": "boundary",
        "title": "Architecture enforcement",
        "area": "Roadmap verification",
        "summary": "Checks package direction and reviewed exception sets.",
        "scenarios": [
          {
            "id": "assertion",
            "title": "Architecture enforcement",
            "dimension": "behavior",
            "summary": "Checks package direction and reviewed exception sets.",
            "status": "tested",
            "flows": [
              "startup"
            ],
            "evidence": [
              "boundary"
            ],
            "limit": "The checker is static. Runtime behavior has separate evidence."
          }
        ]
      },
      {
        "id": "card-save",
        "title": "Preserve hidden card fields",
        "area": "Roadmap verification",
        "summary": "Checks retained fields and rejection of a stale revision.",
        "scenarios": [
          {
            "id": "assertion",
            "title": "Preserve hidden card fields",
            "dimension": "behavior",
            "summary": "Checks retained fields and rejection of a stale revision.",
            "status": "tested",
            "flows": [
              "savecard"
            ],
            "evidence": [
              "card-save"
            ],
            "limit": "API and repository seam."
          }
        ]
      },
      {
        "id": "card-race",
        "title": "Concurrent card saves",
        "area": "Roadmap verification",
        "summary": "Only one of two saves with the same revision succeeds.",
        "scenarios": [
          {
            "id": "assertion",
            "title": "Concurrent card saves",
            "dimension": "behavior",
            "summary": "Only one of two saves with the same revision succeeds.",
            "status": "tested",
            "flows": [
              "savecard"
            ],
            "evidence": [
              "card-race"
            ],
            "limit": "Concurrent service updates in a controlled fixture."
          }
        ]
      },
      {
        "id": "source-edit",
        "title": "Reject a reordered source",
        "area": "Roadmap verification",
        "summary": "Rejects stale row identity after reorder and addresses duplicate rows.",
        "scenarios": [
          {
            "id": "assertion",
            "title": "Reject a reordered source",
            "dimension": "behavior",
            "summary": "Rejects stale row identity after reorder and addresses duplicate rows.",
            "status": "tested",
            "flows": [
              "source"
            ],
            "evidence": [
              "source-edit"
            ],
            "limit": "Backend source mutation path."
          }
        ]
      },
      {
        "id": "conversation",
        "title": "Revalidate scope and conversation",
        "area": "Roadmap verification",
        "summary": "Checks scoped action revalidation and stale conversation choice handling.",
        "scenarios": [
          {
            "id": "assertion",
            "title": "Revalidate scope and conversation",
            "dimension": "behavior",
            "summary": "Checks scoped action revalidation and stale conversation choice handling.",
            "status": "tested",
            "flows": [
              "resume"
            ],
            "evidence": [
              "conversation"
            ],
            "limit": "Controlled backend session fixture."
          }
        ]
      },
      {
        "id": "center-history",
        "title": "Render retained reports after reconstruction",
        "area": "Roadmap verification",
        "summary": "Reads a retained report through a reconstructed manager and coordinator.",
        "scenarios": [
          {
            "id": "assertion",
            "title": "Render retained reports after reconstruction",
            "dimension": "behavior",
            "summary": "Reads a retained report through a reconstructed manager and coordinator.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "center-history"
            ],
            "limit": "Exercises backend/UI seam without rendering a live terminal restart."
          }
        ]
      },
      {
        "id": "center-switch",
        "title": "Reject late center reads",
        "area": "Roadmap verification",
        "summary": "Changes selection during a read and checks stale data cannot populate the new center.",
        "scenarios": [
          {
            "id": "assertion",
            "title": "Reject late center reads",
            "dimension": "behavior",
            "summary": "Changes selection during a read and checks stale data cannot populate the new center.",
            "status": "tested",
            "flows": [
              "center-recovery"
            ],
            "evidence": [
              "center-switch"
            ],
            "limit": "Controlled coordinator timing."
          }
        ]
      },
      {
        "id": "pty-detach",
        "title": "Restore terminal after detach",
        "area": "Roadmap verification",
        "summary": "Checks resize, legacy and Kitty detach, and equality of terminal state before and after.",
        "scenarios": [
          {
            "id": "assertion",
            "title": "Restore terminal after detach",
            "dimension": "interaction",
            "summary": "Checks resize, legacy and Kitty detach, and equality of terminal state before and after.",
            "status": "tested",
            "flows": [
              "attach"
            ],
            "evidence": [
              "pty-detach"
            ],
            "limit": "Linux PTY fixture."
          }
        ]
      },
      {
        "id": "pty-split",
        "title": "Keep a split detach sequence local",
        "area": "Roadmap verification",
        "summary": "Feeds the detach bytes in two reads and checks that only preceding ordinary bytes were forwarded.",
        "scenarios": [
          {
            "id": "assertion",
            "title": "Keep a split detach sequence local",
            "dimension": "interaction",
            "summary": "Feeds the detach bytes in two reads and checks that only preceding ordinary bytes were forwarded.",
            "status": "tested",
            "flows": [
              "attach"
            ],
            "evidence": [
              "pty-split"
            ],
            "limit": "Deterministic decoder test. The full binary run corroborates the behavior."
          }
        ]
      },
      {
        "id": "missing-attach",
        "title": "Close an obsolete attachment",
        "area": "Roadmap verification",
        "summary": "Checks immediate close and rejection after full and sessions-only refresh confirm absence.",
        "scenarios": [
          {
            "id": "assertion",
            "title": "Close an obsolete attachment",
            "dimension": "behavior",
            "summary": "Checks immediate close and rejection after full and sessions-only refresh confirm absence.",
            "status": "tested",
            "flows": [
              "attach-disposal"
            ],
            "evidence": [
              "missing-attach"
            ],
            "limit": "Coordinator fixture with controlled ownership."
          }
        ]
      },
      {
        "id": "mcp-scope",
        "title": "Keep dashboard scope through HTTP",
        "area": "Roadmap verification",
        "summary": "Exercises dashboard scope through the MCP server transport.",
        "scenarios": [
          {
            "id": "assertion",
            "title": "Keep dashboard scope through HTTP",
            "dimension": "behavior",
            "summary": "Exercises dashboard scope through the MCP server transport.",
            "status": "tested",
            "flows": [
              "control"
            ],
            "evidence": [
              "mcp-scope"
            ],
            "limit": "Local transport and test identities."
          }
        ]
      },
      {
        "id": "delete",
        "title": "Retain deletion progress after read failure",
        "area": "Roadmap verification",
        "summary": "Checks execution progress stays visible when a subsequent status read fails.",
        "scenarios": [
          {
            "id": "assertion",
            "title": "Retain deletion progress after read failure",
            "dimension": "behavior",
            "summary": "Checks execution progress stays visible when a subsequent status read fails.",
            "status": "tested",
            "flows": [
              "deletion-outcomes"
            ],
            "evidence": [
              "delete"
            ],
            "limit": "Coordinator and view fixture."
          }
        ]
      },
      {
        "id": "runtime-gaps",
        "title": "Runtime coverage gaps",
        "area": "Verification limits",
        "summary": "The budget-limited recall and verification beyond Linux remain open.",
        "scenarios": [
          {
            "id": "platform",
            "title": "Terminal behavior outside Linux",
            "dimension": "interaction",
            "summary": "Run native lifecycle and terminal restoration on other supported host platforms.",
            "status": "unassessed",
            "flows": [
              "attach",
              "resize"
            ],
            "limit": "All captured runtime evidence is from Linux."
          }
        ]
      },
      {
        "id": "codex-token-free",
        "title": "Codex startup and session lifecycle",
        "area": "Linux terminal E2E",
        "summary": "Real Codex CLI in a disposable workspace with a dummy credential and loopback model endpoint. The terminal harness never submits a prompt.",
        "scenarios": [
          {
            "id": "idle-attach",
            "title": "Launch, name and reattach to Codex",
            "dimension": "interaction",
            "summary": "Ctrl+K returns to the dashboard. Reattach keeps the created process and session selected.",
            "status": "tested",
            "flows": [
              "launch-focus",
              "attach"
            ],
            "evidence": [
              "tui-codex",
              "tui-launch-focus"
            ],
            "captures": [
              "tui-codex-attached",
              "tui-codex-dashboard",
              "tui-codex-reattached"
            ],
            "limit": "No prompt was submitted."
          },
          {
            "id": "resize",
            "title": "Reject undersized terminal and recover",
            "dimension": "layout",
            "summary": "At 60×20, the size warning appears. Returning to 120×40 restores the dashboard.",
            "status": "tested",
            "flows": [
              "resize"
            ],
            "evidence": [
              "tui-codex"
            ],
            "captures": [
              "tui-codex-minimum-size"
            ],
            "limit": "Checks two terminal sizes on Linux."
          },
          {
            "id": "workspace-refresh",
            "title": "Refresh non-Git workspace facts",
            "dimension": "behavior",
            "summary": "A new file appears in the center after the initial directory listing.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "tui-codex"
            ],
            "captures": [
              "tui-codex-workspace-initial",
              "tui-codex-workspace-updated"
            ],
            "limit": "Uses the actual Codex workspace and waits for the live refresh."
          },
          {
            "id": "stop-relaunch",
            "title": "Relaunch the stopped Codex session",
            "dimension": "interaction",
            "summary": "The same Kankuro session ID and working directory acquire a new Codex PID.",
            "status": "tested",
            "flows": [
              "stop-delete",
              "resume"
            ],
            "evidence": [
              "tui-codex"
            ],
            "captures": [
              "tui-codex-stopped",
              "tui-codex-relaunched"
            ],
            "limit": "Relaunch uses a new conversation. Native conversation recall is separate."
          },
          {
            "id": "report-restart",
            "title": "Reconstruct the selected session’s authored report",
            "dimension": "behavior",
            "summary": "The selected author has a seeded report after Codex relaunch and Kankuro restart.",
            "status": "tested",
            "flows": [
              "preview",
              "center-recovery"
            ],
            "evidence": [
              "tui-codex"
            ],
            "captures": [
              "tui-codex-report-after-relaunch",
              "tui-codex-restart-report"
            ],
            "limit": "Seeded a report through the real store. No live report tool call was made."
          },
          {
            "id": "git-refresh",
            "title": "Refresh clean and changed Git state",
            "dimension": "behavior",
            "summary": "The center shows a clean tree, then an untracked file after it is created.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "tui-codex-git"
            ],
            "captures": [
              "tui-codex-git-workspace-initial",
              "tui-codex-git-workspace-updated"
            ],
            "limit": "The live terminal checks an untracked file. Rename and malformed Git parsing remain backend assertions."
          },
          {
            "id": "report-relaunch",
            "title": "Preserve reports through Codex relaunch",
            "dimension": "behavior",
            "summary": "The same stopped report author starts a new Codex process. Its existing report remains visible after detach and Kankuro restart.",
            "status": "tested",
            "flows": [
              "preview",
              "resume"
            ],
            "evidence": [
              "tui-codex-git"
            ],
            "captures": [
              "tui-codex-git-report-after-relaunch",
              "tui-codex-git-restart-report"
            ],
            "limit": "A previous report is seeded and retained by session ID. No model was used to author a new report."
          }
        ]
      },
      {
        "id": "mcp-terminal-registry",
        "title": "MCP configuration screens",
        "area": "Linux terminal E2E",
        "summary": "The project’s model-free tui-test journey drives the registry screens and checks persistence.",
        "scenarios": [
          {
            "id": "registry-edit",
            "title": "Create and edit a custom MCP",
            "dimension": "interaction",
            "summary": "The custom MCP is created, renamed and retains its command.",
            "status": "tested",
            "flows": [
              "mcp",
              "mcp-open"
            ],
            "evidence": [
              "tui-mcp"
            ],
            "captures": [
              "tui-mcp-list",
              "tui-mcp-new",
              "tui-mcp-edit",
              "tui-mcp-edited-list"
            ],
            "limit": "Uses a placeholder command and never starts it."
          },
          {
            "id": "registry-delete",
            "title": "Cancel and confirm MCP deletion",
            "dimension": "interaction",
            "summary": "Cancel retains the custom entry. Confirmation removes it and preserves the protected built-in.",
            "status": "tested",
            "flows": [
              "mcp-delete"
            ],
            "evidence": [
              "tui-mcp"
            ],
            "captures": [
              "tui-mcp-confirm",
              "tui-mcp-deleted"
            ],
            "limit": "Tests one custom server and the built-in at 80×24."
          }
        ]
      },
      {
        "id": "luna-model-journey",
        "title": "Live Codex Luna registration, reporting and recall",
        "area": "Linux terminal E2E",
        "summary": "The opt-in token-usage journey exercises the real Codex CLI with GPT-6 Luna and an authenticated local Kankuro MCP.",
        "scenarios": [
          {
            "id": "registration",
            "title": "Register the actual Codex conversation",
            "dimension": "behavior",
            "summary": "Kankuro persisted a named Codex binding for the selected session. The rollout recorded the first completed turn.",
            "status": "tested",
            "flows": [
              "registration"
            ],
            "evidence": [
              "tui-luna"
            ],
            "limit": "One real Luna conversation. Relaunch with the saved binding remains unverified."
          },
          {
            "id": "live-report",
            "title": "Show an authored progress report in the center",
            "dimension": "interaction",
            "summary": "The first live progress report replaced the workspace directory view. MCP verified the report text, kind, correlation and author.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "tui-luna"
            ],
            "captures": [
              "tui-luna-workspace-before-report",
              "tui-luna-live-report"
            ],
            "limit": "One first report on Linux. The run ended before stop or relaunch."
          },
          {
            "id": "recall",
            "title": "Resume and recall context",
            "dimension": "interaction",
            "summary": "Choose the real saved conversation after stopping Codex and verify recall of the code from its earlier turn.",
            "status": "partial",
            "flows": [
              "resume",
              "attach"
            ],
            "evidence": [
              "tui-luna"
            ],
            "limit": "The binding was persisted. The first turn exceeded the 60,000 total input threshold, so this run ended before resume and recall.",
            "priority": "First",
            "suggestion": "Bound model requests between tool exchanges and capture sanitized per-request usage before another live run.",
            "reason": "One interactive turn used 129,209 input tokens before the total check could stop the test.",
            "captures": [
              "tui-luna-input-overage"
            ]
          }
        ]
      },
      {
        "id": "configuration-startup",
        "title": "Configuration startup",
        "area": "Runtime",
        "summary": "Initialize an empty root and reject unsupported or incomplete layouts.",
        "scenarios": [
          {
            "id": "fresh-root",
            "title": "Initialize fresh state and reject a partial root",
            "dimension": "behavior",
            "summary": "Checks that fresh startup validates and that a partial root fails with the missing-marker diagnostic.",
            "status": "tested",
            "flows": [
              "startup"
            ],
            "evidence": [
              "fresh-initialization-and-partial-root"
            ],
            "limit": "Calls configuration startup directly; does not start the full TUI/MCP composition root."
          },
          {
            "id": "invalid-records",
            "title": "Reject bad membership without rewriting records",
            "dimension": "behavior",
            "summary": "Validates a known-good fixture before each corrupt session/operation/message/cursor/card record, then checks rejection and exact original bytes.",
            "status": "tested",
            "flows": [
              "startup"
            ],
            "evidence": [
              "validate-layout-rejects-invalid-records-without-rewriting"
            ],
            "limit": "Five explicit fixture cases; not arbitrary schema fuzzing."
          },
          {
            "id": "exclusive-writer",
            "title": "Refuse a second writer and release the lock",
            "dimension": "behavior",
            "summary": "A second AcquireLock fails while held and succeeds after Close.",
            "status": "tested",
            "flows": [
              "startup"
            ],
            "evidence": [
              "exclusive-lock"
            ],
            "limit": "Same-process lock contention on Linux; no cross-platform or subprocess assertion."
          },
          {
            "id": "puppet-sync-order",
            "title": "Finish puppet sync before loading the dashboard",
            "dimension": "behavior",
            "summary": "Read the first registry baseline after the puppet registry is ready. A missing source is a warning and the dashboard still opens with valid artwork.",
            "status": "tested",
            "flows": [
              "startup",
              "tick"
            ],
            "evidence": [
              "startup-puppet-ordered",
              "startup-puppet-terminal",
              "refresh-token-retry"
            ],
            "limit": "Linux terminal observation and targeted unit assertions. Source latency was not measured."
          }
        ]
      },
      {
        "id": "dashboard-management",
        "title": "Dashboard forms and deletion feedback",
        "area": "UI forms",
        "summary": "Create and edit dashboard details and present bounded deletion/recovery errors.",
        "scenarios": [
          {
            "id": "dashboard-details",
            "title": "Create and edit dashboard details with stable identity",
            "dimension": "interaction",
            "summary": "The create form validates name and callback payload; the service test checks create/update inventory equality and immutable identity.",
            "status": "partial",
            "flows": [
              "dashboard-details"
            ],
            "evidence": [
              "dashboard-list-creates-named-dashboard",
              "dashboard-details-through-service"
            ],
            "limit": "No direct retained assertion was found for e-driven edit form, selection preservation across SetDashboards reorder, or the whole callback/generation/save-refresh journey.",
            "sources": {
              "refs": [
                "views.DashboardList.Update",
                "views.DashboardList.updateForm",
                "ui.Coordinator.dashboardListCallbacks",
                "api.Service.CreateDashboard",
                "config.FileRepository.UpdateDashboardDetails"
              ]
            },
            "suggestion": "Select a dashboard, edit with e, save and reorder/refresh inventory; assert name/description, immutable selected ID and scoped completion.",
            "priority": "Next",
            "reason": "Name sorting changes can move the cursor to a different dashboard even when saved IDs remain valid.",
            "previousStatus": "partial"
          },
          {
            "id": "list-errors",
            "title": "Display and clear deletion preparation/recovery errors",
            "dimension": "layout",
            "summary": "Direct view assertions cover sanitized bounded preflight/status errors, cancellation/success clearing and long errors when no dashboards remain.",
            "status": "tested",
            "flows": [
              "list-errors"
            ],
            "evidence": [
              "dashboard-list-shows-deletion-preparation-and-status-errors",
              "dashboard-list-bounds-long-deletion-error-when-list-becomes-empty"
            ],
            "limit": "Layout and injected message interaction only; deletion effects are assessed separately.",
            "sources": {
              "refs": [
                "views.DashboardList.Update",
                "ui.Coordinator.dashboardListCallbacks",
                "api.Service.GetDashboardDeletion",
                "views.DashboardList.View"
              ]
            },
            "previousStatus": "tested"
          }
        ]
      },
      {
        "id": "dashboard-ownership",
        "title": "Dashboard ownership and sharing",
        "area": "Policy",
        "summary": "Cards may be shared; sessions retain their dashboard identity.",
        "scenarios": [
          {
            "id": "scoped-inventory",
            "title": "Keep sharing, ownership and deletion scoped",
            "dimension": "behavior",
            "summary": "Checks private/shared visibility, consumer edit denial, dashboard-local sessions, stale/foreign deletion confirmation, surviving foreign sessions and switching away from a deleted card.",
            "status": "tested",
            "flows": [
              "dashboard",
              "savecard",
              "card-delete",
              "deleted-session",
              "caller-scope"
            ],
            "evidence": [
              "read-dashboard-data-scopes-cards-and-sessions"
            ],
            "limit": "Service/repository/manager fixture with cat processes; no keyboard or MCP transport assertions.",
            "previousStatus": "tested"
          },
          {
            "id": "revalidate-launch",
            "title": "Reject a card unshared during launch",
            "dimension": "behavior",
            "summary": "Blocks copying, changes sharing, then asserts no publication and release of the pending reservation.",
            "status": "tested",
            "flows": [
              "launch"
            ],
            "evidence": [
              "scoped-launch-revalidates-changed-card-before-publication"
            ],
            "limit": "Injected workspace-copy gate; does not simulate every registry mutation.",
            "previousStatus": "tested"
          }
        ]
      },
      {
        "id": "dashboard-control",
        "title": "Authenticated dashboard control",
        "area": "Policy",
        "summary": "Authorize dashboard-scoped discovery and lifecycle actions from deployed activation credentials.",
        "scenarios": [
          {
            "id": "control",
            "title": "Authenticate dashboard discovery and lifecycle control",
            "dimension": "behavior",
            "summary": "Retained service and real-HTTP tests assert controller authority, registration-only denial, scoped listings and idempotent scoped creation.",
            "status": "tested",
            "flows": [
              "control"
            ],
            "evidence": [
              "dashboard-scope-over-h-t-t-p",
              "operation-runner-creates-once-and-replays-request",
              "unauthenticated-discovery-is-denied"
            ],
            "limit": "HTTP test directly covers list_sessions; the broader lifecycle authorization checks use the shared service. No arbitrary credential fuzzing.",
            "sources": {
              "refs": [
                "mcp.addDiscoveryTools",
                "api.Service.ResolveCaller",
                "session.Manager.ResolveDashboardCaller",
                "api.Service.GetSession",
                "mcp.addLifecycleTools"
              ]
            }
          }
        ]
      },
      {
        "id": "deletion-recovery",
        "title": "Deletion and restart recovery",
        "area": "Policy",
        "summary": "Durable deletion intent closes admission and resumes cleanup.",
        "scenarios": [
          {
            "id": "pending-launch",
            "title": "Cancel a foreign pending launch before deleting its card",
            "dimension": "behavior",
            "summary": "Asserts foreign pending impact, required confirmation, cancelled launch and no published session.",
            "status": "tested",
            "flows": [
              "card-delete"
            ],
            "evidence": [
              "confirmed-card-deletion-cancels-foreign-pending-launch"
            ],
            "limit": "Controlled copier and service entry points; no confirmation-modal rendering.",
            "previousStatus": "tested"
          },
          {
            "id": "restart-cascade",
            "title": "Resume dashboard deletion after restart",
            "dimension": "behavior",
            "summary": "Persists deleting intent, checks admission denial, rebuilds manager/service, removes owned session while preserving the foreign session and permits idempotent retry.",
            "status": "tested",
            "flows": [
              "delete-dashboard",
              "cascade-recovery"
            ],
            "evidence": [
              "dashboard-cascade-resumes-after-restart"
            ],
            "limit": "In-process restart against real files; no OS kill/power-loss simulation."
          },
          {
            "id": "card-recovery",
            "title": "Retry exact unpublished workspace cleanup after restart",
            "dimension": "behavior",
            "summary": "The recovery regression asserts durable failed-cleanup identity survives card edits and restart, then removes only the captured temporary root and clears its error.",
            "status": "tested",
            "flows": [
              "card-recovery"
            ],
            "evidence": [
              "card-deletion-retries-pending-workspace-cleanup-after-restart"
            ],
            "limit": "Controlled copier and permissions; no general filesystem failure or power-loss claim.",
            "sources": {
              "refs": [
                "session.Manager.ReserveCardDeletion",
                "session.Manager.CancelPendingCardReferences",
                "session.Manager.WaitCardReferences",
                "api.Service.cleanupCardPending",
                "session.Manager.RemoveResidual"
              ]
            },
            "previousStatus": "tested"
          }
        ]
      },
      {
        "id": "dashboard-context",
        "title": "Dashboard context and refresh",
        "area": "UI coordinator",
        "summary": "Reject obsolete asynchronous results and retain independent view contexts.",
        "scenarios": [
          {
            "id": "stale-results",
            "title": "Discard results from a previously opened dashboard",
            "dimension": "behavior",
            "summary": "Checks foreign center rejection, stale refresh/navigation/attach isolation, current cards/sessions and restoration of the original view instance.",
            "status": "tested",
            "flows": [
              "tick",
              "navigation",
              "preview",
              "attach"
            ],
            "evidence": [
              "multi-dashboard-discards-stale-results-and-preserves-views"
            ],
            "limit": "Drives coordinator messages directly with real local cat sessions; not an end-to-end terminal journey.",
            "previousStatus": "tested"
          },
          {
            "id": "fingerprints",
            "title": "Skip reading unchanged registries",
            "dimension": "behavior",
            "summary": "Counts ListIDCards calls and asserts an unchanged fingerprint produces no additional registry read.",
            "status": "tested",
            "flows": [
              "tick"
            ],
            "evidence": [
              "maybe-refresh-config-skips-unchanged-registry"
            ],
            "limit": "Counting repository fixture; does not establish every filesystem timestamp boundary.",
            "previousStatus": "tested"
          }
        ]
      },
      {
        "id": "minimum-terminal",
        "title": "Terminal bounds and keyboard safety",
        "area": "UI layout",
        "summary": "A too-small terminal displays a bounded notice and ignores ordinary action keys.",
        "scenarios": [
          {
            "id": "notice-rectangle",
            "title": "Fit the notice to every tested tiny size",
            "dimension": "layout",
            "summary": "Sweeps widths 1–85 and heights 1–27; checks exact notice rectangle below 80×24 and notice removal at supported sizes.",
            "status": "tested",
            "flows": [
              "resize"
            ],
            "evidence": [
              "coordinator-too-small-screen-is-exact"
            ],
            "limit": "String/cell layout assertions; no emulator font or image comparison."
          },
          {
            "id": "drop-input",
            "title": "Ignore action keys below the minimum",
            "dimension": "interaction",
            "summary": "Checks that n, Enter, Tab, Backspace and c return no commands and do not change the active view at three undersized dimensions.",
            "status": "tested",
            "flows": [
              "resize"
            ],
            "evidence": [
              "coordinator-drops-keys-when-too-small"
            ],
            "limit": "Selected keys and sizes only; paste is not asserted here."
          },
          {
            "id": "explicit-quit",
            "title": "Require q to quit the dashboard list",
            "dimension": "interaction",
            "summary": "Checks repeated Esc/Ctrl+C keep the dashboard list open at normal and small sizes and q returns QuitMsg.",
            "status": "tested",
            "flows": [
              "navigation",
              "quit",
              "resize"
            ],
            "evidence": [
              "dashboard-list-requires-explicit-quit"
            ],
            "limit": "Coordinator-level input assertions; no physical terminal signal delivery."
          }
        ]
      },
      {
        "id": "card-editor",
        "title": "Card editor state and focus",
        "area": "UI forms",
        "summary": "Keep hidden fields and capability selections while supporting explicit controls.",
        "scenarios": [
          {
            "id": "preserve-fields",
            "title": "Edit a name without losing other card fields",
            "dimension": "behavior",
            "summary": "Saves a name-only draft through the editor service, compares every other persisted field apart from UpdatedAt, and rejects a stale revision.",
            "status": "tested",
            "flows": [
              "editcard",
              "savecard"
            ],
            "evidence": [
              "card-editor-preserves-unmanaged-fields"
            ],
            "limit": "Service and repository assertions. The UI draft is covered separately.",
            "previousStatus": "tested"
          },
          {
            "id": "control-toggle",
            "title": "Round-trip Dashboard Control through Space",
            "dimension": "interaction",
            "summary": "Focuses the control row, toggles off/on and checks Card().DashboardControl.",
            "status": "unassessed",
            "flows": [
              "savecard"
            ],
            "evidence": [
              "dashboard-control-card-editor-toggle-round-trip"
            ],
            "limit": "Does not prove deployed activation authority or MCP permission changes. The changed test source needs a fresh assertion review.",
            "previousStatus": "tested"
          },
          {
            "id": "focused-field",
            "title": "Keep the focused field visible while tabbing",
            "dimension": "layout",
            "summary": "Checks every expected focus target remains visible with optional sections on/off, empty/large lists and minimum-sized form heights.",
            "status": "tested",
            "flows": [
              "editcard"
            ],
            "evidence": [
              "card-editor-focused-field-always-visible"
            ],
            "limit": "Rendered text and model focus; no screenshot, native terminal or image baseline.",
            "previousStatus": "tested"
          },
          {
            "id": "skill-filter",
            "title": "Toggle selected-only skills without losing selection",
            "dimension": "interaction",
            "summary": "The retained regression verifies selected-only filtering/cursor behavior across a registry refresh and toggles the visible selection.",
            "status": "partial",
            "flows": [
              "skill-filter"
            ],
            "evidence": [
              "selected-only-skill-cursor-normalizes-after-registry-refresh"
            ],
            "limit": "No direct Ctrl+S assertion, selected-only header assertion or selected-only/search/save round trip was found in editor and capability-list tests.",
            "sources": {
              "refs": [
                "views.CardEditor.handleKey",
                "components.CapabilityList.SetSelectedOnly",
                "components.CapabilityList.FilteredItems",
                "components.CapabilityList.renderHeader",
                "views.CardEditor.handleSpace"
              ]
            },
            "suggestion": "Send Ctrl+S on Skills and outside Skills; check header, search subset, saved selections and second-toggle restoration.",
            "priority": "Next",
            "reason": "SetSelectedOnly coverage alone cannot detect a broken keyboard binding.",
            "previousStatus": "partial"
          },
          {
            "id": "temporary-toggle",
            "title": "Toggle temporary workspace mode with Space and save it",
            "dimension": "interaction",
            "summary": "The editor focuses Use Tmp Workspace, toggles false→true→false with Space and submits both saved values through OnSave.",
            "status": "tested",
            "flows": [
              "temporary-toggle"
            ],
            "evidence": [
              "temporary-workspace-space-save"
            ],
            "limit": "Ordinary Go key/model regression checks the normalized card payload; repository and lifecycle behavior remain separately assessed.",
            "sources": {
              "refs": [
                "views.CardEditor.handleSpace",
                "views.CardEditor.renderTmpWorkspaceLine",
                "views.CardEditor.handleSubmit",
                "views.CardEditor.Draft"
              ]
            },
            "previousStatus": "tested"
          }
        ]
      },
      {
        "id": "workspace-selection",
        "title": "Workspace selection",
        "area": "UI forms",
        "summary": "Choose multiple sub workspaces, preserve provisional selections during refresh and explicitly clear them.",
        "scenarios": [
          {
            "id": "clear-subworkspaces",
            "title": "Clear only sub-workspace selections",
            "dimension": "interaction",
            "summary": "Checks Ctrl+X clears selected/missing IDs, preserves all other card fields and focus, invokes no save/cancel, shows the hint and ignores plain x.",
            "status": "tested",
            "flows": [
              "sub-clear"
            ],
            "evidence": [
              "ctrl-x-clears-only-sub-workspace-selection"
            ],
            "limit": "In-memory editor; filesystem copies and registry writes are outside this test.",
            "previousStatus": "tested"
          },
          {
            "id": "sub-picker",
            "title": "Select, cancel and save multiple sub workspaces",
            "dimension": "interaction",
            "summary": "The keyboard-level view test covers multiple checks across searches, unresolved entries, cancel/confirm and returning to single-selection main mode.",
            "status": "tested",
            "flows": [
              "sub-picker"
            ],
            "evidence": [
              "sub-workspace-picker-multiple-unresolved-and-main-modes"
            ],
            "limit": "Form callback payload and view text only; source copying is a separate runtime flow.",
            "sources": {
              "refs": [
                "views.CardEditor.handleSpace",
                "views.WorkspacePickerModal.SetMultiple",
                "views.WorkspacePickerModal.SetSelectedIDs",
                "views.WorkspacePickerModal.rebuildList",
                "views.WorkspacePickerModal.Update"
              ]
            },
            "previousStatus": "tested"
          },
          {
            "id": "picker-refresh",
            "title": "Preserve provisional checks through registry refresh",
            "dimension": "interaction",
            "summary": "The editor/picker regression directly checks missing rows, cancellation, explicit deselection and retaining a checked ID absent at confirmation.",
            "status": "tested",
            "flows": [
              "picker-refresh"
            ],
            "evidence": [
              "sub-workspace-new-selection-survives-registry-refresh"
            ],
            "limit": "Controlled DataLoadedMsg refreshes in the model; no screenshot or source-registry race claim.",
            "sources": {
              "refs": [
                "views.CardEditor.SetChoices",
                "views.WorkspacePickerModal.SetWorkspaces",
                "views.WorkspacePickerModal.rememberWorkspace",
                "views.WorkspacePickerModal.filterWorkspaces",
                "views.WorkspacePickerModal.refreshItems"
              ]
            },
            "previousStatus": "tested"
          }
        ]
      },
      {
        "id": "registry-refresh",
        "title": "Registry refresh and source management",
        "area": "UI coordinator",
        "summary": "Preserve loaded editor state, manage workspace sources and report synchronization failures.",
        "scenarios": [
          {
            "id": "sync-warning",
            "title": "Warn and refresh after a saved source fails to sync",
            "dimension": "behavior",
            "summary": "Checks emitted sync error, refresh request and durable saved source path despite failed scanning.",
            "status": "unassessed",
            "flows": [
              "source",
              "mutation"
            ],
            "evidence": [
              "save-simple-source-surfaces-sync-failure"
            ],
            "limit": "One missing skill-source path through callbacks; no full source-manager UI journey. The changed test source needs a fresh assertion review.",
            "previousStatus": "tested"
          },
          {
            "id": "workspace-sources",
            "title": "Visit workspace sources and return to the unsaved card",
            "dimension": "interaction",
            "summary": "Repository callbacks preserve workspace-source identity and reject malformed registry mutations. Search reviewed: internal/ui/views/{card_editor*,workspace_picker*,source_manager*}_test.go and internal/ui/{navigation,integration,source_save,source_sync_warning}_test.go; Ctrl+E/OnOpenWorkspaces call sites.",
            "status": "partial",
            "flows": [
              "workspace-sources"
            ],
            "evidence": [
              "workspace-source-i-d-assigned-and-preserved-on-edit",
              "save-workspace-source-aborts-on-malformed-registry",
              "delete-workspace-source-aborts-on-malformed-registry"
            ],
            "limit": "Searched editor/navigation/source tests: no direct assertion of Ctrl+E from both Workspace/Sub Workspaces, source edit, Esc return and preservation of the unsaved card was found. The changed test source needs a fresh assertion review.",
            "sources": {
              "refs": [
                "views.CardEditor.handleCtrlE",
                "ui.Coordinator.cardEditorCallbacks",
                "ui.Coordinator.navigateTo",
                "views.SourceManager.Update",
                "config.FileRepository.UpdateWorkspaceSources"
              ]
            },
            "suggestion": "Drive Ctrl+E from each workspace field, save a source, return with Esc and compare the unsaved card/focus.",
            "priority": "Next",
            "reason": "Cross-view refresh can overwrite unsaved editor state even when repository callbacks work.",
            "previousStatus": "partial"
          },
          {
            "id": "source-crud",
            "title": "Create, edit and delete source rows through keys",
            "dimension": "interaction",
            "summary": "Successful callback/reload transitions and synchronous/asynchronous deletion errors have concrete assertions.",
            "status": "partial",
            "flows": [
              "source",
              "mutation"
            ],
            "evidence": [
              "source-list-key-crud",
              "source-list-delete-error",
              "source-manager-delete-error"
            ],
            "limit": "Component/callback fixtures; full source persistence and unsaved-editor return are separate boundaries. The changed test source needs a fresh assertion review.",
            "previousStatus": "tested"
          },
          {
            "id": "destination-refresh",
            "title": "Apply source and card refreshes to actual views",
            "dimension": "behavior",
            "summary": "Nonempty persisted data is delivered to source lists and dashboard cards; updates and removals cannot be satisfied by repository-only assertions.",
            "status": "partial",
            "flows": [
              "mutation",
              "source"
            ],
            "evidence": [
              "source-refresh-destination",
              "card-refresh-destination"
            ],
            "limit": "In-process coordinator views; not every registry or external filesystem change. The changed test source needs a fresh assertion review.",
            "previousStatus": "tested"
          },
          {
            "id": "corrupt-registry",
            "title": "Identify independently corrupted registries",
            "dimension": "behavior",
            "summary": "Skills, personas and MCP YAML corruption each fails from its own valid baseline and identifies the correct file.",
            "status": "tested",
            "flows": [
              "savecard"
            ],
            "evidence": [
              "isolated-registry-corruption"
            ],
            "limit": "Repository card validation; no TUI save error journey."
          },
          {
            "id": "blank-canvas",
            "title": "Allocate the blank canvas used by puppet rendering",
            "dimension": "layout",
            "summary": "Constructor output must have the expected blank rows, columns and content.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "blank-canvas-representation"
            ],
            "limit": "ASCII representation test; does not compare puppet screenshots."
          }
        ]
      },
      {
        "id": "application-settings",
        "title": "Application settings",
        "area": "UI forms",
        "summary": "Edit and persist independent workspace commands and editor visibility preferences.",
        "scenarios": [
          {
            "id": "settings",
            "title": "Save commands and editor visibility settings",
            "dimension": "behavior",
            "summary": "Retained tests establish settings payloads and exact repository round trips; the new coordinator regression executes a failing save and preserves its visible error and settings state.",
            "status": "partial",
            "flows": [
              "settings"
            ],
            "evidence": [
              "settings-view_-save-in-edit-mode",
              "settings-view_-terminal-command-focus-save-and-cancel",
              "settings-view_-toggle-rows",
              "app-settings-persistence",
              "terminal-launch-through-dashboard",
              "settings-save-failure-visible"
            ],
            "limit": "Failed save composition is now directly asserted. No single successful key→durable YAML→refreshed editor visibility→return-navigation journey is asserted. The changed test source needs a fresh assertion review.",
            "sources": {
              "refs": [
                "views.SettingsView.handleEditEnter",
                "ui.Coordinator.settingsCallbacks",
                "config.FileRepository.SaveAppSettings",
                "config.AppSettings",
                "ui.Coordinator.handleSpecialMessages"
              ]
            },
            "suggestion": "Add a successful coordinator save journey checking durable YAML, editor visibility refresh and dashboard return.",
            "priority": "First",
            "reason": "The reported error test never executes its error command; failed persistence must not look like a successful save.",
            "previousStatus": "partial"
          },
          {
            "id": "settings-edit",
            "title": "Edit, save and cancel independent workspace commands",
            "dimension": "interaction",
            "summary": "View-level assertions cover field focus, independent text, trimming on save and restoration on cancel.",
            "status": "tested",
            "flows": [
              "settings-edit"
            ],
            "evidence": [
              "settings-view_-terminal-command-focus-save-and-cancel",
              "settings-view_-save-in-edit-mode",
              "settings-view_-toggle-rows"
            ],
            "limit": "Tests use in-memory callbacks; durable saving is assessed separately.",
            "sources": {
              "refs": [
                "views.NewSettingsView",
                "views.SettingsView.SetSettings",
                "views.SettingsView.startEdit",
                "views.SettingsView.updateEdit",
                "views.SettingsView.syncFocus"
              ]
            },
            "previousStatus": "tested"
          },
          {
            "id": "settings-layout",
            "title": "Keep settings controls and long command text within bounds",
            "dimension": "layout",
            "summary": "Direct layout assertions cover display/edit modes, sanitized text, visible controls/help and width/height limits.",
            "status": "tested",
            "flows": [
              "settings-edit",
              "settings"
            ],
            "evidence": [
              "settings-view_-long-commands-stay-within-size"
            ],
            "limit": "Two tested view sizes; no screenshot comparison or native terminal font guarantee.",
            "sources": {
              "refs": [
                "views.NewSettingsView",
                "views.SettingsView.SetSettings",
                "views.SettingsView.startEdit",
                "views.SettingsView.updateEdit",
                "views.SettingsView.syncFocus"
              ]
            },
            "previousStatus": "tested"
          }
        ]
      },
      {
        "id": "mcp-form",
        "title": "MCP form serialization",
        "area": "UI forms",
        "summary": "Round-trip all MCP fields including escaped comma values.",
        "scenarios": [
          {
            "id": "all-fields",
            "title": "Submit every loaded field without loss",
            "dimension": "interaction",
            "summary": "Presses Enter and asserts one save with exact ID/name/command/args/CWD/URL/env/headers, including an argument containing a comma.",
            "status": "unassessed",
            "flows": [
              "mcp"
            ],
            "evidence": [
              "m-c-p-form-round-trips-every-field"
            ],
            "limit": "Callback payload only; does not connect to that server or persist the configuration. The changed test source needs a fresh assertion review.",
            "previousStatus": "tested"
          }
        ]
      },
      {
        "id": "workspace-commands",
        "title": "Open workspaces in an IDE or terminal",
        "area": "Workspace tools",
        "summary": "Launch configured commands in the selected session workspace and report configuration or process errors.",
        "scenarios": [
          {
            "id": "ide",
            "title": "Open the selected workspace in an IDE",
            "dimension": "behavior",
            "summary": "The o key targets the selected session; shared command execution validates cwd and argument substitution.",
            "status": "partial",
            "flows": [
              "ide"
            ],
            "evidence": [
              "dashboard_-launch-terminal-routing",
              "launch-workspace-command-executes-in-workspace",
              "terminal-launch-reports-configuration-and-lookup-errors"
            ],
            "limit": "No retained test drives a successful o command through the coordinator and asserts the actual IDE child cwd/argv. The real child test covers the common launcher, and the IDE callback is directly tested for missing configuration. The changed test source needs a fresh assertion review.",
            "sources": {
              "refs": [
                "views.Dashboard.handleKey",
                "views.Dashboard.handleLaunchSession",
                "ui.Coordinator.launchIDE",
                "ui.sessionActions.launchWorkspace",
                "api.Service.CheckSessionInDashboard"
              ]
            },
            "suggestion": "Drive o with an isolated helper executable configured as IDECommand and assert the child cwd/argv and scoped completion.",
            "priority": "Next",
            "reason": "Terminal and IDE settings are independent; shared-launcher coverage does not prove successful IDE wiring.",
            "previousStatus": "partial"
          },
          {
            "id": "terminal",
            "title": "Launch a terminal in the actual selected workspace",
            "dimension": "interaction",
            "summary": "A coordinator key journey launches a real helper child in the temporary session workspace and checks its arguments and successful scoped completion.",
            "status": "partial",
            "flows": [
              "terminal"
            ],
            "evidence": [
              "terminal-launch-through-dashboard",
              "dashboard_-launch-terminal-routing",
              "launch-workspace-command-executes-in-workspace"
            ],
            "limit": "The child is a test executable, not a graphical terminal emulator; only one configured command template is exercised here. The changed test source needs a fresh assertion review.",
            "sources": {
              "refs": [
                "views.Dashboard.handleKey",
                "views.Dashboard.handleLaunchSession",
                "ui.Coordinator.dashboardCallbacks",
                "ui.Coordinator.launchTerminal",
                "ui.sessionActions.launchWorkspace"
              ]
            },
            "previousStatus": "tested"
          },
          {
            "id": "workspace-command-errors",
            "title": "Report workspace command configuration and spawn failures",
            "dimension": "behavior",
            "summary": "Tests check command/path failures and delivery of a spawn failure to the coordinator error banner.",
            "status": "partial",
            "flows": [
              "workspace-command-errors"
            ],
            "evidence": [
              "terminal-launch-reports-configuration-and-lookup-errors",
              "launch-workspace-command-errors"
            ],
            "limit": "Checks the listed API/callback cases and banner presence; not exact error copy or external app startup errors after exec succeeds. The changed test source needs a fresh assertion review.",
            "sources": {
              "refs": [
                "api.Service.CheckSessionInDashboard",
                "ui.sessionActions.launchWorkspace",
                "session.LaunchWorkspaceCommand",
                "messages.ScopedActionFinishedMsg",
                "ui.Coordinator.handleSessionMessages"
              ]
            },
            "previousStatus": "tested"
          }
        ]
      },
      {
        "id": "conversation-choice",
        "title": "Explicit native conversation choice",
        "area": "Sessions",
        "summary": "Permit compatible bindings and reject incompatible driver identities.",
        "scenarios": [
          {
            "id": "compatible-only",
            "title": "Submit only compatible bindings",
            "dimension": "interaction",
            "summary": "Checks a matching Claude binding is submitted and a selected Codex binding cannot be submitted to the Claude activation.",
            "status": "unassessed",
            "flows": [
              "resume"
            ],
            "evidence": [
              "conversation-choice-requires-explicit-compatible-choice"
            ],
            "limit": "Picker callback boundary; native CLI resume is outside scope. The changed test source needs a fresh assertion review.",
            "previousStatus": "tested"
          }
        ]
      },
      {
        "id": "activation-readiness",
        "title": "Activation startup and delivery readiness",
        "area": "Sessions",
        "summary": "Publish processes for immediate attachment while gating message delivery separately.",
        "scenarios": [
          {
            "id": "readiness",
            "title": "Attach immediately while message delivery waits for startup",
            "dimension": "behavior",
            "summary": "Manager tests separately assert immediate attachment, delayed delivery, registration independence and inspection snapshot preservation.",
            "status": "tested",
            "flows": [
              "readiness"
            ],
            "evidence": [
              "activation-published-before-startup-delay",
              "startup-admission-and-registration-are-independent",
              "inspection-conversion-preserves-activation-snapshot"
            ],
            "limit": "Controlled startup timers and local processes; delivery eligibility still does not prove an AI agent can process a task.",
            "sources": {
              "refs": [
                "session.Manager.activateSessionLocked",
                "session.Manager.startStartupLocked",
                "session.Manager.AttachChecked",
                "session.Manager.snapshotLocked",
                "session.Manager.deliveryAvailableLocked"
              ]
            }
          },
          {
            "id": "dashboard-launch",
            "title": "Launch a process through dashboard input",
            "dimension": "interaction",
            "summary": "Executed dashboard-key commands publish a live cat process and persist its session identity and workspace.",
            "status": "unassessed",
            "flows": [
              "launch"
            ],
            "evidence": [
              "dashboard-launch-persisted"
            ],
            "limit": "Local process and application policy; native-agent behavior is assessed separately. The changed test source needs a fresh assertion review.",
            "previousStatus": "tested"
          }
        ]
      },
      {
        "id": "workspace-identity",
        "title": "Saved workspace validation",
        "area": "Sessions",
        "summary": "Relaunch and switch reuse saved paths and do not recreate missing work.",
        "scenarios": [
          {
            "id": "unavailable-path",
            "title": "Fail without spawning until the saved workspace is restored",
            "dimension": "behavior",
            "summary": "For relaunch/switch and original/temporary paths, checks missing/file errors, no extra Spawn, preserved session identity/content, and successful single-spawn retry after restoring the directory.",
            "status": "tested",
            "flows": [
              "workspace",
              "resume"
            ],
            "evidence": [
              "activation-rejects-unavailable-saved-workspace"
            ],
            "limit": "Controlled driver counts and filesystem fixtures; no real Codex/Claude execution."
          },
          {
            "id": "inspect-only",
            "title": "Read sub-workspace descriptors without preparing directories",
            "dimension": "behavior",
            "summary": "Checks descriptor identity/path, absence of a newly created tmp directory and unavailable status for corrupt manifests.",
            "status": "tested",
            "flows": [
              "workspace"
            ],
            "evidence": [
              "read-sub-workspace-descriptors-does-not-prepare-workspace"
            ],
            "limit": "Read-only helper only; copy collision and multi-selection journeys remain separate."
          }
        ]
      },
      {
        "id": "relaunch-mcp",
        "title": "Relaunch deployment and credentials",
        "area": "Sessions",
        "summary": "Preserve workspace-authored files while rotating the Kankuro activation entry.",
        "scenarios": [
          {
            "id": "preserve-and-rotate",
            "title": "Refresh only Kankuro MCP after a manager restart",
            "dimension": "behavior",
            "summary": "Checks invalid config blocks relaunch without rewriting, then preserves instructions/custom server/metadata, rotates token, revokes the old credential and uses the new endpoint.",
            "status": "tested",
            "flows": [
              "mcp-config",
              "resume"
            ],
            "evidence": [
              "relaunch-refreshes-only-kankuro-m-c-p-after-restart"
            ],
            "limit": "Local files and cat process across in-process restart; no native CLI MCP handshake."
          }
        ]
      },
      {
        "id": "workspace-trust",
        "title": "Workspace trust safety",
        "area": "Drivers",
        "summary": "Fail closed on malformed native CLI configuration.",
        "scenarios": [
          {
            "id": "preserve-invalid",
            "title": "Preserve malformed TOML/JSON and wrong-shaped objects",
            "dimension": "behavior",
            "summary": "Checks PrepareWorkspace errors and byte-identical original configuration for Codex TOML, Claude JSON and a wrong-typed projects field.",
            "status": "tested",
            "flows": [
              "trust"
            ],
            "evidence": [
              "trust-invalid-configuration-preserved"
            ],
            "limit": "Isolated HOME; no installed native CLI trust acceptance."
          }
        ]
      },
      {
        "id": "protocol-skills",
        "title": "Embedded agent protocol skills",
        "area": "Drivers",
        "summary": "Select and deploy the embedded MCP access and registration/reporting guidance.",
        "scenarios": [
          {
            "id": "skills",
            "title": "Select and deploy the embedded protocol skills",
            "dimension": "behavior",
            "summary": "Capability assertions cover conditional/deduplicated access+puppet selection; root tests compare actual deployed bytes with the embedded content.",
            "status": "tested",
            "flows": [
              "skills"
            ],
            "evidence": [
              "resolve-capabilities-adds-m-c-p-access-skill-only-when-kankuro-m-c-p-selected",
              "embedded-m-c-p-c-l-i-skill-deploys-through-config",
              "embedded-kankuro-puppet-skill-deploys-through-config"
            ],
            "limit": "Embedded file/selection contract only. No end-to-end assertion that an AI agent follows registration/report instructions.",
            "sources": {
              "refs": [
                "config.SetBuiltinKankuroPuppetSkillContent",
                "api.appendKankuroAccessSkill",
                "driver.GenericCLIDriver.SetSkills",
                "driver.writeBuiltinSkill",
                "driver.syncInstructionFile"
              ]
            }
          }
        ]
      },
      {
        "id": "session-actions",
        "title": "Session attachment, naming and removal",
        "area": "Sessions",
        "summary": "Apply manual actions in the owning dashboard and stop processes before deletion.",
        "scenarios": [
          {
            "id": "manual-session",
            "title": "Attach, name, stop and remove owned sessions from the UI",
            "dimension": "interaction",
            "summary": "Manager/service tests assert attachment replay, remove guards/cleanup and naming failure; the launcher checks pending/retry behavior and the command harness drives captured actions. Existing-session rename now has a coordinator/service/disk/reload regression and view-level cancellation/scope-reset assertions.",
            "status": "partial",
            "flows": [
              "manual-session"
            ],
            "evidence": [
              "session-attach-replay",
              "service-remove-session-delegates-guards-cleanup-and-manager-errors",
              "session-launcher-waits-for-name-result-and-allows-retry",
              "set-name-propagates-persist-failure",
              "rename-rename-restart",
              "rename-cancel-draft"
            ],
            "limit": "Existing-session rename has direct successful refresh/restart and cancellation assertions. A complete dashboard stop/remove journey and raw terminal detach and restoration remain outside the retained assertions. The changed test source needs a fresh assertion review.",
            "sources": {
              "refs": [
                "ui.Coordinator.requestAttach",
                "api.Service.AttachInDashboard",
                "session.Manager.AttachChecked",
                "ui.Coordinator.handleSessionMessages",
                "api.Service.RenameDashboardSession"
              ]
            },
            "suggestion": "Drive stop/remove through the coordinator and assert removal effects; add a project-owned terminal detach/restoration journey.",
            "priority": "First",
            "reason": "Scoped asynchronous actions and raw attachment cross ownership and terminal-state boundaries.",
            "previousStatus": "partial"
          },
          {
            "id": "stop-delete",
            "title": "Stop before removing a session and retain truthful outcomes",
            "dimension": "behavior",
            "summary": "The service lifecycle test asserts stop/switch/relaunch/delete and stable replay; removal tests check running guards/temp cleanup and messaging tests check recipient history deletion.",
            "status": "tested",
            "flows": [
              "stop-delete"
            ],
            "evidence": [
              "operation-runner-creates-once-and-replays-request",
              "service-remove-session-delegates-guards-cleanup-and-manager-errors",
              "messaging-operations-pin-activations-and-persist-explicit-reports"
            ],
            "limit": "Local cat processes; no proof every descendant process group is killed. UI confirmations are a separate interaction boundary.",
            "sources": {
              "refs": [
                "api.Service.StopSession",
                "session.Manager.killDetailed",
                "driver.GenericCLIDriver.Kill",
                "session.Manager.revokeSessionActivationLocked",
                "api.Service.DeleteSession"
              ]
            }
          }
        ]
      },
      {
        "id": "durable-operations",
        "title": "Durable lifecycle operations",
        "area": "Runtime",
        "summary": "Deduplicate requests, bound worker admission, cancel work and recover truthful outcomes.",
        "scenarios": [
          {
            "id": "operations",
            "title": "Replay, bound, cancel and recover durable operations",
            "dimension": "behavior",
            "summary": "Assertions cover concurrent duplicate requests, conflicting fingerprints, bounded queues, disconnected waits, late cancellation, post-effect journal warnings and interrupted/corrupt recovery.",
            "status": "tested",
            "flows": [
              "operations"
            ],
            "evidence": [
              "operation-runner-creates-once-and-replays-request",
              "final-journal-write-failure-returns-outcome-and-recovers-interrupted",
              "operation-queue-bound-and-disconnected-wait-return-promptly",
              "operation-journal-recovers-interrupted-and-fails-closed-on-corruption"
            ],
            "limit": "Controlled local runner/real cat fixtures and constructed durable records; no OS crash or exhaustive scheduling proof.",
            "sources": {
              "refs": [
                "api.Service.CreateSession",
                "api.OperationRunner.Replay",
                "api.OperationRunner.accept",
                "api.OperationRunner.execute",
                "api.controllerAuthority.validate"
              ]
            }
          }
        ]
      },
      {
        "id": "shutdown",
        "title": "Shutdown while activation is in flight",
        "area": "Runtime",
        "summary": "Join in-flight work and clean up processes and temporary directories.",
        "scenarios": [
          {
            "id": "spawn-boundary",
            "title": "Kill a process that finishes spawning during shutdown",
            "dimension": "behavior",
            "summary": "Checks shutdown blocks for an in-flight relaunch, returns ErrManagerClosed after release, kills the spawned driver and leaves terminated status.",
            "status": "tested",
            "flows": [
              "quit"
            ],
            "evidence": [
              "shutdown-during-spawn-kills-the-new-process"
            ],
            "limit": "Gated test driver; does not simulate every OS process-group behavior."
          },
          {
            "id": "copy-boundary",
            "title": "Wait for temporary-workspace preparation and remove its copy",
            "dimension": "behavior",
            "summary": "Blocks the copier, asserts Shutdown waits, then checks closed-manager launch error and an empty temporary root.",
            "status": "tested",
            "flows": [
              "quit",
              "workspace"
            ],
            "evidence": [
              "shutdown-waits-for-temporary-workspace-preparation"
            ],
            "limit": "Injected copier; does not prove native cp cancellation under every filesystem."
          }
        ]
      },
      {
        "id": "session-mapping",
        "title": "Persistence and snapshot field mapping",
        "area": "Storage",
        "summary": "Copy actual nonzero values rather than checking only field names.",
        "scenarios": [
          {
            "id": "nonzero-values",
            "title": "Copy every persisted field into restore and snapshot",
            "dimension": "behavior",
            "summary": "Fills fields with nonzero values and compares all persisted values through data/restore and snapshot, excluding intentional schema/status transformations.",
            "status": "tested",
            "flows": [
              "watch"
            ],
            "evidence": [
              "data-round-trip-copies-every-field"
            ],
            "limit": "Pure mapping; no disk durability assertion or complete clone-alias isolation claim."
          }
        ]
      },
      {
        "id": "task-messaging",
        "title": "Orchestrator task delivery",
        "area": "Messaging",
        "summary": "Pin a recipient activation and publish one replay-safe task through shared admission.",
        "scenarios": [
          {
            "id": "send",
            "title": "Pin and submit an idempotent orchestrator message",
            "dimension": "behavior",
            "summary": "Retained tests assert activation-pinned submission, replay conflict behavior, journal rollback and capacity recovery without leaked admission.",
            "status": "tested",
            "flows": [
              "send"
            ],
            "evidence": [
              "messaging-operations-pin-activations-and-persist-explicit-reports",
              "ordinary-message-journal-rejection-rolls-back-history-and-retries",
              "ordinary-message-queue-rejection-releases-reservation-and-preserves-replay"
            ],
            "limit": "Submission/history invariants under local fixtures; no AI-agent task-processing or external client UX claim.",
            "sources": {
              "refs": [
                "api.Service.SendMessage",
                "session.Manager.ReserveMessageAdmission",
                "api.OperationRunner.SubmitWithAdmission",
                "api.preparedDelivery.publish",
                "session.Manager.PublishMessageAdmission"
              ]
            }
          }
        ]
      },
      {
        "id": "report-intent",
        "title": "Report persistence and silent progress",
        "area": "Messaging",
        "summary": "Durably correlate explicit reports and initial notification obligations.",
        "scenarios": [
          {
            "id": "durable-idempotent",
            "title": "Persist report kind and notification state exactly once",
            "dimension": "behavior",
            "summary": "Checks pending for completed/blocked/question and not_required for progress; reopen/retry returns the same report ID and one stored item with the expected state.",
            "status": "tested",
            "flows": [
              "reports",
              "report-progress",
              "report-inspection"
            ],
            "evidence": [
              "report-notification-initial-state-is-durable-and-idempotent"
            ],
            "limit": "Message store on disk; no prompt delivery or agent processing."
          },
          {
            "id": "progress-then-complete",
            "title": "Retain progress separately from later completion",
            "dimension": "behavior",
            "summary": "Checks progress deduplication, rejection of a changed kind under the same request ID, and ordered progress/completed reports after reopening.",
            "status": "tested",
            "flows": [
              "report-progress",
              "report-inspection"
            ],
            "evidence": [
              "progress-replies-persist-before-completion"
            ],
            "limit": "Does not observe whether a live CLI continues working after progress."
          },
          {
            "id": "interrupted-attempt",
            "title": "Recover admitted notification input as uncertain",
            "dimension": "behavior",
            "summary": "Checks recovery marks admitted delivery uncertain while preserving the accepted obligation and request identity.",
            "status": "tested",
            "flows": [
              "recovery"
            ],
            "evidence": [
              "restart-marks-admitted-notification-uncertain"
            ],
            "limit": "Constructed durable records; no process crash at an actual PTY write."
          }
        ]
      },
      {
        "id": "automatic-wakeups",
        "title": "Automatic report wake-ups",
        "area": "Messaging",
        "summary": "Qualifying reports notify the original controller without giving workers send authority.",
        "scenarios": [
          {
            "id": "controller-notified",
            "title": "Deliver a correlated notification with a private report body",
            "dimension": "behavior",
            "summary": "Denies the worker a general send, observes controller PTY output containing message/reply IDs and no-report-reply guidance, and checks sensitive report content is absent.",
            "status": "tested",
            "flows": [
              "reports",
              "notification-dispatch",
              "notification-authority",
              "notification-write"
            ],
            "evidence": [
              "report-automatically-notifies-original-controller"
            ],
            "limit": "Real local PTY harness; receipt of bytes is not proof an AI agent processed or fetched the report."
          },
          {
            "id": "reporter-stop",
            "title": "Dispatch a saved report after the worker stops",
            "dimension": "behavior",
            "summary": "Stops dispatcher, saves report, kills reporter, restarts runner and waits for notification output.",
            "status": "tested",
            "flows": [
              "notification-dispatch",
              "notification-authority"
            ],
            "evidence": [
              "saved-notification-survives-reporter-stopping"
            ],
            "limit": "Same application process and local harness; not an application-wide restart."
          },
          {
            "id": "sender-replaced",
            "title": "Refuse notification to a replacement controller activation",
            "dimension": "behavior",
            "summary": "Stops/relaunches the original sender, attempts dispatch and checks unavailable notification state.",
            "status": "tested",
            "flows": [
              "notification-authority"
            ],
            "evidence": [
              "stopped-original-activation-makes-notification-unavailable"
            ],
            "limit": "State assertion for one activation replacement; not arbitrary authorization races."
          }
        ]
      },
      {
        "id": "notification-retries",
        "title": "Notification retries and cancellation",
        "area": "Messaging",
        "summary": "Retry safe outcomes, preserve ordering and cancel queued notification work.",
        "scenarios": [
          {
            "id": "queue-full",
            "title": "Retry when queue capacity returns",
            "dimension": "behavior",
            "summary": "Fills queue, checks pending obligation and request ID survive rejection, drains queue, starts workers and observes notification with the same request ID.",
            "status": "tested",
            "flows": [
              "notification-retry"
            ],
            "evidence": [
              "notification-retries-after-queue-capacity-returns"
            ],
            "limit": "Controlled runner/queue; no native agent readiness proof."
          },
          {
            "id": "byte-evidence",
            "title": "Reconcile accepted/submitted/uncertain/zero-byte evidence",
            "dimension": "behavior",
            "summary": "Checks state mapping, advances the attempt only for not-submitted delivery and verifies stale reconciliation does not downgrade or advance it again.",
            "status": "tested",
            "flows": [
              "notification-retry",
              "recovery"
            ],
            "evidence": [
              "notification-reconciliation-uses-delivery-evidence"
            ],
            "limit": "Constructed persisted delivery states; concrete driver byte-write boundaries need their own tests."
          },
          {
            "id": "single-admission",
            "title": "Admit one attempt under concurrent dispatch",
            "dimension": "behavior",
            "summary": "Runs twelve dispatchers against the same obligation and asserts one matching runner operation.",
            "status": "tested",
            "flows": [
              "notification-dispatch",
              "notification-retry"
            ],
            "evidence": [
              "concurrent-notification-dispatch-admits-one-attempt"
            ],
            "limit": "One controlled interleaving; this run did not use the race detector."
          },
          {
            "id": "fairness",
            "title": "Preserve recipient ordering without blocking other recipients",
            "dimension": "behavior",
            "summary": "Checks one recipient backs off without suppressing another and keeps its accepted predecessor ahead of a later pending report.",
            "status": "tested",
            "flows": [
              "notification-retry"
            ],
            "evidence": [
              "notification-scheduling-preserves-recipient-order-and-fairness"
            ],
            "limit": "Pure scheduling function with fixed time; no throughput benchmark."
          },
          {
            "id": "notification-cancel",
            "title": "Cancel queued notifications and join shutdown work",
            "dimension": "behavior",
            "summary": "Queued notification cancellation asserts no terminal submission and released lifecycle reservation; deletion revokes saved authority and ordinary delivery shutdown is deadline-checked.",
            "status": "partial",
            "flows": [
              "notification-cancel"
            ],
            "evidence": [
              "queued-delivery-cancellation-releases-reservation-without-submitting",
              "deleting-report-history-cancels-admitted-notification",
              "queued-message-reservations-cannot-deadlock-stop-workers-or-shutdown"
            ],
            "limit": "The notification case cancels captured authority directly. No dedicated retained assertion was found that runner.Shutdown itself cancels a notification queued or blocked at the PTY boundary and joins its dispatcher.",
            "sources": {
              "refs": [
                "api.OperationRunner.Shutdown",
                "api.preparedDelivery.cancel",
                "session.MessageStore.UpdateDelivery",
                "session.MessageAdmission.Release"
              ]
            },
            "suggestion": "Queue a real notification, call runner.Shutdown while it is gated, then assert dispatcher/workers finish, no bytes are written and the destination reservation is released.",
            "priority": "First",
            "reason": "Shutdown must not leave queued notification work or block application exit."
          }
        ]
      },
      {
        "id": "terminal-journeys",
        "title": "Native-agent acceptance boundary",
        "area": "Integration",
        "summary": "Project-owned Go E2E evidence uses verified current credentials, isolated service state and the requested OpenRouter router. The run failed at register; reached checkpoints and limits are explicit.",
        "scenarios": [
          {
            "id": "native-agent",
            "title": "Launch, attach, register and resume an installed native agent",
            "dimension": "interaction",
            "summary": "The corrected Go fixture verifies current OpenCode credentials, uses a private service port and selects OpenRouter Free Models Router. Launch/attach and the requested model label passed; registration did not complete within the four-minute state wait.",
            "status": "partial",
            "flows": [
              "launch",
              "attach",
              "registration",
              "resume"
            ],
            "evidence": [
              "native-agent-terminal-journey",
              "m-c-p-c-l-i-registers-conversation-for-live-activation",
              "claude-resume-lifecycle-uses-selected-native-i-d",
              "session-attach-replay",
              "tui-luna",
              "tui-codex"
            ],
            "limit": "Single OpenCode 2.2.6/OpenRouter run; no other CLI/provider or image baseline comparison. Failure at register; later journey checkpoints remain unverified. Free Models Router may choose different backing models; responsiveness does not establish tool execution. The separate current Codex runs verify idle lifecycle, live registration and report takeover. Live recall remains open.",
            "sources": {
              "refs": [
                "ui.Coordinator.launchCard",
                "api.Service.LaunchSession",
                "session.Manager.ReserveCardReferenceForDashboardWithCancel",
                "session.Manager.LaunchInDashboardWithIDContext",
                "session.Manager.bindPendingWorkspace"
              ],
              "files": [
                "tests/e2e/environment_test.go",
                "tests/e2e/opencode_test.go",
                "tests/e2e/capture_test.go",
                "tests/e2e/provenance_test.go",
                "tests/e2e/README.md"
              ]
            },
            "suggestion": "Rerun the tagged journey when the selected provider/agent can complete registration within the bounded wait.",
            "priority": "First",
            "reason": "Reached launch/model assertions do not establish native registration, detach/restoration or resume.",
            "captures": [
              "native-launched",
              "native-failure"
            ]
          },
          {
            "id": "registration-fixture",
            "title": "Register and persist a conversation through a real MCP CLI",
            "dimension": "behavior",
            "summary": "The shell-agent interoperability test exercises the deployed endpoint/credential and report persistence/retry through an installed mcp-cli.",
            "status": "tested",
            "flows": [
              "registration"
            ],
            "evidence": [
              "m-c-p-c-l-i-registers-conversation-for-live-activation"
            ],
            "limit": "Shell/cat agent rather than Codex/Claude; does not establish native agent task execution.",
            "sources": {
              "refs": [
                "mcp.Start",
                "mcp.authorize",
                "api.Service.RegisterConversation",
                "session.Manager.RegisterConversation",
                "session.Manager.registerConversationLocked"
              ]
            }
          },
          {
            "id": "resume-driver",
            "title": "Construct selected Claude resume arguments through lifecycle",
            "dimension": "behavior",
            "summary": "Capture-driver assertions cover explicit saved/new choices and per-card IDs while retaining the original workspace.",
            "status": "tested",
            "flows": [
              "resume"
            ],
            "evidence": [
              "claude-resume-lifecycle-uses-selected-native-i-d"
            ],
            "limit": "Spawn is overridden; these are actual manager lifecycle and argument assertions, not a native CLI run.",
            "sources": {
              "refs": [
                "ui.Coordinator.requestRelaunch",
                "session.Manager.GetSession",
                "views.ConversationChoiceView.SetContext",
                "ui.Coordinator.conversationChoiceCallbacks",
                "api.Service.SwitchInDashboard"
              ]
            }
          },
          {
            "id": "attachment-handle",
            "title": "Replay output through manager attachment handles",
            "dimension": "behavior",
            "summary": "Real cat attachment replays the exact input bytes after closing and reopening the handle.",
            "status": "tested",
            "flows": [
              "attach"
            ],
            "evidence": [
              "session-attach-replay"
            ],
            "limit": "Does not exercise attachExec terminal mode, Ctrl+K or SIGWINCH.",
            "sources": {
              "refs": [
                "ui.Coordinator.requestAttach",
                "api.Service.AttachInDashboard",
                "session.Manager.AttachChecked",
                "ui.Coordinator.handleSessionMessages",
                "ui.scopedAttachCommand"
              ]
            }
          }
        ]
      },
      {
        "id": "session-rename",
        "title": "Rename an existing session",
        "area": "Dashboard",
        "summary": "Select a session and press e to edit its saved display name.",
        "sources": {
          "refs": [
            "views.Dashboard.handleEdit",
            "ui.Coordinator.renameDashboardSession"
          ],
          "files": [
            "internal/ui/views/dashboard.go",
            "internal/ui/callbacks.go",
            "internal/ui/coordinator.go"
          ]
        },
        "scenarios": [
          {
            "id": "draft-retry",
            "title": "Edit a stable draft and retry a failure",
            "dimension": "interaction",
            "status": "unassessed",
            "evidence": [
              "rename-draft"
            ],
            "limit": "Stub callback; no filesystem or full cancellation journey. The changed test source needs a fresh assertion review.",
            "flows": [
              "rename-edit"
            ],
            "summary": "Captured target, normalized Unicode input, draft survives refresh, duplicate suppression, retry after error.",
            "previousStatus": "tested"
          },
          {
            "id": "blank-name",
            "title": "Blank name falls back to session ID",
            "dimension": "behavior",
            "status": "unassessed",
            "evidence": [
              "rename-blank"
            ],
            "limit": "Does not prove empty sessions-pane behavior. The changed test source needs a fresh assertion review.",
            "flows": [
              "rename-save"
            ],
            "summary": "Blank input submits the session ID.",
            "previousStatus": "tested"
          },
          {
            "id": "stale-result",
            "title": "Reject stale generation and retain save error",
            "dimension": "behavior",
            "status": "unassessed",
            "evidence": [
              "rename-scoped"
            ],
            "limit": "Missing target fixture; success and stale dashboard ID are not directly asserted. The changed test source needs a fresh assertion review.",
            "flows": [
              "rename-save"
            ],
            "summary": "Real command error stays in dialog; stale generation leaves pending state intact.",
            "previousStatus": "tested"
          },
          {
            "id": "layout",
            "title": "Minimum-size dialog and rename hint",
            "dimension": "layout",
            "status": "tested",
            "evidence": [
              "rename-layout"
            ],
            "limit": "Bounds and text assertions only; no native UI capture.",
            "flows": [
              "rename-layout"
            ],
            "summary": "Long error bounds at 80×24 and 120×40, visible save/cancel controls and scoped rename footer.",
            "previousStatus": "tested"
          },
          {
            "id": "search",
            "title": "Retain search typing of e",
            "dimension": "interaction",
            "status": "tested",
            "evidence": [
              "rename-search"
            ],
            "limit": "Existing cards-pane search test only.",
            "flows": [
              "rename-edit"
            ],
            "summary": "Existing search consumes typed t/e/s/t and Esc exits search.",
            "previousStatus": "tested"
          },
          {
            "id": "persist-error",
            "title": "Surface persistence errors",
            "dimension": "behavior",
            "status": "tested",
            "evidence": [
              "rename-persist"
            ],
            "limit": "Existing manager test; no UI integration.",
            "flows": [
              "rename-save"
            ],
            "summary": "Existing checked rename surfaces a record-write failure."
          },
          {
            "id": "launch-naming",
            "title": "Retain separate launch-name retry flow",
            "dimension": "interaction",
            "status": "unassessed",
            "evidence": [
              "rename-launcher"
            ],
            "limit": "Unchanged launcher test does not assert dashboard success avoids attachment. The changed test source needs a fresh assertion review.",
            "flows": [
              "launch"
            ],
            "summary": "Existing launcher naming waits for result and permits retry.",
            "previousStatus": "tested"
          },
          {
            "id": "success-restart",
            "title": "Successful dashboard rename survives restart",
            "dimension": "behavior",
            "status": "unassessed",
            "evidence": [
              "rename-rename-restart"
            ],
            "limit": "Local cat process and ordinary coordinator events; no native-agent terminal rendering journey. The changed test source needs a fresh assertion review.",
            "flows": [
              "rename-save"
            ],
            "summary": "The coordinator saves through the real service, immediately refreshes the selected name without attachment/navigation, preserves the running process and reloads the name after restart.",
            "previousStatus": "tested"
          },
          {
            "id": "cancel",
            "title": "Cancel before submission and clear scope",
            "dimension": "interaction",
            "status": "tested",
            "evidence": [
              "rename-cancel-draft"
            ],
            "limit": "View-level input events and callback spy; not real terminal input decoding.",
            "flows": [
              "rename-edit"
            ],
            "summary": "Esc and both dashboard-scope changes discard the draft without invoking save; reopening restores the original name.",
            "previousStatus": "tested"
          }
        ]
      },
      {
        "id": "mcp-registry",
        "title": "Manage global MCP servers",
        "area": "Dashboard",
        "summary": "Select a global definition while retaining repository safeguards.",
        "sources": {
          "refs": [
            "views.MCPManager.Update",
            "ui.Coordinator.handleSpecialMessages"
          ],
          "files": [
            "feature-mcpedit.md"
          ]
        },
        "scenarios": [
          {
            "id": "back",
            "title": "Nested cancel navigation",
            "dimension": "interaction",
            "status": "tested",
            "flows": [
              "mcp-open",
              "mcp"
            ],
            "evidence": [
              "mcp-navigation"
            ],
            "limit": "In-process keys and messages; no native terminal.",
            "summary": "Nested cancel navigation",
            "previousStatus": "tested"
          },
          {
            "id": "edit",
            "title": "Edit metadata and create a clean New record",
            "dimension": "behavior",
            "status": "partial",
            "flows": [
              "mcp"
            ],
            "evidence": [
              "mcp-metadata",
              "mcp-parser-retained",
              "mcp-cross-dashboard-saves",
              "mcp-write-failure-retry",
              "mcp-mcp-terminal-journey"
            ],
            "limit": "Isolated real repository; no restart persistence test. The changed test source needs a fresh assertion review.",
            "summary": "Edit metadata and create a clean New record",
            "previousStatus": "tested"
          },
          {
            "id": "delete",
            "title": "Delete admission and delayed request recovery",
            "dimension": "interaction",
            "status": "partial",
            "flows": [
              "mcp-delete",
              "mcp-stale"
            ],
            "evidence": [
              "mcp-stale-commands",
              "mcp-modal-target",
              "mcp-cross-dashboard-requests",
              "mcp-cross-dashboard-deletes",
              "mcp-write-failure-retry",
              "mcp-mcp-terminal-journey"
            ],
            "limit": "Same dashboard only; delayed result must not change current view. The changed test source needs a fresh assertion review.",
            "summary": "Delete admission and delayed request recovery",
            "previousStatus": "tested"
          },
          {
            "id": "policy",
            "title": "Protected and referenced definitions remain intact",
            "dimension": "behavior",
            "status": "partial",
            "flows": [
              "mcp-delete"
            ],
            "evidence": [
              "mcp-referenced",
              "mcp-protected",
              "mcp-repository-retained"
            ],
            "limit": "No injected filesystem write failure or foreign-dashboard reference integration case. The changed test source needs a fresh assertion review.",
            "summary": "Protected and referenced definitions remain intact",
            "previousStatus": "tested"
          },
          {
            "id": "bounds",
            "title": "Long list and modal stay within bounds",
            "dimension": "layout",
            "status": "unassessed",
            "flows": [
              "mcp-open",
              "mcp-delete"
            ],
            "evidence": [
              "mcp-long-list"
            ],
            "limit": "80×22 component bounds; no actual application screenshot or position-baseline comparison. The changed test source needs a fresh assertion review.",
            "summary": "Long list and modal stay within bounds",
            "captures": [
              "mcp-edited-list",
              "mcp-confirm",
              "mcp-deleted"
            ],
            "previousStatus": "tested"
          },
          {
            "id": "source-reuse",
            "title": "Existing source-list navigation remains usable",
            "dimension": "interaction",
            "status": "tested",
            "flows": [
              "mcp-card"
            ],
            "evidence": [
              "mcp-sources-retained"
            ],
            "limit": "Direct component tests only; workspace CRUD covered by suite but not exhaustively assessed here.",
            "summary": "Existing source-list navigation remains usable",
            "previousStatus": "tested"
          },
          {
            "id": "faults",
            "title": "Cross-dashboard delayed results and disk errors",
            "dimension": "behavior",
            "status": "tested",
            "flows": [
              "mcp-stale",
              "mcp-delete"
            ],
            "evidence": [
              "mcp-cross-dashboard-requests",
              "mcp-cross-dashboard-saves",
              "mcp-cross-dashboard-deletes",
              "mcp-write-failure-retry"
            ],
            "limit": "Deterministic coordinator command delivery with real file storage; ENOSPC injected at the repository boundary. Physical disk fault handling is not claimed.",
            "summary": "Cross-dashboard delayed command/result and write-error recovery boundaries now have direct Go assertions.",
            "previousStatus": "tested"
          },
          {
            "id": "visual",
            "title": "Real terminal configuration journey and captures",
            "dimension": "layout",
            "status": "tested",
            "flows": [
              "mcp-open",
              "mcp",
              "mcp-delete"
            ],
            "evidence": [
              "mcp-mcp-terminal-journey"
            ],
            "limit": "List/modal at 80×24 and forms at 120×40. Screenshot checkpoints are supported by text/storage assertions and visual inspection; no pixel-baseline comparison.",
            "summary": "Project E2E checks actual keyboard navigation, persisted create/edit/delete and captures the reached list/form/modal states.",
            "captures": [
              "mcp-list",
              "mcp-new",
              "mcp-edit",
              "mcp-edited-list",
              "mcp-confirm",
              "mcp-deleted"
            ]
          }
        ]
      },
      {
        "id": "launch-focus",
        "title": "Keep launch and naming focus on the created session",
        "area": "Dashboard",
        "summary": "New-session selection survives naming, older refreshes, detach and session-list reordering.",
        "sources": {
          "refs": [
            "views.Dashboard.FocusSession",
            "views.Dashboard.FocusSessionAfterRename",
            "ui.Coordinator.hasStaleSessionSnapshot"
          ],
          "files": [
            "internal/ui/launch_focus_internal_test.go",
            "internal/ui/inline_rename_refresh_internal_test.go",
            "tests/e2e/launch_focus_test.go"
          ]
        },
        "scenarios": [
          {
            "id": "completion",
            "title": "Select the created ID and align center",
            "summary": "Select the created ID and align center",
            "status": "unassessed",
            "dimension": "interaction",
            "evidence": [
              "focus-completion",
              "focus-search"
            ],
            "limit": "Coordinator/view fixtures; actual terminal checkpoint is separate. The changed test source needs a fresh assertion review.",
            "flows": [
              "launch-focus"
            ],
            "captures": [],
            "previousStatus": "tested"
          },
          {
            "id": "epochs",
            "title": "Reject old launch and inline-rename refreshes",
            "summary": "Reject old launch and inline-rename refreshes",
            "status": "unassessed",
            "dimension": "behavior",
            "evidence": [
              "focus-epochs",
              "rename-epochs"
            ],
            "limit": "Covers full/session/fingerprint delivery with deterministic events. The changed test source needs a fresh assertion review.",
            "flows": [
              "launch-focus"
            ],
            "captures": [],
            "previousStatus": "tested"
          },
          {
            "id": "scope",
            "title": "Keep warnings usable and ignore canceled/stale focus",
            "summary": "Keep warnings usable and ignore canceled/stale focus",
            "status": "unassessed",
            "dimension": "behavior",
            "evidence": [
              "focus-warning",
              "focus-stale",
              "focus-cancel"
            ],
            "limit": "Controlled error/scope messages; no physical filesystem outage. The changed test source needs a fresh assertion review.",
            "flows": [
              "launch-focus"
            ],
            "captures": [],
            "previousStatus": "tested"
          },
          {
            "id": "ordering",
            "title": "Keep selection stable when newer sessions arrive",
            "summary": "Keep selection stable when newer sessions arrive",
            "status": "unassessed",
            "dimension": "interaction",
            "evidence": [
              "focus-order"
            ],
            "limit": "List component identity assertion, no terminal key decoder. The changed test source needs a fresh assertion review.",
            "flows": [
              "launch-focus"
            ],
            "captures": [],
            "previousStatus": "tested"
          },
          {
            "id": "terminal",
            "title": "Launch, name, detach and reattach the same process",
            "summary": "Launch, name, detach and reattach the same process",
            "status": "tested",
            "dimension": "interaction",
            "evidence": [
              "launch-focus-terminal",
              "tui-launch-focus"
            ],
            "limit": "Codex startup only at 120×40; no model task or image baseline.",
            "flows": [
              "launch-focus"
            ],
            "captures": [
              "launch-focus-attached",
              "launch-focus-dashboard",
              "launch-focus-reattached",
              "e2e-codex-attached",
              "e2e-codex-reattached"
            ]
          }
        ]
      },
      {
        "id": "session-center",
        "title": "Selected-session center",
        "area": "Dashboard",
        "summary": "Authored reports, bounded workspace facts and cancelable event/recovery orchestration.",
        "sources": {
          "refs": [
            "api.Service.ReadSessionReports",
            "ui.Coordinator.handleCenter",
            "components.SessionCenter.View"
          ]
        },
        "scenarios": [
          {
            "id": "author-order",
            "title": "Filter authors and stable timestamp ties",
            "dimension": "behavior",
            "summary": "Rejects received/foreign reports, checks timestamp/sequence/ID order, omission, detached results, invalid limit and cancellation.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "center-author-order"
            ],
            "limit": "In-memory store fixture; not a full agent journey.",
            "sources": {
              "files": [
                "internal/session/authored_reports_test.go"
              ]
            }
          },
          {
            "id": "reload",
            "title": "Bounds and durable reconstruction across activations",
            "dimension": "behavior",
            "summary": "Reloads retained history, clamps to 256 reports, preserves activation attribution and creates no notifications.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "center-reload"
            ],
            "limit": "Synthetic persisted history; does not execute a process relaunch.",
            "sources": {
              "files": [
                "internal/session/authored_reports_test.go"
              ]
            }
          },
          {
            "id": "takeover",
            "title": "First report rejects late workspace data",
            "dimension": "behavior",
            "summary": "An event reads reports while retaining ongoing workspace work; successful report content cancels it, and late workspace data cannot overwrite reports.",
            "status": "tested",
            "flows": [
              "preview",
              "center-workspace"
            ],
            "evidence": [
              "center-takeover"
            ],
            "limit": "Direct coordinator messages and fake port; no terminal I/O.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "coalesce",
            "title": "Coalesce changes and release obsolete setup",
            "dimension": "behavior",
            "summary": "Multiple changes produce one follow-up and stale setup/current subscriptions are released after clear.",
            "status": "tested",
            "flows": [
              "preview",
              "center-recovery"
            ],
            "evidence": [
              "center-coalesce"
            ],
            "limit": "Deterministic state transitions; not an OS goroutine leak audit.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "scope",
            "title": "Scope removal clears selected content",
            "dimension": "behavior",
            "summary": "Scope rejection clears cached reports and blocks immediate retry for unchanged invalid selection.",
            "status": "tested",
            "flows": [
              "center-recovery"
            ],
            "evidence": [
              "center-scope"
            ],
            "limit": "Fake scope result; real membership separately tested.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "retry",
            "title": "One 30-second recovery timer",
            "dimension": "behavior",
            "summary": "Fake clock observes 30 seconds, excludes overlapping timer/setup and triggers resubscription first.",
            "status": "tested",
            "flows": [
              "center-recovery"
            ],
            "evidence": [
              "center-retry"
            ],
            "limit": "No wall-clock or long-lived live-terminal observation.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "failures",
            "title": "Stale reports, empty success and workspace warnings",
            "dimension": "behavior",
            "summary": "Read failure retains stale reports, confirmed empty clears them, workspace failure retains same-session data/warnings, clear drops content.",
            "status": "tested",
            "flows": [
              "preview",
              "center-workspace",
              "center-recovery"
            ],
            "evidence": [
              "center-failures"
            ],
            "limit": "Selected state-machine paths; not exhaustive cross-product of failures.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "closed",
            "title": "Closed subscription preserves reports and schedules retry",
            "dimension": "behavior",
            "summary": "Unexpected stream closure preserves cached reports and sets warning/retry state.",
            "status": "tested",
            "flows": [
              "center-recovery"
            ],
            "evidence": [
              "center-closed"
            ],
            "limit": "Tests scheduling, not a complete subsequent recovery journey.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "fit",
            "title": "Safe bounded report rendering",
            "dimension": "layout",
            "summary": "ANSI stripped; report text retained within 20×4; changing session clears it; zero size is safe.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "center-fit"
            ],
            "limit": "One explicit size plus zero-size boundary; no pixel comparison or complete resize/refetch journey.",
            "sources": {
              "files": [
                "internal/ui/components/session_center_test.go"
              ]
            }
          },
          {
            "id": "service-scope",
            "title": "Human scoped read and coalesced updates",
            "dimension": "behavior",
            "summary": "Real service reads authored progress after subscription signal, rejects foreign dashboard and allows idempotent release.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "center-service-scope"
            ],
            "limit": "Messaging harness injects records; no native model execution.",
            "sources": {
              "files": [
                "internal/api/center_test.go"
              ]
            }
          },
          {
            "id": "initial-read",
            "title": "Report before initial read and after session stop",
            "dimension": "behavior",
            "summary": "Report inserted after subscribe is read immediately and signals; stopping session keeps reports readable; cancellation closes changes.",
            "status": "tested",
            "flows": [
              "preview",
              "center-recovery"
            ],
            "evidence": [
              "center-initial-read"
            ],
            "limit": "No full relaunch journey or screenshot.",
            "sources": {
              "files": [
                "internal/api/center_test.go"
              ]
            }
          },
          {
            "id": "rearm",
            "title": "Change during signal rearming",
            "dimension": "behavior",
            "summary": "Controlled channels close while watch rearming is paused; a notice is retained and cancellation ends stream.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "center-rearm"
            ],
            "limit": "Direct subscription helper fixture, not a stress/load measurement.",
            "sources": {
              "files": [
                "internal/api/center_test.go"
              ]
            }
          },
          {
            "id": "saved-path",
            "title": "Resolve the saved session workspace",
            "dimension": "behavior",
            "summary": "Uses the session saved directory and rejects foreign scope.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-saved-path"
            ],
            "limit": "Harness has a normal workspace; temporary identity checked separately.",
            "sources": {
              "files": [
                "internal/api/center_test.go"
              ]
            }
          },
          {
            "id": "temporary-path",
            "title": "Resolve a temporary session copy",
            "dimension": "behavior",
            "summary": "Launches a cat session into a temporary copy and reads exactly its saved WorkingDir.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-temporary-path"
            ],
            "limit": "Does not invoke a model or assert recursive copy contents.",
            "sources": {
              "files": [
                "internal/api/center_test.go"
              ]
            }
          },
          {
            "id": "inventory-failure",
            "title": "Membership read errors are not empty history",
            "dimension": "behavior",
            "summary": "Corrupt card inventory makes the report read fail explicitly.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "center-inventory-failure"
            ],
            "limit": "One corrupted registry boundary.",
            "sources": {
              "files": [
                "internal/api/center_test.go"
              ]
            }
          },
          {
            "id": "git-directory",
            "title": "Clean, non-Git and full-repository changes",
            "dimension": "behavior",
            "summary": "Checks clean Git vs non-Git, staged rename plus unstaged modification and newline untracked path while reading from a subdirectory.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-git-directory"
            ],
            "limit": "Temporary real Git fixture; conflict/submodule variants not asserted here.",
            "sources": {
              "files": [
                "internal/api/center_workspace_test.go"
              ]
            }
          },
          {
            "id": "bounded-directory",
            "title": "Bound inventory, reject missing/canceled/malformed reads",
            "dimension": "behavior",
            "summary": "More than 1,000 entries truncates; missing path and canceled inspection error; malformed/partial Git records rejected.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-bounded-directory"
            ],
            "limit": "Metadata access error combinations are not exhaustive.",
            "sources": {
              "files": [
                "internal/api/center_workspace_test.go"
              ]
            }
          },
          {
            "id": "symlink-fallback",
            "title": "Symlink target and missing Git warning",
            "dimension": "behavior",
            "summary": "Immediate hidden/symlink rows retain link target; unavailable Git yields warned directory fallback.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-symlink-fallback"
            ],
            "limit": "Unix fixture, not cross-platform owner/group verification.",
            "sources": {
              "files": [
                "internal/api/center_workspace_test.go"
              ]
            }
          },
          {
            "id": "linked-worktree",
            "title": "Inspect a linked Git worktree",
            "dimension": "behavior",
            "summary": "Real linked worktree is classified as Git and reports its modified tracked file.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-linked-worktree"
            ],
            "limit": "Single worktree and tracked-path example.",
            "sources": {
              "files": [
                "internal/api/center_workspace_test.go"
              ]
            }
          },
          {
            "id": "git-timeout",
            "title": "Bound status records and cancel subprocess",
            "dimension": "behavior",
            "summary": "Status entries cap at 1,000; a busy fake Git is canceled by an explicit short deadline.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-git-timeout"
            ],
            "limit": "Does not wait out the production ten-second timeout.",
            "sources": {
              "files": [
                "internal/api/center_workspace_test.go"
              ]
            }
          },
          {
            "id": "git-output",
            "title": "Terminate Git when output exceeds limit",
            "dimension": "behavior",
            "summary": "Unbounded fake Git output fails on output budget before outer context expires.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-git-output"
            ],
            "limit": "Checks stdout budget, not every stderr variant.",
            "sources": {
              "files": [
                "internal/api/center_workspace_test.go"
              ]
            }
          },
          {
            "id": "live-resize-journey",
            "title": "Live resize, relaunch and all-failure recovery journey",
            "dimension": "interaction",
            "summary": "A complete real terminal journey through resize/refetch, session relaunch and combined failure recovery has not been reviewed for this Appedia refresh.",
            "status": "unassessed",
            "flows": [
              "preview",
              "center-workspace",
              "center-recovery"
            ],
            "evidence": [],
            "limit": "The model-free terminal cases verify resize, relaunch and report retention. A single live all-failure recovery journey remains unverified.",
            "suggestion": "Add a project-owned isolated terminal journey checking growth/refetch, relaunch continuity and recovery labels.",
            "priority": "Next",
            "reason": "Unit/state-machine assertions do not establish the complete visible journey."
          },
          {
            "id": "failed-read-clock",
            "title": "Report failures coalesce changes until timer",
            "dimension": "behavior",
            "summary": "Repeated store notices do not advance report requests while degraded; firing the fake clock allows exactly one retry and repeated failure reschedules recovery.",
            "status": "tested",
            "flows": [
              "center-recovery"
            ],
            "evidence": [
              "center-failed-read-clock"
            ],
            "limit": "Coordinator/fake-clock boundary, no live terminal.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "workspace-clock",
            "title": "Events retain cached and in-flight workspace reads",
            "dimension": "behavior",
            "summary": "Empty event reads do not refresh cached workspace before timer; a due timer refreshes it, and notices do not cancel/restart slow workspace work.",
            "status": "tested",
            "flows": [
              "center-workspace",
              "center-recovery"
            ],
            "evidence": [
              "center-workspace-clock"
            ],
            "limit": "Deterministic coordinator transitions, not real Git load timing.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "failed-workspace-clock",
            "title": "Failed initial workspace without cache waits for retry",
            "dimension": "behavior",
            "summary": "Initial load occurs immediately; after failure without cache, repeated healthy empty report reads cannot retry workspace until fake clock fires.",
            "status": "tested",
            "flows": [
              "center-workspace",
              "center-recovery"
            ],
            "evidence": [
              "center-failed-workspace-clock"
            ],
            "limit": "Tests policy state with fake application capability.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "workspace-freshness",
            "title": "Workspace stale state survives history-only reads",
            "dimension": "behavior",
            "summary": "Failed cached workspace remains stale through successful-empty and failed history reads and retry; successful workspace replacement or report takeover clears it.",
            "status": "tested",
            "flows": [
              "center-workspace",
              "center-recovery"
            ],
            "evidence": [
              "center-workspace-freshness"
            ],
            "limit": "Coordinator state assertions; complete live journey remains unassessed.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "inherited-pipes",
            "title": "Bound inherited pipes on cancellation and output overflow",
            "dimension": "behavior",
            "summary": "Fake Git starts a pipe-holding descendant; both deadline and output-limit paths finish before the child normal exit, using bounded WaitDelay.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-inherited-pipes"
            ],
            "limit": "Unix shell/helper fixture with elapsed-time tolerance; not a process-tree termination guarantee.",
            "sources": {
              "files": [
                "internal/api/center_workspace_test.go"
              ]
            }
          },
          {
            "id": "path-rows",
            "title": "Paths and symlink targets cannot forge rows",
            "dimension": "layout",
            "summary": "Git current/original paths, directory names and symlink targets flatten control whitespace; assertions reject forged status/warning rows and check display width.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-path-rows"
            ],
            "limit": "200-column renderer fixture; not pixel comparison.",
            "sources": {
              "files": [
                "internal/ui/components/session_center_test.go"
              ]
            }
          },
          {
            "id": "multiline",
            "title": "Preserve multiline report bodies separately",
            "dimension": "layout",
            "summary": "Report first/second lines retain their newline after introducing path-specific sanitization.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "center-multiline"
            ],
            "limit": "Component render at 80×8, no native terminal.",
            "sources": {
              "files": [
                "internal/ui/components/session_center_test.go"
              ]
            }
          },
          {
            "id": "navigation-reports",
            "title": "Return to cached reports and clear on empty",
            "dimension": "interaction",
            "summary": "Navigation detaches successful reports and cancels resources; return restores the same session, a failed reread renders cached reports stale, and confirmed empty remains cleared after another round trip.",
            "status": "tested",
            "flows": [
              "preview",
              "center-recovery"
            ],
            "evidence": [
              "center-navigation-reports"
            ],
            "limit": "Coordinator/navigation/component boundary with fake report reads; no native terminal journey.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "navigation-isolation",
            "title": "Navigation cache cannot move to another selection",
            "dimension": "interaction",
            "summary": "Switching selected stable session while away restores no prior-session workspace or reports.",
            "status": "tested",
            "flows": [
              "center-recovery"
            ],
            "evidence": [
              "center-navigation-isolation"
            ],
            "limit": "One dashboard/two-session fixture; not an exhaustive multi-dashboard journey.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "navigation-workspace",
            "title": "Restore workspace fallback and discard on scope loss",
            "dimension": "interaction",
            "summary": "Same-session navigation restores workspace data through a report error; scope rejection clears live and saved content.",
            "status": "tested",
            "flows": [
              "center-workspace",
              "center-recovery"
            ],
            "evidence": [
              "center-navigation-workspace"
            ],
            "limit": "Fake report/scope results, with separate service membership tests.",
            "sources": {
              "files": [
                "internal/ui/center_internal_test.go"
              ]
            }
          },
          {
            "id": "helper-cancel",
            "title": "Unix cancellation terminates owned Git helpers",
            "dimension": "behavior",
            "summary": "A controlled Git fixture starts a child process. Cancellation returns promptly, the child exits, and an unrelated process survives.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-helper-cancel"
            ],
            "limit": "Unix process-group fixture; no equivalent process-tree guarantee is established for non-Unix fallback or helpers that escape the group.",
            "sources": {
              "files": [
                "internal/api/center_process_unix_test.go"
              ]
            }
          },
          {
            "id": "helper-output",
            "title": "Unix output limit terminates helper group",
            "dimension": "behavior",
            "summary": "Fake Git with a sleep child exceeds output budget; command errors and helper is no longer running.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-helper-output"
            ],
            "limit": "Unix helper fixture; zombie state counts as terminated pending OS parent reaping.",
            "sources": {
              "files": [
                "internal/api/center_process_unix_test.go"
              ]
            }
          },
          {
            "id": "header-kind",
            "title": "Keep time and kind in final two rows",
            "dimension": "layout",
            "summary": "80×2 keeps full date/time, blocked kind and body; 20×2 uses compact time while retaining kind and body.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "center-header-kind"
            ],
            "limit": "Two explicit widths; does not assert every kind/width combination or screenshot comparison.",
            "sources": {
              "files": [
                "internal/ui/components/session_center_test.go"
              ]
            }
          },
          {
            "id": "selected-header",
            "title": "Keep the selected-session header above reports and workspace details",
            "dimension": "layout",
            "summary": "The selected session keeps its metadata, badge and puppet when the report body replaces workspace information. At minimum size, the narrower header leaves space for content.",
            "status": "tested",
            "flows": [
              "preview",
              "center-workspace",
              "resize",
              "launch-focus"
            ],
            "evidence": [
              "center-header-render",
              "center-header-size",
              "tui-launch-focus",
              "tui-codex"
            ],
            "captures": [
              "launch-focus-dashboard",
              "tui-header-minimum",
              "tui-codex-report-after-relaunch"
            ],
            "limit": "Component assertions and Linux terminal checkpoints at 120×40 and 80×24. No pixel-baseline comparison."
          },
          {
            "id": "split-index",
            "title": "Read split indexes in ordinary and linked worktrees",
            "dimension": "behavior",
            "summary": "Both ordinary and linked worktrees using split indexes return the modified tracked file as Git status.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-split-index"
            ],
            "limit": "Real local Git on Linux.",
            "sources": {
              "files": [
                "internal/api/center_workspace_test.go"
              ]
            }
          },
          {
            "id": "helper-isolation",
            "title": "Keep Git helpers out of background inspection",
            "dimension": "behavior",
            "summary": "A split-index repository reports clean status with fsmonitor disabled. Configured clean filters trigger the labelled fallback without executing the sentinel helper.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-helper-isolation"
            ],
            "limit": "Covers fsmonitor and the configured clean-filter fallback.",
            "sources": {
              "files": [
                "internal/api/center_workspace_test.go"
              ]
            }
          },
          {
            "id": "metadata-fallback",
            "title": "Retain directory entries after a Git metadata error",
            "dimension": "behavior",
            "summary": "A failing metadata probe yields a warning and preserves a readable file in the directory listing.",
            "status": "tested",
            "flows": [
              "center-workspace"
            ],
            "evidence": [
              "center-metadata-fallback"
            ],
            "limit": "A symlink loop triggers the metadata error.",
            "sources": {
              "files": [
                "internal/api/center_workspace_test.go"
              ]
            }
          },
          {
            "id": "report-signal-eviction",
            "title": "Only delivery changes that evict reports wake the center",
            "dimension": "behavior",
            "summary": "An ordinary delivery update wakes general waiters only. Evicting retained replies wakes report readers.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "center-report-signal-eviction"
            ],
            "limit": "Uses a near-capacity message store with legacy replies.",
            "sources": {
              "files": [
                "internal/session/messaging_review_test.go"
              ]
            }
          },
          {
            "id": "report-signal-commit",
            "title": "Ignore report retries and failed writes",
            "dimension": "behavior",
            "summary": "A persisted report wakes report readers. Replaying the report or failing a new write leaves the signal unchanged.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "center-report-signal-commit"
            ],
            "limit": "Message-store test with a forced persistence error.",
            "sources": {
              "files": [
                "internal/session/messaging_review_test.go"
              ]
            }
          },
          {
            "id": "report-scope-signal",
            "title": "Recheck scope independently of report updates",
            "dimension": "behavior",
            "summary": "The subscription closes on scope loss without requiring a report change.",
            "status": "tested",
            "flows": [
              "center-recovery"
            ],
            "evidence": [
              "center-report-scope-signal"
            ],
            "limit": "Application service fixture.",
            "sources": {
              "files": [
                "internal/api/center_test.go"
              ]
            }
          },
          {
            "id": "report-timestamp-fit",
            "title": "Omit reports that cannot keep a timestamp",
            "dimension": "layout",
            "summary": "When an older report cannot fit with its timestamp, omit it. A newest report can use a compact one-row timestamp.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "center-report-timestamp-fit"
            ],
            "limit": "Component assertions at bounded sizes.",
            "sources": {
              "files": [
                "internal/ui/components/session_center_test.go"
              ]
            }
          }
        ]
      },
      {
        "id": "edit-effort",
        "title": "Configure reasoning effort",
        "area": "Agent cards",
        "summary": "Edit, save and clear effort using the card editor and existing repository.",
        "sources": {
          "refs": [
            "views.CardEditor.Draft",
            "config.IDCard",
            "api.applyCardChanges"
          ]
        },
        "scenarios": [
          {
            "id": "draft",
            "title": "Enter a value and navigate the field",
            "dimension": "layout",
            "summary": "The draft trims surrounding whitespace and the focused row stays visible at the supported minimum.",
            "status": "tested",
            "flows": [
              "editcard"
            ],
            "evidence": [
              "effort-draft",
              "effort-layout"
            ],
            "limit": "Go component assertions. No recorded terminal editor journey for this field."
          },
          {
            "id": "save",
            "title": "Create, preserve and clear effort",
            "dimension": "behavior",
            "summary": "The card value round-trips through the repository. Name-only edits preserve it and an explicit empty edit clears it.",
            "status": "tested",
            "flows": [
              "savecard"
            ],
            "evidence": [
              "effort-create",
              "effort-preserve",
              "effort-clear"
            ],
            "limit": "Service and repository coverage. The terminal form submission is outside these cases."
          },
          {
            "id": "guard",
            "title": "Respect revisions and read-only access",
            "dimension": "interaction",
            "summary": "A concurrent effort edit invalidates a stale save. The shared card consumer retains the read-only editor guard.",
            "status": "tested",
            "flows": [
              "editcard",
              "savecard"
            ],
            "evidence": [
              "effort-clear",
              "editor-read-only"
            ],
            "limit": "Separate service and view-level tests. No effort-specific consumer typing assertion."
          },
          {
            "id": "terminal",
            "title": "Edit and reopen a card in a real terminal",
            "dimension": "interaction",
            "summary": "The existing launch-focus E2E journey sets high, saves the card and verifies high on the Reasoning effort row after reopening.",
            "status": "tested",
            "flows": [
              "editcard",
              "savecard"
            ],
            "limit": "Tests the real terminal at 120×40. Clearing remains covered by the Go service fixture.",
            "evidence": [
              "e2e-card-effort"
            ],
            "captures": [
              "e2e-card-effort-reopened"
            ]
          }
        ]
      },
      {
        "id": "apply-effort",
        "title": "Apply reasoning effort on activation",
        "area": "Sessions",
        "summary": "Argument construction adds the saved effort before conversation arguments at both spawn sites.",
        "sources": {
          "refs": [
            "session.activationArgs",
            "driver.AgentDriver"
          ]
        },
        "scenarios": [
          {
            "id": "dispatch",
            "title": "Dispatch supported mappings and generic no-op",
            "dimension": "behavior",
            "summary": "Claude and Agy return --effort, Codex emits a configuration override, and OpenCode appends no effort arguments.",
            "status": "tested",
            "flows": [
              "reasoning-effort-activation"
            ],
            "evidence": [
              "effort-registry"
            ],
            "limit": "Exact argv at the driver boundary. No native Agy invocation."
          },
          {
            "id": "launch-resume-switch",
            "title": "Capture activation arguments",
            "dimension": "behavior",
            "summary": "Capture arguments on launch, relaunch with a changed effort, and switch to and from a card with empty effort.",
            "status": "tested",
            "flows": [
              "launch",
              "resume",
              "reasoning-effort-activation"
            ],
            "evidence": [
              "effort-claude",
              "effort-codex"
            ],
            "limit": "Fixture process boundary. The Codex ordering test uses a saved-conversation fixture."
          },
          {
            "id": "saved-source",
            "title": "Resolve a saved card through the shared service",
            "dimension": "behavior",
            "summary": "Retained tests check latest saved card identity and deployment behavior while manager fixtures cover effort argv.",
            "status": "partial",
            "flows": [
              "launch",
              "resume",
              "reasoning-effort-activation"
            ],
            "evidence": [
              "latest-launch",
              "latest-switch",
              "effort-claude"
            ],
            "limit": "The tests do not combine a persisted effort edit, shared service resolution and captured argv in one case.",
            "suggestion": "Extend a service fixture to capture argv after changing the persisted effort and relaunching.",
            "priority": "Later",
            "reason": "Joins the independently verified storage and activation boundaries."
          },
          {
            "id": "native",
            "title": "Native CLI acceptance",
            "dimension": "behavior",
            "summary": "The launch-focus journey opens idle Codex after saving high. Its screenshot shows GPT-6-Sol high. Fixtures check the exact effort argv.",
            "status": "partial",
            "flows": [
              "reasoning-effort-activation",
              "resume"
            ],
            "evidence": [
              "effort-registry",
              "effort-codex",
              "e2e-launch-focus"
            ],
            "limit": "The footer is a visual observation. Native Claude, Agy and Codex saved-conversation resume remain outside this run.",
            "captures": [
              "e2e-codex-attached"
            ]
          }
        ]
      }
    ],
    "evidence": [
      {
        "id": "boundary",
        "name": "TestFrontendBoundary",
        "method": "automated",
        "integration": "project",
        "file": "internal/architecture/boundary_test.go",
        "line": 39,
        "endLine": 56,
        "proves": "Checks package direction and reviewed exception sets.",
        "limits": "The checker is static. Runtime behavior has separate evidence.",
        "reviewedHash": "7d5b2dae45b6a152df35434f934a406ff8d571c021100a6313a4ba2f34c99c5f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.012769099Z",
          "scope": "TestFrontendBoundary in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "card-save",
        "name": "TestEditorSavePreservesUneditableFieldsAndRejectsStaleRevision",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/card_editor_test.go",
        "line": 13,
        "endLine": 72,
        "proves": "Checks retained fields and rejection of a stale revision.",
        "limits": "API and repository seam.",
        "reviewedHash": "c10ea2d898eac2df31e809c0a1c13fa0074b198315d96daaa04c3bc9c9162374",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./...",
          "at": "2026-11-21T21:47:16.346390935Z",
          "scope": "TestEditorSavePreservesUneditableFieldsAndRejectsStaleRevision",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "card-race",
        "name": "TestEditorConcurrentSavesOnlyOneSucceeds",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/card_editor_test.go",
        "line": 74,
        "endLine": 109,
        "proves": "Only one of two saves with the same revision succeeds.",
        "limits": "Concurrent service updates in a controlled fixture.",
        "reviewedHash": "c10ea2d898eac2df31e809c0a1c13fa0074b198315d96daaa04c3bc9c9162374",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./...",
          "at": "2026-11-21T21:47:16.353339783Z",
          "scope": "TestEditorConcurrentSavesOnlyOneSucceeds",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "source-edit",
        "name": "TestSourceIdentityRejectsReorderAndAddressesDuplicates",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/configuration_sources_test.go",
        "line": 15,
        "endLine": 50,
        "proves": "Rejects stale row identity after reorder and addresses duplicate rows.",
        "limits": "Backend source mutation path.",
        "reviewedHash": "78aec81ee4767f1306b28c9e041709521bf44e57f3431c8855fb2ee5f7d2bd02",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:00.13931537Z",
          "scope": "TestSourceIdentityRejectsReorderAndAddressesDuplicates in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "conversation",
        "name": "TestSessionContractsRevalidateScopeAndChoice",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/session_actions_internal_test.go",
        "line": 45,
        "endLine": 103,
        "proves": "Checks scoped action revalidation and stale conversation choice handling.",
        "limits": "Controlled backend session fixture.",
        "reviewedHash": "e2870b54ecd260896b9782a8394e38f4a5c45eaa20ac8eb8d091c43fdc74aff7",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.965299561Z",
          "scope": "TestSessionContractsRevalidateScopeAndChoice",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "center-history",
        "name": "TestCenterReportSurvivesBackendReconstruction",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/center_integration_internal_test.go",
        "line": 16,
        "endLine": 72,
        "proves": "Reads a retained report through a reconstructed manager and coordinator.",
        "limits": "Exercises backend/UI seam without rendering a live terminal restart.",
        "reviewedHash": "6f5ad56984d7167bf861bb7b3ff437e1defad2c43b52799d0007756372ac631a",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.765947188Z",
          "scope": "TestCenterReportSurvivesBackendReconstruction in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-switch",
        "name": "TestCenterSwitchWhileReadInFlight",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/center_internal_test.go",
        "line": 513,
        "endLine": 534,
        "proves": "Changes selection during a read and checks stale data cannot populate the new center.",
        "limits": "Controlled coordinator timing.",
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.795438756Z",
          "scope": "TestCenterSwitchWhileReadInFlight in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "pty-detach",
        "name": "TestTerminalPTYDetachAndRestoration",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/terminal/pty_test.go",
        "line": 66,
        "endLine": 145,
        "proves": "Checks resize, legacy and Kitty detach, and equality of terminal state before and after.",
        "limits": "Linux PTY fixture.",
        "reviewedHash": "a4359460934eee6d61f2ffd6dc5b2f4a3b9e340a3b2b7081bbf9484334ff2eb6",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.517785348Z",
          "scope": "TestTerminalPTYDetachAndRestoration in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "pty-split",
        "name": "TestInputDecoderSplitDetach",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/terminal/attach_test.go",
        "line": 116,
        "endLine": 131,
        "proves": "Feeds the detach bytes in two reads and checks that only preceding ordinary bytes were forwarded.",
        "limits": "Deterministic decoder test. The full binary run corroborates the behavior.",
        "reviewedHash": "f54b9ca0a25861678b6ec4d5fe79adcd858b18f35ca732426ae3d101cf0d1630",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.315513593Z",
          "scope": "TestInputDecoderSplitDetach in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "missing-attach",
        "name": "TestMissingSessionClosesQueuedAttachment",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/attachment_missing_internal_test.go",
        "line": 47,
        "endLine": 77,
        "proves": "Checks immediate close and rejection after full and sessions-only refresh confirm absence.",
        "limits": "Coordinator fixture with controlled ownership.",
        "reviewedHash": "2f535a287c4801d876f2f175f4bcc6a8b6887075989db8456c4bb5eb92058aba",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.330566382Z",
          "scope": "TestMissingSessionClosesQueuedAttachment in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "mcp-scope",
        "name": "TestDashboardScopeOverHTTP",
        "method": "automated",
        "integration": "project",
        "file": "internal/mcp/dashboard_scope_test.go",
        "line": 22,
        "endLine": 228,
        "proves": "Exercises dashboard scope through the MCP server transport.",
        "limits": "Local transport and test identities.",
        "reviewedHash": "03413effbbfe574e7fc1d94717cb6a6a1641cfe6cc4541163ed1d518ac2e074e",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:15.089580079Z",
          "scope": "TestDashboardScopeOverHTTP",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "delete",
        "name": "TestDashboardExecutionProgressSurvivesStatusReadFailure",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/deletion_presentation_internal_test.go",
        "line": 97,
        "endLine": 137,
        "proves": "Checks execution progress stays visible when a subsequent status read fails.",
        "limits": "Coordinator and view fixture.",
        "reviewedHash": "5ce881d26236dad53fe14596642d9e4fb51841d8875f19c7640d83b8b2258bf1",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:04.132758481Z",
          "scope": "TestDashboardExecutionProgressSurvivesStatusReadFailure in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "tui-codex",
        "name": "TestCodexOfflineWorkspaceAndReports/directory",
        "kind": "E2E",
        "method": "automated",
        "integration": "project",
        "file": "tests/e2e/codex_offline_test.go",
        "proves": "Launches, names, detaches, reattaches, stops and relaunches idle Codex against a directory workspace. Checks refresh, size recovery, report retention and Kankuro restart.",
        "limits": "Linux only. The report is seeded through the real store. Codex receives no model prompts.",
        "reviewedHash": "86f758dde6cd5f32afdf4c6985b3cc65d73b23ce15c5fc287775fff5e82599ad",
        "result": {
          "status": "passed",
          "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/directory$' -count=1 -timeout=3m -v",
          "at": "2026-11-19T21:26:35.969656818Z",
          "scope": "Launches, names, detaches, reattaches, stops and relaunches idle Codex against a directory workspace. Checks refresh, size recovery, report retention and Kankuro restart.",
          "revision": "WORKTREE",
          "fingerprint": "1594c215e9ef7de8ae058411320bd1b0ffc6420de9a370d636424f33cb67a9ce"
        },
        "line": 19,
        "endLine": 31
      },
      {
        "id": "tui-launch-focus",
        "name": "TestNativeLaunchFocusReturn",
        "kind": "E2E",
        "method": "automated",
        "integration": "project",
        "file": "tests/e2e/launch_focus_test.go",
        "proves": "Checks dashboard focus after detach and reattaches to the same Codex PID.",
        "limits": "Codex starts with a dummy credential and receives no prompt. Execution shown here is from the earlier checkout. This refresh compiled the tagged package and ran the model-free capture check.",
        "reviewedHash": "b3322556b2918002831d1c7c5b35f9616ad746410f268d53a633f8d6ad461edb",
        "result": {
          "status": "passed",
          "command": "go test -tags=e2e -count=1 -timeout=8m -json ./tests/e2e",
          "at": "2026-11-21T21:33:33.71233779Z",
          "scope": "TestNativeLaunchFocusReturn",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        },
        "line": 28,
        "endLine": 33
      },
      {
        "id": "tui-codex-git",
        "name": "TestCodexOfflineWorkspaceAndReports/git",
        "kind": "E2E",
        "method": "automated",
        "integration": "project",
        "file": "tests/e2e/codex_offline_test.go",
        "proves": "Launches, names, detaches, reattaches, stops and relaunches idle Codex against a git workspace. Checks refresh, size recovery, report retention and Kankuro restart.",
        "limits": "Linux only. The report is seeded through the real store. Codex receives no model prompts.",
        "reviewedHash": "86f758dde6cd5f32afdf4c6985b3cc65d73b23ce15c5fc287775fff5e82599ad",
        "result": {
          "status": "passed",
          "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/git$' -count=1 -timeout=3m -v",
          "at": "2026-11-19T21:27:11.45966887Z",
          "scope": "Launches, names, detaches, reattaches, stops and relaunches idle Codex against a git workspace. Checks refresh, size recovery, report retention and Kankuro restart.",
          "revision": "WORKTREE",
          "fingerprint": "1594c215e9ef7de8ae058411320bd1b0ffc6420de9a370d636424f33cb67a9ce"
        },
        "line": 19,
        "endLine": 31
      },
      {
        "id": "tui-mcp",
        "name": "TestMCPManagerJourney",
        "kind": "E2E",
        "method": "automated",
        "integration": "project",
        "file": "tests/e2e/mcp_manager_test.go",
        "proves": "Creates, edits and deletes a custom MCP, checks persistence and cancel, and preserves the protected built-in at 80 by 24 cells.",
        "limits": "Registry interaction only. The custom service is never started.",
        "reviewedHash": "0b63e84ab8c3c871017132e66b9a9c3866d6d83d517e59ca1d71a3710a4b92bf",
        "result": {
          "status": "passed",
          "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
          "at": "2026-11-19T21:27:17.210608058Z",
          "scope": "Creates, edits and deletes a custom MCP, checks persistence and cancel, and preserves the protected built-in at 80 by 24 cells.",
          "revision": "WORKTREE",
          "fingerprint": "1594c215e9ef7de8ae058411320bd1b0ffc6420de9a370d636424f33cb67a9ce"
        },
        "line": 27,
        "endLine": 113
      },
      {
        "id": "tui-luna",
        "name": "TestNativeAgentJourney",
        "kind": "E2E",
        "method": "automated",
        "integration": "project",
        "file": "tests/e2e/codex_token_usage_test.go",
        "proves": "The 60,000-input-budget run registered a Codex conversation and confirmed the first live progress report and its attribution.",
        "limits": "The first turn consumed 129,209 input tokens (117,376 cached) and 1,013 output tokens. The total input usage check failed before resume or recall. No second prompt was sent. Execution shown here is from the earlier checkout. This refresh compiled the tagged package and ran the model-free capture check.",
        "reviewedHash": "32d2db4028633b144215cd8e07c54073a63aca29dce22d8d0f1c5495728d9292",
        "result": {
          "status": "failed",
          "command": "./tests/e2e/run token-usage",
          "at": "2026-11-18T21:47:36.773848345Z",
          "scope": "The 60,000-input-budget run registered a Codex conversation and confirmed the first live progress report and its attribution.",
          "revision": "WORKTREE",
          "fingerprint": "68189939a8365fc1cb81f7b61d700cffdea80f639bca768fdabd0809d0bd2b6e"
        },
        "line": 37,
        "endLine": 91
      },
      {
        "id": "fresh-initialization-and-partial-root",
        "name": "TestFreshInitializationAndPartialRoot",
        "method": "automated",
        "integration": "project",
        "file": "internal/configstartup/startup_test.go",
        "line": 10,
        "endLine": 30,
        "proves": "Checks that fresh startup validates and that a partial root fails with the missing-marker diagnostic.",
        "limits": "Calls configuration startup directly; does not start the full TUI/MCP composition root.",
        "reviewedHash": "23def849d1afd7463ecca53123cd3035b705f5cd4ef39e1b7a2fc5782fd55c31",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.063401182Z",
          "scope": "TestFreshInitializationAndPartialRoot in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "validate-layout-rejects-invalid-records-without-rewriting",
        "name": "TestValidateLayoutRejectsInvalidRecordsWithoutRewriting",
        "method": "automated",
        "integration": "project",
        "file": "internal/configstartup/validate_test.go",
        "line": 20,
        "endLine": 48,
        "proves": "Validates a known-good fixture before each corrupt session/operation/message/cursor/card record, then checks rejection and exact original bytes.",
        "limits": "Five explicit fixture cases; not arbitrary schema fuzzing.",
        "reviewedHash": "a79b37b8df4a3e2f94e517171046fa8d1f212cd9e52cb5075ac05f7c7789b9ca",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.077244132Z",
          "scope": "TestValidateLayoutRejectsInvalidRecordsWithoutRewriting in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "exclusive-lock",
        "name": "TestExclusiveLock",
        "method": "automated",
        "integration": "project",
        "file": "internal/configstartup/startup_test.go",
        "line": 42,
        "endLine": 64,
        "proves": "A second AcquireLock fails while held and succeeds after Close.",
        "limits": "Same-process lock contention on Linux; no cross-platform or subprocess assertion.",
        "reviewedHash": "23def849d1afd7463ecca53123cd3035b705f5cd4ef39e1b7a2fc5782fd55c31",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.063552322Z",
          "scope": "TestExclusiveLock in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "dashboard-list-creates-named-dashboard",
        "name": "TestDashboardListCreatesNamedDashboard",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_list_test.go",
        "reviewedHash": "2217d2cb2378bd49456d51cb12c352bee8f4033d229fe5297211727dea37c806",
        "proves": "Drives n, rejects empty name, edits name/description and asserts callback values.",
        "limits": "Create-form callback only; no edit-form submission, repository write or selected-ID refresh assertion.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.578769989Z",
          "scope": "TestDashboardListCreatesNamedDashboard in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 171,
        "endLine": 190
      },
      {
        "id": "dashboard-details-through-service",
        "name": "TestDashboardDetailsThroughService",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/membership_internal_test.go",
        "reviewedHash": "5c03e0ca7fa22fb2f702ffc7ab2a5ee661a6fb521c75935bdd83dc1719f97f46",
        "proves": "Creates and updates dashboard details, then checks the single returned inventory entry equals the update and retains the created identity.",
        "limits": "Service/repository test; no dashboard-list form key journey.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.427411117Z",
          "scope": "TestDashboardDetailsThroughService in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 348,
        "endLine": 369
      },
      {
        "id": "dashboard-list-shows-deletion-preparation-and-status-errors",
        "name": "TestDashboardListShowsDeletionPreparationAndStatusErrors",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_list_test.go",
        "reviewedHash": "2217d2cb2378bd49456d51cb12c352bee8f4033d229fe5297211727dea37c806",
        "proves": "At three sizes, drives preflight/status keys, feeds multiline/ANSI failure messages and checks visible bounded errors that clear after cancel or successful retry.",
        "limits": "View/message assertions; fake callbacks do not execute the deletion service. Helper assertDashboardDeletionErrorVisible supplies the bounds/sanitization oracle.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.5642564Z",
          "scope": "TestDashboardListShowsDeletionPreparationAndStatusErrors in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 28,
        "endLine": 69
      },
      {
        "id": "dashboard-list-bounds-long-deletion-error-when-list-becomes-empty",
        "name": "TestDashboardListBoundsLongDeletionErrorWhenListBecomesEmpty",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_list_test.go",
        "reviewedHash": "2217d2cb2378bd49456d51cb12c352bee8f4033d229fe5297211727dea37c806",
        "proves": "Keeps a long CJK cleanup error visible after the dashboard inventory becomes empty and checks the view stays within 80×21.",
        "limits": "One narrow view state; not a screenshot or actual retry service run.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.56558524Z",
          "scope": "TestDashboardListBoundsLongDeletionErrorWhenListBecomesEmpty in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 82,
        "endLine": 94
      },
      {
        "id": "read-dashboard-data-scopes-cards-and-sessions",
        "name": "TestReadDashboardDataScopesCardsAndSessions",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/membership_internal_test.go",
        "proves": "Checks private/shared visibility, consumer edit denial, dashboard-local sessions, stale/foreign deletion confirmation, surviving foreign sessions and switching away from a deleted card.",
        "limits": "Service/repository/manager fixture with cat processes; no keyboard or MCP transport assertions.",
        "reviewedHash": "5c03e0ca7fa22fb2f702ffc7ab2a5ee661a6fb521c75935bdd83dc1719f97f46",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.415437928Z",
          "scope": "TestReadDashboardDataScopesCardsAndSessions in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 139,
        "endLine": 308
      },
      {
        "id": "scoped-launch-revalidates-changed-card-before-publication",
        "name": "TestScopedLaunchRevalidatesChangedCardBeforePublication",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/membership_internal_test.go",
        "proves": "Blocks copying, changes sharing, then asserts no publication and release of the pending reservation.",
        "limits": "Injected workspace-copy gate; does not simulate every registry mutation.",
        "reviewedHash": "5c03e0ca7fa22fb2f702ffc7ab2a5ee661a6fb521c75935bdd83dc1719f97f46",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.460837244Z",
          "scope": "TestScopedLaunchRevalidatesChangedCardBeforePublication in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 403,
        "endLine": 458
      },
      {
        "id": "dashboard-scope-over-h-t-t-p",
        "name": "TestDashboardScopeOverHTTP",
        "method": "automated",
        "integration": "project",
        "file": "internal/mcp/dashboard_scope_test.go",
        "line": 22,
        "endLine": 228,
        "reviewedHash": "03413effbbfe574e7fc1d94717cb6a6a1641cfe6cc4541163ed1d518ac2e074e",
        "proves": "Authenticates two dashboards over the real Streamable HTTP SDK client, checks peer session discovery excludes caller and foreign sessions.",
        "limits": "Only list_sessions transport scope is asserted here; no external AI client.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:15.089580079Z",
          "scope": "TestDashboardScopeOverHTTP",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "operation-runner-creates-once-and-replays-request",
        "name": "TestOperationRunnerCreatesOnceAndReplaysRequest",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 164,
        "endLine": 504,
        "reviewedHash": "4c2f3e9342ceea29d5ff1d5ba392ce0d5c67b3e8b03c38bff8421942ca9bc9df",
        "proves": "With real cat sessions, checks deployed controller/registration-only authority, scoped discovery/inspection/create, duplicate concurrent create identity, request conflicts, cross-dashboard denials, lifecycle replay/restart and cancellation.",
        "limits": "In-process service/runner fixture; no TUI or installed native-agent client. Tests the declared finite set of authorization and operation cases.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.434609084Z",
          "scope": "TestOperationRunnerCreatesOnceAndReplaysRequest",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "unauthenticated-discovery-is-denied",
        "name": "TestUnauthenticatedDiscoveryIsDenied",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 150,
        "endLine": 161,
        "reviewedHash": "4c2f3e9342ceea29d5ff1d5ba392ce0d5c67b3e8b03c38bff8421942ca9bc9df",
        "proves": "Checks empty credentials cannot list cards/sessions or inspect a target.",
        "limits": "Service rejection assertions without transport status assertions.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.145536991Z",
          "scope": "TestUnauthenticatedDiscoveryIsDenied",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "confirmed-card-deletion-cancels-foreign-pending-launch",
        "name": "TestConfirmedCardDeletionCancelsForeignPendingLaunch",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/membership_internal_test.go",
        "proves": "Asserts foreign pending impact, required confirmation, cancelled launch and no published session.",
        "limits": "Controlled copier and service entry points; no confirmation-modal rendering.",
        "reviewedHash": "5c03e0ca7fa22fb2f702ffc7ab2a5ee661a6fb521c75935bdd83dc1719f97f46",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.474424083Z",
          "scope": "TestConfirmedCardDeletionCancelsForeignPendingLaunch in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 461,
        "endLine": 501
      },
      {
        "id": "dashboard-cascade-resumes-after-restart",
        "name": "TestDashboardCascadeResumesAfterRestart",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/dashboard_deletion_internal_test.go",
        "line": 132,
        "endLine": 184,
        "proves": "Persists deleting intent, checks admission denial, rebuilds manager/service, removes owned session while preserving the foreign session and permits idempotent retry.",
        "limits": "In-process restart against real files; no OS kill/power-loss simulation.",
        "reviewedHash": "abc20c154ff0e19f4923ad649195d14945e477961cb8114cf97ad6bb7736398a",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.211417864Z",
          "scope": "TestDashboardCascadeResumesAfterRestart in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "card-deletion-retries-pending-workspace-cleanup-after-restart",
        "name": "TestCardDeletionRetriesPendingWorkspaceCleanupAfterRestart",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/card_cleanup_internal_test.go",
        "reviewedHash": "d9286cfb104e002db652973ffe28a6085d9b023f762c4ad4a983146bca3157f7",
        "proves": "Blocks a foreign pending copy, edits the card workspace, checks failed cleanup persists the original exact paths/error in a tombstone, then restarts and retries cleanup without removing a foreign session.",
        "limits": "Permission-based non-root filesystem fixture; in-process restart, not hard OS termination.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.075301782Z",
          "scope": "TestCardDeletionRetriesPendingWorkspaceCleanupAfterRestart in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 16,
        "endLine": 167
      },
      {
        "id": "multi-dashboard-discards-stale-results-and-preserves-views",
        "name": "TestMultiDashboardDiscardsStaleResultsAndPreservesViews",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/multi_dashboard_internal_test.go",
        "proves": "Checks that foreign sessions cannot take over the center, stale settings navigation, refresh and attach cannot cross dashboard scope, and returning restores the original view.",
        "limits": "Drives coordinator messages directly with real local cat sessions; not an end-to-end terminal journey.",
        "reviewedHash": "7749e986fb4965d6a54c9b4ec20cedf1b4fb4a725b5a8604c85a1c90ac011ccb",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:04.40770914Z",
          "scope": "TestMultiDashboardDiscardsStaleResultsAndPreservesViews in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 20,
        "endLine": 111
      },
      {
        "id": "maybe-refresh-config-skips-unchanged-registry",
        "name": "TestMaybeRefreshConfigSkipsUnchangedRegistry",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/refresh_cadence_internal_test.go",
        "proves": "Counts ListIDCards calls and asserts an unchanged fingerprint produces no additional registry read.",
        "limits": "Counting repository fixture; does not establish every filesystem timestamp boundary.",
        "reviewedHash": "afde45f722205d5aefb659b9755f98ae98b0d562fd12fc339b266aef4d0be0df",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:23.432907046Z",
          "scope": "TestMaybeRefreshConfigSkipsUnchangedRegistry",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        },
        "line": 67,
        "endLine": 94
      },
      {
        "id": "coordinator-too-small-screen-is-exact",
        "name": "TestCoordinatorTooSmallScreenIsExact",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/size_guard_test.go",
        "line": 26,
        "endLine": 45,
        "proves": "Sweeps widths 1–85 and heights 1–27; checks exact notice rectangle below 80×24 and notice removal at supported sizes.",
        "limits": "String/cell layout assertions; no emulator font or image comparison.",
        "reviewedHash": "0801b02c0bbba2fc83c94692410ff7ad170275438ea78e8a12e56f587f220950",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.522864064Z",
          "scope": "TestCoordinatorTooSmallScreenIsExact in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "coordinator-drops-keys-when-too-small",
        "name": "TestCoordinatorDropsKeysWhenTooSmall",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/size_guard_test.go",
        "line": 54,
        "endLine": 69,
        "proves": "Checks that n, Enter, Tab, Backspace and c return no commands and do not change the active view at three undersized dimensions.",
        "limits": "Selected keys and sizes only; paste is not asserted here.",
        "reviewedHash": "0801b02c0bbba2fc83c94692410ff7ad170275438ea78e8a12e56f587f220950",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.57031494Z",
          "scope": "TestCoordinatorDropsKeysWhenTooSmall in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "dashboard-list-requires-explicit-quit",
        "name": "TestDashboardListRequiresExplicitQuit",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/dashboard_list_keys_test.go",
        "line": 11,
        "endLine": 44,
        "proves": "Checks repeated Esc/Ctrl+C keep the dashboard list open at normal and small sizes and q returns QuitMsg.",
        "limits": "Coordinator-level input assertions; no physical terminal signal delivery.",
        "reviewedHash": "bf4a7614f4eabbcc3f4871f1baa0fb1bceaa11a211b6a1af3dc8e25c11a552ec",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.067406993Z",
          "scope": "TestDashboardListRequiresExplicitQuit in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "card-editor-preserves-unmanaged-fields",
        "name": "TestEditorSavePreservesUneditableFieldsAndRejectsStaleRevision",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/card_editor_test.go",
        "proves": "Saves a name-only draft through the editor service, compares every other persisted field apart from UpdatedAt, and rejects a stale revision.",
        "limits": "Service and repository assertions. The UI draft is covered separately.",
        "reviewedHash": "c10ea2d898eac2df31e809c0a1c13fa0074b198315d96daaa04c3bc9c9162374",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./...",
          "at": "2026-11-21T21:47:16.346390935Z",
          "scope": "TestEditorSavePreservesUneditableFieldsAndRejectsStaleRevision",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        },
        "line": 13,
        "endLine": 72
      },
      {
        "id": "dashboard-control-card-editor-toggle-round-trip",
        "name": "TestDashboardControlCardEditorToggleRoundTrip",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/card_editor_test.go",
        "proves": "Prior assessment: Focuses the control row, toggles off/on and checks Card().DashboardControl.",
        "limits": "Does not prove deployed activation authority or MCP permission changes. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:58.282823223Z",
          "scope": "TestDashboardControlCardEditorToggleRoundTrip in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 117,
        "endLine": 132
      },
      {
        "id": "card-editor-focused-field-always-visible",
        "name": "TestCardEditorFocusedFieldAlwaysVisible",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/card_editor_layout_test.go",
        "proves": "Checks every expected focus target remains visible with optional sections on/off, empty/large lists and minimum-sized form heights.",
        "limits": "Rendered text and model focus; no screenshot, native terminal or image baseline.",
        "reviewedHash": "5d0e2dc83803768c7b520d21f5cb4ff06b25a5743664836064aaea4c05cf88c6",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./...",
          "at": "2026-11-21T21:47:19.200096895Z",
          "scope": "TestCardEditorFocusedFieldAlwaysVisible",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        },
        "line": 263,
        "endLine": 286
      },
      {
        "id": "selected-only-skill-cursor-normalizes-after-registry-refresh",
        "name": "TestSelectedOnlySkillCursorNormalizesAfterRegistryRefresh",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/card_editor_test.go",
        "reviewedHash": "ea3731dd18253155280647a4c78b98ef9b808b9de498ca3f60f679aa78fdc0e2",
        "proves": "Enables selected-only directly, removes a selected skill during registry refresh, checks cursor normalization to A and that Space toggles A.",
        "limits": "Does not send Ctrl+S or assert header indicator, search intersection or re-enabling all items.",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./...",
          "at": "2026-11-21T21:47:19.880783948Z",
          "scope": "TestSelectedOnlySkillCursorNormalizesAfterRegistryRefresh",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        },
        "line": 138,
        "endLine": 158
      },
      {
        "id": "temporary-workspace-space-save",
        "name": "TestTemporaryWorkspaceSpaceAndSave",
        "method": "automated",
        "integration": "project",
        "kind": "unit",
        "file": "internal/ui/views/card_editor_test.go",
        "proves": "Starts with each temporary-workspace setting, toggles it with Space, and checks the saved draft materializes to the expected card.",
        "limits": "Card editor model/key/callback boundary; no terminal screenshot or repository write.",
        "reviewedHash": "ea3731dd18253155280647a4c78b98ef9b808b9de498ca3f60f679aa78fdc0e2",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./...",
          "at": "2026-11-21T21:47:19.915031553Z",
          "scope": "TestTemporaryWorkspaceSpaceAndSave",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        },
        "line": 581,
        "endLine": 618
      },
      {
        "id": "ctrl-x-clears-only-sub-workspace-selection",
        "name": "TestCtrlXClearsOnlySubWorkspaceSelection",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/card_editor_test.go",
        "proves": "Checks Ctrl+X clears selected/missing IDs, preserves all other card fields and focus, invokes no save/cancel, shows the hint and ignores plain x.",
        "limits": "In-memory editor; filesystem copies and registry writes are outside this test.",
        "reviewedHash": "ea3731dd18253155280647a4c78b98ef9b808b9de498ca3f60f679aa78fdc0e2",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./...",
          "at": "2026-11-21T21:47:19.882619498Z",
          "scope": "TestCtrlXClearsOnlySubWorkspaceSelection",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        },
        "line": 160,
        "endLine": 211
      },
      {
        "id": "sub-workspace-picker-multiple-unresolved-and-main-modes",
        "name": "TestSubWorkspacePickerMultipleUnresolvedAndMainModes",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/workspace_picker_test.go",
        "reviewedHash": "2b3e2e4efc955282f3e19101aa8d0a0576dd8a48c9488d80067e56e1cab7c9b2",
        "proves": "Drives the sub-picker through missing selection repair, search, provisional toggles, cancel, confirmation of two IDs, main-picker mode reset and saving both sub IDs.",
        "limits": "In-memory form/view sequence; does not copy sub-workspace files or persist the card repository.",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./...",
          "at": "2026-11-21T21:47:27.221062022Z",
          "scope": "TestSubWorkspacePickerMultipleUnresolvedAndMainModes",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        },
        "line": 77,
        "endLine": 129
      },
      {
        "id": "sub-workspace-new-selection-survives-registry-refresh",
        "name": "TestSubWorkspaceNewSelectionSurvivesRegistryRefresh",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/workspace_picker_test.go",
        "reviewedHash": "2b3e2e4efc955282f3e19101aa8d0a0576dd8a48c9488d80067e56e1cab7c9b2",
        "proves": "Checks a newly selected workspace remains visible as missing after refresh, cancel does not commit, explicit missing-row deselection works, and confirming a missing checked ID retains it.",
        "limits": "View/registry-message fixture, not concurrent disk registry scanning.",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./...",
          "at": "2026-11-21T21:47:27.228210411Z",
          "scope": "TestSubWorkspaceNewSelectionSurvivesRegistryRefresh",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        },
        "line": 131,
        "endLine": 178
      },
      {
        "id": "save-simple-source-surfaces-sync-failure",
        "name": "TestSaveSimpleSourceSurfacesSyncFailure",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/source_sync_warning_test.go",
        "proves": "Prior assessment: Checks emitted sync error, refresh request and durable saved source path despite failed scanning.",
        "limits": "One missing skill-source path through callbacks; no full source-manager UI journey. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.725791758Z",
          "scope": "TestSaveSimpleSourceSurfacesSyncFailure in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 13,
        "endLine": 28
      },
      {
        "id": "workspace-source-i-d-assigned-and-preserved-on-edit",
        "name": "TestWorkspaceSourceIDAssignedAndPreservedOnEdit",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/source_save_test.go",
        "proves": "Prior assessment: Executes source save/edit callbacks, then checks a nonempty repository-assigned ID survives a change of IsRoot.",
        "limits": "Does not drive Ctrl+E, the source form or navigation back to an unsaved card. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.719932808Z",
          "scope": "TestWorkspaceSourceIDAssignedAndPreservedOnEdit in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 67,
        "endLine": 98
      },
      {
        "id": "save-workspace-source-aborts-on-malformed-registry",
        "name": "TestSourceMutationAbortsOnMalformedRegistry",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/source_save_test.go",
        "reviewedHash": "c48d3ee8f0fe52cdf7e8d02622215662b3eb2a1af2f3218ceb6ddc2b244e5955",
        "proves": "The workspace create case uses the source manager callback and checks a failed result and unchanged malformed YAML.",
        "limits": "Coordinator callback and repository fixture. It does not assert the displayed error.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.713677419Z",
          "scope": "TestSourceMutationAbortsOnMalformedRegistry in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 15,
        "endLine": 65
      },
      {
        "id": "delete-workspace-source-aborts-on-malformed-registry",
        "name": "TestSourceMutationAbortsOnMalformedRegistry",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/source_save_test.go",
        "reviewedHash": "c48d3ee8f0fe52cdf7e8d02622215662b3eb2a1af2f3218ceb6ddc2b244e5955",
        "proves": "The workspace delete case uses a valid source identity and checks a failed result and unchanged malformed YAML.",
        "limits": "Coordinator callback and repository fixture. It does not assert the displayed error.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.713677419Z",
          "scope": "TestSourceMutationAbortsOnMalformedRegistry in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 15,
        "endLine": 65
      },
      {
        "id": "source-list-key-crud",
        "name": "TestSourceListIntegration_CRUD",
        "method": "automated",
        "integration": "project",
        "kind": "integration",
        "file": "internal/ui/components/source_list_integration_test.go",
        "proves": "Prior assessment: Drives new/edit/save/delete keys, checks store/display equality, form/list transitions, callback counts, and automatic refresh/reload after successful callbacks.",
        "limits": "In-memory source store; does not establish filesystem persistence. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:57.750951484Z",
          "scope": "TestSourceListIntegration_CRUD in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 15,
        "endLine": 78
      },
      {
        "id": "source-list-delete-error",
        "name": "TestSourceList_HandleDelete_CallbackError",
        "method": "automated",
        "integration": "project",
        "kind": "unit",
        "file": "internal/ui/components/source_list_test.go",
        "proves": "Drives delete on index 1; checks callback invocation, nil returned command, unchanged cursor/items, and zero refresh/reload calls on synchronous callback failure.",
        "limits": "Component error contract; the callback owns error reporting.",
        "reviewedHash": "9f7c6ae7caf62037b2131c47e7e669a9b1d520fe3f68193add2229f5bee5cf8f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:57.751438435Z",
          "scope": "TestSourceList_HandleDelete_CallbackError in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 285,
        "endLine": 307
      },
      {
        "id": "source-manager-delete-error",
        "name": "TestSourceManager_Delete",
        "method": "automated",
        "integration": "project",
        "kind": "integration",
        "file": "internal/ui/views/source_manager_test.go",
        "proves": "Prior assessment: Drives deletion through SourceManager.Update, checks selected index, returned asynchronous ErrorMsg and preserved items after reload.",
        "limits": "Callback fixture emits the asynchronous error; repository failure coverage is separate. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:10.934234675Z",
          "scope": "TestSourceManager_Delete in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 117,
        "endLine": 148
      },
      {
        "id": "source-refresh-destination",
        "name": "TestSourceManagementFlow",
        "method": "automated",
        "integration": "project",
        "kind": "integration",
        "file": "internal/ui/integration_test.go",
        "proves": "Prior assessment: Navigates to source management, changes persisted sources, delivers refresh results and checks actual list additions/removals.",
        "limits": "Coordinator/repository integration; does not cover the unsaved editor Ctrl+E return journey. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.935623493Z",
          "scope": "TestSourceManagementFlow",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        },
        "line": 214,
        "endLine": 254,
        "reviewedHash": "b490df4ae553b790329f3356c2a722efae0de8f9e8c9dd01b0b0163d12b0b91e"
      },
      {
        "id": "card-refresh-destination",
        "name": "TestCoordinatorDataPropagation",
        "method": "automated",
        "integration": "project",
        "kind": "integration",
        "file": "internal/ui/integration_test.go",
        "proves": "Checks an initially empty dashboard then verifies refreshed cards and changed names in the actual destination view.",
        "limits": "Coordinator/repository integration, not every capability registry.",
        "reviewedHash": "b490df4ae553b790329f3356c2a722efae0de8f9e8c9dd01b0b0163d12b0b91e",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.98485336Z",
          "scope": "TestCoordinatorDataPropagation",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        },
        "line": 500,
        "endLine": 537
      },
      {
        "id": "isolated-registry-corruption",
        "name": "TestValidateCardRegistryIntegrity",
        "method": "automated",
        "integration": "project",
        "kind": "unit",
        "file": "internal/config/refactor_test.go",
        "proves": "Each skills/personas/MCP corruption case starts from a passing repository and requires the failing registry filename.",
        "limits": "Malformed YAML paths; not arbitrary registry semantic corruption.",
        "line": 208,
        "endLine": 225,
        "reviewedHash": "9b01386e6b4ab1b5aa57181b11235d30eb05c52cfc559c5a4c7465b87e767786",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.235127619Z",
          "scope": "TestValidateCardRegistryIntegrity in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "blank-canvas-representation",
        "name": "TestCanvasCreation",
        "method": "automated",
        "integration": "project",
        "kind": "unit",
        "file": "internal/ascii/canvas_test.go",
        "proves": "Checks the exact five blank rows of ten cells consumed by the renderer.",
        "limits": "ASCII constructor representation only.",
        "line": 10,
        "endLine": 16,
        "reviewedHash": "677d53291a799c08ed87d7b4b950f840f3254ccbbe4df7702c7f11587deb17f4",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.022245516Z",
          "scope": "TestCanvasCreation in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "settings-view_-save-in-edit-mode",
        "name": "TestSettingsView_SaveInEditMode",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/settings_test.go",
        "reviewedHash": "a34f2ebbb13bcf7181347b7cf1d0e729344abf8911547b35ab6c847fefd687ad",
        "proves": "Checks Enter calls OnSave with the edited IDE command and returns the view to display mode.",
        "limits": "In-memory callback payload only; does not execute the asynchronous repository save.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:10.928570945Z",
          "scope": "TestSettingsView_SaveInEditMode in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 57,
        "endLine": 79
      },
      {
        "id": "settings-view_-terminal-command-focus-save-and-cancel",
        "name": "TestSettingsView_TerminalCommandFocusSaveAndCancel",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/settings_test.go",
        "reviewedHash": "a34f2ebbb13bcf7181347b7cf1d0e729344abf8911547b35ab6c847fefd687ad",
        "proves": "Checks Tab/Shift+Tab preserve separate command text, Enter trims/saves both fields, and Esc restores both saved inputs after further edits.",
        "limits": "View model callback boundary only; no YAML write.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:10.929459195Z",
          "scope": "TestSettingsView_TerminalCommandFocusSaveAndCancel in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 137,
        "endLine": 164
      },
      {
        "id": "settings-view_-toggle-rows",
        "name": "TestSettingsView_ToggleRows",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/settings_test.go",
        "reviewedHash": "a34f2ebbb13bcf7181347b7cf1d0e729344abf8911547b35ab6c847fefd687ad",
        "proves": "Tab focuses the two visibility toggles, Space changes them, and settings values reflect the changes.",
        "limits": "Does not execute coordinator persistence or check every focus wrap boundary.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:10.928615075Z",
          "scope": "TestSettingsView_ToggleRows in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 81,
        "endLine": 102
      },
      {
        "id": "app-settings-persistence",
        "name": "TestAppSettingsPersistence",
        "method": "automated",
        "integration": "project",
        "file": "internal/config/app_settings_test.go",
        "line": 13,
        "endLine": 48,
        "reviewedHash": "c592744c5bf639742840eff42c89f67ca9d7280245e9bfdfa11301f3914c272d",
        "proves": "Checks default command/visibility settings, saves independent IDE/terminal commands and visibility options, then reads back exact AppSettings equality.",
        "limits": "Repository Save/Get against the same repository instance; no real app restart or coordinated navigation assertion.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.064045382Z",
          "scope": "TestAppSettingsPersistence in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "terminal-launch-through-dashboard",
        "name": "TestTerminalLaunchThroughDashboard",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/terminal_launch_test.go",
        "proves": "Prior assessment: Drives Tab/t through coordinator and executes the command; child record must contain the actual temporary WorkingDir in cwd/args and differ from OriginalDir.",
        "limits": "Real local helper process rather than a graphical terminal emulator; one configured launch shape. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.759954665Z",
          "scope": "TestTerminalLaunchThroughDashboard in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 43,
        "endLine": 128
      },
      {
        "id": "settings-save-failure-visible",
        "name": "TestSettingsSaveFailureRemainsVisible",
        "method": "automated",
        "integration": "project",
        "kind": "integration",
        "file": "internal/ui/settings_error_test.go",
        "proves": "Edits settings through coordinator Update, executes the real save callback command, delivers SettingsSavedMsg, and checks the visible error, retained view/edit and unchanged persisted settings.",
        "limits": "Repository wrapper injects save failure; successful save/editor-refresh composition remains separately partial.",
        "line": 27,
        "endLine": 73,
        "reviewedHash": "81bb1da189891e07b7e291451c878bfd68d116805f8f04aa6691a88193a7586c",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.831886065Z",
          "scope": "TestSettingsSaveFailureRemainsVisible in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "settings-view_-long-commands-stay-within-size",
        "name": "TestSettingsView_LongCommandsStayWithinSize",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/settings_test.go",
        "reviewedHash": "a34f2ebbb13bcf7181347b7cf1d0e729344abf8911547b35ab6c847fefd687ad",
        "proves": "For display/edit mode at 80×20 and 120×36, checks rows/display widths, no injected ANSI/newline text, retained controls and edit cancel help.",
        "limits": "String/cell assertions rather than terminal screenshots; helper assertSettingsFits is at settings_test.go:181.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:10.933549744Z",
          "scope": "TestSettingsView_LongCommandsStayWithinSize in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 166,
        "endLine": 176
      },
      {
        "id": "m-c-p-form-round-trips-every-field",
        "name": "TestMCPFormRoundTripsEveryField",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/mcp_form_test.go",
        "proves": "Prior assessment: Presses Enter and asserts one save with exact ID/name/command/args/CWD/URL/env/headers, including an argument containing a comma.",
        "limits": "Callback payload only; does not connect to that server or persist the configuration. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.643641114Z",
          "scope": "TestMCPFormRoundTripsEveryField in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 148,
        "endLine": 199
      },
      {
        "id": "dashboard_-launch-terminal-routing",
        "name": "TestDashboard_LaunchTerminalRouting",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_test.go",
        "proves": "Prior assessment: Checks the terminal callback receives the selected session for running/waiting/terminated states, ignores no-selection/search/nil callback cases, and checks o selects the expected IDE session.",
        "limits": "View callback assertions only; no command execution or external UI. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.629044715Z",
          "scope": "TestDashboard_LaunchTerminalRouting in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 492,
        "endLine": 540
      },
      {
        "id": "launch-workspace-command-executes-in-workspace",
        "name": "TestLaunchWorkspaceCommandExecutesInWorkspace",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/ide_launcher_test.go",
        "line": 69,
        "endLine": 78,
        "reviewedHash": "77686722e24693b846bff6f8047659d01ba9757df42d8e0fed047c940f32849f",
        "proves": "The helper runWorkspaceLauncherCase starts a real child and asserts absolute cwd, exact placeholder multiplicity, quoted paths with spaces and preservation of ordinary arguments.",
        "limits": "Assertions reside in runWorkspaceLauncherCase at internal/session/ide_launcher_test.go:80, called by this table-driven test. Uses the test executable, not an IDE GUI.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:05.450434799Z",
          "scope": "TestLaunchWorkspaceCommandExecutesInWorkspace in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "terminal-launch-reports-configuration-and-lookup-errors",
        "name": "TestTerminalLaunchReportsConfigurationAndLookupErrors",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/terminal_launch_test.go",
        "proves": "Prior assessment: Checks missing/foreign scope, missing IDE/terminal commands, failed executable and absent workspace; feeds spawn failure back and asserts the coordinator banner is set.",
        "limits": "Most cases invoke callbacks directly. The banner assertion establishes error presence, not its full displayed text. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.777924174Z",
          "scope": "TestTerminalLaunchReportsConfigurationAndLookupErrors in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 164,
        "endLine": 258
      },
      {
        "id": "launch-workspace-command-errors",
        "name": "TestLaunchWorkspaceCommandErrors",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/ide_launcher_test.go",
        "line": 155,
        "endLine": 171,
        "reviewedHash": "77686722e24693b846bff6f8047659d01ba9757df42d8e0fed047c940f32849f",
        "proves": "Checks unconfigured/whitespace command, empty/missing workspace path and missing executable return the expected sentinels or an error.",
        "limits": "Shared launcher API tests, not every caller-specific banner or concurrent filesystem change.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:05.453254779Z",
          "scope": "TestLaunchWorkspaceCommandErrors in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "conversation-choice-requires-explicit-compatible-choice",
        "name": "TestConversationChoiceRequiresExplicitCompatibleChoice",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/conversation_choice_test.go",
        "proves": "Prior assessment: Checks a matching Claude binding is submitted and a selected Codex binding cannot be submitted to the Claude activation.",
        "limits": "Picker callback boundary; native CLI resume is outside scope. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:58.338356869Z",
          "scope": "TestConversationChoiceRequiresExplicitCompatibleChoice in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 16,
        "endLine": 46
      },
      {
        "id": "activation-published-before-startup-delay",
        "name": "TestActivationPublishedBeforeStartupDelay",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/startup_integration_test.go",
        "line": 56,
        "endLine": 108,
        "reviewedHash": "1847caf3b093d5208ab0c5fd95a41f44dededa6ffbab7daa7730fc1e37e35a25",
        "proves": "Checks a published real cat activation reports starting/process-started and permits immediate attachment; controlled timer transitions to active without enabling MCP-disabled delivery.",
        "limits": "Injected startup clock/driver preparation; does not assert native CLI task readiness.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:10.412072375Z",
          "scope": "TestActivationPublishedBeforeStartupDelay in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "startup-admission-and-registration-are-independent",
        "name": "TestStartupAdmissionAndRegistrationAreIndependent",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/startup_boundary_test.go",
        "line": 73,
        "endLine": 113,
        "reviewedHash": "d45a7c426f6cc5146d599ef0de532250f821904149f365d147b78f5dc2666c66",
        "proves": "Registers a conversation during starting, checks early delivery rejection, then advances the controlled startup timer and verifies admission becomes available.",
        "limits": "Controlled clock/manager fixture, not native CLI processing.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:04.213164295Z",
          "scope": "TestStartupAdmissionAndRegistrationAreIndependent in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "inspection-conversion-preserves-activation-snapshot",
        "name": "TestInspectionConversionPreservesActivationSnapshot",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/inspection_runtime_test.go",
        "line": 11,
        "endLine": 39,
        "reviewedHash": "6df6907a9f82ded9c9630b78044cc58fe22938ac554baadb19ec322a3efaa60a",
        "proves": "Checks inspection copies activation/startup/delivery fields, serialized historical effects stay unchanged after source snapshot mutation and old data does not fabricate runtime fields.",
        "limits": "Pure conversion/serialization test, separate from transport.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.285852268Z",
          "scope": "TestInspectionConversionPreservesActivationSnapshot in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "dashboard-launch-persisted",
        "name": "TestFullSessionLaunchFlow",
        "method": "automated",
        "integration": "project",
        "kind": "integration",
        "file": "internal/ui/integration_test.go",
        "proves": "Prior assessment: Drives dashboard Enter, executes launch commands, verifies a running process and reads the persisted session/card/dashboard/workspace identity.",
        "limits": "Real local cat process; no native-agent task execution or raw-terminal attachment. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.184638744Z",
          "scope": "TestFullSessionLaunchFlow in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 23,
        "endLine": 138
      },
      {
        "id": "activation-rejects-unavailable-saved-workspace",
        "name": "TestActivationRejectsUnavailableSavedWorkspace",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/workspace_validation_test.go",
        "line": 27,
        "endLine": 116,
        "proves": "For relaunch/switch and original/temporary paths, checks missing/file errors, no extra Spawn, preserved session identity/content, and successful single-spawn retry after restoring the directory.",
        "limits": "Controlled driver counts and filesystem fixtures; no real Codex/Claude execution.",
        "reviewedHash": "e484a44ee3d38d1551068285e2471f861dd497b11af4cc4503256e6e81574550",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:21.578513021Z",
          "scope": "TestActivationRejectsUnavailableSavedWorkspace in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "read-sub-workspace-descriptors-does-not-prepare-workspace",
        "name": "TestReadSubWorkspaceDescriptorsDoesNotPrepareWorkspace",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/sub_workspaces_inspection_test.go",
        "line": 11,
        "endLine": 40,
        "proves": "Checks descriptor identity/path, absence of a newly created tmp directory and unavailable status for corrupt manifests.",
        "limits": "Read-only helper only; copy collision and multi-selection journeys remain separate.",
        "reviewedHash": "47b304b52e9a80c1587c84b7b469e5cab9c4696c3769199011887bbdeda0c3a7",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:10.437627163Z",
          "scope": "TestReadSubWorkspaceDescriptorsDoesNotPrepareWorkspace in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "relaunch-refreshes-only-kankuro-m-c-p-after-restart",
        "name": "TestRelaunchRefreshesOnlyKankuroMCPAfterRestart",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/relaunch_mcp_test.go",
        "line": 16,
        "endLine": 91,
        "proves": "Checks invalid config blocks relaunch without rewriting, then preserves instructions/custom server/metadata, rotates token, revokes the old credential and uses the new endpoint.",
        "limits": "Local files and cat process across in-process restart; no native CLI MCP handshake.",
        "reviewedHash": "d6d0f89ec7ac16c4c2395b51aeae0d31ba9bcab099c7395fd6806a03cc127526",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.212621762Z",
          "scope": "TestRelaunchRefreshesOnlyKankuroMCPAfterRestart in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "trust-invalid-configuration-preserved",
        "name": "TestTrustInvalidConfigurationPreserved",
        "method": "automated",
        "integration": "project",
        "file": "internal/driver/trust_test.go",
        "line": 86,
        "endLine": 113,
        "proves": "Checks PrepareWorkspace errors and byte-identical original configuration for Codex TOML, Claude JSON and a wrong-typed projects field.",
        "limits": "Isolated HOME; no installed native CLI trust acceptance.",
        "reviewedHash": "c3c5f502e8c165b4b78aeea83d59d1301a13bd91b3dd1233565c1709f8f9c3f2",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.603428821Z",
          "scope": "TestTrustInvalidConfigurationPreserved in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "resolve-capabilities-adds-m-c-p-access-skill-only-when-kankuro-m-c-p-selected",
        "name": "TestResolveCapabilitiesAddsMCPAccessSkillOnlyWhenKankuroMCPSelected",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 1243,
        "endLine": 1272,
        "reviewedHash": "4c2f3e9342ceea29d5ff1d5ba392ce0d5c67b3e8b03c38bff8421942ca9bc9df",
        "proves": "Checks mcp-cli is absent without Kankuro MCP, both protocol skills appear once with Kankuro selection and explicit mcp-cli selection does not duplicate either.",
        "limits": "Tests capability selection, not instruction following by an AI agent.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.753576964Z",
          "scope": "TestResolveCapabilitiesAddsMCPAccessSkillOnlyWhenKankuroMCPSelected",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "embedded-m-c-p-c-l-i-skill-deploys-through-config",
        "name": "TestEmbeddedMCPCLISkillDeploysThroughConfig",
        "method": "automated",
        "integration": "project",
        "file": "main_test.go",
        "line": 13,
        "endLine": 40,
        "reviewedHash": "a7d91caf10f7521dc017c9a9317d878331cf2ebe5dfdc853fea56bbe5ec70b5a",
        "proves": "Checks embedded mcp-cli bytes/metadata and exact deployed SKILL.md bytes through the real driver deployment.",
        "limits": "Uses builtin lookup only as fixture setup for the live deployment path; does not establish any runtime need for that lookup helper. The recorded execution is historical.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.018148696Z",
          "scope": "TestEmbeddedMCPCLISkillDeploysThroughConfig in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "embedded-kankuro-puppet-skill-deploys-through-config",
        "name": "TestEmbeddedKankuroPuppetSkillDeploysThroughConfig",
        "method": "automated",
        "integration": "project",
        "file": "main_test.go",
        "line": 42,
        "endLine": 60,
        "reviewedHash": "a7d91caf10f7521dc017c9a9317d878331cf2ebe5dfdc853fea56bbe5ec70b5a",
        "proves": "Checks canonical embedded puppet skill metadata and byte-identical driver deployment to the expected skill folder.",
        "limits": "File deployment only; no evidence that an installed CLI reads or follows the skill. The recorded execution is historical.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.018649746Z",
          "scope": "TestEmbeddedKankuroPuppetSkillDeploysThroughConfig in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "session-attach-replay",
        "name": "TestSessionAttachReplay",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/attach_test.go",
        "line": 11,
        "endLine": 50,
        "reviewedHash": "50a58baf339a4b8439246f3d2e98a6ec453af1063d2a190efd536e99adb6a4f4",
        "proves": "Writes hello through a real cat attachment, closes it, reattaches and asserts exact replayed scrollback.",
        "limits": "Manager handles only; does not run attachExec raw terminal mode, Ctrl+K or resize synchronization.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:04.473538515Z",
          "scope": "TestSessionAttachReplay in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "service-remove-session-delegates-guards-cleanup-and-manager-errors",
        "name": "TestServiceRemoveSessionDelegatesGuardsCleanupAndManagerErrors",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 1725,
        "endLine": 1785,
        "reviewedHash": "4c2f3e9342ceea29d5ff1d5ba392ce0d5c67b3e8b03c38bff8421942ca9bc9df",
        "proves": "Rejects removal of a running session, removes a stopped temporary workspace, propagates session-record cleanup failure and closed-manager activation error.",
        "limits": "Service/manager and real filesystem fixture; not the dashboard confirmation key sequence.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.916760564Z",
          "scope": "TestServiceRemoveSessionDelegatesGuardsCleanupAndManagerErrors",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "session-launcher-waits-for-name-result-and-allows-retry",
        "name": "TestSessionLauncherWaitsForNameResultAndAllowsRetry",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/session_launcher_test.go",
        "proves": "Prior assessment: Checks name submission starts once, duplicate Enter is suppressed while pending, an error is shown and another Enter can retry.",
        "limits": "In-memory naming callback/result boundary; not a persisted service rename. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.144731611Z",
          "scope": "TestSessionLauncherWaitsForNameResultAndAllowsRetry",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        },
        "line": 136,
        "endLine": 161,
        "reviewedHash": "d2a35f3d77fc4e93d3a29699a8d0fc9aa05f053d9649c768855ac0603b9c8c4b"
      },
      {
        "id": "set-name-propagates-persist-failure",
        "name": "TestSetNamePropagatesPersistFailure",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/lifecycle_test.go",
        "line": 413,
        "endLine": 451,
        "reviewedHash": "8c7059d4e1eccf3c2b67008f3a76171f4d7022ce691d3ddee93a06c0eedc23d8",
        "proves": "Makes the session-record directory unwritable and asserts the checked name operation propagates its persistence error.",
        "limits": "Permission-based test skips as root; manager fixture is not the complete UI rename path.",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./...",
          "at": "2026-11-21T21:47:25.293694226Z",
          "scope": "TestSetNamePropagatesPersistFailure",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "rename-rename-restart",
        "name": "TestCoordinator_DashboardRenameRefreshesWithoutAttachAndSurvivesRestart",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/coordinator_test.go",
        "proves": "Prior assessment: Real cat session renamed through coordinator input, service and disk; returned command only refreshes; selected dashboard name updates, view stays put, process PID/activation identity are retained; fresh repository/manager/coordinator reloads the saved name.",
        "limits": "Ordinary Go integration test using a local cat PTY; does not drive a native AI CLI or real terminal renderer. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:06.826651502Z",
          "scope": "TestCoordinator_DashboardRenameRefreshesWithoutAttachAndSurvivesRestart in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 127,
        "endLine": 189
      },
      {
        "id": "rename-cancel-draft",
        "name": "TestDashboard_RenameSessionDiscardsDraftWithoutSaving",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_test.go",
        "proves": "Esc, dashboard changes and generation changes close an edited draft without saving or changing the session name. Reopening starts from the original name and submits once.",
        "limits": "View-level callback spy; no native terminal key decoding or persistence is involved.",
        "reviewedHash": "fe042e81c3207ef759933378eca1b1ec1ba5f30f1aa3a1079c017b33289025bd",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.623982236Z",
          "scope": "TestDashboard_RenameSessionDiscardsDraftWithoutSaving in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 437,
        "endLine": 441
      },
      {
        "id": "messaging-operations-pin-activations-and-persist-explicit-reports",
        "name": "TestMessagingOperationsPinActivationsAndPersistExplicitReports",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 1835,
        "endLine": 1952,
        "reviewedHash": "4c2f3e9342ceea29d5ff1d5ba392ce0d5c67b3e8b03c38bff8421942ca9bc9df",
        "proves": "Checks target activation inspection, submitted send result, stable replay identity, changed-text conflict, stored text, worker report permission and removal of history on recipient deletion.",
        "limits": "Local service/PTY fixtures; submission is byte evidence, not an agent acknowledgment or processing proof.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.958065002Z",
          "scope": "TestMessagingOperationsPinActivationsAndPersistExplicitReports",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "final-journal-write-failure-returns-outcome-and-recovers-interrupted",
        "name": "TestFinalJournalWriteFailureReturnsOutcomeAndRecoversInterrupted",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 507,
        "endLine": 584,
        "reviewedHash": "4c2f3e9342ceea29d5ff1d5ba392ce0d5c67b3e8b03c38bff8421942ca9bc9df",
        "proves": "Injects failure after a successful creation effect, checks usable target/result with warning, in-memory replay without another session, and interrupted recovery without auto-execution.",
        "limits": "Injected journal write seam and in-process runner restart, not OS crash/power-loss.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.464953432Z",
          "scope": "TestFinalJournalWriteFailureReturnsOutcomeAndRecoversInterrupted",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "operation-queue-bound-and-disconnected-wait-return-promptly",
        "name": "TestOperationQueueBoundAndDisconnectedWaitReturnPromptly",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 1002,
        "endLine": 1170,
        "reviewedHash": "4c2f3e9342ceea29d5ff1d5ba392ce0d5c67b3e8b03c38bff8421942ca9bc9df",
        "proves": "Fills worker/queue gates, checks prompt return on disconnected waits, capacity_exhausted, eventual completed effects and truthful success-with-warning when cancellation arrives after publication.",
        "limits": "Finite controlled worker interleavings and deadlines; not a throughput or starvation proof.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.748273014Z",
          "scope": "TestOperationQueueBoundAndDisconnectedWaitReturnPromptly",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "operation-journal-recovers-interrupted-and-fails-closed-on-corruption",
        "name": "TestOperationJournalRecoversInterruptedAndFailsClosedOnCorruption",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 1172,
        "endLine": 1203,
        "reviewedHash": "4c2f3e9342ceea29d5ff1d5ba392ce0d5c67b3e8b03c38bff8421942ca9bc9df",
        "proves": "Writes a running record, checks interrupted recovery with warning and verifies a corrupt journal sets loadErr to block mutation admission.",
        "limits": "Constructed journal records and runner state; no actual process crash.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.748717374Z",
          "scope": "TestOperationJournalRecoversInterruptedAndFailsClosedOnCorruption",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "shutdown-during-spawn-kills-the-new-process",
        "name": "TestShutdownDuringSpawnKillsTheNewProcess",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/shutdown_test.go",
        "line": 161,
        "endLine": 192,
        "proves": "Checks shutdown blocks for an in-flight relaunch, returns ErrManagerClosed after release, kills the spawned driver and leaves terminated status.",
        "limits": "Gated test driver; does not simulate every OS process-group behavior.",
        "reviewedHash": "c3feacc7073ff6582b64bf6830985cb9c7a26005624752028f9c7dd862320d42",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.876460846Z",
          "scope": "TestShutdownDuringSpawnKillsTheNewProcess in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "shutdown-waits-for-temporary-workspace-preparation",
        "name": "TestShutdownWaitsForTemporaryWorkspacePreparation",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/shutdown_test.go",
        "line": 235,
        "endLine": 275,
        "proves": "Blocks the copier, asserts Shutdown waits, then checks closed-manager launch error and an empty temporary root.",
        "limits": "Injected copier; does not prove native cp cancellation under every filesystem.",
        "reviewedHash": "c3feacc7073ff6582b64bf6830985cb9c7a26005624752028f9c7dd862320d42",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:10.09260827Z",
          "scope": "TestShutdownWaitsForTemporaryWorkspacePreparation in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "data-round-trip-copies-every-field",
        "name": "TestDataRoundTripCopiesEveryField",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/mapping_internal_test.go",
        "line": 12,
        "endLine": 30,
        "proves": "Fills fields with nonzero values and compares all persisted values through data/restore and snapshot, excluding intentional schema/status transformations.",
        "limits": "Pure mapping; no disk durability assertion or complete clone-alias isolation claim.",
        "reviewedHash": "7a24af59808ff28547db62ea979f2028b08a805bd1ae443bbd7bef97a9f77831",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.35291536Z",
          "scope": "TestDataRoundTripCopiesEveryField in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "ordinary-message-journal-rejection-rolls-back-history-and-retries",
        "name": "TestOrdinaryMessageJournalRejectionRollsBackHistoryAndRetries",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_shared_delivery_test.go",
        "line": 78,
        "endLine": 119,
        "reviewedHash": "54f614934f1d4b4db1ce19564d7c19ffc9f3d917ada59b39e62db979117dad09",
        "proves": "Injects journal admission failure, checks no intent/receipt survives, then retries and replays the same request with the original message identity.",
        "limits": "Controlled write failure; no later partial PTY write claim.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:59.022584256Z",
          "scope": "TestOrdinaryMessageJournalRejectionRollsBackHistoryAndRetries in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "ordinary-message-queue-rejection-releases-reservation-and-preserves-replay",
        "name": "TestOrdinaryMessageQueueRejectionReleasesReservationAndPreservesReplay",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_shared_delivery_test.go",
        "line": 34,
        "endLine": 76,
        "reviewedHash": "54f614934f1d4b4db1ce19564d7c19ffc9f3d917ada59b39e62db979117dad09",
        "proves": "Fills operation queue, checks capacity rejection leaves no message, then retries/replays same message ID and stops recipient without a retained reservation.",
        "limits": "Controlled queue/worker fixture rather than unbounded production load.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:58.97638779Z",
          "scope": "TestOrdinaryMessageQueueRejectionReleasesReservationAndPreservesReplay in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "report-notification-initial-state-is-durable-and-idempotent",
        "name": "TestReportNotificationInitialStateIsDurableAndIdempotent",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/messaging_test.go",
        "line": 46,
        "endLine": 90,
        "proves": "Checks pending for completed/blocked/question and not_required for progress; reopen/retry returns the same report ID and one stored item with the expected state.",
        "limits": "Message store on disk; no prompt delivery or agent processing.",
        "reviewedHash": "e853b5124ff8cf803fb6075b18985c8739e9ef4a7d766771631879e2aaff8007",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.146605327Z",
          "scope": "TestReportNotificationInitialStateIsDurableAndIdempotent in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "progress-replies-persist-before-completion",
        "name": "TestProgressRepliesPersistBeforeCompletion",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/messaging_test.go",
        "line": 218,
        "endLine": 258,
        "proves": "Checks progress deduplication, rejection of a changed kind under the same request ID, and ordered progress/completed reports after reopening.",
        "limits": "Does not observe whether a live CLI continues working after progress.",
        "reviewedHash": "e853b5124ff8cf803fb6075b18985c8739e9ef4a7d766771631879e2aaff8007",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.149028367Z",
          "scope": "TestProgressRepliesPersistBeforeCompletion in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "restart-marks-admitted-notification-uncertain",
        "name": "TestRestartMarksAdmittedNotificationUncertain",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/messaging_test.go",
        "line": 260,
        "endLine": 324,
        "proves": "Checks recovery marks admitted delivery uncertain while preserving the accepted obligation and request identity.",
        "limits": "Constructed durable records; no process crash at an actual PTY write.",
        "reviewedHash": "e853b5124ff8cf803fb6075b18985c8739e9ef4a7d766771631879e2aaff8007",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.149838987Z",
          "scope": "TestRestartMarksAdmittedNotificationUncertain in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "report-automatically-notifies-original-controller",
        "name": "TestReportAutomaticallyNotifiesOriginalController",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 481,
        "endLine": 520,
        "proves": "Denies the worker a general send, observes controller PTY output containing message/reply IDs and no-report-reply guidance, and checks sensitive report content is absent.",
        "limits": "Real local PTY harness; receipt of bytes is not proof an AI agent processed or fetched the report.",
        "reviewedHash": "2b7b06d4b5a4145f38f48757741fc9829481eb3fc735a0f393d676016724df42",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:57.328946127Z",
          "scope": "TestReportAutomaticallyNotifiesOriginalController in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "saved-notification-survives-reporter-stopping",
        "name": "TestSavedNotificationSurvivesReporterStopping",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 656,
        "endLine": 686,
        "proves": "Stops dispatcher, saves report, kills reporter, restarts runner and waits for notification output.",
        "limits": "Same application process and local harness; not an application-wide restart.",
        "reviewedHash": "2b7b06d4b5a4145f38f48757741fc9829481eb3fc735a0f393d676016724df42",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:57.664044621Z",
          "scope": "TestSavedNotificationSurvivesReporterStopping in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "stopped-original-activation-makes-notification-unavailable",
        "name": "TestStoppedOriginalActivationMakesNotificationUnavailable",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 688,
        "endLine": 710,
        "proves": "Stops/relaunches the original sender, attempts dispatch and checks unavailable notification state.",
        "limits": "State assertion for one activation replacement; not arbitrary authorization races.",
        "reviewedHash": "2b7b06d4b5a4145f38f48757741fc9829481eb3fc735a0f393d676016724df42",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:57.711248038Z",
          "scope": "TestStoppedOriginalActivationMakesNotificationUnavailable in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "notification-retries-after-queue-capacity-returns",
        "name": "TestNotificationRetriesAfterQueueCapacityReturns",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 522,
        "endLine": 577,
        "proves": "Fills queue, checks pending obligation and request ID survive rejection, drains queue, starts workers and observes notification with the same request ID.",
        "limits": "Controlled runner/queue; no native agent readiness proof.",
        "reviewedHash": "2b7b06d4b5a4145f38f48757741fc9829481eb3fc735a0f393d676016724df42",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:57.388070533Z",
          "scope": "TestNotificationRetriesAfterQueueCapacityReturns in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "notification-reconciliation-uses-delivery-evidence",
        "name": "TestNotificationReconciliationUsesDeliveryEvidence",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 580,
        "endLine": 654,
        "proves": "Checks state mapping, advances the attempt only for not-submitted delivery and verifies stale reconciliation does not downgrade or advance it again.",
        "limits": "Constructed persisted delivery states; concrete driver byte-write boundaries need their own tests.",
        "reviewedHash": "2b7b06d4b5a4145f38f48757741fc9829481eb3fc735a0f393d676016724df42",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:57.617779535Z",
          "scope": "TestNotificationReconciliationUsesDeliveryEvidence in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "concurrent-notification-dispatch-admits-one-attempt",
        "name": "TestConcurrentNotificationDispatchAdmitsOneAttempt",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 832,
        "endLine": 861,
        "proves": "Runs twelve dispatchers against the same obligation and asserts one matching runner operation.",
        "limits": "One controlled interleaving; this run did not use the race detector.",
        "reviewedHash": "2b7b06d4b5a4145f38f48757741fc9829481eb3fc735a0f393d676016724df42",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:58.869197168Z",
          "scope": "TestConcurrentNotificationDispatchAdmitsOneAttempt in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "notification-scheduling-preserves-recipient-order-and-fairness",
        "name": "TestNotificationSchedulingPreservesRecipientOrderAndFairness",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 929,
        "endLine": 954,
        "proves": "Checks one recipient backs off without suppressing another and keeps its accepted predecessor ahead of a later pending report.",
        "limits": "Pure scheduling function with fixed time; no throughput benchmark.",
        "reviewedHash": "2b7b06d4b5a4145f38f48757741fc9829481eb3fc735a0f393d676016724df42",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:58.927311823Z",
          "scope": "TestNotificationSchedulingPreservesRecipientOrderAndFairness in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "queued-delivery-cancellation-releases-reservation-without-submitting",
        "name": "TestQueuedDeliveryCancellationReleasesReservationWithoutSubmitting",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_shared_delivery_test.go",
        "line": 122,
        "endLine": 230,
        "reviewedHash": "54f614934f1d4b4db1ce19564d7c19ffc9f3d917ada59b39e62db979117dad09",
        "proves": "For ordinary and notification queues, cancels before execution, starts a worker, checks not-submitted delivery, successful target stop and absence of message markers in captured output.",
        "limits": "Controlled runner/worker plus PTY bytes; ordinary uses public cancellation, notification cancels the captured task authority directly.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:59.10229428Z",
          "scope": "TestQueuedDeliveryCancellationReleasesReservationWithoutSubmitting in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "deleting-report-history-cancels-admitted-notification",
        "name": "TestDeletingReportHistoryCancelsAdmittedNotification",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 755,
        "endLine": 820,
        "reviewedHash": "2b7b06d4b5a4145f38f48757741fc9829481eb3fc735a0f393d676016724df42",
        "proves": "Publishes admitted notification intent, deletes dashboard history, then checks submission rejects and notifications stay absent rather than recreating deleted history.",
        "limits": "Direct manager/store admission sequence; no raw agent prompt-processing claim.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:58.816056242Z",
          "scope": "TestDeletingReportHistoryCancelsAdmittedNotification in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "queued-message-reservations-cannot-deadlock-stop-workers-or-shutdown",
        "name": "TestQueuedMessageReservationsCannotDeadlockStopWorkersOrShutdown",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 179,
        "endLine": 288,
        "reviewedHash": "2b7b06d4b5a4145f38f48757741fc9829481eb3fc735a0f393d676016724df42",
        "proves": "Queues delivery reservations alongside stop operations, cancels sends, waits for stop and checks runner Shutdown returns before deadline with all targets terminated.",
        "limits": "Finite controlled worker interleaving for ordinary sends; dispatcher shutdown during a real notification write is not directly asserted.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.813462167Z",
          "scope": "TestQueuedMessageReservationsCannotDeadlockStopWorkersOrShutdown in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "native-agent-terminal-journey",
        "name": "TestNativeAgentJourney",
        "method": "automated",
        "integration": "project",
        "kind": "e2e",
        "file": "appedia/historical-sources/native_agent_test.go",
        "proves": "Uses tui-test Go to launch real Kankuro/OpenCode, verify imported credentials, assign an isolated service port and select OpenRouter Free Models Router. Reached assertions: Dashboard Enter launched a persisted native-agent process and Enter attached to the native UI.",
        "limits": "Explicit e2e tag required. Single OpenCode 2.2.6/OpenRouter router run; no screenshot baseline comparison. The full test file and configuration, capture and provenance helpers are retained in the inventory. The run failed at register; subsequent assertions were not reached. Historical native-agent result/captures were retained from 2026-11-09; this refresh did not execute that journey or rebind it to the current source fingerprint. Historical observation retained in Appedia.",
        "result": {
          "status": "failed",
          "command": "go test -tags=e2e ./tests/e2e -run '^TestNativeAgentJourney$' -count=1 -timeout=10m -v",
          "at": "2026-11-09T22:42:26.268507608Z",
          "scope": "TestNativeAgentJourney; openrouter/openrouter/free; final stage register",
          "revision": "WORKTREE",
          "fingerprint": "22ce60949b80ba3d6d306b85a4c88683bf040372efe7dbbff5d879030f24365a"
        },
        "reviewedHash": "41f3ea85716fea5e9811a386f25b55b6c75f2921f7352de6c63eef22f5e091df"
      },
      {
        "id": "m-c-p-c-l-i-registers-conversation-for-live-activation",
        "name": "TestMCPCLIRegistersConversationForLiveActivation",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/conversation_test.go",
        "line": 63,
        "endLine": 333,
        "reviewedHash": "67dbeb8f4c470db473e3a379e88aa76d2a2e187a372bea1e622d224a5a6ac758",
        "proves": "Runs a shell child using the installed mcp-cli against the real MCP endpoint, checks reserved-env isolation, binding persistence/warnings/retry, stale credentials, explicit choice and card/session-local bindings across restart.",
        "limits": "Interoperability test uses a shell/cat agent; it is not an installed Codex/Claude CLI accepting native resume. Skips when mcp-cli is absent; the recorded endpoint run passed.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:05.229418366Z",
          "scope": "TestMCPCLIRegistersConversationForLiveActivation in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "claude-resume-lifecycle-uses-selected-native-i-d",
        "name": "TestClaudeResumeLifecycleUsesSelectedNativeID",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/claude_resume_test.go",
        "line": 70,
        "endLine": 186,
        "reviewedHash": "1bc77dd8da8a6f189300421e5801adc407c37c2dfe40d5ccc7818e3df9d316dc",
        "proves": "Captures fresh/saved/new and A→B→A Claude arguments, checks original workspace marker survives and conversation IDs stay card-local.",
        "limits": "Overrides Spawn and PrepareWorkspace with a capture driver; never starts the installed Claude CLI.",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./...",
          "at": "2026-11-21T21:47:23.507794977Z",
          "scope": "TestClaudeResumeLifecycleUsesSelectedNativeID",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "rename-draft",
        "name": "TestDashboard_RenameSession",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_test.go",
        "proves": "Prior assessment: Captured target, normalized Unicode input, draft survives refresh, duplicate suppression, retry after error.",
        "limits": "View callback is a stub; no service write or cancellation assertion before submit. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.614062867Z",
          "scope": "TestDashboard_RenameSession in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 362,
        "endLine": 409
      },
      {
        "id": "rename-blank",
        "name": "TestDashboard_RenameSessionEmptyUsesIDAndNoSelectionIsNoop",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_test.go",
        "proves": "Prior assessment: Blank input submits the session ID.",
        "limits": "No-selection probe starts in the cards pane; it does not establish empty sessions-pane behavior. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.615748347Z",
          "scope": "TestDashboard_RenameSessionEmptyUsesIDAndNoSelectionIsNoop in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 411,
        "endLine": 435
      },
      {
        "id": "rename-scoped",
        "name": "TestCoordinator_DashboardRenameFailureStaysInDialogAndIgnoresStaleResult",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/coordinator_test.go",
        "proves": "Prior assessment: Real command error stays in dialog; stale generation leaves pending state intact.",
        "limits": "Missing session fixture; does not establish successful persistence or stale dashboard identity separately. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:06.796265085Z",
          "scope": "TestCoordinator_DashboardRenameFailureStaysInDialogAndIgnoresStaleResult in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 59,
        "endLine": 98
      },
      {
        "id": "rename-layout",
        "name": "TestDashboardRenameLayoutAndScopedFooter",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_layout_test.go",
        "proves": "Long error bounds at 80×24 and 120×40, visible save/cancel controls and scoped rename footer.",
        "limits": "String and cell bounds, not a native terminal screenshot or pixel comparison.",
        "reviewedHash": "34242af9c0592577eeee58c84686ed8d3399ce6110722c11838c87111097c43d",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:58.373479516Z",
          "scope": "TestDashboardRenameLayoutAndScopedFooter in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 149,
        "endLine": 174
      },
      {
        "id": "rename-search",
        "name": "TestDashboard_Search",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_test.go",
        "proves": "Existing search consumes typed t/e/s/t and Esc exits search.",
        "limits": "Cards-pane search; does not explicitly exercise session-pane e search.",
        "reviewedHash": "fe042e81c3207ef759933378eca1b1ec1ba5f30f1aa3a1079c017b33289025bd",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.607776067Z",
          "scope": "TestDashboard_Search in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 315,
        "endLine": 347
      },
      {
        "id": "rename-persist",
        "name": "TestSetNamePropagatesPersistFailure",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/lifecycle_test.go",
        "line": 413,
        "endLine": 451,
        "proves": "Existing checked rename surfaces a record-write failure.",
        "limits": "Manager boundary only; no complete dashboard rename/restart journey.",
        "reviewedHash": "8c7059d4e1eccf3c2b67008f3a76171f4d7022ce691d3ddee93a06c0eedc23d8",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./...",
          "at": "2026-11-21T21:47:25.293694226Z",
          "scope": "TestSetNamePropagatesPersistFailure",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "rename-launcher",
        "name": "TestSessionLauncherWaitsForNameResultAndAllowsRetry",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/session_launcher_test.go",
        "proves": "Prior assessment: Existing launcher naming waits for result and permits retry.",
        "limits": "Launch naming form, not dashboard completion routing. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.144731611Z",
          "scope": "TestSessionLauncherWaitsForNameResultAndAllowsRetry",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        },
        "line": 136,
        "endLine": 161,
        "reviewedHash": "d2a35f3d77fc4e93d3a29699a8d0fc9aa05f053d9649c768855ac0603b9c8c4b"
      },
      {
        "id": "mcp-navigation",
        "name": "TestMCPManagerNestedNavigation",
        "file": "internal/ui/navigation_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Repeated form cancel → manager → dashboard journeys preserve destinations.",
        "limits": "No real terminal or rendered key delivery.",
        "reviewedHash": "9fa0fed09a03953dbb6e6fefe0bea9efb4a9106d62f33f92cd956280309652c7",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.505161Z",
          "scope": "TestMCPManagerNestedNavigation in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 112,
        "endLine": 137
      },
      {
        "id": "mcp-metadata",
        "name": "TestMCPManagerEditPreservesMetadataAndNewIsBlank",
        "file": "internal/ui/mcp_manager_review_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Existing ID/description survives edit; New clears fields and metadata.",
        "limits": "In-process coordinator with isolated repository; no process restart. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.473940762Z",
          "scope": "TestMCPManagerEditPreservesMetadataAndNewIsBlank in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 74,
        "endLine": 126
      },
      {
        "id": "mcp-parser-retained",
        "name": "TestMCPFormRoundTripsEveryField",
        "file": "internal/ui/views/mcp_form_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Retained assertions round-trip visible fields including escaped args/env/headers.",
        "limits": "Direct form callback; no application navigation. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.643641114Z",
          "scope": "TestMCPFormRoundTripsEveryField in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 148,
        "endLine": 199
      },
      {
        "id": "mcp-cross-dashboard-saves",
        "name": "TestMCPManagerCrossDashboardSaveResults",
        "file": "internal/ui/mcp_manager_gaps_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Delayed successful/failed saves preserve the other dashboard's form draft, navigation and current error; successful global mutation refreshes with current scope.",
        "limits": "ENOSPC is injected at the repository boundary; does not fill a disk or exercise physical-media failure.",
        "reviewedHash": "cd0f5c2a0cfded6055d3b2e8893ed28939d09bd36151e962666f312e21f0dc38",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.394234128Z",
          "scope": "TestMCPManagerCrossDashboardSaveResults in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 176,
        "endLine": 182
      },
      {
        "id": "mcp-write-failure-retry",
        "name": "TestMCPManagerWriteFailuresRetainDataAndAllowRetry",
        "file": "internal/ui/mcp_manager_gaps_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Save/delete PathError(ENOSPC) remains visible, registry bytes stay identical, drafts/rows survive, and retry persists successfully.",
        "limits": "Injects errors at the repository mutation boundary to avoid chmod/root dependence; atomic filesystem fault internals are outside this test.",
        "reviewedHash": "cd0f5c2a0cfded6055d3b2e8893ed28939d09bd36151e962666f312e21f0dc38",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.459333413Z",
          "scope": "TestMCPManagerWriteFailuresRetainDataAndAllowRetry in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 268,
        "endLine": 274
      },
      {
        "id": "mcp-mcp-terminal-journey",
        "name": "TestMCPManagerJourney",
        "file": "tests/e2e/mcp_manager_test.go",
        "method": "automated",
        "integration": "project",
        "kind": "E2E",
        "proves": "Actual executable: dashboard list → New/create → populated edit → rename preserving ID/command → Delete/cancel → Delete/confirm → return/reopen; six screenshots and saved-file assertions.",
        "limits": "No agents/model calls or screenshot baseline comparison. List/modal checked at 80×24; forms at 120×40. Historical observation retained in Appedia.",
        "reviewedHash": "0b63e84ab8c3c871017132e66b9a9c3866d6d83d517e59ca1d71a3710a4b92bf",
        "result": {
          "status": "passed",
          "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
          "at": "2026-11-11T22:34:16.106878795Z",
          "scope": "Real executable configuration journey, no agent/model calls",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        },
        "line": 27,
        "endLine": 113
      },
      {
        "id": "mcp-stale-commands",
        "name": "TestMCPManagerStaleRequests",
        "file": "internal/ui/mcp_manager_review_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Old edit/delete requests rejected; new deletion succeeds and delayed result does not navigate.",
        "limits": "Same-dashboard leave/reopen; no explicit cross-dashboard delayed-command assertion. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.466708603Z",
          "scope": "TestMCPManagerStaleRequests in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 14,
        "endLine": 72
      },
      {
        "id": "mcp-modal-target",
        "name": "TestMCPManagerConfirmationKeepsTargetAcrossRefresh",
        "file": "internal/ui/views/mcp_manager_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Modal captures ID across reorder, Escape cancels, nil data retains loaded rows.",
        "limits": "View-level messages; repository result is covered separately. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.028448498Z",
          "scope": "TestMCPManagerConfirmationKeepsTargetAcrossRefresh in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 15,
        "endLine": 53
      },
      {
        "id": "mcp-cross-dashboard-requests",
        "name": "TestMCPManagerCrossDashboardRequests",
        "file": "internal/ui/mcp_manager_gaps_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Old edit/delete messages are rejected after dashboard switch without clearing the new visit's confirmation; fresh deletion succeeds.",
        "limits": "Real coordinator and file repository; command delivery is held deterministically rather than scheduled by the terminal.",
        "reviewedHash": "cd0f5c2a0cfded6055d3b2e8893ed28939d09bd36151e962666f312e21f0dc38",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.360397551Z",
          "scope": "TestMCPManagerCrossDashboardRequests in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 135,
        "endLine": 141
      },
      {
        "id": "mcp-cross-dashboard-deletes",
        "name": "TestMCPManagerCrossDashboardDeleteResults",
        "file": "internal/ui/mcp_manager_gaps_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Delayed success/failure cannot navigate or overwrite the new dashboard's error; success refreshes its global registry; failure releases pending state for retry.",
        "limits": "Two actual persisted dashboards and an isolated file registry; injected repository failure, no real concurrent disk outage.",
        "reviewedHash": "cd0f5c2a0cfded6055d3b2e8893ed28939d09bd36151e962666f312e21f0dc38",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.428606066Z",
          "scope": "TestMCPManagerCrossDashboardDeleteResults in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 224,
        "endLine": 230
      },
      {
        "id": "mcp-referenced",
        "name": "TestMCPManagerReferencedDeleteReportsError",
        "file": "internal/ui/mcp_manager_review_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Saved card reference blocks deletion, retains server and displays an error.",
        "limits": "Reference belongs to default dashboard; no injected disk-write failure. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:07.480588502Z",
          "scope": "TestMCPManagerReferencedDeleteReportsError in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 128,
        "endLine": 180
      },
      {
        "id": "mcp-protected",
        "name": "TestMCPManagerProtectedAndNew",
        "file": "internal/ui/views/mcp_manager_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Built-in Enter/Delete return errors; New opens blank and cannot be deleted.",
        "limits": "No real CLI or remote MCP server. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.028542928Z",
          "scope": "TestMCPManagerProtectedAndNew in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 55,
        "endLine": 91
      },
      {
        "id": "mcp-repository-retained",
        "name": "TestKankuroMCPIsOfferedOnceAndProtected",
        "file": "internal/config/builtin_mcp_test.go",
        "line": 11,
        "endLine": 44,
        "method": "automated",
        "integration": "project",
        "proves": "Exactly one built-in; repository blocks edits, deletion and name shadowing.",
        "limits": "Isolated file repository; no UI.",
        "reviewedHash": "b28e8471e9e7f5e3c6ef47d89afdbdee27458b6a74c3da4576c34a705ca1f7b6",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.070560682Z",
          "scope": "TestKankuroMCPIsOfferedOnceAndProtected in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "mcp-long-list",
        "name": "TestMCPManagerLongListFits",
        "file": "internal/ui/views/mcp_manager_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Forty long Unicode/ANSI names fit 80×22; New reachable; confirmation fits.",
        "limits": "Bounds assertions only, no screenshot baseline or stable row-position assertion. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.075495084Z",
          "scope": "TestMCPManagerLongListFits in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 93,
        "endLine": 111
      },
      {
        "id": "mcp-sources-retained",
        "name": "TestSourceList_Navigation",
        "file": "internal/ui/components/source_list_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Retained source-list arrow navigation and New position.",
        "limits": "Does not assert every source-manager CRUD path.",
        "reviewedHash": "9f7c6ae7caf62037b2131c47e7e669a9b1d520fe3f68193add2229f5bee5cf8f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:57.751196515Z",
          "scope": "TestSourceList_Navigation in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 106,
        "endLine": 138
      },
      {
        "id": "focus-completion",
        "name": "TestLaunchFocusSnapshotBeforeCompletion",
        "file": "internal/ui/launch_focus_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Launch completion selects the exact created ID, including a snapshot already loaded.",
        "limits": "Coordinator events with controlled snapshots; native attachment is assessed separately. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:23.342542172Z",
          "scope": "TestLaunchFocusSnapshotBeforeCompletion",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        },
        "line": 97,
        "endLine": 104,
        "reviewedHash": "e86f812b702fe9c0f28a16945fb4a45c793251c98d110b7aa096e251f2e46874"
      },
      {
        "id": "focus-search",
        "name": "TestDashboard_FocusSessionWaitsForDataAndClearsOnlySessionSearch",
        "file": "internal/ui/views/dashboard_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Focus waits for target data, clears hiding session search, selects sessions pane and aligns preview; later refreshes do not steal focus.",
        "limits": "Direct view-level data refresh. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:09.603456457Z",
          "scope": "TestDashboard_FocusSessionWaitsForDataAndClearsOnlySessionSearch in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 232,
        "endLine": 269
      },
      {
        "id": "focus-epochs",
        "name": "TestLaunchFocusRejectsRefreshStartedBeforeLaunchAndName",
        "file": "internal/ui/launch_focus_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Old session epochs cannot displace the created or confirmed-name target.",
        "limits": "Deterministic snapshot ordering, not real scheduler timing. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:23.346874942Z",
          "scope": "TestLaunchFocusRejectsRefreshStartedBeforeLaunchAndName",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        },
        "line": 106,
        "endLine": 142,
        "reviewedHash": "e86f812b702fe9c0f28a16945fb4a45c793251c98d110b7aa096e251f2e46874"
      },
      {
        "id": "rename-epochs",
        "name": "TestInlineRenameRejectsOlderRefreshSnapshots",
        "file": "internal/ui/inline_rename_refresh_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Full, sessions and fingerprint snapshots started before inline rename do not restore the old name.",
        "limits": "In-process command/result delivery. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:04.265524491Z",
          "scope": "TestInlineRenameRejectsOlderRefreshSnapshots in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 17,
        "endLine": 76
      },
      {
        "id": "focus-warning",
        "name": "TestLaunchFocusPersistenceWarningStillSelectsSession",
        "file": "internal/ui/launch_focus_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Usable launch with metadata persistence warning keeps selected focus and exposes the warning.",
        "limits": "Controlled ErrPersistFailed result, not a physical disk fault. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:23.359379541Z",
          "scope": "TestLaunchFocusPersistenceWarningStillSelectsSession",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        },
        "line": 164,
        "endLine": 176,
        "reviewedHash": "e86f812b702fe9c0f28a16945fb4a45c793251c98d110b7aa096e251f2e46874"
      },
      {
        "id": "focus-stale",
        "name": "TestLaunchFocusIgnoresStaleResults",
        "file": "internal/ui/launch_focus_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Stale dashboard/generation results cannot move session focus.",
        "limits": "Does not launch a native agent. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:23.367669381Z",
          "scope": "TestLaunchFocusIgnoresStaleResults",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        },
        "line": 178,
        "endLine": 204,
        "reviewedHash": "e86f812b702fe9c0f28a16945fb4a45c793251c98d110b7aa096e251f2e46874"
      },
      {
        "id": "focus-cancel",
        "name": "TestLaunchFocusPendingRequestIsCancelledOnDashboardExit",
        "file": "internal/ui/launch_focus_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Leaving a dashboard clears pending focus; delayed launch cannot apply in a different context.",
        "limits": "Controlled coordinator events. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:23.37147723Z",
          "scope": "TestLaunchFocusPendingRequestIsCancelledOnDashboardExit",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        },
        "line": 206,
        "endLine": 225,
        "reviewedHash": "e86f812b702fe9c0f28a16945fb4a45c793251c98d110b7aa096e251f2e46874"
      },
      {
        "id": "focus-order",
        "name": "TestSessionList_SetSessionsPreservesSelectionByID",
        "file": "internal/ui/components/session_list_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Prior assessment: Newer session insertion preserves the selected ID rather than its row index.",
        "limits": "Direct list component assertion. The current source differs from that assessment and needs a fresh assertion review.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:57.748437855Z",
          "scope": "TestSessionList_SetSessionsPreservesSelectionByID in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        },
        "line": 284,
        "endLine": 304
      },
      {
        "id": "launch-focus-terminal",
        "name": "TestNativeLaunchFocusReturn",
        "file": "tests/e2e/launch_focus_test.go",
        "method": "automated",
        "integration": "project",
        "kind": "E2E",
        "proves": "Real Codex startup screen after launch/name; detach restores selected created session; Enter reattaches to same live process, preserving exactly one older terminated session.",
        "limits": "Startup-only; no agent prompt/provider inference or pixel-baseline comparison. Requires Codex installed. Historical observation retained in Appedia. Execution shown here is from the earlier checkout. This refresh compiled the tagged package and ran the model-free capture check.",
        "result": {
          "status": "passed",
          "command": "go test -tags=e2e -count=1 -timeout=8m -json ./tests/e2e",
          "at": "2026-11-21T21:33:33.71233779Z",
          "scope": "TestNativeLaunchFocusReturn",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        },
        "line": 28,
        "endLine": 33,
        "reviewedHash": "b3322556b2918002831d1c7c5b35f9616ad746410f268d53a633f8d6ad461edb"
      },
      {
        "id": "center-author-order",
        "name": "TestRecentAuthoredReports",
        "file": "internal/session/authored_reports_test.go",
        "method": "automated",
        "integration": "project",
        "line": 15,
        "endLine": 60,
        "proves": "Rejects received/foreign reports, checks timestamp/sequence/ID order, omission, detached results, invalid limit and cancellation.",
        "limits": "In-memory store fixture; not a full agent journey.",
        "reviewedHash": "6bf7c1ee54b32e350c9c409601cbe6659113b6ffb871592640520920cbd0e28d",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.34838219Z",
          "scope": "TestRecentAuthoredReports in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-reload",
        "name": "TestRecentAuthoredReportsBoundsAndReload",
        "file": "internal/session/authored_reports_test.go",
        "method": "automated",
        "integration": "project",
        "line": 62,
        "endLine": 106,
        "proves": "Reloads retained history, clamps to 256 reports, preserves activation attribution and creates no notifications.",
        "limits": "Synthetic persisted history; does not execute a process relaunch.",
        "reviewedHash": "6bf7c1ee54b32e350c9c409601cbe6659113b6ffb871592640520920cbd0e28d",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.34929185Z",
          "scope": "TestRecentAuthoredReportsBoundsAndReload in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-takeover",
        "name": "TestCenterReportTakesOverAndRejectsLateWorkspace",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "line": 59,
        "endLine": 90,
        "proves": "An event reads reports while retaining ongoing workspace work; successful report content cancels it, and late workspace data cannot overwrite reports.",
        "limits": "Direct coordinator messages and fake port; no terminal I/O.",
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.768171718Z",
          "scope": "TestCenterReportTakesOverAndRejectsLateWorkspace in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-coalesce",
        "name": "TestCenterCoalescesAndReleasesStaleSetup",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "line": 91,
        "endLine": 113,
        "proves": "Multiple changes produce one follow-up and stale setup/current subscriptions are released after clear.",
        "limits": "Deterministic state transitions; not an OS goroutine leak audit.",
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.770112628Z",
          "scope": "TestCenterCoalescesAndReleasesStaleSetup in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-scope",
        "name": "TestCenterScopeFailureClearsCachedReports",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "line": 114,
        "endLine": 135,
        "proves": "Scope rejection clears cached reports and blocks immediate retry for unchanged invalid selection.",
        "limits": "Fake scope result; real membership separately tested.",
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.773279328Z",
          "scope": "TestCenterScopeFailureClearsCachedReports in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-retry",
        "name": "TestCenterRetryClockIsNonOverlappingAndCanceled",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "line": 151,
        "endLine": 171,
        "proves": "Fake clock observes 30 seconds, excludes overlapping timer/setup and triggers resubscription first.",
        "limits": "No wall-clock or long-lived live-terminal observation.",
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.773312617Z",
          "scope": "TestCenterRetryClockIsNonOverlappingAndCanceled in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-failures",
        "name": "TestCenterFailurePolicyRetainsOnlySameSessionData",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "line": 173,
        "endLine": 205,
        "proves": "Read failure retains stale reports, confirmed empty clears them, workspace failure retains same-session data/warnings, clear drops content.",
        "limits": "Selected state-machine paths; not exhaustive cross-product of failures.",
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.776354937Z",
          "scope": "TestCenterFailurePolicyRetainsOnlySameSessionData in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-closed",
        "name": "TestCenterClosedSubscriptionUsesRetry",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "line": 207,
        "endLine": 217,
        "proves": "Unexpected stream closure preserves cached reports and sets warning/retry state.",
        "limits": "Tests scheduling, not a complete subsequent recovery journey.",
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.779775817Z",
          "scope": "TestCenterClosedSubscriptionUsesRetry in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-fit",
        "name": "TestSessionCenterReportsFitAndSanitize",
        "file": "internal/ui/components/session_center_test.go",
        "method": "automated",
        "integration": "project",
        "line": 14,
        "endLine": 48,
        "proves": "ANSI stripped; report text retained within 20×4; changing session clears it; zero size is safe.",
        "limits": "One explicit size plus zero-size boundary; no pixel comparison or complete resize/refetch journey.",
        "reviewedHash": "12aa0d07db33ff7d1122a48372e2d09222ba950c4a3fa4d12c8927ae466c86b6",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.77678563Z",
          "scope": "TestSessionCenterReportsFitAndSanitize in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-service-scope",
        "name": "TestCenterReportsSubscriptionAndScope",
        "file": "internal/api/center_test.go",
        "line": 18,
        "endLine": 72,
        "method": "automated",
        "integration": "project",
        "proves": "Real service reads authored progress after subscription signal, rejects foreign dashboard and allows idempotent release.",
        "limits": "Messaging harness injects records; no native model execution.",
        "reviewedHash": "02587d33e3c40f27ea04b389dbf5bd33c173e2bb5461c8a4af449282e4fe40ca",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.315938123Z",
          "scope": "TestCenterReportsSubscriptionAndScope in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-initial-read",
        "name": "TestCenterReportsChangeBeforeInitialReadAndStop",
        "file": "internal/api/center_test.go",
        "line": 74,
        "endLine": 129,
        "method": "automated",
        "integration": "project",
        "proves": "Report inserted after subscribe is read immediately and signals; stopping session keeps reports readable; cancellation closes changes.",
        "limits": "No full relaunch journey or screenshot.",
        "reviewedHash": "02587d33e3c40f27ea04b389dbf5bd33c173e2bb5461c8a4af449282e4fe40ca",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.348425441Z",
          "scope": "TestCenterReportsChangeBeforeInitialReadAndStop in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-rearm",
        "name": "TestCenterReportChangeDuringRearm",
        "file": "internal/api/center_test.go",
        "line": 144,
        "endLine": 188,
        "method": "automated",
        "integration": "project",
        "proves": "Controlled channels close while watch rearming is paused; a notice is retained and cancellation ends stream.",
        "limits": "Direct subscription helper fixture, not a stress/load measurement.",
        "reviewedHash": "02587d33e3c40f27ea04b389dbf5bd33c173e2bb5461c8a4af449282e4fe40ca",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.389389327Z",
          "scope": "TestCenterReportChangeDuringRearm in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-saved-path",
        "name": "TestCenterWorkspaceUsesSavedSessionDirectory",
        "file": "internal/api/center_test.go",
        "line": 131,
        "endLine": 142,
        "method": "automated",
        "integration": "project",
        "proves": "Uses the session saved directory and rejects foreign scope.",
        "limits": "Harness has a normal workspace; temporary identity checked separately.",
        "reviewedHash": "02587d33e3c40f27ea04b389dbf5bd33c173e2bb5461c8a4af449282e4fe40ca",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.389319507Z",
          "scope": "TestCenterWorkspaceUsesSavedSessionDirectory in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-temporary-path",
        "name": "TestCenterWorkspaceUsesTemporaryCopy",
        "file": "internal/api/center_test.go",
        "line": 190,
        "endLine": 212,
        "method": "automated",
        "integration": "project",
        "proves": "Launches a cat session into a temporary copy and reads exactly its saved WorkingDir.",
        "limits": "Does not invoke a model or assert recursive copy contents.",
        "reviewedHash": "02587d33e3c40f27ea04b389dbf5bd33c173e2bb5461c8a4af449282e4fe40ca",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.407353816Z",
          "scope": "TestCenterWorkspaceUsesTemporaryCopy in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-inventory-failure",
        "name": "TestCenterReportsInventoryFailureIsNotEmptyHistory",
        "file": "internal/api/center_test.go",
        "line": 214,
        "endLine": 225,
        "method": "automated",
        "integration": "project",
        "proves": "Corrupt card inventory makes the report read fail explicitly.",
        "limits": "One corrupted registry boundary.",
        "reviewedHash": "02587d33e3c40f27ea04b389dbf5bd33c173e2bb5461c8a4af449282e4fe40ca",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.439811993Z",
          "scope": "TestCenterReportsInventoryFailureIsNotEmptyHistory in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-git-directory",
        "name": "TestCenterWorkspaceGitAndDirectory",
        "file": "internal/api/center_workspace_test.go",
        "line": 32,
        "endLine": 67,
        "method": "automated",
        "integration": "project",
        "proves": "Checks clean Git vs non-Git, staged rename plus unstaged modification and newline untracked path while reading from a subdirectory.",
        "limits": "Temporary real Git fixture; conflict/submodule variants not asserted here.",
        "reviewedHash": "e40352e1edec3c9bd82223668bf76eed046bc407c10d343fe0bb862e700ab60f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.540855456Z",
          "scope": "TestCenterWorkspaceGitAndDirectory in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-bounded-directory",
        "name": "TestCenterWorkspaceBoundedListingAndErrors",
        "file": "internal/api/center_workspace_test.go",
        "line": 68,
        "endLine": 98,
        "method": "automated",
        "integration": "project",
        "proves": "More than 1,000 entries truncates; missing path and canceled inspection error; malformed/partial Git records rejected.",
        "limits": "Metadata access error combinations are not exhaustive.",
        "reviewedHash": "e40352e1edec3c9bd82223668bf76eed046bc407c10d343fe0bb862e700ab60f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.581393123Z",
          "scope": "TestCenterWorkspaceBoundedListingAndErrors in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-symlink-fallback",
        "name": "TestCenterWorkspaceSymlinkAndMissingGit",
        "file": "internal/api/center_workspace_test.go",
        "line": 100,
        "endLine": 126,
        "method": "automated",
        "integration": "project",
        "proves": "Immediate hidden/symlink rows retain link target; unavailable Git yields warned directory fallback.",
        "limits": "Unix fixture, not cross-platform owner/group verification.",
        "reviewedHash": "e40352e1edec3c9bd82223668bf76eed046bc407c10d343fe0bb862e700ab60f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.583544412Z",
          "scope": "TestCenterWorkspaceSymlinkAndMissingGit in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-linked-worktree",
        "name": "TestCenterWorkspaceLinkedWorktree",
        "file": "internal/api/center_workspace_test.go",
        "line": 128,
        "endLine": 146,
        "method": "automated",
        "integration": "project",
        "proves": "Real linked worktree is classified as Git and reports its modified tracked file.",
        "limits": "Single worktree and tracked-path example.",
        "reviewedHash": "e40352e1edec3c9bd82223668bf76eed046bc407c10d343fe0bb862e700ab60f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.662919756Z",
          "scope": "TestCenterWorkspaceLinkedWorktree in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-git-timeout",
        "name": "TestCenterWorkspaceGitTimeoutAndTruncation",
        "file": "internal/api/center_workspace_test.go",
        "line": 148,
        "endLine": 167,
        "method": "automated",
        "integration": "project",
        "proves": "Status entries cap at 1,000; a busy fake Git is canceled by an explicit short deadline.",
        "limits": "Does not wait out the production ten-second timeout.",
        "reviewedHash": "e40352e1edec3c9bd82223668bf76eed046bc407c10d343fe0bb862e700ab60f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.715558772Z",
          "scope": "TestCenterWorkspaceGitTimeoutAndTruncation in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-git-output",
        "name": "TestCenterWorkspaceKillsGitOnOutputLimit",
        "file": "internal/api/center_workspace_test.go",
        "line": 169,
        "endLine": 182,
        "method": "automated",
        "integration": "project",
        "proves": "Unbounded fake Git output fails on output budget before outer context expires.",
        "limits": "Checks stdout budget, not every stderr variant.",
        "reviewedHash": "e40352e1edec3c9bd82223668bf76eed046bc407c10d343fe0bb862e700ab60f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.763620438Z",
          "scope": "TestCenterWorkspaceKillsGitOnOutputLimit in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-failed-read-clock",
        "name": "TestCenterFailedReadCoalescesChangesUntilRetry",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Repeated store notices do not advance report requests while degraded; firing the fake clock allows exactly one retry and repeated failure reschedules recovery.",
        "limits": "Coordinator/fake-clock boundary, no live terminal.",
        "line": 233,
        "endLine": 261,
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.783345877Z",
          "scope": "TestCenterFailedReadCoalescesChangesUntilRetry in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-workspace-clock",
        "name": "TestCenterChangesDoNotRestartWorkspace",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Empty event reads do not refresh cached workspace before timer; a due timer refreshes it, and notices do not cancel/restart slow workspace work.",
        "limits": "Deterministic coordinator transitions, not real Git load timing.",
        "line": 263,
        "endLine": 304,
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.785393067Z",
          "scope": "TestCenterChangesDoNotRestartWorkspace in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-failed-workspace-clock",
        "name": "TestCenterFailedWorkspaceWithoutCacheWaitsForTimer",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Initial load occurs immediately; after failure without cache, repeated healthy empty report reads cannot retry workspace until fake clock fires.",
        "limits": "Tests policy state with fake application capability.",
        "line": 306,
        "endLine": 341,
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.788011396Z",
          "scope": "TestCenterFailedWorkspaceWithoutCacheWaitsForTimer in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-workspace-freshness",
        "name": "TestCenterWorkspaceStaleUntilSuccessfulWorkspaceOrReport",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Failed cached workspace remains stale through successful-empty and failed history reads and retry; successful workspace replacement or report takeover clears it.",
        "limits": "Coordinator state assertions; complete live journey remains unassessed.",
        "line": 343,
        "endLine": 394,
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.789513446Z",
          "scope": "TestCenterWorkspaceStaleUntilSuccessfulWorkspaceOrReport in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-inherited-pipes",
        "name": "TestCenterWorkspaceBoundsInheritedPipes",
        "file": "internal/api/center_workspace_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Fake Git starts a pipe-holding descendant; both deadline and output-limit paths finish before the child normal exit, using bounded WaitDelay.",
        "limits": "Unix shell/helper fixture with elapsed-time tolerance; not a process-tree termination guarantee.",
        "line": 184,
        "endLine": 216,
        "reviewedHash": "e40352e1edec3c9bd82223668bf76eed046bc407c10d343fe0bb862e700ab60f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.927339996Z",
          "scope": "TestCenterWorkspaceBoundsInheritedPipes in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-path-rows",
        "name": "TestSessionCenterWorkspaceNamesCannotForgeRows",
        "file": "internal/ui/components/session_center_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Git current/original paths, directory names and symlink targets flatten control whitespace; assertions reject forged status/warning rows and check display width.",
        "limits": "200-column renderer fixture; not pixel comparison.",
        "line": 150,
        "endLine": 203,
        "reviewedHash": "12aa0d07db33ff7d1122a48372e2d09222ba950c4a3fa4d12c8927ae466c86b6",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.78001102Z",
          "scope": "TestSessionCenterWorkspaceNamesCannotForgeRows in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-multiline",
        "name": "TestSessionCenterPreservesMultilineReport",
        "file": "internal/ui/components/session_center_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Report first/second lines retain their newline after introducing path-specific sanitization.",
        "limits": "Component render at 80×8, no native terminal.",
        "line": 205,
        "endLine": 215,
        "reviewedHash": "12aa0d07db33ff7d1122a48372e2d09222ba950c4a3fa4d12c8927ae466c86b6",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.78017745Z",
          "scope": "TestSessionCenterPreservesMultilineReport in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-navigation-reports",
        "name": "TestCenterNavigationRestoresSameSessionAndClearsOnEmpty",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Navigation detaches successful reports and cancels resources; return restores the same session, a failed reread renders cached reports stale, and confirmed empty remains cleared after another round trip.",
        "limits": "Coordinator/navigation/component boundary with fake report reads; no native terminal journey.",
        "line": 396,
        "endLine": 440,
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.791010466Z",
          "scope": "TestCenterNavigationRestoresSameSessionAndClearsOnEmpty in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-navigation-isolation",
        "name": "TestCenterNavigationCacheNeverMovesToAnotherSelection",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Switching selected stable session while away restores no prior-session workspace or reports.",
        "limits": "One dashboard/two-session fixture; not an exhaustive multi-dashboard journey.",
        "line": 442,
        "endLine": 464,
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.792232236Z",
          "scope": "TestCenterNavigationCacheNeverMovesToAnotherSelection in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-navigation-workspace",
        "name": "TestCenterNavigationRestoresWorkspaceAndClearsOnScopeLoss",
        "file": "internal/ui/center_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Same-session navigation restores workspace data through a report error; scope rejection clears live and saved content.",
        "limits": "Fake report/scope results, with separate service membership tests.",
        "line": 466,
        "endLine": 496,
        "reviewedHash": "6e2a95606fcf1ac1d054b976b2fdbdd78e831a3a6d3e45f8151f8c243c789cee",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.794419676Z",
          "scope": "TestCenterNavigationRestoresWorkspaceAndClearsOnScopeLoss in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-helper-cancel",
        "name": "TestCenterWorkspaceCancellationKillsGitHelpers",
        "file": "internal/api/center_process_unix_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "A controlled Git fixture starts a child process. Cancellation returns promptly, the child exits, and an unrelated process survives.",
        "limits": "Unix process-group fixture; no equivalent process-tree guarantee is established for non-Unix fallback or helpers that escape the group.",
        "line": 21,
        "endLine": 77,
        "reviewedHash": "1c2284cc957f4bc6934a88255e19f0d179da9ad024804115d46152e912c2be04",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.154761785Z",
          "scope": "TestCenterWorkspaceCancellationKillsGitHelpers in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-helper-output",
        "name": "TestCenterWorkspaceOutputLimitKillsGitHelpers",
        "file": "internal/api/center_process_unix_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Fake Git with a sleep child exceeds output budget; command errors and helper is no longer running.",
        "limits": "Unix helper fixture; zombie state counts as terminated pending OS parent reaping.",
        "line": 79,
        "endLine": 109,
        "reviewedHash": "1c2284cc957f4bc6934a88255e19f0d179da9ad024804115d46152e912c2be04",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.253318758Z",
          "scope": "TestCenterWorkspaceOutputLimitKillsGitHelpers in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-header-kind",
        "name": "TestSessionCenterReportHeaderKeepsKindWhenItFits",
        "file": "internal/ui/components/session_center_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "80×2 keeps full date/time, blocked kind and body; 20×2 uses compact time while retaining kind and body.",
        "limits": "Two explicit widths; does not assert every kind/width combination or screenshot comparison.",
        "line": 217,
        "endLine": 238,
        "reviewedHash": "12aa0d07db33ff7d1122a48372e2d09222ba950c4a3fa4d12c8927ae466c86b6",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.78022512Z",
          "scope": "TestSessionCenterReportHeaderKeepsKindWhenItFits in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-header-render",
        "name": "TestSessionCenterKeepsSessionHeaderAboveBody",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/components/session_center_test.go",
        "proves": "Verifies the legacy metadata order and artwork, live refresh of the same row, workspace-to-report takeover beneath the header, and rejection of old-session data.",
        "limits": "Component output and selection assertions. No pixel baseline comparison.",
        "reviewedHash": "12aa0d07db33ff7d1122a48372e2d09222ba950c4a3fa4d12c8927ae466c86b6",
        "line": 50,
        "endLine": 111,
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.77776365Z",
          "scope": "TestSessionCenterKeepsSessionHeaderAboveBody in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-header-size",
        "name": "TestSessionCenterReservesSpaceForBody",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/components/session_center_test.go",
        "proves": "Checks bounded rows and columns, workspace-tail truncation, optional art and the body-derived report limit across sizes.",
        "limits": "Component dimensions. The separate E2E screenshot shows the 80×24 dashboard.",
        "reviewedHash": "12aa0d07db33ff7d1122a48372e2d09222ba950c4a3fa4d12c8927ae466c86b6",
        "line": 113,
        "endLine": 147,
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.7787781Z",
          "scope": "TestSessionCenterReservesSpaceForBody in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "startup-puppet-ordered",
        "name": "TestCoordinatorStartupReadsRegistriesWithoutResyncingPuppets",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/refresh_certification_internal_test.go",
        "proves": "Executes the UI startup batch, confirms a valid initial registry baseline and checks that the UI never starts a puppet sync.",
        "limits": "Coordinator commands and repository fixture. The separate terminal smoke drives the real executable.",
        "reviewedHash": "3f8d69d3a93233374abe267c1fdd44aa4c276723ab0dc922b1872cb87f140d56",
        "line": 26,
        "endLine": 46,
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.48073483Z",
          "scope": "TestCoordinatorStartupReadsRegistriesWithoutResyncingPuppets",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "refresh-token-retry",
        "name": "TestRefreshTokenCertificationFailuresRetryAndKeepInventory",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/refresh_certification_internal_test.go",
        "proves": "For before/after token errors and mismatches, verifies usable inventory remains visible and a later refresh retries and establishes a stable token.",
        "limits": "Controlled token sequence. Filesystem startup ordering is covered separately.",
        "reviewedHash": "3f8d69d3a93233374abe267c1fdd44aa4c276723ab0dc922b1872cb87f140d56",
        "line": 77,
        "endLine": 112,
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.495405299Z",
          "scope": "TestRefreshTokenCertificationFailuresRetryAndKeepInventory",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "startup-puppet-terminal",
        "name": "TestPuppetStartupSmoke",
        "kind": "E2E",
        "method": "automated",
        "integration": "fallback",
        "fallbackReason": "A temporary Go test overlay reuses the existing tui-test harness so the Appedia refresh could verify the real binary without changing project test files.",
        "file": "appedia/evidence/current/startup-smoke_test.go",
        "reviewedHash": "205032814eca6e1c3b96e71cbff72264f1dddc938901bf61b9e24bd4ce3a264b",
        "line": 278,
        "endLine": 372,
        "proves": "Starts a fresh real binary with a blank puppet registry, observes custom artwork on first dashboard open, verifies a missing source raises a nonfatal warning, and refreshes edited artwork through Source Manager.",
        "limits": "Two isolated Linux startups. No model provider, boot timing measurements or pixel comparison.",
        "result": {
          "status": "passed",
          "command": "go test -json -overlay <temporary fixture> -tags=e2e ./tests/e2e -run '^TestPuppetStartupSmoke$' -count=1",
          "at": "2026-11-19T21:27:07.942396604Z",
          "scope": "TestPuppetStartupSmoke",
          "revision": "WORKTREE",
          "fingerprint": "1594c215e9ef7de8ae058411320bd1b0ffc6420de9a370d636424f33cb67a9ce"
        }
      },
      {
        "id": "center-split-index",
        "name": "TestCenterWorkspaceSplitIndexStatus",
        "file": "internal/api/center_workspace_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Both ordinary and linked worktrees using split indexes return the modified tracked file as Git status.",
        "limits": "Real local Git on Linux.",
        "line": 289,
        "endLine": 314,
        "reviewedHash": "e40352e1edec3c9bd82223668bf76eed046bc407c10d343fe0bb862e700ab60f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.108573592Z",
          "scope": "TestCenterWorkspaceSplitIndexStatus in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-helper-isolation",
        "name": "TestCenterWorkspaceIgnoresConfiguredGitHelpers",
        "file": "internal/api/center_workspace_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "A split-index repository reports clean status with fsmonitor disabled. Configured clean filters trigger the labelled fallback without executing the sentinel helper.",
        "limits": "Covers fsmonitor and the configured clean-filter fallback.",
        "line": 218,
        "endLine": 250,
        "reviewedHash": "e40352e1edec3c9bd82223668bf76eed046bc407c10d343fe0bb862e700ab60f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.00230673Z",
          "scope": "TestCenterWorkspaceIgnoresConfiguredGitHelpers in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-metadata-fallback",
        "name": "TestCenterWorkspaceFallsBackAfterMetadataProbeFailure",
        "file": "internal/api/center_workspace_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "A failing metadata probe yields a warning and preserves a readable file in the directory listing.",
        "limits": "A symlink loop triggers the metadata error.",
        "line": 252,
        "endLine": 271,
        "reviewedHash": "e40352e1edec3c9bd82223668bf76eed046bc407c10d343fe0bb862e700ab60f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.00312919Z",
          "scope": "TestCenterWorkspaceFallsBackAfterMetadataProbeFailure in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-report-signal-eviction",
        "name": "TestReportSignalTracksDeliveryEvictionOfLegacyReplies",
        "file": "internal/session/messaging_review_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "An ordinary delivery update wakes general waiters only. Evicting retained replies wakes report readers.",
        "limits": "Uses a near-capacity message store with legacy replies.",
        "line": 598,
        "endLine": 628,
        "reviewedHash": "db2f9231e1ba49bda149ca04d6fa940b16b9cb1ab02f671f7701fc476e62c8d8",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:04.201030896Z",
          "scope": "TestReportSignalTracksDeliveryEvictionOfLegacyReplies in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-report-signal-commit",
        "name": "TestReportSignalSkipsReplayAndFailedWrite",
        "file": "internal/session/messaging_review_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "A persisted report wakes report readers. Replaying the report or failing a new write leaves the signal unchanged.",
        "limits": "Message-store test with a forced persistence error.",
        "line": 630,
        "endLine": 664,
        "reviewedHash": "db2f9231e1ba49bda149ca04d6fa940b16b9cb1ab02f671f7701fc476e62c8d8",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:07:04.202174476Z",
          "scope": "TestReportSignalSkipsReplayAndFailedWrite in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-report-scope-signal",
        "name": "TestCenterReportScopeLossWithoutReportChanges",
        "file": "internal/api/center_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "The subscription closes on scope loss without requiring a report change.",
        "limits": "Application service fixture.",
        "line": 227,
        "endLine": 249,
        "reviewedHash": "02587d33e3c40f27ea04b389dbf5bd33c173e2bb5461c8a4af449282e4fe40ca",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:55.467178521Z",
          "scope": "TestCenterReportScopeLossWithoutReportChanges in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "center-report-timestamp-fit",
        "name": "TestSessionCenterOmitsUntimestampedReport",
        "file": "internal/ui/components/session_center_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "When an older report cannot fit with its timestamp, omit it. A newest report can use a compact one-row timestamp.",
        "limits": "Component assertions at bounded sizes.",
        "line": 240,
        "endLine": 269,
        "reviewedHash": "12aa0d07db33ff7d1122a48372e2d09222ba950c4a3fa4d12c8927ae466c86b6",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-20T00:06:56.78134072Z",
          "scope": "TestSessionCenterOmitsUntimestampedReport in the ordinary Go suite",
          "revision": "5c80ec87119819563ad7b2253ed9f58cbedbd257",
          "fingerprint": "8083453596feae03e5a7fd245afac917c73bab578245d127b786000ba2b31a66"
        }
      },
      {
        "id": "effort-draft",
        "name": "TestCardEditor_Card",
        "file": "internal/ui/views/card_editor_test.go",
        "kind": "unit",
        "method": "automated",
        "integration": "project",
        "line": 83,
        "endLine": 119,
        "proves": "The draft trims a padded effort string to high.",
        "limits": "Sets the input directly in the view fixture.",
        "reviewedHash": "ea3731dd18253155280647a4c78b98ef9b808b9de498ca3f60f679aa78fdc0e2",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./internal/driver ./internal/session ./internal/api ./internal/ui/views -run '^(TestReasoningEffortRegistry|TestCodexResumeEffortOrdering|TestClaudeResumeLifecycleUsesSelectedNativeID|TestEditor(SavePreservesUneditableFieldsAndRejectsStaleRevision|ExplicitClearAndHiddenFieldConflict|SaveResultRevisionMatchesPersistedCard)|TestCardEditor(_Card|FocusedFieldAlwaysVisible|SharedAndConsumerReadOnly)|TestService(LaunchResolvesCurrentSavedCard|RelaunchAndSwitchResolveLatestSavedCards))$'",
          "at": "2026-11-21T21:32:36.971590408Z",
          "scope": "TestCardEditor_Card",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "effort-layout",
        "name": "TestCardEditorFocusedFieldAlwaysVisible",
        "file": "internal/ui/views/card_editor_layout_test.go",
        "kind": "unit",
        "method": "automated",
        "integration": "project",
        "line": 263,
        "endLine": 286,
        "proves": "Cycles all focus targets and asserts that the focused row stays visible and within the view at 80×24 and the smaller component fixture.",
        "limits": "Measures rendered text in the component.",
        "reviewedHash": "5d0e2dc83803768c7b520d21f5cb4ff06b25a5743664836064aaea4c05cf88c6",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./internal/driver ./internal/session ./internal/api ./internal/ui/views -run '^(TestReasoningEffortRegistry|TestCodexResumeEffortOrdering|TestClaudeResumeLifecycleUsesSelectedNativeID|TestEditor(SavePreservesUneditableFieldsAndRejectsStaleRevision|ExplicitClearAndHiddenFieldConflict|SaveResultRevisionMatchesPersistedCard)|TestCardEditor(_Card|FocusedFieldAlwaysVisible|SharedAndConsumerReadOnly)|TestService(LaunchResolvesCurrentSavedCard|RelaunchAndSwitchResolveLatestSavedCards))$'",
          "at": "2026-11-21T21:32:36.971344847Z",
          "scope": "TestCardEditorFocusedFieldAlwaysVisible",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "effort-create",
        "name": "TestEditorSaveResultRevisionMatchesPersistedCard",
        "file": "internal/api/card_editor_test.go",
        "kind": "integration",
        "method": "automated",
        "integration": "project",
        "line": 176,
        "endLine": 229,
        "proves": "Creates a card with high, reloads the saved value and checks its revision.",
        "limits": "Exercises the service and repository without a terminal.",
        "reviewedHash": "c10ea2d898eac2df31e809c0a1c13fa0074b198315d96daaa04c3bc9c9162374",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./internal/driver ./internal/session ./internal/api ./internal/ui/views -run '^(TestReasoningEffortRegistry|TestCodexResumeEffortOrdering|TestClaudeResumeLifecycleUsesSelectedNativeID|TestEditor(SavePreservesUneditableFieldsAndRejectsStaleRevision|ExplicitClearAndHiddenFieldConflict|SaveResultRevisionMatchesPersistedCard)|TestCardEditor(_Card|FocusedFieldAlwaysVisible|SharedAndConsumerReadOnly)|TestService(LaunchResolvesCurrentSavedCard|RelaunchAndSwitchResolveLatestSavedCards))$'",
          "at": "2026-11-21T21:32:36.902736219Z",
          "scope": "TestEditorSaveResultRevisionMatchesPersistedCard",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "effort-preserve",
        "name": "TestEditorSavePreservesUneditableFieldsAndRejectsStaleRevision",
        "file": "internal/api/card_editor_test.go",
        "kind": "integration",
        "method": "automated",
        "integration": "project",
        "line": 13,
        "endLine": 72,
        "proves": "A name-only save compares all other stored fields, including effort.",
        "limits": "Uses the service directly.",
        "reviewedHash": "c10ea2d898eac2df31e809c0a1c13fa0074b198315d96daaa04c3bc9c9162374",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./internal/driver ./internal/session ./internal/api ./internal/ui/views -run '^(TestReasoningEffortRegistry|TestCodexResumeEffortOrdering|TestClaudeResumeLifecycleUsesSelectedNativeID|TestEditor(SavePreservesUneditableFieldsAndRejectsStaleRevision|ExplicitClearAndHiddenFieldConflict|SaveResultRevisionMatchesPersistedCard)|TestCardEditor(_Card|FocusedFieldAlwaysVisible|SharedAndConsumerReadOnly)|TestService(LaunchResolvesCurrentSavedCard|RelaunchAndSwitchResolveLatestSavedCards))$'",
          "at": "2026-11-21T21:32:36.891130291Z",
          "scope": "TestEditorSavePreservesUneditableFieldsAndRejectsStaleRevision",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "effort-clear",
        "name": "TestEditorExplicitClearAndHiddenFieldConflict",
        "file": "internal/api/card_editor_test.go",
        "kind": "integration",
        "method": "automated",
        "integration": "project",
        "line": 111,
        "endLine": 174,
        "proves": "Clears a present effort and rejects the next edit after another writer changes effort.",
        "limits": "Uses the service directly.",
        "reviewedHash": "c10ea2d898eac2df31e809c0a1c13fa0074b198315d96daaa04c3bc9c9162374",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./internal/driver ./internal/session ./internal/api ./internal/ui/views -run '^(TestReasoningEffortRegistry|TestCodexResumeEffortOrdering|TestClaudeResumeLifecycleUsesSelectedNativeID|TestEditor(SavePreservesUneditableFieldsAndRejectsStaleRevision|ExplicitClearAndHiddenFieldConflict|SaveResultRevisionMatchesPersistedCard)|TestCardEditor(_Card|FocusedFieldAlwaysVisible|SharedAndConsumerReadOnly)|TestService(LaunchResolvesCurrentSavedCard|RelaunchAndSwitchResolveLatestSavedCards))$'",
          "at": "2026-11-21T21:32:36.89831121Z",
          "scope": "TestEditorExplicitClearAndHiddenFieldConflict",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "editor-read-only",
        "name": "TestCardEditorSharedAndConsumerReadOnly",
        "file": "internal/ui/views/card_editor_test.go",
        "kind": "unit",
        "method": "automated",
        "integration": "project",
        "line": 326,
        "endLine": 355,
        "proves": "Checks the consumer form is read-only and the save shortcut does not save.",
        "limits": "Retained test of the shared guard. It does not attempt to type in the new input.",
        "reviewedHash": "ea3731dd18253155280647a4c78b98ef9b808b9de498ca3f60f679aa78fdc0e2",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./internal/driver ./internal/session ./internal/api ./internal/ui/views -run '^(TestReasoningEffortRegistry|TestCodexResumeEffortOrdering|TestClaudeResumeLifecycleUsesSelectedNativeID|TestEditor(SavePreservesUneditableFieldsAndRejectsStaleRevision|ExplicitClearAndHiddenFieldConflict|SaveResultRevisionMatchesPersistedCard)|TestCardEditor(_Card|FocusedFieldAlwaysVisible|SharedAndConsumerReadOnly)|TestService(LaunchResolvesCurrentSavedCard|RelaunchAndSwitchResolveLatestSavedCards))$'",
          "at": "2026-11-21T21:32:36.975188927Z",
          "scope": "TestCardEditorSharedAndConsumerReadOnly",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "e2e-card-effort",
        "name": "editAndReopenReasoningEffort",
        "file": "tests/e2e/launch_focus_test.go",
        "kind": "E2E",
        "method": "automated",
        "integration": "project",
        "line": 35,
        "endLine": 56,
        "reviewedHash": "b3322556b2918002831d1c7c5b35f9616ad746410f268d53a633f8d6ad461edb",
        "proves": "Saves high with the terminal editor, reopens the saved card and matches high on the Reasoning effort row.",
        "limits": "Covers an existing card at 120×40. Clearing and stale saves are exercised by Go service tests. Execution shown here is from the earlier checkout. This refresh compiled the tagged package and ran the model-free capture check.",
        "result": {
          "status": "passed",
          "command": "go test -tags=e2e -count=1 -timeout=8m -json ./tests/e2e",
          "at": "2026-11-21T21:33:33.71233779Z",
          "scope": "TestNativeLaunchFocusReturn",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "effort-registry",
        "name": "TestReasoningEffortRegistry",
        "file": "internal/driver/driver_test.go",
        "kind": "unit",
        "method": "automated",
        "integration": "project",
        "line": 479,
        "endLine": 498,
        "proves": "Checks exact argument slices for the named drivers, OpenCode, another generic driver and empty effort.",
        "limits": "Calls driver methods. It does not execute native CLIs.",
        "reviewedHash": "2dd48a6061f3cb2c74e770edb92439288b14d703a7e982544523b48abfd2f540",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./internal/driver ./internal/session ./internal/api ./internal/ui/views -run '^(TestReasoningEffortRegistry|TestCodexResumeEffortOrdering|TestClaudeResumeLifecycleUsesSelectedNativeID|TestEditor(SavePreservesUneditableFieldsAndRejectsStaleRevision|ExplicitClearAndHiddenFieldConflict|SaveResultRevisionMatchesPersistedCard)|TestCardEditor(_Card|FocusedFieldAlwaysVisible|SharedAndConsumerReadOnly)|TestService(LaunchResolvesCurrentSavedCard|RelaunchAndSwitchResolveLatestSavedCards))$'",
          "at": "2026-11-21T21:32:36.592144072Z",
          "scope": "TestReasoningEffortRegistry",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "effort-claude",
        "name": "TestClaudeResumeLifecycleUsesSelectedNativeID",
        "file": "internal/session/claude_resume_test.go",
        "kind": "integration",
        "method": "automated",
        "integration": "project",
        "line": 70,
        "endLine": 186,
        "proves": "Captures model, changed effort and conversation arguments on launch, resume, switch and switch back.",
        "limits": "Uses a process fixture in place of native Claude.",
        "reviewedHash": "1bc77dd8da8a6f189300421e5801adc407c37c2dfe40d5ccc7818e3df9d316dc",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./internal/driver ./internal/session ./internal/api ./internal/ui/views -run '^(TestReasoningEffortRegistry|TestCodexResumeEffortOrdering|TestClaudeResumeLifecycleUsesSelectedNativeID|TestEditor(SavePreservesUneditableFieldsAndRejectsStaleRevision|ExplicitClearAndHiddenFieldConflict|SaveResultRevisionMatchesPersistedCard)|TestCardEditor(_Card|FocusedFieldAlwaysVisible|SharedAndConsumerReadOnly)|TestService(LaunchResolvesCurrentSavedCard|RelaunchAndSwitchResolveLatestSavedCards))$'",
          "at": "2026-11-21T21:32:36.908850248Z",
          "scope": "TestClaudeResumeLifecycleUsesSelectedNativeID",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "effort-codex",
        "name": "TestCodexResumeEffortOrdering",
        "file": "internal/session/reasoning_effort_test.go",
        "kind": "integration",
        "method": "automated",
        "integration": "project",
        "line": 28,
        "endLine": 78,
        "proves": "Checks the complete argument slice at Spawn for a saved Codex conversation.",
        "limits": "Uses a capture driver and does not exercise Codex’s parser.",
        "reviewedHash": "c9908e303ff40235c98eb6221d33a7775d5a4aa906f22d830cde63373d484939",
        "result": {
          "status": "passed",
          "command": "go test -json -count=1 ./internal/driver ./internal/session ./internal/api ./internal/ui/views -run '^(TestReasoningEffortRegistry|TestCodexResumeEffortOrdering|TestClaudeResumeLifecycleUsesSelectedNativeID|TestEditor(SavePreservesUneditableFieldsAndRejectsStaleRevision|ExplicitClearAndHiddenFieldConflict|SaveResultRevisionMatchesPersistedCard)|TestCardEditor(_Card|FocusedFieldAlwaysVisible|SharedAndConsumerReadOnly)|TestService(LaunchResolvesCurrentSavedCard|RelaunchAndSwitchResolveLatestSavedCards))$'",
          "at": "2026-11-21T21:32:36.932873804Z",
          "scope": "TestCodexResumeEffortOrdering",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "latest-launch",
        "name": "TestServiceLaunchResolvesCurrentSavedCard",
        "file": "internal/api/service_test.go",
        "kind": "integration",
        "method": "automated",
        "integration": "project",
        "line": 1356,
        "endLine": 1378,
        "proves": "Launch uses the current saved card name and preserves card identity, dashboard and workspace.",
        "limits": "Retained service regression. It does not inspect effort arguments.",
        "reviewedHash": "4c2f3e9342ceea29d5ff1d5ba392ce0d5c67b3e8b03c38bff8421942ca9bc9df",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.769695223Z",
          "scope": "TestServiceLaunchResolvesCurrentSavedCard",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "latest-switch",
        "name": "TestServiceRelaunchAndSwitchResolveLatestSavedCards",
        "file": "internal/api/service_test.go",
        "kind": "integration",
        "method": "automated",
        "integration": "project",
        "line": 1381,
        "endLine": 1423,
        "proves": "A relaunch retains deployed instructions and a switch syncs the latest destination instructions in the existing workspace.",
        "limits": "Retained service regression. It does not inspect effort arguments.",
        "reviewedHash": "4c2f3e9342ceea29d5ff1d5ba392ce0d5c67b3e8b03c38bff8421942ca9bc9df",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.781326522Z",
          "scope": "TestServiceRelaunchAndSwitchResolveLatestSavedCards",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "e2e-launch-focus",
        "name": "TestNativeLaunchFocusReturn",
        "file": "tests/e2e/launch_focus_test.go",
        "kind": "E2E",
        "method": "automated",
        "integration": "project",
        "line": 28,
        "endLine": 33,
        "reviewedHash": "b3322556b2918002831d1c7c5b35f9616ad746410f268d53a633f8d6ad461edb",
        "proves": "Opens the card editor, saves high and reopens the card before launching offline Codex. Detach and reattach preserve the same live process.",
        "limits": "Uses idle Codex and submits no prompt. The automated assertions check the editor value and launch behavior. Execution shown here is from the earlier checkout. This refresh compiled the tagged package and ran the model-free capture check.",
        "result": {
          "status": "passed",
          "command": "go test -tags=e2e -count=1 -timeout=8m -json ./tests/e2e",
          "at": "2026-11-21T21:33:33.71233779Z",
          "scope": "TestNativeLaunchFocusReturn",
          "revision": "WORKTREE",
          "fingerprint": "7a82c552124894775cf130a3a7bf0a9ca46f46b8cb29ba57ba60130da4aa129f"
        }
      },
      {
        "id": "workspace-feature-display",
        "name": "TestLaunchUsesDisplayedWorkspaceAfterCardEdit",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/launch_focus_internal_test.go",
        "line": 415,
        "endLine": 457,
        "proves": "A real local process launches in the displayed directory after the card default changes.",
        "limits": "Uses cat and a non-temporary launch.",
        "reviewedHash": "e86f812b702fe9c0f28a16945fb4a45c793251c98d110b7aa096e251f2e46874",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:23.404402668Z",
          "scope": "TestLaunchUsesDisplayedWorkspaceAfterCardEdit",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-cached",
        "name": "TestLaunchPreparationWorkspaceErrorKeepsCachedChoices",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/launch_focus_internal_test.go",
        "line": 459,
        "endLine": 491,
        "proves": "A stale result is ignored and an accepted read failure shows an error with cached options.",
        "limits": "Injects a registry read failure into the coordinator.",
        "reviewedHash": "e86f812b702fe9c0f28a16945fb4a45c793251c98d110b7aa096e251f2e46874",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:23.410633988Z",
          "scope": "TestLaunchPreparationWorkspaceErrorKeepsCachedChoices",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-cancel",
        "name": "TestWorkspaceDraftCancelBlocksQueuedEnter",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/launch_focus_internal_test.go",
        "line": 332,
        "endLine": 354,
        "proves": "Enter after Esc does not issue a launch.",
        "limits": "Uses the fake session action port.",
        "reviewedHash": "e86f812b702fe9c0f28a16945fb4a45c793251c98d110b7aa096e251f2e46874",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:23.38008506Z",
          "scope": "TestWorkspaceDraftCancelBlocksQueuedEnter",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-sources",
        "name": "TestWorkspaceSourcesNavigationIsScopedToDraft",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/launch_focus_internal_test.go",
        "line": 356,
        "endLine": 389,
        "proves": "Stale navigation is ignored and a valid sources visit returns to the same draft.",
        "limits": "Exercises message routing without native terminal input.",
        "reviewedHash": "e86f812b702fe9c0f28a16945fb4a45c793251c98d110b7aa096e251f2e46874",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:23.38595706Z",
          "scope": "TestWorkspaceSourcesNavigationIsScopedToDraft",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-sources-submit",
        "name": "TestWorkspaceSourcesNavigationRejectedAfterSubmission",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/launch_focus_internal_test.go",
        "line": 391,
        "endLine": 400,
        "proves": "A queued source navigation cannot interrupt an already submitted launch.",
        "limits": "Uses the fake session action port.",
        "reviewedHash": "e86f812b702fe9c0f28a16945fb4a45c793251c98d110b7aa096e251f2e46874",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:23.390011779Z",
          "scope": "TestWorkspaceSourcesNavigationRejectedAfterSubmission",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-quit",
        "name": "TestWorkspaceDraftCtrlCQuits",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/launch_focus_internal_test.go",
        "line": 402,
        "endLine": 413,
        "proves": "Ctrl+C from the form produces a quit message.",
        "limits": "Checks the coordinator command.",
        "reviewedHash": "e86f812b702fe9c0f28a16945fb4a45c793251c98d110b7aa096e251f2e46874",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:23.394079309Z",
          "scope": "TestWorkspaceDraftCtrlCQuits",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-draft",
        "name": "TestSessionLauncherWorkspaceDraft",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/session_launcher_test.go",
        "line": 181,
        "endLine": 220,
        "proves": "Initial labels match defaults and picker Enter is independent from form Enter.",
        "limits": "Checks view strings and callbacks at a configured minimum-sized view.",
        "reviewedHash": "d2a35f3d77fc4e93d3a29699a8d0fc9aa05f053d9649c768855ac0603b9c8c4b",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.147513171Z",
          "scope": "TestSessionLauncherWorkspaceDraft",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-promote",
        "name": "TestSessionLauncherPromotesSubWorkspace",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/session_launcher_test.go",
        "line": 222,
        "endLine": 243,
        "proves": "Promoting a sub-workspace submits it once as the main workspace.",
        "limits": "Checks the view callback payload.",
        "reviewedHash": "d2a35f3d77fc4e93d3a29699a8d0fc9aa05f053d9649c768855ac0603b9c8c4b",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.149348921Z",
          "scope": "TestSessionLauncherPromotesSubWorkspace",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-editor-promote",
        "name": "TestCardEditorPromotesSubWorkspace",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/card_editor_draft_test.go",
        "line": 170,
        "endLine": 188,
        "proves": "The same role change removes the duplicate in the editor draft.",
        "limits": "Does not save a real card.",
        "reviewedHash": "5eb342ddc0109a7773f52f8c8a593f25bce2c7c0f8e3ce53c1e09a4b8688ce06",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:15.791595756Z",
          "scope": "TestCardEditorPromotesSubWorkspace",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-picker-quit",
        "name": "TestWorkspacePickerCtrlCMatchesEditor",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/session_launcher_test.go",
        "line": 245,
        "endLine": 275,
        "proves": "Ctrl+C produces the same command type in both pickers.",
        "limits": "Checks the main workspace picker.",
        "reviewedHash": "d2a35f3d77fc4e93d3a29699a8d0fc9aa05f053d9649c768855ac0603b9c8c4b",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.153486031Z",
          "scope": "TestWorkspacePickerCtrlCMatchesEditor",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-picker-search",
        "name": "TestWorkspacePicker_EscapeBehavior",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/workspace_picker_test.go",
        "line": 180,
        "endLine": 208,
        "proves": "Escape clears the query before cancelling the shared picker.",
        "limits": "Calls the picker model directly.",
        "reviewedHash": "2b3e2e4efc955282f3e19101aa8d0a0576dd8a48c9488d80067e56e1cab7c9b2",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.181530749Z",
          "scope": "TestWorkspacePicker_EscapeBehavior",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-picker-size",
        "name": "TestWorkspacePicker_SmallTerminal",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/workspace_picker_test.go",
        "line": 341,
        "endLine": 361,
        "proves": "Exercises the shared picker with a small terminal size.",
        "limits": "Does not assert the full launch form layout.",
        "reviewedHash": "2b3e2e4efc955282f3e19101aa8d0a0576dd8a48c9488d80067e56e1cab7c9b2",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.196160378Z",
          "scope": "TestWorkspacePicker_SmallTerminal",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-transport",
        "name": "TestRealServiceLifecycleResultsOverMCP",
        "method": "automated",
        "integration": "project",
        "file": "internal/mcp/server_test.go",
        "line": 352,
        "endLine": 982,
        "proves": "Uses real HTTP discovery to launch with an unreferenced main workspace and two ordered sub-workspaces. Checks the copied files, excluded default, identical replay, reordered-ID conflict and no-grant discovery denial.",
        "limits": "Uses local directories and a generic test process without a model provider.",
        "reviewedHash": "52bac4fc72177e4e8a0510855a3b7e48253422c639ce4dc506012e43419d7ee8",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:15.698834472Z",
          "scope": "TestRealServiceLifecycleResultsOverMCP",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-retry",
        "name": "TestOperationRunnerCreatesOnceAndReplaysRequest",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 164,
        "endLine": 504,
        "proves": "Repeated create requests replay, and changed workspace IDs conflict.",
        "limits": "Checks the service operation runner.",
        "reviewedHash": "4c2f3e9342ceea29d5ff1d5ba392ce0d5c67b3e8b03c38bff8421942ca9bc9df",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.434609084Z",
          "scope": "TestOperationRunnerCreatesOnceAndReplaysRequest",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-admission",
        "name": "TestScopedLaunchRevalidatesChangedCardBeforePublication",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/membership_internal_test.go",
        "line": 403,
        "endLine": 458,
        "proves": "A card changed during activation cannot publish an obsolete launch.",
        "limits": "Exercises the retained admission rule on the default path.",
        "reviewedHash": "5c03e0ca7fa22fb2f702ffc7ab2a5ee661a6fb521c75935bdd83dc1719f97f46",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:15.968048495Z",
          "scope": "TestScopedLaunchRevalidatesChangedCardBeforePublication",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-manifest",
        "name": "TestSharedWorkspaceLaunchMetadataTracksSelection",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/workspace_test.go",
        "line": 1051,
        "endLine": 1101,
        "proves": "A later main-only launch reports no selected copies while preserving the shared manifest.",
        "limits": "Uses the manager directly.",
        "reviewedHash": "6c86561d1c8644673dd403c6ee0d8c5ea5e455e501864ff5436e3f1a23c9d52e",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:37.961604614Z",
          "scope": "TestSharedWorkspaceLaunchMetadataTracksSelection",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-copy",
        "name": "TestLaunchCopiesSubWorkspaceAndRetainsManagedEdits",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/workspace_test.go",
        "line": 402,
        "endLine": 471,
        "proves": "Managed copies persist and retain edits through another launch.",
        "limits": "Uses the manager directly.",
        "reviewedHash": "6c86561d1c8644673dd403c6ee0d8c5ea5e455e501864ff5436e3f1a23c9d52e",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:37.663282543Z",
          "scope": "TestLaunchCopiesSubWorkspaceAndRetainsManagedEdits",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-cleanup",
        "name": "TestFailedTemporaryLaunchRemovesSessionRoot",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/workspace_test.go",
        "line": 473,
        "endLine": 506,
        "proves": "Failed temporary preparation removes the new session root.",
        "limits": "Uses an induced manager failure.",
        "reviewedHash": "6c86561d1c8644673dd403c6ee0d8c5ea5e455e501864ff5436e3f1a23c9d52e",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:37.673558972Z",
          "scope": "TestFailedTemporaryLaunchRemovesSessionRoot",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-relaunch-override",
        "name": "TestLaunchWorkspaceOverridesDoNotChangeCard",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/session_actions_internal_test.go",
        "line": 233,
        "endLine": 293,
        "proves": "The overridden directory survives stop and relaunch and the saved card retains its original workspace.",
        "limits": "Uses a single explicit ID.",
        "reviewedHash": "e2870b54ecd260896b9782a8394e38f4a5c45eaa20ac8eb8d091c43fdc74aff7",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.995025269Z",
          "scope": "TestLaunchWorkspaceOverridesDoNotChangeCard",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-relaunch-baseline",
        "name": "TestServiceRelaunchPreservesSavedWorkspaceIdentity",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 1495,
        "endLine": 1573,
        "proves": "Relaunch reuses saved paths after card changes.",
        "limits": "Exercises the retained default launch path.",
        "reviewedHash": "4c2f3e9342ceea29d5ff1d5ba392ce0d5c67b3e8b03c38bff8421942ca9bc9df",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:18.857033758Z",
          "scope": "TestServiceRelaunchPreservesSavedWorkspaceIdentity",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-launch-flow",
        "name": "TestFullSessionLaunchFlow",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/integration_test.go",
        "line": 23,
        "endLine": 151,
        "proves": "The confirmation reaches launch completion through coordinator and service.",
        "limits": "Local test fixture without a native provider.",
        "reviewedHash": "b490df4ae553b790329f3356c2a722efae0de8f9e8c9dd01b0b0163d12b0b91e",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.921893534Z",
          "scope": "TestFullSessionLaunchFlow",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-warning",
        "name": "TestSessionLauncher_PersistFailureStillOffersAttach",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/session_launcher_test.go",
        "line": 53,
        "endLine": 97,
        "proves": "A persistence warning still permits naming and attachment.",
        "limits": "Uses a supplied lifecycle result.",
        "reviewedHash": "d2a35f3d77fc4e93d3a29699a8d0fc9aa05f053d9649c768855ac0603b9c8c4b",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:25.143805611Z",
          "scope": "TestSessionLauncher_PersistFailureStillOffersAttach",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-workspace-service",
        "name": "TestListWorkspacesAuthorizationAndRegistry",
        "file": "internal/api/workspace_discovery_test.go",
        "proves": "Checks authorization before registry reads, same-name ordering by ID, projected fields, empty results and registry errors.",
        "limits": "Uses a service-level registry probe and local activation credentials.",
        "method": "automated",
        "integration": "project",
        "line": 26,
        "endLine": 114,
        "reviewedHash": "1eb2ad707fba0f9ea2e834c060e43bcfc66b24ba51842ed270ef37bf192295f0",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:19.009873408Z",
          "scope": "TestListWorkspacesAuthorizationAndRegistry",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-workspace-scopes",
        "name": "TestDashboardScopeOverHTTP",
        "file": "internal/mcp/dashboard_scope_test.go",
        "proves": "Two callers in separate dashboards get the same workspace list and only their scoped card and session data.",
        "limits": "Uses local fixtures for two dashboards and an unreferenced workspace.",
        "method": "automated",
        "integration": "project",
        "line": 22,
        "endLine": 228,
        "reviewedHash": "03413effbbfe574e7fc1d94717cb6a6a1641cfe6cc4541163ed1d518ac2e074e",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:15.089580079Z",
          "scope": "TestDashboardScopeOverHTTP",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-workspace-registration",
        "name": "TestMCPRegistersLifecycleAndDiscoveryTools",
        "file": "internal/mcp/server_test.go",
        "proves": "The tools list includes list_workspaces, which returns an empty items array for an empty fixture.",
        "limits": "Uses the discovery registrar fixture.",
        "method": "automated",
        "integration": "project",
        "line": 229,
        "endLine": 349,
        "reviewedHash": "52bac4fc72177e4e8a0510855a3b7e48253422c639ce4dc506012e43419d7ee8",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:15.157986395Z",
          "scope": "TestMCPRegistersLifecycleAndDiscoveryTools",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      },
      {
        "id": "workspace-feature-workspace-error",
        "name": "TestMCPWorkspaceDiscoveryReadFailure",
        "file": "internal/mcp/server_test.go",
        "proves": "The tool reports a registry read failure as an MCP error.",
        "limits": "Uses an injected error from the discovery registrar fixture.",
        "method": "automated",
        "integration": "project",
        "line": 1052,
        "endLine": 1085,
        "reviewedHash": "52bac4fc72177e4e8a0510855a3b7e48253422c639ce4dc506012e43419d7ee8",
        "result": {
          "status": "passed",
          "command": "go test -count=1 -json ./...",
          "at": "2026-11-23T21:51:15.72933502Z",
          "scope": "TestMCPWorkspaceDiscoveryReadFailure",
          "revision": "WORKTREE",
          "fingerprint": "307dedfa4db0c67f4f2a7d141769139101e0b6dfdc8b836c25be25024f0c7767"
        }
      }
    ],
    "captures": [
      {
        "id": "tui-mcp-list",
        "title": "List",
        "file": "appedia/evidence/current/mcp/list.png",
        "state": "Only the protected built-in and New are initially available.",
        "observation": "Only the protected built-in and New are initially available.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:17.210608058Z",
        "evidence": [
          "tui-mcp"
        ],
        "assertions": "Only the protected built-in and New are initially available."
      },
      {
        "id": "tui-mcp-new",
        "title": "New",
        "file": "appedia/evidence/current/mcp/new.png",
        "state": "New opens a blank MCP form with save/back instructions.",
        "observation": "New opens a blank MCP form with save/back instructions.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:17.210608058Z",
        "evidence": [
          "tui-mcp"
        ],
        "assertions": "New opens a blank MCP form with save/back instructions."
      },
      {
        "id": "tui-mcp-edit",
        "title": "Edit",
        "file": "appedia/evidence/current/mcp/edit.png",
        "state": "Existing entry loads its name and command in the editor.",
        "observation": "Existing entry loads its name and command in the editor.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:17.210608058Z",
        "evidence": [
          "tui-mcp"
        ],
        "assertions": "Existing entry loads its name and command in the editor."
      },
      {
        "id": "tui-mcp-edited-list",
        "title": "Edited list",
        "file": "appedia/evidence/current/mcp/edited-list.png",
        "state": "At 80×24 the renamed server, protected built-in, New and help are visible.",
        "observation": "At 80×24 the renamed server, protected built-in, New and help are visible.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          80,
          24
        ],
        "unit": "cells",
        "environment": "Linux, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:17.210608058Z",
        "evidence": [
          "tui-mcp"
        ],
        "assertions": "At 80×24 the renamed server, protected built-in, New and help are visible."
      },
      {
        "id": "tui-mcp-confirm",
        "title": "Confirm",
        "file": "appedia/evidence/current/mcp/confirm.png",
        "state": "Delete opens a named confirmation with confirm/cancel instructions at 80×24.",
        "observation": "Delete opens a named confirmation with confirm/cancel instructions at 80×24.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          80,
          24
        ],
        "unit": "cells",
        "environment": "Linux, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:17.210608058Z",
        "evidence": [
          "tui-mcp"
        ],
        "assertions": "Delete opens a named confirmation with confirm/cancel instructions at 80×24."
      },
      {
        "id": "tui-mcp-deleted",
        "title": "Deleted",
        "file": "appedia/evidence/current/mcp/deleted.png",
        "state": "Confirmed deletion removes the custom definition and retains built-in/New at 80×24.",
        "observation": "Confirmed deletion removes the custom definition and retains built-in/New at 80×24.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          80,
          24
        ],
        "unit": "cells",
        "environment": "Linux, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:17.210608058Z",
        "evidence": [
          "tui-mcp"
        ],
        "assertions": "Confirmed deletion removes the custom definition and retains built-in/New at 80×24."
      },
      {
        "id": "tui-codex-attached",
        "title": "Attached",
        "file": "appedia/evidence/current/codex/directory/attached.png",
        "state": "The real Codex input screen is visible after card launch and name confirmation; no agent prompt was entered.",
        "observation": "The real Codex input screen is visible after card launch and name confirmation; no agent prompt was entered.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/directory$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:26:35.969656818Z",
        "evidence": [
          "tui-codex"
        ],
        "assertions": "The real Codex input screen is visible after card launch and name confirmation; no agent prompt was entered."
      },
      {
        "id": "tui-codex-dashboard",
        "title": "Dashboard",
        "file": "appedia/evidence/current/codex/directory/dashboard.png",
        "state": "Ctrl+K restores the dashboard with both sessions, action hints and the selected session header.",
        "observation": "Ctrl+K restores the dashboard with both sessions, action hints and the selected session header.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/directory$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:26:35.969656818Z",
        "evidence": [
          "tui-codex"
        ],
        "assertions": "Ctrl+K restores the dashboard with both sessions, action hints and the selected session header."
      },
      {
        "id": "tui-codex-reattached",
        "title": "Reattached",
        "file": "appedia/evidence/current/codex/directory/reattached.png",
        "state": "Enter without pane navigation reattaches to the same live process; exactly two persisted sessions remain and the older session stays terminated.",
        "observation": "Enter without pane navigation reattaches to the same live process; exactly two persisted sessions remain and the older session stays terminated.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/directory$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:26:35.969656818Z",
        "evidence": [
          "tui-codex"
        ],
        "assertions": "Enter without pane navigation reattaches to the same live process; exactly two persisted sessions remain and the older session stays terminated."
      },
      {
        "id": "tui-codex-workspace-initial",
        "title": "Workspace initial",
        "file": "appedia/evidence/current/codex/directory/workspace-initial.png",
        "state": "The center displays Directory: for the selected idle Codex workspace.",
        "observation": "The center displays Directory: for the selected idle Codex workspace.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/directory$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:26:35.969656818Z",
        "evidence": [
          "tui-codex"
        ],
        "assertions": "The center displays Directory: for the selected idle Codex workspace."
      },
      {
        "id": "tui-codex-workspace-updated",
        "title": "Workspace updated",
        "file": "appedia/evidence/current/codex/directory/workspace-updated.png",
        "state": "The live workspace refresh displays the new file.",
        "observation": "The live workspace refresh displays the new file.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/directory$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:26:35.969656818Z",
        "evidence": [
          "tui-codex"
        ],
        "assertions": "The live workspace refresh displays the new file."
      },
      {
        "id": "tui-codex-minimum-size",
        "title": "Minimum size",
        "file": "appedia/evidence/current/codex/directory/minimum-size.png",
        "state": "At 60×20 cells the dashboard displays its minimum-size warning.",
        "observation": "At 60×20 cells the dashboard displays its minimum-size warning.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          60,
          20
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/directory$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:26:35.969656818Z",
        "evidence": [
          "tui-codex"
        ],
        "assertions": "At 60×20 cells the dashboard displays its minimum-size warning."
      },
      {
        "id": "tui-codex-stopped",
        "title": "Stopped",
        "file": "appedia/evidence/current/codex/directory/stopped.png",
        "state": "The selected session is stopped and its persisted workspace is unchanged.",
        "observation": "The selected session is stopped and its persisted workspace is unchanged.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/directory$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:26:35.969656818Z",
        "evidence": [
          "tui-codex"
        ],
        "assertions": "The selected session is stopped and its persisted workspace is unchanged."
      },
      {
        "id": "tui-codex-relaunched",
        "title": "Relaunched",
        "file": "appedia/evidence/current/codex/directory/relaunched.png",
        "state": "The same session and workspace have a new Codex process.",
        "observation": "The same session and workspace have a new Codex process.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/directory$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:26:35.969656818Z",
        "evidence": [
          "tui-codex"
        ],
        "assertions": "The same session and workspace have a new Codex process."
      },
      {
        "id": "tui-codex-report-after-relaunch",
        "title": "Report after relaunch",
        "file": "appedia/evidence/current/codex/directory/report-after-relaunch.png",
        "state": "The seeded report remains visible after the author session relaunches with Codex.",
        "observation": "The seeded report remains visible after the author session relaunches with Codex.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/directory$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:26:35.969656818Z",
        "evidence": [
          "tui-codex"
        ],
        "assertions": "The seeded report remains visible after the author session relaunches with Codex."
      },
      {
        "id": "tui-codex-restart-report",
        "title": "Restart report",
        "file": "appedia/evidence/current/codex/directory/restart-report.png",
        "state": "Kankuro reconstructs the selected session’s persisted report after restart.",
        "observation": "Kankuro reconstructs the selected session’s persisted report after restart.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/directory$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:26:35.969656818Z",
        "evidence": [
          "tui-codex"
        ],
        "assertions": "Kankuro reconstructs the selected session’s persisted report after restart."
      },
      {
        "id": "tui-codex-git-attached",
        "title": "Attached",
        "file": "appedia/evidence/current/codex/git/attached.png",
        "state": "The real Codex input screen is visible after card launch and name confirmation; no agent prompt was entered.",
        "observation": "The real Codex input screen is visible after card launch and name confirmation; no agent prompt was entered.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/git$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:11.45966887Z",
        "evidence": [
          "tui-codex-git"
        ],
        "assertions": "The real Codex input screen is visible after card launch and name confirmation; no agent prompt was entered."
      },
      {
        "id": "tui-codex-git-dashboard",
        "title": "Dashboard",
        "file": "appedia/evidence/current/codex/git/dashboard.png",
        "state": "Ctrl+K restores the dashboard with both sessions, action hints and the selected session header.",
        "observation": "Ctrl+K restores the dashboard with both sessions, action hints and the selected session header.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/git$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:11.45966887Z",
        "evidence": [
          "tui-codex-git"
        ],
        "assertions": "Ctrl+K restores the dashboard with both sessions, action hints and the selected session header."
      },
      {
        "id": "tui-codex-git-reattached",
        "title": "Reattached",
        "file": "appedia/evidence/current/codex/git/reattached.png",
        "state": "Enter without pane navigation reattaches to the same live process; exactly two persisted sessions remain and the older session stays terminated.",
        "observation": "Enter without pane navigation reattaches to the same live process; exactly two persisted sessions remain and the older session stays terminated.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/git$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:11.45966887Z",
        "evidence": [
          "tui-codex-git"
        ],
        "assertions": "Enter without pane navigation reattaches to the same live process; exactly two persisted sessions remain and the older session stays terminated."
      },
      {
        "id": "tui-codex-git-workspace-initial",
        "title": "Workspace initial",
        "file": "appedia/evidence/current/codex/git/workspace-initial.png",
        "state": "The center displays Working tree clean for the selected idle Codex workspace.",
        "observation": "The center displays Working tree clean for the selected idle Codex workspace.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/git$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:11.45966887Z",
        "evidence": [
          "tui-codex-git"
        ],
        "assertions": "The center displays Working tree clean for the selected idle Codex workspace."
      },
      {
        "id": "tui-codex-git-workspace-updated",
        "title": "Workspace updated",
        "file": "appedia/evidence/current/codex/git/workspace-updated.png",
        "state": "The live workspace refresh displays the new file.",
        "observation": "The live workspace refresh displays the new file.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/git$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:11.45966887Z",
        "evidence": [
          "tui-codex-git"
        ],
        "assertions": "The live workspace refresh displays the new file."
      },
      {
        "id": "tui-codex-git-minimum-size",
        "title": "Minimum size",
        "file": "appedia/evidence/current/codex/git/minimum-size.png",
        "state": "At 60×20 cells the dashboard displays its minimum-size warning.",
        "observation": "At 60×20 cells the dashboard displays its minimum-size warning.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          60,
          20
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/git$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:11.45966887Z",
        "evidence": [
          "tui-codex-git"
        ],
        "assertions": "At 60×20 cells the dashboard displays its minimum-size warning."
      },
      {
        "id": "tui-codex-git-stopped",
        "title": "Stopped",
        "file": "appedia/evidence/current/codex/git/stopped.png",
        "state": "The selected session is stopped and its persisted workspace is unchanged.",
        "observation": "The selected session is stopped and its persisted workspace is unchanged.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/git$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:11.45966887Z",
        "evidence": [
          "tui-codex-git"
        ],
        "assertions": "The selected session is stopped and its persisted workspace is unchanged."
      },
      {
        "id": "tui-codex-git-relaunched",
        "title": "Relaunched",
        "file": "appedia/evidence/current/codex/git/relaunched.png",
        "state": "The same session and workspace have a new Codex process.",
        "observation": "The same session and workspace have a new Codex process.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/git$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:11.45966887Z",
        "evidence": [
          "tui-codex-git"
        ],
        "assertions": "The same session and workspace have a new Codex process."
      },
      {
        "id": "tui-codex-git-report-after-relaunch",
        "title": "Report after relaunch",
        "file": "appedia/evidence/current/codex/git/report-after-relaunch.png",
        "state": "The seeded report remains visible after the author session relaunches with Codex.",
        "observation": "The seeded report remains visible after the author session relaunches with Codex.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/git$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:11.45966887Z",
        "evidence": [
          "tui-codex-git"
        ],
        "assertions": "The seeded report remains visible after the author session relaunches with Codex."
      },
      {
        "id": "tui-codex-git-restart-report",
        "title": "Restart report",
        "file": "appedia/evidence/current/codex/git/restart-report.png",
        "state": "Kankuro reconstructs the selected session’s persisted report after restart.",
        "observation": "Kankuro reconstructs the selected session’s persisted report after restart.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/git$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:27:11.45966887Z",
        "evidence": [
          "tui-codex-git"
        ],
        "assertions": "Kankuro reconstructs the selected session’s persisted report after restart."
      },
      {
        "id": "tui-luna-launched",
        "title": "Launched",
        "file": "tmp/appediff/tui-evidence/project-luna-60k/launched.png",
        "state": "Dashboard Enter launched a persisted native-agent process and Enter attached to the native UI.",
        "observation": "Dashboard Enter launched a persisted native-agent process and Enter attached to the native UI.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "./tests/e2e/run token-usage",
        "at": "2026-11-18T21:47:36.773848345Z",
        "evidence": [
          "tui-luna"
        ],
        "assertions": "Dashboard Enter launched a persisted native-agent process and Enter attached to the native UI."
      },
      {
        "id": "tui-luna-dashboard",
        "title": "Dashboard",
        "file": "tmp/appediff/tui-evidence/project-luna-60k/dashboard.png",
        "state": "Ctrl+K restored the dashboard; c opened settings and Escape returned using ordinary keyboard input.",
        "observation": "Ctrl+K restored the dashboard; c opened settings and Escape returned using ordinary keyboard input.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "./tests/e2e/run token-usage",
        "at": "2026-11-18T21:47:36.773848345Z",
        "evidence": [
          "tui-luna"
        ],
        "assertions": "Ctrl+K restored the dashboard; c opened settings and Escape returned using ordinary keyboard input."
      },
      {
        "id": "tui-luna-workspace-before-report",
        "title": "Workspace before report",
        "file": "tmp/appediff/tui-evidence/project-luna-60k/workspace-before-report.png",
        "state": "The selected Codex session initially displays its saved workspace directory.",
        "observation": "The selected Codex session initially displays its saved workspace directory.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "./tests/e2e/run token-usage",
        "at": "2026-11-18T21:47:36.773848345Z",
        "evidence": [
          "tui-luna"
        ],
        "assertions": "The selected Codex session initially displays its saved workspace directory."
      },
      {
        "id": "tui-luna-live-report",
        "title": "Live report",
        "file": "tmp/appediff/tui-evidence/project-luna-60k/live-report.png",
        "state": "A live Codex progress report replaces the selected session’s workspace listing.",
        "observation": "A live Codex progress report replaces the selected session’s workspace listing.",
        "limits": "One tui-test screenshot at this checkpoint. No image baseline comparison was performed.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "./tests/e2e/run token-usage",
        "at": "2026-11-18T21:47:36.773848345Z",
        "evidence": [
          "tui-luna"
        ],
        "assertions": "A live Codex progress report replaces the selected session’s workspace listing."
      },
      {
        "id": "tui-luna-input-overage",
        "title": "Dashboard checkpoint when the total input check failed",
        "file": "tmp/appediff/tui-evidence/project-luna-60k/failure.png",
        "state": "The dashboard still shows the authored progress report.",
        "observation": "The first progress report remains visible when the test ends on the measured input overage.",
        "limits": "The terminal capture shows the dashboard. The run record measures the overage and identifies where the test ended.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux, Codex CLI, tui-test Go binding",
        "command": "./tests/e2e/run token-usage",
        "at": "2026-11-18T21:47:36.773848345Z",
        "evidence": [
          "tui-luna"
        ]
      },
      {
        "id": "native-launched",
        "title": "OpenCode with OpenRouter Free Models Router (historical)",
        "file": "tests/e2e/artifacts/native-agent/launched.png",
        "state": "Native UI reached through Kankuro launch/attach",
        "observation": "The native model label shows Free Models Router OpenRouter.",
        "limits": "Launch/model/attachment checkpoint only. Retained from the project Appedia as historical evidence.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux/amd64; tui-test Go v0.1.1; Noto Sans Symbols2 fallback; OpenCode 2.2.6; openrouter/openrouter/free; isolated HOME/config/workspace/service",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestNativeAgentJourney$' -count=1 -timeout=10m -v",
        "at": "2026-11-09T22:38:25.698926+00:00",
        "assertions": [
          "Dashboard Enter launched a persisted native-agent process and Enter attached to the native UI."
        ],
        "evidence": [
          "native-agent-terminal-journey"
        ]
      },
      {
        "id": "native-failure",
        "title": "Native journey failure at register (historical)",
        "file": "tests/e2e/artifacts/native-agent/failure.png",
        "state": "Terminal at failed register checkpoint",
        "observation": "The inspected image shows the explicit registration task and Free Models Router after a completed response, with no registration result or executed shell command visible. No quota error is shown.",
        "limits": "Failure evidence; subsequent checkpoints are not established. Retained from the project Appedia as historical evidence.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Linux/amd64; tui-test Go v0.1.1; Noto Sans Symbols2 fallback; OpenCode 2.2.6; openrouter/openrouter/free; isolated HOME/config/workspace/service",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestNativeAgentJourney$' -count=1 -timeout=10m -v",
        "at": "2026-11-09T22:42:25.738775+00:00",
        "assertions": [
          "The recorded run failed at register."
        ],
        "evidence": [
          "native-agent-terminal-journey"
        ]
      },
      {
        "id": "mcp-list",
        "title": "MCP configuration · list (historical)",
        "file": "tests/e2e/artifacts/mcp-manager/list.png",
        "state": "list",
        "observation": "Only the protected built-in and New are initially available.",
        "limits": "Project-owned asserted checkpoint; no pixel baseline comparison. Visual inspection documents the reached state, not all terminal sizes or themes. Retained from the project Appedia as historical evidence.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "tui-test Go v0.1.1; real local Kankuro executable; isolated temporary config/HOME/XDG; no agent/model",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
        "at": "2026-11-11T22:34:15.147987645Z",
        "assertions": [
          "Only the protected built-in and New are initially available."
        ],
        "evidence": [
          "mcp-mcp-terminal-journey"
        ]
      },
      {
        "id": "mcp-new",
        "title": "MCP configuration · new (historical)",
        "file": "tests/e2e/artifacts/mcp-manager/new.png",
        "state": "new",
        "observation": "New opens a blank MCP form with save/back instructions.",
        "limits": "Project-owned asserted checkpoint; no pixel baseline comparison. Visual inspection documents the reached state, not all terminal sizes or themes. Retained from the project Appedia as historical evidence.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "tui-test Go v0.1.1; real local Kankuro executable; isolated temporary config/HOME/XDG; no agent/model",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
        "at": "2026-11-11T22:34:15.291908278Z",
        "assertions": [
          "New opens a blank MCP form with save/back instructions."
        ],
        "evidence": [
          "mcp-mcp-terminal-journey"
        ]
      },
      {
        "id": "mcp-edit",
        "title": "MCP configuration · edit (historical)",
        "file": "tests/e2e/artifacts/mcp-manager/edit.png",
        "state": "edit",
        "observation": "Existing entry loads its name and command in the editor.",
        "limits": "Project-owned asserted checkpoint; no pixel baseline comparison. Visual inspection documents the reached state, not all terminal sizes or themes. Retained from the project Appedia as historical evidence.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "tui-test Go v0.1.1; real local Kankuro executable; isolated temporary config/HOME/XDG; no agent/model",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
        "at": "2026-11-11T22:34:15.43776469Z",
        "assertions": [
          "Existing entry loads its name and command in the editor."
        ],
        "evidence": [
          "mcp-mcp-terminal-journey"
        ]
      },
      {
        "id": "mcp-edited-list",
        "title": "MCP configuration · edited-list (historical)",
        "file": "tests/e2e/artifacts/mcp-manager/edited-list.png",
        "state": "edited-list",
        "observation": "At 80×24 the renamed server, protected built-in, New and help are visible.",
        "limits": "Project-owned asserted checkpoint; no pixel baseline comparison. Visual inspection documents the reached state, not all terminal sizes or themes. Retained from the project Appedia as historical evidence.",
        "size": [
          80,
          24
        ],
        "unit": "cells",
        "environment": "tui-test Go v0.1.1; real local Kankuro executable; isolated temporary config/HOME/XDG; no agent/model",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
        "at": "2026-11-11T22:34:15.507906977Z",
        "assertions": [
          "At 80×24 the renamed server, protected built-in, New and help are visible."
        ],
        "evidence": [
          "mcp-mcp-terminal-journey"
        ]
      },
      {
        "id": "mcp-confirm",
        "title": "MCP configuration · confirm (historical)",
        "file": "tests/e2e/artifacts/mcp-manager/confirm.png",
        "state": "confirm",
        "observation": "Delete opens a named confirmation with confirm/cancel instructions at 80×24.",
        "limits": "Project-owned asserted checkpoint; no pixel baseline comparison. Visual inspection documents the reached state, not all terminal sizes or themes. Retained from the project Appedia as historical evidence.",
        "size": [
          80,
          24
        ],
        "unit": "cells",
        "environment": "tui-test Go v0.1.1; real local Kankuro executable; isolated temporary config/HOME/XDG; no agent/model",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
        "at": "2026-11-11T22:34:15.571669443Z",
        "assertions": [
          "Delete opens a named confirmation with confirm/cancel instructions at 80×24."
        ],
        "evidence": [
          "mcp-mcp-terminal-journey"
        ]
      },
      {
        "id": "mcp-deleted",
        "title": "MCP configuration · deleted (historical)",
        "file": "tests/e2e/artifacts/mcp-manager/deleted.png",
        "state": "deleted",
        "observation": "Confirmed deletion removes the custom definition and retains built-in/New at 80×24.",
        "limits": "Project-owned asserted checkpoint; no pixel baseline comparison. Visual inspection documents the reached state, not all terminal sizes or themes. Retained from the project Appedia as historical evidence.",
        "size": [
          80,
          24
        ],
        "unit": "cells",
        "environment": "tui-test Go v0.1.1; real local Kankuro executable; isolated temporary config/HOME/XDG; no agent/model",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
        "at": "2026-11-11T22:34:15.789639732Z",
        "assertions": [
          "Confirmed deletion removes the custom definition and retains built-in/New at 80×24."
        ],
        "evidence": [
          "mcp-mcp-terminal-journey"
        ]
      },
      {
        "id": "launch-focus-attached",
        "title": "Launch focus · attached (historical)",
        "file": "tests/e2e/artifacts/launch-focus/attached.png",
        "state": "attached",
        "observation": "The isolated Codex startup input screen is visible.",
        "limits": "Startup/attachment checkpoint only; no prompt or baseline-image comparison. Retained from the project Appedia as historical evidence.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "tui-test Go; isolated HOME/config/workspace; Codex startup only",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestNativeLaunchFocusReturn$' -count=1 -timeout=3m -v",
        "assertions": [
          "The real Codex input screen is visible after card launch and name confirmation; no agent prompt was entered."
        ],
        "evidence": [
          "launch-focus-terminal"
        ]
      },
      {
        "id": "launch-focus-dashboard",
        "title": "Launch focus · dashboard",
        "file": "appedia/evidence/current/codex/dashboard.png",
        "state": "Ctrl+K restores the dashboard with both sessions, action hints and the selected session header.",
        "observation": "Ctrl+K restores the dashboard with both sessions, action hints and the selected session header.",
        "limits": "Current Linux launch-focus checkpoint. No model prompt or pixel-baseline comparison.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "tui-test Go; isolated HOME/config/workspace; Codex startup only",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestNativeLaunchFocusReturn$' -count=1 -timeout=3m -v",
        "assertions": "Ctrl+K restores the dashboard with both sessions, action hints and the selected session header.",
        "evidence": [
          "tui-launch-focus"
        ],
        "at": "2026-11-19T21:27:15.182353193Z"
      },
      {
        "id": "launch-focus-reattached",
        "title": "Launch focus · reattached (historical)",
        "file": "tests/e2e/artifacts/launch-focus/reattached.png",
        "state": "reattached",
        "observation": "The isolated Codex startup input screen is visible.",
        "limits": "Startup/attachment checkpoint only; no prompt or baseline-image comparison. Retained from the project Appedia as historical evidence.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "tui-test Go; isolated HOME/config/workspace; Codex startup only",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestNativeLaunchFocusReturn$' -count=1 -timeout=3m -v",
        "assertions": [
          "Enter without pane navigation reattaches to the same live process; exactly two persisted sessions remain and the older session stays terminated."
        ],
        "evidence": [
          "launch-focus-terminal"
        ]
      },
      {
        "id": "tui-header-minimum",
        "title": "Session header at 80×24",
        "unit": "cells",
        "environment": "Linux, Codex CLI and tui-test",
        "limits": "Minimum supported terminal size. Artwork is omitted and metadata is clipped to the narrow center.",
        "evidence": [
          "tui-codex"
        ],
        "file": "appedia/evidence/current/codex/directory/supported-minimum.png",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestCodexOfflineWorkspaceAndReports$/directory$' -count=1 -timeout=3m -v",
        "at": "2026-11-19T21:26:35.969656818Z",
        "size": [
          80,
          24
        ],
        "state": "At 80×24 the selected session header and directory body remain visible.",
        "observation": "At 80×24 the selected session header and directory body remain visible.",
        "assertions": "At 80×24 the selected session header and directory body remain visible."
      },
      {
        "id": "launch-focus-dashboard-historical",
        "title": "Launch focus · dashboard (historical terminal preview)",
        "file": "tests/e2e/artifacts/launch-focus/dashboard.png",
        "state": "dashboard",
        "observation": "The created session and older session are visible with session-pane actions after detach.",
        "limits": "Startup/attachment checkpoint only; no prompt or baseline-image comparison. Retained from the project Appedia as historical evidence.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "tui-test Go; isolated HOME/config/workspace; Codex startup only",
        "command": "go test -tags=e2e ./tests/e2e -run '^TestNativeLaunchFocusReturn$' -count=1 -timeout=3m -v",
        "assertions": [
          "Ctrl+K restores the dashboard with both sessions visible and session-pane action hints, without pressing Tab."
        ],
        "evidence": [
          "launch-focus-terminal"
        ]
      },
      {
        "id": "e2e-codex-attached",
        "title": "Codex after card launch",
        "file": "tmp/appediff/runs/codex/attached.png",
        "state": "Idle Codex after naming and attaching to a fresh session.",
        "observation": "The Codex screen shows an empty input prompt and GPT-6-Sol high in the footer.",
        "limits": "The footer value is a visual observation. The test asserts the saved card value and the CLI reaching its idle prompt.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Project tui-test Go harness and Codex CLI 0.171.0 with an unreachable loopback provider",
        "command": "go test -tags=e2e -count=1 -timeout=8m -json ./tests/e2e",
        "assertions": [
          "The real Codex input screen is visible after card launch and name confirmation; no agent prompt was entered."
        ],
        "evidence": [
          "e2e-launch-focus"
        ],
        "at": "2026-11-21T21:33:33.089Z"
      },
      {
        "id": "e2e-codex-reattached",
        "title": "Codex after reattaching",
        "file": "tmp/appediff/runs/codex/reattached.png",
        "state": "The original agent process is active after detach and reattach.",
        "observation": "The same workspace path and GPT-6-Sol high appear after reattaching.",
        "limits": "The fixture verifies session identity. The screenshot does not establish the effort flag.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Project tui-test Go harness and Codex CLI 0.171.0 with an unreachable loopback provider",
        "command": "go test -tags=e2e -count=1 -timeout=8m -json ./tests/e2e",
        "assertions": [
          "Enter without pane navigation reattaches to the same live process; exactly two persisted sessions remain and the older session stays terminated."
        ],
        "evidence": [
          "e2e-launch-focus"
        ],
        "at": "2026-11-21T21:33:33.375Z"
      },
      {
        "id": "e2e-card-effort-reopened",
        "title": "Saved reasoning effort after reopening",
        "file": "tmp/appediff/runs/codex/card-effort-reopened.png",
        "state": "Edit ID Card after saving and reopening the launch card.",
        "observation": "The Reasoning effort row reads high, directly below the Model row.",
        "limits": "The checkpoint covers saving and reopening an existing card at 120×40.",
        "size": [
          120,
          40
        ],
        "unit": "cells",
        "environment": "Kankuro terminal driven by the project tui-test Go harness",
        "command": "go test -tags=e2e -count=1 -timeout=8m -json ./tests/e2e",
        "assertions": [
          "Reopening the saved card shows high on the Reasoning effort row."
        ],
        "evidence": [
          "e2e-card-effort"
        ],
        "at": "2026-11-21T21:33:32.383Z"
      }
    ]
  }
};
