package session

import (
	"encoding/json"
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
	Status          Status    `json:"status"`
}

func loadSessions(dir string) ([]*Data, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	var out []*Data
	for _, path := range matches {
		fileData, errRead := os.ReadFile(path)
		if errRead != nil {
			continue // skip corrupted
		}
		var s Data
		if errUnmarshal := json.Unmarshal(fileData, &s); errUnmarshal != nil {
			continue
		}
		out = append(out, &s)
	}
	return out, nil
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
