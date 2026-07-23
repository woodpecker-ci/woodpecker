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

package tenki

import (
	"context"
	"io"
	"time"

	"github.com/TenkiCloud/tenki-sdk-go/sandbox"
)

// The interfaces below wrap the parts of the Tenki SDK the backend uses. The
// SDK exposes concrete structs (Client/Session/Command/RunHandle) that talk to
// the network, so these seams let the lifecycle be unit-tested with fakes.
// Production code uses the thin real* adapters at the bottom of this file.

type sandboxClient interface {
	CreateAndWait(ctx context.Context, timeout time.Duration, opts ...sandbox.CreateOption) (sandboxSession, error)
	ListProjectSandboxes(ctx context.Context, projectID string, opts ...sandbox.ListOption) ([]sandboxSession, error)
	WhoAmI(ctx context.Context) (*sandbox.Identity, error)
}

type sandboxSession interface {
	ID() string
	Name() string
	Command(argv []string, opts ...sandbox.RunOptions) sandboxCommand
	CloseIfOpen(ctx context.Context) error
}

type sandboxCommand interface {
	Stream(ctx context.Context) (sandboxRunHandle, error)
	Exec(ctx context.Context) (*sandbox.Result, error)
}

type sandboxRunHandle interface {
	Stdin() io.WriteCloser
	Stdout() io.Reader
	Stderr() io.Reader
	Wait() (*sandbox.Result, error)
	Kill() error
}

// Real adapters wrapping the concrete Tenki SDK types.

type realClient struct{ c *sandbox.Client }

func (r realClient) CreateAndWait(ctx context.Context, timeout time.Duration, opts ...sandbox.CreateOption) (sandboxSession, error) {
	s, err := r.c.CreateAndWait(ctx, timeout, opts...)
	if err != nil {
		return nil, err
	}
	return realSession{s}, nil
}

func (r realClient) ListProjectSandboxes(ctx context.Context, projectID string, opts ...sandbox.ListOption) ([]sandboxSession, error) {
	sessions, err := r.c.ListProjectSandboxes(ctx, projectID, opts...)
	if err != nil {
		return nil, err
	}
	out := make([]sandboxSession, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, realSession{s})
	}
	return out, nil
}

func (r realClient) WhoAmI(ctx context.Context) (*sandbox.Identity, error) {
	return r.c.WhoAmI(ctx)
}

type realSession struct{ s *sandbox.Session }

func (r realSession) ID() string   { return r.s.ID }
func (r realSession) Name() string { return r.s.Name }

func (r realSession) Command(argv []string, opts ...sandbox.RunOptions) sandboxCommand {
	return realCommand{r.s.Command(argv, opts...)}
}

func (r realSession) CloseIfOpen(ctx context.Context) error { return r.s.CloseIfOpen(ctx) }

type realCommand struct{ c *sandbox.Command }

func (r realCommand) Stream(ctx context.Context) (sandboxRunHandle, error) {
	h, err := r.c.Stream(ctx)
	if err != nil {
		return nil, err
	}
	return realRunHandle{h}, nil
}

func (r realCommand) Exec(ctx context.Context) (*sandbox.Result, error) { return r.c.Exec(ctx) }

type realRunHandle struct{ h *sandbox.RunHandle }

func (r realRunHandle) Stdin() io.WriteCloser          { return r.h.Stdin }
func (r realRunHandle) Stdout() io.Reader              { return r.h.Stdout }
func (r realRunHandle) Stderr() io.Reader              { return r.h.Stderr }
func (r realRunHandle) Wait() (*sandbox.Result, error) { return r.h.Wait() }
func (r realRunHandle) Kill() error                    { return r.h.Kill() }
