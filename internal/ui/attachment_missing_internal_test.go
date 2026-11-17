package ui

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

type controlledAttachmentPort struct {
	resource contracts.Attachment
	entered  chan context.Context
	release  chan struct{}
}

func (p *controlledAttachmentPort) AttachInDashboard(ctx context.Context, _, _ string) (contracts.Attachment, error) {
	if p.entered != nil {
		p.entered <- ctx
		<-p.release
	}
	return p.resource, nil
}
func (p *controlledAttachmentPort) AttachExpectedProcessInDashboard(ctx context.Context, d, s, _ string) (contracts.Attachment, error) {
	return p.AttachInDashboard(ctx, d, s)
}

func attachmentInventory(c *Coordinator, request, epoch uint64, ids ...string) messages.SessionsRefreshedMsg {
	sessions := make([]contracts.SessionSummary, 0, len(ids))
	for _, id := range ids {
		sessions = append(sessions, contracts.SessionSummary{ID: id, DashboardID: c.activeDashboard})
	}
	return messages.SessionsRefreshedMsg{
		DashboardID:        c.activeDashboard,
		Generation:         c.viewGeneration,
		SessionEpoch:       epoch,
		InventoryRequestID: request,
		Inventory: contracts.ReadResult[contracts.DashboardInventory]{
			Attempted: true,
			Value:     contracts.DashboardInventory{Sessions: sessions},
		},
	}
}

func TestMissingSessionClosesQueuedAttachment(t *testing.T) {
	for _, full := range []bool{false, true} {
		t.Run(map[bool]string{false: "sessions", true: "full"}[full], func(t *testing.T) {
			c, _ := newLaunchFocusCoordinator(t)
			closed := false
			c.attachment = &controlledAttachmentPort{resource: &fakeAttachment{closed: &closed}}
			ready := c.requestAttach("gone")().(messages.ScopedAttachReadyMsg)
			request := c.lastAppliedInventoryRequest + 1
			if full {
				c.Update(
					messages.DataLoadedMsg{
						DashboardID:        c.activeDashboard,
						Generation:         c.viewGeneration,
						SessionEpoch:       c.sessionEpoch,
						InventoryRequestID: request,
						Inventory:          contracts.ReadResult[contracts.DashboardInventory]{Attempted: true},
					},
				)
			} else {
				c.Update(attachmentInventory(c, request, c.sessionEpoch))
			}
			if !closed {
				t.Fatal("queued attachment was not closed on confirmed absence")
			}
			_, cmd := c.Update(ready)
			if cmd != nil {
				t.Fatal("removed session started terminal")
			}
		})
	}
}

func TestMissingSessionCancelsAcquisitionAndClosesLateResult(t *testing.T) {
	for _, trigger := range []string{"inventory", "removal"} {
		t.Run(trigger, func(t *testing.T) {
			c, _ := newLaunchFocusCoordinator(t)
			closed := false
			port := &controlledAttachmentPort{
				resource: &fakeAttachment{closed: &closed},
				entered:  make(chan context.Context, 1),
				release:  make(chan struct{}),
			}
			c.attachment = port
			cmd := c.requestAttach("gone")
			done := make(chan messages.ScopedAttachReadyMsg, 1)
			go func() { done <- cmd().(messages.ScopedAttachReadyMsg) }()
			ctx := <-port.entered
			if trigger == "inventory" {
				c.Update(attachmentInventory(c, c.lastAppliedInventoryRequest+1, c.sessionEpoch))
			} else {
				c.Update(messages.SessionRemovedMsg{DashboardID: c.activeDashboard, Generation: c.viewGeneration,
					SessionID: "gone", Result: contracts.RemoveSessionResult{InMemoryRemoved: true}})
			}
			select {
			case <-ctx.Done():
			case <-time.After(time.Second):
				t.Fatal("acquisition not cancelled")
			}
			close(port.release)
			ready := <-done
			if !closed || ready.Handle != nil {
				t.Fatalf("late resource: closed=%t handle=%v", closed, ready.Handle)
			}
			if _, next := c.Update(ready); next != nil {
				t.Fatal("late result started terminal")
			}
		})
	}
}

