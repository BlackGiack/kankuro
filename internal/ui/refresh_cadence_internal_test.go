package ui

// Whitebox tests for the idle tick's config half
// (maybeRefreshConfig/Fingerprint). These live in
// package app (not app_test) because they need direct access to
// configFingerprintMsg and c.configFingerprint, which are intentionally
// unexported.

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/api"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/configstartup"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/models"
)

// countingRepo wraps a real Repository and counts calls to the List* methods
// refreshConfig/loadConfigData reads, so a test can assert a given check did
// or did not re-read the registries.
type countingRepo struct {
	config.Repository

	cardListCalls   int
	sourceListCalls int
}

func (r *countingRepo) ListIDCards() ([]config.IDCard, error) {
	r.cardListCalls++
	return r.Repository.ListIDCards()
}

func (r *countingRepo) ListSkillSources() ([]string, error) {
	r.sourceListCalls++
	return r.Repository.ListSkillSources()
}

func newCountingCoordinator(t *testing.T) (*Coordinator, *countingRepo, *config.FileRepository) {
	t.Helper()
	root := t.TempDir()
	if err := configstartup.PrepareStartup(root); err != nil {
		t.Fatal(err)
	}
	base, err := config.NewFileRepository(root)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	repo := &countingRepo{Repository: base}
	sm := newTestManager(t, base.Dir(), base)
	c := NewCoordinator(testCapabilities(api.New(repo, sm)), sm.LoadErrors())
	c.Update(messages.DashboardOpenRequestedMsg{
		DashboardID: "default", DashboardName: "Default", DashboardState: contracts.DashboardActive,
	})
	return c, repo, base
}

// TestMaybeRefreshConfigSkipsUnchangedRegistry pins step 1: a second
// maybeRefreshConfig check with nothing changed on disk must not re-read (and
// re-parse) the registries.
func TestMaybeRefreshConfigSkipsUnchangedRegistry(t *testing.T) {
	c, repo, _ := newCountingCoordinator(t)

	msg1, ok := c.maybeRefreshConfig()().(configFingerprintMsg)
	if !ok {
		t.Fatal("expected a configFingerprintMsg")
	}
	if msg1.data == nil {
		t.Fatal("expected the first check (no baseline yet) to read the registries")
	}
	c.configFingerprint = msg1.token
	callsAfterFirst := repo.cardListCalls
	if callsAfterFirst == 0 {
		t.Fatal("expected ListIDCards to have been called at least once")
	}

	msg2, ok := c.maybeRefreshConfig()().(configFingerprintMsg)
	if !ok {
		t.Fatal("expected a configFingerprintMsg")
	}
	if msg2.data != nil {
		t.Fatal("expected an unchanged registry to skip the reload")
	}
	if repo.cardListCalls != callsAfterFirst {
		t.Fatalf("expected no additional ListIDCards call, got %d calls (was %d)",
			repo.cardListCalls, callsAfterFirst)
	}
}

// TestMaybeRefreshConfigPicksUpFileChange pins step 1 move 2: an out-of-band
// edit to a registry file (different mtime+size) must cause the next check to
// re-read.
func TestMaybeRefreshConfigPicksUpFileChange(t *testing.T) {
	c, repo, base := newCountingCoordinator(t)

	msg1, ok := c.maybeRefreshConfig()().(configFingerprintMsg)
	if !ok {
		t.Fatal("expected a configFingerprintMsg")
	}
	c.configFingerprint = msg1.token
	callsAfterFirst := repo.cardListCalls

	idCardsPath := filepath.Join(base.Dir(), config.IDCardsFilename)
	data, err := os.ReadFile(idCardsPath)
	if err != nil {
		t.Fatalf("read id_cards.yaml: %v", err)
	}
	data = append(data, []byte("\n# touched by test\n")...)
	if writeErr := os.WriteFile(idCardsPath, data, 0o600); writeErr != nil {
		t.Fatalf("write id_cards.yaml: %v", writeErr)
	}
	// Force the mtime forward: same-second writes can otherwise be missed
	// by mtime-only comparison (documented residual - see Fingerprint).
	future := time.Now().Add(2 * time.Second)
	if chErr := os.Chtimes(idCardsPath, future, future); chErr != nil {
		t.Fatalf("chtimes id_cards.yaml: %v", chErr)
	}

	msg2, ok := c.maybeRefreshConfig()().(configFingerprintMsg)
	if !ok {
		t.Fatal("expected a configFingerprintMsg")
	}
	if msg2.data == nil {
		t.Fatal("expected the changed registry file to trigger a reload")
	}
	if repo.cardListCalls == callsAfterFirst {
		t.Fatal("expected an additional ListIDCards call after the on-disk edit")
	}
}

