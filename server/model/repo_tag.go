// Copyright 2026 Woodpecker Authors
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

package model

// RepoTag is a repository tag as returned by the forge.
type RepoTag struct {
	Name      string `json:"name"`
	SHA       string `json:"sha,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"` // unix seconds when known
} //	@name	RepoTag
