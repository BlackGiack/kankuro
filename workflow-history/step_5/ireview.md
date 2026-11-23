# Independent review — feature 5

Decision: **CHANGES REQUIRED. Not ready to advance to feature 6.**

Reviewed against `feature_5.md`, including the configuration contracts, source
identity checks, application operations, frontend commands/views and tests. Did
not read `impl_plan.md` or implementation-plan/review summaries. This file was
created because two changes are required. No implementation files changed and
no commit was made.

## Findings

1. **[P2] Reconcile a late initial settings load with the editor's baseline.**
   `internal/app/coordinator.go:494` applies SettingsLoadedMsg to SettingsView only
   when it is in display mode. If the user starts editing before the initial
   asynchronous load arrives, preserving the edited input is correct, but the
   view's untouched values and cancellation baseline remain the constructor
   defaults. SaveSettings writes the entire Settings value, so an edit to just
   IDECommand can then erase a stored TerminalCommand and reset the two existing
   visibility preferences. Cancelling instead also leaves the settings screen
   showing stale defaults. The coordinator has the successfully loaded values,
   but later SettingsView.Init does not reconcile them, so this is not just a
   transient display issue.

   Independent probes used the real configuration service and repository:
   store an existing IDE command, terminal command and false visibility settings;
   capture the initial settings read result; open Settings and edit only the IDE
   command; then deliver the delayed read and save. The persisted terminal command
   became empty and both preferences became true. A separate cancel probe also
   restored defaults rather than the successful read. The existing
   TestSettingsLateLoadPreservesDraft checks only the edited input and misses the
   loss of untouched values.

   Keep a current loaded baseline separate from user edits and merge a late read
   into untouched fields, preserving actual edits and using the loaded baseline
   on cancel. Alternatively, prevent editing/saving until the initial usable
   settings snapshot is available. Add regressions for both save and cancel after
   a delayed initial read, with non-default values in fields the user did not edit.

2. **[P2] Route both MCP deletion paths through the configuration capability.**
   `internal/app/legacy_actions.go:18` and `:28` invoke
   `a.service.DeleteConfiguredMCP` on a concrete `*api.Service`. The coordinator
   still calls these helpers for card-editor MCP deletion and MCP-manager
   deletion. `configurationPort` already declares DeleteConfiguredMCP, but these
   paths bypass it. Removing direct repository calls alone does not complete the
   specified configuration boundary: feature 5 requires a complete slice through
   consumer capabilities and no remaining registry bridge. The feature-7-owned
   concrete-service field exception for remaining attachment/deletion composition
   must not conceal unfinished MCP configuration operations.

   An independent probe injected a configurationPort overriding
   DeleteConfiguredMCP, drove MCPDeleteRequestedMsg through Coordinator.Update,
   and executed its returned command. The injected capability was never called;
   deletion went to the concrete legacy service. Route both helpers through
   c.configuration or a narrower consumer interface, capture it before command
   execution, and remove the MCP deletion methods from the concrete legacy path.
   Keep backend built-in/reference protections and existing request/scope guards.
   Add routing tests for both entry points with a fake configuration capability.

## Verification

Passed independently:

- `go test ./internal/config ./internal/api ./internal/app ./internal/ui/views ./internal/ui/components ./internal/architecture -count=1`
- `go test -race ./internal/config ./internal/api ./internal/app -run 'Source|MCP|Settings|Refresh|CommandClosures' -count=1`
- `go build ./...`
- `go vet ./...`
- `git diff --check`
- `golangci-lint run --new-from-rev=HEAD --whole-files=false` — zero issues

The additional delayed-load save/cancel probes and deletion-routing probe failed
with the results above. They ran using temporary Go overlays outside the
workspace; temporary probes were removed. The checked-in suite remains green
without these missing scenarios. Full lint was not rerun, and previously recorded
architecture-checker lint debt is not claimed resolved.

The source implementation validates path-list kind/list/position identities or
workspace ID/content identities inside repository updates. Tests cover duplicates,
reorder/conflict behavior, concurrent creates, cancellation, pre-save failure and
saved-but-sync-failed outcomes. MCP list rows omit connection maps/arguments, while
editor inputs retain them with copy boundaries. Built-in protections and parsing
tests pass. These positives do not resolve the two findings above.

No native-model or manual terminal E2E run was performed. Only this review file
was added; no commit was made.

## Readiness and blockers

- `ireview.md` **was created**.
- **Not ready to advance:** resolve findings 1–2 and verify the added regressions
  before final approval.
- No user question or environment blocker.
