package driver_test

import (
	"errors"
	"kankuro/internal/config"
	"kankuro/internal/driver"
	"testing"
)

var errStubSpawn = errors.New("stub spawn")

type stubDriver struct{}

func (s *stubDriver) Spawn(_ driver.AgentConfig) (*driver.SessionHandle, error) {
	return nil, errStubSpawn
}
func (s *stubDriver) Kill() error { return nil }
func (s *stubDriver) SetAgentsMD(_ config.IDCard, _ []config.Persona, _ []config.MCPServer, _ string, _ bool) error {
	return nil
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

func (s *stubDriver) SetModel(_ string) []string      { return nil }
func (s *stubDriver) SendInput(_ string) error        { return nil }
func (s *stubDriver) GetStatus() driver.SessionStatus { return driver.StatusRunning }

func TestNewCLIDriverFallsBackToGeneric(t *testing.T) {
	//nolint:reassign // This is a test resetting the global registry.
	driver.DriverConstructors = map[string]func() driver.AgentDriver{}
	d := driver.NewCLIDriver("unknown")
	if _, ok := d.(*driver.GenericCLIDriver); !ok {
		t.Fatalf("expected generic fallback")
	}
}

func TestNewCLIDriverUsesRegisteredConstructor(t *testing.T) {
	//nolint:reassign // This is a test resetting the global registry.
	driver.DriverConstructors = map[string]func() driver.AgentDriver{}
	driver.RegisterDriver("special", func() driver.AgentDriver { return &stubDriver{} })
	d := driver.NewCLIDriver("special")
	if _, ok := d.(*stubDriver); !ok {
		t.Fatalf("expected registered constructor")
	}
}
