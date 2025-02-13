package model

import (
	"errors"
	"fmt"
)

var (
	ErrParameterNameInvalid = errors.New("invalid parameter name")
	ErrParameterTypeInvalid = errors.New("invalid parameter type")
)

type ParameterType string

const (
	ParameterTypeBoolean        ParameterType = "boolean"
	ParameterTypeSingleChoice   ParameterType = "single_choice"
	ParameterTypeMultipleChoice ParameterType = "multiple_choice"
	ParameterTypeString         ParameterType = "string"
	ParameterTypeText           ParameterType = "text"
	ParameterTypePassword       ParameterType = "password"
)

// Parameter represents a configurable parameter for a repository.
type Parameter struct {
	ID           int64         `json:"id"            xorm:"pk autoincr 'parameter_id'"`
	RepoID       int64         `json:"repo_id"       xorm:"UNIQUE(s) 'parameter_repo_id'"`
	Name         string        `json:"name"          xorm:"UNIQUE(s) 'parameter_name'"`
	Branch       string        `json:"branch"        xorm:"UNIQUE(s) 'parameter_branch'"`
	Type         ParameterType `json:"type"          xorm:"'parameter_type'"`
	Description  string        `json:"description"   xorm:"TEXT 'parameter_description'"`
	DefaultValue string        `json:"default_value" xorm:"TEXT 'parameter_default_value'"`
	TrimString   bool          `json:"trim_string"   xorm:"'parameter_trim_string'"`
}

// TableName return database table name for xorm.
func (Parameter) TableName() string {
	return "parameters"
}

// Validate validates the required fields and formats.
func (p *Parameter) Validate() error {
	switch {
	case len(p.Name) == 0:
		return fmt.Errorf("%w: empty name", ErrParameterNameInvalid)
	case len(p.Branch) == 0:
		return fmt.Errorf("%w: empty branch", ErrParameterNameInvalid)
	case !validParameterType(p.Type):
		return fmt.Errorf("%w: %s", ErrParameterTypeInvalid, p.Type)
	default:
		return nil
	}
}

func validParameterType(t ParameterType) bool {
	switch t {
	case ParameterTypeBoolean, ParameterTypeSingleChoice, ParameterTypeMultipleChoice,
		ParameterTypeString, ParameterTypeText, ParameterTypePassword:
		return true
	default:
		return false
	}
}
