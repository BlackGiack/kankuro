package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/BlackGiack/kankuro/internal/session"
	"github.com/BlackGiack/kankuro/internal/ui/messages"
)

// previewSub owns the dashboard preview's live-output subscription. It is
// coordinator-goroutine state: only call its methods from Update/handlers,
// never from inside a tea.Cmd closure. next is the one method that returns a
// closure, and that closure captures only the channel and the session ID,
// never the previewSub itself.
//
// It returns data and never broadcasts: the coordinator passes what
// subscribe returns to broadcastPreview, so previewSub needs no access to
// views.
type previewSub struct {
	sessions  *session.Manager
	sessionID string
	ch        chan []byte
}

// subscribe switches the subscription to session id: it drops the previous
// subscription, takes a scrollback snapshot, then subscribes to live output.
// The returned scrollback is what the caller should broadcast immediately;
// the Cmd, when non-nil, waits for the first live chunk. A failed subscribe still records
// id (so the dashboard's auto-selection is not re-requested every tick) but
// leaves no channel, so accepts rejects any output for it.
func (p *previewSub) subscribe(id string) ([]byte, tea.Cmd) {
	p.release()
	p.sessionID = id

	scrollback := p.sessions.GetScrollback(id)

	ch, err := p.sessions.SubscribeOutput(id)
	if err != nil {
		return scrollback, nil
	}
	p.ch = ch
	return scrollback, p.next()
}

// unsubscribe clears the session ID and subscription together so queued output
// from the previous selection cannot rearm a stale subscription.
func (p *previewSub) unsubscribe() {
	p.release()
	p.sessionID = ""
}

// release drops the live-output channel, leaving sessionID alone.
func (p *previewSub) release() {
	if p.ch != nil {
		p.sessions.UnsubscribeOutput(p.sessionID, p.ch)
		p.ch = nil
	}
}

// current is the session the subscription is for, or "" when there is none.
func (p *previewSub) current() string {
	return p.sessionID
}

// accepts reports whether m is output for the session we are still
// subscribed to.
func (p *previewSub) accepts(m messages.SessionOutputMsg) bool {
	return m.SessionID == p.sessionID && p.ch != nil
}

// next waits for the next chunk of live output on the current channel. It
// returns nil when there is no subscription. The closure captures the
// channel and ID by value, so it never reads the previewSub from its own
// goroutine.
func (p *previewSub) next() tea.Cmd {
	ch, id := p.ch, p.sessionID
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		output, ok := <-ch
		if !ok {
			return nil
		}
		return messages.SessionOutputMsg{
			SessionID: id,
			Output:    output,
		}
	}
}
