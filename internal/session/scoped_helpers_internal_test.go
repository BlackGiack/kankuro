package session

import (
	"context"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/google/uuid"
)

func testSetName(m *Manager, id, name string) error {
	return m.SetNameChecked(id, name, func([]Snapshot) error { return nil })
}

func testAttach(m *Manager, id string) (*AttachHandle, error) {
	return m.AttachChecked(id, func([]Snapshot) error { return nil })
}

func testLaunch(m *Manager, card config.IDCard, skills []config.Skill, personas []config.Persona, mcps []config.MCPServer) (string, error) {
	return testLaunchWithIDContext(m, context.Background(), uuid.NewString(), card, skills, personas, mcps)
}

//nolint:revive // Keep the original call shape for this test-only adapter.
func testLaunchWithIDContext(
	m *Manager,
	ctx context.Context,
	id string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
	name ...*string,
) (string, error) {
	if card.DashboardID == "" {
		card.DashboardID = "default"
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	release, err := m.ReserveCardReferenceForDashboardWithCancel(
		card.ID,
		card.DashboardID,
		id,
		cancel,
		func([]Snapshot) (bool, error) { return true, nil },
	)
	if err != nil {
		return "", err
	}
	defer release()
	return m.LaunchInDashboardWithIDContext(
		ctx,
		id,
		card.DashboardID,
		card,
		skills,
		personas,
		mcps,
		func([]Snapshot) error { return nil },
		name...)
}

func testRelaunch(m *Manager, id string, card config.IDCard) error {
	return testRelaunchWithChoice(m, id, card, ConversationChoice{})
}

func testRelaunchWithChoice(m *Manager, id string, card config.IDCard, choice ConversationChoice) error {
	return m.RelaunchWithChoiceCheckedContext(context.Background(), id, card, choice, func([]Snapshot) error { return nil })
}

func testSwitchAgent(
	m *Manager,
	id string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
) error {
	return testSwitchAgentWithChoice(m, id, card, skills, personas, mcps, ConversationChoice{})
}

func testSwitchAgentWithChoice(
	m *Manager,
	id string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
	choice ConversationChoice,
) error {
	return m.SwitchAgentWithChoiceCheckedContext(
		context.Background(),
		id,
		card,
		skills,
		personas,
		mcps,
		choice,
		func([]Snapshot) error { return nil },
	)
}
