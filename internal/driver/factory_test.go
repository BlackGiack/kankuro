package driver_test

import (
	"context"
	"errors"
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
)

var errStubSpawn = errors.New("stub spawn")

type stubDriver struct{}

func (s *stubDriver) Spawn(_ driver.AgentConfig) (*driver.SessionHandle, error) {
	return nil, errStubSpawn
}
func (s *stubDriver) PrepareWorkspace(_ context.Context, _ string) error { return nil }
func (s *stubDriver) Kill() error                                        { return nil }

func (s *stubDriver) SetAgentsMD(
	_ config.IDCard,
	_ []config.Persona,
	_ []config.MCPServer,
	_ string,
	_ driver.InstructionSyncMode,
) (bool, error) {
	return false, nil
}
func (s *stubDriver) SetSkills(_ []config.Skill, _ ...string) error {
	return nil
}
func (s *stubDriver) SetPersonas(_ []config.Persona, _ ...string) error {
	return nil
}
func (s *stubDriver) SelectMCPs(_ []config.MCPServer, _ ...string) error {
	return nil
}

func (s *stubDriver) SetModel(_ string) []string { return nil }
func (s *stubDriver) SetSession(id string) []string {
	return (&driver.GenericCLIDriver{}).SetSession(id)
}
func (s *stubDriver) SendInput(_ string) error             { return nil }
func (s *stubDriver) SendInputBytes(_ []byte) (int, error) { return 0, nil }
func (s *stubDriver) SubmitPrompt(_ context.Context, _, _, _ string) (driver.SubmissionEvidence, error) {
	return driver.SubmissionEvidence{State: driver.SubmissionNoBytes}, nil
}
func (s *stubDriver) GetStatus() driver.SessionStatus { return driver.StatusRunning }
func (s *stubDriver) GetAgentFolder() string          { return ".agents" }
func (s *stubDriver) GetInstructionFileName() string  { return "AGENTS.md" }

func TestNewRegistry(t *testing.T) {
	r := driver.NewRegistry()
	r.Register("special", func() driver.AgentDriver { return &stubDriver{} })

	d := r.Get("special")
	if _, ok := d.(*stubDriver); !ok {
		t.Fatal("expected registered constructor")
	}
}

func TestRegistryFallbackToGeneric(t *testing.T) {
	r := driver.NewRegistry()
	d := r.Get("unknown")
	if _, ok := d.(*driver.GenericCLIDriver); !ok {
		t.Fatal("expected GenericCLIDriver fallback")
	}
}

func TestBuiltinRegistry(t *testing.T) {
	r := driver.NewBuiltinRegistry()
	d := r.Get("claude")
	if _, ok := d.(*driver.ClaudeCodeDriver); !ok {
		t.Fatal("expected ClaudeCodeDriver from builtin registry")
	}
	codex := r.Get("codex")
	if _, ok := codex.(*driver.CodexCLIDriver); !ok {
		t.Fatal("expected CodexCLIDriver from builtin registry")
	}
}
