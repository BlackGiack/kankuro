package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type SessionData struct {
	ID              string    `json:"id"`
	CardID          string    `json:"card_id"`
	Name            string    `json:"name"`
	PID             int       `json:"pid"`
	WorkingDir      string    `json:"working_dir"`
	OriginalDir     string    `json:"original_dir"`
	UseTmpWorkspace bool      `json:"use_tmp_workspace"`
	StartedAt       time.Time `json:"started_at"`
	LastWordAt      time.Time `json:"last_word_at"`
	Status          Status    `json:"status"`
}

func loadSessions(dir string) ([]*SessionData, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	var out []*SessionData
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue // skip corrupted
		}
		var s SessionData
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		out = append(out, &s)
	}
	return out, nil
}

func saveSession(dir string, s *SessionData) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, s.ID+".json")
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func deleteSession(dir string, id string) error {
	return os.Remove(filepath.Join(dir, id+".json"))
}
