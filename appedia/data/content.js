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
  "snapshot": "Final worktree on UX commit 474ae3a: MCP registry management, launch/naming focus and session-epoch refresh protection; five suggested dependency updates validated. Ordinary tests and two startup/configuration E2E journeys passed. Current ordinary-suite statement coverage is collected separately; older model-driven native evidence remains historical.",
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
      "role": "Coordinates screens, snapshots and asynchronous commands. Launch completion focuses the created session; session epochs discard refresh snapshots that predate a launch or rename.",
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
          "app.sessionActions.switchAgent",
          "app.Coordinator.renameDashboardSession"
        ]
      }
    },
    "views": {
      "label": "Terminal views",
      "sub": "forms, lists and modals",
      "role": "Owns terminal screens and their interaction state. The dashboard preserves launch focus through refreshed naming, filters and session-list rebuilds, and hosts the existing-session rename dialog.",
      "files": {
        "dashboard.go": "Three panes, keys, search, launch-focus reconciliation and existing-session rename modal with pending-save/error handling",
        "card_editor.go": "card fields, one key switch, focused sub workspace clearing, source shortcuts, skills filter",
        "mcp_form.go": "Create/edit form; overlays visible fields to preserve loaded metadata",
        "source_manager.go": "source directories per registry",
        "settings.go": "Independent IDE/terminal command inputs, focus, save/cancel and bounded rendering",
        "session_launcher.go": "progress while a session starts",
        "agent_switch.go": "pick a card for a terminated session",
        "workspace_picker_modal.go": "single/multiple selection, provisional checks, missing IDs",
        "puppet_selector_modal.go": "modal: choose a puppet",
        "agentsmd_editor_modal.go": "modal: AGENTS.md content or path",
        "constants.go": "MinTerminalWidth/Height, CoordinatorRows",
        "source_type.go": "per-registry labels",
        "dashboard_list.go": "list-first navigation, forms and delete/retry controls",
        "mcp_manager.go": "Global MCP list, stable selection, protected built-in handling and scoped delete confirmation"
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
          "views.workspaceItem.Title",
          "views.NewDashboard",
          "views.dashboardModalInputWidth",
          "views.Dashboard.CompleteRename",
          "views.Dashboard.updateRename",
          "views.Dashboard.closeRename"
        ]
      }
    },
    "messages": {
      "label": "UI messages",
      "sub": "scoped async results",
      "role": "Defines scoped launch, naming and refresh events; refresh messages carry a session epoch so pre-mutation snapshots can be rejected.",
      "files": {
        "messages.go": "all message types, DataReceiver, PreviewReceiver"
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
        "Selected-only mode changes visible items, not the full item collection or selection map. Normalize the cursor after registry and selection changes.",
        "SourceList retains source-manager defaults while providing configurable labels and bounded MCP-list rendering."
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
        "launch-focus",
        "attach",
        "ide",
        "terminal",
        "workspace-command-errors",
        "rename-edit",
        "rename-save",
        "rename-layout"
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
        "mcp-open",
        "mcp",
        "mcp-delete",
        "mcp-stale",
        "mcp-card",
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
      "summary": "How views get their size, and what happens below 80×24. Rename input follows modal width; long save errors are sanitized and truncated to keep save/cancel controls visible at 80×24.",
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
      "summary": "Human and MCP requests share dashboard-scoped admission and process publication; the human launch result hands selection to the launch-focus flow.",
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
        ],
        [
          "mgr",
          "c",
          "msg",
          "SessionLaunchCompletedMsg(session ID)",
          "A usable launch advances the session epoch and requests focus for the exact ID.",
          "app.Coordinator.handleSessionMessages"
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
          "app.Coordinator.handleDataLoaded",
          "app.Coordinator.handleSessionMessages",
          "app.Coordinator.handleSessionsRefreshed",
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
          "app.Coordinator.hasStaleSessionSnapshot",
          "messages.SessionsRefreshedMsg"
        ],
        "files": [
          "internal/app/launch_focus_internal_test.go",
          "internal/app/inline_rename_refresh_internal_test.go",
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
      "summary": "Press o in Sessions to start the configured IDE in the selected session’s actual workspace through shared command execution. Completion is a ScopedActionFinishedMsg carrying captured dashboard identity and generation.",
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
          "ScopedActionFinishedMsg{Err}",
          "The shared workspace command returns captured dashboard identity and generation with its error.",
          "app.sessionActions.launchWorkspace"
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
          "views.Dashboard.SetSessions",
          "app.Coordinator.hasStaleSessionSnapshot",
          "app.refresher.scopedConfigIfChanged",
          "app.refresher.scopedSessions"
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
          "scopedFull(scope, generation, epoch, previewID)",
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
          "Reject obsolete dashboard/generation and session-epoch snapshots before applying any data; nil failed fields preserve prior values.",
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
      "summary": "The safe way to change screens: prepare the target view in the callback, navigate from a message. The e shortcut still edits cards in the cards pane; in the sessions pane it opens the separate rename dialog.",
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
          "app.Coordinator"
        ],
        [
          "f",
          "views.MCPForm"
        ],
        [
          "r",
          "config.FileRepository"
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
          "app.Coordinator.matchesMCPManagerScope"
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
          "Command captures repository, scope and form request.",
          "app.Coordinator.mcpCallbacks"
        ],
        [
          "r",
          "r",
          "call",
          "Persist validated global definition",
          "Built-in/shadow protection remains authoritative.",
          "config.FileRepository.UpsertMCP"
        ],
        [
          "c",
          "m",
          "msg",
          "Return to caller and refresh data",
          "Manager return path avoids previous-view bouncing.",
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
      "summary": "The opened dashboard shows its owned sessions and exclusive/shared available cards, with native conversation choice and bounded metadata. Session-pane help includes e=Rename, and its modal owns input until save or cancel.",
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
          "views.Dashboard.renderPreviewPane",
          "views.Dashboard.footerText",
          "views.Dashboard.View"
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
      "title": "Attach, name a launch, stop or remove an owned session",
      "summary": "Manual session actions use shared dashboard policy. Launch-time naming can attach; existing-session rename has a separate scoped result and is traced in the rename flows.",
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
          "Name a newly launched session",
          "Launcher naming still attaches after success. Existing-session e rename uses DashboardSessionRenamedMsg and only refreshes.",
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
          "app.Coordinator"
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
          "app.Coordinator.dashboardCallbacks"
        ],
        [
          "coord",
          "svc",
          "cmd",
          "renameDashboardSession",
          "Execute RenameInDashboard using captured values.",
          "app.Coordinator.renameDashboardSession"
        ],
        [
          "svc",
          "manager",
          "call",
          "RenameInDashboard → SetNameChecked",
          "Existing membership policy runs under manager admission.",
          "api.Service.RenameInDashboard"
        ],
        [
          "manager",
          "disk",
          "io",
          "SetNameChecked: persist session",
          "Reject lifecycle conflicts; update memory before saving, so save failure is not rollback.",
          "session.Manager.SetNameChecked"
        ],
        [
          "coord",
          "coord",
          "msg",
          "handleSessionMessages(result)",
          "DashboardSessionRenamedMsg must match dashboard, generation and current view. A successful inline rename also advances the session epoch so older refreshes cannot restore the old name.",
          "app.Coordinator.handleSessionMessages"
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
          "app.Coordinator.refreshDataSeeded"
        ]
      ],
      "sources": {
        "refs": [
          "messages.DashboardSessionRenamedMsg"
        ],
        "files": [
          "internal/app/coordinator_test.go",
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
          "app.Coordinator"
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
          "app.Coordinator.dashboardCallbacks"
        ],
        [
          "c",
          "m",
          "call",
          "Open manager; assign visit scope",
          "Navigation remains on the coordinator goroutine.",
          "app.Coordinator.navigateTo"
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
          "app.Coordinator.initViews",
          "app.convertViewType"
        ],
        "files": [
          "feature-mcpedit.md",
          "internal/ui/views/mcp_manager_test.go",
          "internal/app/navigation_test.go"
        ]
      }
    },
    "mcp-delete": {
      "title": "Confirm deletion of a global MCP",
      "summary": "Delete opens a named modal; confirmation invokes repository policy and refreshes consumers.",
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
          "app.Coordinator"
        ],
        [
          "a",
          "app.sessionActions"
        ],
        [
          "r",
          "config.FileRepository"
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
          "app.Coordinator.matchesMCPManagerScope"
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
          "app.sessionActions.deleteMCPFromManager"
        ],
        [
          "a",
          "r",
          "call",
          "DeleteMCP(id)",
          "Rejects referenced cards, built-in and missing IDs.",
          "config.FileRepository.DeleteMCP"
        ],
        [
          "a",
          "c",
          "msg",
          "Deletion result",
          "Success triggers global refresh; current errors stay visible.",
          "app.Coordinator.handleSpecialMessages"
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
          "app.Coordinator"
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
          "app.Coordinator.matchesMCPManagerScope"
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
          "app.Coordinator.handleSpecialMessages"
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
          "internal/app/mcp_manager_gaps_test.go",
          "internal/app/mcp_manager_review_test.go"
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
          "app.Coordinator"
        ],
        [
          "a",
          "app.sessionActions"
        ],
        [
          "r",
          "config.FileRepository"
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
          "Coordinator delegates repository work.",
          "app.Coordinator.handleSpecialMessages"
        ],
        [
          "c",
          "a",
          "cmd",
          "deleteMCP",
          "Existing card reference policy is reused.",
          "app.sessionActions.deleteMCP"
        ],
        [
          "a",
          "r",
          "call",
          "DeleteMCP",
          "Referenced saved cards still prevent deletion.",
          "config.FileRepository.DeleteMCP"
        ],
        [
          "a",
          "c",
          "msg",
          "Centralize completion and refresh",
          "Preserve editor deselection, refresh the registry globally.",
          "app.Coordinator.handleSpecialMessages"
        ]
      ]
    },
    "launch-focus": {
      "title": "Keep the created session selected through naming and detach",
      "summary": "Launch/name completion and fresh snapshots maintain one selected ID and preview; old epochs cannot displace it.",
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
          "mgr",
          "session.Manager"
        ],
        [
          "launcher",
          "views.SessionLauncher"
        ],
        [
          "ref",
          "app.refresher"
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
          "components.SessionPreview"
        ]
      ],
      "steps": [
        [
          "mgr",
          "c",
          "msg",
          "SessionLaunchCompletedMsg(session ID)",
          "A usable launch advances the session epoch and requests focus for the exact ID.",
          "app.Coordinator.handleSessionMessages"
        ],
        [
          "c",
          "dash",
          "call",
          "Focus the new session",
          "If its snapshot is not loaded yet, the one-shot focus waits for session data.",
          "app.Coordinator.handleSessionMessages"
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
          "app.Coordinator.handleSessionMessages"
        ],
        [
          "ref",
          "c",
          "msg",
          "Delayed old-epoch refresh",
          "Reject snapshots started before launch or rename, even within the same dashboard generation.",
          "app.Coordinator.handleSessionsRefreshed"
        ],
        [
          "ref",
          "dash",
          "msg",
          "Current session snapshot",
          "Select by ID, clear search only when it hides the target, and sync preview.",
          "app.Coordinator.handleDataLoaded"
        ],
        [
          "u",
          "c",
          "key",
          "Ctrl+K to detach",
          "The detach refresh retains the selected ID; later refreshes do not steal focus again.",
          "app.Coordinator.handleSessionMessages"
        ],
        [
          "dash",
          "list",
          "call",
          "Rebuild session list and preview by selected ID",
          "Session-list ordering changes do not displace the created session.",
          "views.Dashboard.SetSessions"
        ],
        [
          "dash",
          "preview",
          "call",
          "Sync preview with focused session",
          "Session action hints and preview agree with the selected session.",
          "views.Dashboard.syncPreviewSelection"
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
          "app.Coordinator.handleDataLoaded",
          "app.Coordinator.handleSessionMessages",
          "app.Coordinator.handleSessionsRefreshed",
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
          "app.Coordinator.hasStaleSessionSnapshot",
          "messages.SessionsRefreshedMsg"
        ],
        "files": [
          "internal/app/launch_focus_internal_test.go",
          "internal/app/inline_rename_refresh_internal_test.go",
          "internal/ui/components/session_list_test.go",
          "tests/e2e/launch_focus_test.go"
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
          "internal/app/mcp_manager_review_test.go",
          "tests/e2e/mcp_manager_test.go"
        ]
      }
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
      "a": "Declare it in messages.go and route it through the coordinator. Partial data gets its own type; async view results carry dashboard ID and generation. Separate dashboard rename completion from launch naming because their success actions differ.",
      "refs": [
        "app.Coordinator.handleSpecialMessages",
        "app.Coordinator.handleSessionMessages",
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
      "a": "views/constants.go and app/size_guard.go.",
      "refs": [
        "app.Coordinator.tooSmall"
      ],
      "flow": "resize"
    },
    {
      "q": "Change refresh behaviour",
      "a": "refresh.go for the commands, handlers.go for applying results, and config.Fingerprint. Capture session epochs with snapshots and reject old epochs before data or fingerprints are applied.",
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
      "a": "For existing-session e rename, start at Dashboard.handleEdit/updateRename, the captured coordinator callback and its dedicated completion message. Reuse RenameInDashboard and SetNameChecked; refresh on either result and preserve process identity. Launch naming keeps its separate attach flow. For operation snapshots, keep inspectionFromSnapshot as the single conversion. Launch naming retains focus by ID and confirmed name; inline rename invalidates older session refreshes.",
      "refs": [
        "views.Dashboard.handleEdit",
        "views.Dashboard.updateRename",
        "app.Coordinator.renameDashboardSession",
        "messages.DashboardSessionRenamedMsg",
        "api.Service.RenameInDashboard",
        "session.Manager.SetNameChecked",
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
        "app.Coordinator.mcpCallbacks",
        "config.FileRepository.DeleteMCP"
      ],
      "flow": "mcp-open"
    },
    {
      "q": "Change launch selection or stale-refresh protection",
      "a": "The coordinator advances session epochs on launch and rename. Dashboard focus waits for the exact ID/name, clears only a hiding session search, and is consumed once. SessionList preserves selected ID across reordering.",
      "refs": [
        "app.Coordinator.hasStaleSessionSnapshot",
        "views.Dashboard.FocusSession",
        "views.Dashboard.FocusSessionAfterRename",
        "components.SessionList.refreshItems"
      ],
      "flow": "launch"
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
    "Validation retained one intermittent TestPreviewImmediateUpdate race-run failure; ten focused repetitions and a full app race rerun passed. This is not reported as a clean initial broad race run."
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
    "scope": "Project-wide functional assessment retained and reassessed against current source; MCP management and launch-focus added with fresh ordinary and E2E evidence. Current ordinary-suite statement coverage is recorded separately from these functional assessments. The older model-driven native registration failure remains explicitly historical.",
    "features": [
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
                "app.Coordinator.dashboardListCallbacks",
                "api.Service.CreateDashboard",
                "config.FileRepository.UpdateDashboardDetails"
              ]
            },
            "suggestion": "Select a dashboard, edit with e, save and reorder/refresh inventory; assert name/description, immutable selected ID and scoped completion.",
            "priority": "Next",
            "reason": "Name sorting changes can move the cursor to a different dashboard even when saved IDs remain valid."
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
                "app.Coordinator.dashboardListCallbacks",
                "api.Service.GetDashboardDeletion",
                "views.DashboardList.View"
              ]
            }
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
            "limit": "Service/repository/manager fixture with cat processes; no keyboard or MCP transport assertions."
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
            "limit": "Injected workspace-copy gate; does not simulate every registry mutation."
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
            "limit": "Controlled copier and service entry points; no confirmation-modal rendering."
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
            }
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
            "summary": "Checks foreign preview rejection, stale refresh/navigation/attach isolation, current cards/sessions and restoration of the original view instance.",
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
            "limit": "Drives coordinator messages directly with real local cat sessions; not an end-to-end terminal journey."
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
            "limit": "Counting repository fixture; does not establish every filesystem timestamp boundary."
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
            "summary": "Seeds every card field with nonzero data, edits Name, and compares the entire returned card with the expected overlay.",
            "status": "tested",
            "flows": [
              "editcard",
              "savecard"
            ],
            "evidence": [
              "card-editor-preserves-unmanaged-fields"
            ],
            "limit": "Editor model round trip; repository persistence and read-only consumer policy are assessed elsewhere."
          },
          {
            "id": "control-toggle",
            "title": "Round-trip Dashboard Control through Space",
            "dimension": "interaction",
            "summary": "Focuses the control row, toggles off/on and checks Card().DashboardControl.",
            "status": "tested",
            "flows": [
              "savecard"
            ],
            "evidence": [
              "dashboard-control-card-editor-toggle-round-trip"
            ],
            "limit": "Does not prove deployed activation authority or MCP permission changes."
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
            "limit": "Rendered text and model focus; no screenshot, native terminal or image baseline."
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
            "reason": "SetSelectedOnly coverage alone cannot detect a broken keyboard binding."
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
                "views.CardEditor.Card"
              ]
            }
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
            "limit": "In-memory editor; filesystem copies and registry writes are outside this test."
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
            }
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
                "views.CardEditor.SetData",
                "views.WorkspacePickerModal.SetWorkspaces",
                "views.WorkspacePickerModal.rememberWorkspace",
                "views.WorkspacePickerModal.filterWorkspaces",
                "views.WorkspacePickerModal.refreshItems"
              ]
            }
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
            "id": "preview-broadcast",
            "title": "Preserve skills when queued preview output arrives",
            "dimension": "behavior",
            "summary": "Asserts a loaded builtin skill selection survives an injected preview/output message while the editor is open.",
            "status": "tested",
            "flows": [
              "preview",
              "mutation"
            ],
            "evidence": [
              "partial-preview-broadcast-does-not-wipe-card-editor"
            ],
            "limit": "Deliberately re-arms a subscription to model late messages; does not prove normal editor preview subscription behavior."
          },
          {
            "id": "sync-warning",
            "title": "Warn and refresh after a saved source fails to sync",
            "dimension": "behavior",
            "summary": "Checks emitted sync error, refresh request and durable saved source path despite failed scanning.",
            "status": "tested",
            "flows": [
              "source",
              "mutation"
            ],
            "evidence": [
              "save-simple-source-surfaces-sync-failure"
            ],
            "limit": "One missing skill-source path through callbacks; no full source-manager UI journey."
          },
          {
            "id": "workspace-sources",
            "title": "Visit workspace sources and return to the unsaved card",
            "dimension": "interaction",
            "summary": "Repository callbacks preserve workspace-source identity and reject malformed registry mutations. Search reviewed: internal/ui/views/{card_editor*,workspace_picker*,source_manager*}_test.go and internal/app/{navigation,integration,source_save,source_sync_warning}_test.go; Ctrl+E/OnOpenWorkspaces call sites.",
            "status": "partial",
            "flows": [
              "workspace-sources"
            ],
            "evidence": [
              "workspace-source-i-d-assigned-and-preserved-on-edit",
              "save-workspace-source-aborts-on-malformed-registry",
              "delete-workspace-source-aborts-on-malformed-registry"
            ],
            "limit": "Searched editor/navigation/source tests: no direct assertion of Ctrl+E from both Workspace/Sub Workspaces, source edit, Esc return and preservation of the unsaved card was found.",
            "sources": {
              "refs": [
                "views.CardEditor.handleCtrlE",
                "app.Coordinator.cardEditorCallbacks",
                "app.Coordinator.navigateTo",
                "views.SourceManager.Update",
                "config.FileRepository.UpdateWorkspaceSources"
              ]
            },
            "suggestion": "Drive Ctrl+E from each workspace field, save a source, return with Esc and compare the unsaved card/focus.",
            "priority": "Next",
            "reason": "Cross-view refresh can overwrite unsaved editor state even when repository callbacks work."
          },
          {
            "id": "source-crud",
            "title": "Create, edit and delete source rows through keys",
            "dimension": "interaction",
            "summary": "Successful callback/reload transitions and synchronous/asynchronous deletion errors have concrete assertions.",
            "status": "tested",
            "flows": [
              "source",
              "mutation"
            ],
            "evidence": [
              "source-list-key-crud",
              "source-list-delete-error",
              "source-manager-delete-error"
            ],
            "limit": "Component/callback fixtures; full source persistence and unsaved-editor return are separate boundaries."
          },
          {
            "id": "destination-refresh",
            "title": "Apply source and card refreshes to actual views",
            "dimension": "behavior",
            "summary": "Nonempty persisted data is delivered to source lists and dashboard cards; updates and removals cannot be satisfied by repository-only assertions.",
            "status": "tested",
            "flows": [
              "mutation",
              "source"
            ],
            "evidence": [
              "source-refresh-destination",
              "card-refresh-destination"
            ],
            "limit": "In-process coordinator views; not every registry or external filesystem change."
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
            "id": "preview-sanitization",
            "title": "Preserve preview text while removing control effects",
            "dimension": "behavior",
            "summary": "Rendering retains concrete output and rejects OSC title/Kitty control payloads.",
            "status": "tested",
            "flows": [
              "preview"
            ],
            "evidence": [
              "preview-sanitization-contract"
            ],
            "limit": "Styled component output, not native-terminal appearance."
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
            "limit": "Failed save composition is now directly asserted. No single successful key→durable YAML→refreshed editor visibility→return-navigation journey is asserted.",
            "sources": {
              "refs": [
                "views.SettingsView.handleEditEnter",
                "app.Coordinator.settingsCallbacks",
                "config.FileRepository.SaveAppSettings",
                "config.AppSettings",
                "app.Coordinator.handleSpecialMessages"
              ]
            },
            "suggestion": "Add a successful coordinator save journey checking durable YAML, editor visibility refresh and dashboard return.",
            "priority": "First",
            "reason": "The reported error test never executes its error command; failed persistence must not look like a successful save."
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
            }
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
            }
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
            "status": "tested",
            "flows": [
              "mcp"
            ],
            "evidence": [
              "m-c-p-form-round-trips-every-field"
            ],
            "limit": "Callback payload only; does not connect to that server or persist the configuration."
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
            "limit": "No retained test drives a successful o command through the coordinator and asserts the actual IDE child cwd/argv. The real child test covers the common launcher, and the IDE callback is directly tested for missing configuration.",
            "sources": {
              "refs": [
                "views.Dashboard.handleKey",
                "views.Dashboard.handleLaunchSession",
                "app.Coordinator.launchIDE",
                "app.sessionActions.launchIDE",
                "api.Service.CheckSessionInDashboard"
              ]
            },
            "suggestion": "Drive o with an isolated helper executable configured as IDECommand and assert the child cwd/argv and scoped completion.",
            "priority": "Next",
            "reason": "Terminal and IDE settings are independent; shared-launcher coverage does not prove successful IDE wiring."
          },
          {
            "id": "terminal",
            "title": "Launch a terminal in the actual selected workspace",
            "dimension": "interaction",
            "summary": "A coordinator key journey launches a real helper child in the temporary session workspace and checks its arguments and successful scoped completion.",
            "status": "tested",
            "flows": [
              "terminal"
            ],
            "evidence": [
              "terminal-launch-through-dashboard",
              "dashboard_-launch-terminal-routing",
              "launch-workspace-command-executes-in-workspace"
            ],
            "limit": "The child is a test executable, not a graphical terminal emulator; only one configured command template is exercised here.",
            "sources": {
              "refs": [
                "views.Dashboard.handleKey",
                "views.Dashboard.handleLaunchSession",
                "app.Coordinator.dashboardCallbacks",
                "app.Coordinator.launchTerminal",
                "app.sessionActions.launchTerminal"
              ]
            }
          },
          {
            "id": "workspace-command-errors",
            "title": "Report workspace command configuration and spawn failures",
            "dimension": "behavior",
            "summary": "Tests check command/path failures and delivery of a spawn failure to the coordinator error banner.",
            "status": "tested",
            "flows": [
              "workspace-command-errors"
            ],
            "evidence": [
              "terminal-launch-reports-configuration-and-lookup-errors",
              "launch-workspace-command-errors"
            ],
            "limit": "Checks the listed API/callback cases and banner presence; not exact error copy or external app startup errors after exec succeeds.",
            "sources": {
              "refs": [
                "api.Service.CheckSessionInDashboard",
                "app.sessionActions.launchTerminal",
                "session.LaunchWorkspaceCommand",
                "messages.ScopedActionFinishedMsg",
                "app.Coordinator.handleSessionMessages"
              ]
            }
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
            "status": "tested",
            "flows": [
              "resume"
            ],
            "evidence": [
              "conversation-choice-requires-explicit-compatible-choice"
            ],
            "limit": "Picker callback boundary; native CLI resume is outside scope."
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
            "status": "tested",
            "flows": [
              "launch"
            ],
            "evidence": [
              "dashboard-launch-persisted"
            ],
            "limit": "Local process and application policy; native-agent behavior is assessed separately."
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
            "limit": "Existing-session rename has direct successful refresh/restart and cancellation assertions. A complete dashboard stop/remove journey and raw attachExec detach/terminal restoration remain outside the retained assertions.",
            "sources": {
              "refs": [
                "app.Coordinator.requestAttach",
                "api.Service.AttachInDashboard",
                "session.Manager.AttachChecked",
                "app.Coordinator.handleSessionMessages",
                "api.Service.RenameInDashboard"
              ]
            },
            "suggestion": "Drive stop/remove through the coordinator and assert removal effects; add a project-owned terminal detach/restoration journey.",
            "priority": "First",
            "reason": "Scoped asynchronous actions and raw attachment cross ownership and terminal-state boundaries."
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
              "session-attach-replay"
            ],
            "limit": "Single OpenCode 2.2.6/OpenRouter run; no other CLI/provider or image baseline comparison. Failure at register; later journey checkpoints remain unverified. Free Models Router may choose different backing models; responsiveness does not establish tool execution.",
            "sources": {
              "refs": [
                "app.Coordinator.launchCard",
                "api.Service.LaunchInDashboard",
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
                "app.Coordinator.requestRelaunch",
                "session.Manager.GetSession",
                "views.ConversationChoiceView.SetContext",
                "app.Coordinator.conversationChoiceCallbacks",
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
                "app.Coordinator.requestAttach",
                "api.Service.AttachInDashboard",
                "session.Manager.AttachChecked",
                "app.Coordinator.handleSessionMessages",
                "app.scopedAttachCommand"
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
            "app.Coordinator.renameDashboardSession"
          ],
          "files": [
            "internal/ui/views/dashboard.go",
            "internal/app/callbacks.go",
            "internal/app/coordinator.go"
          ]
        },
        "scenarios": [
          {
            "id": "draft-retry",
            "title": "Edit a stable draft and retry a failure",
            "dimension": "interaction",
            "status": "tested",
            "evidence": [
              "rename-draft"
            ],
            "limit": "Stub callback; no filesystem or full cancellation journey.",
            "flows": [
              "rename-edit"
            ],
            "summary": "Captured target, normalized Unicode input, draft survives refresh, duplicate suppression, retry after error."
          },
          {
            "id": "blank-name",
            "title": "Blank name falls back to session ID",
            "dimension": "behavior",
            "status": "tested",
            "evidence": [
              "rename-blank"
            ],
            "limit": "Does not prove empty sessions-pane behavior.",
            "flows": [
              "rename-save"
            ],
            "summary": "Blank input submits the session ID."
          },
          {
            "id": "stale-result",
            "title": "Reject stale generation and retain save error",
            "dimension": "behavior",
            "status": "tested",
            "evidence": [
              "rename-scoped"
            ],
            "limit": "Missing target fixture; success and stale dashboard ID are not directly asserted.",
            "flows": [
              "rename-save"
            ],
            "summary": "Real command error stays in dialog; stale generation leaves pending state intact."
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
            "summary": "Long error bounds at 80×24 and 120×40, visible save/cancel controls and scoped rename footer."
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
            "summary": "Existing search consumes typed t/e/s/t and Esc exits search."
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
            "status": "tested",
            "evidence": [
              "rename-launcher"
            ],
            "limit": "Unchanged launcher test does not assert dashboard success avoids attachment.",
            "flows": [
              "launch"
            ],
            "summary": "Existing launcher naming waits for result and permits retry."
          },
          {
            "id": "success-restart",
            "title": "Successful dashboard rename survives restart",
            "dimension": "behavior",
            "status": "tested",
            "evidence": [
              "rename-rename-restart"
            ],
            "limit": "Local cat process and ordinary coordinator events; no native-agent terminal rendering journey.",
            "flows": [
              "rename-save"
            ],
            "summary": "The coordinator saves through the real service, immediately refreshes the selected name without attachment/navigation, preserves the running process and reloads the name after restart."
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
            "summary": "Esc and both dashboard-scope changes discard the draft without invoking save; reopening restores the original name."
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
            "app.Coordinator.handleSpecialMessages"
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
            "summary": "Nested cancel navigation"
          },
          {
            "id": "edit",
            "title": "Edit metadata and create a clean New record",
            "dimension": "behavior",
            "status": "tested",
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
            "limit": "Isolated real repository; no restart persistence test.",
            "summary": "Edit metadata and create a clean New record"
          },
          {
            "id": "delete",
            "title": "Delete admission and delayed request recovery",
            "dimension": "interaction",
            "status": "tested",
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
            "limit": "Same dashboard only; delayed result must not change current view.",
            "summary": "Delete admission and delayed request recovery"
          },
          {
            "id": "policy",
            "title": "Protected and referenced definitions remain intact",
            "dimension": "behavior",
            "status": "tested",
            "flows": [
              "mcp-delete"
            ],
            "evidence": [
              "mcp-referenced",
              "mcp-protected",
              "mcp-repository-retained"
            ],
            "limit": "No injected filesystem write failure or foreign-dashboard reference integration case.",
            "summary": "Protected and referenced definitions remain intact"
          },
          {
            "id": "bounds",
            "title": "Long list and modal stay within bounds",
            "dimension": "layout",
            "status": "tested",
            "flows": [
              "mcp-open",
              "mcp-delete"
            ],
            "evidence": [
              "mcp-long-list"
            ],
            "limit": "80×22 component bounds; no actual application screenshot or position-baseline comparison.",
            "summary": "Long list and modal stay within bounds",
            "captures": [
              "mcp-edited-list",
              "mcp-confirm",
              "mcp-deleted"
            ]
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
            "summary": "Existing source-list navigation remains usable"
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
            "summary": "Cross-dashboard delayed command/result and write-error recovery boundaries now have direct Go assertions."
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
            "app.Coordinator.hasStaleSessionSnapshot"
          ],
          "files": [
            "internal/app/launch_focus_internal_test.go",
            "internal/app/inline_rename_refresh_internal_test.go",
            "tests/e2e/launch_focus_test.go"
          ]
        },
        "scenarios": [
          {
            "id": "completion",
            "title": "Select the created ID and align preview",
            "summary": "Select the created ID and align preview",
            "status": "tested",
            "dimension": "interaction",
            "evidence": [
              "focus-completion",
              "focus-search"
            ],
            "limit": "Coordinator/view fixtures; actual terminal checkpoint is separate.",
            "flows": [
              "launch-focus"
            ],
            "captures": []
          },
          {
            "id": "epochs",
            "title": "Reject old launch and inline-rename refreshes",
            "summary": "Reject old launch and inline-rename refreshes",
            "status": "tested",
            "dimension": "behavior",
            "evidence": [
              "focus-epochs",
              "rename-epochs"
            ],
            "limit": "Covers full/session/fingerprint delivery with deterministic events.",
            "flows": [
              "launch-focus"
            ],
            "captures": []
          },
          {
            "id": "scope",
            "title": "Keep warnings usable and ignore canceled/stale focus",
            "summary": "Keep warnings usable and ignore canceled/stale focus",
            "status": "tested",
            "dimension": "behavior",
            "evidence": [
              "focus-warning",
              "focus-stale",
              "focus-cancel"
            ],
            "limit": "Controlled error/scope messages; no physical filesystem outage.",
            "flows": [
              "launch-focus"
            ],
            "captures": []
          },
          {
            "id": "ordering",
            "title": "Keep selection stable when newer sessions arrive",
            "summary": "Keep selection stable when newer sessions arrive",
            "status": "tested",
            "dimension": "interaction",
            "evidence": [
              "focus-order"
            ],
            "limit": "List component identity assertion, no terminal key decoder.",
            "flows": [
              "launch-focus"
            ],
            "captures": []
          },
          {
            "id": "terminal",
            "title": "Launch, name, detach and reattach the same process",
            "summary": "Launch, name, detach and reattach the same process",
            "status": "tested",
            "dimension": "interaction",
            "evidence": [
              "launch-focus-terminal"
            ],
            "limit": "Codex startup only at 120×40; no model task or image baseline.",
            "flows": [
              "launch-focus"
            ],
            "captures": [
              "launch-focus-attached",
              "launch-focus-dashboard",
              "launch-focus-reattached"
            ]
          }
        ]
      }
    ],
    "evidence": [
      {
        "id": "fresh-initialization-and-partial-root",
        "name": "TestFreshInitializationAndPartialRoot",
        "method": "automated",
        "integration": "project",
        "file": "internal/configstartup/startup_test.go",
        "line": 10,
        "endLine": 31,
        "proves": "Checks that fresh startup validates and that a partial root fails with the missing-marker diagnostic.",
        "limits": "Calls configuration startup directly; does not start the full TUI/MCP composition root.",
        "reviewedHash": "23def849d1afd7463ecca53123cd3035b705f5cd4ef39e1b7a2fc5782fd55c31",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.715942183Z",
          "scope": "TestFreshInitializationAndPartialRoot; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "validate-layout-rejects-invalid-records-without-rewriting",
        "name": "TestValidateLayoutRejectsInvalidRecordsWithoutRewriting",
        "method": "automated",
        "integration": "project",
        "file": "internal/configstartup/validate_test.go",
        "line": 20,
        "endLine": 49,
        "proves": "Validates a known-good fixture before each corrupt session/operation/message/cursor/card record, then checks rejection and exact original bytes.",
        "limits": "Five explicit fixture cases; not arbitrary schema fuzzing.",
        "reviewedHash": "a79b37b8df4a3e2f94e517171046fa8d1f212cd9e52cb5075ac05f7c7789b9ca",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.727411172Z",
          "scope": "TestValidateLayoutRejectsInvalidRecordsWithoutRewriting; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "exclusive-lock",
        "name": "TestExclusiveLock",
        "method": "automated",
        "integration": "project",
        "file": "internal/configstartup/startup_test.go",
        "line": 42,
        "endLine": 65,
        "proves": "A second AcquireLock fails while held and succeeds after Close.",
        "limits": "Same-process lock contention on Linux; no cross-platform or subprocess assertion.",
        "reviewedHash": "23def849d1afd7463ecca53123cd3035b705f5cd4ef39e1b7a2fc5782fd55c31",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.716070833Z",
          "scope": "TestExclusiveLock; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "read-dashboard-data-scopes-cards-and-sessions",
        "name": "TestReadDashboardDataScopesCardsAndSessions",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/membership_internal_test.go",
        "line": 139,
        "endLine": 309,
        "proves": "Checks private/shared visibility, consumer edit denial, dashboard-local sessions, stale/foreign deletion confirmation, surviving foreign sessions and switching away from a deleted card.",
        "limits": "Service/repository/manager fixture with cat processes; no keyboard or MCP transport assertions.",
        "reviewedHash": "7788a680a0548d68f4a1d3ee5108c3ca6f219d435c34e55cd3ce6306b42e5cb2",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.914872882Z",
          "scope": "TestReadDashboardDataScopesCardsAndSessions; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "scoped-launch-revalidates-changed-card-before-publication",
        "name": "TestScopedLaunchRevalidatesChangedCardBeforePublication",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/membership_internal_test.go",
        "line": 403,
        "endLine": 460,
        "proves": "Blocks copying, changes sharing, then asserts no publication and release of the pending reservation.",
        "limits": "Injected workspace-copy gate; does not simulate every registry mutation.",
        "reviewedHash": "7788a680a0548d68f4a1d3ee5108c3ca6f219d435c34e55cd3ce6306b42e5cb2",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.932543752Z",
          "scope": "TestScopedLaunchRevalidatesChangedCardBeforePublication; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "confirmed-card-deletion-cancels-foreign-pending-launch",
        "name": "TestConfirmedCardDeletionCancelsForeignPendingLaunch",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/membership_internal_test.go",
        "line": 461,
        "endLine": 502,
        "proves": "Asserts foreign pending impact, required confirmation, cancelled launch and no published session.",
        "limits": "Controlled copier and service entry points; no confirmation-modal rendering.",
        "reviewedHash": "7788a680a0548d68f4a1d3ee5108c3ca6f219d435c34e55cd3ce6306b42e5cb2",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.938806011Z",
          "scope": "TestConfirmedCardDeletionCancelsForeignPendingLaunch; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "dashboard-cascade-resumes-after-restart",
        "name": "TestDashboardCascadeResumesAfterRestart",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/dashboard_deletion_internal_test.go",
        "line": 132,
        "endLine": 186,
        "proves": "Persists deleting intent, checks admission denial, rebuilds manager/service, removes owned session while preserving the foreign session and permits idempotent retry.",
        "limits": "In-process restart against real files; no OS kill/power-loss simulation.",
        "reviewedHash": "ba7b77503150f21a09e4c47a8818dc09064dfaf30c0fe1b292e713807c78af6a",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.77520059Z",
          "scope": "TestDashboardCascadeResumesAfterRestart; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "multi-dashboard-discards-stale-results-and-preserves-views",
        "name": "TestMultiDashboardDiscardsStaleResultsAndPreservesViews",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/multi_dashboard_internal_test.go",
        "line": 19,
        "endLine": 111,
        "proves": "Checks foreign preview rejection, stale refresh/navigation/attach isolation, current cards/sessions and restoration of the original view instance.",
        "limits": "Drives coordinator messages directly with real local cat sessions; not an end-to-end terminal journey.",
        "reviewedHash": "d946df2c92262a773d61a2e7670f54cfdfaf4dc7e2a780b6ebeb11ec3466b8a5",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.666897041Z",
          "scope": "TestMultiDashboardDiscardsStaleResultsAndPreservesViews; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "maybe-refresh-config-skips-unchanged-registry",
        "name": "TestMaybeRefreshConfigSkipsUnchangedRegistry",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/refresh_cadence_internal_test.go",
        "line": 63,
        "endLine": 94,
        "proves": "Counts ListIDCards calls and asserts an unchanged fingerprint produces no additional registry read.",
        "limits": "Counting repository fixture; does not establish every filesystem timestamp boundary.",
        "reviewedHash": "5949b61fea145c448014776b5d2b37e945c67497ad99f79d7be77c5515ba94ae",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.672500281Z",
          "scope": "TestMaybeRefreshConfigSkipsUnchangedRegistry; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "coordinator-too-small-screen-is-exact",
        "name": "TestCoordinatorTooSmallScreenIsExact",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/size_guard_test.go",
        "line": 26,
        "endLine": 46,
        "proves": "Sweeps widths 1–85 and heights 1–27; checks exact notice rectangle below 80×24 and notice removal at supported sizes.",
        "limits": "String/cell layout assertions; no emulator font or image comparison.",
        "reviewedHash": "fcb14dea1373d1c821a7ef0039bfa0ec28484fb5005502e0546068b3366028b3",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:05.684013886Z",
          "scope": "TestCoordinatorTooSmallScreenIsExact; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "coordinator-drops-keys-when-too-small",
        "name": "TestCoordinatorDropsKeysWhenTooSmall",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/size_guard_test.go",
        "line": 54,
        "endLine": 70,
        "proves": "Checks that n, Enter, Tab, Backspace and c return no commands and do not change the active view at three undersized dimensions.",
        "limits": "Selected keys and sizes only; paste is not asserted here.",
        "reviewedHash": "fcb14dea1373d1c821a7ef0039bfa0ec28484fb5005502e0546068b3366028b3",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:05.705936575Z",
          "scope": "TestCoordinatorDropsKeysWhenTooSmall; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "dashboard-list-requires-explicit-quit",
        "name": "TestDashboardListRequiresExplicitQuit",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/dashboard_list_keys_test.go",
        "line": 11,
        "endLine": 45,
        "proves": "Checks repeated Esc/Ctrl+C keep the dashboard list open at normal and small sizes and q returns QuitMsg.",
        "limits": "Coordinator-level input assertions; no physical terminal signal delivery.",
        "reviewedHash": "80762c5bb8bd9b5ed61a61ebd9268f056982d976e359b06c63dee8726b588763",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.017818592Z",
          "scope": "TestDashboardListRequiresExplicitQuit; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "card-editor-preserves-unmanaged-fields",
        "name": "TestCardEditorPreservesUnmanagedFields",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/card_editor_test.go",
        "line": 122,
        "endLine": 183,
        "proves": "Seeds every card field with nonzero data, edits Name, and compares the entire returned card with the expected overlay.",
        "limits": "Editor model round trip; repository persistence and read-only consumer policy are assessed elsewhere.",
        "reviewedHash": "2a7332fafb47cd5df346dea1eff1d549fd80c41d83b4c9122dd3cd1490fda622",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:58.638108503Z",
          "scope": "TestCardEditorPreservesUnmanagedFields; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "dashboard-control-card-editor-toggle-round-trip",
        "name": "TestDashboardControlCardEditorToggleRoundTrip",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/card_editor_test.go",
        "line": 184,
        "endLine": 200,
        "proves": "Focuses the control row, toggles off/on and checks Card().DashboardControl.",
        "limits": "Does not prove deployed activation authority or MCP permission changes.",
        "reviewedHash": "2a7332fafb47cd5df346dea1eff1d549fd80c41d83b4c9122dd3cd1490fda622",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:58.639084563Z",
          "scope": "TestDashboardControlCardEditorToggleRoundTrip; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "ctrl-x-clears-only-sub-workspace-selection",
        "name": "TestCtrlXClearsOnlySubWorkspaceSelection",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/card_editor_test.go",
        "line": 223,
        "endLine": 275,
        "proves": "Checks Ctrl+X clears selected/missing IDs, preserves all other card fields and focus, invokes no save/cancel, shows the hint and ignores plain x.",
        "limits": "In-memory editor; filesystem copies and registry writes are outside this test.",
        "reviewedHash": "2a7332fafb47cd5df346dea1eff1d549fd80c41d83b4c9122dd3cd1490fda622",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:58.643277152Z",
          "scope": "TestCtrlXClearsOnlySubWorkspaceSelection; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "card-editor-focused-field-always-visible",
        "name": "TestCardEditorFocusedFieldAlwaysVisible",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/card_editor_layout_test.go",
        "line": 262,
        "endLine": 286,
        "proves": "Checks every expected focus target remains visible with optional sections on/off, empty/large lists and minimum-sized form heights.",
        "limits": "Rendered text and model focus; no screenshot, native terminal or image baseline.",
        "reviewedHash": "36ac1bf09e7ef1d2e10115628f1ca6f164065619c422eaeb2021556770442626",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:58.030879304Z",
          "scope": "TestCardEditorFocusedFieldAlwaysVisible; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "partial-preview-broadcast-does-not-wipe-card-editor",
        "name": "TestPartialPreviewBroadcastDoesNotWipeCardEditor",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/preview_broadcast_test.go",
        "line": 19,
        "endLine": 72,
        "proves": "Asserts a loaded builtin skill selection survives an injected preview/output message while the editor is open.",
        "limits": "Deliberately re-arms a subscription to model late messages; does not prove normal editor preview subscription behavior.",
        "reviewedHash": "74eb2a13274325ddfa009e0aae169d37bbe0febb6d2828e54c54d22c3cf08ffa",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.300699448Z",
          "scope": "TestPartialPreviewBroadcastDoesNotWipeCardEditor; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "save-simple-source-surfaces-sync-failure",
        "name": "TestSaveSimpleSourceSurfacesSyncFailure",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/source_sync_warning_test.go",
        "line": 17,
        "endLine": 70,
        "proves": "Checks emitted sync error, refresh request and durable saved source path despite failed scanning.",
        "limits": "One missing skill-source path through callbacks; no full source-manager UI journey.",
        "reviewedHash": "16ae0ac11a321a83ff10b4b5d3cffb060f66a816805c5a30b497a8ee890a2f9d",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:05.80242298Z",
          "scope": "TestSaveSimpleSourceSurfacesSyncFailure; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "m-c-p-form-round-trips-every-field",
        "name": "TestMCPFormRoundTripsEveryField",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/mcp_form_test.go",
        "line": 148,
        "endLine": 203,
        "proves": "Presses Enter and asserts one save with exact ID/name/command/args/CWD/URL/env/headers, including an argument containing a comma.",
        "limits": "Callback payload only; does not connect to that server or persist the configuration.",
        "reviewedHash": "3e58b28d6fdcfbbc1c0f45984f19da8e223ef6c872f2212d1393492235248e6f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.513482413Z",
          "scope": "TestMCPFormRoundTripsEveryField; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "conversation-choice-requires-explicit-compatible-choice",
        "name": "TestConversationChoiceRequiresExplicitCompatibleChoice",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/conversation_choice_test.go",
        "line": 16,
        "endLine": 47,
        "proves": "Checks a matching Claude binding is submitted and a selected Codex binding cannot be submitted to the Claude activation.",
        "limits": "Picker callback boundary; native CLI resume is outside scope.",
        "reviewedHash": "bd540f8c7e3ac081fad2910f159272fde91d53d1b07cbbf26797de2d877b1021",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:58.674249451Z",
          "scope": "TestConversationChoiceRequiresExplicitCompatibleChoice; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "activation-rejects-unavailable-saved-workspace",
        "name": "TestActivationRejectsUnavailableSavedWorkspace",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/workspace_validation_test.go",
        "line": 27,
        "endLine": 117,
        "proves": "For relaunch/switch and original/temporary paths, checks missing/file errors, no extra Spawn, preserved session identity/content, and successful single-spawn retry after restoring the directory.",
        "limits": "Controlled driver counts and filesystem fixtures; no real Codex/Claude execution.",
        "reviewedHash": "e484a44ee3d38d1551068285e2471f861dd497b11af4cc4503256e6e81574550",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:18.030479824Z",
          "scope": "TestActivationRejectsUnavailableSavedWorkspace; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "read-sub-workspace-descriptors-does-not-prepare-workspace",
        "name": "TestReadSubWorkspaceDescriptorsDoesNotPrepareWorkspace",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/sub_workspaces_inspection_test.go",
        "line": 11,
        "endLine": 41,
        "proves": "Checks descriptor identity/path, absence of a newly created tmp directory and unavailable status for corrupt manifests.",
        "limits": "Read-only helper only; copy collision and multi-selection journeys remain separate.",
        "reviewedHash": "47b304b52e9a80c1587c84b7b469e5cab9c4696c3769199011887bbdeda0c3a7",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:07.018106997Z",
          "scope": "TestReadSubWorkspaceDescriptorsDoesNotPrepareWorkspace; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "relaunch-refreshes-only-kankuro-m-c-p-after-restart",
        "name": "TestRelaunchRefreshesOnlyKankuroMCPAfterRestart",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/relaunch_mcp_test.go",
        "line": 16,
        "endLine": 92,
        "proves": "Checks invalid config blocks relaunch without rewriting, then preserves instructions/custom server/metadata, rotates token, revokes the old credential and uses the new endpoint.",
        "limits": "Local files and cat process across in-process restart; no native CLI MCP handshake.",
        "reviewedHash": "d6d0f89ec7ac16c4c2395b51aeae0d31ba9bcab099c7395fd6806a03cc127526",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.821670683Z",
          "scope": "TestRelaunchRefreshesOnlyKankuroMCPAfterRestart; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "trust-invalid-configuration-preserved",
        "name": "TestTrustInvalidConfigurationPreserved",
        "method": "automated",
        "integration": "project",
        "file": "internal/driver/trust_test.go",
        "line": 86,
        "endLine": 115,
        "proves": "Checks PrepareWorkspace errors and byte-identical original configuration for Codex TOML, Claude JSON and a wrong-typed projects field.",
        "limits": "Isolated HOME; no installed native CLI trust acceptance.",
        "reviewedHash": "c3c5f502e8c165b4b78aeea83d59d1301a13bd91b3dd1233565c1709f8f9c3f2",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:57.242431346Z",
          "scope": "TestTrustInvalidConfigurationPreserved; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "shutdown-during-spawn-kills-the-new-process",
        "name": "TestShutdownDuringSpawnKillsTheNewProcess",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/shutdown_test.go",
        "line": 161,
        "endLine": 196,
        "proves": "Checks shutdown blocks for an in-flight relaunch, returns ErrManagerClosed after release, kills the spawned driver and leaves terminated status.",
        "limits": "Gated test driver; does not simulate every OS process-group behavior.",
        "reviewedHash": "c3feacc7073ff6582b64bf6830985cb9c7a26005624752028f9c7dd862320d42",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.481817694Z",
          "scope": "TestShutdownDuringSpawnKillsTheNewProcess; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "shutdown-waits-for-temporary-workspace-preparation",
        "name": "TestShutdownWaitsForTemporaryWorkspacePreparation",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/shutdown_test.go",
        "line": 235,
        "endLine": 276,
        "proves": "Blocks the copier, asserts Shutdown waits, then checks closed-manager launch error and an empty temporary root.",
        "limits": "Injected copier; does not prove native cp cancellation under every filesystem.",
        "reviewedHash": "c3feacc7073ff6582b64bf6830985cb9c7a26005624752028f9c7dd862320d42",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.689039183Z",
          "scope": "TestShutdownWaitsForTemporaryWorkspacePreparation; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "data-round-trip-copies-every-field",
        "name": "TestDataRoundTripCopiesEveryField",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/mapping_internal_test.go",
        "line": 12,
        "endLine": 35,
        "proves": "Fills fields with nonzero values and compares all persisted values through data/restore and snapshot, excluding intentional schema/status transformations.",
        "limits": "Pure mapping; no disk durability assertion or complete clone-alias isolation claim.",
        "reviewedHash": "7a24af59808ff28547db62ea979f2028b08a805bd1ae443bbd7bef97a9f77831",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.97018627Z",
          "scope": "TestDataRoundTripCopiesEveryField; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "report-notification-initial-state-is-durable-and-idempotent",
        "name": "TestReportNotificationInitialStateIsDurableAndIdempotent",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/messaging_test.go",
        "line": 46,
        "endLine": 91,
        "proves": "Checks pending for completed/blocked/question and not_required for progress; reopen/retry returns the same report ID and one stored item with the expected state.",
        "limits": "Message store on disk; no prompt delivery or agent processing.",
        "reviewedHash": "e853b5124ff8cf803fb6075b18985c8739e9ef4a7d766771631879e2aaff8007",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.791244424Z",
          "scope": "TestReportNotificationInitialStateIsDurableAndIdempotent; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "progress-replies-persist-before-completion",
        "name": "TestProgressRepliesPersistBeforeCompletion",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/messaging_test.go",
        "line": 218,
        "endLine": 259,
        "proves": "Checks progress deduplication, rejection of a changed kind under the same request ID, and ordered progress/completed reports after reopening.",
        "limits": "Does not observe whether a live CLI continues working after progress.",
        "reviewedHash": "e853b5124ff8cf803fb6075b18985c8739e9ef4a7d766771631879e2aaff8007",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.793450414Z",
          "scope": "TestProgressRepliesPersistBeforeCompletion; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "restart-marks-admitted-notification-uncertain",
        "name": "TestRestartMarksAdmittedNotificationUncertain",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/messaging_test.go",
        "line": 260,
        "endLine": 325,
        "proves": "Checks recovery marks admitted delivery uncertain while preserving the accepted obligation and request identity.",
        "limits": "Constructed durable records; no process crash at an actual PTY write.",
        "reviewedHash": "e853b5124ff8cf803fb6075b18985c8739e9ef4a7d766771631879e2aaff8007",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.794096964Z",
          "scope": "TestRestartMarksAdmittedNotificationUncertain; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
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
        "reviewedHash": "47cb7f6b304ebb0ed39e27062af620dc9b6db936838017137cd4e931b076e6d3",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:57.619733686Z",
          "scope": "TestReportAutomaticallyNotifiesOriginalController; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "saved-notification-survives-reporter-stopping",
        "name": "TestSavedNotificationSurvivesReporterStopping",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 654,
        "endLine": 684,
        "proves": "Stops dispatcher, saves report, kills reporter, restarts runner and waits for notification output.",
        "limits": "Same application process and local harness; not an application-wide restart.",
        "reviewedHash": "47cb7f6b304ebb0ed39e27062af620dc9b6db936838017137cd4e931b076e6d3",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:57.870628913Z",
          "scope": "TestSavedNotificationSurvivesReporterStopping; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "stopped-original-activation-makes-notification-unavailable",
        "name": "TestStoppedOriginalActivationMakesNotificationUnavailable",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 685,
        "endLine": 708,
        "proves": "Stops/relaunches the original sender, attempts dispatch and checks unavailable notification state.",
        "limits": "State assertion for one activation replacement; not arbitrary authorization races.",
        "reviewedHash": "47cb7f6b304ebb0ed39e27062af620dc9b6db936838017137cd4e931b076e6d3",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:57.91535088Z",
          "scope": "TestStoppedOriginalActivationMakesNotificationUnavailable; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "notification-retries-after-queue-capacity-returns",
        "name": "TestNotificationRetriesAfterQueueCapacityReturns",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 521,
        "endLine": 577,
        "proves": "Fills queue, checks pending obligation and request ID survive rejection, drains queue, starts workers and observes notification with the same request ID.",
        "limits": "Controlled runner/queue; no native agent readiness proof.",
        "reviewedHash": "47cb7f6b304ebb0ed39e27062af620dc9b6db936838017137cd4e931b076e6d3",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:57.666367803Z",
          "scope": "TestNotificationRetriesAfterQueueCapacityReturns; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "notification-reconciliation-uses-delivery-evidence",
        "name": "TestNotificationReconciliationUsesDeliveryEvidence",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 578,
        "endLine": 653,
        "proves": "Checks state mapping, advances the attempt only for not-submitted delivery and verifies stale reconciliation does not downgrade or advance it again.",
        "limits": "Constructed persisted delivery states; concrete driver byte-write boundaries need their own tests.",
        "reviewedHash": "47cb7f6b304ebb0ed39e27062af620dc9b6db936838017137cd4e931b076e6d3",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:57.814122415Z",
          "scope": "TestNotificationReconciliationUsesDeliveryEvidence; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "concurrent-notification-dispatch-admits-one-attempt",
        "name": "TestConcurrentNotificationDispatchAdmitsOneAttempt",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 828,
        "endLine": 858,
        "proves": "Runs twelve dispatchers against the same obligation and asserts one matching runner operation.",
        "limits": "One controlled interleaving; this run did not use the race detector.",
        "reviewedHash": "47cb7f6b304ebb0ed39e27062af620dc9b6db936838017137cd4e931b076e6d3",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:59.039145432Z",
          "scope": "TestConcurrentNotificationDispatchAdmitsOneAttempt; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "notification-scheduling-preserves-recipient-order-and-fairness",
        "name": "TestNotificationSchedulingPreservesRecipientOrderAndFairness",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 925,
        "endLine": 951,
        "proves": "Checks one recipient backs off without suppressing another and keeps its accepted predecessor ahead of a later pending report.",
        "limits": "Pure scheduling function with fixed time; no throughput benchmark.",
        "reviewedHash": "47cb7f6b304ebb0ed39e27062af620dc9b6db936838017137cd4e931b076e6d3",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:59.07518189Z",
          "scope": "TestNotificationSchedulingPreservesRecipientOrderAndFairness; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "launch-workspace-command-executes-in-workspace",
        "name": "TestLaunchWorkspaceCommandExecutesInWorkspace",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/ide_launcher_test.go",
        "line": 69,
        "endLine": 79,
        "reviewedHash": "77686722e24693b846bff6f8047659d01ba9757df42d8e0fed047c940f32849f",
        "proves": "The helper runWorkspaceLauncherCase starts a real child and asserts absolute cwd, exact placeholder multiplicity, quoted paths with spaces and preservation of ordinary arguments.",
        "limits": "Assertions reside in runWorkspaceLauncherCase at internal/session/ide_launcher_test.go:80, called by this table-driven test. Uses the test executable, not an IDE GUI.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:02.15047283Z",
          "scope": "TestLaunchWorkspaceCommandExecutesInWorkspace; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "dashboard_-launch-terminal-routing",
        "name": "TestDashboard_LaunchTerminalRouting",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_test.go",
        "line": 322,
        "endLine": 371,
        "reviewedHash": "c920b81707a08ff6e6af8c1a5cc5d4a9e5f7a702861887fe5797bfa1b68d4b48",
        "proves": "Checks the terminal callback receives the selected session for running/waiting/terminated states, ignores no-selection/search/nil callback cases, and checks o selects the expected IDE session.",
        "limits": "View callback assertions only; no command execution or external UI.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.506123793Z",
          "scope": "TestDashboard_LaunchTerminalRouting; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "terminal-launch-through-dashboard",
        "name": "TestTerminalLaunchThroughDashboard",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/terminal_launch_test.go",
        "line": 43,
        "endLine": 115,
        "reviewedHash": "0c4f6d7c1e26657e6fdcff703800126c243b5fe46651caf4ac4098da1a673cb7",
        "proves": "Drives Tab/t through coordinator and executes the command; child record must contain the actual temporary WorkingDir in cwd/args and differ from OriginalDir.",
        "limits": "Real local helper process rather than a graphical terminal emulator; one configured launch shape.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:05.827426238Z",
          "scope": "TestTerminalLaunchThroughDashboard; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "terminal-launch-reports-configuration-and-lookup-errors",
        "name": "TestTerminalLaunchReportsConfigurationAndLookupErrors",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/terminal_launch_test.go",
        "line": 149,
        "endLine": 213,
        "reviewedHash": "0c4f6d7c1e26657e6fdcff703800126c243b5fe46651caf4ac4098da1a673cb7",
        "proves": "Checks missing/foreign scope, missing IDE/terminal commands, failed executable and absent workspace; feeds spawn failure back and asserts the coordinator banner is set.",
        "limits": "Most cases invoke callbacks directly. The banner assertion establishes error presence, not its full displayed text.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:05.836167808Z",
          "scope": "TestTerminalLaunchReportsConfigurationAndLookupErrors; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "settings-view_-save-in-edit-mode",
        "name": "TestSettingsView_SaveInEditMode",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/settings_test.go",
        "line": 57,
        "endLine": 80,
        "reviewedHash": "85c62ee5ef6b2dd5b79feec61a07cc1209a8b03a085c634307c2587c5eec7d7d",
        "proves": "Checks Enter calls OnSave with the edited IDE command and returns the view to display mode.",
        "limits": "In-memory callback payload only; does not execute the asynchronous repository save.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:07.248780925Z",
          "scope": "TestSettingsView_SaveInEditMode; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "settings-view_-terminal-command-focus-save-and-cancel",
        "name": "TestSettingsView_TerminalCommandFocusSaveAndCancel",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/settings_test.go",
        "line": 136,
        "endLine": 164,
        "reviewedHash": "85c62ee5ef6b2dd5b79feec61a07cc1209a8b03a085c634307c2587c5eec7d7d",
        "proves": "Checks Tab/Shift+Tab preserve separate command text, Enter trims/saves both fields, and Esc restores both saved inputs after further edits.",
        "limits": "View model callback boundary only; no YAML write.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:07.249061595Z",
          "scope": "TestSettingsView_TerminalCommandFocusSaveAndCancel; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "settings-view_-toggle-rows",
        "name": "TestSettingsView_ToggleRows",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/settings_test.go",
        "line": 81,
        "endLine": 102,
        "reviewedHash": "85c62ee5ef6b2dd5b79feec61a07cc1209a8b03a085c634307c2587c5eec7d7d",
        "proves": "Tab focuses the two visibility toggles, Space changes them, and settings values reflect the changes.",
        "limits": "Does not execute coordinator persistence or check every focus wrap boundary.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:07.248995744Z",
          "scope": "TestSettingsView_ToggleRows; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "settings-view_-long-commands-stay-within-size",
        "name": "TestSettingsView_LongCommandsStayWithinSize",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/settings_test.go",
        "line": 165,
        "endLine": 176,
        "reviewedHash": "85c62ee5ef6b2dd5b79feec61a07cc1209a8b03a085c634307c2587c5eec7d7d",
        "proves": "For display/edit mode at 80×20 and 120×36, checks rows/display widths, no injected ANSI/newline text, retained controls and edit cancel help.",
        "limits": "String/cell assertions rather than terminal screenshots; helper assertSettingsFits is at settings_test.go:181.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:07.250798044Z",
          "scope": "TestSettingsView_LongCommandsStayWithinSize; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "app-settings-persistence",
        "name": "TestAppSettingsPersistence",
        "method": "automated",
        "integration": "project",
        "file": "internal/config/app_settings_test.go",
        "line": 13,
        "endLine": 56,
        "reviewedHash": "c592744c5bf639742840eff42c89f67ca9d7280245e9bfdfa11301f3914c272d",
        "proves": "Checks default command/visibility settings, saves independent IDE/terminal commands and visibility options, then reads back exact AppSettings equality.",
        "limits": "Repository Save/Get against the same repository instance; no real app restart or coordinated navigation assertion.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.716000153Z",
          "scope": "TestAppSettingsPersistence; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "workspace-source-i-d-assigned-and-preserved-on-edit",
        "name": "TestWorkspaceSourceIDAssignedAndPreservedOnEdit",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/source_save_test.go",
        "line": 203,
        "endLine": 257,
        "reviewedHash": "ad2cf273e0e9665d3397e7fa3d5c16d937c303dc3b5fb1233031eceb6a8bf3c5",
        "proves": "Executes source save/edit callbacks, then checks a nonempty repository-assigned ID survives a change of IsRoot.",
        "limits": "Does not drive Ctrl+E, the source form or navigation back to an unsaved card.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:05.79729125Z",
          "scope": "TestWorkspaceSourceIDAssignedAndPreservedOnEdit; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "save-workspace-source-aborts-on-malformed-registry",
        "name": "TestSaveWorkspaceSourceAbortsOnMalformedRegistry",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/source_save_test.go",
        "line": 115,
        "endLine": 155,
        "reviewedHash": "ad2cf273e0e9665d3397e7fa3d5c16d937c303dc3b5fb1233031eceb6a8bf3c5",
        "proves": "Executes the workspace save command against malformed source YAML, checks ErrorMsg and byte-identical failed-read input.",
        "limits": "Callback error boundary only; no source-manager key journey.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:05.786941181Z",
          "scope": "TestSaveWorkspaceSourceAbortsOnMalformedRegistry; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "delete-workspace-source-aborts-on-malformed-registry",
        "name": "TestDeleteWorkspaceSourceAbortsOnMalformedRegistry",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/source_save_test.go",
        "line": 156,
        "endLine": 202,
        "reviewedHash": "ad2cf273e0e9665d3397e7fa3d5c16d937c303dc3b5fb1233031eceb6a8bf3c5",
        "proves": "Executes delete callback against malformed workspace sources and checks ErrorMsg without modifying the original bytes.",
        "limits": "Does not check successful deletion refresh or return navigation.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:05.79153707Z",
          "scope": "TestDeleteWorkspaceSourceAbortsOnMalformedRegistry; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "selected-only-skill-cursor-normalizes-after-registry-refresh",
        "name": "TestSelectedOnlySkillCursorNormalizesAfterRegistryRefresh",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/card_editor_test.go",
        "line": 201,
        "endLine": 222,
        "reviewedHash": "2a7332fafb47cd5df346dea1eff1d549fd80c41d83b4c9122dd3cd1490fda622",
        "proves": "Enables selected-only directly, removes a selected skill during registry refresh, checks cursor normalization to A and that Space toggles A.",
        "limits": "Does not send Ctrl+S or assert header indicator, search intersection or re-enabling all items.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:58.641615782Z",
          "scope": "TestSelectedOnlySkillCursorNormalizesAfterRegistryRefresh; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "sub-workspace-picker-multiple-unresolved-and-main-modes",
        "name": "TestSubWorkspacePickerMultipleUnresolvedAndMainModes",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/workspace_picker_test.go",
        "line": 77,
        "endLine": 128,
        "reviewedHash": "56f8d29c821f80c7d198b88b8a45ce36502fea738c742a25e62a36fc438a756d",
        "proves": "Drives the sub-picker through missing selection repair, search, provisional toggles, cancel, confirmation of two IDs, main-picker mode reset and saving both sub IDs.",
        "limits": "In-memory form/view sequence; does not copy sub-workspace files or persist the card repository.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:07.263170084Z",
          "scope": "TestSubWorkspacePickerMultipleUnresolvedAndMainModes; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "sub-workspace-new-selection-survives-registry-refresh",
        "name": "TestSubWorkspaceNewSelectionSurvivesRegistryRefresh",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/workspace_picker_test.go",
        "line": 129,
        "endLine": 177,
        "reviewedHash": "56f8d29c821f80c7d198b88b8a45ce36502fea738c742a25e62a36fc438a756d",
        "proves": "Checks a newly selected workspace remains visible as missing after refresh, cancel does not commit, explicit missing-row deselection works, and confirming a missing checked ID retains it.",
        "limits": "View/registry-message fixture, not concurrent disk registry scanning.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:07.269702663Z",
          "scope": "TestSubWorkspaceNewSelectionSurvivesRegistryRefresh; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "launch-workspace-command-errors",
        "name": "TestLaunchWorkspaceCommandErrors",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/ide_launcher_test.go",
        "line": 155,
        "endLine": 172,
        "reviewedHash": "77686722e24693b846bff6f8047659d01ba9757df42d8e0fed047c940f32849f",
        "proves": "Checks unconfigured/whitespace command, empty/missing workspace path and missing executable return the expected sentinels or an error.",
        "limits": "Shared launcher API tests, not every caller-specific banner or concurrent filesystem change.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:02.15061457Z",
          "scope": "TestLaunchWorkspaceCommandErrors; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "activation-published-before-startup-delay",
        "name": "TestActivationPublishedBeforeStartupDelay",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/startup_integration_test.go",
        "line": 56,
        "endLine": 110,
        "reviewedHash": "1847caf3b093d5208ab0c5fd95a41f44dededa6ffbab7daa7730fc1e37e35a25",
        "proves": "Checks a published real cat activation reports starting/process-started and permits immediate attachment; controlled timer transitions to active without enabling MCP-disabled delivery.",
        "limits": "Injected startup clock/driver preparation; does not assert native CLI task readiness.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:07.000527357Z",
          "scope": "TestActivationPublishedBeforeStartupDelay; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "startup-admission-and-registration-are-independent",
        "name": "TestStartupAdmissionAndRegistrationAreIndependent",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/startup_boundary_test.go",
        "line": 73,
        "endLine": 114,
        "reviewedHash": "deb0728efc7f561909d61e939dfae8e0ea17909c240d0f18e74c6ba591b576fd",
        "proves": "Registers a conversation during starting, checks early delivery rejection, then advances the controlled startup timer and verifies admission becomes available.",
        "limits": "Controlled clock/manager fixture, not native CLI processing.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:01.336299952Z",
          "scope": "TestStartupAdmissionAndRegistrationAreIndependent; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "inspection-conversion-preserves-activation-snapshot",
        "name": "TestInspectionConversionPreservesActivationSnapshot",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/inspection_runtime_test.go",
        "line": 11,
        "endLine": 40,
        "reviewedHash": "6df6907a9f82ded9c9630b78044cc58fe22938ac554baadb19ec322a3efaa60a",
        "proves": "Checks inspection copies activation/startup/delivery fields, serialized historical effects stay unchanged after source snapshot mutation and old data does not fabricate runtime fields.",
        "limits": "Pure conversion/serialization test, separate from transport.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.805084578Z",
          "scope": "TestInspectionConversionPreservesActivationSnapshot; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "dashboard-scope-over-h-t-t-p",
        "name": "TestDashboardScopeOverHTTP",
        "method": "automated",
        "integration": "project",
        "file": "internal/mcp/dashboard_scope_test.go",
        "line": 21,
        "endLine": 172,
        "reviewedHash": "9b762f25ed9999928f3bcb4d29513636fb34986351fef01c4e11f5689e31efa4",
        "proves": "Authenticates two dashboards over the real Streamable HTTP SDK client, checks peer session discovery excludes caller and foreign sessions.",
        "limits": "Only list_sessions transport scope is asserted here; no external AI client.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:57.060497875Z",
          "scope": "TestDashboardScopeOverHTTP; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "operation-runner-creates-once-and-replays-request",
        "name": "TestOperationRunnerCreatesOnceAndReplaysRequest",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 156,
        "endLine": 455,
        "reviewedHash": "592e2d9e09e30a20069d19b9395846eea1222bdc6a1fc71dbfa75c3165318792",
        "proves": "With real cat sessions, checks deployed controller/registration-only authority, scoped discovery/inspection/create, duplicate concurrent create identity, request conflicts, cross-dashboard denials, lifecycle replay/restart and cancellation.",
        "limits": "In-process service/runner fixture; no TUI or installed native-agent client. Tests the declared finite set of authorization and operation cases.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:59.524544126Z",
          "scope": "TestOperationRunnerCreatesOnceAndReplaysRequest; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "unauthenticated-discovery-is-denied",
        "name": "TestUnauthenticatedDiscoveryIsDenied",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 142,
        "endLine": 155,
        "reviewedHash": "592e2d9e09e30a20069d19b9395846eea1222bdc6a1fc71dbfa75c3165318792",
        "proves": "Checks empty credentials cannot list cards/sessions or inspect a target.",
        "limits": "Service rejection assertions without transport status assertions.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:59.251895841Z",
          "scope": "TestUnauthenticatedDiscoveryIsDenied; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "final-journal-write-failure-returns-outcome-and-recovers-interrupted",
        "name": "TestFinalJournalWriteFailureReturnsOutcomeAndRecoversInterrupted",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 456,
        "endLine": 535,
        "reviewedHash": "592e2d9e09e30a20069d19b9395846eea1222bdc6a1fc71dbfa75c3165318792",
        "proves": "Injects failure after a successful creation effect, checks usable target/result with warning, in-memory replay without another session, and interrupted recovery without auto-execution.",
        "limits": "Injected journal write seam and in-process runner restart, not OS crash/power-loss.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:59.556599765Z",
          "scope": "TestFinalJournalWriteFailureReturnsOutcomeAndRecoversInterrupted; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "operation-queue-bound-and-disconnected-wait-return-promptly",
        "name": "TestOperationQueueBoundAndDisconnectedWaitReturnPromptly",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 951,
        "endLine": 1120,
        "reviewedHash": "592e2d9e09e30a20069d19b9395846eea1222bdc6a1fc71dbfa75c3165318792",
        "proves": "Fills worker/queue gates, checks prompt return on disconnected waits, capacity_exhausted, eventual completed effects and truthful success-with-warning when cancellation arrives after publication.",
        "limits": "Finite controlled worker interleavings and deadlines; not a throughput or starvation proof.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:59.84234434Z",
          "scope": "TestOperationQueueBoundAndDisconnectedWaitReturnPromptly; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "operation-journal-recovers-interrupted-and-fails-closed-on-corruption",
        "name": "TestOperationJournalRecoversInterruptedAndFailsClosedOnCorruption",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 1121,
        "endLine": 1154,
        "reviewedHash": "592e2d9e09e30a20069d19b9395846eea1222bdc6a1fc71dbfa75c3165318792",
        "proves": "Writes a running record, checks interrupted recovery with warning and verifies a corrupt journal sets loadErr to block mutation admission.",
        "limits": "Constructed journal records and runner state; no actual process crash.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:59.8428063Z",
          "scope": "TestOperationJournalRecoversInterruptedAndFailsClosedOnCorruption; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "messaging-operations-pin-activations-and-persist-explicit-reports",
        "name": "TestMessagingOperationsPinActivationsAndPersistExplicitReports",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 1782,
        "endLine": 1900,
        "reviewedHash": "592e2d9e09e30a20069d19b9395846eea1222bdc6a1fc71dbfa75c3165318792",
        "proves": "Checks target activation inspection, submitted send result, stable replay identity, changed-text conflict, stored text, worker report permission and removal of history on recipient deletion.",
        "limits": "Local service/PTY fixtures; submission is byte evidence, not an agent acknowledgment or processing proof.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:00.044800369Z",
          "scope": "TestMessagingOperationsPinActivationsAndPersistExplicitReports; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "ordinary-message-journal-rejection-rolls-back-history-and-retries",
        "name": "TestOrdinaryMessageJournalRejectionRollsBackHistoryAndRetries",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_shared_delivery_test.go",
        "line": 78,
        "endLine": 121,
        "reviewedHash": "cbc127688b33201d49f397a3c95256bac975c8ca541718803b12274124b47fd7",
        "proves": "Injects journal admission failure, checks no intent/receipt survives, then retries and replays the same request with the original message identity.",
        "limits": "Controlled write failure; no later partial PTY write claim.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:59.147339046Z",
          "scope": "TestOrdinaryMessageJournalRejectionRollsBackHistoryAndRetries; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "ordinary-message-queue-rejection-releases-reservation-and-preserves-replay",
        "name": "TestOrdinaryMessageQueueRejectionReleasesReservationAndPreservesReplay",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_shared_delivery_test.go",
        "line": 34,
        "endLine": 77,
        "reviewedHash": "cbc127688b33201d49f397a3c95256bac975c8ca541718803b12274124b47fd7",
        "proves": "Fills operation queue, checks capacity rejection leaves no message, then retries/replays same message ID and stops recipient without a retained reservation.",
        "limits": "Controlled queue/worker fixture rather than unbounded production load.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:59.112647198Z",
          "scope": "TestOrdinaryMessageQueueRejectionReleasesReservationAndPreservesReplay; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "service-remove-session-delegates-guards-cleanup-and-manager-errors",
        "name": "TestServiceRemoveSessionDelegatesGuardsCleanupAndManagerErrors",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 1672,
        "endLine": 1733,
        "reviewedHash": "592e2d9e09e30a20069d19b9395846eea1222bdc6a1fc71dbfa75c3165318792",
        "proves": "Rejects removal of a running session, removes a stopped temporary workspace, propagates session-record cleanup failure and closed-manager activation error.",
        "limits": "Service/manager and real filesystem fixture; not the dashboard confirmation key sequence.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:59.998265502Z",
          "scope": "TestServiceRemoveSessionDelegatesGuardsCleanupAndManagerErrors; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "resolve-capabilities-adds-m-c-p-access-skill-only-when-kankuro-m-c-p-selected",
        "name": "TestResolveCapabilitiesAddsMCPAccessSkillOnlyWhenKankuroMCPSelected",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/service_test.go",
        "line": 1192,
        "endLine": 1222,
        "reviewedHash": "592e2d9e09e30a20069d19b9395846eea1222bdc6a1fc71dbfa75c3165318792",
        "proves": "Checks mcp-cli is absent without Kankuro MCP, both protocol skills appear once with Kankuro selection and explicit mcp-cli selection does not duplicate either.",
        "limits": "Tests capability selection, not instruction following by an AI agent.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:59.84684967Z",
          "scope": "TestResolveCapabilitiesAddsMCPAccessSkillOnlyWhenKankuroMCPSelected; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "embedded-m-c-p-c-l-i-skill-deploys-through-config",
        "name": "TestEmbeddedMCPCLISkillDeploysThroughConfig",
        "method": "automated",
        "integration": "project",
        "file": "main_test.go",
        "line": 13,
        "endLine": 41,
        "reviewedHash": "a7d91caf10f7521dc017c9a9317d878331cf2ebe5dfdc853fea56bbe5ec70b5a",
        "proves": "Checks embedded mcp-cli bytes/metadata and exact deployed SKILL.md bytes through the real driver deployment.",
        "limits": "Uses builtin lookup only as fixture setup for the live deployment path; does not establish any runtime need for that lookup helper.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.696139914Z",
          "scope": "TestEmbeddedMCPCLISkillDeploysThroughConfig; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "embedded-kankuro-puppet-skill-deploys-through-config",
        "name": "TestEmbeddedKankuroPuppetSkillDeploysThroughConfig",
        "method": "automated",
        "integration": "project",
        "file": "main_test.go",
        "line": 42,
        "endLine": 62,
        "reviewedHash": "a7d91caf10f7521dc017c9a9317d878331cf2ebe5dfdc853fea56bbe5ec70b5a",
        "proves": "Checks canonical embedded puppet skill metadata and byte-identical driver deployment to the expected skill folder.",
        "limits": "File deployment only; no evidence that an installed CLI reads or follows the skill.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.696863434Z",
          "scope": "TestEmbeddedKankuroPuppetSkillDeploysThroughConfig; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "dashboard-list-creates-named-dashboard",
        "name": "TestDashboardListCreatesNamedDashboard",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_list_test.go",
        "line": 159,
        "endLine": 179,
        "reviewedHash": "f5fe6b22f5ee6c2d978eef517df6c89fb4791816878d6f24d9c114ab7e8d2fee",
        "proves": "Drives n, rejects empty name, edits name/description and asserts callback values.",
        "limits": "Create-form callback only; no edit-form submission, repository write or selected-ID refresh assertion.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.485135534Z",
          "scope": "TestDashboardListCreatesNamedDashboard; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "dashboard-details-through-service",
        "name": "TestDashboardDetailsThroughService",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/membership_internal_test.go",
        "line": 348,
        "endLine": 370,
        "reviewedHash": "7788a680a0548d68f4a1d3ee5108c3ca6f219d435c34e55cd3ce6306b42e5cb2",
        "proves": "Creates and updates dashboard details, then checks the single returned inventory entry equals the update and retains the created identity.",
        "limits": "Service/repository test; no dashboard-list form key journey.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.914962342Z",
          "scope": "TestDashboardDetailsThroughService; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "session-attach-replay",
        "name": "TestSessionAttachReplay",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/attach_test.go",
        "line": 11,
        "endLine": 52,
        "reviewedHash": "50a58baf339a4b8439246f3d2e98a6ec453af1063d2a190efd536e99adb6a4f4",
        "proves": "Writes hello through a real cat attachment, closes it, reattaches and asserts exact replayed scrollback.",
        "limits": "Manager handles only; does not run attachExec raw terminal mode, Ctrl+K or resize synchronization.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:01.489860464Z",
          "scope": "TestSessionAttachReplay; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "session-launcher-waits-for-name-result-and-allows-retry",
        "name": "TestSessionLauncherWaitsForNameResultAndAllowsRetry",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/session_launcher_test.go",
        "line": 121,
        "endLine": 147,
        "reviewedHash": "b868fdd3452c53f4f8f4967f63bf6084d84bb354fa020067ec9d3446a866c4e4",
        "proves": "Checks name submission starts once, duplicate Enter is suppressed while pending, an error is shown and another Enter can retry.",
        "limits": "In-memory naming callback/result boundary; not a persisted service rename.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:07.248557364Z",
          "scope": "TestSessionLauncherWaitsForNameResultAndAllowsRetry; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "set-name-propagates-persist-failure",
        "name": "TestSetNamePropagatesPersistFailure",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/lifecycle_test.go",
        "line": 412,
        "endLine": 456,
        "reviewedHash": "20a260206ef6895110aafb86e5bd2811799a452e27066a795d75135ee00cff4b",
        "proves": "Makes the session-record directory unwritable and asserts the checked name operation propagates its persistence error.",
        "limits": "Permission-based test skips as root; manager fixture is not the complete UI rename path.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.359369767Z",
          "scope": "TestSetNamePropagatesPersistFailure; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "card-deletion-retries-pending-workspace-cleanup-after-restart",
        "name": "TestCardDeletionRetriesPendingWorkspaceCleanupAfterRestart",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/card_cleanup_internal_test.go",
        "line": 15,
        "endLine": 152,
        "reviewedHash": "cf170f60e2b5f59b84be6a7001d43fb756a631a4b235a31129fac5b71032392b",
        "proves": "Blocks a foreign pending copy, edits the card workspace, checks failed cleanup persists the original exact paths/error in a tombstone, then restarts and retries cleanup without removing a foreign session.",
        "limits": "Permission-based non-root filesystem fixture; in-process restart, not hard OS termination.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.722598682Z",
          "scope": "TestCardDeletionRetriesPendingWorkspaceCleanupAfterRestart; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "dashboard-list-shows-deletion-preparation-and-status-errors",
        "name": "TestDashboardListShowsDeletionPreparationAndStatusErrors",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_list_test.go",
        "line": 16,
        "endLine": 58,
        "reviewedHash": "f5fe6b22f5ee6c2d978eef517df6c89fb4791816878d6f24d9c114ab7e8d2fee",
        "proves": "At three sizes, drives preflight/status keys, feeds multiline/ANSI failure messages and checks visible bounded errors that clear after cancel or successful retry.",
        "limits": "View/message assertions; fake callbacks do not execute the deletion service. Helper assertDashboardDeletionErrorVisible supplies the bounds/sanitization oracle.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.474655015Z",
          "scope": "TestDashboardListShowsDeletionPreparationAndStatusErrors; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "dashboard-list-bounds-long-deletion-error-when-list-becomes-empty",
        "name": "TestDashboardListBoundsLongDeletionErrorWhenListBecomesEmpty",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_list_test.go",
        "line": 70,
        "endLine": 83,
        "reviewedHash": "f5fe6b22f5ee6c2d978eef517df6c89fb4791816878d6f24d9c114ab7e8d2fee",
        "proves": "Keeps a long CJK cleanup error visible after the dashboard inventory becomes empty and checks the view stays within 80×21.",
        "limits": "One narrow view state; not a screenshot or actual retry service run.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.475441804Z",
          "scope": "TestDashboardListBoundsLongDeletionErrorWhenListBecomesEmpty; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
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
        "reviewedHash": "cbc127688b33201d49f397a3c95256bac975c8ca541718803b12274124b47fd7",
        "proves": "For ordinary and notification queues, cancels before execution, starts a worker, checks not-submitted delivery, successful target stop and absence of message markers in captured output.",
        "limits": "Controlled runner/worker plus PTY bytes; ordinary uses public cancellation, notification cancels the captured task authority directly.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:59.197479973Z",
          "scope": "TestQueuedDeliveryCancellationReleasesReservationWithoutSubmitting; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "queued-message-reservations-cannot-deadlock-stop-workers-or-shutdown",
        "name": "TestQueuedMessageReservationsCannotDeadlockStopWorkersOrShutdown",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 179,
        "endLine": 289,
        "reviewedHash": "47cb7f6b304ebb0ed39e27062af620dc9b6db936838017137cd4e931b076e6d3",
        "proves": "Queues delivery reservations alongside stop operations, cancels sends, waits for stop and checks runner Shutdown returns before deadline with all targets terminated.",
        "limits": "Finite controlled worker interleaving for ordinary sends; dispatcher shutdown during a real notification write is not directly asserted.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:57.187253348Z",
          "scope": "TestQueuedMessageReservationsCannotDeadlockStopWorkersOrShutdown; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "deleting-report-history-cancels-admitted-notification",
        "name": "TestDeletingReportHistoryCancelsAdmittedNotification",
        "method": "automated",
        "integration": "project",
        "file": "internal/api/messaging_review_test.go",
        "line": 751,
        "endLine": 817,
        "reviewedHash": "47cb7f6b304ebb0ed39e27062af620dc9b6db936838017137cd4e931b076e6d3",
        "proves": "Publishes admitted notification intent, deletes dashboard history, then checks submission rejects and notifications stay absent rather than recreating deleted history.",
        "limits": "Direct manager/store admission sequence; no raw agent prompt-processing claim.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:58.995301444Z",
          "scope": "TestDeletingReportHistoryCancelsAdmittedNotification; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "m-c-p-c-l-i-registers-conversation-for-live-activation",
        "name": "TestMCPCLIRegistersConversationForLiveActivation",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/conversation_test.go",
        "line": 63,
        "endLine": 334,
        "reviewedHash": "67dbeb8f4c470db473e3a379e88aa76d2a2e187a372bea1e622d224a5a6ac758",
        "proves": "Runs a shell child using the installed mcp-cli against the real MCP endpoint, checks reserved-env isolation, binding persistence/warnings/retry, stale credentials, explicit choice and card/session-local bindings across restart.",
        "limits": "Interoperability test uses a shell/cat agent; it is not an installed Codex/Claude CLI accepting native resume. Skips when mcp-cli is absent; the recorded endpoint run passed.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:02.018737467Z",
          "scope": "TestMCPCLIRegistersConversationForLiveActivation; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "claude-resume-lifecycle-uses-selected-native-i-d",
        "name": "TestClaudeResumeLifecycleUsesSelectedNativeID",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/claude_resume_test.go",
        "line": 70,
        "endLine": 182,
        "reviewedHash": "7f6768bfaccb7c57e03c66b33d35eb4165bbced868a119c602240ac4500aa144",
        "proves": "Captures fresh/saved/new and A→B→A Claude arguments, checks original workspace marker survives and conversation IDs stay card-local.",
        "limits": "Overrides Spawn and PrepareWorkspace with a capture driver; never starts the installed Claude CLI.",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:01.510357493Z",
          "scope": "TestClaudeResumeLifecycleUsesSelectedNativeID; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "temporary-workspace-space-save",
        "name": "TestTemporaryWorkspaceSpaceAndSave",
        "method": "automated",
        "integration": "project",
        "kind": "unit",
        "file": "internal/ui/views/card_editor_test.go",
        "proves": "Focuses Use Tmp Workspace, sends Space for false→true→false, then saves each boolean through OnSave and checks the complete normalized card.",
        "limits": "Card editor model/key/callback boundary; no terminal screenshot or repository write.",
        "line": 644,
        "endLine": 681,
        "reviewedHash": "2a7332fafb47cd5df346dea1eff1d549fd80c41d83b4c9122dd3cd1490fda622",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:58.673864221Z",
          "scope": "TestTemporaryWorkspaceSpaceAndSave; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "settings-save-failure-visible",
        "name": "TestSettingsSaveFailureRemainsVisible",
        "method": "automated",
        "integration": "project",
        "kind": "integration",
        "file": "internal/app/settings_error_test.go",
        "proves": "Edits settings through coordinator Update, executes the real save callback command, delivers SettingsSavedMsg, and checks the visible error, retained view/edit and unchanged persisted settings.",
        "limits": "Repository wrapper injects save failure; successful save/editor-refresh composition remains separately partial.",
        "line": 27,
        "endLine": 74,
        "reviewedHash": "dc56c003ff77c0b7e63ab091d05d3ab6458ad83c6ef408368d8f9ce58c5e6a27",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.620934791Z",
          "scope": "TestSettingsSaveFailureRemainsVisible; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "source-list-key-crud",
        "name": "TestSourceListIntegration_CRUD",
        "method": "automated",
        "integration": "project",
        "kind": "integration",
        "file": "internal/ui/components/source_list_integration_test.go",
        "proves": "Drives new/edit/save/delete keys, checks store/display equality, form/list transitions, callback counts, and automatic refresh/reload after successful callbacks.",
        "limits": "In-memory source store; does not establish filesystem persistence.",
        "line": 15,
        "endLine": 76,
        "reviewedHash": "8db1b7fa376767dcf736a0916f53f142d2545069e37a5b34c9bd81ce29d26668",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:01.844135306Z",
          "scope": "TestSourceListIntegration_CRUD; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
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
        "line": 285,
        "endLine": 308,
        "reviewedHash": "5a3a7e6203cc8b929a5aefff15098c9871e12b589602a6f1d8dcaf190487d212",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:01.844695986Z",
          "scope": "TestSourceList_HandleDelete_CallbackError; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "source-manager-delete-error",
        "name": "TestSourceManager_Delete",
        "method": "automated",
        "integration": "project",
        "kind": "integration",
        "file": "internal/ui/views/source_manager_test.go",
        "proves": "Drives deletion through SourceManager.Update, checks selected index, returned asynchronous ErrorMsg and preserved items after reload.",
        "limits": "Callback fixture emits the asynchronous error; repository failure coverage is separate.",
        "line": 195,
        "endLine": 233,
        "reviewedHash": "d26d45240dbdbbf44017bd99a3953a3deab304cef8933da3aa067ad27710d60d",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:07.251624274Z",
          "scope": "TestSourceManager_Delete; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "preview-sanitization-contract",
        "name": "TestSessionPreview_ScrollbackSanitization",
        "method": "automated",
        "integration": "project",
        "kind": "unit",
        "file": "internal/ui/components/session_preview_integration_test.go",
        "proves": "Checks visible Success/Error/Done text survives while OSC title and Kitty payload/control effects are removed.",
        "limits": "Styled component rendering; no native terminal or every possible escape sequence.",
        "line": 12,
        "endLine": 31,
        "reviewedHash": "2a3683ddd211d7ab5ba94831b6878a090d83919999138b64de833222108871c7",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:58.518516179Z",
          "scope": "TestSessionPreview_ScrollbackSanitization; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "dashboard-launch-persisted",
        "name": "TestFullSessionLaunchFlow",
        "method": "automated",
        "integration": "project",
        "kind": "integration",
        "file": "internal/app/integration_test.go",
        "proves": "Drives dashboard Enter, executes launch commands, verifies a running process and reads the persisted session/card/dashboard/workspace identity.",
        "limits": "Real local cat process; no native-agent task execution or raw-terminal attachment.",
        "line": 19,
        "endLine": 134,
        "reviewedHash": "1707c86e4737e0fd967dd80bd5ea6dd74f8ca4a07c3159cdf0431dd505fa98a5",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.07210313Z",
          "scope": "TestFullSessionLaunchFlow; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "source-refresh-destination",
        "name": "TestSourceManagementFlow",
        "method": "automated",
        "integration": "project",
        "kind": "integration",
        "file": "internal/app/integration_test.go",
        "proves": "Navigates to source management, changes persisted sources, delivers refresh results and checks actual list additions/removals.",
        "limits": "Coordinator/repository integration; does not cover the unsaved editor Ctrl+E return journey.",
        "line": 186,
        "endLine": 215,
        "reviewedHash": "1707c86e4737e0fd967dd80bd5ea6dd74f8ca4a07c3159cdf0431dd505fa98a5",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.084304499Z",
          "scope": "TestSourceManagementFlow; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "card-refresh-destination",
        "name": "TestCoordinatorDataPropagation",
        "method": "automated",
        "integration": "project",
        "kind": "integration",
        "file": "internal/app/integration_test.go",
        "proves": "Checks an initially empty dashboard then verifies refreshed cards and changed names in the actual destination view.",
        "limits": "Coordinator/repository integration, not every capability registry.",
        "line": 462,
        "endLine": 500,
        "reviewedHash": "1707c86e4737e0fd967dd80bd5ea6dd74f8ca4a07c3159cdf0431dd505fa98a5",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.137399637Z",
          "scope": "TestCoordinatorDataPropagation; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
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
        "endLine": 226,
        "reviewedHash": "9b01386e6b4ab1b5aa57181b11235d30eb05c52cfc559c5a4c7465b87e767786",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.798574308Z",
          "scope": "TestValidateCardRegistryIntegrity; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
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
        "endLine": 17,
        "reviewedHash": "677d53291a799c08ed87d7b4b950f840f3254ccbbe4df7702c7f11587deb17f4",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.702230263Z",
          "scope": "TestCanvasCreation; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "native-agent-terminal-journey",
        "name": "TestNativeAgentJourney",
        "method": "automated",
        "integration": "project",
        "kind": "e2e",
        "file": "tests/e2e/native_agent_test.go",
        "proves": "Uses tui-test Go to launch real Kankuro/OpenCode, verify imported credentials, assign an isolated service port and select OpenRouter Free Models Router. Reached assertions: Dashboard Enter launched a persisted native-agent process and Enter attached to the native UI.",
        "limits": "Explicit e2e tag required. Single OpenCode 2.2.6/OpenRouter router run; no screenshot baseline comparison. The full test file and configuration, capture and provenance helpers are retained in the inventory. The run failed at register; subsequent assertions were not reached. Historical native-agent result/captures were retained from 2026-11-09; this refresh did not execute that journey or rebind it to the current source fingerprint.",
        "line": 29,
        "endLine": 48,
        "reviewedHash": "41f3ea85716fea5e9811a386f25b55b6c75f2921f7352de6c63eef22f5e091df",
        "result": {
          "status": "failed",
          "command": "go test -tags=e2e ./tests/e2e -run '^TestNativeAgentJourney$' -count=1 -timeout=10m -v",
          "at": "2026-11-09T22:42:26.268507608Z",
          "scope": "TestNativeAgentJourney; openrouter/openrouter/free; final stage register",
          "revision": "WORKTREE",
          "fingerprint": "22ce60949b80ba3d6d306b85a4c88683bf040372efe7dbbff5d879030f24365a"
        }
      },
      {
        "id": "rename-draft",
        "name": "TestDashboard_RenameSession",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_test.go",
        "line": 192,
        "endLine": 240,
        "proves": "Captured target, normalized Unicode input, draft survives refresh, duplicate suppression, retry after error.",
        "limits": "View callback is a stub; no service write or cancellation assertion before submit.",
        "reviewedHash": "c920b81707a08ff6e6af8c1a5cc5d4a9e5f7a702861887fe5797bfa1b68d4b48",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.495640404Z",
          "scope": "TestDashboard_RenameSession; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "rename-blank",
        "name": "TestDashboard_RenameSessionEmptyUsesIDAndNoSelectionIsNoop",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_test.go",
        "line": 241,
        "endLine": 266,
        "proves": "Blank input submits the session ID.",
        "limits": "No-selection probe starts in the cards pane; it does not establish empty sessions-pane behavior.",
        "reviewedHash": "c920b81707a08ff6e6af8c1a5cc5d4a9e5f7a702861887fe5797bfa1b68d4b48",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.498136103Z",
          "scope": "TestDashboard_RenameSessionEmptyUsesIDAndNoSelectionIsNoop; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "rename-scoped",
        "name": "TestCoordinator_DashboardRenameFailureStaysInDialogAndIgnoresStaleResult",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/coordinator_test.go",
        "line": 58,
        "endLine": 93,
        "proves": "Real command error stays in dialog; stale generation leaves pending state intact.",
        "limits": "Missing session fixture; does not establish successful persistence or stale dashboard identity separately.",
        "reviewedHash": "96e217a0912e57be8507792ab1b6aea0fc02ec3b008cc54af0d9a6f50c01e51d",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.776549585Z",
          "scope": "TestCoordinator_DashboardRenameFailureStaysInDialogAndIgnoresStaleResult; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "rename-layout",
        "name": "TestDashboardRenameLayoutAndScopedFooter",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_layout_test.go",
        "line": 143,
        "endLine": 169,
        "proves": "Long error bounds at 80×24 and 120×40, visible save/cancel controls and scoped rename footer.",
        "limits": "String and cell bounds, not a native terminal screenshot or pixel comparison.",
        "reviewedHash": "b41624ff511eda1e9749598a3ad94f2ac00fa6611bae993a2f416c85bbc83afe",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:58.68897797Z",
          "scope": "TestDashboardRenameLayoutAndScopedFooter; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "rename-search",
        "name": "TestDashboard_Search",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_test.go",
        "line": 145,
        "endLine": 178,
        "proves": "Existing search consumes typed t/e/s/t and Esc exits search.",
        "limits": "Cards-pane search; does not explicitly exercise session-pane e search.",
        "reviewedHash": "c920b81707a08ff6e6af8c1a5cc5d4a9e5f7a702861887fe5797bfa1b68d4b48",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.493343354Z",
          "scope": "TestDashboard_Search; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "rename-persist",
        "name": "TestSetNamePropagatesPersistFailure",
        "method": "automated",
        "integration": "project",
        "file": "internal/session/lifecycle_test.go",
        "line": 412,
        "endLine": 456,
        "proves": "Existing checked rename surfaces a record-write failure.",
        "limits": "Manager boundary only; no complete dashboard rename/restart journey.",
        "reviewedHash": "20a260206ef6895110aafb86e5bd2811799a452e27066a795d75135ee00cff4b",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.359369767Z",
          "scope": "TestSetNamePropagatesPersistFailure; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "rename-launcher",
        "name": "TestSessionLauncherWaitsForNameResultAndAllowsRetry",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/session_launcher_test.go",
        "line": 121,
        "endLine": 147,
        "proves": "Existing launcher naming waits for result and permits retry.",
        "limits": "Launch naming form, not dashboard completion routing.",
        "reviewedHash": "b868fdd3452c53f4f8f4967f63bf6084d84bb354fa020067ec9d3446a866c4e4",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:07.248557364Z",
          "scope": "TestSessionLauncherWaitsForNameResultAndAllowsRetry; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "rename-cancel-draft",
        "name": "TestDashboard_RenameSessionDiscardsDraftWithoutSaving",
        "method": "automated",
        "integration": "project",
        "file": "internal/ui/views/dashboard_test.go",
        "proves": "Esc, dashboard changes and generation changes close an edited draft without saving or changing the session name. Reopening starts from the original name and submits once.",
        "limits": "View-level callback spy; no native terminal key decoding or persistence is involved.",
        "line": 267,
        "endLine": 272,
        "reviewedHash": "c920b81707a08ff6e6af8c1a5cc5d4a9e5f7a702861887fe5797bfa1b68d4b48",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.502853093Z",
          "scope": "TestDashboard_RenameSessionDiscardsDraftWithoutSaving; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "rename-rename-restart",
        "name": "TestCoordinator_DashboardRenameRefreshesWithoutAttachAndSurvivesRestart",
        "method": "automated",
        "integration": "project",
        "file": "internal/app/coordinator_test.go",
        "proves": "Real cat session renamed through coordinator input, service and disk; returned command only refreshes; selected dashboard name updates, view stays put, process PID/activation identity are retained; fresh repository/manager/coordinator reloads the saved name.",
        "limits": "Ordinary Go integration test using a local cat PTY; does not drive a native AI CLI or real terminal renderer.",
        "line": 121,
        "endLine": 184,
        "reviewedHash": "96e217a0912e57be8507792ab1b6aea0fc02ec3b008cc54af0d9a6f50c01e51d",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.795700044Z",
          "scope": "TestCoordinator_DashboardRenameRefreshesWithoutAttachAndSurvivesRestart; ordinary suite excludes E2E-tagged tests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-navigation",
        "name": "TestMCPManagerNestedNavigation",
        "file": "internal/app/navigation_test.go",
        "line": 115,
        "endLine": 141,
        "method": "automated",
        "integration": "project",
        "proves": "Repeated form cancel → manager → dashboard journeys preserve destinations.",
        "limits": "No real terminal or rendered key delivery.",
        "reviewedHash": "5bd7b802a0fc7b6156a004e3e478dc6f00653207d4fdd8c3b909b6f02fbc4968",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.289821179Z",
          "scope": "TestMCPManagerNestedNavigation",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-metadata",
        "name": "TestMCPManagerEditPreservesMetadataAndNewIsBlank",
        "file": "internal/app/mcp_manager_review_test.go",
        "line": 73,
        "endLine": 124,
        "method": "automated",
        "integration": "project",
        "proves": "Existing ID/description survives edit; New clears fields and metadata.",
        "limits": "In-process coordinator with isolated repository; no process restart.",
        "reviewedHash": "3e229a9ac835d5a1fe54d866630ba78b245dde7df4dd03cc2a27fda3e3a39992",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.273107329Z",
          "scope": "TestMCPManagerEditPreservesMetadataAndNewIsBlank",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-stale-commands",
        "name": "TestMCPManagerStaleRequests",
        "file": "internal/app/mcp_manager_review_test.go",
        "line": 13,
        "endLine": 72,
        "method": "automated",
        "integration": "project",
        "proves": "Old edit/delete requests rejected; new deletion succeeds and delayed result does not navigate.",
        "limits": "Same-dashboard leave/reopen; no explicit cross-dashboard delayed-command assertion.",
        "reviewedHash": "3e229a9ac835d5a1fe54d866630ba78b245dde7df4dd03cc2a27fda3e3a39992",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.26764511Z",
          "scope": "TestMCPManagerStaleRequests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-referenced",
        "name": "TestMCPManagerReferencedDeleteReportsError",
        "file": "internal/app/mcp_manager_review_test.go",
        "line": 125,
        "endLine": 178,
        "method": "automated",
        "integration": "project",
        "proves": "Saved card reference blocks deletion, retains server and displays an error.",
        "limits": "Reference belongs to default dashboard; no injected disk-write failure.",
        "reviewedHash": "3e229a9ac835d5a1fe54d866630ba78b245dde7df4dd03cc2a27fda3e3a39992",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.279805509Z",
          "scope": "TestMCPManagerReferencedDeleteReportsError",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-modal-target",
        "name": "TestMCPManagerConfirmationKeepsTargetAcrossRefresh",
        "file": "internal/ui/views/mcp_manager_test.go",
        "line": 15,
        "endLine": 40,
        "method": "automated",
        "integration": "project",
        "proves": "Modal captures ID across reorder, Escape cancels, nil data retains loaded rows.",
        "limits": "View-level messages; repository result is covered separately.",
        "reviewedHash": "193db5f05aee84e02967a53609e5cf7eef5760d029be9474a5b48e1d9fb297d6",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.970635879Z",
          "scope": "TestMCPManagerConfirmationKeepsTargetAcrossRefresh",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-protected",
        "name": "TestMCPManagerProtectedAndNew",
        "file": "internal/ui/views/mcp_manager_test.go",
        "line": 41,
        "endLine": 66,
        "method": "automated",
        "integration": "project",
        "proves": "Built-in Enter/Delete return errors; New opens blank and cannot be deleted.",
        "limits": "No real CLI or remote MCP server.",
        "reviewedHash": "193db5f05aee84e02967a53609e5cf7eef5760d029be9474a5b48e1d9fb297d6",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.970815169Z",
          "scope": "TestMCPManagerProtectedAndNew",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-long-list",
        "name": "TestMCPManagerLongListFits",
        "file": "internal/ui/views/mcp_manager_test.go",
        "line": 67,
        "endLine": 86,
        "method": "automated",
        "integration": "project",
        "proves": "Forty long Unicode/ANSI names fit 80×22; New reachable; confirmation fits.",
        "limits": "Bounds assertions only, no screenshot baseline or stable row-position assertion.",
        "reviewedHash": "193db5f05aee84e02967a53609e5cf7eef5760d029be9474a5b48e1d9fb297d6",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:57.001711328Z",
          "scope": "TestMCPManagerLongListFits",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-parser-retained",
        "name": "TestMCPFormRoundTripsEveryField",
        "file": "internal/ui/views/mcp_form_test.go",
        "line": 148,
        "endLine": 203,
        "method": "automated",
        "integration": "project",
        "proves": "Retained assertions round-trip visible fields including escaped args/env/headers.",
        "limits": "Direct form callback; no application navigation.",
        "reviewedHash": "3e58b28d6fdcfbbc1c0f45984f19da8e223ef6c872f2212d1393492235248e6f",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.513482413Z",
          "scope": "TestMCPFormRoundTripsEveryField",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-sources-retained",
        "name": "TestSourceList_Navigation",
        "file": "internal/ui/components/source_list_test.go",
        "line": 106,
        "endLine": 139,
        "method": "automated",
        "integration": "project",
        "proves": "Retained source-list arrow navigation and New position.",
        "limits": "Does not assert every source-manager CRUD path.",
        "reviewedHash": "5a3a7e6203cc8b929a5aefff15098c9871e12b589602a6f1d8dcaf190487d212",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:01.844405616Z",
          "scope": "TestSourceList_Navigation",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-repository-retained",
        "name": "TestKankuroMCPIsOfferedOnceAndProtected",
        "file": "internal/config/builtin_mcp_test.go",
        "line": 11,
        "endLine": 45,
        "method": "automated",
        "integration": "project",
        "proves": "Exactly one built-in; repository blocks edits, deletion and name shadowing.",
        "limits": "Isolated file repository; no UI.",
        "reviewedHash": "b28e8471e9e7f5e3c6ef47d89afdbdee27458b6a74c3da4576c34a705ca1f7b6",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:56.722894302Z",
          "scope": "TestKankuroMCPIsOfferedOnceAndProtected",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-cross-dashboard-requests",
        "name": "TestMCPManagerCrossDashboardRequests",
        "file": "internal/app/mcp_manager_gaps_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Old edit/delete messages are rejected after dashboard switch without clearing the new visit's confirmation; fresh deletion succeeds.",
        "limits": "Real coordinator and file repository; command delivery is held deterministically rather than scheduled by the terminal.",
        "reviewedHash": "13720406996dc7dc1f91c9a07c7b23d46f05aba708710a59c135c7e0973567ae",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.191154644Z",
          "scope": "TestMCPManagerCrossDashboardRequests",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-cross-dashboard-saves",
        "name": "TestMCPManagerCrossDashboardSaveResults",
        "file": "internal/app/mcp_manager_gaps_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Delayed successful/failed saves preserve the other dashboard's form draft, navigation and current error; successful global mutation refreshes with current scope.",
        "limits": "ENOSPC is injected at the repository boundary; does not fill a disk or exercise physical-media failure.",
        "reviewedHash": "13720406996dc7dc1f91c9a07c7b23d46f05aba708710a59c135c7e0973567ae",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.224576772Z",
          "scope": "TestMCPManagerCrossDashboardSaveResults",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-cross-dashboard-deletes",
        "name": "TestMCPManagerCrossDashboardDeleteResults",
        "file": "internal/app/mcp_manager_gaps_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Delayed success/failure cannot navigate or overwrite the new dashboard's error; success refreshes its global registry; failure releases pending state for retry.",
        "limits": "Two actual persisted dashboards and an isolated file registry; injected repository failure, no real concurrent disk outage.",
        "reviewedHash": "13720406996dc7dc1f91c9a07c7b23d46f05aba708710a59c135c7e0973567ae",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.245453271Z",
          "scope": "TestMCPManagerCrossDashboardDeleteResults",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-write-failure-retry",
        "name": "TestMCPManagerWriteFailuresRetainDataAndAllowRetry",
        "file": "internal/app/mcp_manager_gaps_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Save/delete PathError(ENOSPC) remains visible, registry bytes stay identical, drafts/rows survive, and retry persists successfully.",
        "limits": "Injects errors at the repository mutation boundary to avoid chmod/root dependence; atomic filesystem fault internals are outside this test.",
        "reviewedHash": "13720406996dc7dc1f91c9a07c7b23d46f05aba708710a59c135c7e0973567ae",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:04.2620067Z",
          "scope": "TestMCPManagerWriteFailuresRetainDataAndAllowRetry",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "mcp-mcp-terminal-journey",
        "name": "TestMCPManagerJourney",
        "file": "tests/e2e/mcp_manager_test.go",
        "method": "automated",
        "integration": "project",
        "kind": "E2E",
        "proves": "Actual executable: dashboard list → New/create → populated edit → rename preserving ID/command → Delete/cancel → Delete/confirm → return/reopen; six screenshots and saved-file assertions.",
        "limits": "No agents/model calls or screenshot baseline comparison. List/modal checked at 80×24; forms at 120×40.",
        "reviewedHash": "0b63e84ab8c3c871017132e66b9a9c3866d6d83d517e59ca1d71a3710a4b92bf",
        "result": {
          "status": "passed",
          "command": "go test -tags=e2e ./tests/e2e -run '^TestMCPManagerJourney$' -count=1 -timeout=3m -v",
          "at": "2026-11-11T22:34:16.106878795Z",
          "scope": "Real executable configuration journey, no agent/model calls",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "focus-completion",
        "name": "TestLaunchFocusSnapshotBeforeCompletion",
        "file": "internal/app/launch_focus_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Launch completion selects the exact created ID, including a snapshot already loaded.",
        "limits": "Coordinator events with controlled snapshots; native attachment is assessed separately.",
        "reviewedHash": "aae09330a865a90c2d04b0eaba6f10cc57d93ea9bfd3822741aef414b8d36c25",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.618484733Z",
          "scope": "TestLaunchFocusSnapshotBeforeCompletion",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "focus-epochs",
        "name": "TestLaunchFocusRejectsRefreshStartedBeforeLaunchAndName",
        "file": "internal/app/launch_focus_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Old session epochs cannot displace the created or confirmed-name target.",
        "limits": "Deterministic snapshot ordering, not real scheduler timing.",
        "reviewedHash": "aae09330a865a90c2d04b0eaba6f10cc57d93ea9bfd3822741aef414b8d36c25",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.624696933Z",
          "scope": "TestLaunchFocusRejectsRefreshStartedBeforeLaunchAndName",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "rename-epochs",
        "name": "TestInlineRenameRejectsOlderRefreshSnapshots",
        "file": "internal/app/inline_rename_refresh_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Full, sessions and fingerprint snapshots started before inline rename do not restore the old name.",
        "limits": "In-process command/result delivery.",
        "reviewedHash": "746d0f43d4a7f49e31e8949800f1a5d6812a68c3845149af004a85220b4695fd",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.612712594Z",
          "scope": "TestInlineRenameRejectsOlderRefreshSnapshots",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "focus-warning",
        "name": "TestLaunchFocusPersistenceWarningStillSelectsSession",
        "file": "internal/app/launch_focus_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Usable launch with metadata persistence warning keeps selected focus and exposes the warning.",
        "limits": "Controlled ErrPersistFailed result, not a physical disk fault.",
        "reviewedHash": "aae09330a865a90c2d04b0eaba6f10cc57d93ea9bfd3822741aef414b8d36c25",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.637971252Z",
          "scope": "TestLaunchFocusPersistenceWarningStillSelectsSession",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "focus-stale",
        "name": "TestLaunchFocusIgnoresStaleResults",
        "file": "internal/app/launch_focus_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Stale dashboard/generation results cannot move session focus.",
        "limits": "Does not launch a native agent.",
        "reviewedHash": "aae09330a865a90c2d04b0eaba6f10cc57d93ea9bfd3822741aef414b8d36c25",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.648507662Z",
          "scope": "TestLaunchFocusIgnoresStaleResults",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "focus-cancel",
        "name": "TestLaunchFocusPendingRequestIsCancelledOnDashboardExit",
        "file": "internal/app/launch_focus_internal_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Leaving a dashboard clears pending focus; delayed launch cannot apply in a different context.",
        "limits": "Controlled coordinator events.",
        "reviewedHash": "aae09330a865a90c2d04b0eaba6f10cc57d93ea9bfd3822741aef414b8d36c25",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:03.653821262Z",
          "scope": "TestLaunchFocusPendingRequestIsCancelledOnDashboardExit",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "focus-order",
        "name": "TestSessionList_SetSessionsPreservesSelectionByID",
        "file": "internal/ui/components/session_list_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Newer session insertion preserves the selected ID rather than its row index.",
        "limits": "Direct list component assertion.",
        "reviewedHash": "115125cb966d130cfc83351a245e2760226e2b3f6bd726fc22e9ed1a62dde95c",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:34:58.517980309Z",
          "scope": "TestSessionList_SetSessionsPreservesSelectionByID",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "focus-search",
        "name": "TestDashboard_FocusSessionWaitsForDataAndClearsOnlySessionSearch",
        "file": "internal/ui/views/dashboard_test.go",
        "method": "automated",
        "integration": "project",
        "proves": "Focus waits for target data, clears hiding session search, selects sessions pane and aligns preview; later refreshes do not steal focus.",
        "limits": "Direct view-level data refresh.",
        "reviewedHash": "c920b81707a08ff6e6af8c1a5cc5d4a9e5f7a702861887fe5797bfa1b68d4b48",
        "result": {
          "status": "passed",
          "command": "go test -json ./... -count=1",
          "at": "2026-11-11T22:35:06.491274094Z",
          "scope": "TestDashboard_FocusSessionWaitsForDataAndClearsOnlySessionSearch",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      },
      {
        "id": "launch-focus-terminal",
        "name": "TestNativeLaunchFocusReturn",
        "file": "tests/e2e/launch_focus_test.go",
        "method": "automated",
        "integration": "project",
        "kind": "E2E",
        "proves": "Real Codex startup screen after launch/name; detach restores selected created session; Enter reattaches to same live process, preserving exactly one older terminated session.",
        "limits": "Startup-only; no agent prompt/provider inference or pixel-baseline comparison. Requires Codex installed.",
        "reviewedHash": "e060a644c780ae1bc26cc70174c10c78034d26265c2c4eff54d39e19eb746a4e",
        "result": {
          "status": "passed",
          "command": "go test -tags=e2e ./tests/e2e -run '^TestNativeLaunchFocusReturn$' -count=1 -timeout=3m -v",
          "at": "2026-11-11T22:34:13.965777217Z",
          "scope": "Actual executable launch/name/detach/reattach with isolated native CLI state",
          "revision": "WORKTREE",
          "fingerprint": "8995dacba1c91320604c55e44cb8736879e336aac4d1920c2bb31617e27fd09c"
        }
      }
    ],
    "captures": [
      {
        "id": "native-launched",
        "title": "OpenCode with OpenRouter Free Models Router",
        "file": "tests/e2e/artifacts/native-agent/launched.png",
        "state": "Native UI reached through Kankuro launch/attach",
        "observation": "The native model label shows Free Models Router OpenRouter.",
        "limits": "Launch/model/attachment checkpoint only.",
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
        "title": "Native journey failure at register",
        "file": "tests/e2e/artifacts/native-agent/failure.png",
        "state": "Terminal at failed register checkpoint",
        "observation": "The inspected image shows the explicit registration task and Free Models Router after a completed response, with no registration result or executed shell command visible. No quota error is shown.",
        "limits": "Failure evidence; subsequent checkpoints are not established.",
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
        "title": "MCP configuration · list",
        "file": "tests/e2e/artifacts/mcp-manager/list.png",
        "state": "list",
        "observation": "Only the protected built-in and New are initially available.",
        "limits": "Project-owned asserted checkpoint; no pixel baseline comparison. Visual inspection documents the reached state, not all terminal sizes or themes.",
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
        "title": "MCP configuration · new",
        "file": "tests/e2e/artifacts/mcp-manager/new.png",
        "state": "new",
        "observation": "New opens a blank MCP form with save/back instructions.",
        "limits": "Project-owned asserted checkpoint; no pixel baseline comparison. Visual inspection documents the reached state, not all terminal sizes or themes.",
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
        "title": "MCP configuration · edit",
        "file": "tests/e2e/artifacts/mcp-manager/edit.png",
        "state": "edit",
        "observation": "Existing entry loads its name and command in the editor.",
        "limits": "Project-owned asserted checkpoint; no pixel baseline comparison. Visual inspection documents the reached state, not all terminal sizes or themes.",
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
        "title": "MCP configuration · edited-list",
        "file": "tests/e2e/artifacts/mcp-manager/edited-list.png",
        "state": "edited-list",
        "observation": "At 80×24 the renamed server, protected built-in, New and help are visible.",
        "limits": "Project-owned asserted checkpoint; no pixel baseline comparison. Visual inspection documents the reached state, not all terminal sizes or themes.",
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
        "title": "MCP configuration · confirm",
        "file": "tests/e2e/artifacts/mcp-manager/confirm.png",
        "state": "confirm",
        "observation": "Delete opens a named confirmation with confirm/cancel instructions at 80×24.",
        "limits": "Project-owned asserted checkpoint; no pixel baseline comparison. Visual inspection documents the reached state, not all terminal sizes or themes.",
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
        "title": "MCP configuration · deleted",
        "file": "tests/e2e/artifacts/mcp-manager/deleted.png",
        "state": "deleted",
        "observation": "Confirmed deletion removes the custom definition and retains built-in/New at 80×24.",
        "limits": "Project-owned asserted checkpoint; no pixel baseline comparison. Visual inspection documents the reached state, not all terminal sizes or themes.",
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
        "title": "Launch focus · attached",
        "file": "tests/e2e/artifacts/launch-focus/attached.png",
        "state": "attached",
        "observation": "The isolated Codex startup input screen is visible.",
        "limits": "Startup/attachment checkpoint only; no prompt or baseline-image comparison.",
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
        "file": "tests/e2e/artifacts/launch-focus/dashboard.png",
        "state": "dashboard",
        "observation": "The created session and older session are visible with session-pane actions after detach.",
        "limits": "Startup/attachment checkpoint only; no prompt or baseline-image comparison.",
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
        "id": "launch-focus-reattached",
        "title": "Launch focus · reattached",
        "file": "tests/e2e/artifacts/launch-focus/reattached.png",
        "state": "reattached",
        "observation": "The isolated Codex startup input screen is visible.",
        "limits": "Startup/attachment checkpoint only; no prompt or baseline-image comparison.",
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
      }
    ]
  }
};
