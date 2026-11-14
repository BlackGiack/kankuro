package contracts

// SourceKind identifies a configuration source inventory.
type SourceKind string

const (
	SourceSkills     SourceKind = "skills"
	SourcePersonas   SourceKind = "personas"
	SourcePuppets    SourceKind = "puppets"
	SourceWorkspaces SourceKind = "workspaces"
)

// SourceRow is an owned snapshot. Identity must be supplied unchanged for edits and deletes.
type SourceRow struct {
	Identity string
	Path     string
	IsRoot   bool
}

// SourceMutation describes an explicit create, edit or delete. An empty identity creates.
type SourceMutation struct {
	Kind     SourceKind
	Identity string
	Path     string
	IsRoot   bool
	Delete   bool
}

// SourceMutationResult distinguishes a saved source list from subsequent sync failure.
type SourceMutationResult struct {
	Saved   bool
	Refresh bool
	Reload  bool
	Warning error
	Err     error
}

// SourceSyncResult describes an explicit registry synchronization, independently of source saves.
type SourceSyncResult struct {
	Refresh bool
	Err     error
}
