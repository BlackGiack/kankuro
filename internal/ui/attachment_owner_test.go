//nolint:testpackage // Exercises package-private resource ownership.
package ui

import (
	"errors"
	"testing"

	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

func TestAttachmentOwnerDisposesLateAndSupersededResults(t *testing.T) {
	o := &attachmentOwner{}
	ctx, old := o.begin(attachmentTarget{})
	_, current := o.begin(attachmentTarget{})
	select {
	case <-ctx.Done():
	default:
		t.Fatal("superseded request was not canceled")
	}
	closed := false
	if o.publish(old, &fakeAttachment{closed: &closed}) || !closed {
		t.Fatal("late resource was not closed")
	}
	closed = false
	if !o.publish(current, &fakeAttachment{closed: &closed}) {
		t.Fatal("current resource rejected")
	}
	o.close()
	if !closed {
		t.Fatal("shutdown did not close queued resource")
	}
}

func TestAttachmentResultIsDiscardedAfterLeavingAndReturning(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	_, request := c.attachmentOwner.begin(attachmentTarget{})
	c.navigateTo(ViewSettings)
	c.navigateTo(ViewDashboard)
	closed := false
	if c.attachmentOwner.publish(request, &fakeAttachment{closed: &closed}) || !closed {
		t.Fatal("attachment from previous visit was retained")
	}
}

func TestAttachmentResultWithErrorClosesResource(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	closed := false
	_, cmd := c.Update(
		messages.ScopedAttachReadyMsg{
			DashboardID: c.activeDashboard,
			Generation:  c.viewGeneration,
			SessionID:   "s",
			Handle:      &fakeAttachment{closed: &closed},
			Err:         errors.New("failed"),
		},
	)
	if !closed || cmd != nil {
		t.Fatalf("closed=%t command=%v", closed, cmd != nil)
	}
}
