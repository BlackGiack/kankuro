package app

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	cterm "github.com/charmbracelet/x/term"
	"github.com/creack/pty"
	"golang.org/x/sys/unix"

	"kankuro/internal/session"
	"kankuro/internal/ui/messages"

	"github.com/charmbracelet/lipgloss"
)

const (
	detachKey = 11

	inputBufferSize = 256
	pollTimeout     = 100
	eofSleep        = 10 * time.Millisecond
)

type attachExec struct {
	sessionID string
	handle    *session.AttachHandle
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer
	stop      chan struct{}
	stopOnce  sync.Once
}

var (
	errDetach = errors.New("detach")
)

func (c *Coordinator) attachSession(id string) tea.Cmd {
	handle, err := c.session.Attach(id)
	if err != nil {
		return func() tea.Msg {
			return messages.AttachFinishedMsg{SessionID: id, Err: err}
		}
	}

	cmd := &attachExec{
		sessionID: id,
		handle:    handle,
	}
	return tea.Exec(cmd, func(err error) tea.Msg {
		return messages.AttachFinishedMsg{SessionID: id, Err: err}
	})
}

func (e *attachExec) SetStdin(r io.Reader)  { e.stdin = r }
func (e *attachExec) SetStdout(w io.Writer) { e.stdout = w }
func (e *attachExec) SetStderr(w io.Writer) { e.stderr = w }

func (e *attachExec) Run() error {
	e.stop = make(chan struct{})
	defer e.stopOutput()
	defer e.handle.Close()

	if err := e.printHeader(); err != nil {
		return err
	}

	stdinFile, ok := e.stdin.(*os.File)
	if !ok || stdinFile == nil {
		return errors.New("attach requires terminal stdin")
	}

	oldState, err := cterm.MakeRaw(stdinFile.Fd())
	if err != nil {
		return err
	}

	// 1. ON ATTACH: Push the state and enable the protocol.
	// This ensures that even on RE-ATTACH, the terminal is synced
	// to the "modern" mode the agent expects.
	_, _ = e.stdout.Write([]byte("\x1b[>1u"))

	defer func() {
		// 2. ON DETACH: Pop the state back to what it was.
		// This restores the terminal to the "classic" state
		// that Kankuro's Bubble Tea dashboard expects.
		_, _ = e.stdout.Write([]byte("\x1b[<1u"))

		_ = cterm.Restore(stdinFile.Fd(), oldState)
	}()

	// 3. THE RE-SYNC TRIGGER:
	// Most modern TUIs re-emit their terminal-init sequences when they receive a SIGWINCH.
	if errSync := syncPTYSize(stdinFile, e.handle.PTY); errSync != nil {
		return errSync
	}

	resizeCh := make(chan os.Signal, 1)
	signal.Notify(resizeCh, syscall.SIGWINCH)
	defer signal.Stop(resizeCh)

	writeErr := make(chan error, 1)
	go e.pumpOutput(writeErr)

	return e.pumpInput(stdinFile, resizeCh, writeErr)
}

const (
	primaryColor = lipgloss.Color("#8E44AD")
)

func (e *attachExec) printHeader() error {
	msg := fmt.Sprintf("\r\n[Attached to session %s. Ctrl+K detaches.]\r\n", e.sessionID)
	styledMsg := lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(msg)
	if _, err := fmt.Fprint(e.stdout, styledMsg); err != nil {
		return err
	}
	if len(e.handle.Scrollback) > 0 {
		if _, err := e.stdout.Write(e.handle.Scrollback); err != nil {
			return err
		}
		if e.handle.Scrollback[len(e.handle.Scrollback)-1] != '\n' {
			if _, err := io.WriteString(e.stdout, "\r\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *attachExec) pumpOutput(writeErr chan<- error) {
	sendErr := func(err error) {
		select {
		case writeErr <- err:
		default:
		}
	}
	for {
		select {
		case <-e.stop:
			return
		case chunk, ok := <-e.handle.Output:
			if !ok {
				sendErr(nil)
				return
			}
			if len(chunk) == 0 {
				continue
			}
			if _, err := e.stdout.Write(chunk); err != nil {
				sendErr(err)
				return
			}
		}
	}
}

//nolint:gocognit // Complex event loop
func (e *attachExec) pumpInput(stdinFile *os.File, resizeCh <-chan os.Signal, writeErr <-chan error) error {
	buf := make([]byte, inputBufferSize)
	fdUint := stdinFile.Fd()
	if fdUint > math.MaxInt32 {
		return fmt.Errorf("file descriptor %d out of range", fdUint)
	}
	fd := int32(fdUint)

	for {
		select {
		case err := <-writeErr:
			return err
		case <-e.handle.Done:
			return nil
		case <-resizeCh:
			if err := syncPTYSize(stdinFile, e.handle.PTY); err != nil {
				return err
			}
		default:
		}

		poll := []unix.PollFd{{Fd: fd, Events: unix.POLLIN}}
		n, err := unix.Poll(poll, pollTimeout)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			return err
		}
		if n == 0 {
			continue
		}
		readN, err := stdinFile.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				time.Sleep(eofSleep)
				continue
			}
			return err
		}
		if readN == 0 {
			continue
		}
		if errForward := e.forwardInput(buf[:readN]); errForward != nil {
			if errors.Is(errForward, errDetach) {
				return nil
			}
			return errForward
		}
	}
}

func (e *attachExec) forwardInput(buf []byte) error {
	idx := -1

	// Standard legacy byte for Ctrl+K (11).
	legacyDetach := []byte{byte(detachKey)}
	// Kitty Protocol sequence for Ctrl+K (\x1b[107;5u).
	kittyDetach := []byte("\x1b[107;5u")

	// 1. Prioritize the modern sequence check.
	if i := bytes.Index(buf, kittyDetach); i != -1 {
		idx = i
	} else if j := bytes.Index(buf, legacyDetach); j != -1 {
		idx = j
	}

	// 2. No signal? Send the whole buffer to the agent.
	if idx == -1 {
		if len(buf) > 0 {
			_, err := e.handle.Input.Write(buf)
			return err
		}
		return nil
	}

	// 3. Signal found!
	// Write only the bytes that came BEFORE the hotkey to the agent.
	if idx > 0 {
		if _, err := e.handle.Input.Write(buf[:idx]); err != nil {
			e.stopOutput() // we still want to stop the output if we have an error
			return err
		}
	}

	// 4. Trigger the detach logic immediately.
	e.stopOutput()
	return errDetach
}

func (e *attachExec) stopOutput() {
	if e.stop == nil {
		return
	}
	e.stopOnce.Do(func() {
		close(e.stop)
	})
}

func syncPTYSize(termFile, ptyFile *os.File) error {
	if termFile == nil || ptyFile == nil {
		return nil
	}
	size, err := pty.GetsizeFull(termFile)
	if err != nil {
		return err
	}
	return pty.Setsize(ptyFile, size)
}
