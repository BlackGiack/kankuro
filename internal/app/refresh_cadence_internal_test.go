package app

// Whitebox tests for debt_impl_plan_5.md: the idle tick's config half
// (maybeRefreshConfig/Fingerprint, step 1) and the single-pass capability
// resolver (resolveCard/resolveCapabilities, steps 2-3). These live in
// package app (not app_test) because they need direct access to
// configFingerprintMsg and c.configFingerprint, which are intentionally
// unexported.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
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
	base, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	repo := &countingRepo{Repository: base}
	sm := newTestManager(t, base.Dir(), base)
	return NewCoordinator(repo, sm), repo, base
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
	c.configFingerprint = msg1.stamps
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
	c.configFingerprint = msg1.stamps
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
		Name:        "Mutation Card",
		Driver:      "echo",
		WorkspaceID: workspaces[0].ID,
	})
	if err != nil {
		t.Fatalf("create card: %v", err)
	}

	_, cmd := c.Update(messages.CardSavedMsg{Card: saved})
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
		Cards() []config.IDCard
	})
	if !ok {
		t.Fatal("expected dashboard view with a Cards() accessor")
	}
	if got := dashboard.Cards(); len(got) != 1 || got[0].ID != saved.ID {
		t.Fatalf("expected the dashboard to have the saved card immediately, got: %v", got)
	}
}

// TestResolveCapabilitiesPreservesRegistryOrder pins step 2: resolveCard's
// single-pass resolver must still return skills in registry (name-sorted)
// order, not the order they appear in card.SkillIDs - a map-based rewrite
// that iterated card.SkillIDs instead would silently change the order
// skills are written to disk / AGENTS.md.
func TestResolveCapabilitiesPreservesRegistryOrder(t *testing.T) {
	base, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := newTestManager(t, base.Dir(), base)
	c := NewCoordinator(base, sm)

	// Upsert out of name order; ListSkills always returns them sorted by
	// Name, so the registry order is: alpha, mid, zeta.
	zeta, err := base.UpsertSkill(config.Skill{Name: "zeta"})
	if err != nil {
		t.Fatalf("upsert zeta: %v", err)
	}
	alpha, err := base.UpsertSkill(config.Skill{Name: "alpha"})
	if err != nil {
		t.Fatalf("upsert alpha: %v", err)
	}
	mid, err := base.UpsertSkill(config.Skill{Name: "mid"})
	if err != nil {
		t.Fatalf("upsert mid: %v", err)
	}

	// Reference them in a different order on the card.
	card := config.IDCard{SkillIDs: []string{zeta.ID, alpha.ID, mid.ID}}

	skills, _, _, err := c.actions.resolveCapabilities(card)
	if err != nil {
		t.Fatalf("resolveCapabilities: %v", err)
	}
	if len(skills) != 3 {
		t.Fatalf("expected 3 skills, got %d: %v", len(skills), skills)
	}
	gotNames := []string{skills[0].Name, skills[1].Name, skills[2].Name}
	wantNames := []string{"alpha", "mid", "zeta"}
	for i := range wantNames {
		if gotNames[i] != wantNames[i] {
			t.Fatalf("expected registry (name-sorted) order %v, got %v", wantNames, gotNames)
		}
	}
}

