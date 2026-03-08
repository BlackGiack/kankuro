package driver

import (
	"kankuro/internal/config"
	"testing"
)

type stubDriver struct{}

func (s *stubDriver) Spawn(config AgentConfig) (*SessionHandle, error) { return nil, nil }
func (s *stubDriver) Kill() error                                      { return nil }
func (s *stubDriver) SetSkills(skills []config.Skill) error            { return nil }
func (s *stubDriver) SetMcps(mcps []config.MCPServer) error            { return nil }
func (s *stubDriver) SendInput(data string) error                      { return nil }
func (s *stubDriver) CaptureOutput() (chan []byte, error)              { return nil, nil }
func (s *stubDriver) GetStatus() SessionStatus                         { return StatusRunning }

func TestNewCLIDriverFallsBackToGeneric(t *testing.T) {
	driverConstructors = map[string]func() AgentDriver{}
	d := NewCLIDriver("unknown")
	if _, ok := d.(*GenericCLIDriver); !ok {
		t.Fatalf("expected generic fallback")
	}
}

func TestNewCLIDriverUsesRegisteredConstructor(t *testing.T) {
	driverConstructors = map[string]func() AgentDriver{}
	RegisterDriver("special", func() AgentDriver { return &stubDriver{} })
	d := NewCLIDriver("special")
	if _, ok := d.(*stubDriver); !ok {
		t.Fatalf("expected registered constructor")
	}
}
