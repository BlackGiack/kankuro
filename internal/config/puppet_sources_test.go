package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
)

func TestListPuppetSources_EmptyDefault(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sources, err := repo.ListPuppetSources()
	if err != nil {
		t.Fatalf("list puppet sources: %v", err)
	}
	if len(sources) != 0 {
		t.Fatalf("expected 0 sources, got %d", len(sources))
	}
}

func TestSavePuppetSources_Persists(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	paths := []string{"/tmp/puppets/one", "/tmp/puppets/two"}
	if err = repo.SavePuppetSources(paths); err != nil {
		t.Fatalf("save puppet sources: %v", err)
	}
	got, err := repo.ListPuppetSources()
	if err != nil {
		t.Fatalf("list puppet sources: %v", err)
	}
	if len(got) != 2 || got[0] != "/tmp/puppets/one" || got[1] != "/tmp/puppets/two" {
		t.Fatalf("expected [one two], got %v", got)
	}
}

func TestSyncPuppetsFromSources_DiscoversValidFiles(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	sourceRoot := t.TempDir()
	crow := filepath.Join(sourceRoot, "crow.puppet")
	fox := filepath.Join(sourceRoot, "fox.puppet")
	if err = os.WriteFile(crow, []byte("  ^ ^  \n (o o) \n ( vvv )\n \\_____/"), 0o644); err != nil {
		t.Fatalf("write crow: %v", err)
	}
	if err = os.WriteFile(fox, []byte(" /\\_/\\ \n( -.- )\n >(T)< \n || || "), 0o644); err != nil {
		t.Fatalf("write fox: %v", err)
	}

	if err = repo.WritePuppetSourcesFile(config.PuppetSourcesFile{
		SchemaVersion: config.SchemaVersion,
		Paths:         []string{sourceRoot},
	}); err != nil {
		t.Fatalf("write sources: %v", err)
	}

	if err = repo.SyncPuppetsFromSources(); err != nil {
		t.Fatalf("sync: %v", err)
	}

	puppets, err := repo.ListPuppets()
	if err != nil {
		t.Fatalf("list puppets: %v", err)
	}

	builtinCount := 6
	if len(puppets) != builtinCount+2 {
		t.Fatalf("expected %d puppets (6 builtins + crow + fox), got %d", builtinCount+2, len(puppets))
	}

	found := map[string]bool{}
	for _, p := range puppets {
		found[p.Name] = true
		if p.Name == "crow" && len(p.Lines) != 4 {
			t.Errorf("crow should have 4 lines, got %d", len(p.Lines))
		}
		if p.Name == "fox" && len(p.Lines) != 4 {
			t.Errorf("fox should have 4 lines, got %d", len(p.Lines))
		}
	}
	if !found["crow"] || !found["fox"] {
		t.Error("expected crow and fox to be discovered")
	}
}

func TestSyncPuppetsFromSources_IgnoresHiddenAndNonPuppet(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	sourceRoot := t.TempDir()
	if err = os.WriteFile(filepath.Join(sourceRoot, ".hidden.puppet"), []byte("data"), 0o644); err != nil {
		t.Fatalf("write hidden: %v", err)
	}
	if err = os.WriteFile(filepath.Join(sourceRoot, "readme.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write txt: %v", err)
	}
	if err = os.MkdirAll(filepath.Join(sourceRoot, "subdir.puppet"), 0o755); err != nil {
		t.Fatalf("mkdir dir: %v", err)
	}

	if err = repo.WritePuppetSourcesFile(config.PuppetSourcesFile{
		SchemaVersion: config.SchemaVersion,
		Paths:         []string{sourceRoot},
	}); err != nil {
		t.Fatalf("write sources: %v", err)
	}

	if err = repo.SyncPuppetsFromSources(); err != nil {
		t.Fatalf("sync: %v", err)
	}

	puppets, err := repo.ListPuppets()
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(puppets) != 6 {
		t.Fatalf("expected 6 built-in puppets only, got %d", len(puppets))
	}
}

func TestSyncPuppetsFromSources_TruncatesOversized(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	sourceRoot := t.TempDir()
	tooManyLines := strings.Repeat("x\n", 7)
	tooWide := "1234567890123\n"

	if err = os.WriteFile(filepath.Join(sourceRoot, "tall.puppet"), []byte(tooManyLines), 0o644); err != nil {
		t.Fatalf("write tall: %v", err)
	}
	if err = os.WriteFile(filepath.Join(sourceRoot, "wide.puppet"), []byte(tooWide), 0o644); err != nil {
		t.Fatalf("write wide: %v", err)
	}

	if err = repo.WritePuppetSourcesFile(config.PuppetSourcesFile{
		SchemaVersion: config.SchemaVersion,
		Paths:         []string{sourceRoot},
	}); err != nil {
		t.Fatalf("write sources: %v", err)
	}

	if err = repo.SyncPuppetsFromSources(); err != nil {
		t.Fatalf("sync: %v", err)
	}

	puppets, err := repo.ListPuppets()
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	for _, p := range puppets {
		switch p.Name {
		case "tall":
			if len(p.Lines) != 5 {
				t.Errorf("tall: expected 5 lines, got %d", len(p.Lines))
			}
		case "wide":
			if len(p.Lines) != 1 || len(p.Lines[0]) > 11 {
				t.Errorf("wide: expected 1 line <=11 chars, got %d lines, first line %d chars", len(p.Lines), len(p.Lines[0]))
			}
		}
	}
}

