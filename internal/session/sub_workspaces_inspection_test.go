package session_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlackGiack/kankuro/internal/session"
)

func TestReadSubWorkspaceDescriptorsDoesNotPrepareWorkspace(t *testing.T) {
	workingDir := t.TempDir()
	metadataDir := filepath.Join(workingDir, ".kankuro")
	if err := os.Mkdir(metadataDir, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema_version":1,"copies":{"project":{"workspace_id":"workspace-id","source_path":"/source/project"}}}`
	if err := os.WriteFile(filepath.Join(metadataDir, "sub-workspaces.json"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	descriptors, available := session.ReadSubWorkspaceDescriptors(workingDir)
	if !available || len(descriptors) != 1 {
		t.Fatalf("descriptors = %#v, available=%v", descriptors, available)
	}
	wantDestination := filepath.Join(workingDir, "tmp", "project")
	if descriptors[0].WorkspaceID != "workspace-id" ||
		descriptors[0].SourcePath != "/source/project" || descriptors[0].DestinationPath != wantDestination {
		t.Fatalf("descriptor = %#v", descriptors[0])
	}
	if _, err := os.Stat(filepath.Join(workingDir, "tmp")); !os.IsNotExist(err) {
		t.Fatalf("inspection prepared a destination directory (stat error %v)", err)
	}
	if err := os.WriteFile(filepath.Join(metadataDir, "sub-workspaces.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	corruptDescriptors, corruptAvailable := session.ReadSubWorkspaceDescriptors(workingDir)
	if corruptAvailable || len(corruptDescriptors) != 0 {
		t.Fatalf("corrupt manifest was treated as available: %#v", corruptDescriptors)
	}
}
