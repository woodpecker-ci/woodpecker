// Copyright 2022 Woodpecker Authors
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

package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/server/queue"
)

func TestCreateFilterFunc(t *testing.T) {
	t.Parallel()

	type filterTests struct {
		tsk queue.Task
		exp bool
	}

	tests := []struct {
		struc rpc.Filter
		ft    []filterTests
	}{{
		struc: rpc.Filter{},
		ft: []filterTests{{
			tsk: queue.Task{
				Labels: map[string]string{"platform": "", "repo": "test/woodpecker"},
			},
			exp: true,
		}, {
			tsk: queue.Task{
				Labels: map[string]string{"platform": ""},
			},
			exp: true,
		}},
	}, {
		struc: rpc.Filter{
			Labels: map[string]string{"platform": "abc"},
		},
		ft: []filterTests{{
			tsk: queue.Task{
				Labels: map[string]string{"platform": "def"},
			},
			exp: false,
		}, {
			tsk: queue.Task{
				Labels: map[string]string{"platform": ""},
			},
			exp: true,
		}},
	}, {
		struc: rpc.Filter{
			Expr: "platform = 'abc' OR repo = 'test/woodpecker'",
		},
		ft: []filterTests{{
			tsk: queue.Task{
				Labels: map[string]string{"platform": "", "repo": "test/woodpecker"},
			},
			exp: true,
		}, {
			tsk: queue.Task{
				Labels: map[string]string{"platform": "abc", "repo": "else"},
			},
			exp: true,
		}, {
			tsk: queue.Task{
				Labels: map[string]string{"platform": "also", "repo": "else"},
			},
			exp: false,
		}},
	}}

	for _, test := range tests {
		fn, err := createFilterFunc(test.struc)
		if !assert.NoError(t, err) {
			t.Fail()
		}

		for _, ft := range test.ft {
			assert.EqualValues(t, ft.exp, fn(&ft.tsk))
		}
	}
}
