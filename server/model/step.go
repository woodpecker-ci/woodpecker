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

import "fmt"

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

// Step represents a process in the pipeline.
type Step struct {
	ID         int64             `json:"id"                   xorm:"pk autoincr 'step_id'"`
	PipelineID int64             `json:"pipeline_id"          xorm:"UNIQUE(s) INDEX 'step_pipeline_id'"`
	PID        int               `json:"pid"                  xorm:"UNIQUE(s) 'step_pid'"`
	PPID       int               `json:"ppid"                 xorm:"step_ppid"`
	PGID       int               `json:"pgid"                 xorm:"step_pgid"`
	Name       string            `json:"name"                 xorm:"step_name"`
	State      StatusValue       `json:"state"                xorm:"step_state"`
	Error      string            `json:"error,omitempty"      xorm:"VARCHAR(500) step_error"`
	ExitCode   int               `json:"exit_code"            xorm:"step_exit_code"`
	Started    int64             `json:"start_time,omitempty" xorm:"step_started"`
	Stopped    int64             `json:"end_time,omitempty"   xorm:"step_stopped"`
	AgentID    int64             `json:"agent_id,omitempty"   xorm:"step_agent_id"`
	Platform   string            `json:"platform,omitempty"   xorm:"step_platform"`
	Environ    map[string]string `json:"environ,omitempty"    xorm:"json 'step_environ'"`
	Children   []*Step           `json:"children,omitempty"   xorm:"-"`
} //	@name	Step

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
	return p.State == StatusError || p.State == StatusKilled || p.State == StatusFailure
}

// IsParent returns true if the process is a parent process.
func (p *Step) IsParent() bool {
	return p.PPID == 0
}

// IsMultiPipeline checks if step list contain more than one parent step
func IsMultiPipeline(steps []*Step) bool {
	c := 0
	for _, step := range steps {
		if step.IsParent() {
			c++
		}
		if c > 1 {
			return true
		}
	}
	return false
}

// Tree creates a process tree from a flat process list.
func Tree(steps []*Step) ([]*Step, error) {
	var nodes []*Step

	// init parent nodes
	for i := range steps {
		if steps[i].IsParent() {
			nodes = append(nodes, steps[i])
		}
	}

	// assign children to parents
	for i := range steps {
		if !steps[i].IsParent() {
			parent, err := findNode(nodes, steps[i].PPID)
			if err != nil {
				return nil, err
			}
			parent.Children = append(parent.Children, steps[i])
		}
	}

	return nodes, nil
}

// PipelineStatus determine pipeline status based on corresponding step list
func PipelineStatus(steps []*Step) StatusValue {
	status := StatusSuccess

	for _, p := range steps {
		if p.IsParent() && p.Failing() {
			status = p.State
		}
	}

	return status
}

// IsThereRunningStage determine if it contains steps running or pending to run
// TODO: return false based on depends_on (https://github.com/woodpecker-ci/woodpecker/pull/730#discussion_r795681697)
func IsThereRunningStage(steps []*Step) bool {
	for _, p := range steps {
		if p.IsParent() {
			if p.Running() {
				return true
			}
		}
	}
	return false
}

func findNode(nodes []*Step, pid int) (*Step, error) {
	for _, node := range nodes {
		if node.PID == pid {
			return node, nil
		}
	}

	return nil, fmt.Errorf("Corrupt step structure")
}