// TestMutationRefreshDoesNotWaitForTick pins step 1's other half: a
// mutation-triggered refresh (here, CardSavedMsg after a save) must reflect
// immediately, without going through maybeRefreshConfig/the tick at all.
func TestMutationRefreshDoesNotWaitForTick(t *testing.T) {
	c, _, base := newCountingCoordinator(t)

	wd := t.TempDir()
	if err := base.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd}}); err != nil {
		t.Fatalf("save workspace sources: %v", err)
	}
	workspaces, err := base.ListWorkspaces()
	if err != nil || len(workspaces) == 0 {
		t.Fatalf("list workspaces: %v (%d)", err, len(workspaces))
	}

	saved, err := base.CreateIDCard(config.IDCard{
		DashboardID: "default",
		Name:        "Mutation Card",
		Driver:      "echo",
		WorkspaceID: workspaces[0].ID,
	})
	if err != nil {
		t.Fatalf("create card: %v", err)
	}

	c.currentView = ViewCardEditor
	_, cmd := c.Update(messages.CardSavedMsg{DashboardID: "default", Generation: c.viewGeneration})
	if cmd == nil {
		t.Fatal("expected CardSavedMsg to produce a follow-up cmd")
	}
	// Drive every message the batch produces back through Update, the way
	// the real Bubble Tea runtime would - CardSavedMsg batches a refresh
	// with a navigate, and tea.Batch gives no ordering guarantee.
	for _, m := range collectMsgs(cmd) {
		c.Update(m)
	}

	dashboard, ok := c.views[ViewDashboard].(interface {
		Cards() []models.CardRow
	})
	if !ok {
		t.Fatal("expected dashboard view with a Cards() accessor")
	}
	if got := dashboard.Cards(); len(got) != 1 || got[0].ID != saved.ID {
		t.Fatalf("expected the dashboard to have the saved card immediately, got: %v", got)
	}
}

// TestIdleTickDoesNotTriggerSourceReload pins review finding 1: the idle
// tick's sessions-only refresh must never reach a config-registry view like
// SourceManager, whose SetData reloads its source files from disk regardless
// of what changed. Before the fix, refreshSessions returned a DataLoadedMsg
// broadcast to every DataReceiver, so this failed.
func TestIdleTickDoesNotTriggerSourceReload(t *testing.T) {
	c, repo, _ := newCountingCoordinator(t)

	for range 2 {
		c.Update(c.refreshSessions()())
	}

	if repo.sourceListCalls != 0 {
		t.Fatalf("expected the idle tick to never read source files, got %d ListSkillSources calls",
			repo.sourceListCalls)
	}
}

// TestFullRefreshSeedsFingerprintBaseline pins review finding 2: a
// mutation-triggered refresh (refreshDataSeeded, used by Init and every
// mutation path) must seed c.configFingerprint, so the very next idle tick
// sees the baseline as unchanged and does not reload everything a second
// time.
func TestFullRefreshSeedsFingerprintBaseline(t *testing.T) {
	c, repo, _ := newCountingCoordinator(t)

	msg, ok := c.refreshDataSeeded()().(configFingerprintMsg)
	if !ok {
		t.Fatal("expected refreshDataSeeded to return a configFingerprintMsg")
	}
	c.Update(msg)
	callsAfterMutation := repo.cardListCalls
	if callsAfterMutation == 0 {
		t.Fatal("expected the mutation refresh to have read the registries at least once")
	}

	tickMsg, ok := c.maybeRefreshConfig()().(configFingerprintMsg)
	if !ok {
		t.Fatal("expected a configFingerprintMsg")
	}
	if tickMsg.data != nil {
		t.Fatal("expected the tick right after a mutation refresh to see an unchanged baseline")
	}
	if tickMsg.token != "" {
		t.Fatal("expected an unchanged result to carry no token")
	}
	if repo.cardListCalls != callsAfterMutation {
		t.Fatalf("expected no additional ListIDCards call from the tick, got %d (was %d)",
			repo.cardListCalls, callsAfterMutation)
	}
}

// An unchanged tick result must not overwrite a newer baseline from a mutation refresh.
func TestUnchangedTickResultDoesNotClobberNewerBaseline(t *testing.T) {
	c, _, _ := newCountingCoordinator(t)

	newerBaseline := contracts.RegistryChangeToken("opaque-newer-token")
	c.Update(configFingerprintMsg{
		token:     newerBaseline,
		requestID: 1,
		data: &messages.DataLoadedMsg{
			DashboardID: "default",
			Generation:  c.viewGeneration,
		},
		dashboardID: "default",
		generation:  c.viewGeneration,
	})
	if c.configFingerprint != newerBaseline {
		t.Fatalf("expected newer baseline to be adopted, got %q", c.configFingerprint)
	}

	// A stale tick result, captured before the mutation, arrives after it.
	c.Update(configFingerprintMsg{requestID: 0})

	if c.configFingerprint != newerBaseline {
		t.Fatalf("expected stale unchanged result not to clobber the newer baseline, got %q", c.configFingerprint)
	}
}

