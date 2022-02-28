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
	"github.com/woodpecker-ci/expr"

	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/server/queue"
)

func createFilterFunc(filter rpc.Filter) (queue.Filter, error) {
	var st *expr.Selector
	var err error

	if filter.Expr != "" {
		st, err = expr.ParseString(filter.Expr)
		if err != nil {
			return nil, err
		}
	}

	return func(task *queue.Task) bool {
		if st != nil {
			match, _ := st.Eval(expr.NewRow(task.Labels))
			return match
		}

		for k, v := range filter.Labels {
			// if platform is not set ignore that filter
			if k == "platform" && task.Labels[k] == "" {
				continue
			}

			if task.Labels[k] != v {
				return false
			}
		}
		return true
	}, nil
}
