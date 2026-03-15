package driver_test

import (
	"kankuro/internal/driver"
	"testing"
	"time"
)

func TestIsPromptLike(t *testing.T) {
	now := time.Now()
	// Test immediate prompt detection
	if !driver.IsPromptLike("Continue?", now, now) {
		t.Fatalf("expected prompt-like line")
	}
	// Test idle word-silence detection (waiting after 15s)
	if !driver.IsPromptLike("done", now.Add(-20*time.Second), now) {
		t.Fatalf("old output SHOULD be prompt-like (waiting) due to word silence")
	}
	// Test that it's NOT prompt-like if recent words were seen and no prompt char
	if driver.IsPromptLike("processing", now.Add(-5*time.Second), now) {
		t.Fatalf("recent output without prompt char should not be prompt-like")
	}
}

func TestSetModel(t *testing.T) {
	d := driver.NewGenericCLIDriver()
	args := d.SetModel("opus")
	if len(args) != 2 || args[0] != "--model" || args[1] != "opus" {
		t.Fatalf("unexpected SetModel result: %v", args)
	}
}
