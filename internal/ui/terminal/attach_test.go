//nolint:testpackage // Exercises package-private resource ownership.
package terminal

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
)

type testAttachment struct {
	output chan []byte
	input  bytes.Buffer
	closed bool
	short  bool
	writes []string
}

func (*testAttachment) Initial() contracts.AttachmentInitial { return contracts.AttachmentInitial{} }
func (a *testAttachment) Next(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case data, ok := <-a.output:
		if !ok {
			return nil, io.EOF
		}
		return data, nil
	}
}
func (a *testAttachment) Write(p []byte) (int, error) {
	if a.short {
		return 0, nil
	}
	a.writes = append(a.writes, string(p))
	return a.input.Write(p)
}
func (*testAttachment) Resize(int, int) error { return nil }
func (a *testAttachment) Close() error        { a.closed = true; return nil }

func TestOutputDrainsBeforeEOF(t *testing.T) {
	a := &testAttachment{output: make(chan []byte, 1)}
	a.output <- []byte("tail")
	close(a.output)
	var output bytes.Buffer
	c := &Command{handle: a, stdout: &output, stop: make(chan struct{})}
	errors := make(chan error, 1)
	c.pumpOutput(errors)
	if err := <-errors; err != nil || output.String() != "tail" {
		t.Fatalf("output = %q, error = %v", output.String(), err)
	}
}

func TestDetachAndShortInput(t *testing.T) {
	a := &testAttachment{}
	c := &Command{handle: a, stop: make(chan struct{})}
	if err := c.forwardInput([]byte("hi\x0bignored")); !errors.Is(err, errDetach) {
		t.Fatalf("detach = %v", err)
	}
	if a.input.String() != "hi" {
		t.Fatalf("input = %q", a.input.String())
	}
	a.short = true
	if err := c.forwardInput([]byte("x")); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short write = %v", err)
	}
}

func TestSetupFailureClosesAttachment(t *testing.T) {
	a := &testAttachment{}
	c := New(a)
	c.SetStdin(bytes.NewBuffer(nil))
	c.SetStdout(io.Discard)
	if err := c.Run(); err == nil || !a.closed {
		t.Fatalf("error = %v, closed = %v", err, a.closed)
	}
}

func TestInputDecoderPreservesNonmatchingAndTimedOutPrefix(t *testing.T) {
	a := &testAttachment{}
	d := inputDecoder{command: &Command{handle: a}}
	for _, chunk := range []string{"a\x1b[10", "8;5ub", "\x1b"} {
		if err := d.forward([]byte(chunk)); err != nil {
			t.Fatal(err)
		}
	}
	if got := a.input.String(); got != "a\x1b[108;5ub" {
		t.Fatalf("forwarded %q", got)
	}
	if err := d.flush(); err != nil {
		t.Fatal(err)
	}
	if got := a.input.String(); got != "a\x1b[108;5ub\x1b" {
		t.Fatalf("flushed %q", got)
	}
}

func TestInputDecoderPreservesWholeOrdinaryKeyWrite(t *testing.T) {
	for _, input := range []string{"\x1b[A", "text\x1b[Aafter", "\x1bx"} {
		t.Run(input, func(t *testing.T) {
			a := &testAttachment{}
			d := inputDecoder{command: &Command{handle: a}}
			if err := d.forward([]byte(input)); err != nil {
				t.Fatal(err)
			}
			if len(a.writes) != 1 || a.writes[0] != input {
				t.Fatalf("writes = %q, want one write of %q", a.writes, input)
			}
		})
	}
}

func TestInputDecoderSplitDetach(t *testing.T) {
	a := &testAttachment{}
	d := inputDecoder{command: &Command{handle: a, stop: make(chan struct{})}}
	if err := d.forward([]byte("before\x1b[107;")); err != nil {
		t.Fatal(err)
	}
	if len(a.writes) != 1 || a.writes[0] != "before" {
		t.Fatalf("first writes = %q", a.writes)
	}
	if err := d.forward([]byte("5uafter")); !errors.Is(err, errDetach) {
		t.Fatalf("second chunk: %v", err)
	}
	if len(a.writes) != 1 {
		t.Fatalf("key or trailing bytes forwarded: %q", a.writes)
	}
}
