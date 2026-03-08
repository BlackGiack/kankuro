package driver

import "strings"

var driverConstructors = map[string]func() AgentDriver{}

// RegisterDriver binds a driver name to a concrete AgentDriver constructor.
func RegisterDriver(name string, ctor func() AgentDriver) {
	if ctor == nil {
		return
	}
	driverConstructors[strings.ToLower(strings.TrimSpace(name))] = ctor
}

// NewCLIDriver returns a dedicated driver implementation when registered, otherwise generic.
func NewCLIDriver(driver string) AgentDriver {
	if ctor, ok := driverConstructors[strings.ToLower(strings.TrimSpace(driver))]; ok {
		return ctor()
	}
	return NewGenericCLIDriver()
}