func TestSyncPuppetsFromSources_BuiltinsAlwaysPresent(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	puppets, err := repo.ListPuppets()
	if err != nil {
		t.Fatalf("list puppets: %v", err)
	}

	if len(puppets) != 6 {
		t.Fatalf("expected 6 built-in puppets with no sources configured, got %d", len(puppets))
	}
}

func TestSyncPuppetsFromSources_CustomBuiltinCollision(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	sourceRoot := t.TempDir()
	colliding := filepath.Join(sourceRoot, "kankuro.puppet")
	if err = os.WriteFile(colliding, []byte("custom\n"), 0o644); err != nil {
		t.Fatalf("write collision: %v", err)
	}

	if err = repo.WritePuppetSourcesFile(config.PuppetSourcesFile{
		SchemaVersion: config.SchemaVersion,
		Paths:         []string{sourceRoot},
	}); err != nil {
		t.Fatalf("write sources: %v", err)
	}

	err = repo.SyncPuppetsFromSources()
	if err == nil {
		t.Fatal("expected error for collision, got nil")
	}
	if !strings.Contains(err.Error(), "kankuro") {
		t.Errorf("expected error to mention colliding name, got: %v", err)
	}

	puppets, err := repo.ListPuppets()
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	for _, p := range puppets {
		if p.Name == "kankuro" && !p.Builtin {
			t.Error("built-in kankuro should win over custom")
		}
	}
}

func TestSyncPuppetsFromSources_DuplicateCustomNames(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	srcA := t.TempDir()
	srcB := t.TempDir()
	if err = os.WriteFile(filepath.Join(srcA, "crow.puppet"), []byte("aaa\nbbb\n"), 0o644); err != nil {
		t.Fatalf("write srcA/crow: %v", err)
	}
	if err = os.WriteFile(filepath.Join(srcB, "crow.puppet"), []byte("ccc\nddd\n"), 0o644); err != nil {
		t.Fatalf("write srcB/crow: %v", err)
	}

	if err = repo.WritePuppetSourcesFile(config.PuppetSourcesFile{
		SchemaVersion: config.SchemaVersion,
		Paths:         []string{srcA, srcB},
	}); err != nil {
		t.Fatalf("write sources: %v", err)
	}

	err = repo.SyncPuppetsFromSources()
	if err == nil {
		t.Fatal("expected error for duplicate names, got nil")
	}
	if !strings.Contains(err.Error(), "crow") {
		t.Errorf("expected error to mention duplicate name, got: %v", err)
	}

	puppets, err := repo.ListPuppets()
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	count := 0
	for _, p := range puppets {
		if p.Name == "crow" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 crow puppet (no duplicates), got %d", count)
	}
}

func TestListPuppets_Sorting(t *testing.T) {
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	sourceRoot := t.TempDir()
	if err = os.WriteFile(filepath.Join(sourceRoot, "zebra.puppet"), []byte("z\n"), 0o644); err != nil {
		t.Fatalf("write zebra: %v", err)
	}
	if err = os.WriteFile(filepath.Join(sourceRoot, "aardvark.puppet"), []byte("a\n"), 0o644); err != nil {
		t.Fatalf("write aardvark: %v", err)
	}

	if err = repo.WritePuppetSourcesFile(config.PuppetSourcesFile{
		SchemaVersion: config.SchemaVersion,
		Paths:         []string{sourceRoot},
	}); err != nil {
		t.Fatalf("write sources: %v", err)
	}
	if err = repo.SyncPuppetsFromSources(); err != nil {
		t.Fatalf("sync: %v", err)
	}

	puppets, err := repo.ListPuppets()
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	builtinNames := []string{"cat", "darth", "default", "kankuro", "strawhat", "totoro"}
	customIdx := len(builtinNames)
	for i, name := range builtinNames {
		if puppets[i].Name != name {
			t.Errorf("expected builtin %q at index %d, got %q", name, i, puppets[i].Name)
		}
	}
	if len(puppets) != 8 {
		t.Fatalf("expected 8 puppets (6 builtins + aardvark + zebra), got %d", len(puppets))
	}
	rest := puppets[customIdx:]
	if rest[0].Name != "aardvark" || rest[1].Name != "zebra" {
		t.Errorf("expected custom puppets sorted alphabetically, got %v", rest)
	}
}
