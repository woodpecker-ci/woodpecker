// Copyright 2021 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

// StepStore persists process information to storage.
type StepStore interface {
	StepLoad(int64) (*Step, error)
	StepFind(*Pipeline, int) (*Step, error)
	StepChild(*Pipeline, int, string) (*Step, error)
	StepList(*Pipeline) ([]*Step, error)
	StepCreate([]*Step) error
	StepUpdate(*Step) error
	StepClear(*Pipeline) error
}

// Different ways to handle failure states
const (
	FailureIgnore = "ignore"
	FailureFail   = "fail"
	// FailureCancel = "cancel" // Not implemented yet
)

// Step represents a process in the pipeline.
type Step struct {
	ID         int64       `json:"id"                   xorm:"pk autoincr 'step_id'"`
	UUID       string      `json:"uuid"                 xorm:"INDEX 'step_uuid'"`
	PipelineID int64       `json:"pipeline_id"          xorm:"UNIQUE(s) INDEX 'step_pipeline_id'"`
	PID        int         `json:"pid"                  xorm:"UNIQUE(s) 'step_pid'"`
	PPID       int         `json:"ppid"                 xorm:"step_ppid"`
	Name       string      `json:"name"                 xorm:"step_name"`
	State      StatusValue `json:"state"                xorm:"step_state"`
	Error      string      `json:"error,omitempty"      xorm:"TEXT 'step_error'"`
	Failure    string      `json:"-"                    xorm:"step_failure"`
	ExitCode   int         `json:"exit_code"            xorm:"step_exit_code"`
	Started    int64       `json:"start_time,omitempty" xorm:"step_started"`
	Stopped    int64       `json:"end_time,omitempty"   xorm:"step_stopped"`
	Type       StepType    `json:"type,omitempty"       xorm:"step_type"`
} //	@name Step

type UpdateStepStore interface {
	StepUpdate(*Step) error
}

// TableName return database table name for xorm
func (Step) TableName() string {
	return "steps"
}

// Running returns true if the process state is pending or running.
func (p *Step) Running() bool {
	return p.State == StatusPending || p.State == StatusRunning
}

// Failing returns true if the process state is failed, killed or error.
func (p *Step) Failing() bool {
	return p.Failure == FailureFail && (p.State == StatusError || p.State == StatusKilled || p.State == StatusFailure)
}

// StepType identifies the type of step
type StepType string //	@name StepType

const (
	StepTypeClone    StepType = "clone"
	StepTypeService  StepType = "service"
	StepTypePlugin   StepType = "plugin"
	StepTypeCommands StepType = "commands"
	StepTypeCache    StepType = "cache"
)
