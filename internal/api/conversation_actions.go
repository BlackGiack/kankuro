package api

import (
	"context"
	"errors"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/session"
)

// ReadConversationPreparation returns detached options; execution revalidates
// the chosen identity under activation admission.
func (s *Service) ReadConversationPreparation(
	ctx context.Context,
	request contracts.ConversationPreparationRequest,
) (contracts.ConversationPreparation, error) {
	result := contracts.ConversationPreparation{DashboardID: request.DashboardID, SessionID: request.SessionID, CardID: request.CardID}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	err := s.sessions.WithAdmissionSnapshot(func(admission session.AdmissionSnapshot) error {
		var target *session.Snapshot
		for i := range admission.Sessions {
			if admission.Sessions[i].ID == request.SessionID {
				target = &admission.Sessions[i]
				break
			}
		}
		if target == nil || target.DashboardID != request.DashboardID {
			return errors.New("session is outside dashboard scope")
		}
		if result.CardID == "" {
			result.CardID = target.CardID
		}
		card, available, err := s.availableDashboardCard(request.DashboardID, result.CardID, admission.Sessions)
		if err != nil {
			return err
		}
		switch {
		case !available:
			result.Availability = disabled("card or dashboard is unavailable")
		case target.Status != session.StatusTerminated:
			result.Availability = disabled("session is already active")
		default:
			result.Availability = enabled()
		}
		bindings := target.Conversations[result.CardID]
		result.HasSaved = len(bindings) > 0
		result.Options = make([]contracts.ConversationOption, 0, len(bindings))
		for _, binding := range bindings {
			option := contracts.ConversationOption{
				SelectionID: session.ConversationOptionIdentity(target.ID, result.CardID, binding),
				NativeID:    binding.ID,
				Name:        binding.Name,
				Driver:      binding.Driver,
				Compatible:  binding.Driver == card.Driver,
			}
			if !option.Compatible {
				option.Reason = "conversation belongs to a different driver"
			}
			result.Options = append(result.Options, option)
		}
		return nil
	})
	return result, err
}

// ReadConversationPreparationResult maps read failures to detached diagnostics.
func (s *Service) ReadConversationPreparationResult(
	ctx context.Context,
	request contracts.ConversationPreparationRequest,
) contracts.ConversationPreparation {
	result, err := s.ReadConversationPreparation(ctx, request)
	if err != nil {
		result.Diagnostic = err.Error()
		result.DiagnosticCode = diagnosticCode(err)
	}
	return result
}

func sessionChoice(choice contracts.ConversationChoice) (session.ConversationChoice, error) {
	switch choice.Kind {
	case contracts.ConversationNew:
		if choice.SelectionID == "" {
			return session.ConversationChoice{New: true}, nil
		}
	case contracts.ConversationResume:
		if choice.SelectionID != "" {
			return session.ConversationChoice{SelectionIdentity: choice.SelectionID}, nil
		}
	}
	return session.ConversationChoice{}, errors.New("invalid conversation choice")
}

func (s *Service) RelaunchDashboardSession(ctx context.Context, request contracts.RelaunchSessionRequest) contracts.LifecycleResult {
	if request.CardID == "" {
		return lifecycleResult(session.ActivationResult{}, errors.New("expected card is required"))
	}
	choice, err := sessionChoice(request.Choice)
	if err != nil {
		return lifecycleResult(session.ActivationResult{}, err)
	}
	var evidence session.ActivationResult
	err = s.relaunchInDashboardEvidence(ctx, request.DashboardID, request.SessionID, request.CardID, choice, &evidence)
	return lifecycleResult(evidence, err)
}

func (s *Service) SwitchSession(ctx context.Context, request contracts.SwitchSessionRequest) contracts.LifecycleResult {
	choice, err := sessionChoice(request.Choice)
	if err != nil {
		return lifecycleResult(session.ActivationResult{}, err)
	}
	var evidence session.ActivationResult
	err = s.switchInDashboardEvidence(ctx, request.DashboardID, request.SessionID, request.CardID, choice, &evidence)
	return lifecycleResult(evidence, err)
}
