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

// Package jsonnet compiles jsonnet pipeline configs into workflow configs.
package jsonnet

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"
)

const (
	fileExtension = ".jsonnet"
	nameKey       = "name"
	maxStack      = 500
	// Limits the compiled JSON size to protect the server from configs
	// that expand into unreasonably large documents.
	maxOutputSize = 1 << 20 // 1 MiB
)

// File is a single workflow config produced by compiling a jsonnet config.
type File struct {
	Name string
	Data []byte
}

// IsJsonnetFile reports whether name refers to a jsonnet config file.
func IsJsonnetFile(name string) bool {
	return filepath.Ext(name) == fileExtension
}

// Compile evaluates a jsonnet config and returns one workflow config per
// produced pipeline object. The jsonnet may evaluate to a single object
// (one workflow, optionally named via a top-level "name" field) or to an
// array of objects (each requiring a "name" field). The "name" field is
// stripped from the resulting workflow config. Imports, external variables
// and native functions are not supported.
func Compile(name string, data []byte) ([]*File, error) {
	vm := jsonnet.MakeVM()
	vm.MaxStack = maxStack
	vm.Importer(&noImporter{})

	output, err := vm.EvaluateAnonymousSnippet(filepath.Base(name), string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to compile jsonnet config %s: %w", name, err)
	}
	if len(output) > maxOutputSize {
		return nil, fmt.Errorf("jsonnet config %s compiles to %d bytes, exceeding the limit of %d bytes", name, len(output), maxOutputSize)
	}

	var compiled any
	if err := json.Unmarshal([]byte(output), &compiled); err != nil {
		return nil, fmt.Errorf("failed to parse compiled jsonnet config %s: %w", name, err)
	}

	switch value := compiled.(type) {
	case map[string]any:
		file, err := workflowFile(name, value, defaultWorkflowName(name), 0)
		if err != nil {
			return nil, err
		}
		return []*File{file}, nil
	case []any:
		files := make([]*File, 0, len(value))
		seen := make(map[string]struct{}, len(value))
		for i, element := range value {
			workflow, ok := element.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("jsonnet config %s: element %d is not an object", name, i)
			}
			file, err := workflowFile(name, workflow, "", i)
			if err != nil {
				return nil, err
			}
			if _, exists := seen[file.Name]; exists {
				return nil, fmt.Errorf("jsonnet config %s: duplicate workflow name %q", name, file.Name)
			}
			seen[file.Name] = struct{}{}
			files = append(files, file)
		}
		return files, nil
	default:
		return nil, fmt.Errorf("jsonnet config %s must evaluate to an object or an array of objects", name)
	}
}

func workflowFile(configName string, workflow map[string]any, fallbackName string, index int) (*File, error) {
	name := fallbackName
	if rawName, ok := workflow[nameKey]; ok {
		stringName, ok := rawName.(string)
		if !ok {
			return nil, fmt.Errorf("jsonnet config %s: workflow %d has a non-string name", configName, index)
		}
		name = stringName
		delete(workflow, nameKey)
	}
	if name == "" {
		return nil, fmt.Errorf("jsonnet config %s: workflow %d has no name", configName, index)
	}
	if strings.ContainsAny(name, "/\\") || strings.HasPrefix(name, ".") {
		return nil, fmt.Errorf("jsonnet config %s: invalid workflow name %q", configName, name)
	}

	data, err := json.Marshal(workflow)
	if err != nil {
		return nil, fmt.Errorf("jsonnet config %s: failed to encode workflow %q: %w", configName, name, err)
	}
	return &File{Name: name, Data: data}, nil
}

func defaultWorkflowName(configName string) string {
	name := strings.TrimSuffix(filepath.Base(configName), fileExtension)
	return strings.TrimPrefix(name, ".")
}

type noImporter struct{}

func (*noImporter) Import(_, importedPath string) (jsonnet.Contents, string, error) {
	return jsonnet.Contents{}, "", fmt.Errorf("imports are not supported in jsonnet configs (import of %q)", importedPath)
}
