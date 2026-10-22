# Feature 3: Session ancestry and subtree lifecycle

Status: canceled before implementation (2026-10-22). Skip this feature; feature 4
now depends directly on accepted feature 2. Feature numbering is retained.

## Decision

Factory membership will group agents and sessions. Each factory will have its own
orchestrator, responsible for the sessions in that factory. A separate persistent
parent/child session hierarchy is unnecessary for this model.

Cancel parent relationships, cycle/orphan handling, recursive session deletion,
and descendant-based deletion guards and activity indicators. Deleting an
orchestrator session does not implicitly delete the other sessions it manages.

## Requirements retained elsewhere

- Reuse ordinary sessions and existing card-scoped workspace preparation (feature 4).
- Keep temporary workspace storage independent per session.
- Preserve existing individual-session lifecycle/deletion guards and shutdown rules.
- Authenticate MCP callers and define their control scope (feature 4); authentication
  remains necessary without ancestry. Do not build a temporary ownership tree.
- Introduce factory membership and its orchestrator association in feature 7.
  Decide factory deletion and orchestrator replacement behavior there; do not
  silently translate the canceled recursive-deletion rule into factory deletion.

The user's workflow remains configurable. This cancellation does not introduce
new ticket types or require factories to be implemented before MCP session control.
