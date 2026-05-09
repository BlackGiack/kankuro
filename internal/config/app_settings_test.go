package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"kankuro/internal/config"
)

func TestAppSettingsPersistence(t *testing.T) {
	dir := t.TempDir()
	repo, err := config.NewFileRepository(dir)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	cfg, err := repo.GetAppSettings()
	if err != nil {
		t.Fatalf("GetAppSettings failed: %v", err)
	}
	if cfg.IDECommand != "" {
		t.Fatalf("expected empty command, got %q", cfg.IDECommand)
	}
	if !cfg.ShowPersonasInCardEditor || !cfg.ShowMCPsInCardEditor {
		t.Fatal("expected card editor lists to be visible by default")
	}

	expected := config.AppSettings{
		IDECommand:               "code $PATH",
		ShowPersonasInCardEditor: false,
		ShowMCPsInCardEditor:     true,
	}
	if saveErr := repo.SaveAppSettings(expected); saveErr != nil {
		t.Fatalf("SaveAppSettings failed: %v", saveErr)
	}

	got, err := repo.GetAppSettings()
	if err != nil {
		t.Fatalf("GetAppSettings after save failed: %v", err)
	}
	if got != expected {
		t.Fatalf("expected settings %+v, got %+v", expected, got)
	}
}

func TestAppSettingsFileCreated(t *testing.T) {
	dir := t.TempDir()
	if _, err := config.NewFileRepository(dir); err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	path := filepath.Join(dir, config.AppSettingsFilename)
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		t.Fatalf("expected app settings file to be created, but it does not exist")
	}
}
