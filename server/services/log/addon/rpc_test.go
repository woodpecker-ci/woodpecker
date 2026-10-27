// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package addon

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	service_log "go.woodpecker-ci.org/woodpecker/v3/server/services/log"
	log_mocks "go.woodpecker-ci.org/woodpecker/v3/server/services/log/mocks"
)

// newRPCService connects the addon RPC client to an RPC server backed by impl
// in process, so every call takes the same JSON and net/rpc path as with a
// real addon binary.
func newRPCService(t *testing.T, impl service_log.Service) service_log.Service {
	t.Helper()

	client, _ := plugin.TestPluginRPCConn(t, map[string]plugin.Plugin{
		pluginKey: &Plugin{Impl: impl},
	}, nil)
	t.Cleanup(func() { _ = client.Close() })

	raw, err := client.Dispense(pluginKey)
	require.NoError(t, err)
	s, ok := raw.(service_log.Service)
	require.True(t, ok)
	return s
}

func TestRPCRoundTrip(t *testing.T) {
	step := &model.Step{ID: 1, UUID: "uuid", PipelineID: 2, PID: 3, Name: "build", State: model.StatusRunning}
	entries := []*model.LogEntry{
		{ID: 1, StepID: 1, Time: 10, Line: 0, Data: []byte("hello\n"), Type: model.LogEntryStdout},
		{ID: 2, StepID: 1, Time: 11, Line: 1, Data: []byte{0x00, 0xff}, Type: model.LogEntryExitCode},
	}

	t.Run("log find", func(t *testing.T) {
		impl := log_mocks.NewMockService(t)
		impl.On("LogFind", mock.Anything, step).Return(entries, nil)

		got, err := newRPCService(t, impl).LogFind(t.Context(), step)
		require.NoError(t, err)
		assert.Equal(t, entries, got)
	})

	t.Run("log append", func(t *testing.T) {
		impl := log_mocks.NewMockService(t)
		impl.On("LogAppend", mock.Anything, step, entries).Return(nil)

		assert.NoError(t, newRPCService(t, impl).LogAppend(t.Context(), step, entries))
	})

	t.Run("log delete", func(t *testing.T) {
		impl := log_mocks.NewMockService(t)
		impl.On("LogDelete", mock.Anything, step).Return(nil)

		assert.NoError(t, newRPCService(t, impl).LogDelete(t.Context(), step))
	})

	t.Run("step finished", func(t *testing.T) {
		impl := log_mocks.NewMockService(t)
		impl.On("StepFinished", mock.Anything, step).Return(nil)

		assert.NoError(t, newRPCService(t, impl).StepFinished(t.Context(), step))
	})
}

func TestRPCErrors(t *testing.T) {
	errStore := errors.New("log store unavailable")

	impl := log_mocks.NewMockService(t)
	impl.On("LogFind", mock.Anything, mock.Anything).Return(nil, errStore)
	impl.On("LogAppend", mock.Anything, mock.Anything, mock.Anything).Return(errStore)
	impl.On("LogDelete", mock.Anything, mock.Anything).Return(errStore)
	impl.On("StepFinished", mock.Anything, mock.Anything).Return(errStore)

	s := newRPCService(t, impl)
	step := &model.Step{ID: 1}

	// the error message is the only thing that crosses the RPC boundary
	_, err := s.LogFind(t.Context(), step)
	assert.ErrorContains(t, err, "log store unavailable")
	assert.ErrorContains(t, s.LogAppend(t.Context(), step, []*model.LogEntry{{Line: 1}}), "log store unavailable")
	assert.ErrorContains(t, s.LogDelete(t.Context(), step), "log store unavailable")
	assert.ErrorContains(t, s.StepFinished(t.Context(), step), "log store unavailable")
}

// TestRPCCanceled ensures a hung addon cannot block the server: a call returns once its context is done.
func TestRPCCanceled(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	called := make(chan struct{})

	impl := log_mocks.NewMockService(t)
	impl.On("StepFinished", mock.Anything, mock.Anything).Run(func(mock.Arguments) {
		close(called)
		<-release // the addon never answers
	}).Return(nil)
	s := newRPCService(t, impl)

	ctx, cancel := context.WithCancelCause(t.Context())
	go func() {
		<-called
		cancel(errors.New("server shutting down"))
	}()
	errc := make(chan error, 1)
	go func() { errc <- s.StepFinished(ctx, &model.Step{ID: 1}) }()
	select {
	case err := <-errc:
		assert.EqualError(t, err, "server shutting down")
	case <-time.After(5 * time.Second):
		t.Fatal("call still blocked after cancel")
	}
}
