package session

import (
	"context"
	"errors"
	"strings"

	"github.com/BlackGiack/kankuro/internal/config"
)

// ActivationEffect describes what is known about an activation attempt.
type ActivationEffect uint8

const (
	ActivationNotPublished ActivationEffect = iota
	ActivationPublished
	ActivationUncertain
)

// ActivationResult is backend evidence captured during an activation attempt.
// Published means usable at publication, not necessarily alive at return or
// ready for input. ProcessActivationID exists only for a published process.
// SessionID is a target identity and alone does not establish an effect.
type ActivationResult struct {
	Effect              ActivationEffect
	SessionID           string
	ProcessActivationID string
	PID                 int
	WorkingDir          string
	Spawned             bool
	Err                 error
}

// LaunchInDashboardWithIDResult returns publication-time evidence through the
// same checked launch path as LaunchInDashboardWithIDContext.
func (m *Manager) LaunchInDashboardWithIDResult(
	ctx context.Context,
	sid, dashboardID string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
	check func([]Snapshot) error,
	name ...*string,
) ActivationResult {
	result := ActivationResult{SessionID: sid}
	switch {
	case strings.TrimSpace(dashboardID) == "":
		result.Err = errors.New("dashboard ID is required")
	case check == nil:
		result.Err = errors.New("dashboard publication check is required")
	default:
		result = m.launchWithIDContext(ctx, sid, dashboardID, card, skills, personas, mcps, check, name...)
	}
	return result
}

// RelaunchWithChoiceCheckedResult returns the historical effect of this attempt.
func (m *Manager) RelaunchWithChoiceCheckedResult(
	ctx context.Context,
	id string,
	card config.IDCard,
	choice ConversationChoice,
	check func([]Snapshot) error,
) ActivationResult {
	result := ActivationResult{SessionID: id}
	if check == nil {
		result.Err = errors.New("dashboard publication check is required")
	} else {
		result = m.relaunchWithChoiceContext(ctx, id, card, choice, check)
	}
	return result
}

// SwitchAgentWithChoiceCheckedResult returns the historical effect of this attempt.
func (m *Manager) SwitchAgentWithChoiceCheckedResult(
	ctx context.Context,
	id string,
	card config.IDCard,
	skills []config.Skill,
	personas []config.Persona,
	mcps []config.MCPServer,
	choice ConversationChoice,
	check func([]Snapshot) error,
) ActivationResult {
	result := ActivationResult{SessionID: id}
	if check == nil {
		result.Err = errors.New("dashboard publication check is required")
	} else {
		result = m.switchAgentWithChoiceContext(ctx, id, card, skills, personas, mcps, choice, check)
	}
	return result
}
