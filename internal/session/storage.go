package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BlackGiack/kankuro/internal/fsutil"
)

const (
	sessionsDirMode = 0o700
	sessionFileMode = 0o600
)

type Data struct {
	ID                   string                           `json:"id"`
	CardID               string                           `json:"card_id"`
	Name                 string                           `json:"name"`
	PID                  int                              `json:"pid"`
	WorkingDir           string                           `json:"working_dir"`
	OriginalDir          string                           `json:"original_dir"`
	UseTmpWorkspace      bool                             `json:"use_tmp_workspace"`
	KankuroMCPEnabled    bool                             `json:"kankuro_mcp_enabled,omitempty"`
	InstructionsManaged  bool                             `json:"instructions_managed"`
	InstructionOwnership map[string]bool                  `json:"instruction_ownership,omitempty"`
	StartedAt            time.Time                        `json:"started_at"`
	LastWordAt           time.Time                        `json:"last_word_at"`
	LastOutput           string                           `json:"last_output,omitempty"`
	Status               Status                           `json:"status"`
	ExitCode             int                              `json:"exit_code,omitempty"`
	Conversations        map[string][]ConversationBinding `json:"conversations,omitempty"`
}

// restore builds a fresh in-memory Session from a persisted Data record.
// Status is always forced to StatusTerminated: a process cannot survive a
// restart, so any loaded session is by definition not running.
func (d *Data) restore() *Session {
	return &Session{
		ID:                      d.ID,
		CardID:                  d.CardID,
		Name:                    d.Name,
		PID:                     d.PID,
		WorkingDir:              d.WorkingDir,
		OriginalDir:             d.OriginalDir,
		UseTmpWorkspace:         d.UseTmpWorkspace,
		KankuroMCPEnabled:       d.KankuroMCPEnabled,
		InstructionsManaged:     d.InstructionsManaged,
		InstructionOwnership:    cloneInstructionOwnership(d.InstructionOwnership),
		StartedAt:               d.StartedAt,
		LastWordAt:              d.LastWordAt,
		LastOutput:              d.LastOutput,
		Status:                  StatusTerminated,
		Registration:            RegistrationUnavailable,
		RegistrationPersistence: RegistrationNotAttempted,
		ExitCode:                d.ExitCode,
		Conversations:           cloneConversations(d.Conversations),
		Subscribers:             make(map[chan []byte]struct{}),
	}
}

// loadSessions reads every persisted session file in dir. Unreadable or
// unparseable files are skipped rather than aborting the whole load (one bad
// file must not block startup), but every skip is recorded and returned as a
// joined error so the caller can surface it once.
func loadSessions(dir string) ([]*Data, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	var out []*Data
	var errs []error
	for _, path := range matches {
		fileData, errRead := os.ReadFile(path)
		if errRead != nil {
			errs = append(errs, fmt.Errorf("read session file %s: %w", path, errRead))
			continue
		}
		var s Data
		if errUnmarshal := json.Unmarshal(fileData, &s); errUnmarshal != nil {
			errs = append(errs, fmt.Errorf("parse session file %s: %w", path, errUnmarshal))
			continue
		}
		out = append(out, &s)
	}
	return out, errors.Join(errs...)
}

// saveSession writes the record atomically (see [fsutil.WriteFileAtomic]): a
// crash mid-write must not truncate the only pointer to a session's name and
// temporary workspace.
func saveSession(dir string, s *Data) error {
	if err := os.MkdirAll(dir, sessionsDirMode); err != nil {
		return err
	}
	path := filepath.Join(dir, s.ID+".json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session %s: %w", s.ID, err)
	}
	return fsutil.WriteFileAtomic(path, data, sessionFileMode)
}

func deleteSession(dir string, id string) error {
	return os.Remove(filepath.Join(dir, id+".json"))
}
