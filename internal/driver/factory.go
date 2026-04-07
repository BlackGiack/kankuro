package driver

// NewCLIDriverFromRegistry creates a driver using an explicit registry.
func NewCLIDriverFromRegistry(reg Registry, driverName string) AgentDriver {
	return reg.Get(driverName)
}
