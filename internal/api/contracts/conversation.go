package contracts

import "errors"

// ConversationOption contains detached display facts and an opaque selection ID.
type ConversationOption struct {
	SelectionID, NativeID, Name, Driver string
	Compatible                          bool
	Reason                              string
}

type ConversationPreparationRequest struct{ DashboardID, SessionID, CardID string }
type ConversationPreparation struct {
	DashboardID, SessionID, CardID string
	Availability                   Availability
	HasSaved                       bool
	Options                        []ConversationOption
	Diagnostic                     string
	DiagnosticCode                 DiagnosticCode
}

// Validate rejects contradictory or incomplete preparation data.
func (p ConversationPreparation) Validate() error {
	if (p.Diagnostic == "") != (p.DiagnosticCode == "") {
		return errors.New("invalid conversation diagnostic")
	}
	if p.Diagnostic != "" {
		if p.DiagnosticCode != DiagnosticFailure && p.DiagnosticCode != DiagnosticCancelled {
			return errors.New("invalid conversation diagnostic code")
		}
		return nil
	}
	if p.DashboardID == "" || p.SessionID == "" || p.CardID == "" || p.HasSaved != (len(p.Options) > 0) ||
		(!p.Availability.Enabled && p.Availability.Reason == "") {
		return errors.New("invalid conversation preparation")
	}
	seen := make(map[string]bool, len(p.Options))
	for _, option := range p.Options {
		if option.SelectionID == "" || seen[option.SelectionID] || (!option.Compatible && option.Reason == "") {
			return errors.New("invalid conversation option")
		}
		seen[option.SelectionID] = true
	}
	return nil
}

type ConversationChoiceKind string

const (
	ConversationNew    ConversationChoiceKind = "new"
	ConversationResume ConversationChoiceKind = "resume"
)

type ConversationChoice struct {
	Kind        ConversationChoiceKind
	SelectionID string
}
type RelaunchSessionRequest struct {
	DashboardID, SessionID, CardID string
	Choice                         ConversationChoice
}
type SwitchSessionRequest struct {
	DashboardID, SessionID, CardID string
	Choice                         ConversationChoice
}
