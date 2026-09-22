package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Data struct {
	ID              string    `json:"id"`
	CardID          string    `json:"card_id"`
	Name            string    `json:"name"`
	PID             int       `json:"pid"`
	WorkingDir      string    `json:"working_dir"`
	OriginalDir     string    `json:"original_dir"`
	UseTmpWorkspace bool      `json:"use_tmp_workspace"`
	StartedAt       time.Time `json:"started_at"`
	LastWordAt      time.Time `json:"last_word_at"`
	LastOutput      string    `json:"last_output,omitempty"`
	Status          Status    `json:"status"`
	ExitCode        int       `json:"exit_code,omitempty"`
}

// restore builds a fresh in-memory Session from a persisted Data record.
// Status is always forced to StatusTerminated: a process cannot survive a
// restart, so any loaded session is by definition not running.
func (d *Data) restore() *Session {
	return &Session{
		ID:              d.ID,
		CardID:          d.CardID,
		Name:            d.Name,
		PID:             d.PID,
		WorkingDir:      d.WorkingDir,
		OriginalDir:     d.OriginalDir,
		UseTmpWorkspace: d.UseTmpWorkspace,
		StartedAt:       d.StartedAt,
		LastWordAt:      d.LastWordAt,
		LastOutput:      d.LastOutput,
		Status:          StatusTerminated,
		ExitCode:        d.ExitCode,
		Subscribers:     make(map[chan []byte]struct{}),
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

func saveSession(dir string, s *Data) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(dir, s.ID+".json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func deleteSession(dir string, id string) error {
	return os.Remove(filepath.Join(dir, id+".json"))
}
