package contracts

// DeletionReference describes an affected session or pending activation.
type DeletionReference struct {
	DashboardID, SessionID  string
	Current, Saved, Pending bool
}

type DeletionSession struct {
	ID, Name, Status, Diagnostic                      string
	Stopped, Removed, RecordDeleted, TemporaryRemoved bool
}

type DeletionCard struct {
	ID, Name, Diagnostic string
	Shared, Deleted      bool
	References           []DeletionReference
}

type DeletionPending struct {
	CardID, DashboardID, SessionID, Diagnostic string
	Cleaned                                    bool
}

type DeletionPreparation struct {
	DashboardID, CardID, Name, Token string
	RequiresForeignConfirmation      bool
	Sessions                         []DeletionSession
	Cards                            []DeletionCard
	References                       []DeletionReference
	Pending                          []DeletionPending
}

type DeletionOutcome string

const (
	DeletionNoIntent       DeletionOutcome = "no_intent"
	DeletionPendingOutcome DeletionOutcome = "pending"
	DeletionCompleted      DeletionOutcome = "completed"
	DeletionUncertain      DeletionOutcome = "uncertain"
)

type DeletionProgress struct {
	DashboardID, CardID               string
	Outcome                           DeletionOutcome
	Diagnostic                        string
	AdmissionsDrained, HistoryRemoved bool
	Sessions                          []DeletionSession
	Cards                             []DeletionCard
	Pending                           []DeletionPending
}
