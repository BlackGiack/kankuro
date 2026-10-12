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

func newBannerCoordinator(t *testing.T) *app.Coordinator {
	t.Helper()
	repo, err := config.NewFileRepository(t.TempDir())
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	sm := newTestManager(t, repo.Dir(), repo)
	coord := app.NewCoordinator(repo, sm)
	coord.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return coord
}

func assertBannerShown(t *testing.T, coord *app.Coordinator, want error, context string) {
	t.Helper()
	if coord.Err() == nil {
		t.Fatalf("%s: expected the error banner to be set, got nil", context)
	}
	if !strings.Contains(coord.View(), want.Error()) {
		t.Fatalf("%s: expected View() to contain %q", context, want.Error())
	}
}

// TestErrorSurvivesNavigationInEitherOrder pins tech_debt_act_2.md §3a:
// navigation must not clear the error banner, whichever order tea.Batch
// delivers an ErrorMsg and a NavigateMsg in.
func TestErrorSurvivesNavigationInEitherOrder(t *testing.T) {
	bannerErr := errors.New("persist failed: disk on fire")

	t.Run("error then navigate", func(t *testing.T) {
		coord := newBannerCoordinator(t)
		coord.Update(messages.ErrorMsg{Err: bannerErr})
		coord.Update(messages.NavigateMsg{View: messages.ViewSettings})
		if coord.CurrentView() != app.ViewSettings {
			t.Fatalf("expected navigation to have happened, got view %v", coord.CurrentView())
		}
		assertBannerShown(t, coord, bannerErr, "error then navigate")
	})

	t.Run("navigate then error", func(t *testing.T) {
		coord := newBannerCoordinator(t)
		coord.Update(messages.NavigateMsg{View: messages.ViewSettings})
		coord.Update(messages.ErrorMsg{Err: bannerErr})
		if coord.CurrentView() != app.ViewSettings {
			t.Fatalf("expected navigation to have happened, got view %v", coord.CurrentView())
		}
		assertBannerShown(t, coord, bannerErr, "navigate then error")
	})
}

// TestKeyPressClearsBanner: the banner goes away when the user acts.
func TestKeyPressClearsBanner(t *testing.T) {
	coord := newBannerCoordinator(t)
	coord.Update(messages.ErrorMsg{Err: errors.New("stale")})
	if coord.Err() == nil {
		t.Fatal("precondition: banner should be set")
	}

	coord.Update(tea.KeyPressMsg{Code: 0, Text: "j"})

	if coord.Err() != nil {
		t.Fatalf("expected a keypress to clear the banner, got: %v", coord.Err())
	}
}

// TestKeyPressThenErrorShowsNewError: an error that is the result of a
// keypress arrives as a later message, so clearing on the keypress must not
// hide it.
func TestKeyPressThenErrorShowsNewError(t *testing.T) {
	coord := newBannerCoordinator(t)
	coord.Update(messages.ErrorMsg{Err: errors.New("old")})

	fresh := errors.New("fresh failure")
	coord.Update(tea.KeyPressMsg{Code: 0, Text: "j"})
	coord.Update(messages.ErrorMsg{Err: fresh})

	assertBannerShown(t, coord, fresh, "keypress then error")
}
