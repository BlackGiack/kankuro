# Files Touched
- internal/session/manager.go: added pending line cap constant and trim logic.
- internal/ui/attach.go: added stop channel to halt output writes immediately on detach.
- internal/session/manager_test.go: added pending line cap test.
- internal/driver/driver.go: set `PWD` env to working dir when spawning processes.
