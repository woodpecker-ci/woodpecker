// Copyright 2021 Woodpecker Authors
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

package datastore_xorm

import (
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) ProcLoad(id int64) (*model.Proc, error) {
	proc := new(model.Proc)
	return proc, wrapGet(s.engine.ID(id).Get(proc))
}

func (s storage) ProcFind(build *model.Build, pid int) (*model.Proc, error) {
	proc := &model.Proc{
		BuildID: build.ID,
		PID:     pid,
	}
	return proc, wrapGet(s.engine.Get(proc))
}

func (s storage) ProcChild(build *model.Build, ppid int, child string) (*model.Proc, error) {
	proc := &model.Proc{
		BuildID: build.ID,
		PPID:    ppid,
		Name:    child,
	}
	return proc, wrapGet(s.engine.Get(proc))
}

func (s storage) ProcList(build *model.Build) ([]*model.Proc, error) {
	procList := make([]*model.Proc, 0, perPage)
	return procList, s.engine.Where("proc_build_id = ?", build.ID).Find(&procList)
}

func (s storage) ProcCreate(procs []*model.Proc) error {
	_, err := s.engine.Insert(procs)
	return err
}

func (s storage) ProcUpdate(proc *model.Proc) error {
	_, err := s.engine.ID(proc.ID).Update(proc)
	return err
}

func (s storage) ProcClear(build *model.Build) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Where("file_build_id = ?", build.ID).Delete(new(model.File)); err != nil {
		return err
	}

	if _, err := sess.Where("proc_build_id = ?", build.ID).Delete(new(model.Proc)); err != nil {
		return err
	}

	return sess.Commit()
}
