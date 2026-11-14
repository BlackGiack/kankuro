//go:build linux

//nolint:testpackage // Exercises terminal resource lifecycle with a PTY.
package terminal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

type ptyAttachment struct {
	output chan []byte
	sizes  chan [2]int
	input  bytes.Buffer
}

func (*ptyAttachment) Initial() contracts.AttachmentInitial {
	return contracts.AttachmentInitial{Scrollback: []byte("initial"), Header: []byte("header\r\n")}
}
func (a *ptyAttachment) Next(ctx context.Context) ([]byte, error) {
	select {
	case b, ok := <-a.output:
		if !ok {
			return nil, io.EOF
		}
		return b, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func (a *ptyAttachment) Write(p []byte) (int, error) { return a.input.Write(p) }
func (a *ptyAttachment) Resize(r, c int) error {
	select {
	case a.sizes <- [2]int{r, c}:
	default:
	}
	return nil
}
func (*ptyAttachment) Close() error { return nil }

type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}
func (b *lockedBuffer) String() string { b.mu.Lock(); defer b.mu.Unlock(); return b.b.String() }

//nolint:gocognit // Exercises complete PTY setup and restoration.
func TestTerminalPTYDetachAndRestoration(t *testing.T) {
	for _, key := range []string{"\x0b", "\x1b[107;5u"} {
		t.Run(key, func(t *testing.T) {
			master, slave, err := pty.Open()
			if err != nil {
				t.Fatal(err)
			}
			defer master.Close()
			defer slave.Close()
			if err = pty.Setsize(slave, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
				t.Fatal(err)
			}
			before, err := unix.IoctlGetTermios(int(slave.Fd()), unix.TCGETS)
			if err != nil {
				t.Fatal(err)
			}
			a := &ptyAttachment{output: make(chan []byte, 1), sizes: make(chan [2]int, 1)}
			var output lockedBuffer
			c := New(a)
			c.SetStdin(slave)
			c.SetStdout(&output)
			done := make(chan error, 1)
			go func() { done <- c.Run() }()
			select {
			case size := <-a.sizes:
				if size != [2]int{24, 80} {
					t.Fatalf("size %v", size)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("terminal setup timed out")
			}
			if err = pty.Setsize(slave, &pty.Winsize{Rows: 30, Cols: 100}); err != nil {
				t.Fatal(err)
			}
			deadline := time.After(3 * time.Second)
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()
		resize:
			for {
				select {
				case <-ticker.C:
					if err = unix.Kill(unix.Getpid(), unix.SIGWINCH); err != nil {
						t.Fatal(err)
					}
				case size := <-a.sizes:
					if size != [2]int{30, 100} {
						t.Fatalf("resized to %v", size)
					}
					break resize
				case <-deadline:
					t.Fatal("terminal resize timed out")
				}
			}
			if _, err = master.WriteString("x" + key); err != nil {
				t.Fatal(err)
			}
			select {
			case err = <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("detach timed out")
			}
			after, err := unix.IoctlGetTermios(int(slave.Fd()), unix.TCGETS)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatal("terminal state was not restored")
			}
			if a.input.String() != "x" {
				t.Fatalf("forwarded input %q", a.input.String())
			}
			if !strings.Contains(output.String(), "initial\r\nheader\r\n") {
				t.Fatalf("header output %q", output.String())
			}
		})
	}
}

type failAfterWriter struct{ writes int }

func (w *failAfterWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == 5 {
		return 0, io.ErrClosedPipe
	}
	return len(p), nil
}

//nolint:gocognit // Exercises terminal EOF and failure restoration.
func TestTerminalPTYOutputEOFAndErrorRestore(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint("write-failure-", fail), func(t *testing.T) {
			master, slave, err := pty.Open()
			if err != nil {
				t.Fatal(err)
			}
			defer master.Close()
			defer slave.Close()
			if err = pty.Setsize(slave, &pty.Winsize{Rows: 40, Cols: 120}); err != nil {
				t.Fatal(err)
			}
			before, err := unix.IoctlGetTermios(int(slave.Fd()), unix.TCGETS)
			if err != nil {
				t.Fatal(err)
			}
			a := &ptyAttachment{output: make(chan []byte, 1), sizes: make(chan [2]int, 1)}
			var out lockedBuffer
			c := New(a)
			c.SetStdin(slave)
			if fail {
				c.SetStdout(&failAfterWriter{})
			} else {
				c.SetStdout(&out)
			}
			done := make(chan error, 1)
			go func() { done <- c.Run() }()
			select {
			case size := <-a.sizes:
				if size != [2]int{40, 120} {
					t.Fatalf("size %v", size)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("setup timed out")
			}
			a.output <- []byte("trailing")
			close(a.output)
			select {
			case err = <-done:
				if fail && !errors.Is(err, io.ErrClosedPipe) {
					t.Fatalf("error %v", err)
				}
				if !fail && err != nil {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("output EOF timed out")
			}
			after, err := unix.IoctlGetTermios(int(slave.Fd()), unix.TCGETS)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) {
				t.Fatal("terminal state was not restored")
			}
			if !fail && !strings.Contains(out.String(), "trailing") {
				t.Fatalf("missing output %q", out.String())
			}
		})
	}
}

