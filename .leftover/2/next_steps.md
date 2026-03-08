# Next Steps
1. Add richer CRUD for sessions/config objects if needed, especially delete flows for ID cards, skills, and MCPs to match the new session cleanup controls.
2. Add UI-model tests for dashboard focus switching, session attach command dispatch, and kill/remove keyboard actions.
3. Decide whether detached background sessions should receive resize updates from the dashboard window or only on attach.
4. Consider whether terminated sessions should show exit code / last output summary before removal.
5. Reconcile `kankuro_specs.md` with any remaining implementation differences as behavior continues to evolve.

# Blockers / Open Questions
- No current blocker for the implemented PTY/session fixes.
- Open product question: should terminated sessions be auto-pruned after a timeout, or remain until explicit delete?
- Open UX question: do we want a dedicated visual cue for which dashboard pane is focused beyond the current border highlight?
