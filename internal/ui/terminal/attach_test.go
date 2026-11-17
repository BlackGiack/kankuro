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
