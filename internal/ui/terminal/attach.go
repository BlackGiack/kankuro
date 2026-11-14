package terminal

import (
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
	"time"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	cterm "github.com/charmbracelet/x/term"
	"golang.org/x/sys/unix"
)

const (
	detachKey = 11

	inputBufferSize        = 256
	pollTimeout            = 100
	escapeAmbiguityTimeout = 30 * time.Millisecond
	kittyPrefixTimeout     = 750 * time.Millisecond
	kittyDetachPrefix      = "\x1b[107;"
	kittyMaxModifierDigits = 9
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
	input := inputDecoder{command: e}

	for {
		if len(input.pending) > 0 && !time.Now().Before(input.deadline()) {
			if err := input.flush(); err != nil {
				return err
			}
		}
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
		n, err := unix.Poll(poll, input.pollTimeout())
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
		if errForward := input.forward(buf[:readN]); errForward != nil {
			if errors.Is(errForward, errDetach) {
				return nil
			}
			return errForward
		}
	}
}

// inputDecoder holds only a possible Kitty Ctrl+K prefix. A timeout releases
// incomplete sequences so an ordinary Escape key does not remain buffered.
type inputDecoder struct {
	command      *Command
	pending      []byte
	pendingSince time.Time
}

func (d *inputDecoder) flush() error {
	if len(d.pending) == 0 {
		return nil
	}
	err := writeAllOnce(d.command.handle, d.pending)
	d.pending = nil
	return err
}

func (d *inputDecoder) deadline() time.Time {
	timeout := escapeAmbiguityTimeout
	if len(d.pending) >= len(kittyDetachPrefix) {
		timeout = kittyPrefixTimeout
	}
	return d.pendingSince.Add(timeout)
}

func (d *inputDecoder) pollTimeout() int {
	if len(d.pending) == 0 {
		return pollTimeout
	}
	remaining := time.Until(d.deadline())
	if remaining <= 0 {
		return 0
	}
	milliseconds := int((remaining + time.Millisecond - 1) / time.Millisecond)
	return min(pollTimeout, milliseconds)
}

//nolint:gocognit,nestif // Streaming key parsing preserves ordering across partial reads.
func (d *inputDecoder) forward(data []byte) error {
	plain := make([]byte, 0, len(d.pending)+len(data))
	for _, b := range data {
		// A rejected prefix joins this read's ordinary bytes before the shared gate.
		for {
			if len(d.pending) == 0 {
				if b == detachKey {
					if len(plain) > 0 {
						if err := writeAllOnce(d.command.handle, plain); err != nil {
							return err
						}
					}
					d.command.stopOutput()
					return errDetach
				}
				if b == kittyDetachPrefix[0] {
					d.pending = append(d.pending, b)
					d.pendingSince = time.Now()
				} else {
					plain = append(plain, b)
				}
				break
			}
			n := len(d.pending)
			if n < len(kittyDetachPrefix) && b == kittyDetachPrefix[n] {
				d.pending = append(d.pending, b)
				if len(d.pending) == len(kittyDetachPrefix) {
					d.pendingSince = time.Now()
				}
				break
			}
			if n >= len(kittyDetachPrefix) && b >= '0' && b <= '9' && n-len(kittyDetachPrefix) < kittyMaxModifierDigits {
				d.pending = append(d.pending, b)
				break
			}
			if n > len(kittyDetachPrefix) && b == 'u' {
				mod, parseErr := strconv.Atoi(string(d.pending[len(kittyDetachPrefix):]))
				if parseErr == nil && mod > 0 && (mod-1)&4 != 0 {
					if len(plain) > 0 {
						if err := writeAllOnce(d.command.handle, plain); err != nil {
							return err
						}
					}
					d.pending = nil
					d.command.stopOutput()
					return errDetach
				}
			}
			plain = append(plain, d.pending...)
			d.pending = nil
			// Reconsider the mismatching byte: it can itself begin a detach sequence.
		}
	}
	if len(plain) > 0 {
		return writeAllOnce(d.command.handle, plain)
	}
	return nil
}

func (e *Command) forwardInput(buf []byte) error {
	d := inputDecoder{command: e}
	if err := d.forward(buf); err != nil {
		return err
	}
	return d.flush()
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
