# Verification — 2026-11-11 session rename refresh

The current source is commit `7a58943` (`feat: sessions rename`). This documentation
refresh adds the existing-session rename interaction, scoped persistence/result
handling, responsive modal/help, and nine supported test scenarios. Production
code and tests were not changed during this appedia update.

## Current execution

`go test -json ./...` passed. Per-test results were read from that run's JSONL and
bound to the fresh collector fingerprint. This ordinary suite excludes E2E-tagged
sources. The source inventory and relevant existing assertions were reassessed;
the moved terminal-action test body is unchanged, and its reviewed file hash was
updated. The nine rename scenarios include direct cancellation/scope-reset and
coordinator/service/disk/reload assertions.

The site's 29 features contain 78 scoped scenarios: 70 tested and eight partial.
The rename feature has nine supported scenarios and no partial/missing evidence.
Broader stop/remove/raw-terminal and native-agent boundaries retain their own
limits; these totals are not a statement or functional coverage percentage.

## Historical evidence

The embedded 2026-11-09 native-agent captures/result retain their original run,
endpoint fingerprint and failed-registration outcome. They were recovered from
the prior evidence inventory because the ignored source PNGs are absent in this
checkout. No new native-agent/model-driven run was performed, and detach/resume
assertions were not reached in that historical attempt.

The old atomic statement profile remains available with its original provenance;
its fingerprint differs from current head, so it is not current-head coverage.
No new statement profile or coverage delta was collected. Detailed historical
results are retained in
[.maintenance/native-agent-verification-2026-11-09.md](.maintenance/native-agent-verification-2026-11-09.md).

## Documentation checks

The accepted rename impact review is retained in
[.maintenance/session-rename-review.json](.maintenance/session-rename-review.json).
Reference/evidence validation and all-route smoke checks passed during finalize.
The updated baseline was accepted at 2026-11-11T21:20:40.117Z. All 114 initial impact
items were reviewed, followed by a source-stable association refinement. The
rename-save flow now links only to its five supported scenarios; broader partial
stop/remove/attach assessments stay on the manual-session flow.

Chrome captures covered the map, changed code, rename flow, feature, evidence
source and coverage pages in light and dark. Additional 900px captures checked
the rename flow/feature and retained historical failure image. Representative
captures were visually inspected. The coverage page correctly refuses to assign
the old profile to current head. Captures are under
`tmp/appedia-session-rename-shots/`. Browser screenshots of the documentation
site are separate from terminal-application E2E evidence.

`git diff --check` passed. No production source or test files changed.

No commits were made.
