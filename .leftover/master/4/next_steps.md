# Next Steps
1. If the residual buffer still appears when switching sessions, consider clearing the terminal on attach or adding an optional "drain for N ms" mode for detach.
2. If any CLI still ignores the working directory, verify whether it requires explicit flags and add driver-specific wrappers if needed.
3. Run a manual stress test with OpenCode + Gemini to confirm no regressions.
