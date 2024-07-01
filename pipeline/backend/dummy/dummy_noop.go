// Copyright 2024 Woodpecker Authors
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

//go:build !test
// +build !test

package dummy

import (
	"context"
	"errors"
	"io"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

type noop struct{}

var ErrOnCompileExcluded = errors.New("the dummy backend engine was excluded on compile time")

// New returns a dummy backend.
func New() types.Backend {
	return &noop{}
}

func (e *noop) Name() string {
	return "dummy"
}

func (e *noop) IsAvailable(context.Context) bool {
	return false
}

func (e *noop) Flags() []cli.Flag {
	return nil
}

// Load new client for Docker Backend using environment variables.
func (e *noop) Load(context.Context) (*types.BackendInfo, error) {
	return nil, ErrOnCompileExcluded
}

func (e *noop) SetupWorkflow(context.Context, *types.Config, string) error {
	return ErrOnCompileExcluded
}

func (e *noop) StartStep(context.Context, *types.Step, string) error {
	return ErrOnCompileExcluded
}

func (e *noop) WaitStep(context.Context, *types.Step, string) (*types.State, error) {
	return nil, ErrOnCompileExcluded
}

func (e *noop) TailStep(context.Context, *types.Step, string) (io.ReadCloser, error) {
	return nil, ErrOnCompileExcluded
}

func (e *noop) DestroyStep(context.Context, *types.Step, string) error {
	return ErrOnCompileExcluded
}

func (e *noop) DestroyWorkflow(context.Context, *types.Config, string) error {
	return ErrOnCompileExcluded
}
