package layout

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
)

// ValidateLayout reads the membership-bearing v2 layout without normalizing it.
//
//nolint:gocognit,gocyclo,cyclop,funlen // Validate all membership domains before startup.
func ValidateLayout(dir string) error {
	marker, err := os.ReadFile(filepath.Join(dir, layoutMarker))
	if err != nil {
		return errors.New("dashboard layout marker is missing or unsupported")
	}
	if string(marker) != currentMarker {
		if string(marker) != "2\n" {
			return errors.New("unsupported dashboard layout marker")
		}
		used, usedErr := os.ReadFile(filepath.Join(dir, layoutUsedMarker))
		if usedErr != nil || string(used) != "used\n" {
			return errors.New("legacy migration may be incomplete; run migrate-dashboards -apply")
		}
	}
	for _, name := range []string{"id_cards.yaml", config.DashboardsFilename, config.CardTombstonesFilename, config.DashboardTombstonesFilename} {
		info, statErr := os.Lstat(filepath.Join(dir, name))
		if statErr != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("required dashboard file %s is missing or invalid", name)
		}
	}
	repo, err := config.OpenExistingFileRepository(dir)
	if err != nil {
		return err
	}
	dashboards, err := repo.ListDashboards()
	if err != nil {
		return err
	}
	deletions, err := repo.ListDashboardDeletions()
	if err != nil {
		return err
	}
	dashboardSet := make(map[string]bool)
	deleting := make(map[string]bool)
	for _, dashboard := range dashboards {
		dashboardSet[dashboard.ID] = true
		if dashboard.State == config.DashboardDeleting {
			deleting[dashboard.ID] = true
		}
	}
	for _, deletion := range deletions {
		dashboardSet[deletion.DashboardID] = true
		deleting[deletion.DashboardID] = true
	}
	for id := range deleting {
		found := false
		for _, deletion := range deletions {
			if deletion.DashboardID == id {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("deleting dashboard %s has no recovery manifest", id)
		}
	}
	dashboardTombstones, err := repo.ListDashboardTombstones()
	if err != nil {
		return err
	}
	historicalDashboards := make(map[string]bool)
	for _, item := range dashboardTombstones {
		if dashboardSet[item.ID] && !deleting[item.ID] {
			return errors.New("active dashboard conflicts with deleted identity")
		}
		historicalDashboards[item.ID] = true
	}
	cardFile, err := repo.LoadIDCardsFile()
	if err != nil {
		return err
	}
	if cardFile.SchemaVersion != config.CardSchemaVersion {
		return errors.New("dashboard card schema is not migrated")
	}
	cards := make(map[string]config.IDCard)
	for _, card := range cardFile.IDCards {
		if !dashboardSet[card.DashboardID] {
			return fmt.Errorf("card %s has unknown dashboard", card.ID)
		}
		cards[card.ID] = card
	}
	tombstones, err := repo.ListCardTombstones()
	if err != nil {
		return err
	}
	deletedCards := make(map[string]bool)
	for _, item := range tombstones {
		if card, exists := cards[item.ID]; exists && (item.State == "deleted" || card.DashboardID != item.DashboardID) {
			return fmt.Errorf("card %s conflicts with tombstone", item.ID)
		}
		deletedCards[item.ID] = true
	}
	if errAt105 := validateLayoutSessions(dir, dashboardSet, cards, deletedCards); errAt105 != nil {
		return errAt105
	}
	return validateLayoutHistory(dir, dashboardSet, historicalDashboards)
}

//nolint:gocognit // Validate each persisted session and all saved card references.
func validateLayoutSessions(dir string, dashboards map[string]bool, cards map[string]config.IDCard, tombstones map[string]bool) error {
	entries, err := os.ReadDir(filepath.Join(dir, "sessions"))
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
			return fmt.Errorf("session %s is not a regular file", entry.Name())
		}
		raw, readErr := os.ReadFile(filepath.Join(dir, "sessions", entry.Name()))
		if readErr != nil {
			return readErr
		}
		if validateErr := validateJSON(raw); validateErr != nil {
			return fmt.Errorf("invalid session %s: %w", entry.Name(), validateErr)
		}
		var item session.Data
		if decodeErr := json.Unmarshal(raw, &item); decodeErr != nil {
			return fmt.Errorf("invalid session %s: %w", entry.Name(), decodeErr)
		}
		if item.SchemaVersion != 2 || item.ID == "" || entry.Name() != item.ID+jsonExtension || seen[item.ID] ||
			!dashboards[item.DashboardID] {
			return fmt.Errorf("invalid session identity or dashboard in %s", entry.Name())
		}
		seen[item.ID] = true
		references := make([]string, 0, len(item.Conversations)+1)
		references = append(references, item.CardID)
		for id := range item.Conversations {
			references = append(references, id)
		}
		for _, id := range references {
			card, exists := cards[id]
			if !exists && !tombstones[id] {
				return fmt.Errorf("session %s references unknown card %s", item.ID, id)
			}
			if exists && !tombstones[id] && !card.Shared && card.DashboardID != item.DashboardID {
				return fmt.Errorf("session %s references foreign exclusive card %s", item.ID, id)
			}
		}
	}
	return nil
}

//nolint:gocognit // Validate both durable history formats before startup.
func validateLayoutHistory(dir string, dashboards, historical map[string]bool) error {
	entries, err := os.ReadDir(filepath.Join(dir, "operations"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	seen := make(map[string]bool)
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != jsonExtension {
			continue
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("operation %s is not a regular file", entry.Name())
		}
		raw, readErr := os.ReadFile(filepath.Join(dir, "operations", entry.Name()))
		if readErr != nil {
			return readErr
		}
		if validateErr := validateJSON(raw); validateErr != nil {
			return validateErr
		}
		var operation struct {
			ID          string `json:"operation_id"`
			DashboardID string `json:"dashboard_id"`
		}
		if decodeErr := json.Unmarshal(raw, &operation); decodeErr != nil {
			return decodeErr
		}
		if operation.ID == "" || seen[operation.ID] || (!dashboards[operation.DashboardID] && !historical[operation.DashboardID]) {
			return fmt.Errorf("invalid operation scope in %s", entry.Name())
		}
		seen[operation.ID] = true
	}
	raw, err := os.ReadFile(filepath.Join(dir, "messaging", "history.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if errAt201 := validateJSON(raw); errAt201 != nil {
		return errAt201
	}
	var history struct {
		Messages []struct {
			DashboardID string `json:"dashboard_id"`
		} `json:"messages"`
		Generations map[string]uint64 `json:"dashboard_generations"`
	}
	if errAt210 := json.Unmarshal(raw, &history); errAt210 != nil {
		return errAt210
	}
	for _, message := range history.Messages {
		if !dashboards[message.DashboardID] {
			return errors.New("message references unknown or deleted dashboard")
		}
	}
	for id := range history.Generations {
		if !dashboards[id] && !historical[id] {
			return errors.New("message cursor references unknown dashboard")
		}
	}
	return nil
}
