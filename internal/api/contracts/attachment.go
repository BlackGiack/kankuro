package contracts

import "context"

// AttachmentInitial contains detached output captured at subscription time.
type AttachmentInitial struct {
	Scrollback []byte
	Header     []byte
}

// Attachment is an owned subscription to one process activation.
type Attachment interface {
	Initial() AttachmentInitial
	Next(context.Context) ([]byte, error)
	Write([]byte) (int, error)
	Resize(rows, columns int) error
	Close() error
}
