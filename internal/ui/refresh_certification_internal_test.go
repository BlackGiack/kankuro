package ui

import (
	"context"
	"errors"
	"testing"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/views"
)

type startupConfigurationSpy struct {
	configurationPort
	puppetSyncs int
}

func (s *startupConfigurationSpy) SyncSources(ctx context.Context, kind contracts.SourceKind) contracts.SourceSyncResult {
	if kind == contracts.SourcePuppets {
		s.puppetSyncs++
	}
	return s.configurationPort.SyncSources(ctx, kind)
}

func TestCoordinatorStartupReadsRegistriesWithoutResyncingPuppets(t *testing.T) {
	_, _, repo := newCountingCoordinator(t)
	service := api.New(repo, nil)
	ports := testCapabilities(service)
	spy := &startupConfigurationSpy{configurationPort: service}
	ports.Configuration = spy
	c := NewCoordinator(ports, nil)

	var refreshed bool
	for _, msg := range collectMsgs(c.Init()) {
		if refresh, ok := msg.(configFingerprintMsg); ok {
			if refresh.data == nil || refresh.data.Err != nil || refresh.token == "" {
				t.Fatalf("initial registry read failed to establish a baseline: %+v", refresh)
			}
			refreshed = true
		}
	}
	if !refreshed || spy.puppetSyncs != 0 {
		t.Fatalf("initial refresh completed = %t, UI puppet syncs = %d", refreshed, spy.puppetSyncs)
	}
}

type certificationReader struct {
	tokens       []tokenResult
	tokenCalls   int
	choices      contracts.DisplayChoicesResult
	inventory    contracts.DashboardInventory
	inventoryErr error
}

type tokenResult struct {
	token contracts.RegistryChangeToken
	err   error
}

func (r *certificationReader) ReadDashboardInventory() ([]contracts.DashboardSummary, error) {
	return nil, nil
}
func (r *certificationReader) ReadDashboardOverview(string) (contracts.DashboardInventory, error) {
	return r.inventory, r.inventoryErr
}
func (r *certificationReader) ReadDisplayChoices() contracts.DisplayChoicesResult { return r.choices }
func (r *certificationReader) ReadRegistryChangeToken() (contracts.RegistryChangeToken, error) {
	index := r.tokenCalls
	r.tokenCalls++
	if index >= len(r.tokens) {
		return "stable", nil
	}
	return r.tokens[index].token, r.tokens[index].err
}

