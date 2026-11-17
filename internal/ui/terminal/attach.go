package terminal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	cterm "github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

const (
	detachKey = 11

	inputBufferSize = 256
	pollTimeout     = 100
)

type Command struct {
	handle   contracts.Attachment
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	stop     chan struct{}
	stopOnce sync.Once
}

var (
	errDetach = errors.New("detach")
)

// New creates a terminal command that owns the attachment.
func New(handle contracts.Attachment) *Command { return &Command{handle: handle} }

func (e *Command) SetStdin(r io.Reader)  { e.stdin = r }
func (e *Command) SetStdout(w io.Writer) { e.stdout = w }
func (e *Command) SetStderr(w io.Writer) { e.stderr = w }

func (e *Command) Run() (result error) {
	e.stop = make(chan struct{})
	defer func() { result = errors.Join(result, e.handle.Close()) }()

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
	defer func() {
		// 2. ON DETACH: Pop the state back to what it was.
		// This restores the terminal to the "classic" state
		// that Kankuro's Bubble Tea dashboard expects.
		result = errors.Join(result, writeAllOnce(e.stdout, []byte("\x1b[<1u")))
		// Belt and braces: also explicitly disable the Kitty keyboard
		// protocol on this transition back to the dashboard. Dashboard.Init
		// does the same on the "navigate back" transition; this covers the
		// "detach a PTY" transition, which does not go through Init.
		result = errors.Join(result, writeAllOnce(e.stdout, []byte("\x1b[=0u")))

		result = errors.Join(result, cterm.Restore(stdinFile.Fd(), oldState))
	}()
	if writeErr := writeAllOnce(e.stdout, []byte("\x1b[>1u")); writeErr != nil {
		return writeErr
	}

	// 3. THE RE-SYNC TRIGGER:
	// Most modern TUIs re-emit their terminal-init sequences when they receive a SIGWINCH.
	if errSync := syncPTYSize(stdinFile, e.handle); errSync != nil {
		return errSync
	}

	resizeCh := make(chan os.Signal, 1)
	signal.Notify(resizeCh, syscall.SIGWINCH)
	defer signal.Stop(resizeCh)

	writeErr := make(chan error, 1)
	outputDone := make(chan struct{})
	go func() {
		defer close(outputDone)
		e.pumpOutput(writeErr)
	}()
	defer func() {
		e.stopOutput()
		<-outputDone
	}()

	return e.pumpInput(stdinFile, resizeCh, writeErr)
}

func (e *Command) printHeader() error {
	initial := e.handle.Initial()
	if len(initial.Scrollback) > 0 {
		if err := writeAllOnce(e.stdout, initial.Scrollback); err != nil {
			return err
		}
		if initial.Scrollback[len(initial.Scrollback)-1] != '\n' {
			if _, err := io.WriteString(e.stdout, "\r\n"); err != nil {
				return err
			}
		}
	}
	if len(initial.Header) > 0 {
		if err := writeAllOnce(e.stdout, initial.Header); err != nil {
			return err
		}
	}
	return nil
}

func (e *Command) pumpOutput(writeErr chan<- error) {
	sendErr := func(err error) {
		select {
		case writeErr <- err:
		default:
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	waiterDone := make(chan struct{})
	defer func() { cancel(); <-waiterDone }()
	go func() {
		defer close(waiterDone)
		select {
		case <-e.stop:
			cancel()
		case <-ctx.Done():
		}
	}()
	for {
		chunk, err := e.handle.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				sendErr(nil)
			} else {
				sendErr(err)
			}
			return
		}
		if len(chunk) > 0 {
			if writeErr := writeAllOnce(e.stdout, chunk); writeErr != nil {
				sendErr(writeErr)
				return
			}
		}
	}
}

//nolint:gocognit // Complex event loop
func (e *Command) pumpInput(stdinFile *os.File, resizeCh <-chan os.Signal, writeErr <-chan error) error {
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
		case <-resizeCh:
			if err := syncPTYSize(stdinFile, e.handle); err != nil {
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

func detectKittyDetach(buf []byte) int {
	kittyPrefix := []byte("\x1b[107;")
	i := bytes.Index(buf, kittyPrefix)
	if i == -1 {
		return -1
	}
	remaining := buf[i+len(kittyPrefix):]
	modBytes, _, found := bytes.Cut(remaining, []byte("u"))
	if !found {
		return -1
	}
	mod, err := strconv.Atoi(string(modBytes))
	if err != nil {
		return -1
	}
	if (mod-1)&4 != 0 {
		return i
	}
	return -1
}

func (e *Command) forwardInput(buf []byte) error {
	// 1. Check for Kitty Protocol sequence: \x1b[107;<MODIFIER>u
	idx := detectKittyDetach(buf)

	// 2. Fallback to legacy Ctrl+K (0x0B / 11)
	if idx == -1 {
		idx = bytes.IndexByte(buf, byte(detachKey))
	}

	// 3. No signal? Send the whole buffer to the agent.
	if idx == -1 {
		if len(buf) > 0 {
			return writeAllOnce(e.handle, buf)
		}
		return nil
	}

	// 4. Signal found!
	// Write only the bytes that came BEFORE the hotkey to the agent.
	if idx > 0 {
		if err := writeAllOnce(e.handle, buf[:idx]); err != nil {
			e.stopOutput()
			return err
		}
	}

	// 5. Trigger the detach logic immediately.
	e.stopOutput()
	return errDetach
}

func writeAllOnce(w io.Writer, data []byte) error {
	n, err := w.Write(data)
	if err == nil && n != len(data) {
		return io.ErrShortWrite
	}
	return err
}

func (e *Command) stopOutput() {
	if e.stop == nil {
		return
	}
	e.stopOnce.Do(func() {
		close(e.stop)
	})
}

func syncPTYSize(termFile *os.File, attachment contracts.Attachment) error {
	if termFile == nil {
		return nil
	}
	columns, rows, err := cterm.GetSize(termFile.Fd())
	if err != nil {
		return err
	}
	return attachment.Resize(rows, columns)
}
