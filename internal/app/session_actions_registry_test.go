package app_test

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/app"
	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

type capabilityFailingRepository struct {
	config.Repository

	cardErr         error
	cardsRead       int
	capabilityReads int
}

func (r *capabilityFailingRepository) ListIDCards() ([]config.IDCard, error) {
	r.cardsRead++
	if r.cardErr != nil {
		return nil, r.cardErr
	}
	return r.Repository.ListIDCards()
}

func (r *capabilityFailingRepository) capabilityFailure() error {
	r.capabilityReads++
	return errors.New("capability registry unavailable")
}

func (r *capabilityFailingRepository) ListSkills() ([]config.Skill, error) {
	return nil, r.capabilityFailure()
}

func (r *capabilityFailingRepository) ListPersonas() ([]config.Persona, error) {
	return nil, r.capabilityFailure()
}

func (r *capabilityFailingRepository) ListMCPs() ([]config.MCPServer, error) {
	return nil, r.capabilityFailure()
}

func TestRelaunchLoadsCardWithoutCapabilityRegistries(t *testing.T) {
	repo, repoErr := config.NewFileRepository(t.TempDir())
	if repoErr != nil {
		t.Fatal(repoErr)
	}
	workspace := t.TempDir()
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, workspaceErr := repo.ListWorkspaces()
	if workspaceErr != nil {
		t.Fatal(workspaceErr)
	}
	card, cardErr := repo.CreateIDCard(config.IDCard{Name: "relaunch registry", Driver: "cat", WorkspaceID: workspaces[0].ID})
	if cardErr != nil {
		t.Fatal(cardErr)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	sid, launchErr := sm.Launch(card, nil, nil, nil)
	if launchErr != nil {
		t.Fatal(launchErr)
	}
	if err := sm.Kill(sid); err != nil {
		t.Fatal(err)
	}

	failingRepo := &capabilityFailingRepository{Repository: repo}
	coord := app.NewCoordinator(failingRepo, sm)
	var result tea.Msg
	for _, msg := range collectMsgs(t, coord.DashboardCallbacks().OnRelaunch(sid)) {
		result = msg
	}
	if _, ok := result.(messages.SessionAttachRequestedMsg); !ok {
		t.Fatalf("relaunch result = %T (%+v), want SessionAttachRequestedMsg", result, result)
	}
	if failingRepo.cardsRead != 1 || failingRepo.capabilityReads != 0 {
		t.Fatalf("relaunch reads: cards=%d capabilities=%d, want 1 and 0", failingRepo.cardsRead, failingRepo.capabilityReads)
	}
	if err := sm.Kill(sid); err != nil {
		t.Fatal(err)
	}
}

func TestCardAndCapabilityLookupFailuresRemainVisible(t *testing.T) {
	repo, repoErr := config.NewFileRepository(t.TempDir())
	if repoErr != nil {
		t.Fatal(repoErr)
	}
	workspace := t.TempDir()
	if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: workspace}}); err != nil {
		t.Fatal(err)
	}
	workspaces, workspaceErr := repo.ListWorkspaces()
	if workspaceErr != nil {
		t.Fatal(workspaceErr)
	}
	card, cardErr := repo.CreateIDCard(config.IDCard{Name: "lookup", Driver: "cat", WorkspaceID: workspaces[0].ID})
	if cardErr != nil {
		t.Fatal(cardErr)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	failingRepo := &capabilityFailingRepository{Repository: repo}
	coord := app.NewCoordinator(failingRepo, sm)
	sid, launchErr := sm.Launch(card, nil, nil, nil)
	if launchErr != nil {
		t.Fatal(launchErr)
	}
	if err := sm.Kill(sid); err != nil {
		t.Fatal(err)
	}

	failingRepo.cardErr = errors.New("cards unavailable")
	msg := collectMsgs(t, coord.DashboardCallbacks().OnRelaunch(sid))
	if len(msg) != 1 {
		t.Fatalf("card read failure messages = %v", msg)
	}
	errMsg, ok := msg[0].(messages.ErrorMsg)
	if !ok || !strings.Contains(errMsg.Err.Error(), "list cards: cards unavailable") {
		t.Fatalf("card read failure = %T (%+v)", msg[0], msg[0])
	}
	failingRepo.cardErr = nil
	failingRepo.Repository = repositoryWithoutCard{Repository: repo, excludedID: card.ID}
	msg = collectMsgs(t, coord.DashboardCallbacks().OnRelaunch(sid))
	errMsg, ok = msg[0].(messages.ErrorMsg)
	if !ok || !strings.Contains(errMsg.Err.Error(), "card not found: "+card.ID) {
		t.Fatalf("missing card result = %T (%+v)", msg[0], msg[0])
	}

	failingRepo.Repository = repo
	failingRepo.capabilityReads = 0
	launchMsgs := collectMsgs(t, coord.DashboardCallbacks().OnLaunchCard(card))
	if !containsCapabilityError(launchMsgs) || failingRepo.capabilityReads == 0 {
		t.Fatalf("launch did not propagate capability failure: %v", launchMsgs)
	}
	switchMsgs := collectMsgs(t, coord.AgentSwitchCallbacks().OnSwitch(sid, card.ID, ""))
	if !containsSwitchCapabilityError(switchMsgs) || failingRepo.capabilityReads == 0 {
		t.Fatalf("switch did not propagate capability failure: %v", switchMsgs)
	}
	if err := sm.Remove(sid); err != nil {
		t.Fatal(err)
	}
}

type repositoryWithoutCard struct {
	config.Repository

	excludedID string
}

func (r repositoryWithoutCard) ListIDCards() ([]config.IDCard, error) {
	cards, err := r.Repository.ListIDCards()
	if err != nil {
		return nil, err
	}
	out := cards[:0]
	for _, card := range cards {
		if card.ID != r.excludedID {
			out = append(out, card)
		}
	}
	return out, nil
}

func containsCapabilityError(msgs []tea.Msg) bool {
	for _, msg := range msgs {
		if launch, ok := msg.(messages.SessionLaunchCompletedMsg); ok &&
			launch.Err != nil && strings.Contains(launch.Err.Error(), "list skills: capability registry unavailable") {
			return true
		}
	}
	return false
}

func containsSwitchCapabilityError(msgs []tea.Msg) bool {
	for _, msg := range msgs {
		if switched, ok := msg.(messages.AgentSwitchedMsg); ok && switched.Err != nil &&
			strings.Contains(switched.Err.Error(), "list skills: capability registry unavailable") {
			return true
		}
	}
	return false
}
