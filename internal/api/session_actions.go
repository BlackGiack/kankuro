package api

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/google/uuid"
)

// LaunchSession publishes one scoped activation and returns its captured effect.
func (s *Service) LaunchSession(ctx context.Context, request contracts.LaunchSessionRequest) contracts.LifecycleResult {
	var evidence session.ActivationResult
	_, err := s.launchDashboardWithIDEvidence(
		ctx,
		request.DashboardID,
		request.CardID,
		uuid.NewString(),
		nil,
		request.WorkspaceIDs,
		&evidence,
	)
	return lifecycleResult(evidence, err)
}

func lifecycleResult(evidence session.ActivationResult, err error) contracts.LifecycleResult {
	result := contracts.LifecycleResult{}
	if err != nil {
		result.Diagnostic = err.Error()
		result.DiagnosticCode = diagnosticCode(err)
	}
	switch evidence.Effect {
	case session.ActivationPublished:
		result.Effect = contracts.UsableActivation
		result.SessionID = evidence.SessionID
		result.ProcessActivationID = evidence.ProcessActivationID
		result.PID = evidence.PID
	case session.ActivationUncertain:
		result.Effect = contracts.UncertainEffect
		if err != nil {
			result.DiagnosticCode = contracts.DiagnosticUncertain
		}
		result.SessionID = evidence.SessionID
		result.PID = evidence.PID
	case session.ActivationNotPublished:
		result.Effect = contracts.NoEffect
	default:
		return contracts.LifecycleResult{
			Effect:     contracts.UncertainEffect,
			SessionID:  evidence.SessionID,
			PID:        evidence.PID,
			Diagnostic: "unknown activation evidence", DiagnosticCode: contracts.DiagnosticInvalid,
		}
	}
	if err == nil && result.Effect != contracts.UsableActivation {
		result.Diagnostic = "activation did not produce a usable process"
		result.DiagnosticCode = contracts.DiagnosticFailure
	}
	if validationErr := result.Validate(); validationErr != nil {
		return contracts.LifecycleResult{
			Effect:     contracts.UncertainEffect,
			SessionID:  evidence.SessionID,
			PID:        evidence.PID,
			Diagnostic: validationErr.Error(), DiagnosticCode: contracts.DiagnosticInvalid,
		}
	}
	return result
}

// OpenSessionWorkspace loads the current configured command and starts it in
// the saved workspace while deletion and scope changes are excluded.
func (s *Service) OpenSessionWorkspace(
	ctx context.Context,
	request contracts.OpenSessionWorkspaceRequest,
) contracts.OpenSessionWorkspaceResult {
	err := s.openSessionWorkspace(ctx, request)
	if err != nil {
		return contracts.OpenSessionWorkspaceResult{Diagnostic: err.Error(), DiagnosticCode: diagnosticCode(err)}
	}
	return contracts.OpenSessionWorkspaceResult{Started: true}
}

func (s *Service) openSessionWorkspace(ctx context.Context, request contracts.OpenSessionWorkspaceRequest) error {
	if request.Kind != contracts.WorkspaceIDE && request.Kind != contracts.WorkspaceTerminal {
		return errors.New("unknown workspace command kind")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	settingsRepo, ok := s.repo.(interface {
		GetAppSettings() (config.AppSettings, error)
	})
	if !ok {
		return errors.New("application settings unavailable")
	}
	settings, err := settingsRepo.GetAppSettings()
	if err != nil {
		return err
	}
	command := settings.IDECommand
	if request.Kind == contracts.WorkspaceTerminal {
		command = settings.TerminalCommand
	}
	return s.sessions.OpenWorkspaceChecked(ctx, request.SessionID, command, func(sessions []session.Snapshot) error {
		inventory, inventoryErr := s.loadMembershipInventory(sessions)
		if inventoryErr != nil {
			return inventoryErr
		}
		if !inventory.allowsSession(request.DashboardID, request.SessionID) {
			return errors.New("session is outside dashboard scope")
		}
		if strings.TrimSpace(command) == "" {
			if request.Kind == contracts.WorkspaceIDE {
				return errors.New("IDE command not configured")
			}
			return fmt.Errorf("%s command not configured", "Terminal")
		}
		return nil
	})
}

func (s *Service) RenameDashboardSession(ctx context.Context, request contracts.RenameSessionRequest) contracts.RenameSessionResult {
	if err := ctx.Err(); err != nil {
		return contracts.RenameSessionResult{Diagnostic: err.Error(), DiagnosticCode: diagnosticCode(err)}
	}
	var result session.RenameResult
	var err error
	if request.ExpectedProcessID != "" {
		result, err = s.sessions.SetNameExpectedProcessChecked(
			request.SessionID,
			request.Name,
			request.ExpectedProcessID,
			s.sessionScopeCheckContext(ctx, request.DashboardID, request.SessionID),
		)
	} else {
		result, err = s.sessions.SetNameDetailedChecked(
			request.SessionID,
			request.Name,
			s.sessionScopeCheckContext(ctx, request.DashboardID, request.SessionID),
		)
	}
	output := contracts.RenameSessionResult{Applied: result.Applied, Persisted: result.Persisted}
	if err != nil {
		output.Diagnostic = err.Error()
		output.DiagnosticCode = diagnosticCode(err)
		if result.Applied {
			output.DiagnosticCode = contracts.DiagnosticPersistence
		}
	}
	return output
}

func (s *Service) StopDashboardSession(ctx context.Context, request contracts.ScopedSessionRequest) contracts.StopSessionResult {
	if err := ctx.Err(); err != nil {
		return contracts.StopSessionResult{Diagnostic: err.Error(), DiagnosticCode: diagnosticCode(err)}
	}
	result, err := s.sessions.KillDetailedChecked(
		request.SessionID,
		s.sessionScopeCheckContext(ctx, request.DashboardID, request.SessionID),
	)
	output := contracts.StopSessionResult{ProcessTerminated: result.ProcessTerminated, RecordPersisted: result.SessionRecordPersisted}
	if err != nil {
		output.Diagnostic = err.Error()
		output.DiagnosticCode = diagnosticCode(err)
		if result.ProcessTerminated {
			output.DiagnosticCode = contracts.DiagnosticPartial
		}
	}
	return output
}

func (s *Service) RemoveDashboardSession(ctx context.Context, request contracts.ScopedSessionRequest) contracts.RemoveSessionResult {
	if err := ctx.Err(); err != nil {
		return contracts.RemoveSessionResult{Diagnostic: err.Error(), DiagnosticCode: diagnosticCode(err)}
	}
	result, err := s.sessions.RemoveDetailedChecked(
		request.SessionID,
		s.sessionScopeCheckContext(ctx, request.DashboardID, request.SessionID),
	)
	output := contracts.RemoveSessionResult{
		InMemoryRemoved:  result.InMemoryRemoved,
		RecordDeleted:    result.SessionRecordDeleted,
		WorkspaceRemoved: result.TemporaryWorkspaceRemoved,
	}
	if err != nil {
		output.Diagnostic = err.Error()
		output.DiagnosticCode = diagnosticCode(err)
		if result.InMemoryRemoved || result.SessionRecordDeleted || result.TemporaryWorkspaceRemoved {
			output.DiagnosticCode = contracts.DiagnosticPartial
		}
	}
	return output
}

func diagnosticCode(err error) contracts.DiagnosticCode {
	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return contracts.DiagnosticCancelled
	case errors.Is(err, session.ErrPersistFailed):
		return contracts.DiagnosticPersistence
	default:
		return contracts.DiagnosticFailure
	}
}
