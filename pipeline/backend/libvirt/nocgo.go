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

//go:build !cgo || !libvirt

package libvirt

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

const engineName = "libvirt"

type stub struct {
	backend_types.Backend
}

func New() backend_types.Backend {
	return &stub{}
}

func (*stub) Name() string {
	return engineName
}

// IsAvailable is true if explizite selected, so Load can return error.
func (*stub) IsAvailable(ctx context.Context) bool {
	c, ok := ctx.Value(backend_types.CliCommand).(*cli.Command)
	return ok && c.String("backend-engine") == engineName
}

func (*stub) Flags() []cli.Flag { return []cli.Flag }

func (*stub) Load(context.Context) (*backend_types.BackendInfo, error) {
	return nil, fmt.Errorf("libvirt backend is not compiled in, build with LIBVIRT_ENABLED=1")
}
