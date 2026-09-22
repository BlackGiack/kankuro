package driver

import (
	"strings"
	"testing"
)

func TestMergeEnvWithPwd(t *testing.T) {
	t.Setenv("KANKURO_TEST_EXISTING", "inherited")

	extra := map[string]string{
		"KANKURO_TEST_EXISTING": "overridden",
		"KANKURO_TEST_FRESH":    "new-value",
	}

	env := mergeEnvWithPwd(extra, "/tmp/workdir")

	counts := map[string]int{}
	values := map[string]string{}
	for _, entry := range env {
		k, v, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		counts[k]++
		values[k] = v
	}

	if counts["KANKURO_TEST_EXISTING"] != 1 {
		t.Fatalf("expected KANKURO_TEST_EXISTING to appear exactly once, got %d", counts["KANKURO_TEST_EXISTING"])
	}
	if values["KANKURO_TEST_EXISTING"] != "overridden" {
		t.Fatalf("expected override value to win, got %q", values["KANKURO_TEST_EXISTING"])
	}

	if counts["KANKURO_TEST_FRESH"] != 1 {
		t.Fatalf("expected fresh key to be appended once, got %d", counts["KANKURO_TEST_FRESH"])
	}
	if values["KANKURO_TEST_FRESH"] != "new-value" {
		t.Fatalf("expected fresh key value, got %q", values["KANKURO_TEST_FRESH"])
	}

	if counts["PWD"] != 1 {
		t.Fatalf("expected PWD to appear exactly once, got %d", counts["PWD"])
	}
	if values["PWD"] != "/tmp/workdir" {
		t.Fatalf("expected PWD to win with workdir value, got %q", values["PWD"])
	}
}
