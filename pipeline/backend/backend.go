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

package backend

import (
	"context"
	"fmt"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func FindBackend(ctx context.Context, backends []types.Backend, backendName string) (types.Backend, error) {
	if backendName == "auto-detect" {
		for _, engine := range backends {
			if engine.IsAvailable(ctx) {
				return engine, nil
			}
		}

		return nil, fmt.Errorf("can't detect an available backend engine")
	}

	for _, engine := range backends {
		if engine.Name() == backendName {
			return engine, nil
		}
	}

	return nil, fmt.Errorf("backend engine '%s' not found", backendName)
}
