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

package model

// TaskStore defines storage for scheduled Tasks.
type TaskStore interface {
	TaskList() ([]*Task, error)
	TaskInsert(*Task) error
	TaskDelete(string) error
}

// Task defines scheduled pipeline Task.
type Task struct {
	ID           string            `xorm:"PK UNIQUE 'task_id'"`
	Data         []byte            `xorm:"'task_data'"`
	Labels       map[string]string `xorm:"json 'task_labels'"`
	Dependencies []string          `xorm:"json 'task_dependencies'"`
	RunOn        []string          `xorm:"json 'task_run_on'"`
}

// TableName return database table name for xorm
func (Task) TableName() string {
	return "tasks"
}
