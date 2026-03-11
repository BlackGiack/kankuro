# Commands
## Repo discovery and spec read
- `pwd && ls -la`
- `sed -n '1,260p' kankuro_specs.md`
- `rg --files -g '!**/.git/**'`
- `find . -maxdepth 4 -type f | sort`

## Implementation and refactor
- Created directories/files under `cmd/` and `internal/` via `mkdir` and heredoc `cat > file`.
- Applied targeted edits via `apply_patch`.
- Split config package files and added comments.
- Added driver factory and updated session manager to use `AgentDriver`.

## Formatting / dependency / verification
- `gofmt -w $(find cmd internal -name '*.go' | tr '\n' ' ')`
- `go mod tidy`
- `go test ./...`
- `go build ./...`

## Handover package
- `python /home/giack/.codex/skills/handover/scripts/init_handover.py --root /home/giack/Projects/kankuro`
- Created `.leftover/1/*` and replaced placeholders with concrete session context.

## Important outputs / side effects
- Tests are passing.
- Build is passing.
- Local binary artifact generated at `/home/giack/Projects/kankuro/cmd/kankuro/kankuro`.
