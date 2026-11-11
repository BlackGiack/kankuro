# Verification — 2026-11-09 OpenCode fixture refresh

## Fixture and checks

Current OpenCode credentials are exported/imported through the installed CLI
and integration/value/active metadata is compared in memory. Credential values
are never printed or published. The conversation database is not copied.
Service configuration is excluded; the fixture uses and stops its own service
on a separate loopback port.

The requested model is openrouter/openrouter/free (Free Models Router).
Its visible model label is asserted before the task is submitted.
Every E2E Go source has the e2e build tag.

The capture-glyph preflight passed using tui-test Go and an installed
Noto Sans Symbols2 fallback for U+2B1D. It creates only a temporary PNG and runs
before model calls. Direct-process exit checks read State.Exited, as required
for Run processes.

Fresh ordinary tests passed with package-local atomic coverage: 7,897 /
10,760 statements (73.4%). Tagged vet and lint passed.
The prior cleanup's production build/vet/race results remain historical checks;
this refresh changes test support and documentation, not production behavior.

## Actual native run

    go test -tags=e2e ./tests/e2e -run '^TestNativeAgentJourney$' -count=1 -timeout=10m -v

Started 2026-11-09T22:38:19.351021609Z; finished 2026-11-09T22:42:26.268507608Z.
Result: failed; final stage: register.
Environment: Linux/amd64, OpenCode 2.2.6, tui-test Go v0.1.1, 120×40 cells,
isolated HOME/config/workspace/service and the configured OpenRouter credential.

The run launched and attached with the requested model, but no conversation binding was saved within the four-minute registration wait. Detach/restoration and resume were not reached.

Read-only native-session diagnostics during this attempt showed a completed
assistant turn without a tool call for the explicit registration task.
A prior router attempt loaded the registration skill, then stopped without
registering. Model inference was responding; this is not evidence of a failed
HTTP connection. The router may choose different backing models.

The current PNGs and result.json are embedded in data/test-evidence.json and
the normalized model. The failure capture preserves the terminal at the
registration checkpoint. No registration, Ctrl+K restoration, resume or
image-baseline comparison is claimed for this failed run.

The initial Big Pickle quota screenshot came from incomplete fixture setup.
It does not describe the user's account state and is superseded. The later
NVIDIA timeouts are also historical attempts; the current evidence uses the
router confirmed by the user.

## Provenance and assessments

Source fingerprint: 22ce60949b80ba3d6d306b85a4c88683bf040372efe7dbbff5d879030f24365a
Revision: WORKTREE.

The ordinary profile, native result and collected endpoint match. The snapshot
contains the raw profile, source hashes and actual execution metadata. Curated
ordinary results come from the run's JSONL events, separate from the failed E2E
result. No earlier profile was relabeled as current and no coverage delta is
claimed.

There are 28 features, 69 scenarios and 94 reviewed evidence records:
61 tested and 8 partial. Temporary-workspace
input/save is tested. The native journey remains partial with actual reached
assertions, implemented later checkpoints, and explicit limits.

Finalization passed reference/evidence validation and the all-route smoke
check. The refreshed baseline was accepted at 2026-11-09T22:46:01.092Z.
Eight Chrome captures covered feature, source, failure-image and coverage pages
in both themes; the feature and failure-image pages were visually inspected.
Those documentation checks are separate from terminal E2E proof.
The accepted impact review is retained in
.maintenance/opencode-fixture-review.json. Prior cleanup reviews remain history.
No commits were made.
