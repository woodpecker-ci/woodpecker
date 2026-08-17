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

package compile

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"go.uber.org/multierr"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml"
)

// LintOption configures LintEmitted.
type LintOption func(*lintConfig)

type lintConfig struct {
	maxConfigs int
	maxSize    int
}

// WithMaxConfigs overrides how many configs a compile phase may emit.
func WithMaxConfigs(count int) LintOption {
	return func(c *lintConfig) {
		c.maxConfigs = count
	}
}

// WithMaxSize overrides the cap on the total size of the emitted configs.
func WithMaxSize(size int) LintOption {
	return func(c *lintConfig) {
		c.maxSize = size
	}
}

// LintEmitted validates what a compile phase emitted, before it is merged into
// the pipeline's config set.
//
// Emitted names become Config.Name and are used to build paths, so they are
// held to the same standard as a config file name in the repository. A config
// that itself declares compile steps is rejected: that keeps the number of
// phases statically at two and rules out recursion.
func LintEmitted(emitted []Config, opts ...LintOption) error {
	cfg := &lintConfig{maxConfigs: DefaultMaxConfigs, maxSize: DefaultMaxPayloadSize}
	for _, opt := range opts {
		opt(cfg)
	}

	if len(emitted) > cfg.maxConfigs {
		return fmt.Errorf("compile phase emitted %d configs, the limit is %d", len(emitted), cfg.maxConfigs)
	}

	var (
		errs  error
		total int
		seen  = make(map[string]struct{}, len(emitted))
	)

	for _, config := range emitted {
		if err := lintName(config.Name); err != nil {
			errs = multierr.Append(errs, err)
			continue
		}

		if _, duplicate := seen[config.Name]; duplicate {
			errs = multierr.Append(errs, fmt.Errorf("%w %q", ErrDuplicateName, config.Name))
			continue
		}
		seen[config.Name] = struct{}{}

		total += len(config.Data)
		if total > cfg.maxSize {
			return multierr.Append(errs, fmt.Errorf(
				"compile phase emitted more than %d bytes of configuration", cfg.maxSize,
			))
		}

		// a removal carries no data to inspect
		if len(config.Data) == 0 {
			continue
		}

		if err := lintNoNestedCompile(config); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return errs
}

func lintName(name string) error {
	switch {
	case name == "":
		return fmt.Errorf("emitted config has an empty name")
	case !utf8.ValidString(name):
		return fmt.Errorf("emitted config name is not valid utf-8")
	case strings.ContainsAny(name, `/\`):
		return fmt.Errorf("emitted config name %q must not contain a path separator", name)
	case strings.Contains(name, ".."):
		return fmt.Errorf("emitted config name %q must not contain %q", name, "..")
	case strings.ContainsRune(name, 0):
		return fmt.Errorf("emitted config name %q must not contain a null byte", name)
	case name != strings.TrimSpace(name):
		return fmt.Errorf("emitted config name %q must not be padded with whitespace", name)
	}

	return nil
}

func lintNoNestedCompile(config Config) error {
	parsed, err := yaml.ParseBytes(config.Data)
	if err != nil {
		return fmt.Errorf("emitted config %q is not valid yaml: %w", config.Name, err)
	}

	if len(parsed.Compile.ContainerList) > 0 {
		return fmt.Errorf("emitted config %q declares a compile section, which is not allowed: "+
			"a pipeline has exactly one compile phase", config.Name)
	}

	return nil
}

// Merge applies what a compile phase emitted to the pipeline's source config
// set.
//
// An emitted config overwrites one of the same name, adds a new one, or, when
// its data is empty, removes it. Order is preserved: an overwritten config
// keeps its position, an added one is appended in the order it was emitted.
// Position decides the positional ids of the resulting workflows, so sorting
// here would reorder a pipeline the user wrote deliberately.
//
// Removing a config that is not in the source set is an error. A removal that
// silently does nothing means an extra workflow runs, which is the same class
// of failure as a mistyped depends_on: the pipeline stays green while doing
// more than the configuration asked for. Emitted names come from a generator,
// so a name that matches nothing is a bug in the generator and should surface
// once, loudly.
func Merge(source, emitted []Config) ([]Config, error) {
	index := make(map[string]int, len(source))
	merged := make([]Config, 0, len(source)+len(emitted))
	for _, config := range source {
		index[config.Name] = len(merged)
		merged = append(merged, config)
	}

	removed := make(map[int]struct{})
	seen := make(map[string]struct{}, len(emitted))

	for _, config := range emitted {
		if _, duplicate := seen[config.Name]; duplicate {
			return nil, fmt.Errorf("%w %q", ErrDuplicateName, config.Name)
		}
		seen[config.Name] = struct{}{}

		pos, exists := index[config.Name]

		switch {
		case len(config.Data) == 0 && !exists:
			return nil, fmt.Errorf("%w %q", ErrRemoveUnknownConfig, config.Name)
		case len(config.Data) == 0:
			removed[pos] = struct{}{}
		case exists:
			merged[pos] = config
		default:
			index[config.Name] = len(merged)
			merged = append(merged, config)
		}
	}

	if len(removed) == 0 {
		return merged, nil
	}

	kept := make([]Config, 0, len(merged)-len(removed))
	for pos, config := range merged {
		if _, gone := removed[pos]; gone {
			continue
		}
		kept = append(kept, config)
	}

	return kept, nil
}