func TestAttachmentSurvivesNonAuthoritativeInventory(t *testing.T) {
	for _, kind := range []string{"failed", "unattempted", "stale_epoch", "stale_request", "other_dashboard", "old_generation", "present"} {
		t.Run(kind, func(t *testing.T) {
			c, _ := newLaunchFocusCoordinator(t)
			c.sessionEpoch = 3
			c.lastAppliedInventoryRequest = 10
			closed := false
			c.attachment = &controlledAttachmentPort{resource: &fakeAttachment{closed: &closed}}
			ready := c.requestAttachExpected("new", "process", messages.AttachRequestLaunch)().(messages.ScopedAttachReadyMsg)
			snapshot := attachmentInventory(c, 11, 3)
			switch kind {
			case "failed":
				snapshot.Inventory.Err = errors.New("read failed")
				snapshot.Err = snapshot.Inventory.Err
			case "unattempted":
				snapshot.Inventory.Attempted = false
			case "stale_epoch":
				snapshot.SessionEpoch = 2
			case "stale_request":
				snapshot.InventoryRequestID = 10
			case "other_dashboard":
				snapshot.DashboardID = "other"
			case "old_generation":
				snapshot.Generation--
			case "present":
				snapshot = attachmentInventory(c, 11, 3, "new")
			}
			c.Update(snapshot)
			if closed || !c.attachmentOwner.current(ready.Sequence, c.activeDashboard, "new") {
				t.Fatal("non-authoritative absence invalidated attachment")
			}
			c.attachmentOwner.close()
		})
	}
}

func TestRemovalResultCancelsMatchingAttachment(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	closed := false
	c.attachment = &controlledAttachmentPort{resource: &fakeAttachment{closed: &closed}}
	ready := c.requestAttach("gone")().(messages.ScopedAttachReadyMsg)
	c.Update(
		messages.SessionRemovedMsg{
			DashboardID: c.activeDashboard,
			Generation:  c.viewGeneration,
			SessionID:   "other",
			Result:      contracts.RemoveSessionResult{InMemoryRemoved: true},
		},
	)
	if closed {
		t.Fatal("unrelated removal closed attachment")
	}
	c.Update(
		messages.SessionRemovedMsg{
			DashboardID: c.activeDashboard,
			Generation:  c.viewGeneration,
			SessionID:   "gone",
			Result:      contracts.RemoveSessionResult{InMemoryRemoved: true},
		},
	)
	if !closed {
		t.Fatal("removed target attachment remained open")
	}
	if _, cmd := c.Update(ready); cmd != nil {
		t.Fatal("removed target started terminal")
	}
}

func TestNewLaunchAttachmentIgnoresPrePublicationInventory(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	oldEpoch := c.sessionEpoch
	oldSnapshot := attachmentInventory(c, c.lastAppliedInventoryRequest+1, oldEpoch)
	c.Update(launchCompleted(c, "new", nil))
	if c.sessionEpoch <= oldEpoch {
		t.Fatal("launch did not advance session epoch")
	}
	closed := false
	c.attachment = &controlledAttachmentPort{resource: &fakeAttachment{closed: &closed}}
	ready := c.requestAttachExpected("new", "process", messages.AttachRequestLaunch)().(messages.ScopedAttachReadyMsg)
	c.Update(oldSnapshot)
	if closed {
		t.Fatal("pre-publication inventory closed new launch")
	}
	if _, cmd := c.Update(ready); cmd == nil {
		t.Fatal("new launch attachment was rejected")
	}
	c.attachmentOwner.close()
}

func TestSuccessfulInventoryWithOtherReadErrorStillClosesMissingAttachment(t *testing.T) {
	c, _ := newLaunchFocusCoordinator(t)
	closed := false
	c.attachment = &controlledAttachmentPort{resource: &fakeAttachment{closed: &closed}}
	ready := c.requestAttach("gone")().(messages.ScopedAttachReadyMsg)
	failure := errors.New("registry read failed")
	c.Update(messages.DataLoadedMsg{DashboardID: c.activeDashboard, Generation: c.viewGeneration, SessionEpoch: c.sessionEpoch,
		InventoryRequestID: c.lastAppliedInventoryRequest + 1, Err: failure, RegistryErr: failure,
		Inventory: contracts.ReadResult[contracts.DashboardInventory]{Attempted: true}})
	if !closed {
		t.Fatal("successful empty inventory was ignored because another read failed")
	}
	if _, cmd := c.Update(ready); cmd != nil {
		t.Fatal("removed session started terminal")
	}
}
