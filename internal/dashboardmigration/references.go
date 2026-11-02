package dashboardmigration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//nolint:gocognit // Validate each saved session and all of its bindings together.
func inspectSessionReferences(dir string) error {
	cards := make(map[string]bool)
	raw, err := os.ReadFile(filepath.Join(dir, "id_cards.yaml"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err == nil {
		var file struct {
			Cards []struct {
				ID string `yaml:"id"`
			} `yaml:"id_cards"`
		}
		if errAt25 := yaml.Unmarshal(raw, &file); errAt25 != nil {
			return errAt25
		}
		for _, card := range file.Cards {
			cards[card.ID] = true
		}
	}
	entries, err := os.ReadDir(filepath.Join(dir, "sessions"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != jsonExtension {
			continue
		}
		rawAt43, errAt43 := os.ReadFile(filepath.Join(dir, "sessions", entry.Name()))
		if errAt43 != nil {
			return errAt43
		}
		var record struct {
			CardID        string                     `json:"card_id"`
			Conversations map[string]json.RawMessage `json:"conversations"`
		}
		if errAt51 := json.Unmarshal(rawAt43, &record); errAt51 != nil {
			return errAt51
		}
		if !cards[record.CardID] {
			return fmt.Errorf("session %s references missing card %q", entry.Name(), record.CardID)
		}
		for id := range record.Conversations {
			if !cards[id] {
				return fmt.Errorf("session %s has binding for missing card %q", entry.Name(), id)
			}
		}
	}
	return nil
}
