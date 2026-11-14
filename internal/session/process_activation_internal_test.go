package session

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/driver"
	"github.com/google/uuid"
)

func TestProcessActivationIdentityIsFreshAndNotPersisted(t *testing.T) {
	m := &Manager{}
	s := &Session{ID: "session"}
	h := &driver.SessionHandle{}
	m.activateSessionLocked(s, config.IDCard{}, nil, h, 0, time.Now())
	first := s.ProcessActivationID
	if _, err := uuid.Parse(first); err != nil {
		t.Fatalf("first process identity %q: %v", first, err)
	}
	m.activateSessionLocked(s, config.IDCard{}, nil, h, 0, time.Now())
	if s.ProcessActivationID == first || s.ProcessActivationID == "" {
		t.Fatalf("replacement process identity was not fresh: %q", s.ProcessActivationID)
	}
	data, err := json.Marshal(s.data())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), first) || strings.Contains(string(data), s.ProcessActivationID) {
		t.Fatal("process identity leaked into persisted session data")
	}
}

type teardownEvidenceDriver struct {
	driver.AgentDriver

	status driver.SessionStatus
	err    error
}

func (d teardownEvidenceDriver) Kill() error                     { return d.err }
func (d teardownEvidenceDriver) GetStatus() driver.SessionStatus { return d.status }

func TestUnpublishedTeardownEvidence(t *testing.T) {
	killErr := errors.New("kill warning")
	for _, tc := range []struct {
		name   string
		status driver.SessionStatus
		want   ActivationEffect
	}{{"confirmed", driver.StatusTerminated, ActivationNotPublished}, {"uncertain", driver.StatusRunning, ActivationUncertain}} {
		t.Run(tc.name, func(t *testing.T) {
			var result ActivationResult
			err := teardownActivation(
				teardownEvidenceDriver{status: tc.status, err: killErr},
				&driver.SessionHandle{},
				&result,
				"s",
				"/workspace",
				context.Canceled,
			)
			if result.Effect != tc.want || !result.Spawned || result.WorkingDir != "/workspace" || !errors.Is(err, killErr) ||
				!errors.Is(err, context.Canceled) {
				t.Fatalf("result=%+v error=%v", result, err)
			}
		})
	}
}

func TestConversationOptionIdentityIsScopedAndRevalidated(t *testing.T) {
	binding := ConversationBinding{ID: "same", Driver: "codex", Name: "old"}
	other := ConversationBinding{ID: "same", Driver: "claude"}
	s := &Session{ID: "s", Conversations: map[string][]ConversationBinding{"card": {other, binding}}}
	token := ConversationOptionIdentity(s.ID, "card", binding)
	if token == ConversationOptionIdentity(s.ID, "card", other) || token == ConversationOptionIdentity("other", "card", binding) {
		t.Fatal("identity is not scoped")
	}
	binding.Name = "renamed"
	if token != ConversationOptionIdentity(s.ID, "card", binding) {
		t.Fatal("display rename changed identity")
	}
	choice := ConversationChoice{SelectionIdentity: token}
	if id, err := validateConversationChoice(s, "card", "codex", choice); err != nil || id != "same" {
		t.Fatalf("id=%q error=%v", id, err)
	}
	if _, err := validateConversationChoice(s, "card", "claude", choice); err == nil {
		t.Fatal("accepted incompatible driver")
	}
	s.Conversations["card"] = []ConversationBinding{other}
	if _, err := validateConversationChoice(s, "card", "codex", choice); err == nil {
		t.Fatal("accepted removed binding")
	}
}
