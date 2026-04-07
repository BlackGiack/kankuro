package config

import (
	"errors"
	"fmt"
)

// validateCardFields checks that mandatory card fields are set.
func validateCardFields(card IDCard) error {
	if card.Name == "" {
		return errors.New("id card name is required")
	}
	if card.Driver == "" {
		return errors.New("driver is required")
	}
	if card.WorkspaceID == "" {
		return errors.New("workspace is required")
	}
	return nil
}

// idSet builds a set of IDs from items.
func idSet[T any](items []T, getID func(T) string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, item := range items {
		set[getID(item)] = struct{}{}
	}
	return set
}

// validateRefs checks that all IDs in refs exist in the set.
func validateRefs(refs []string, set map[string]struct{}, typeName string) error {
	for _, id := range refs {
		if _, ok := set[id]; !ok {
			return fmt.Errorf("%s reference not found: %s", typeName, id)
		}
	}
	return nil
}
