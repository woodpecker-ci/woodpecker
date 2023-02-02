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

package datastore

import (
	"bytes"
	"io"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) FileList(pipeline *model.Pipeline) ([]*model.File, error) {
	files := make([]*model.File, 0, perPage)
	return files, s.engine.Where("file_pipeline_id = ?", pipeline.ID).Find(&files)
}

func (s storage) FileFind(step *model.Step, name string) (*model.File, error) {
	file := &model.File{
		StepID: step.ID,
		Name:   name,
	}
	return file, wrapGet(s.engine.Get(file))
}

func (s storage) FileRead(step *model.Step, name string) (io.ReadCloser, error) {
	file, err := s.FileFind(step, name)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(file.Data)
	return io.NopCloser(buf), err
}

func (s storage) FileCreate(file *model.File, reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	file.Data = data
	// only Insert set auto created ID back to object
	_, err = s.engine.Insert(file)
	return err
}
