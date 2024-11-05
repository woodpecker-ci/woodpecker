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
	ID         int64             `json:"id"                   xorm:"pk autoincr 'id'"`
	PipelineID int64             `json:"pipeline_id"          xorm:"UNIQUE(s) INDEX 'pipeline_id'"`
	PID        int               `json:"pid"                  xorm:"UNIQUE(s) 'pid'"`
	Name       string            `json:"name"                 xorm:"name"`
	State      StatusValue       `json:"state"                xorm:"state"`
	Error      string            `json:"error,omitempty"      xorm:"TEXT 'error'"`
	Started    int64             `json:"started,omitempty"    xorm:"started"`
	Finished   int64             `json:"finished,omitempty"   xorm:"finished"`
	AgentID    int64             `json:"agent_id,omitempty"   xorm:"agent_id"`
	Platform   string            `json:"platform,omitempty"   xorm:"platform"`
	Environ    map[string]string `json:"environ,omitempty"    xorm:"json 'environ'"`
	AxisID     int               `json:"-"                    xorm:"axis_id"`
	Children   []*Step           `json:"children,omitempty"   xorm:"-"`
}

// TableName return database table name for xorm.
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

// IsThereRunningStage determine if it contains workflows running or pending to run.
// TODO: return false based on depends_on (https://github.com/woodpecker-ci/woodpecker/pull/730#discussion_r795681697)
func IsThereRunningStage(workflows []*Workflow) bool {
	for _, p := range workflows {
		if p.Running() {
			return true
		}
	}
	return false
}

// PipelineStatus determine pipeline status based on corresponding workflow list.
func PipelineStatus(workflows []*Workflow) StatusValue {
	status := StatusSuccess

	for _, p := range workflows {
		if p.Failing() {
			status = p.State
		}
	}

	return status
}

// WorkflowStatus determine workflow status based on corresponding step list.
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
