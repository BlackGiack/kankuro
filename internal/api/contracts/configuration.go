package contracts

// Settings contains the values used by the application and its settings editor.
type Settings struct {
	IDECommand               string
	TerminalCommand          string
	ShowPersonasInCardEditor bool
	ShowMCPsInCardEditor     bool
}

// MCPRow contains display and availability facts, never connection credentials.
type MCPRow struct {
	ID      string
	Name    string
	Builtin bool
	Edit    Availability
	Delete  Availability
}

// MCPEditor contains editable connection data. Maps and slices belong to the recipient.
type MCPEditor struct {
	ID          string
	Name        string
	Command     string
	Args        []string
	Env         map[string]string
	CWD         string
	URL         string
	Headers     map[string]string
	Description string
}
