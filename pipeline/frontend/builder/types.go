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

package builder

import (
	"sort"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

type Item struct {
	Workflow  *Workflow
	Labels    map[string]string
	DependsOn []string
	RunsOn    []string
	Config    *backend_types.Config
}

type Workflow struct {
	ID      int64             `json:"id"`
	PID     int               `json:"pid"`
	Name    string            `json:"name"`
	Environ map[string]string `json:"environ,omitempty"`
	AxisID  int               `json:"-"`
}

type YamlFile struct {
	Name string
	Data []byte
}

type yamlFileList []*YamlFile

func (a yamlFileList) Len() int           { return len(a) }
func (a yamlFileList) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a yamlFileList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func SortYamlFilesByName(fm []*YamlFile) []*YamlFile {
	l := yamlFileList(fm)
	sort.Sort(l)
	return l
}
