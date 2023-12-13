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

package environments

import (
	"strings"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type builtin struct {
	globals []*model.Environ
}

// Parse returns a EnvironService based on a string slice where key and value are separated by a ":" delimiter.
func Parse(params []string) model.EnvironService {
	var globals []*model.Environ

	for _, item := range params {
		kvPair := strings.SplitN(item, ":", 2)
		if len(kvPair) != 2 {
			// ignore items only containing a key and no value
			log.Warn().Msgf("key '%s' has no value, will be ignored", kvPair[0])
			continue
		}
		globals = append(globals, &model.Environ{Name: kvPair[0], Value: kvPair[1]})
	}
	return &builtin{globals}
}

func (b *builtin) EnvironList(_ *model.Repo) ([]*model.Environ, error) {
	return b.globals, nil
}
