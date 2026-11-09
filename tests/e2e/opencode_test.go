//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

const (
	openCodeTestModel     = "openrouter/openrouter/free"
	openCodeTestModelName = "Free Models Router"
)

type openCodeCredential struct {
	Active        bool   `json:"active"`
	IntegrationID string `json:"integrationID"`
	Value         any    `json:"value"`
}

// Use the installed CLI's supported transfer format, keeping secrets in memory.
// Copying auth.json alone misses credentials migrated into OpenCode's v2 database.
func copyOpenCodeCredentials(t *testing.T, environment map[string]string) {
	t.Helper()
	exported := openCodeCommand(t, nil, nil, "auth", "export")
	var original []openCodeCredential
	must(t, json.Unmarshal(exported, &original))
	// A standalone invocation closes its private server after import/export.
	openCodeCommand(t, environment, exported, "auth", "import", "--standalone")
	imported := openCodeCommand(t, environment, nil, "auth", "export", "--standalone")
	var copied []openCodeCredential
	must(t, json.Unmarshal(imported, &copied))
	configureOpenCodeService(t, environment)
	if !sameOpenCodeCredentials(original, copied) {
		t.Fatal("isolated OpenCode credential import differs from the configured integrations")
	}
}

// Compare by integration/value rather than IDs or ordering chosen by import.
// Never include credential material in assertion errors.
func sameOpenCodeCredentials(original, copied []openCodeCredential) bool {
	if len(original) != len(copied) {
		return false
	}
	remaining := append([]openCodeCredential(nil), copied...)
	for _, credential := range original {
		found := -1
		for i, candidate := range remaining {
			if reflect.DeepEqual(credential, candidate) {
				found = i
				break
			}
		}
		if found < 0 {
			return false
		}
		remaining = append(remaining[:found], remaining[found+1:]...)
	}
	return true
}

func openCodeCommand(t *testing.T, environment map[string]string, input []byte, arguments ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "opencode", arguments...)
	command.Env = os.Environ()
	for key, value := range environment {
		command.Env = append(command.Env, key+"="+value)
	}
	if home := environment["HOME"]; home != "" {
		command.Dir = home
	}
	command.Stdin = bytes.NewReader(input)
	// Do not capture stderr or print stdout: either can include exported secrets.
	output, err := command.Output()
	if err != nil {
		t.Fatalf("OpenCode %v failed: %v", arguments, err)
	}
	return output
}

func configureOpenCodeWorkspace(t *testing.T, workspace string) string {
	t.Helper()
	model := openCodeTestModel
	settings := map[string]any{"permission": "allow"}
	if model != "" {
		settings["model"] = model
	}
	encoded, err := json.Marshal(settings)
	must(t, err)
	must(t, os.WriteFile(filepath.Join(workspace, "opencode.json"), encoded, 0o600))
	return model
}

// Service endpoints are runtime state: a copied port can target the user's live server.
func copyOpenCodeConfiguration(t *testing.T, source, destination string) {
	t.Helper()
	entries, err := os.ReadDir(source)
	if os.IsNotExist(err) {
		return
	}
	must(t, err)
	for _, entry := range entries {
		if entry.Name() == "service.json" {
			continue
		}
		copyConfiguration(t, filepath.Join(source, entry.Name()), filepath.Join(destination, entry.Name()))
	}
}

func configureOpenCodeService(t *testing.T, environment map[string]string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	must(t, err)
	address, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatal("test listener did not have a TCP address")
	}
	port := strconv.Itoa(address.Port)
	must(t, listener.Close())
	openCodeCommand(t, environment, nil, "service", "set", "port", port)
	t.Cleanup(func() { openCodeCommand(t, environment, nil, "service", "stop") })
}
