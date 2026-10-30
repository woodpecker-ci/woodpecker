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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepoUpdate_Visibility(t *testing.T) {
	tests := []struct {
		name           string
		stored         Repo
		from           Repo
		wantVisibility RepoVisibility
		wantPrivate    bool
	}{
		{
			name:           "empty source visibility preserves stored value",
			stored:         Repo{Visibility: VisibilityPrivate, IsSCMPrivate: true},
			from:           Repo{Visibility: "", IsSCMPrivate: false},
			wantVisibility: VisibilityPrivate,
			wantPrivate:    true,
		},
		{
			name:           "empty source visibility preserves stored public value",
			stored:         Repo{Visibility: VisibilityPublic, IsSCMPrivate: false},
			from:           Repo{Visibility: "", IsSCMPrivate: false},
			wantVisibility: VisibilityPublic,
			wantPrivate:    false,
		},
		{
			name:           "source can change public to private",
			stored:         Repo{Visibility: VisibilityPublic, IsSCMPrivate: false},
			from:           Repo{Visibility: VisibilityPrivate, IsSCMPrivate: true},
			wantVisibility: VisibilityPrivate,
			wantPrivate:    true,
		},
		{
			name:           "source can change private to public",
			stored:         Repo{Visibility: VisibilityPrivate, IsSCMPrivate: true},
			from:           Repo{Visibility: VisibilityPublic, IsSCMPrivate: false},
			wantVisibility: VisibilityPublic,
			wantPrivate:    false,
		},
		{
			name:           "internal visibility is preserved (not collapsed to private)",
			stored:         Repo{Visibility: VisibilityPublic, IsSCMPrivate: false},
			from:           Repo{Visibility: VisibilityInternal, IsSCMPrivate: true},
			wantVisibility: VisibilityInternal,
			wantPrivate:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.stored
			r.Update(&tt.from)
			assert.Equal(t, tt.wantVisibility, r.Visibility)
			assert.Equal(t, tt.wantPrivate, r.IsSCMPrivate)
		})
	}
}
