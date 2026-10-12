# KANKURO

KANKURO is a terminal dashboard for running and switching between AI coding-agent
CLI sessions.

## Install

Requires Go **1.26.1**. The Go command's default `GOTOOLCHAIN=auto` behavior can
automatically download and select the version required by this module.

```bash
go install github.com/BlackGiack/kankuro@latest
```

Go installs the executable into `GOBIN`, or the default Go bin directory when
`GOBIN` is unset. Add that directory to `PATH` to run `kankuro`.

## Local development

From the repository root:

```bash
go run .
go build .
go install .
```

For a manual run, isolate the app's data from your real user configuration:

```bash
KANKURO_TMP_DIR=/tmp/kankuro-test/tmp \
KANKURO_CONFIG_DIR=/tmp/kankuro-test/config \
  go run .
```
