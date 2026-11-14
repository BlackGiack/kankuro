package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
	"github.com/BlackGiack/kankuro/internal/ui/terminal"
)

func (c *Coordinator) invalidateMissingAttachment(
	dashboardID string, epoch, request uint64, inventory contracts.ReadResult[contracts.DashboardInventory],
) {
	if c.attachmentOwner == nil || !inventory.Attempted || inventory.Err != nil {
		return
	}
	c.attachmentOwner.invalidateMissing(dashboardID, epoch, request, inventory.Value.Sessions)
}

func (c *Coordinator) requestAttachExpected(id, processID, kind string) tea.Cmd {
	request := c.lifecycleRequest
	if kind == messages.AttachRequestLaunch {
		request = c.launchRequest
	}
	dashboardID, generation, service := c.activeDashboard, c.viewGeneration, c.attachment
	owner := c.attachmentOwner
	ctx, sequence := owner.begin(
		attachmentTarget{
			dashboardID:      dashboardID,
			sessionID:        id,
			sessionEpoch:     c.sessionEpoch,
			inventoryRequest: c.lastAppliedInventoryRequest,
		},
	)
	return func() tea.Msg {
		handle, err := service.AttachExpectedProcessInDashboard(ctx, dashboardID, id, processID)
		if !owner.publish(sequence, handle) {
			handle = nil
		}
		return messages.ScopedAttachReadyMsg{
			Sequence:    sequence,
			DashboardID: dashboardID,
			Generation:  generation,
			SessionID:   id,
			Request:     request,
			RequestKind: kind,
			Handle:      handle,
			Err:         err,
		}
	}
}

func (c *Coordinator) requestAttach(id string) tea.Cmd {
	dashboardID, generation, service := c.activeDashboard, c.viewGeneration, c.attachment
	owner := c.attachmentOwner
	ctx, sequence := owner.begin(
		attachmentTarget{
			dashboardID:      dashboardID,
			sessionID:        id,
			sessionEpoch:     c.sessionEpoch,
			inventoryRequest: c.lastAppliedInventoryRequest,
		},
	)
	return func() tea.Msg {
		handle, err := service.AttachInDashboard(ctx, dashboardID, id)
		if !owner.publish(sequence, handle) {
			handle = nil
		}
		return messages.ScopedAttachReadyMsg{
			Sequence:    sequence,
			DashboardID: dashboardID,
			Generation:  generation,
			SessionID:   id,
			Handle:      handle,
			Err:         err,
		}
	}
}

func scopedAttachCommand(msg messages.ScopedAttachReadyMsg) tea.Cmd {
	command := terminal.New(msg.Handle)
	return tea.Exec(command, func(err error) tea.Msg {
		return messages.AttachFinishedMsg{
			Sequence:    msg.Sequence,
			SessionID:   msg.SessionID,
			DashboardID: msg.DashboardID,
			Generation:  msg.Generation,
			Err:         err,
		}
	})
}
