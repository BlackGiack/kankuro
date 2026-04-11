package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"kankuro/internal/config"
)

func TestIDEConfigPersistence(t *testing.T) {
	dir := t.TempDir()
	repo, err := config.NewFileRepository(dir)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	// Verify default (empty) config
	cfg, err := repo.GetIDEConfig()
	if err != nil {
		t.Fatalf("GetIDEConfig failed: %v", err)
	}
	if cfg.Command != "" {
		t.Fatalf("expected empty command, got %q", cfg.Command)
	}

	// Save and reload
	expected := config.IDEConfig{Command: "code $PATH"}
	if saveErr := repo.SaveIDEConfig(expected); saveErr != nil {
		t.Fatalf("SaveIDEConfig failed: %v", saveErr)
	}

	got, err := repo.GetIDEConfig()
	if err != nil {
		t.Fatalf("GetIDEConfig after save failed: %v", err)
	}
	if got.Command != expected.Command {
		t.Fatalf("expected command %q, got %q", expected.Command, got.Command)
	}
}

func TestIDEConfigFileCreated(t *testing.T) {
	dir := t.TempDir()
	if _, err := config.NewFileRepository(dir); err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	path := filepath.Join(dir, config.IDEConfigFilename)
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		t.Fatalf("expected IDE config file to be created, but it does not exist")
	}
}