func TestRefreshTokenCertificationFailuresRetryAndKeepInventory(t *testing.T) {
	tests := []struct {
		name   string
		tokens []tokenResult
	}{
		{"token read before load fails", []tokenResult{{err: errors.New("before failed")}, {token: "stable"}}},
		{"token read after load fails", []tokenResult{{token: "stable"}, {err: errors.New("after failed")}}},
		{"token changes during load", []tokenResult{{token: "before"}, {token: "after"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _, _ := newCountingCoordinator(t)
			reader := successfulCertificationReader()
			reader.choices = emptySuccessfulChoices()
			reader.tokens = tt.tokens
			c.reads = reader
			msg := c.refreshDataSeeded()()
			c.Update(msg)
			if c.configFingerprint != "" || c.err == nil {
				t.Fatalf("uncertified registry baseline accepted: token=%q err=%v", c.configFingerprint, c.err)
			}
			if rows := c.views[ViewDashboard].(*views.Dashboard).Sessions(); len(rows) != 1 || rows[0].Name != "usable" {
				t.Fatalf("useful inventory was lost: %+v", rows)
			}
			// With no valid baseline, an unchanged token still triggers a new read.
			retry := c.maybeRefreshConfig()().(configFingerprintMsg)
			if retry.data == nil {
				t.Fatal("retry skipped despite invalid baseline")
			}
			c.Update(retry)
			if c.configFingerprint != "stable" {
				t.Fatalf("retry did not certify unchanged metadata: %q", c.configFingerprint)
			}
		})
	}
}

func TestFirstInventorySurvivesFailedDisplayChoices(t *testing.T) {
	c, _, _ := newCountingCoordinator(t)
	reader := successfulCertificationReader()
	choiceErr := errors.New("choice registry unavailable")
	reader.choices = contracts.DisplayChoicesResult{
		Skills:   contracts.ReadResult[[]contracts.NamedChoice]{Attempted: true, Err: choiceErr},
		Personas: contracts.ReadResult[[]contracts.NamedChoice]{Attempted: true, Err: choiceErr},
		Puppets:  contracts.ReadResult[[]contracts.PuppetChoice]{Attempted: true, Err: choiceErr},
	}
	reader.inventory.Cards = []contracts.CardSummary{{ID: "c", OwnerDashboardID: "default", Name: "Card", SkillIDs: []string{"missing"}}}
	c.reads = reader
	c.Update(c.refreshDataSeeded()())
	dashboard := c.views[ViewDashboard].(*views.Dashboard)
	if len(dashboard.Cards()) != 1 || dashboard.Cards()[0].ID != "c" || len(dashboard.Sessions()) != 1 {
		t.Fatalf("first successful inventory was withheld on choice failure: cards=%+v sessions=%+v",
			dashboard.Cards(), dashboard.Sessions())
	}
	if c.err == nil || c.configFingerprint != "" {
		t.Fatalf("choice failure did not surface or invalidate baseline: err=%v token=%q", c.err, c.configFingerprint)
	}
}

func TestRefreshChoiceFailuresAreIndependentAndRecover(t *testing.T) {
	for _, failed := range []string{"skills", "personas", "puppets"} {
		t.Run(failed, func(t *testing.T) { assertChoiceFailureAndRecovery(t, failed) })
	}
}

func assertChoiceFailureAndRecovery(t *testing.T, failed string) {
	t.Helper()
	c, _, _ := newCountingCoordinator(t)
	c.displayChoices = contracts.DisplayChoices{
		Skills:   []contracts.NamedChoice{{ID: "skill", Name: "prior skill"}},
		Personas: []contracts.NamedChoice{{ID: "persona", Name: "prior persona"}},
		Puppets:  []contracts.PuppetChoice{{ID: "puppet", Name: "Prior Puppet", Lines: []string{"prior art"}}},
	}
	reader := successfulCertificationReader()
	reader.choices = emptySuccessfulChoices()
	setChoiceFailure(reader, failed, errors.New("choice read failed"))
	c.reads = reader
	c.Update(c.refreshDataSeeded()())
	if c.err == nil || c.configFingerprint != "" {
		t.Fatalf("choice failure did not surface/invalidate token: err=%v token=%q", c.err, c.configFingerprint)
	}
	if rows := c.views[ViewDashboard].(*views.Dashboard).Sessions(); len(rows) != 1 || rows[0].Name != "usable" {
		t.Fatalf("inventory was discarded with independent choice failure: %+v", rows)
	}
	assertFailedChoiceRetained(t, c.displayChoices, failed)

	reader.choices = emptySuccessfulChoices()
	reader.tokens = []tokenResult{{token: "stable"}, {token: "stable"}}
	c.Update(c.refreshDataSeeded()())
	assertSuccessfulEmptyClears(t, c.displayChoices)
	if c.configFingerprint != "stable" {
		t.Fatalf("recovered reads did not certify baseline: err=%v token=%q", c.err, c.configFingerprint)
	}
}

func setChoiceFailure(reader *certificationReader, failed string, err error) {
	switch failed {
	case "skills":
		reader.choices.Skills = contracts.ReadResult[[]contracts.NamedChoice]{Attempted: true, Err: err}
	case "personas":
		reader.choices.Personas = contracts.ReadResult[[]contracts.NamedChoice]{Attempted: true, Err: err}
	case "puppets":
		reader.choices.Puppets = contracts.ReadResult[[]contracts.PuppetChoice]{Attempted: true, Err: err}
	}
}

func assertFailedChoiceRetained(t *testing.T, choices contracts.DisplayChoices, failed string) {
	t.Helper()
	if failed == "skills" && (len(choices.Skills) != 1 || choices.Skills[0].Name != "prior skill") {
		t.Fatalf("failed skill read discarded cache: %+v", choices.Skills)
	}
	if failed == "personas" && (len(choices.Personas) != 1 || choices.Personas[0].Name != "prior persona") {
		t.Fatalf("failed persona read discarded cache: %+v", choices.Personas)
	}
	if failed == "puppets" && (len(choices.Puppets) != 1 || choices.Puppets[0].Lines[0] != "prior art") {
		t.Fatalf("failed puppet read discarded artwork cache: %+v", choices.Puppets)
	}
}

func assertSuccessfulEmptyClears(t *testing.T, choices contracts.DisplayChoices) {
	t.Helper()
	if len(choices.Skills) != 0 {
		t.Fatalf("successful empty skills did not clear: %+v", choices.Skills)
	}
	if len(choices.Personas) != 0 {
		t.Fatalf("successful empty personas did not clear: %+v", choices.Personas)
	}
	if len(choices.Puppets) != 0 {
		t.Fatalf("successful empty puppets did not clear: %+v", choices.Puppets)
	}
}

func successfulCertificationReader() *certificationReader {
	return &certificationReader{
		tokens: []tokenResult{{token: "stable"}, {token: "stable"}},
		inventory: contracts.DashboardInventory{Sessions: []contracts.SessionSummary{{
			ID: "s", DashboardID: "default", CardID: "c", Name: "usable",
		}}},
		choices: emptySuccessfulChoices(),
	}
}

func emptySuccessfulChoices() contracts.DisplayChoicesResult {
	return contracts.DisplayChoicesResult{
		Skills:   contracts.ReadResult[[]contracts.NamedChoice]{Attempted: true, Value: []contracts.NamedChoice{}},
		Personas: contracts.ReadResult[[]contracts.NamedChoice]{Attempted: true, Value: []contracts.NamedChoice{}},
		Puppets:  contracts.ReadResult[[]contracts.PuppetChoice]{Attempted: true, Value: []contracts.PuppetChoice{}},
	}
}
