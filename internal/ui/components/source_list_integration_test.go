package components_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/ui/components"
)

var errTest = errors.New("test")

func TestSourceListIntegration_CRUD(t *testing.T) {
	var sources []string
	refreshes, reloads, saves, deletes := 0, 0, 0, 0
	sl := components.NewSourceList(components.SourceListConfig{Title: "Sources", FormFields: 1}, components.SourceListCallbacks{
		OnSave: func(item components.SourceItem, index int) (tea.Cmd, error) {
			saves++
			if index == -1 {
				sources = append(sources, item.Display())
			} else {
				sources[index] = item.Display()
			}
			return nil, nil //nolint:nilnil // Source callbacks may succeed without a follow-up command.
		},
		OnDelete: func(index int) (tea.Cmd, error) {
			deletes++
			sources = append(sources[:index], sources[index+1:]...)
			return nil, nil //nolint:nilnil // Source callbacks may succeed without a follow-up command.
		},
		OnRefresh: func() tea.Cmd { refreshes++; return nil },
	})
	sl.Callbacks.OnReload = func() {
		reloads++
		sl.SetItems(components.ToSimpleItems(sources))
	}
	key := func(msg tea.KeyPressMsg) { t.Helper(); sl.Update(msg) }
	check := func(want []string, mutations int) {
		t.Helper()
		var got []string
		for _, item := range sl.Items {
			got = append(got, item.Display())
		}
		if !slices.Equal(got, want) || !slices.Equal(sources, want) || sl.Mode != components.SourceListModeList {
			t.Fatalf("store=%v displayed=%v mode=%v, want %v in list mode", sources, got, sl.Mode, want)
		}
		if refreshes != mutations || reloads != mutations {
			t.Fatalf("refreshes=%d reloads=%d, want %d", refreshes, reloads, mutations)
		}
		assertSourceListText(t, sl, want)
	}
	key(tea.KeyPressMsg{Text: "n"})
	if sl.Mode != components.SourceListModeForm || sl.EditingIndex != -1 {
		t.Fatal("new did not open new-item form")
	}
	key(tea.KeyPressMsg{Text: "/new/source"})
	key(tea.KeyPressMsg{Code: tea.KeyEnter})
	check([]string{"/new/source"}, 1)
	key(tea.KeyPressMsg{Code: tea.KeyEnter})
	if sl.Mode != components.SourceListModeForm || sl.EditingIndex != 0 || sl.Inputs[0].Value() != "/new/source" {
		t.Fatal("edit did not load selected source")
	}
	key(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
	key(tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl})
	key(tea.KeyPressMsg{Text: "/updated/source"})
	key(tea.KeyPressMsg{Code: tea.KeyEnter})
	check([]string{"/updated/source"}, 2)
	key(tea.KeyPressMsg{Text: "d"})
	check(nil, 3)
	if saves != 2 || deletes != 1 || sl.Cursor != 0 {
		t.Fatalf("saves=%d deletes=%d cursor=%d", saves, deletes, sl.Cursor)
	}
}

func assertSourceListText(t *testing.T, list *components.SourceList, sources []string) {
	t.Helper()
	for _, source := range sources {
		if !strings.Contains(list.View(), source) {
			t.Fatalf("view missing %q", source)
		}
	}
}
