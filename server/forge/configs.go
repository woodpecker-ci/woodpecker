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

package forge

import (
	"context"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// ConfigsFetcher is an optional interface for forges that can fetch all
// pipeline config candidates with a single request instead of trying them one
// by one with File and Dir.
//
// Implementations: GitHub.
type ConfigsFetcher interface {
	// Configs fetches the config candidates at the pipeline commit (b.Commit).
	// A candidate with a trailing slash is a folder, any other is a file.
	//
	// The result holds one entry per candidate, in the order of candidates:
	//  - nil if the candidate does not exist
	//  - the files of a folder, like Dir returns them
	//  - a single entry with the content of a file, like File returns it
	//
	// On error the caller falls back to File and Dir, so anything those would
	// answer differently must be returned as an error.
	Configs(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, candidates []string) ([][]*types.FileMeta, error)
}
