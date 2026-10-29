//nolint:testpackage // Exercises locked activation transitions and stale callbacks.
package session

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/driver"
)

type controlledStartupTimer struct {
	ready   chan time.Time
	stopped chan struct{}
	once    sync.Once
}

func (t *controlledStartupTimer) C() <-chan time.Time { return t.ready }
func (t *controlledStartupTimer) Stop() bool          { t.once.Do(func() { close(t.stopped) }); return true }

type controlledStartupClock struct {
	mu        sync.Mutex
	now       time.Time
	created   chan *controlledStartupTimer
	durations chan time.Duration
}

func (c *controlledStartupClock) Now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.now }
func (c *controlledStartupClock) NewTimer(d time.Duration) StartupTimer {
	timer := &controlledStartupTimer{ready: make(chan time.Time, 1), stopped: make(chan struct{})}
	c.durations <- d
	c.created <- timer
	return timer
}
func (c *controlledStartupClock) advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	c.mu.Unlock()
}
func waitStartupSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("timer did not stop")
	}
}
func startControlledTimer(t *testing.T, m *Manager, id string, clock *controlledStartupClock, deadline time.Time) *controlledStartupTimer {
	t.Helper()
	m.mu.Lock()
	m.startupClock = clock
	m.startupDelay = 10 * time.Second
	m.startStartupLocked(m.sessions[id], deadline)
	m.mu.Unlock()
	select {
	case timer := <-clock.created:
		return timer
	case <-time.After(3 * time.Second):
		t.Fatal("timer not created")
		return nil
	}
}
func newControlledStartupClock() *controlledStartupClock {
	return &controlledStartupClock{
		now: time.Unix(100, 0), created: make(chan *controlledStartupTimer, 4),
		durations: make(chan time.Duration, 4),
	}
}

func TestStartupAdmissionAndRegistrationAreIndependent(t *testing.T) {
	m, _, senderToken, recipientID, activation := newReviewMessagingManager(t)
	clock := newControlledStartupClock()
	deadline := clock.Now().Add(10 * time.Second)
	clock.advance(4 * time.Second)
	timer := startControlledTimer(t, m, recipientID, clock, deadline)
	if duration := <-clock.durations; duration != 6*time.Second {
		t.Fatalf("timer duration = %v, want remaining six seconds", duration)
	}
	m.mu.RLock()
	recipientToken := m.sessions[recipientID].ActivationToken
	m.mu.RUnlock()
	result, err := m.RegisterConversation(recipientToken, "native-id", "recipient")
	if err != nil || result.Status != RegistrationRegistered || result.Persistence != RegistrationSaved {
		t.Fatalf("registration = %+v, %v", result, err)
	}
	snap, _ := m.GetSession(recipientID)
	if snap.Phase != "starting" || !snap.ProcessStarted || snap.MessageDeliveryAvailable {
		t.Fatalf("starting snapshot = %+v", snap)
	}
	if admission, admissionErr := m.ReserveMessageAdmission(senderToken, recipientID, activation); admissionErr == nil {
		admission.Release()
		t.Fatal("starting recipient accepted a message")
	} else {
		var failure *MessagingError
		if !errors.As(admissionErr, &failure) || failure.Code != "starting" {
			t.Fatalf("starting rejection = %v", admissionErr)
		}
	}
	timer.ready <- deadline
	waitStartupSignal(t, timer.stopped)
	snap, _ = m.GetSession(recipientID)
	if snap.Phase != "active" || !snap.MessageDeliveryAvailable {
		t.Fatalf("active snapshot = %+v", snap)
	}
	admission, err := m.ReserveMessageAdmission(senderToken, recipientID, activation)
	if err != nil {
		t.Fatal(err)
	}
	admission.Release()
}

func TestStaleStartupCallbackCannotEnableReplacement(t *testing.T) {
	m, _, _, recipientID, _ := newReviewMessagingManager(t)
	clock := newControlledStartupClock()
	first := startControlledTimer(t, m, recipientID, clock, clock.Now().Add(10*time.Second))
	<-clock.durations
	m.mu.Lock()
	item := m.sessions[recipientID]
	first.ready <- clock.Now()
	m.cancelStartupLocked(item)
	item.generation++
	m.startStartupLocked(item, clock.Now().Add(10*time.Second))
	m.mu.Unlock()
	second := <-clock.created
	<-clock.durations
	waitStartupSignal(t, first.stopped)
	snap, _ := m.GetSession(recipientID)
	if snap.Phase != "starting" || snap.MessageDeliveryAvailable {
		t.Fatalf("stale timer enabled replacement: %+v", snap)
	}
	second.ready <- clock.Now()
	waitStartupSignal(t, second.stopped)
	snap, _ = m.GetSession(recipientID)
	if snap.Phase != "active" || !snap.MessageDeliveryAvailable {
		t.Fatalf("replacement did not become active: %+v", snap)
	}
}

