// Copyright 2024 Woodpecker Authors
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
	"errors"
	"fmt"
)

var (
	ErrVariableNameInvalid  = errors.New("invalid variable name")
	ErrVariableValueInvalid = errors.New("invalid variable value")
)

// VariableStore persists variable information to storage.
type VariableStore interface {
	VariableFind(*Repo, string) (*Variable, error)
	VariableList(*Repo, bool, *ListOptions) ([]*Variable, error)
	VariableCreate(*Variable) error
	VariableUpdate(*Variable) error
	VariableDelete(*Variable) error
	OrgVariableFind(int64, string) (*Variable, error)
	OrgVariableList(int64, *ListOptions) ([]*Variable, error)
	GlobalVariableFind(string) (*Variable, error)
	GlobalVariableList(*ListOptions) ([]*Variable, error)
	VariableListAll() ([]*Variable, error)
}

type Variable struct {
	ID     int64  `json:"id"              xorm:"pk autoincr 'id'"`
	OrgID  int64  `json:"org_id"          xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'org_id'"`
	RepoID int64  `json:"repo_id"         xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'repo_id'"`
	Name   string `json:"name"            xorm:"NOT NULL UNIQUE(s) INDEX 'name'"`
	Value  string `json:"value,omitempty" xorm:"TEXT 'value'"`
} //	@name Variable

// TableName return database table name for xorm.
func (Variable) TableName() string {
	return "variables"
}

// Global variable.
func (s Variable) IsGlobal() bool {
	return s.RepoID == 0 && s.OrgID == 0
}

// Organization variable.
func (s Variable) IsOrganization() bool {
	return s.RepoID == 0 && s.OrgID != 0
}

// Repository variable.
func (s Variable) IsRepository() bool {
	return s.RepoID != 0 && s.OrgID == 0
}

// Validate validates the required fields and formats.
func (s *Variable) Validate() error {
	switch {
	case len(s.Name) == 0:
		return fmt.Errorf("%w: empty name", ErrVariableNameInvalid)
	case len(s.Value) == 0:
		return fmt.Errorf("%w: empty value", ErrVariableValueInvalid)
	default:
		return nil
	}
}