// TestResolveCapabilitiesAbortsOnRegistryReadError pins step 2's other
// requirement: a failed registry read must abort (return an error) instead
// of proceeding capability-less.
func TestResolveCapabilitiesAbortsOnRegistryReadError(t *testing.T) {
	base, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := newTestManager(t, base.Dir(), base)
	c := NewCoordinator(base, sm)

	skillsPath := filepath.Join(base.Dir(), config.SkillsFilename)
	if writeErr := os.WriteFile(skillsPath, []byte("not: valid: yaml: :::"), 0o600); writeErr != nil {
		t.Fatalf("corrupt skills.yaml: %v", writeErr)
	}

	_, _, _, err = c.actions.resolveCapabilities(config.IDCard{SkillIDs: []string{"whatever"}})
	if err == nil {
		t.Fatal("expected resolveCapabilities to surface the corrupt skills.yaml read error")
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
	if tickMsg.stamps != nil {
		t.Fatal("expected an unchanged result to carry nil stamps (review_dp_5_2.md finding 1)")
	}
	if repo.cardListCalls != callsAfterMutation {
		t.Fatalf("expected no additional ListIDCards call from the tick, got %d (was %d)",
			repo.cardListCalls, callsAfterMutation)
	}
}

// TestUnchangedTickResultDoesNotClobberNewerBaseline pins
// review_dp_5_2.md finding 1: an "unchanged" configFingerprintMsg (nil
// stamps) delivered after a newer baseline was already adopted must not
// reset that baseline. This can happen when a tick's fingerprint check and a
// mutation's refreshDataSeeded are both in flight and the tick's (stale)
// result is delivered last - without the fix, it would stomp the newer
// baseline and force one extra full reload on the following tick.
func TestUnchangedTickResultDoesNotClobberNewerBaseline(t *testing.T) {
	c, _, _ := newCountingCoordinator(t)

	newerBaseline := map[string]config.FileStamp{
		config.IDCardsFilename: {Size: 123},
	}
	c.Update(configFingerprintMsg{stamps: newerBaseline})
	if len(c.configFingerprint) != 1 {
		t.Fatalf("expected the newer baseline to be adopted, got %v", c.configFingerprint)
	}

	// A stale tick result, captured before the mutation, arrives after it.
	c.Update(configFingerprintMsg{})

	if len(c.configFingerprint) != 1 || c.configFingerprint[config.IDCardsFilename].Size != 123 {
		t.Fatalf("expected the stale unchanged tick result not to clobber the newer baseline, got %v",
			c.configFingerprint)
	}
}

// TestLaunchCardAbortsOnRegistryReadError pins review finding 4: a corrupt
// registry must abort launchCard (no session created, an error surfaced via
// SessionLaunchCompletedMsg) rather than launching capability-less.
func TestLaunchCardAbortsOnRegistryReadError(t *testing.T) {
	base, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := newTestManager(t, base.Dir(), base)
	c := NewCoordinator(base, sm)

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

	cmd := c.dashboardCallbacks().OnLaunchCard(card)
	if cmd == nil {
		t.Fatal("expected OnLaunchCard to return a cmd")
	}

	var launchMsg *messages.SessionLaunchCompletedMsg
	for _, m := range collectMsgs(cmd) {
		if lm, ok := m.(messages.SessionLaunchCompletedMsg); ok {
			launchMsg = &lm
		}
	}
	if launchMsg == nil {
		t.Fatal("expected a SessionLaunchCompletedMsg")
	}
	if launchMsg.Err == nil {
		t.Fatal("expected launchCard to surface the corrupt skills.yaml read error")
	}
	if launchMsg.SessionID != "" {
		t.Fatalf("expected no session ID on an aborted launch, got %q", launchMsg.SessionID)
	}
	if got := sm.List(); len(got) != 0 {
		t.Fatalf("expected no session to have been created, got %d", len(got))
	}
}

// TestResolveCardSurfacesListCardsErrorNotNotFound pins review finding 4 /
// step 3: a failed ListIDCards must report "list cards: ...", not be
// misreported as "card not found".
func TestResolveCardSurfacesListCardsErrorNotNotFound(t *testing.T) {
	base, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := newTestManager(t, base.Dir(), base)
	c := NewCoordinator(base, sm)

	idCardsPath := filepath.Join(base.Dir(), config.IDCardsFilename)
	if writeErr := os.WriteFile(idCardsPath, []byte("not: valid: yaml: :::"), 0o600); writeErr != nil {
		t.Fatalf("corrupt id_cards.yaml: %v", writeErr)
	}

	_, err = c.actions.resolveCard("some-card-id")
	if err == nil {
		t.Fatal("expected resolveCard to surface the corrupt id_cards.yaml read error")
	}
	if !strings.Contains(err.Error(), "list cards:") {
		t.Fatalf("expected error to report the read failure (\"list cards: ...\"), got: %v", err)
	}
	if strings.Contains(err.Error(), "card not found") {
		t.Fatalf("expected the read failure not to be misreported as \"card not found\", got: %v", err)
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
