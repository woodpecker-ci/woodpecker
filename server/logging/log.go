package logging

import (
	"context"
	"sync"

	"go.woodpecker-ci.org/woodpecker/server/model"
)

// TODO (bradrydzewski) writing to subscribers is currently a blocking
// operation and does not protect against slow clients from locking
// the stream. This should be resolved.

// TODO (bradrydzewski) implement a mux.Info to fetch information and
// statistics for the multiplexer. Streams, subscribers, etc
// mux.Info()

// TODO (bradrydzewski) refactor code to place publisher and subscriber
// operations in separate files with more encapsulated logic.
// sub.push()
// sub.join()
// sub.start()... event loop

type subscriber struct {
	handler Handler
}

type stream struct {
	sync.Mutex

	stepID int64
	list   []*model.LogEntry
	subs   map[*subscriber]struct{}
	done   chan struct{}
}

type log struct {
	sync.Mutex

	streams map[int64]*stream
}

// New returns a new logger.
func New() Log {
	return &log{
		streams: map[int64]*stream{},
	}
}

func (l *log) Open(_ context.Context, stepID int64) error {
	l.Lock()
	_, ok := l.streams[stepID]
	if !ok {
		l.streams[stepID] = &stream{
			stepID: stepID,
			subs:   make(map[*subscriber]struct{}),
			done:   make(chan struct{}),
		}
	}
	l.Unlock()
	return nil
}

func (l *log) Write(ctx context.Context, stepID int64, logEntry *model.LogEntry) error {
	l.Lock()
	s, ok := l.streams[stepID]
	l.Unlock()
	if !ok {
		return l.Open(ctx, stepID)
	}
	s.Lock()
	s.list = append(s.list, logEntry)
	for sub := range s.subs {
		go sub.handler(logEntry)
	}
	s.Unlock()
	return nil
}

func (l *log) Tail(c context.Context, stepID int64, handler Handler) error {
	l.Lock()
	s, ok := l.streams[stepID]
	l.Unlock()
	if !ok {
		return ErrNotFound
	}

	sub := &subscriber{
		handler: handler,
	}
	s.Lock()
	if len(s.list) != 0 {
		sub.handler(s.list...)
	}
	s.subs[sub] = struct{}{}
	s.Unlock()

	select {
	case <-c.Done():
	case <-s.done:
	}

	s.Lock()
	delete(s.subs, sub)
	s.Unlock()
	return nil
}

func (l *log) Close(_ context.Context, stepID int64) error {
	l.Lock()
	s, ok := l.streams[stepID]
	l.Unlock()
	if !ok {
		return ErrNotFound
	}

	s.Lock()
	close(s.done)
	s.Unlock()

	l.Lock()
	delete(l.streams, stepID)
	l.Unlock()
	return nil
}
