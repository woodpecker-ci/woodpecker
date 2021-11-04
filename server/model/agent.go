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

// TODO: check if it is actually used or just some relict from the past

type Agent struct {
	ID       int64  `xorm:"pk autoincr 'agent_id'"`
	Addr     string `xorm:"UNIQUE VARCHAR(250) 'agent_addr'"`
	Platform string `xorm:"VARCHAR(500) 'agent_platform'"`
	Capacity int64  `xorm:"agent_capacity"`
	Created  int64  `xorm:"created 'agent_created'"`
	Updated  int64  `xorm:"updated 'agent_updated'"`
}
