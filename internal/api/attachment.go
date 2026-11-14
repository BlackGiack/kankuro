package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"

	"github.com/BlackGiack/kankuro/internal/api/contracts"
	"github.com/BlackGiack/kankuro/internal/session"
)

type attachment struct {
	handle *session.AttachHandle
	closed chan struct{}
	once   sync.Once
}

func (a *attachment) Initial() contracts.AttachmentInitial {
	return contracts.AttachmentInitial{Scrollback: bytes.Clone(a.handle.Scrollback), Header: bytes.Clone(a.handle.Header)}
}

func (a *attachment) Next(ctx context.Context) ([]byte, error) {
	select {
	case <-a.closed:
		return nil, io.EOF
	default:
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-a.closed:
		return nil, io.EOF
	case chunk, ok := <-a.handle.Output:
		if !ok {
			return nil, io.EOF
		}
		return chunk, nil
	}
}

func (a *attachment) Write(data []byte) (int, error) {
	select {
	case <-a.closed:
		return 0, errors.New("attachment closed")
	default:
	}
	return a.handle.WriteChecked(data)
}

func (a *attachment) Resize(rows, columns int) error {
	select {
	case <-a.closed:
		return errors.New("attachment closed")
	default:
	}
	return a.handle.ResizeChecked(rows, columns)
}

func (a *attachment) Close() error {
	a.once.Do(func() { close(a.closed); a.handle.Close() })
	return nil
}
