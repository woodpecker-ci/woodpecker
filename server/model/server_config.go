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

// ServerConfigStore persists key-value pairs for storing server configurations.
type ServerConfigStore interface {
	ServerConfigGet(key string) (string, error)
	ServerConfigSet(key int64, value string) error
}

// ServerConfig represents a key-value pair for storing server configurations.
type ServerConfig struct {
	Key   string `json:"key"   xorm:"pk"`
	Value string `json:"value" xorm:""`
}
