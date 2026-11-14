//nolint:testpackage // Exercises package-private resource ownership.
package api

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/BlackGiack/kankuro/internal/session"
)

func TestAttachmentDrainsOutputAndOwnsInitial(t *testing.T) {
	output := make(chan []byte, 1)
	output <- []byte("trailing")
	close(output)
	closed := 0
	a := &attachment{
		handle: &session.AttachHandle{Output: output, Scrollback: []byte("initial"), Close: func() { closed++ }},
		closed: make(chan struct{}),
	}
	initial := a.Initial()
	initial.Scrollback[0] = 'X'
	if string(a.Initial().Scrollback) != "initial" {
		t.Fatal("initial output was shared")
	}
	chunk, err := a.Next(context.Background())
	if err != nil || string(chunk) != "trailing" {
		t.Fatalf("output = %q, %v", chunk, err)
	}
	if _, err = a.Next(context.Background()); !errors.Is(err, io.EOF) {
		t.Fatalf("EOF = %v", err)
	}
	_ = a.Close()
	_ = a.Close()
	if closed != 1 {
		t.Fatalf("close called %d times", closed)
	}
}

func TestAttachmentNextCancellation(t *testing.T) {
	a := &attachment{handle: &session.AttachHandle{Output: make(chan []byte), Close: func() {}}, closed: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := a.Next(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Next error = %v", err)
	}
	_ = a.Close()
}
