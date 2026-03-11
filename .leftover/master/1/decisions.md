# Decisions
1. Use Bubble Tea + creack/pty for MVP runtime.
- Rationale: Matches spec and gives interactive terminal wrapping quickly.
- Tradeoff: Terminal rendering is currently line-based and not full VT emulation.

2. Persist config as YAML files under user config dir.
- Rationale: Human-readable, easy manual editing, simple bootstrapping.
- Tradeoff: No migration framework yet beyond schema version field.

3. Keep generic driver as default fallback.
- Rationale: Supports unknown CLI tools immediately with minimal assumptions.
- Tradeoff: Provider-specific skill/MCP wiring remains TODO.

4. Introduce driver factory map for extensibility.
- Rationale: Allows per-driver constructors while keeping interface-based orchestration.
- Tradeoff: Global constructor map currently has no built-in specialized entries yet.

5. Refactor config package by subject.
- Rationale: Improve readability/maintainability as requested.
- Tradeoff: Slightly larger file count.
