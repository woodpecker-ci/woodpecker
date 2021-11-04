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
	ProcFind(*Build, int) (*Proc, error)
	ProcChild(*Build, int, string) (*Proc, error)
	ProcList(*Build) ([]*Proc, error)
	ProcCreate([]*Proc) error
	ProcUpdate(*Proc) error
	ProcClear(*Build) error
}

// Proc represents a process in the build pipeline.
// swagger:model proc
type Proc struct {
	ID       int64             `json:"id"                   xorm:"pk autoincr 'proc_id'"`
	BuildID  int64             `json:"build_id"             xorm:"UNIQUE(s) INDEX 'proc_build_id'"`
	PID      int               `json:"pid"                  xorm:"UNIQUE(s) 'proc_pid'"`
	PPID     int               `json:"ppid"                 xorm:"proc_ppid"`
	PGID     int               `json:"pgid"                 xorm:"proc_pgid"`
	Name     string            `json:"name"                 xorm:"proc_name"`
	State    string            `json:"state"                xorm:"proc_state"`
	Error    string            `json:"error,omitempty"      xorm:"proc_error"`
	ExitCode int               `json:"exit_code"            xorm:"proc_exit_code"`
	Started  int64             `json:"start_time,omitempty" xorm:"proc_started"`
	Stopped  int64             `json:"end_time,omitempty"   xorm:"proc_stopped"`
	Machine  string            `json:"machine,omitempty"    xorm:"proc_machine"`
	Platform string            `json:"platform,omitempty"   xorm:"proc_platform"`
	Environ  map[string]string `json:"environ,omitempty"    xorm:"json 'proc_environ'"`
	Children []*Proc           `json:"children,omitempty"   xorm:"-"`
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

// Tree creates a process tree from a flat process list.
func Tree(procs []*Proc) []*Proc {
	var nodes []*Proc
	for _, proc := range procs {
		if proc.PPID == 0 {
			nodes = append(nodes, proc)
		} else {
			parent, _ := findNode(nodes, proc.PPID)
			parent.Children = append(parent.Children, proc)
		}
	}
	return nodes
}

func findNode(nodes []*Proc, pid int) (*Proc, error) {
	for _, node := range nodes {
		if node.PID == pid {
			return node, nil
		}
	}

	return nil, fmt.Errorf("Corrupt proc structure")
}