func TestNewerUnchangedRefreshInvalidatesOlderChangedResult(t *testing.T) {
	c, _, _ := newCountingCoordinator(t)
	c.configFingerprint = "seed"
	generation := c.viewGeneration
	c.Update(configFingerprintMsg{
		requestID: 2, dashboardID: "default", generation: generation,
	})
	c.Update(configFingerprintMsg{
		token: "older-change", requestID: 1, dashboardID: "default", generation: generation,
		data: &messages.DataLoadedMsg{DashboardID: "default", Generation: generation},
	})
	if c.configFingerprint != "seed" {
		t.Fatalf("older changed refresh replaced the baseline after a newer unchanged result: %q", c.configFingerprint)
	}
}

// TestLaunchCardAbortsOnRegistryReadError pins review finding 4: a corrupt
// registry must abort launchCard (no session created, an error surfaced via
// SessionLaunchCompletedMsg) rather than launching capability-less.
func TestLaunchCardAbortsOnRegistryReadError(t *testing.T) {
	root := t.TempDir()
	if err := configstartup.PrepareStartup(root); err != nil {
		t.Fatal(err)
	}
	base, err := config.NewFileRepository(root)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := newTestManager(t, base.Dir(), base)
	c := NewCoordinator(testCapabilities(api.New(base, sm)), sm.LoadErrors())
	c.Update(messages.DashboardOpenRequestedMsg{
		DashboardID: "default", DashboardName: "Default", DashboardState: contracts.DashboardActive,
	})

	wd := t.TempDir()
	if saveErr := base.SaveWorkspaceSources([]config.WorkspaceSource{{Path: wd}}); saveErr != nil {
		t.Fatalf("save workspace sources: %v", saveErr)
	}
	workspaces, err := base.ListWorkspaces()
	if err != nil || len(workspaces) == 0 {
		t.Fatalf("list workspaces: %v (%d)", err, len(workspaces))
	}

	skill, err := base.UpsertSkill(config.Skill{Name: "a skill"})
	if err != nil {
		t.Fatalf("upsert skill: %v", err)
	}

	card, err := base.CreateIDCard(config.IDCard{
		DashboardID: "default",
		Name:        "Launch Card",
		Driver:      "echo",
		WorkspaceID: workspaces[0].ID,
		SkillIDs:    []string{skill.ID},
	})
	if err != nil {
		t.Fatalf("create card: %v", err)
	}

	skillsPath := filepath.Join(base.Dir(), config.SkillsFilename)
	if writeErr := os.WriteFile(skillsPath, []byte("not: valid: yaml: :::"), 0o600); writeErr != nil {
		t.Fatalf("corrupt skills.yaml: %v", writeErr)
	}

	loadCmd := c.dashboardCallbacks().OnLaunchCard(card.ID)
	if loadCmd == nil {
		t.Fatal("expected OnLaunchCard to return a cmd")
	}
	loaded, loadSucceeded := loadCmd().(messages.DashboardCardLoadedMsg)
	if !loadSucceeded {
		t.Fatalf("expected card-load result, got %T", loaded)
	}
	_, cmd := c.Update(loaded)

	var launchMsg *messages.SessionLaunchCompletedMsg
	for _, m := range collectMsgs(cmd) {
		if lm, ok := m.(messages.SessionLaunchCompletedMsg); ok {
			launchMsg = &lm
		}
	}
	if launchMsg == nil {
		t.Fatal("expected a SessionLaunchCompletedMsg")
	}
	if launchMsg.Result.Diagnostic == "" {
		t.Fatal("expected launchCard to surface the corrupt skills.yaml read error")
	}
	if launchMsg.Result.SessionID != "" {
		t.Fatalf("expected no session ID on an aborted launch, got %q", launchMsg.Result.SessionID)
	}
	if got := sm.List(); len(got) != 0 {
		t.Fatalf("expected no session to have been created, got %d", len(got))
	}
}

// collectMsgs executes cmd, recursing into tea.BatchMsg, and returns every
// leaf message it produces.
func collectMsgs(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if msg == nil {
		return nil
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, c := range batch {
			out = append(out, collectMsgs(c)...)
		}
		return out
	}
	return []tea.Msg{msg}
}