func TestStartupTimerStopsOnKillAndShutdown(t *testing.T) {
	for _, shutdown := range []bool{false, true} {
		t.Run(map[bool]string{false: "kill", true: "shutdown"}[shutdown], func(t *testing.T) {
			m, _, _, id, _ := newReviewMessagingManager(t)
			clock := newControlledStartupClock()
			timer := startControlledTimer(t, m, id, clock, clock.Now().Add(10*time.Second))
			<-clock.durations
			if shutdown {
				m.Shutdown()
			} else if err := m.Kill(id); err != nil {
				t.Fatal(err)
			}
			waitStartupSignal(t, timer.stopped)
			snap, _ := m.GetSession(id)
			if snap.ProcessStarted || snap.MessageDeliveryAvailable || snap.Phase != "terminated" {
				t.Fatalf("terminated snapshot = %+v", snap)
			}
		})
	}
}

func TestStartupNaturalExitRevokesCredentialAndStopsTimer(t *testing.T) {
	m, _, _, id, _ := newReviewMessagingManager(t)
	clock := newControlledStartupClock()
	timer := startControlledTimer(t, m, id, clock, clock.Now().Add(10*time.Second))
	<-clock.durations
	m.mu.RLock()
	item := m.sessions[id]
	token, d := item.ActivationToken, item.Driver
	m.mu.RUnlock()
	if err := d.Kill(); err != nil {
		t.Fatal(err)
	}
	waitStartupSignal(t, timer.stopped)
	deadline := time.After(3 * time.Second)
	for {
		snap, _ := m.GetSession(id)
		if snap.Status == StatusTerminated {
			if snap.ProcessStarted || snap.MessageDeliveryAvailable || m.HasActivationCredential(token) {
				t.Fatalf("exit left live activation: %+v", snap)
			}
			break
		}
		select {
		case <-deadline:
			t.Fatalf("natural exit not observed: %+v", snap)
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

func TestStartupInvalidationPreventsIntentPublication(t *testing.T) {
	m, _, senderToken, id, activation := newReviewMessagingManager(t)
	admission, err := m.ReserveMessageAdmission(senderToken, id, activation)
	if err != nil {
		t.Fatal(err)
	}
	defer admission.Release()
	if err = m.Kill(id); err != nil {
		t.Fatal(err)
	}
	published := false
	err = m.PublishMessageAdmission(admission, func() error { published = true; return nil })
	if err == nil || published {
		t.Fatalf("invalidated admission published: %v", err)
	}
}

func TestStartupOutputUsesNormalBoundedStream(t *testing.T) {
	m, _, _, id, _ := newReviewMessagingManager(t)
	clock := newControlledStartupClock()
	timer := startControlledTimer(t, m, id, clock, clock.Now().Add(10*time.Second))
	<-clock.durations
	output := make(chan []byte, 2)
	subscription, err := m.SubscribeOutput(id)
	if err != nil {
		t.Fatal(err)
	}
	defer m.UnsubscribeOutput(id, subscription)
	m.mu.RLock()
	generation := m.sessions[id].generation
	m.mu.RUnlock()
	done := make(chan struct{})
	go func() { defer close(done); m.trackOutput(id, generation, output) }()
	payload := []byte("startup-marker\n")
	output <- payload
	select {
	case got := <-subscription:
		if string(got) != string(payload) {
			t.Fatalf("subscriber output = %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("startup output not broadcast")
	}
	if got := string(m.GetScrollback(id)); !strings.Contains(got, string(payload)) {
		t.Fatalf("startup output absent from scrollback: %q", got)
	}
	output <- bytes.Repeat([]byte("x"), ScrollbackLimit+10)
	close(output)
	waitStartupSignal(t, done)
	if got := m.GetScrollback(id); len(got) != ScrollbackLimit {
		t.Fatalf("scrollback length = %d", len(got))
	}
	timer.ready <- clock.Now()
	waitStartupSignal(t, timer.stopped)
}

type drainBlockingDriver struct {
	driver.AgentDriver

	drained <-chan struct{}
}

func (d *drainBlockingDriver) Kill() error { <-d.drained; return nil }
func TestUnpublishedTeardownDrainsAndJoins(t *testing.T) {
	output := make(chan []byte)
	drained := make(chan struct{})
	go func() { output <- []byte("blocked output"); close(drained) }()
	done := make(chan error, 1)
	go func() { done <- teardownUnpublished(&drainBlockingDriver{drained: drained}, output) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("teardown failed to drain output")
	}
}
