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

// Different ways to handle failure states.
const (
	FailureIgnore = "ignore"
	FailureFail   = "fail"
	FailureCancel = "cancel"
)

// Step represents a process in the pipeline.
type Step struct {
	ID         int64       `json:"id"                   xorm:"pk autoincr 'id'"`
	UUID       string      `json:"uuid"                 xorm:"INDEX 'uuid'"`
	PipelineID int64       `json:"pipeline_id"          xorm:"UNIQUE(s) INDEX 'pipeline_id'"`
	PID        int         `json:"pid"                  xorm:"UNIQUE(s) 'pid'"`
	PPID       int         `json:"ppid"                 xorm:"ppid"`
	Name       string      `json:"name"                 xorm:"name"`
	State      StatusValue `json:"state"                xorm:"state"`
	Error      string      `json:"error,omitempty"      xorm:"TEXT 'error'"`
	Failure    string      `json:"-"                    xorm:"failure"`
	ExitCode   int         `json:"exit_code"            xorm:"exit_code"`
	Started    int64       `json:"started,omitempty"    xorm:"started"`
	Finished   int64       `json:"finished,omitempty"   xorm:"finished"`
	Type       StepType    `json:"type,omitempty"       xorm:"type"`
	DependsOn  []string    `json:"depends_on,omitempty" xorm:"json 'depends_on'"`
	LogState   LogState    `json:"-"                    xorm:"VARCHAR(16) NOT NULL DEFAULT '' INDEX 'log_state'"`
	LogClaimed int64       `json:"-"                    xorm:"NOT NULL DEFAULT 0 'log_claimed'"`
} //	@name	Step

// LogState tells who works on the log of a step. It only changes with store.StepLogStateSwap,
// which claims the log atomically, so several servers can share the database.
type LogState string

// The zero value is open, so new steps and steps from before the upgrade start open.
const (
	LogStateOpen       LogState = ""           // log lines may still come
	LogStateFinalizing LogState = "finalizing" // claimed by a log finalizer
	LogStateFinalized  LogState = "finalized"  // complete, may be packed or moved
	LogStateDeleting   LogState = "deleting"   // claimed by a log delete
)

// TableName return database table name for xorm.
func (Step) TableName() string {
	return "steps"
}

// Running returns true if the process state is pending or running.
func (p *Step) Running() bool {
	return p.State == StatusPending || p.State == StatusRunning
}

// Failing returns true if the process state is failed, killed or error.
func (p *Step) Failing() bool {
	return p.State == StatusError || p.State == StatusKilled || p.State == StatusFailure
}

// StepType identifies the type of step.
type StepType string //	@name	StepType

const (
	StepTypeClone    StepType = "clone"
	StepTypeService  StepType = "service"
	StepTypePlugin   StepType = "plugin"
	StepTypeCommands StepType = "commands"
	StepTypeCache    StepType = "cache"
)
