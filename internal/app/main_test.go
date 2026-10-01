package app_test

import (
	"os"
	"testing"
)

// TestMain points KANKURO_TMP_DIR at a throwaway directory for the whole
// package. Tests that launch with UseTmpWorkspace otherwise inherit the
// developer's KANKURO_TMP_DIR from the shell and leave a workspace copy in
// their real tmp root on every run. A test may still override it with
// t.Setenv.
func TestMain(m *testing.M) {
	os.Exit(runWithTmpDir(m))
}

func runWithTmpDir(m *testing.M) int {
	dir, err := os.MkdirTemp("", "kankuro-app-test-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	if err = os.Setenv("KANKURO_TMP_DIR", dir); err != nil {
		panic(err)
	}
	return m.Run()
}
