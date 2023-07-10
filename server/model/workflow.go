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

// Workflow represents a workflow in the pipeline.
type Workflow struct {
	ID         int64             `json:"id"                   xorm:"pk autoincr 'workflow_id'"`
	PipelineID int64             `json:"pipeline_id"          xorm:"UNIQUE(s) INDEX 'workflow_pipeline_id'"`
	PID        int               `json:"pid"                  xorm:"UNIQUE(s) 'workflow_pid'"`
	Name       string            `json:"name"                 xorm:"workflow_name"`
	State      StatusValue       `json:"state"                xorm:"workflow_state"`
	Error      string            `json:"error,omitempty"      xorm:"TEXT 'workflow_error'"`
	Started    int64             `json:"start_time,omitempty" xorm:"workflow_started"`
	Stopped    int64             `json:"end_time,omitempty"   xorm:"workflow_stopped"`
	AgentID    int64             `json:"agent_id,omitempty"   xorm:"workflow_agent_id"`
	Platform   string            `json:"platform,omitempty"   xorm:"workflow_platform"`
	Environ    map[string]string `json:"environ,omitempty"    xorm:"json 'workflow_environ'"`
	Children   []*Step           `json:"children,omitempty"   xorm:"-"`
}

type UpdateWorkflowStore interface {
	WorkflowUpdate(*Workflow) error
}

// TableName return database table name for xorm
func (Workflow) TableName() string {
	return "workflows"
}

// Running returns true if the process state is pending or running.
func (p *Workflow) Running() bool {
	return p.State == StatusPending || p.State == StatusRunning
}

// Failing returns true if the process state is failed, killed or error.
func (p *Workflow) Failing() bool {
	return p.State == StatusError || p.State == StatusKilled || p.State == StatusFailure
}

// IsThereRunningStage determine if it contains workflows running or pending to run
// TODO: return false based on depends_on (https://github.com/woodpecker-ci/woodpecker/pull/730#discussion_r795681697)
func IsThereRunningStage(workflows []*Workflow) bool {
	for _, p := range workflows {
		if p.Running() {
			return true
		}
	}
	return false
}

// PipelineStatus determine pipeline status based on corresponding workflow list
func PipelineStatus(workflows []*Workflow) StatusValue {
	status := StatusSuccess

	for _, p := range workflows {
		if p.Failing() {
			status = p.State
		}
	}

	return status
}

// WorkflowStatus determine workflow status based on corresponding step list
func WorkflowStatus(steps []*Step) StatusValue {
	status := StatusSuccess

	for _, p := range steps {
		if p.Failing() {
			status = p.State
			break
		}
	}

	return status
}
