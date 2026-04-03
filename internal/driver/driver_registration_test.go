package driver_test

import "kankuro/internal/driver"

//nolint:gochecknoinits // Tests need to mirror main.go driver registration.
func init() {
	driver.RegisterDriver("claude", func() driver.AgentDriver { return driver.NewClaudeCodeDriver() })
}
