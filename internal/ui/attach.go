package ui

import (
	"fmt"
	"io"
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
)

const detachKey = 11

type attachExec struct {
	sessionID string
	handle    *session.AttachHandle
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer
	stop      chan struct{}
	stopOnce  sync.Once
}

var errDetach = fmt.Errorf("detach")

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

	if _, err := fmt.Fprintf(e.stdout, "\r\n[Attached to session %s. Ctrl+K detaches.]\r\n", e.sessionID); err != nil {
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

	stdinFile, _ := e.stdin.(*os.File)
	if stdinFile == nil {
		return fmt.Errorf("attach requires terminal stdin")
	}

	oldState, err := cterm.MakeRaw(stdinFile.Fd())
	if err != nil {
		return err
	}
	defer func() {
		_ = cterm.Restore(stdinFile.Fd(), oldState)
	}()

	if err := syncPTYSize(stdinFile, e.handle.PTY); err != nil {
		return err
	}

	resizeCh := make(chan os.Signal, 1)
	signal.Notify(resizeCh, syscall.SIGWINCH)
	defer signal.Stop(resizeCh)

	writeErr := make(chan error, 1)
	sendErr := func(err error) {
		select {
		case writeErr <- err:
		default:
		}
	}
	go func() {
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
	}()

	buf := make([]byte, 256)
	fd := int(stdinFile.Fd())
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

		poll := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		n, err := unix.Poll(poll, 100)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return err
		}
		if n == 0 {
			continue
		}
		readN, err := stdinFile.Read(buf)
		if err != nil {
			if err == io.EOF {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return err
		}
		if readN == 0 {
			continue
		}
		if err := e.forwardInput(buf[:readN]); err != nil {
			if err == errDetach {
				return nil
			}
			return err
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
