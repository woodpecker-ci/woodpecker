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

import "io"

// FileStore persists pipeline artifacts to storage.
type FileStore interface {
	FileList(*Pipeline, *ListOptions) ([]*File, error)
	FileFind(*Step, string) (*File, error)
	FileRead(*Step, string) (io.ReadCloser, error)
	FileCreate(*File, io.Reader) error
}

// File represents a pipeline artifact.
type File struct {
	ID         int64  `json:"id"      xorm:"pk autoincr 'file_id'"`
	PipelineID int64  `json:"-"       xorm:"INDEX 'file_pipeline_id'"`
	StepID     int64  `json:"step_id" xorm:"UNIQUE(s) INDEX 'file_step_id'"`
	PID        int    `json:"pid"     xorm:"file_pid"`
	Name       string `json:"name"    xorm:"UNIQUE(s) file_name"`
	Size       int    `json:"size"    xorm:"file_size"`
	Mime       string `json:"mime"    xorm:"file_mime"`
	Time       int64  `json:"time"    xorm:"file_time"`
	Passed     int    `json:"passed"  xorm:"file_meta_passed"`
	Failed     int    `json:"failed"  xorm:"file_meta_failed"`
	Skipped    int    `json:"skipped" xorm:"file_meta_skipped"`
	Data       []byte `json:"-"       xorm:"file_data"` // TODO: don't store in db but object storage?
}

// TableName return database table name for xorm
func (File) TableName() string {
	return "files"
}
