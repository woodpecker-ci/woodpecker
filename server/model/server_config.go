// Copyright 2022 Woodpecker Authors
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

// ServerConfig represents a key-value pair for storing server configurations.
type ServerConfig struct {
	Key   string `json:"key"   xorm:"pk 'key'"`
	Value string `json:"value" xorm:"value"`
}

// TableName return database table name for xorm.
func (ServerConfig) TableName() string {
	return "server_configs"
}
