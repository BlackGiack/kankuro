package driver

import (
	"testing"
	"time"
)

func TestIsPromptLike(t *testing.T) {
	now := time.Now()
	if !IsPromptLike("Continue?", now, now) {
		t.Fatalf("expected prompt-like line")
	}
	if IsPromptLike("done", now.Add(-20*time.Second), now) {
		t.Fatalf("old output should not be prompt-like")
	}
}
