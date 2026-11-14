package models

import (
	"slices"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

// CardEditorDraft contains only editable values and the user's intended changes.
// The editor owns it; callers must copy its slices when retaining it.
type CardEditorDraft struct {
	Values  contracts.CardValues
	Changes contracts.CardChanges
}

type EditorSettings struct{ ShowPersonas, ShowMCPs bool }
type WorkspaceOption struct{ ID, Name, Path string }
type EditorOption struct{ ID, Name, Description string }
type PuppetOption struct {
	ID, Name string
	Lines    []string
	Builtin  bool
}
type EditorChoices struct {
	Workspaces             []WorkspaceOption
	Skills, Personas, MCPs []EditorOption
	Puppets                []PuppetOption
}

// CloneCardEditorDraft detaches mutable selections from an asynchronous command.
func CloneCardEditorDraft(draft CardEditorDraft) CardEditorDraft {
	draft.Values.SubWorkspaceIDs = slices.Clone(draft.Values.SubWorkspaceIDs)
	draft.Values.SkillIDs = slices.Clone(draft.Values.SkillIDs)
	draft.Values.PersonaIDs = slices.Clone(draft.Values.PersonaIDs)
	draft.Values.MCPIDs = slices.Clone(draft.Values.MCPIDs)
	draft.Changes.SubWorkspaceIDs.Value = slices.Clone(draft.Changes.SubWorkspaceIDs.Value)
	draft.Changes.SkillIDs.Value = slices.Clone(draft.Changes.SkillIDs.Value)
	draft.Changes.PersonaIDs.Value = slices.Clone(draft.Changes.PersonaIDs.Value)
	draft.Changes.MCPIDs.Value = slices.Clone(draft.Changes.MCPIDs.Value)
	return draft
}
