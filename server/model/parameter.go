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
	"fmt"
	"regexp"
	"slices"
	"strconv"
)

type ParameterType string

const (
	ParameterTypeString  ParameterType = "string"
	ParameterTypeNumber  ParameterType = "number"
	ParameterTypeBoolean ParameterType = "boolean"
	ParameterTypeChoice  ParameterType = "choice"
)

var parameterTypes = []ParameterType{
	ParameterTypeString,
	ParameterTypeNumber,
	ParameterTypeBoolean,
	ParameterTypeChoice,
}

const (
	// ParameterSourceRepoConfig marks parameters defined via the repo settings UI.
	// A future source "workflow" (definitions declared in the pipeline YAML) can be added
	// without changing the data model. Intended precedence once both exist: workflow-defined
	// definitions are authoritative for type/options, repo_config provides defaults/overrides
	// and acts as fallback (e.g. for repos using an external HTTP config service where no
	// static YAML exists).
	ParameterSourceRepoConfig = "repo_config"
)

var parameterNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Parameter is a typed input definition for manual pipeline runs. The chosen values are
// injected into the pipeline as environment variables via the pipeline's additional variables.
type Parameter struct {
	ID          int64         `json:"id"          xorm:"pk autoincr 'id'"`
	RepoID      int64         `json:"repo_id"     xorm:"UNIQUE(s) INDEX 'repo_id'"`
	Name        string        `json:"name"        xorm:"UNIQUE(s) INDEX 'name'"`
	Type        ParameterType `json:"type"        xorm:"'param_type'"`
	Description string        `json:"description" xorm:"TEXT 'description'"`
	Default     string        `json:"default"     xorm:"TEXT 'default_value'"`
	Options     []string      `json:"options"     xorm:"json 'options'"`
	Required    bool          `json:"required"    xorm:"required"`
	Order       int           `json:"order"       xorm:"display_order"`
	Source      string        `json:"source"      xorm:"source"`
} //	@name	Parameter

// TableName returns the database table name for xorm.
func (Parameter) TableName() string {
	return "parameters"
}

// Validate ensures the parameter definition is consistent.
func (p *Parameter) Validate() error {
	if !parameterNameRegex.MatchString(p.Name) {
		return fmt.Errorf("parameter name %q must be a valid environment variable identifier", p.Name)
	}

	if !slices.Contains(parameterTypes, p.Type) {
		return fmt.Errorf("invalid parameter type %q", p.Type)
	}

	// a boolean always has a value (checked or not), so "required" is meaningless
	if p.Type == ParameterTypeBoolean && p.Required {
		return fmt.Errorf("boolean parameter %q cannot be required", p.Name)
	}

	switch p.Type {
	case ParameterTypeChoice:
		if len(p.Options) == 0 {
			return fmt.Errorf("parameter %q of type %q requires options", p.Name, p.Type)
		}
		for _, option := range p.Options {
			if option == "" {
				return fmt.Errorf("parameter %q must not have empty options", p.Name)
			}
		}
	default:
		// non-choice types must not carry options
	}

	if p.Default != "" {
		if err := p.ValidateValue(p.Default); err != nil {
			return fmt.Errorf("invalid default: %w", err)
		}
	}

	return nil
}

// ValidateValue checks a submitted (or default) value against the parameter type.
func (p *Parameter) ValidateValue(value string) error {
	switch p.Type {
	case ParameterTypeNumber:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("value %q of parameter %q is not a number", value, p.Name)
		}
	case ParameterTypeBoolean:
		if value != "true" && value != "false" {
			return fmt.Errorf("value %q of parameter %q is not a boolean", value, p.Name)
		}
	case ParameterTypeChoice:
		if !slices.Contains(p.Options, value) {
			return fmt.Errorf("value %q of parameter %q is not one of the allowed options", value, p.Name)
		}
	case ParameterTypeString:
		// free-form value
	}
	return nil
}

// ValidateParameterValues validates submitted manual-run values against the given parameter
// definitions and fills in defaults for missing values. It stays lenient on purpose: values
// without a matching definition are left untouched so ad-hoc variables keep working.
// This is shared logic so it applies regardless of the parameter source.
func ValidateParameterValues(params []*Parameter, values map[string]string) error {
	for _, param := range params {
		value, ok := values[param.Name]
		if !ok || value == "" {
			if param.Default != "" {
				values[param.Name] = param.Default
				continue
			}
			if param.Required {
				return fmt.Errorf("required parameter %q is missing", param.Name)
			}
			continue
		}

		if err := param.ValidateValue(value); err != nil {
			return err
		}
	}
	return nil
}

type ParameterPatch struct {
	Name        *string        `json:"name"`
	Type        *ParameterType `json:"type"`
	Description *string        `json:"description"`
	Default     *string        `json:"default"`
	Options     []string       `json:"options"`
	Required    *bool          `json:"required"`
	Order       *int           `json:"order"`
} //	@name	ParameterPatch
