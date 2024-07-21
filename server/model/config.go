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

// Config represents a pipeline configuration.
type Config struct {
	ID     int64  `json:"-"    xorm:"pk autoincr 'id'"`
	RepoID int64  `json:"-"    xorm:"UNIQUE(s) 'repo_id'"`
	Hash   string `json:"hash" xorm:"UNIQUE(s) 'hash'"`
	Name   string `json:"name" xorm:"UNIQUE(s) 'name'"`
	Data   []byte `json:"data" xorm:"LONGBLOB 'data'"`
} //	@name Config

func (Config) TableName() string {
	return "configs"
}

// PipelineConfig is the n:n relation between Pipeline and Config.
type PipelineConfig struct {
	ConfigID   int64 `json:"-"   xorm:"UNIQUE(s) NOT NULL 'config_id'"`
	PipelineID int64 `json:"-"   xorm:"UNIQUE(s) NOT NULL 'pipeline_id'"`
}

func (PipelineConfig) TableName() string {
	return "pipeline_configs"
}
