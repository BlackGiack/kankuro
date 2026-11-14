package contracts

import session "github.com/BlackGiack/kankuro/internal/session"

type Alias = session.Snapshot
type Inner struct{ Values map[string][]*session.Snapshot }
type Outer struct{ Inner }
type Box[T any] struct{ Value T }

func (b Box[T]) Result() session.Snapshot { return session.Snapshot{} }

type CycleA struct{ Next *CycleB }
type CycleB struct {
	Next  *CycleA
	Value session.Snapshot
}
