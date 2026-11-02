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
  "snapshot": "Commit 60e5808, 2026-11-07; dashboard ownership, shared message delivery and automatic report wake-ups.",
  "pitch": "Organize agent configurations and sessions into independent dashboards. Share selected cards across dashboards while keeping each session, workspace identity and conversation binding in its creation dashboard.",
  "pitch2": "The dashboard-list TUI and authenticated loopback MCP endpoint share one scoped service. The session manager owns processes, credentials and workspace preparation; repositories persist registries, membership and checkpointed recovery. A report dispatcher feeds the existing operation runner and shared message delivery hooks. Configstartup locks and validates configuration before runtime begins.",
  "concepts": [
    [
      "Dashboard",
      "Persistent ID, name and description owning sessions and card creation.",
      "config.Dashboard"
    ],
    [
      "ID card",
      "Owned agent template; shared cards are available across dashboards.",
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
      "Trace a scoped launch through admission and process publication.",
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
      ],
      "configstartup": [
        580,
        220
      ]
    }
  },
  "packages": {
    "app": {
      "label": "UI coordinator",
      "sub": "commands and navigation",
      "role": "Owns the opened dashboard and view generation. Builds asynchronous scoped commands and rejects stale results before changing views.",
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
        "app.go": "tea.Model wrapper",
        "dashboard_list.go": "dashboard inventory, scope-aware callbacks and cascade confirmation",
        "scoped_refresh.go": "captured scope, generation and fingerprint-aware commands"
      },
      "notes": [
        "Commands capture dashboard ID, view generation and settings before running off the coordinator goroutine.",
        "Test-only adapters in export_test.go exercise the current scoped handlers."
      ],
      "rules": [],
      "key": [
        "Coordinator",
        "sessionActions",
        "refresher"
      ],
      "sources": {
        "refs": [
          "app.Coordinator.Init",
          "app.Coordinator.agentSwitchCallbacks",
          "app.Coordinator.cardEditorCallbacks",
          "app.Coordinator.conversationChoiceCallbacks",
          "app.Coordinator.dashboardCallbacks",
          "app.Coordinator.dashboardListCallbacks",
          "app.Coordinator.guardKey",
          "app.Coordinator.handleAgentSwitched",
          "app.Coordinator.handleConfigFingerprint",
          "app.Coordinator.handleConversationChoiceRequested",
          "app.Coordinator.handleDataLoaded",
          "app.Coordinator.handlePreviewRequest",
          "app.Coordinator.handleSessionMessages",
          "app.Coordinator.handleSessionsRefreshed",
          "app.Coordinator.handleSpecialMessages",
          "app.Coordinator.launchCard",
          "app.Coordinator.launchIDE",
          "app.Coordinator.launchTerminal",
          "app.Coordinator.maybeRefreshConfig",
          "app.Coordinator.maybeSubscribeAutoSelectedPreview",
          "app.Coordinator.mcpCallbacks",
          "app.Coordinator.navigateTo",
          "app.Coordinator.navigationCommand",
          "app.Coordinator.refreshDashboardList",
          "app.Coordinator.refreshDataSeeded",
          "app.Coordinator.refreshSessions",
          "app.Coordinator.requestAttach",
          "app.Coordinator.requestRelaunch",
          "app.Coordinator.requestSwitch",
          "app.Coordinator.resizeAll",
          "app.Coordinator.sessionLauncherCallbacks",
          "app.Coordinator.settingsCallbacks",
          "app.New",
          "app.NewCoordinator",
          "app.cardDeletionPreflightCommand",
          "app.confirmCardDeletionCommand",
          "app.convertViewType",
          "app.dashboardDeletionStatus",
          "app.dashboardDeletionSummary",
          "app.dashboardListCommand",
          "app.previewSub.unsubscribe",
          "app.refresher.scopedConfigIfChanged",
          "app.refresher.scopedFull",
          "app.refresher.scopedSessions",
          "app.retryCardDeletionCommand",
          "app.scopedAttachCommand",
          "app.sessionActions.launch",
          "app.sessionActions.launchIDE",
          "app.sessionActions.launchTerminal",
          "app.sessionActions.launchWorkspace",
          "app.sessionActions.relaunch",
          "app.sessionActions.switchAgent"
        ]
      }
    },
    "views": {
      "label": "Terminal views",
      "sub": "forms, lists and modals",
      "role": "Renders the initial dashboard list, scoped card/session screens, shared-card editor and explicit cross-dashboard deletion confirmations.",
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
        "source_type.go": "per-registry labels",
        "dashboard_list.go": "list-first navigation, forms and delete/retry controls"
      },
      "notes": [
        "The entry screen is the dashboard list. Escape cancels a dialog or stays on the list; q quits there.",
        "Delete/Backspace resumes an incomplete dashboard deletion; Enter inspects its status."
      ],
      "rules": [],
      "key": [
        "DashboardList",
        "Dashboard",
        "CardEditor"
      ],
      "sources": {
        "refs": [
          "views.CardEditor.Card",
          "views.CardEditor.Reset",
          "views.CardEditor.SetCard",
          "views.CardEditor.SetDashboardContext",
          "views.CardEditor.SetData",
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
          "views.ConversationChoiceView.SelectedBinding",
          "views.ConversationChoiceView.Update",
          "views.Dashboard.SetDashboardName",
          "views.Dashboard.SetDashboardScope",
          "views.Dashboard.SetSessions",
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
          "views.Dashboard.renderPreviewPane",
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
          "views.workspaceItem.Title"
        ]
      }
    },
    "messages": {
      "label": "UI messages",
      "sub": "scoped async results",
      "role": "Carries dashboard IDs and view generations between commands and the coordinator, alongside existing full and partial snapshots.",
      "files": {
        "messages.go": "all message types, DataReceiver, PreviewReceiver"
      },
      "notes": [
        "DataLoadedMsg is always a full snapshot; partial updates get their own type."
      ],
      "rules": [],
      "key": [],
      "sources": {
        "refs": [
          "messages.ScopedActionFinishedMsg"
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
          "session.Manager.SetNameChecked",
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
          "session.validateConversationChoice"
        ]
      }
    },
    "components": {
      "label": "UI components",
      "sub": "lists, text and previews",
      "role": "Reusable terminal widgets; deleted-card sessions remain visible with an explanatory description.",
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
          "components.SessionPreview.handleStateCSI",
          "components.TruncateHead",
          "components.cardDelegate.Render",
          "components.renderPreviewField",
          "components.renderedBlockWidth",
          "components.wrapPreviewWords"
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
        "factory.go": "construction helpers",
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
          "driver.writePromptPhases"
        ]
      }
    },
    "ascii": {
      "label": "ASCII rendering",
      "sub": "terminal canvas",
      "role": "Supplies the canvas and drawing primitives used by puppet rendering.",
      "files": {
        "canvas.go": "drawing surface",
        "shape.go": "Shape interface",
        "options.go": "config options"
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
      "notes": [],
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
          "config.validateNewCardPending"
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
      "role": "Shared dashboard policy and operation execution. Ordinary sends and report wake-ups share prepared delivery hooks and the bounded runner; narrowly scoped report authority grants no general worker control.",
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
        "notifications.go": "owned dispatcher, retry scheduling and report authority"
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
          "api.Service.CreateCardInDashboard",
          "api.Service.CreateDashboard",
          "api.Service.GetDashboardDeletion",
          "api.Service.GetMessage",
          "api.Service.GetSession",
          "api.Service.LaunchInDashboard",
          "api.Service.ListDashboards",
          "api.Service.ListSessions",
          "api.Service.PreflightCardDeletion",
          "api.Service.PreflightDashboardDeletion",
          "api.Service.ReadDashboardData",
          "api.Service.RelaunchInDashboard",
          "api.Service.RemoveInDashboard",
          "api.Service.RenameInDashboard",
          "api.Service.ResolveCaller",
          "api.Service.RetryCardDeletion",
          "api.Service.RetryDashboardDeletion",
          "api.Service.StopInDashboard",
          "api.Service.SwitchInDashboard",
          "api.Service.UpdateCardInDashboard",
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
          "api.sessionUsesCard"
        ]
      }
    },
    "mcp": {
      "label": "MCP transport",
      "sub": "loopback tools",
      "role": "Authenticated loopback tools. Report input and reply wrapper stay compatible; tool descriptions explain wake-ups and additive delivery status.",
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
        "refs": []
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
        "list-errors"
      ]
    },
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
          "app.Coordinator"
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
          "",
          "app.New"
        ],
        [
          "boot",
          "http",
          "io",
          "Start loopback MCP",
          "Authenticated tools use the same service as the UI.",
          "mcp.Start"
        ],
        [
          "boot",
          "mgr",
          "call",
          "Set MCP endpoint",
          "Set before any session activation.",
          "session.Manager.SetMCPEndpoint"
        ],
        [
          "boot",
          "ui",
          "cmd",
          "Initialize entry view and refresh",
          "Bubble Tea starts from the dashboard list.",
          "app.Coordinator.Init"
        ]
      ],
      "sources": {
        "refs": [
          "app.New",
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
          "root.run"
        ]
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
      "summary": "Human and MCP requests enter the same dashboard-scoped activation path.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "c",
          "app.Coordinator"
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
          "Launch selected card",
          "",
          "app.Coordinator.launchCard"
        ],
        [
          "c",
          "svc",
          "call",
          "LaunchInDashboard(capturedID, cardID)",
          "",
          "api.Service.LaunchInDashboard"
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
          "mgr",
          "call",
          "LaunchInDashboardWithIDContext",
          "",
          "session.Manager.LaunchInDashboardWithIDContext"
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
          "",
          "session.Manager.startSession"
        ],
        [
          "mgr",
          "disk",
          "io",
          "publish + save with immutable dashboard_id",
          "Revalidates card/dashboard under admission; rejects deletion races.",
          "session.Session.data"
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
          "session.Manager.SetNameChecked",
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
          "views.SessionLauncherCallbacks"
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
          "app.Coordinator"
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
          "app.attachExec"
        ]
      ],
      "steps": [
        [
          "user",
          "ui",
          "key",
          "Attach selected session",
          "",
          "app.Coordinator.requestAttach"
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
          "app.Coordinator.handleSessionMessages"
        ],
        [
          "ui",
          "exec",
          "cmd",
          "Release UI terminal; run attach",
          "",
          "app.scopedAttachCommand"
        ],
        [
          "exec",
          "exec",
          "io",
          "Prepare raw terminal and output",
          "",
          "app.attachExec.Run"
        ],
        [
          "exec",
          "exec",
          "go",
          "Pump process output",
          "",
          "app.attachExec.pumpOutput"
        ],
        [
          "exec",
          "exec",
          "call",
          "Pump terminal input",
          "",
          "app.attachExec.pumpInput"
        ],
        [
          "exec",
          "exec",
          "call",
          "Forward through locked driver input",
          "",
          "app.attachExec.forwardInput"
        ],
        [
          "exec",
          "exec",
          "call",
          "Recognize detach key",
          "Restore terminal state on return.",
          "app.detectKittyDetach"
        ],
        [
          "exec",
          "ui",
          "msg",
          "AttachFinishedMsg",
          "Only matching scope/generation may update the view.",
          "app.Coordinator.handleSessionMessages"
        ]
      ],
      "sources": {
        "refs": [
          "app.scopedAttachCommand",
          "session.driverInputWriter"
        ]
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
      "summary": "Press o in Sessions to start the configured IDE in the selected session’s actual workspace through shared command execution. Commands capture dashboard identity, generation and command settings.",
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
          "call",
          "Check captured membership",
          "Validate the session against the dashboard captured by the command.",
          "api.Service.CheckSessionInDashboard"
        ],
        [
          "act",
          "ide",
          "call",
          "Use saved WorkingDir",
          "Resolve the existing session workspace after authorization.",
          "app.sessionActions.launchWorkspaceCommand"
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
          "Apply scoped completion",
          "Reject stale dashboard/generation results; display current errors without navigation.",
          "app.Coordinator.handleSessionMessages"
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
      "summary": "A single seeded refresh captures dashboard and generation; global capability registries remain global.",
      "participants": [
        [
          "loop",
          "tea"
        ],
        [
          "c",
          "app.Coordinator"
        ],
        [
          "refresh",
          "app.refresher"
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
          "loop",
          "c",
          "msg",
          "Refresh tick",
          "tick requests sessions/config refresh.",
          "app.Coordinator.handleSpecialMessages"
        ],
        [
          "c",
          "c",
          "call",
          "Seed baseline",
          "Seed refresh fingerprint baseline.",
          "app.Coordinator.refreshDataSeeded"
        ],
        [
          "c",
          "refresh",
          "call",
          "scopedFull",
          "No seeded flag or unseeded production branch. Failed reads preserve previous data; successful empty reads clear it. Scoped refresh; fingerprint envelope.",
          "app.refresher.scopedFull"
        ],
        [
          "refresh",
          "svc",
          "call",
          "Read scope",
          "ReadDashboardData(capturedID).",
          "api.Service.ReadDashboardData"
        ],
        [
          "refresh",
          "c",
          "msg",
          "Apply if current",
          "Ignores results from another dashboard or an obsolete generation. handleDataLoaded checks scope/generation.",
          "app.Coordinator.handleDataLoaded"
        ],
        [
          "c",
          "dash",
          "call",
          "Apply scoped sessions",
          "handleSessionsRefreshed applies partial data.",
          "app.Coordinator.handleSessionsRefreshed"
        ],
        [
          "c",
          "c",
          "call",
          "Owned preview",
          "preview subscription only for owned selection.",
          "app.Coordinator.handlePreviewRequest"
        ]
      ],
      "sources": {
        "refs": [
          "app.Coordinator.handleConfigFingerprint",
          "app.Coordinator.handleDataLoaded",
          "app.Coordinator.handlePreviewRequest",
          "app.Coordinator.handleSessionsRefreshed",
          "app.Coordinator.maybeRefreshConfig",
          "app.Coordinator.maybeSubscribeAutoSelectedPreview",
          "app.Coordinator.refreshDataSeeded",
          "app.Coordinator.refreshSessions",
          "app.previewSub.unsubscribe",
          "app.refresher.scopedConfigIfChanged",
          "app.refresher.scopedFull",
          "app.refresher.scopedSessions",
          "components.SessionPreview.handleStateCSI",
          "config.registryFiles",
          "messages.AgentSwitchRequestedMsg",
          "messages.AgentSwitchedMsg",
          "messages.AttachFinishedMsg",
          "messages.CardDeletionPreparedMsg",
          "messages.CardSavedMsg",
          "messages.ConversationChoiceRequestedMsg",
          "messages.DashboardDeletionPreparedMsg",
          "messages.DashboardDeletionStatusMsg",
          "messages.DashboardListLoadedMsg",
          "messages.DashboardListRequestedMsg",
          "messages.DashboardMutationFinishedMsg",
          "messages.DashboardOpenRequestedMsg",
          "messages.MCPSavedMsg",
          "messages.ScopedActionFinishedMsg",
          "messages.ScopedAttachReadyMsg",
          "messages.ScopedNavigateMsg",
          "messages.ScopedSessionNamedMsg",
          "messages.SessionAttachRequestedMsg",
          "messages.SessionLaunchCompletedMsg",
          "messages.SettingsSavedMsg",
          "session.Data.restore",
          "session.loadSessions",
          "session.saveSession",
          "views.Dashboard.SetSessions"
        ]
      }
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
          "scopedFull(scope, generation, previewID)",
          "Load global capabilities and scoped cards/sessions, then seed the fingerprint baseline.",
          "app.refresher.scopedFull"
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
          "Reject obsolete dashboard/generation data; nil failed fields preserve previous values.",
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
          "app.Coordinator.handleSessionMessages"
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
          "views.CardEditor.Reset",
          "views.CardEditor.SetCard",
          "views.CardEditor.fieldBlocks",
          "views.editorKeyIdentity"
        ],
        "files": [
          "internal/ui/views/card_editor.go"
        ]
      }
    },
    "savecard": {
      "title": "Configure exclusive or shared cards",
      "summary": "Only the immutable owner edits a card; consumers can use shared cards.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "edit",
          "views.CardEditor"
        ],
        [
          "c",
          "app.Coordinator"
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
          "edit",
          "key",
          "Shared toggle; owner/read-only context",
          "",
          "views.CardEditor.SetDashboardContext"
        ],
        [
          "edit",
          "edit",
          "call",
          "Card overlays loaded configuration",
          "Preserves fields outside the form.",
          "views.CardEditor.Card"
        ],
        [
          "edit",
          "c",
          "cmd",
          "OnSave captures expected card revision",
          "",
          "app.Coordinator.cardEditorCallbacks"
        ],
        [
          "c",
          "svc",
          "call",
          "CreateCardInDashboard for new cards",
          "",
          "api.Service.CreateCardInDashboard"
        ],
        [
          "c",
          "svc",
          "call",
          "UpdateCardInDashboard for saved cards",
          "",
          "api.Service.UpdateCardInDashboard"
        ],
        [
          "svc",
          "mgr",
          "call",
          "WithCardReferencesAndSnapshot",
          "Serializes owner, revision and foreign-reference checks with activation admission.",
          "session.Manager.WithCardReferencesAndSnapshot"
        ],
        [
          "svc",
          "repo",
          "call",
          "UpdateIDCard validates tombstones",
          "",
          "config.FileRepository.UpdateIDCard"
        ]
      ],
      "sources": {
        "refs": [
          "api.Service.AttachInDashboard",
          "api.Service.CheckSessionInDashboard",
          "api.Service.CreateCardInDashboard",
          "api.Service.CreateDashboard",
          "api.Service.LaunchInDashboard",
          "api.Service.ListDashboards",
          "api.Service.ReadDashboardData",
          "api.Service.RelaunchInDashboard",
          "api.Service.RemoveInDashboard",
          "api.Service.RenameInDashboard",
          "api.Service.StopInDashboard",
          "api.Service.SwitchInDashboard",
          "api.Service.UpdateCardInDashboard",
          "api.Service.UpdateDashboardDetails",
          "api.Service.availableDashboardCard",
          "api.Service.launchDashboardWithID",
          "api.Service.listCardTombstones",
          "api.Service.loadMembershipInventory",
          "api.hasForeignCardReferences",
          "api.membershipInventory.activeDashboard",
          "api.membershipInventory.allowsCard",
          "api.membershipInventory.allowsSession",
          "api.membershipInventory.hasTombstone",
          "api.newMembershipInventory",
          "api.sessionInDashboard",
          "api.sessionUsesCard",
          "app.Coordinator.agentSwitchCallbacks",
          "app.Coordinator.cardEditorCallbacks",
          "app.Coordinator.conversationChoiceCallbacks",
          "app.Coordinator.dashboardCallbacks",
          "app.Coordinator.launchCard",
          "app.Coordinator.launchIDE",
          "app.Coordinator.launchTerminal",
          "app.Coordinator.mcpCallbacks",
          "app.Coordinator.navigationCommand",
          "app.Coordinator.requestRelaunch",
          "app.Coordinator.requestSwitch",
          "app.Coordinator.sessionLauncherCallbacks",
          "app.Coordinator.settingsCallbacks",
          "components.Filter",
          "config.FileRepository.CreateIDCard",
          "config.FileRepository.LoadIDCardsFile",
          "config.FileRepository.UpdateIDCard",
          "config.FileRepository.WriteIDCardsFile",
          "config.validateCardMembershipFormat",
          "views.CardEditor.Card",
          "views.CardEditor.Reset",
          "views.CardEditor.SetCard",
          "views.CardEditor.SetDashboardContext",
          "views.CardEditor.SetData",
          "views.CardEditor.SetSettings",
          "views.CardEditor.Update",
          "views.CardEditor.fieldBlocks",
          "views.CardEditor.handleSpace",
          "views.CardEditor.renderDashboardControlLine",
          "views.CardEditor.renderSharedLine",
          "views.CardEditor.renderTmpWorkspaceLine",
          "views.CardEditor.visibleFocusOrder",
          "views.CardEditorCallbacks",
          "views.focusTargetTable"
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
          "views.CardEditor.SetCard",
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
      "summary": "Press t in the Sessions pane to launch an independently configured terminal in the actual session workspace. Commands capture dashboard identity, generation and command settings.",
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
          "Launch command in session.WorkingDir",
          "Use the saved session WorkingDir after the membership check.",
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
          "ScopedActionFinishedMsg{Err}",
          "Reports process startup, not the external application’s eventual exit.",
          "app.Coordinator.handleSpecialMessages"
        ],
        [
          "coord",
          "coord",
          "self",
          "Apply scoped completion",
          "Reject stale dashboard/generation results; display current errors without navigation.",
          "app.Coordinator.handleSessionMessages"
        ]
      ],
      "sources": {
        "refs": [
          "messages.ScopedActionFinishedMsg",
          "views.DashboardCallbacks"
        ],
        "files": [
          "internal/app/terminal_launch_test.go",
          "internal/ui/views/dashboard_test.go",
          "internal/session/ide_launcher.go"
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
          "Check scope and session",
          "Membership must still be valid when the command executes.",
          "api.Service.CheckSessionInDashboard"
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
          "app.Coordinator.handleSessionMessages"
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
          "messages.ScopedActionFinishedMsg"
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
          "app.Coordinator.requestRelaunch"
        ],
        [
          "ui",
          "snap",
          "call",
          "Read per-card saved bindings",
          "Bindings belong to the existing session; even one binding requires explicit choice.",
          "session.Manager.GetSession"
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
          "RelaunchInDashboard / SwitchInDashboard",
          "Pass the captured dashboard, existing session and explicit choice.",
          "api.Service.SwitchInDashboard"
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
          "app.Coordinator.handleConversationChoiceRequested",
          "app.Coordinator.requestSwitch",
          "session.savedTemporaryRoot",
          "session.validateConversationChoice",
          "views.AgentSwitchView.SetData",
          "views.AgentSwitchView.handleKey",
          "views.ConversationChoiceView.RefreshContext",
          "views.ConversationChoiceView.SelectedBinding",
          "views.ConversationChoiceView.Update"
        ]
      }
    },
    "control": {
      "title": "Authenticated discovery and session control",
      "summary": "An authenticated activation with deployed Dashboard Control manages available cards and sessions inside its creation dashboard.",
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
          "Launch in caller dashboard",
          "Caller scope comes from immutable session membership, never human navigation.",
          "api.Service.launchDashboardWithID"
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
          "app.Coordinator"
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
          "PreflightCardDeletion inventories references",
          "",
          "api.Service.PreflightCardDeletion"
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
          "ConfirmCardDeletion(token, consent)",
          "Rechecks owner, revision, expiry and reference-set digest before effects.",
          "api.Service.ConfirmCardDeletion"
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
      "title": "Render owned sessions and available cards",
      "summary": "The opened dashboard shows its owned sessions and exclusive/shared available cards, with native conversation choice and bounded metadata.",
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
          "components.SessionList.refreshItems",
          "components.cardDelegate.Render",
          "components.renderPreviewField",
          "components.renderedBlockWidth",
          "components.wrapPreviewWords",
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
          "app.Coordinator"
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
          "app.Coordinator.dashboardListCallbacks"
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
      "summary": "Enter opens its card/session view; Esc returns without stopping sessions.",
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
          "app.Coordinator"
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
          "app.Coordinator.handleSpecialMessages"
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
          "ReadDashboardData(openedID)",
          "",
          "api.Service.ReadDashboardData"
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
          "Drops the preview subscription and invalidates old scoped results.",
          "app.Coordinator.handleSpecialMessages"
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
          "app.Coordinator.Init",
          "app.Coordinator.dashboardListCallbacks",
          "app.Coordinator.guardKey",
          "app.Coordinator.handleAgentSwitched",
          "app.Coordinator.handleConversationChoiceRequested",
          "app.Coordinator.handleSessionMessages",
          "app.Coordinator.handleSpecialMessages",
          "app.Coordinator.navigateTo",
          "app.Coordinator.refreshDashboardList",
          "app.Coordinator.resizeAll",
          "app.NewCoordinator",
          "app.cardDeletionPreflightCommand",
          "app.confirmCardDeletionCommand",
          "app.convertViewType",
          "app.dashboardDeletionStatus",
          "app.dashboardDeletionSummary",
          "app.dashboardListCommand",
          "app.retryCardDeletionCommand",
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
          "api.DashboardData",
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
          "session.Manager.SetNameChecked",
          "session.Manager.ShutdownStarted",
          "session.Manager.listLocked",
          "session.NewManager",
          "session.Session.data",
          "session.Session.snapshot"
        ]
      }
    },
    "manual-session": {
      "title": "Attach, rename, stop or remove an owned session",
      "summary": "Human actions are checked against the dashboard captured when invoked.",
      "participants": [
        [
          "u",
          "user"
        ],
        [
          "c",
          "app.Coordinator"
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
          "app.Coordinator.requestAttach"
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
          "app.Coordinator.handleSessionMessages"
        ],
        [
          "launch",
          "svc",
          "call",
          "RenameInDashboard after async name submit",
          "",
          "api.Service.RenameInDashboard"
        ],
        [
          "c",
          "svc",
          "call",
          "StopInDashboard validates owned target",
          "",
          "api.Service.StopInDashboard"
        ],
        [
          "c",
          "svc",
          "call",
          "RemoveInDashboard retains removal guards",
          "",
          "api.Service.RemoveInDashboard"
        ]
      ],
      "sources": {
        "refs": [
          "app.Coordinator.requestAttach",
          "app.scopedAttachCommand",
          "app.sessionActions.launch",
          "app.sessionActions.launchIDE",
          "app.sessionActions.launchTerminal",
          "app.sessionActions.launchWorkspace",
          "app.sessionActions.relaunch",
          "app.sessionActions.switchAgent",
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
          "app.Coordinator"
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
          "api.Service.ReadDashboardData"
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
          "app.Coordinator.dashboardListCallbacks",
          "app.Coordinator.refreshDashboardList",
          "app.cardDeletionPreflightCommand",
          "app.confirmCardDeletionCommand",
          "app.dashboardDeletionStatus",
          "app.dashboardDeletionSummary",
          "app.dashboardListCommand",
          "app.retryCardDeletionCommand",
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
          "app.Coordinator"
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
          "app.Coordinator.dashboardListCallbacks"
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
    }
  },
  "rules": [
    {
      "n": "1",
      "title": "No tea.Cmd closure touches coordinator state",
      "ref": "app.Coordinator",
      "why": "A command's closure runs on its own goroutine. Reading or writing c.views, c.preview, c.currentView, c.err or c.settings there is a concurrent access that can crash the app. Snapshot what you need before returning the closure.",
      "check": "Capture IDs, generation and settings by value. Run the coordinator command-closure race regression; external test adapters call the real scoped route.",
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
          "app.Coordinator.handleSessionMessages"
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
      "ref": "app.Coordinator.handleDataLoaded",
      "why": "Capture dashboard and generation before commands. Late refresh, attachment and workspace-command results must not retarget the current view.",
      "check": "Scoped refresh, attachment and coordinator command-race tests.",
      "flows": [
        "navigation",
        "tick",
        "attach"
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
      "a": "Declare it in messages.go and route it through the coordinator. Partial data gets its own type; async view results carry dashboard ID and generation.",
      "refs": [
        "app.Coordinator.handleSpecialMessages",
        "app.Coordinator.handleSessionMessages",
        "messages.ScopedActionFinishedMsg"
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
      "a": "Dashboard panes use dashboard.go; the entry list, forms and deletion/status controls use dashboard_list.go. Keep the legend visible and Escape safe on return.",
      "refs": [
        "views.Dashboard.handleKey",
        "views.Dashboard.footerText",
        "views.DashboardList.Update",
        "views.DashboardList.renderPage"
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
      "a": "Capture settings and scope, check session membership, then launch the configured command in the saved WorkingDir. Completion stays tied to its originating view.",
      "refs": [
        "app.sessionActions.launchWorkspace",
        "session.LaunchWorkspaceCommand"
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
    "The shared-delivery suite exercises queue rejection, journal rollback, cancellation and stable replay through ordinary sends and notification entry points. Notification reconciliation must never turn uncertain bytes into an automatic retry."
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
  ]
};
