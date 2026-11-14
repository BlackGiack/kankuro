package ui

import (
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

func (r refresher) scopedFull(
	service dashboardInventoryReader,
	dashboardID string,
	generation uint64,
	requestID uint64,
	inventoryRequestID uint64,
) tea.Cmd {
	return func() tea.Msg {
		before, beforeErr := service.ReadRegistryChangeToken()
		data := r.loadGlobalConfigData()
		data.DashboardID, data.Generation, data.SessionEpoch = dashboardID, generation, r.sessionEpoch
		data.InventoryRequestID = inventoryRequestID
		data.Inventory, data.Choices, _ = readFocusedDashboard(service, dashboardID)
		choiceErr := displayChoiceError(data.Choices)
		after, afterErr := service.ReadRegistryChangeToken()
		tokenErr := errors.Join(tokenReadError("before", beforeErr), tokenReadError("after", afterErr))
		var certificationErr error
		var token contracts.RegistryChangeToken
		if beforeErr == nil && afterErr == nil && before == after && data.Err == nil &&
			data.Inventory.Err == nil && choiceErr == nil {
			token = after
		} else if beforeErr == nil && afterErr == nil && before != after {
			certificationErr = errors.New("registries changed during refresh")
		}
		data.RegistryErr = errors.Join(data.RegistryErr, choiceErr, tokenErr, certificationErr)
		data.Err = errors.Join(data.RegistryErr, data.Inventory.Err)
		return configFingerprintMsg{
			token: token, requestID: requestID, inventoryRequestID: inventoryRequestID, dashboardID: dashboardID,
			generation: generation, sessionEpoch: r.sessionEpoch, data: &data,
		}
	}
}

func (r refresher) scopedSessions(
	service dashboardInventoryReader,
	dashboardID string,
	generation uint64,
	inventoryRequestID uint64,
) tea.Cmd {
	return func() tea.Msg {
		result := messages.SessionsRefreshedMsg{
			DashboardID: dashboardID, Generation: generation, SessionEpoch: r.sessionEpoch,
			InventoryRequestID: inventoryRequestID,
		}
		if dashboardID == "" {
			return result
		}
		inventory, err := service.ReadDashboardOverview(dashboardID)
		result.Inventory.Attempted = true
		if err != nil {
			result.Err = err
			result.Inventory.Err = err
			return result
		}
		result.Inventory = contracts.ReadResult[contracts.DashboardInventory]{Attempted: true, Value: inventory}
		return result
	}
}

func (r refresher) scopedConfigIfChanged(
	service dashboardInventoryReader,
	dashboardID string,
	generation uint64,
	prev contracts.RegistryChangeToken,
	requestID uint64,
	inventoryRequestID uint64,
) tea.Cmd {
	return func() tea.Msg {
		before, beforeErr := service.ReadRegistryChangeToken()
		if beforeErr == nil && prev != "" && before == prev {
			return configFingerprintMsg{
				requestID: requestID, inventoryRequestID: inventoryRequestID, dashboardID: dashboardID,
				generation: generation, sessionEpoch: r.sessionEpoch,
			}
		}
		data := r.loadGlobalConfigData()
		data.DashboardID, data.Generation, data.SessionEpoch = dashboardID, generation, r.sessionEpoch
		data.InventoryRequestID = inventoryRequestID
		data.Inventory, data.Choices, _ = readFocusedDashboard(service, dashboardID)
		choiceErr := displayChoiceError(data.Choices)
		after, afterErr := service.ReadRegistryChangeToken()
		tokenErr := errors.Join(tokenReadError("before", beforeErr), tokenReadError("after", afterErr))
		var certificationErr error
		var token contracts.RegistryChangeToken
		if beforeErr == nil && afterErr == nil && before == after && data.Err == nil &&
			data.Inventory.Err == nil && choiceErr == nil {
			token = after
		} else if beforeErr == nil && afterErr == nil && before != after {
			certificationErr = errors.New("registries changed during refresh")
		}
		data.RegistryErr = errors.Join(data.RegistryErr, choiceErr, tokenErr, certificationErr)
		data.Err = errors.Join(data.RegistryErr, data.Inventory.Err)
		return configFingerprintMsg{
			token: token, requestID: requestID, inventoryRequestID: inventoryRequestID, dashboardID: dashboardID,
			generation: generation, sessionEpoch: r.sessionEpoch, data: &data,
		}
	}
}

func tokenReadError(stage string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("read registry change token %s load: %w", stage, err)
}

func displayChoiceError(choices contracts.DisplayChoicesResult) error {
	return errors.Join(choices.Skills.Err, choices.Personas.Err, choices.Puppets.Err)
}

func readFocusedInventory(
	service dashboardInventoryReader,
	dashboardID string,
) (contracts.ReadResult[contracts.DashboardInventory], error) {
	result := contracts.ReadResult[contracts.DashboardInventory]{}
	if dashboardID == "" {
		return result, nil
	}
	result.Attempted = true
	inventory, inventoryErr := service.ReadDashboardOverview(dashboardID)
	if inventoryErr == nil {
		result.Value = inventory
	} else {
		result.Err = inventoryErr
	}
	return result, inventoryErr
}

func readFocusedDashboard(
	service dashboardInventoryReader,
	dashboardID string,
) (contracts.ReadResult[contracts.DashboardInventory], contracts.DisplayChoicesResult, error) {
	inventory, inventoryErr := readFocusedInventory(service, dashboardID)
	if dashboardID == "" {
		return inventory, contracts.DisplayChoicesResult{}, inventoryErr
	}
	choices := service.ReadDisplayChoices()
	return inventory, choices, errors.Join(inventoryErr, choices.Skills.Err, choices.Personas.Err, choices.Puppets.Err)
}
