package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
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
	if cfg.IDECommand != "" || cfg.TerminalCommand != "" {
		t.Fatalf("expected empty command, got %q", cfg.IDECommand)
	}
	if !cfg.ShowPersonasInCardEditor || !cfg.ShowMCPsInCardEditor {
		t.Fatal("expected card editor lists to be visible by default")
	}

	expected := config.AppSettings{
		IDECommand:               "code $PATH",
		TerminalCommand:          "xterm \"$PATH\"",
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

// TestGetAppSettingsCoversEveryField is a regression guard against
// GetAppSettings re-declaring the settings schema by hand: it sets every
// field of the real AppSettings struct via reflection (so a field added
// later is automatically included, not silently skipped) to a non-default
// value, round-trips it through Save/Get, and requires an exact match. If
// GetAppSettings ever goes back to unmarshalling into a hand-written local
// struct, a newly added field falls out here.
func TestGetAppSettingsCoversEveryField(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	var want config.AppSettings
	v := reflect.ValueOf(&want).Elem()
	typ := v.Type()
	for i := range typ.NumField() {
		field := v.Field(i)
		switch {
		case field.Kind() == reflect.String:
			field.SetString(fmt.Sprintf("value-for-%s", typ.Field(i).Name))
		case field.Kind() == reflect.Bool:
			// Every bool field defaults to true today (DefaultAppSettings);
			// pin to false so a round-trip that silently falls back to the
			// default is caught.
			field.SetBool(false)
		default:
			t.Fatalf(
				"AppSettings field %q has kind %s: extend this test to cover it before relying on the round-trip guarantee",
				typ.Field(i).Name, field.Kind(),
			)
		}
	}

	if saveErr := repo.SaveAppSettings(want); saveErr != nil {
		t.Fatalf("SaveAppSettings failed: %v", saveErr)
	}
	got, err := repo.GetAppSettings()
	if err != nil {
		t.Fatalf("GetAppSettings failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetAppSettings dropped or mismatched a field: want %+v, got %+v", want, got)
	}
}

// TestGetAppSettingsDefaultsAbsentFields checks the other half of the
// contract: a field genuinely absent from the on-disk YAML (e.g. a file
// written before that field existed) must fall back to DefaultAppSettings,
// not zero out.
func TestGetAppSettingsDefaultsAbsentFields(t *testing.T) {
	dir := t.TempDir()
	repo, err := config.NewFileRepository(dir)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	path := filepath.Join(dir, config.AppSettingsFilename)
	partial := []byte("schema_version: 1\napp_settings:\n  ide_command: \"code $PATH\"\n")
	if writeErr := os.WriteFile(path, partial, 0o600); writeErr != nil {
		t.Fatalf("write partial settings: %v", writeErr)
	}

	got, err := repo.GetAppSettings()
	if err != nil {
		t.Fatalf("GetAppSettings failed: %v", err)
	}
	want := config.DefaultAppSettings()
	want.IDECommand = "code $PATH"
	if got != want {
		t.Fatalf("expected absent bool fields to default: want %+v, got %+v", want, got)
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
