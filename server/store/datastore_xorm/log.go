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
	"bytes"
	"io"
	"io/ioutil"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) LogFind(proc *model.Proc) (io.ReadCloser, error) {
	logs := &model.Logs{
		ProcID: proc.ID,
	}
	if err := wrapGet(s.engine.Get(logs)); err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(logs.Data)
	return ioutil.NopCloser(buf), nil
}

func (s storage) LogSave(proc *model.Proc, reader io.Reader) error {
	data, _ := ioutil.ReadAll(reader)

	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	logs := new(model.Logs)
	exist, err := sess.Where("log_job_id = ?", proc.ID).Get(logs)
	if err != nil {
		return err
	}

	if exist {
		if _, err := sess.ID(logs.ID).Cols("log_data").Update(&model.Logs{Data: data}); err != nil {
			return err
		}
	} else {
		if _, err := sess.InsertOne(&model.Logs{
			ProcID: proc.ID,
			Data:   data,
		}); err != nil {
			return err
		}
	}

	return sess.Commit()
}
