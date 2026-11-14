package api_test

import (
	"testing"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
)

func TestConfigurationMCPProtectionAndCopies(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := api.New(repo, nil)
	if _, err = service.SaveMCP(contracts.MCPEditor{Name: "kankuro", Command: "x"}); err == nil {
		t.Fatal("builtin shadow was accepted")
	}
	if err = service.DeleteConfiguredMCP(config.BuiltinMCPKankuro); err == nil {
		t.Fatal("builtin delete was accepted")
	}
	input := contracts.MCPEditor{Name: "test", Command: "test", Args: []string{"original"}, Env: map[string]string{"KEY": "value"}}
	saved, err := service.SaveMCP(input)
	if err != nil {
		t.Fatal(err)
	}
	input.Args[0] = "changed"
	input.Env["KEY"] = "changed"
	loaded, err := service.LoadMCPEditor(saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Args[0] != "original" || loaded.Env["KEY"] != "value" {
		t.Fatalf("input aliased: %+v", loaded)
	}
	loaded.Args[0] = "changed"
	loaded.Env["KEY"] = "changed"
	loaded, err = service.LoadMCPEditor(saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Args[0] != "original" || loaded.Env["KEY"] != "value" {
		t.Fatalf("output aliased: %+v", loaded)
	}
	rows, err := service.ReadMCPRows()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range rows {
		if row.ID == config.BuiltinMCPKankuro {
			found = true
			if row.Edit.Enabled || row.Delete.Enabled {
				t.Fatalf("builtin available: %+v", row)
			}
		}
	}
	if !found {
		t.Fatal("builtin missing")
	}
}

func TestConfigurationSettingsRoundTrip(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := api.New(repo, nil)
	want := contracts.Settings{IDECommand: "editor", TerminalCommand: "", ShowPersonasInCardEditor: false, ShowMCPsInCardEditor: true}
	if saveErr := service.SaveSettings(want); saveErr != nil {
		t.Fatal(saveErr)
	}
	got, err := service.ReadSettings()
	if err != nil || got != want {
		t.Fatalf("settings: %+v, %v", got, err)
	}
}
