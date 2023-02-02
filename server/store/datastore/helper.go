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
	"github.com/woodpecker-ci/woodpecker/server/store/types"
)

// wrapGet return error if err not nil or if requested entry do not exist
func wrapGet(exist bool, err error) error {
	if err != nil {
		return err
	}
	if !exist {
		return types.RecordNotExist
	}
	return nil
}
