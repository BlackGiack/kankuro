package ui

import (
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

var errDetach = errors.New("detach")

func (a *App) attachSession(id string) tea.Cmd {
	handle, err := a.session.Attach(id)
	if err != nil {
		return func() tea.Msg {
			return attachFinishedMsg{sessionID: id, err: err}
		}
	}

	cmd := &attachExec{
		sessionID: id,
		handle:    handle,
	}
	return tea.Exec(cmd, func(err error) tea.Msg {
		return attachFinishedMsg{sessionID: id, err: err}
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
	defer func() {
		_ = cterm.Restore(stdinFile.Fd(), oldState)
	}()

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
	start := 0
	for i, b := range buf {
		if b != detachKey {
			continue
		}
		if i > start {
			if _, err := e.handle.Input.Write(buf[start:i]); err != nil {
				return err
			}
		}
		e.stopOutput()
		return errDetach
	}
	if len(buf) > 0 {
		_, err := e.handle.Input.Write(buf)
		return err
	}
	return nil
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
