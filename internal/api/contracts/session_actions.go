package contracts

import "errors"

// DiagnosticCode gives the semantic class of a user-facing diagnostic.
type DiagnosticCode string

const (
	DiagnosticFailure     DiagnosticCode = "failure"
	DiagnosticCancelled   DiagnosticCode = "cancelled"
	DiagnosticPersistence DiagnosticCode = "persistence_warning"
	DiagnosticPartial     DiagnosticCode = "partial_effect"
	DiagnosticUncertain   DiagnosticCode = "uncertain_effect"
	DiagnosticInvalid     DiagnosticCode = "invalid_result"
)

// LifecycleEffect classifies the observed effect of a single activation attempt.
type LifecycleEffect string

const (
	NoEffect         LifecycleEffect = "no_effect"
	UsableActivation LifecycleEffect = "usable_activation"
	UncertainEffect  LifecycleEffect = "uncertain_effect"
)

// LifecycleResult is a detached value; Diagnostic is a user-facing message.
// NoEffect has a diagnostic and no identities. UsableActivation has both
// identities and an optional warning. UncertainEffect has a diagnostic and
// may contain a target session ID and observed PID, but no process identity.
type LifecycleResult struct {
	Effect              LifecycleEffect
	SessionID           string
	ProcessActivationID string
	PID                 int
	Diagnostic          string
	DiagnosticCode      DiagnosticCode
}

func (r LifecycleResult) Validate() error {
	if (r.Diagnostic == "") != (r.DiagnosticCode == "") {
		return errors.New("invalid lifecycle diagnostic")
	}
	switch r.Effect {
	case NoEffect:
		if (r.DiagnosticCode == DiagnosticFailure || r.DiagnosticCode == DiagnosticCancelled) && r.Diagnostic != "" && r.SessionID == "" &&
			r.ProcessActivationID == "" &&
			r.PID == 0 {
			return nil
		}
	case UsableActivation:
		if r.SessionID != "" && r.ProcessActivationID != "" &&
			(r.DiagnosticCode == "" || r.DiagnosticCode == DiagnosticPersistence || r.DiagnosticCode == DiagnosticFailure || r.DiagnosticCode == DiagnosticCancelled || r.DiagnosticCode == DiagnosticPartial) {
			return nil
		}
	case UncertainEffect:
		if (r.DiagnosticCode == DiagnosticUncertain || r.DiagnosticCode == DiagnosticInvalid) && r.Diagnostic != "" &&
			r.ProcessActivationID == "" {
			return nil
		}
	}
	return errors.New("invalid lifecycle result")
}

type LaunchSessionRequest struct {
	DashboardID, CardID string
	WorkspaceIDs        []string
}

type WorkspaceCommandKind string

const (
	WorkspaceIDE      WorkspaceCommandKind = "ide"
	WorkspaceTerminal WorkspaceCommandKind = "terminal"
)

type OpenSessionWorkspaceRequest struct {
	DashboardID, SessionID string
	Kind                   WorkspaceCommandKind
}

type ScopedSessionRequest struct{ DashboardID, SessionID string }
type RenameSessionRequest struct{ DashboardID, SessionID, Name, ExpectedProcessID string }
type RenameSessionResult struct {
	Applied, Persisted bool
	Diagnostic         string
	DiagnosticCode     DiagnosticCode
}
type StopSessionResult struct {
	ProcessTerminated, RecordPersisted bool
	Diagnostic                         string
	DiagnosticCode                     DiagnosticCode
}
type RemoveSessionResult struct {
	InMemoryRemoved, RecordDeleted, WorkspaceRemoved bool
	Diagnostic                                       string
	DiagnosticCode                                   DiagnosticCode
}

// OpenSessionWorkspaceResult distinguishes a started child from a pre-start failure.
type OpenSessionWorkspaceResult struct {
	Started        bool
	Diagnostic     string
	DiagnosticCode DiagnosticCode
}

// LaunchPreparation is a detached workspace draft for a scoped card.
type LaunchPreparation struct {
	Card            CardSummary
	WorkspaceID     string
	SubWorkspaceIDs []string
	Workspaces      ReadResult[[]WorkspaceOption]
}
