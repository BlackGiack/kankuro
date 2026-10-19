package driver

import "strings"

// Registry maps driver names to constructors.
type Registry interface {
	Get(name string) AgentDriver
}

// MutableRegistry allows registration of drivers.
type MutableRegistry interface {
	Registry
	Register(name string, ctor func() AgentDriver)
}

// MapRegistry is an in-memory driver registry.
type MapRegistry struct {
	constructors map[string]func() AgentDriver
}

// NewRegistry creates an empty registry.
func NewRegistry() *MapRegistry {
	return &MapRegistry{
		constructors: make(map[string]func() AgentDriver),
	}
}

// NewBuiltinRegistry creates a registry preloaded with built-in drivers.
func NewBuiltinRegistry() *MapRegistry {
	r := NewRegistry()
	r.Register("claude", func() AgentDriver {
		return NewClaudeCodeDriver()
	})
	r.Register("codex", func() AgentDriver {
		return NewCodexCLIDriver()
	})
	return r
}

// Register adds a driver constructor to the registry.
func (r *MapRegistry) Register(name string, ctor func() AgentDriver) {
	if ctor == nil {
		return
	}
	r.constructors[strings.ToLower(strings.TrimSpace(name))] = ctor
}

// Get returns a driver by name, falling back to generic.
func (r *MapRegistry) Get(name string) AgentDriver {
	key := strings.ToLower(strings.TrimSpace(name))
	if ctor, ok := r.constructors[key]; ok {
		return ctor()
	}
	return NewGenericCLIDriver()
}
