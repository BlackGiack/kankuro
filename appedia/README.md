# KANKURO Appedia

Open [index.html](index.html) directly in a browser. This snapshot describes
the reasoning-effort feature on frontend boundary roadmap commit `110865d`.

The card editor loads and persists an optional effort beside Model. Follow
[the activation flow](index.html#flow-reasoning-effort-activation) to see how the
saved value becomes CLI arguments on launch, relaunch and switch. The Tests
tab includes the model-free terminal journey that saves and reopens a card.

The selected-session center keeps its session details and puppet above reports
or workspace information. The `flow-preview` URL opens the current report-center
flow. The startup flow shows puppet sync completing before the UI reads registry
data and explains how a missing source becomes a startup warning. The center flows
cover report-only notifications, timestamp fitting and isolated Git status,
including repositories with split indexes.

The Tests tab links inspected assertions and recorded runs. Historical coverage
and terminal captures keep their original provenance. Live model recall, the
combined all-failure terminal recovery journey, and non-Linux terminal behavior
remain open.

## Updating the snapshot

Edit `data/content.js` for explanations, flows and assessments. Run the
create-appedia `appedia.mjs` workflow with `--action prepare --repo .`, review
the generated impact report, then build with `--action rebuild --repo .` and
accept with `--action finalize --repo . --jsdom <installation>`. Use
`--test-evidence <snapshot>` on rebuild and finalize to retain curated terminal
captures and bind current execution results to their source endpoint.

Architecture diagrams, the source model and `data/content-index.json` are
generated. Native model E2E runs require the approval described in
[tests/e2e/AGENTS.md](../tests/e2e/AGENTS.md).
