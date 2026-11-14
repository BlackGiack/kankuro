package contracts

import "time"

type Availability struct {
	Enabled bool
	Reason  string
}

type DashboardState string

const (
	DashboardActive   DashboardState = "active"
	DashboardDeleting DashboardState = "deleting"
	DashboardUnknown  DashboardState = "unknown"
)

type DashboardSummary struct {
	ID, Name, Description              string
	State                              DashboardState
	Recoverable                        bool
	Open, Edit, Delete, Inspect, Retry Availability
}

type CardSummary struct {
	ID, OwnerDashboardID, Name, Driver, Model string
	Shared                                    bool
	SkillIDs, PersonaIDs                      []string
	Puppet, FaceStyle                         string
	DeletionState                             string
	Launch, Edit, Delete, Retry, Inspect      Availability
}

type SessionSummary struct {
	ID, DashboardID, CardID, Name                            string
	Status                                                   string
	StartedAt                                                time.Time
	PID                                                      int
	WorkingDir, OriginalDir                                  string
	TemporaryWorkspace                                       bool
	CardDeleted, CardUnavailable                             bool
	Registration, RegistrationPersistence, RegistrationError string
	ActiveConversationID, ActiveConversationName             string
	FaceStyle, Puppet                                        string
	Attach, Relaunch, Stop, Remove, Switch, Rename           Availability
	OpenIDE, OpenTerminal                                    Availability
}

type NamedChoice struct{ ID, Name string }

type PuppetChoice struct {
	ID, Name string
	Lines    []string
	Builtin  bool
}

type DisplayChoices struct {
	Skills, Personas []NamedChoice
	Puppets          []PuppetChoice
}

type ReadResult[T any] struct {
	Attempted bool
	Value     T
	Err       error
}

type DisplayChoicesResult struct {
	Skills   ReadResult[[]NamedChoice]
	Personas ReadResult[[]NamedChoice]
	Puppets  ReadResult[[]PuppetChoice]
}

// RegistryChangeToken is an opaque metadata-based invalidation hint. It is not
// a revision or authorization token. Its zero value means no baseline exists.
type RegistryChangeToken string

type DashboardInventory struct {
	Dashboard DashboardSummary
	Cards     []CardSummary
	Sessions  []SessionSummary
}
