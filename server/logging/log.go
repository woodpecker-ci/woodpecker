// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"context"
	"sync"

	logger "github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

// TODO: (bradrydzewski) writing to subscribers is currently a blocking
// operation and does not protect against slow clients from locking
// the stream. This should be resolved.

//nolint:godot
// TODO: (bradrydzewski) implement a mux.Info to fetch information and
// statistics for the multiplexer. Streams, subscribers, etc
// mux.Info()

//nolint:godot
// TODO: (bradrydzewski) refactor code to place publisher and subscriber
// operations in separate files with more encapsulated logic.
// sub.push()
// sub.join()
// sub.start()... event loop

type subscriber struct {
	receiver LogChan
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

func (l *log) Write(ctx context.Context, stepID int64, entries []*model.LogEntry) error {
	l.Lock()
	s, ok := l.streams[stepID]
	l.Unlock()

	// auto open the stream if it does not exist
	if !ok {
		err := l.Open(ctx, stepID)
		if err != nil {
			return err
		}
		s = l.streams[stepID]
	}

	s.Lock()
	s.list = append(s.list, entries...)
	for sub := range s.subs {
		select {
		case sub.receiver <- entries:
		default:
			logger.Info().Msgf("subscriber channel is full -- dropping logs for step %d", stepID)
		}
	}
	s.Unlock()

	return nil
}

func (l *log) Tail(c context.Context, stepID int64, receiver LogChan) error {
	l.Lock()
	s, ok := l.streams[stepID]
	l.Unlock()
	if !ok {
		return ErrNotFound
	}

	sub := &subscriber{
		receiver: receiver,
	}
	s.Lock()
	if len(s.list) != 0 {
		sub.receiver <- s.list
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
