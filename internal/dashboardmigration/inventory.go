// Package dashboardmigration contains offline inspection for the dashboard layout migration.
package dashboardmigration

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	dashboardFileName           = "dashboards.yaml"
	tombstonesFileName          = "card_tombstones.yaml"
	dashboardTombstonesFileName = "dashboard_tombstones.yaml"
	defaultID                   = "default"
	jsonExtension               = ".json"
	privateFileMode             = 0o600
)

// Inventory summarizes legacy records without changing them.
type Inventory struct {
	Cards      int
	Sessions   int
	Operations int
	Messages   int
}

type legacyMessage struct {
	ID          string `json:"message_id"`
	DashboardID string `json:"dashboard_id"`
	Replies     []struct {
		ID string `json:"reply_id"`
	} `json:"replies"`
}

// InspectLegacy validates the membership-bearing portions of a legacy layout.
// It is read-only and does not substitute for the migration lock or backup.
func InspectLegacy(dir string) (Inventory, error) {
	var inventory Inventory
	for _, name := range []string{dashboardFileName, tombstonesFileName, dashboardTombstonesFileName, layoutMarker, layoutUsedMarker, "dashboard_deletions", "dashboard-migration.json"} {
		_, err := os.Lstat(filepath.Join(dir, name))
		if err == nil {
			return Inventory{}, fmt.Errorf("%s exists: layout is not unambiguously legacy", name)
		}
		if !errors.Is(err, os.ErrNotExist) {
			return Inventory{}, err
		}
	}
	if err := inspectCards(dir, &inventory); err != nil {
		return Inventory{}, err
	}
	if err := inspectRecords(filepath.Join(dir, "sessions"), recordKindSession, &inventory.Sessions); err != nil {
		return Inventory{}, err
	}
	if err := inspectSessionReferences(dir); err != nil {
		return Inventory{}, err
	}
	if err := inspectRecords(filepath.Join(dir, "operations"), "operation", &inventory.Operations); err != nil {
		return Inventory{}, err
	}
	if err := inspectMessages(dir, &inventory); err != nil {
		return Inventory{}, err
	}
	return inventory, nil
}

func inspectCards(dir string, inventory *Inventory) error {
	data, err := os.ReadFile(filepath.Join(dir, "id_cards.yaml"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	transformed, err := TransformCards(data)
	if err != nil {
		return fmt.Errorf("invalid card document: %w", err)
	}
	var file struct {
		Cards []map[string]any `yaml:"id_cards"`
	}
	if decodeErr := yaml.Unmarshal(transformed, &file); decodeErr != nil {
		return fmt.Errorf("parse transformed id_cards.yaml: %w", decodeErr)
	}
	inventory.Cards = len(file.Cards)
	return nil
}

const recordKindSession = "session"

//nolint:gocognit // All records must be checked before migration can proceed.
func inspectRecords(dir, kind string, count *int) error {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	seen := make(map[string]bool)
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != jsonExtension {
			continue
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("%s record %s is not a regular file", kind, entry.Name())
		}
		data, errAt109 := os.ReadFile(filepath.Join(dir, entry.Name()))
		if errAt109 != nil {
			return errAt109
		}
		if errAt113 := validateJSON(data); errAt113 != nil {
			return fmt.Errorf("invalid %s record %s: %w", kind, entry.Name(), errAt113)
		}
		var record map[string]json.RawMessage
		if errAt117 := json.Unmarshal(data, &record); errAt117 != nil || record == nil {
			return fmt.Errorf("invalid %s record %s", kind, entry.Name())
		}
		key := "id"
		if kind == "operation" {
			key = "operation_id"
		}
		var id string
		if errAt125 := json.Unmarshal(record[key], &id); errAt125 != nil || strings.TrimSpace(id) == "" || seen[id] {
			return fmt.Errorf("missing or duplicate %s ID in %s", kind, entry.Name())
		}
		if kind == recordKindSession && entry.Name() != id+jsonExtension {
			return fmt.Errorf("session ID does not match filename: %s", entry.Name())
		}
		if _, exists := record["schema_version"]; kind == recordKindSession && exists {
			return fmt.Errorf("session %s already has a schema version", entry.Name())
		}
		if raw, exists := record["dashboard_id"]; exists {
			var dashboard string
			if kind == recordKindSession || json.Unmarshal(raw, &dashboard) != nil || dashboard != defaultID {
				return fmt.Errorf("unexpected dashboard membership in %s", entry.Name())
			}
		} else if kind == "operation" {
			return fmt.Errorf("missing dashboard membership in %s", entry.Name())
		}
		seen[id] = true
		*count++
	}
	return nil
}

func inspectMessages(dir string, inventory *Inventory) error {
	data, err := os.ReadFile(filepath.Join(dir, "messaging", "history.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if errAt153 := validateJSON(data); errAt153 != nil {
		return fmt.Errorf("invalid message history: %w", errAt153)
	}
	var document map[string]json.RawMessage
	if decodeErr := json.Unmarshal(data, &document); decodeErr != nil || document == nil {
		return errors.New("message history must be an object")
	}
	var history struct {
		Messages             []legacyMessage   `json:"messages"`
		DashboardGenerations map[string]uint64 `json:"dashboard_generations"`
	}
	if errAt162 := json.Unmarshal(data, &history); errAt162 != nil {
		return fmt.Errorf("parse message history: %w", errAt162)
	}
	if validationErr := validateMessages(history.Messages); validationErr != nil {
		return validationErr
	}
	for id := range history.DashboardGenerations {
		if id != defaultID {
			return errors.New("message history has unexpected dashboard generation")
		}
	}
	inventory.Messages = len(history.Messages)
	return nil
}

func validateMessages(messages []legacyMessage) error {
	seenMessages := make(map[string]bool)
	seenReplies := make(map[string]bool)
	for _, message := range messages {
		if message.DashboardID != defaultID {
			return errors.New("message history has unexpected dashboard membership")
		}
		if message.ID == "" || seenMessages[message.ID] {
			return errors.New("message history has missing or duplicate message ID")
		}
		seenMessages[message.ID] = true
		for _, reply := range message.Replies {
			if reply.ID == "" || seenReplies[reply.ID] {
				return errors.New("message history has missing or duplicate reply ID")
			}
			seenReplies[reply.ID] = true
		}
	}
	return nil
}
