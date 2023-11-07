// Copyright 2021 Woodpecker Authors
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

package datastore

import (
	"fmt"
	"runtime"
	"strings"

	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/server/model"
	"go.woodpecker-ci.org/woodpecker/server/store/types"
)

// wrapGet return error if err not nil or if requested entry do not exist
func wrapGet(exist bool, err error) error {
	if !exist {
		return types.RecordNotExist
	}
	if err != nil {
		// we only ask for the function's name if needed, as it's not as preformatted as to just execute it
		fnName := callerName(2)
		return fmt.Errorf("%s: %w", fnName, err)
	}
	return nil
}

// wrapDelete return error if err not nil or if requested entry do not exist
func wrapDelete(c int64, err error) error {
	if c == 0 {
		return types.RecordNotExist
	}
	if err != nil {
		// we only ask for the function's name if needed, as it's not as preformatted as to just execute it
		fnName := callerName(2)
		return fmt.Errorf("%s: %w", fnName, err)
	}
	return nil
}

func (s storage) paginate(p *model.ListOptions) *xorm.Session {
	if p.All {
		return s.engine.NewSession()
	}
	if p.PerPage < 1 {
		p.PerPage = 1
	}
	if p.Page < 1 {
		p.Page = 1
	}
	return s.engine.Limit(p.PerPage, p.PerPage*(p.Page-1))
}

func callerName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	fnName := runtime.FuncForPC(pc).Name()
	pIndex := strings.LastIndex(fnName, ".")
	if pIndex != -1 {
		fnName = fnName[pIndex+1:]
	}
	return fnName
}
