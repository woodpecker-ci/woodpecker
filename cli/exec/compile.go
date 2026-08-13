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

package exec

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"sync"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/compile"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/logging"
	pipeline_utils "go.woodpecker-ci.org/woodpecker/v3/pipeline/utils"
)

// compileCollector gathers what the compile steps of a local run emitted.
//
// It reuses the same package the agent does, so `cli exec` and a real run apply
// the same rules to the same bytes. Locally there are no secrets to mask, and
// there is no separate log view to fold the block into, so the block lines are
// simply not shown.
type compileCollector struct {
	mu      sync.Mutex
	configs map[string][]compile.Config
	errs    map[string]error
}

func newCompileCollector() *compileCollector {
	return &compileCollector{
		configs: map[string][]compile.Config{},
		errs:    map[string]error{},
	}
}

func (c *compileCollector) logger() logging.Logger {
	return func(step *backend_types.Step, rc io.ReadCloser) error {
		logWriter := NewLineWriter(step.Name, step.UUID)

		if step.Type == backend_types.StepTypeClone {
			return pipeline_utils.CopyLineByLine(logWriter, rc, pipeline.MaxLogLineLength)
		}

		src, emitted := compile.ScanWriter(logWriter, nil)
		copyErr := pipeline_utils.CopyLineByLine(src, rc, pipeline.MaxLogLineLength)

		configs, err := emitted()
		c.set(step.UUID, configs, err)

		return copyErr
	}
}

func (c *compileCollector) set(stepUUID string, configs []compile.Config, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err != nil {
		c.errs[stepUUID] = err
		return
	}
	c.configs[stepUUID] = configs
}

// emitted folds the per-step results together, in UUID order so the merge does
// not depend on which step happened to finish first. A step that failed to
// produce a usable response fails the whole phase.
func (c *compileCollector) emitted() ([]compile.Config, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if failed := slices.Sorted(maps.Keys(c.errs)); len(failed) > 0 {
		return nil, c.errs[failed[0]]
	}

	var configs []compile.Config
	for _, stepUUID := range slices.Sorted(maps.Keys(c.configs)) {
		configs = append(configs, c.configs[stepUUID]...)
	}

	return configs, nil
}

// mergeCompiled applies what a local compile phase emitted to the yaml files the
// run phase is built from.
func mergeCompiled(yamls []*builder.YamlFile, emitted []compile.Config) ([]*builder.YamlFile, error) {
	if err := compile.LintEmitted(emitted); err != nil {
		return nil, err
	}

	source := make([]compile.Config, 0, len(yamls))
	for _, yaml := range yamls {
		source = append(source, compile.Config{Name: builder.SanitizePath(yaml.Name), Data: yaml.Data})
	}

	merged, err := compile.Merge(source, emitted)
	if err != nil {
		return nil, err
	}

	files := make([]*builder.YamlFile, 0, len(merged))
	for _, config := range merged {
		files = append(files, &builder.YamlFile{Name: config.Name, Data: config.Data})
	}

	return files, nil
}

// compilePhaseError wraps a local compile failure so the message names the
// phase rather than looking like an ordinary workflow error.
func compilePhaseError(err error) error {
	return fmt.Errorf("compile phase failed: %w", err)
}
