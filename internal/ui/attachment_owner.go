package ui

import (
	"context"
	"sync"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

// attachmentOwner retains resources while their results are queued or in use.
type attachmentOwner struct {
	mu       sync.Mutex
	sequence uint64
	cancel   context.CancelFunc
	resource contracts.Attachment
	target   attachmentTarget
	closed   bool
}

type attachmentTarget struct {
	dashboardID, sessionID         string
	sessionEpoch, inventoryRequest uint64
}

func (o *attachmentOwner) begin(target attachmentTarget) (context.Context, uint64) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.clearLocked()
	o.sequence++
	o.target = target
	ctx, cancel := context.WithCancel(context.Background())
	if o.closed {
		cancel()
	} else {
		o.cancel = cancel
	}
	return ctx, o.sequence
}

func (o *attachmentOwner) publish(sequence uint64, resource contracts.Attachment) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.closed || sequence != o.sequence || o.cancel == nil {
		if resource != nil {
			_ = resource.Close()
		}
		return false
	}
	o.resource = resource
	return true
}

func (o *attachmentOwner) current(sequence uint64, dashboardID, sessionID string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return !o.closed && sequence == o.sequence && o.cancel != nil &&
		o.target.dashboardID == dashboardID && o.target.sessionID == sessionID
}

// invalidateMissing only consumes an inventory newer than the one observed at
// request creation and from no earlier session epoch.
func (o *attachmentOwner) invalidateMissing(dashboardID string, epoch, request uint64, sessions []contracts.SessionSummary) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.cancel == nil || o.target.dashboardID != dashboardID || epoch < o.target.sessionEpoch || request <= o.target.inventoryRequest {
		return
	}
	for _, item := range sessions {
		if item.ID == o.target.sessionID {
			return
		}
	}
	o.clearLocked()
	o.sequence++
}

func (o *attachmentOwner) invalidateSession(dashboardID, sessionID string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.cancel != nil && o.target.dashboardID == dashboardID && o.target.sessionID == sessionID {
		o.clearLocked()
		o.sequence++
	}
}

func (o *attachmentOwner) clear() { o.mu.Lock(); defer o.mu.Unlock(); o.clearLocked(); o.sequence++ }
func (o *attachmentOwner) finish(sequence uint64) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if sequence == o.sequence {
		o.clearLocked()
	}
}
func (o *attachmentOwner) close() { o.mu.Lock(); defer o.mu.Unlock(); o.closed = true; o.clearLocked() }
func (o *attachmentOwner) clearLocked() {
	if o.cancel != nil {
		o.cancel()
		o.cancel = nil
	}
	if o.resource != nil {
		_ = o.resource.Close()
		o.resource = nil
	}
	o.target = attachmentTarget{}
}
