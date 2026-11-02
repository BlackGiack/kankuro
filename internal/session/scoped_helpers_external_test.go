package session_test

import (
	"context"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

// These unit-test predicates exercise the checked manager entry points without
// requiring a service or transport; integration tests use persisted policy.
func testAttach(m *session.Manager, id string) (*session.AttachHandle, error) {
	return m.AttachChecked(id, func([]session.Snapshot) error { return nil })
}

func testSetName(m *session.Manager, id, name string) error {
	return m.SetNameChecked(id, name, func([]session.Snapshot) error { return nil })
}

func testLaunch(
	m *session.Manager,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
) (string, error) {
	return testLaunchWithIDContext(m, context.Background(), uuid.NewString(), card, skills, personas, mcps)
}

//nolint:revive // Keep the original call shape for this test-only adapter.
func testLaunchWithIDContext(
	m *session.Manager,
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
		func([]session.Snapshot) (bool, error) { return true, nil },
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
		func([]session.Snapshot) error { return nil },
		name...)
}

func testRelaunch(m *session.Manager, id string, card config.IDCard) error {
	return testRelaunchWithChoice(m, id, card, session.ConversationChoice{})
}

func testRelaunchWithChoice(m *session.Manager, id string, card config.IDCard, choice session.ConversationChoice) error {
	return m.RelaunchWithChoiceCheckedContext(context.Background(), id, card, choice, func([]session.Snapshot) error { return nil })
}

//nolint:unparam // Preserve test cases for the complete manager signature.
func testSwitchAgent(
	m *session.Manager,
	id string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
) error {
	return testSwitchAgentWithChoice(m, id, card, skills, personas, mcps, session.ConversationChoice{})
}

func testSwitchAgentWithChoice(
	m *session.Manager,
	id string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
	choice session.ConversationChoice,
) error {
	return m.SwitchAgentWithChoiceCheckedContext(
		context.Background(),
		id,
		card,
		skills,
		personas,
		mcps,
		choice,
		func([]session.Snapshot) error { return nil },
	)
}
