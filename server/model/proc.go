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

// ProcStore persists process information to storage.
type ProcStore interface {
	ProcLoad(int64) (*Proc, error)
	ProcFind(*Pipeline, int) (*Proc, error)
	ProcChild(*Pipeline, int, string) (*Proc, error)
	ProcList(*Pipeline) ([]*Proc, error)
	ProcCreate([]*Proc) error
	ProcUpdate(*Proc) error
	ProcClear(*Pipeline) error
}

// Proc represents a process in the build pipeline.
// swagger:model proc
type Proc struct {
	ID         int64             `json:"id"                   xorm:"pk autoincr 'proc_id'"`
	PipelineID int64             `json:"build_id"             xorm:"UNIQUE(s) INDEX 'proc_build_id'"`
	PID        int               `json:"pid"                  xorm:"UNIQUE(s) 'proc_pid'"`
	PPID       int               `json:"ppid"                 xorm:"proc_ppid"`
	PGID       int               `json:"pgid"                 xorm:"proc_pgid"`
	Name       string            `json:"name"                 xorm:"proc_name"`
	State      StatusValue       `json:"state"                xorm:"proc_state"`
	Error      string            `json:"error,omitempty"      xorm:"VARCHAR(500) proc_error"`
	ExitCode   int               `json:"exit_code"            xorm:"proc_exit_code"`
	Started    int64             `json:"start_time,omitempty" xorm:"proc_started"`
	Stopped    int64             `json:"end_time,omitempty"   xorm:"proc_stopped"`
	Machine    string            `json:"machine,omitempty"    xorm:"proc_machine"`
	Platform   string            `json:"platform,omitempty"   xorm:"proc_platform"`
	Environ    map[string]string `json:"environ,omitempty"    xorm:"json 'proc_environ'"`
	Children   []*Proc           `json:"children,omitempty"   xorm:"-"`
}

type UpdateProcStore interface {
	ProcUpdate(*Proc) error
}

// TableName return database table name for xorm
func (Proc) TableName() string {
	return "procs"
}

// Running returns true if the process state is pending or running.
func (p *Proc) Running() bool {
	return p.State == StatusPending || p.State == StatusRunning
}

// Failing returns true if the process state is failed, killed or error.
func (p *Proc) Failing() bool {
	return p.State == StatusError || p.State == StatusKilled || p.State == StatusFailure
}

// IsParent returns true if the process is a parent process.
func (p *Proc) IsParent() bool {
	return p.PPID == 0
}

// IsMultiPipeline checks if proc list contain more than one parent proc
func IsMultiPipeline(procs []*Proc) bool {
	c := 0
	for _, proc := range procs {
		if proc.IsParent() {
			c++
		}
		if c > 1 {
			return true
		}
	}
	return false
}

// Tree creates a process tree from a flat process list.
func Tree(procs []*Proc) ([]*Proc, error) {
	var nodes []*Proc

	// init parent nodes
	for i := range procs {
		if procs[i].IsParent() {
			nodes = append(nodes, procs[i])
		}
	}

	// assign children to parrents
	for i := range procs {
		if !procs[i].IsParent() {
			parent, err := findNode(nodes, procs[i].PPID)
			if err != nil {
				return nil, err
			}
			parent.Children = append(parent.Children, procs[i])
		}
	}

	return nodes, nil
}

// BuildStatus determine build status based on corresponding proc list
func BuildStatus(procs []*Proc) StatusValue {
	status := StatusSuccess

	for _, p := range procs {
		if p.IsParent() && p.Failing() {
			status = p.State
		}
	}

	return status
}

// IsThereRunningStage determine if it contains procs running or pending to run
// TODO: return false based on depends_on (https://github.com/woodpecker-ci/woodpecker/pull/730#discussion_r795681697)
func IsThereRunningStage(procs []*Proc) bool {
	for _, p := range procs {
		if p.IsParent() {
			if p.Running() {
				return true
			}
		}
	}
	return false
}

func findNode(nodes []*Proc, pid int) (*Proc, error) {
	for _, node := range nodes {
		if node.PID == pid {
			return node, nil
		}
	}

	return nil, fmt.Errorf("Corrupt proc structure")
}