func TestTerminalPTYSplitKittyDetach(t *testing.T) {
	master, slave, err := pty.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer master.Close()
	defer slave.Close()
	if err = pty.Setsize(slave, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		t.Fatal(err)
	}
	a := &ptyAttachment{output: make(chan []byte), sizes: make(chan [2]int, 1)}
	var output lockedBuffer
	c := New(a)
	c.SetStdin(slave)
	c.SetStdout(&output)
	done := make(chan error, 1)
	go func() { done <- c.Run() }()
	select {
	case <-a.sizes:
	case <-time.After(3 * time.Second):
		t.Fatal("setup timeout")
	}
	if _, err = master.WriteString("before\x1b[107;"); err != nil {
		t.Fatal(err)
	}
	// Exercise a delayed continuation; the decoder test forces the split deterministically.
	time.Sleep(250 * time.Millisecond)
	if _, err = master.WriteString("5uafter"); err != nil {
		t.Fatal(err)
	}
	select {
	case err = <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("split key did not detach")
	}
	if got := a.input.String(); got != "before" {
		t.Fatalf("child input = %q, want only preceding input", got)
	}
}

type observedInputAttachment struct {
	*ptyAttachment

	writes chan string
}

func (a *observedInputAttachment) Write(p []byte) (int, error) {
	a.writes <- string(p)
	return len(p), nil
}

func TestTerminalPTYReleasesIncompleteEscape(t *testing.T) {
	master, slave, err := pty.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer master.Close()
	defer slave.Close()
	if err = pty.Setsize(slave, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		t.Fatal(err)
	}
	a := &observedInputAttachment{
		ptyAttachment: &ptyAttachment{output: make(chan []byte), sizes: make(chan [2]int, 1)},
		writes:        make(chan string, 2),
	}
	var output lockedBuffer
	c := New(a)
	c.SetStdin(slave)
	c.SetStdout(&output)
	done := make(chan error, 1)
	go func() { done <- c.Run() }()
	select {
	case <-a.sizes:
	case <-time.After(3 * time.Second):
		t.Fatal("setup timeout")
	}
	start := time.Now()
	if _, err = master.WriteString("\x1b"); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-a.writes:
		if got != "\x1b" {
			t.Fatalf("input %q", got)
		}
		if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
			t.Fatalf("Escape took %v to reach the child", elapsed)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("incomplete Escape was not released")
	}
	if _, err = master.WriteString("\x0b"); err != nil {
		t.Fatal(err)
	}
	select {
	case err = <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("detach timeout")
	}
}
