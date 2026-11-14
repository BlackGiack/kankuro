package session

import (
	"testing"
	"time"

	"github.com/creack/pty"
)

func TestOldAttachmentDrainsAfterReplacement(t *testing.T) {
	m, _, _, id, _ := newReviewMessagingManager(t)
	oldOutput := make(chan []byte, 1)
	oldSubscriber := make(chan []byte, 1)
	newSubscriber := make(chan []byte, 1)
	m.mu.Lock()
	s := m.sessions[id]
	generation := s.generation
	oldSubscribers := s.Subscribers
	oldSubscribers[oldSubscriber] = struct{}{}
	s.generation++
	s.Subscribers = map[chan []byte]struct{}{newSubscriber: {}}
	m.mu.Unlock()
	done := make(chan struct{})
	go func() { m.trackOutput(id, generation, oldOutput, oldSubscribers); close(done) }()
	oldOutput <- []byte("tail")
	close(oldOutput)
	select {
	case got := <-oldSubscriber:
		if string(got) != "tail" {
			t.Fatalf("old output = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("old attachment did not receive trailing output")
	}
	<-done
	m.mu.RLock()
	replacementComplete := s.outputComplete
	m.mu.RUnlock()
	if replacementComplete {
		t.Fatal("old stream completion marked the replacement complete")
	}
	if _, ok := <-oldSubscriber; ok {
		t.Fatal("old subscription remained open")
	}
	select {
	case got := <-newSubscriber:
		t.Fatalf("new attachment received old output: %q", got)
	default:
	}
}

func TestAttachmentDrainsAfterSessionRemoved(t *testing.T) {
	m, _, _, id, _ := newReviewMessagingManager(t)
	output := make(chan []byte, 1)
	subscriber := make(chan []byte, 1)
	m.mu.Lock()
	s := m.sessions[id]
	generation, subscribers := s.generation, s.Subscribers
	subscribers[subscriber] = struct{}{}
	delete(m.sessions, id)
	m.mu.Unlock()
	done := make(chan struct{})
	go func() { m.trackOutput(id, generation, output, subscribers); close(done) }()
	output <- []byte("last")
	close(output)
	if got := <-subscriber; string(got) != "last" {
		t.Fatalf("output = %q", got)
	}
	<-done
	if _, ok := <-subscriber; ok {
		t.Fatal("subscription remained open")
	}
}

func TestAttachmentResizeUpdatesPTYAndRejectsStoppedProcess(t *testing.T) {
	m, _, _, id, _ := newReviewMessagingManager(t)
	handle, err := m.AttachChecked(id, func([]Snapshot) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Close()
	if err = handle.ResizeChecked(40, 120); err != nil {
		t.Fatal(err)
	}
	rows, columns, err := pty.Getsize(handle.PTY)
	if err != nil || rows != 40 || columns != 120 {
		t.Fatalf("size %d x %d: %v", rows, columns, err)
	}
	if err = handle.ResizeChecked(0, 120); err == nil {
		t.Fatal("invalid size accepted")
	}
	if err = m.Kill(id); err != nil {
		t.Fatal(err)
	}
	if err = handle.ResizeChecked(24, 80); err == nil {
		t.Fatal("stopped attachment accepted resize")
	}
	if _, err = handle.WriteChecked([]byte("input")); err == nil {
		t.Fatal("stopped attachment accepted input")
	}
}

func TestAttachmentStreamCompletionAdmission(t *testing.T) {
	output := make(chan []byte, 2)
	done := make(chan struct{})
	s := &Session{
		ID:          "stream",
		Status:      StatusRunning,
		generation:  1,
		Output:      output,
		Done:        make(chan struct{}),
		Subscribers: make(map[chan []byte]struct{}),
	}
	m := &Manager{sessions: map[string]*Session{"stream": s}}
	attach := func() *AttachHandle {
		t.Helper()
		h, err := m.AttachChecked("stream", func([]Snapshot) error { return nil })
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	before := attach()
	defer before.Close()
	trackerDone := make(chan struct{})
	go func() { m.trackOutput("stream", 1, output, s.Subscribers); close(trackerDone) }()
	output <- []byte("first")
	if got := <-before.Output; string(got) != "first" {
		t.Fatalf("first output %q", got)
	}
	during := attach()
	defer during.Close()
	cancelled := attach()
	cancelled.Close()
	output <- []byte("tail")
	close(output)
	<-trackerDone
	for _, h := range []*AttachHandle{before, during} {
		if got := <-h.Output; string(got) != "tail" {
			t.Fatalf("trailing output %q", got)
		}
		if _, ok := <-h.Output; ok {
			t.Fatal("stream did not end")
		}
	}
	select {
	case <-s.Done:
		t.Fatal("process should still be running")
	default:
	}
	after := attach()
	defer after.Close()
	select {
	case _, ok := <-after.Output:
		if ok {
			t.Fatal("post-EOF subscription produced data")
		}
	case <-time.After(time.Second):
		t.Fatal("post-EOF subscription was orphaned")
	}
	select {
	case _, ok := <-cancelled.Output:
		if ok {
			t.Fatal("closed subscription produced data")
		}
	default:
		t.Fatal("closed subscription remained open")
	}
	// A replacement owns a fresh completion state and its own subscription map.
	m.mu.Lock()
	s.generation++
	s.outputComplete = false
	s.Subscribers = make(map[chan []byte]struct{})
	newSubscribers := s.Subscribers
	m.mu.Unlock()
	replacement := attach()
	defer replacement.Close()
	newOutput := make(chan []byte, 1)
	m.mu.Lock()
	s.Output = newOutput
	m.mu.Unlock()
	newOutput <- []byte("new")
	close(newOutput)
	go func() { m.trackOutput("stream", 2, newOutput, newSubscribers); close(done) }()
	if got := <-replacement.Output; string(got) != "new" {
		t.Fatalf("replacement output %q", got)
	}
	<-done
	if _, ok := <-replacement.Output; ok {
		t.Fatal("replacement stream did not end")
	}
	m.mu.Lock()
	delete(m.sessions, "stream")
	m.mu.Unlock()
	if _, err := m.AttachChecked("stream", func([]Snapshot) error { return nil }); err == nil {
		t.Fatal("removed session accepted attachment")
	}
}

func TestAttachmentAdmissionRacesStreamCompletion(t *testing.T) {
	for range 100 {
		output := make(chan []byte)
		s := &Session{ID: "race", Status: StatusRunning, generation: 1, Output: output, Subscribers: make(map[chan []byte]struct{})}
		m := &Manager{sessions: map[string]*Session{"race": s}}
		trackerDone := make(chan struct{})
		go func() { m.trackOutput("race", 1, output, s.Subscribers); close(trackerDone) }()
		type result struct {
			handle *AttachHandle
			err    error
		}
		attached := make(chan result, 1)
		go func() {
			handle, err := m.AttachChecked("race", func([]Snapshot) error { return nil })
			attached <- result{handle, err}
		}()
		close(output)
		outcome := <-attached
		if outcome.err != nil {
			t.Fatal(outcome.err)
		}
		<-trackerDone
		select {
		case _, ok := <-outcome.handle.Output:
			if ok {
				t.Fatal("unexpected output after EOF")
			}
		case <-time.After(time.Second):
			t.Fatal("attachment was orphaned during EOF")
		}
		outcome.handle.Close()
	}
}

func TestOldStreamCompletionDoesNotCloseReusedSessionID(t *testing.T) {
	oldOutput, newOutput := make(chan []byte), make(chan []byte)
	old := &Session{ID: "reused", Status: StatusRunning, generation: 1, Output: oldOutput, Subscribers: make(map[chan []byte]struct{})}
	current := &Session{ID: "reused", Status: StatusRunning, generation: 1, Output: newOutput, Subscribers: make(map[chan []byte]struct{})}
	m := &Manager{sessions: map[string]*Session{"reused": old}}
	oldDone := make(chan struct{})
	go func() { m.trackOutput("reused", 1, oldOutput, old.Subscribers); close(oldDone) }()
	m.mu.Lock()
	m.sessions["reused"] = current
	m.mu.Unlock()
	close(oldOutput)
	<-oldDone
	h, err := m.AttachChecked("reused", func([]Snapshot) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	select {
	case <-h.Output:
		t.Fatal("old EOF closed replacement stream")
	default:
	}
	newDone := make(chan struct{})
	go func() { m.trackOutput("reused", 1, newOutput, current.Subscribers); close(newDone) }()
	close(newOutput)
	<-newDone
	if _, ok := <-h.Output; ok {
		t.Fatal("replacement output remained open")
	}
}
