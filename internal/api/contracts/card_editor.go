package contracts

// Change distinguishes an unchanged field from an intentional zero value.
type Change[T any] struct {
	Present bool
	Value   T
}

// CardValues contains only fields that can be edited. Slice values are owned by
// the recipient and must be copied before retaining them across boundaries.
type CardValues struct {
	Name, Driver, Model, ReasoningEffort, WorkspaceID string
	SubWorkspaceIDs                                   []string
	FaceStyle, Puppet                                 string
	UseTmpWorkspace, DashboardControl, Shared         bool
	AgentsMD, AgentsMDPath                            string
	AgentsMDPreferPath                                bool
	SkillIDs, PersonaIDs, MCPIDs                      []string
}

type CardChanges struct {
	Name, Driver, Model, ReasoningEffort, WorkspaceID Change[string]
	SubWorkspaceIDs                                   Change[[]string]
	FaceStyle, Puppet                                 Change[string]
	UseTmpWorkspace, DashboardControl, Shared         Change[bool]
	AgentsMD, AgentsMDPath                            Change[string]
	AgentsMDPreferPath                                Change[bool]
	SkillIDs, PersonaIDs, MCPIDs                      Change[[]string]
}

type CardEditorData struct {
	DashboardID, OwnerDashboardID, CardID, Revision string
	CanEdit                                         bool
	Reason                                          string
	Values                                          CardValues
}

type CreateCardInput struct {
	DashboardID string
	Values      CardValues
}
type UpdateCardInput struct {
	DashboardID, CardID, ExpectedRevision string
	Changes                               CardChanges
}
type CardSaveStatus string

const (
	CardSaveSucceeded CardSaveStatus = "succeeded"
	CardSaveConflict  CardSaveStatus = "conflict"
)

type CardSaveResult struct {
	CardID, Revision string
	Status           CardSaveStatus
	Reason           string
}

// EditorOption excludes registry execution settings and secrets.
type EditorOption struct{ ID, Name, Description string }
type WorkspaceOption struct{ ID, Name, Path string }
type CardEditorChoices struct {
	Workspaces             ReadResult[[]WorkspaceOption]
	Skills, Personas, MCPs ReadResult[[]EditorOption]
	Puppets                ReadResult[[]PuppetChoice]
}
