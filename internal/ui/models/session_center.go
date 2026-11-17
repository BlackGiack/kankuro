package models

import "github.com/BlackGiack/kankuro/internal/api/contracts"

// SessionCenterData is the selected session's detached panel state.
type SessionCenterData struct {
	SessionID string
	Reports   []contracts.SessionReport
	Omitted   int
	Workspace *contracts.SessionWorkspaceResult
	Warnings  []string
	Stale     bool
	Loading   bool
}

// CloneSessionCenterData detaches data retained by a presentation component.
func CloneSessionCenterData(data SessionCenterData) SessionCenterData {
	data.Reports = append([]contracts.SessionReport(nil), data.Reports...)
	data.Warnings = append([]string(nil), data.Warnings...)
	if data.Workspace != nil {
		workspace := *data.Workspace
		workspace.GitEntries = append([]contracts.GitEntry(nil), workspace.GitEntries...)
		workspace.Entries = append([]contracts.DirectoryEntry(nil), workspace.Entries...)
		workspace.Warnings = append([]string(nil), workspace.Warnings...)
		for i := range workspace.Entries {
			entry := &workspace.Entries[i]
			if entry.LinkCount != nil {
				v := *entry.LinkCount
				entry.LinkCount = &v
			}
			if entry.Owner != nil {
				v := *entry.Owner
				entry.Owner = &v
			}
			if entry.Group != nil {
				v := *entry.Group
				entry.Group = &v
			}
			if entry.SymlinkTarget != nil {
				v := *entry.SymlinkTarget
				entry.SymlinkTarget = &v
			}
		}
		data.Workspace = &workspace
	}
	return data
}
