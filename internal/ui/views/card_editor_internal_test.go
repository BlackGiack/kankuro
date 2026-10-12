package views

import (
	"testing"

	"github.com/BlackGiack/kankuro/internal/config"
)

// TestFocusCountMatchesTargets guards against the FocusCount/focusTargets
// drift that hand-maintained parallel index systems are prone to: if a new
// Focus* constant is added to the iota block without a matching entry in
// focusTargets (or vice versa), this fails instead of silently making a
// field unreachable via Tab.
func TestFocusCountMatchesTargets(t *testing.T) {
	if FocusCount != len(focusTargetTable()) {
		t.Fatalf("FocusCount = %d, len(focusTargetTable()) = %d", FocusCount, len(focusTargetTable()))
	}
	seen := make(map[int]bool, len(focusTargetTable()))
	for _, tgt := range focusTargetTable() {
		if seen[tgt.focus] {
			t.Fatalf("duplicate focus target %d", tgt.focus)
		}
		seen[tgt.focus] = true
	}
	for i := range FocusCount {
		if !seen[i] {
			t.Fatalf("focus index %d has no entry in focusTargets", i)
		}
	}
}

// TestInputIndexConstantsMatchTargets guards the other half of the drift #1
// left unclosed: nameInputIdx..faceStyleInputIdx must cover exactly the input
// slots focusTargetTable() allocates for, or a constant added without a
// matching table entry leaves a hole (and an index-out-of-range at
// construction if it is the highest).
func TestInputIndexConstantsMatchTargets(t *testing.T) {
	if got, want := inputCountFromTargets(), faceStyleInputIdx+1; got != want {
		t.Fatalf("input index constants and focusTargets disagree: %d vs %d", want, got)
	}
}

func TestVisibleFocusOrder(t *testing.T) {
	e := NewCardEditor(CardEditorCallbacks{})

	e.settings = config.AppSettings{ShowPersonasInCardEditor: true, ShowMCPsInCardEditor: true}
	full := e.visibleFocusOrder()
	if len(full) != len(focusTargetTable()) {
		t.Fatalf("with both sections enabled expected all %d targets visible, got %d", len(focusTargetTable()), len(full))
	}

	e.settings = config.AppSettings{ShowPersonasInCardEditor: false, ShowMCPsInCardEditor: false}
	reduced := e.visibleFocusOrder()
	for _, f := range reduced {
		if f == FocusPersonas || f == FocusMCPs {
			t.Fatalf("expected FocusPersonas/FocusMCPs excluded when disabled, got order %v", reduced)
		}
	}
	if len(reduced) != len(focusTargetTable())-2 {
		t.Fatalf("expected %d visible targets with both sections disabled, got %d", len(focusTargetTable())-2, len(reduced))
	}
}

// TestCycleFocusWrapsWithBothSectionsHidden is the concrete bug FocusCount
// drift would cause: tabbing forward from the last visible target must wrap
// to the first, even when the table's tail entries (Personas/MCPs) are
// hidden.
func TestCycleFocusWrapsWithBothSectionsHidden(t *testing.T) {
	e := NewCardEditor(CardEditorCallbacks{})
	e.SetSettings(config.AppSettings{ShowPersonasInCardEditor: false, ShowMCPsInCardEditor: false})

	order := e.visibleFocusOrder()
	e.setFocus(order[len(order)-1])

	e.CycleFocus(1)
	if e.focusIndex != order[0] {
		t.Fatalf("expected wrap to first visible focus %d, got %d", order[0], e.focusIndex)
	}

	e.CycleFocus(-1)
	if e.focusIndex != order[len(order)-1] {
		t.Fatalf("expected wrap to last visible focus %d, got %d", order[len(order)-1], e.focusIndex)
	}
}
