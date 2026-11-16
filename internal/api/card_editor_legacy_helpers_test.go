//nolint:testpackage // Tests exercise package-private revision and service fixtures.
package api

import (
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
)

func createCardForTest(s *Service, dashboard string, c config.IDCard) (config.IDCard, error) {
	result, err := s.CreateEditorCard(contracts.CreateCardInput{DashboardID: dashboard, Values: cardValues(c)})
	if err != nil {
		return config.IDCard{}, err
	}
	return s.loadCard(result.CardID)
}
func updateCardForTest(s *Service, dashboard string, expected, updated config.IDCard) (config.IDCard, error) {
	revision, err := cardRevision(expected)
	if err != nil {
		return config.IDCard{}, err
	}
	changes := contracts.CardChanges{
		Shared: contracts.Change[bool]{Present: true, Value: updated.Shared},
		Name:   contracts.Change[string]{Present: true, Value: updated.Name},
	}
	result, err := s.UpdateEditorCard(
		contracts.UpdateCardInput{DashboardID: dashboard, CardID: expected.ID, ExpectedRevision: revision, Changes: changes},
	)
	if err != nil {
		return config.IDCard{}, err
	}
	return s.loadCard(result.CardID)
}
