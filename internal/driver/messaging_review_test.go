//nolint:testpackage // These cancellation checks need to occupy the private input gate deterministically.
package driver

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

//nolint:gocognit // Both cancellation boundaries share the same no-byte assertion.
func TestSubmitPromptCanceledBeforeOrWhileWaitingWritesNoBytes(t *testing.T) {
	for _, test := range []struct {
		name       string
		occupyGate bool
	}{
		{name: "already canceled"},
		{name: "canceled while waiting", occupyGate: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			reader, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()
			defer writer.Close()

			d := NewGenericCLIDriver()
			d.pty = writer
			d.inputGate = make(chan struct{}, 1)
			if test.occupyGate {
				d.inputGate <- struct{}{}
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if !test.occupyGate {
				cancel()
			}
			done := make(chan error, 1)
			started := make(chan struct{})
			go func() {
				close(started)
				_, submitErr := d.SubmitPrompt(ctx, "message", "sender", "do not execute")
				done <- submitErr
			}()
			if test.occupyGate {
				<-started
				time.Sleep(20 * time.Millisecond)
				cancel()
			}
			if err = <-done; !errors.Is(err, context.Canceled) {
				t.Fatalf("SubmitPrompt error = %v, want context.Canceled", err)
			}
			if err = reader.SetReadDeadline(time.Now().Add(50 * time.Millisecond)); err != nil {
				t.Fatal(err)
			}
			buffer := make([]byte, 256)
			if count, readErr := reader.Read(buffer); count != 0 || !isTimeout(readErr) {
				t.Fatalf("canceled prompt wrote %d bytes (read error %v): %q", count, readErr, buffer[:count])
			}
		})
	}
}

func TestCodexPromptCanceledBetweenTextAndEnterIsUncertain(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()

	d := NewGenericCLIDriver()
	d.pty = writer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan struct {
		evidence SubmissionEvidence
		err      error
	}, 1)
	go func() {
		evidence, submitErr := d.submitPrompt(ctx, "message", "sender", "prompt text", 250*time.Millisecond)
		result <- struct {
			evidence SubmissionEvidence
			err      error
		}{evidence, submitErr}
	}()

	prompt := make([]byte, 256)
	if err = reader.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	count, err := reader.Read(prompt)
	if err != nil {
		t.Fatalf("read text phase: %v", err)
	}
	if strings.Contains(string(prompt[:count]), "\r") {
		t.Fatalf("text phase included Enter: %q", prompt[:count])
	}
	cancel()
	completed := <-result
	if !errors.Is(completed.err, context.Canceled) || completed.evidence.State != SubmissionUncertain ||
		completed.evidence.BytesWritten != count {
		t.Fatalf("canceled between phases = %#v, %v", completed.evidence, completed.err)
	}
	if err = reader.SetReadDeadline(time.Now().Add(50 * time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	var trailing [1]byte
	if trailingCount, readErr := reader.Read(trailing[:]); trailingCount != 0 || !isTimeout(readErr) {
		t.Fatalf("canceled Codex submission wrote Enter: count=%d err=%v", trailingCount, readErr)
	}
}

func TestCodexPromptCancellationAtEnterBoundaryDoesNotWriteEnter(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	text := []byte("marked prompt")
	written, err := writeAll(writer, text)
	if err != nil || written != len(text) {
		t.Fatalf("write prompt text: bytes=%d err=%v", written, err)
	}
	if err = waitForPromptEnter(ctx, time.Millisecond); err != nil {
		t.Fatalf("wait for Enter pause: %v", err)
	}
	// Model cancellation after the pause timer fires and before Enter is admitted.
	cancel()
	enterWritten, err := writePromptEnter(ctx, writer)
	evidence := classifySubmission(written+enterWritten, len(text), time.Millisecond, err)
	if !errors.Is(err, context.Canceled) || enterWritten != 0 || evidence.State != SubmissionUncertain {
		t.Fatalf("Enter boundary = bytes %d, err %v; want zero bytes and cancellation", enterWritten, err)
	}

	buffer := make([]byte, len(text)+1)
	if err = reader.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	count, err := io.ReadFull(reader, buffer[:len(text)])
	if err != nil || count != len(text) || string(buffer[:count]) != string(text) {
		t.Fatalf("text phase = %q, bytes %d, err %v", buffer[:count], count, err)
	}
	if err = reader.SetReadDeadline(time.Now().Add(50 * time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	if count, err = reader.Read(buffer[len(text):]); count != 0 || !isTimeout(err) {
		t.Fatalf("canceled boundary wrote extra input: bytes %d err %v", count, err)
	}
}

func TestCodexPromptPauseSerializesAttachedInput(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	d := NewGenericCLIDriver()
	d.pty = writer

	submission := make(chan error, 1)
	go func() {
		evidence, submitErr := d.submitPrompt(context.Background(), "message", "sender", "prompt text", 250*time.Millisecond)
		if submitErr == nil && evidence.State != SubmissionComplete {
			submitErr = errors.New("prompt was not completely submitted")
		}
		submission <- submitErr
	}()
	prompt := make([]byte, 256)
	if err = reader.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	count, err := reader.Read(prompt)
	if err != nil {
		t.Fatalf("read text phase: %v", err)
	}
	if strings.Contains(string(prompt[:count]), "\r") {
		t.Fatalf("text phase included Enter: %q", prompt[:count])
	}
	attached := make(chan error, 1)
	go func() { attached <- d.SendInput("human input") }()
	if err = <-submission; err != nil {
		t.Fatalf("submit prompt: %v", err)
	}
	if err = <-attached; err != nil {
		t.Fatalf("send attached input: %v", err)
	}
	wantTail := "\rhuman input"
	tail := make([]byte, len(wantTail))
	if err = reader.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err = io.ReadFull(reader, tail); err != nil {
		t.Fatalf("read Enter and attached input: %v", err)
	}
	if string(tail) != wantTail {
		t.Fatalf("input after prompt = %q, want %q", tail, wantTail)
	}
}

func isTimeout(err error) bool {
	var netErr interface{ Timeout() bool }
	return errors.As(err, &netErr) && netErr.Timeout()
}
