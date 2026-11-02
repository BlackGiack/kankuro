package app_test

import (
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

// Card deletion always requires a preflight and confirmation, including when a
// session references the card. The existing session must survive deletion.
//
//nolint:gocognit,govet // Cover both deletion cases and each operation error.
func TestOnDeleteCardPreflightsAndPreservesReferencedSession(t *testing.T) {
	for _, referenced := range []bool{false, true} {
		t.Run(map[bool]string{false: "unreferenced", true: "referenced"}[referenced], func(t *testing.T) {
			repo, err := newCurrentRepository(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			sm := newTestManager(t, repo.Dir(), repo)
			if err := repo.SaveWorkspaceSources([]config.WorkspaceSource{{Path: t.TempDir()}}); err != nil {
				t.Fatal(err)
			}
			workspaces, err := repo.ListWorkspaces()
			if err != nil || len(workspaces) != 1 {
				t.Fatalf("workspaces: %v %v", workspaces, err)
			}
			card, err := createCurrentCard(repo, config.IDCard{Name: "Card", Driver: "cat", WorkspaceID: workspaces[0].ID})
			if err != nil {
				t.Fatal(err)
			}
			var sessionID string
			if referenced {
				sessionID, err = launchCurrentSession(repo, sm, card)
				if err != nil {
					t.Fatal(err)
				}
			}
			coord := newOpenedCoordinator(repo, sm)
			callbacks := coord.DashboardCallbacks()
			cmd, err := callbacks.OnDeleteCard(card.ID)
			if err != nil || cmd == nil {
				t.Fatalf("prepare deletion: %v", err)
			}
			prepared, ok := cmd().(messages.CardDeletionPreparedMsg)
			if !ok || prepared.Err != nil || prepared.Token == "" {
				t.Fatalf("preflight: %#v", prepared)
			}
			cards, err := repo.ListIDCards()
			if err != nil || len(cards) != 1 {
				t.Fatalf("preflight changed cards: %v %v", cards, err)
			}
			result, ok := callbacks.OnConfirmCardDeletion(card.ID, prepared.Token, prepared.RequiresForeignConfirmation)().(messages.ScopedActionFinishedMsg)
			if !ok || result.Err != nil {
				t.Fatalf("confirm: %#v", result)
			}
			cards, err = repo.ListIDCards()
			if err != nil || len(cards) != 0 {
				t.Fatalf("card remains: %v %v", cards, err)
			}
			if referenced {
				if _, ok := sm.GetSession(sessionID); !ok {
					t.Fatal("referenced session was removed")
				}
			}
		})
	}
}
