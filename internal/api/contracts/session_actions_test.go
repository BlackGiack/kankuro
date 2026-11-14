package contracts_test

import (
	"testing"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

func TestLifecycleResultValidation(t *testing.T) {
	tests := []struct {
		name  string
		value contracts.LifecycleResult
		valid bool
	}{
		{"usable", contracts.LifecycleResult{Effect: contracts.UsableActivation, SessionID: "s", ProcessActivationID: "p"}, true},
		{
			"warning",
			contracts.LifecycleResult{
				Effect:              contracts.UsableActivation,
				SessionID:           "s",
				ProcessActivationID: "p",
				Diagnostic:          "disk",
				DiagnosticCode:      contracts.DiagnosticPersistence,
			},
			true,
		},
		{
			"failure",
			contracts.LifecycleResult{Effect: contracts.NoEffect, Diagnostic: "failed", DiagnosticCode: contracts.DiagnosticFailure},
			true,
		},
		{
			"uncertain",
			contracts.LifecycleResult{
				Effect:         contracts.UncertainEffect,
				SessionID:      "s",
				PID:            12,
				Diagnostic:     "unknown",
				DiagnosticCode: contracts.DiagnosticUncertain,
			},
			true,
		},
		{"zero", contracts.LifecycleResult{}, false},
		{
			"unknown tag",
			contracts.LifecycleResult{Effect: "other", Diagnostic: "failure", DiagnosticCode: contracts.DiagnosticFailure},
			false,
		},
		{"missing process", contracts.LifecycleResult{Effect: contracts.UsableActivation, SessionID: "s"}, false},
		{
			"failure with identity",
			contracts.LifecycleResult{
				Effect:         contracts.NoEffect,
				SessionID:      "s",
				Diagnostic:     "failed",
				DiagnosticCode: contracts.DiagnosticFailure,
			},
			false,
		},
		{
			"uncertain with process",
			contracts.LifecycleResult{
				Effect:              contracts.UncertainEffect,
				ProcessActivationID: "p",
				Diagnostic:          "unknown",
				DiagnosticCode:      contracts.DiagnosticUncertain,
			},
			false,
		},
		{
			"warning without code",
			contracts.LifecycleResult{Effect: contracts.UsableActivation, SessionID: "s", ProcessActivationID: "p", Diagnostic: "warning"},
			false,
		},
		{
			"contradictory code",
			contracts.LifecycleResult{
				Effect:              contracts.UsableActivation,
				SessionID:           "s",
				ProcessActivationID: "p",
				Diagnostic:          "unknown",
				DiagnosticCode:      contracts.DiagnosticUncertain,
			},
			false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if (tc.value.Validate() == nil) != tc.valid {
				t.Fatalf("result %+v valid=%t", tc.value, tc.valid)
			}
		})
	}
}
