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

package rpc

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
)

// versionMismatchAction is the policy applied when the server reports a
// different application version than the agent runs.
type versionMismatchAction int

const (
	versionActionWarn     versionMismatchAction = iota // log only, keep going
	versionActionShutdown                              // log error and ask the agent to stop
	versionActionIgnore                                // do nothing
)

// versionMismatchPolicy maps gRPC FullMethod to the action to take on a
// mismatch. Methods that pull or start new work (Next/Init) trigger shutdown
// so the agent does not race ahead under a mismatched runtime; methods that
// belong to an in-flight workflow (Extend, Done, Update, Log, …) fall through
// to warn so the workflow can finish and report. Version() is excluded
// because the maintainer requires it to keep working as the discovery path.
var versionMismatchPolicy = map[string]versionMismatchAction{
	"/proto.Woodpecker/Version": versionActionIgnore,
	"/proto.Woodpecker/Next":    versionActionShutdown,
	"/proto.Woodpecker/Init":    versionActionShutdown,
}

// VersionInterceptor compares the `server-version` response header against
// the agent's own version on every unary call and dispatches per-method
// action via versionMismatchPolicy.
type VersionInterceptor struct {
	agentVersion string
	shutdown     atomic.Pointer[shutdownFunc]
	once         sync.Once
}

// shutdownFunc is the callback invoked at most once when a shutdown-policy
// method observes a mismatch. The reported server version is forwarded so
// the caller can confirm via Version() before tearing things down.
type shutdownFunc func(serverVersion string)

func NewVersionInterceptor(agentVersion string) *VersionInterceptor {
	return &VersionInterceptor{agentVersion: agentVersion}
}

// SetShutdownHandler installs the callback fired on a shutdown-policy
// mismatch. Safe to call after the interceptor is already wired into the
// gRPC client; until set, mismatches are logged but no shutdown is triggered.
func (v *VersionInterceptor) SetShutdownHandler(fn shutdownFunc) {
	v.shutdown.Store(&fn)
}

func (v *VersionInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var md metadata.MD
		opts = append(opts, grpc.Header(&md))
		err := invoker(ctx, method, req, reply, cc, opts...)
		v.check(method, md)
		return err
	}
}

func (v *VersionInterceptor) check(method string, md metadata.MD) {
	if v.agentVersion == "" || v.agentVersion == "dev" {
		return
	}
	values := md.Get(rpc.MetadataKeyServerVersion)
	if len(values) == 0 {
		return
	}
	serverVersion := values[0]
	if serverVersion == "" || serverVersion == "dev" || serverVersion == v.agentVersion {
		return
	}

	action, ok := versionMismatchPolicy[method]
	if !ok {
		action = versionActionWarn
	}

	switch action {
	case versionActionIgnore:
		return
	case versionActionWarn:
		log.Warn().
			Str("method", method).
			Str("server-version", serverVersion).
			Str("agent-version", v.agentVersion).
			Msg("server and agent versions do not match — letting current work finish")
	case versionActionShutdown:
		v.once.Do(func() {
			log.Error().
				Str("method", method).
				Str("server-version", serverVersion).
				Str("agent-version", v.agentVersion).
				Msg("server and agent versions do not match — initiating graceful shutdown")
			if p := v.shutdown.Load(); p != nil && *p != nil {
				(*p)(serverVersion)
			}
		})
	}
}
