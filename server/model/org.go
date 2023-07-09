// Copyright 2023 Woodpecker Authors
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

// Org represents an organization.
type Org struct {
	ID   int64   `json:"id,omitempty"                    xorm:"pk autoincr"`
	Name string  `json:"name"`
	Type OrgType `json:"type"                            xorm:"index"`
} //	@name Org

// OrgType represents the type of an organization.
type OrgType string

const (
	// OrgTypeUser represents a user organization.
	OrgTypeUser OrgType = "user"
	// OrgTypeTeam represents a team organization.
	OrgTypeTeam OrgType = "team"
)
