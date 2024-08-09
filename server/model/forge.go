// Copyright 2024 Woodpecker Authors
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

type ForgeType string

const (
	ForgeTypeGithub              ForgeType = "github"
	ForgeTypeGitlab              ForgeType = "gitlab"
	ForgeTypeGitea               ForgeType = "gitea"
	ForgeTypeForgejo             ForgeType = "forgejo"
	ForgeTypeBitbucket           ForgeType = "bitbucket"
	ForgeTypeBitbucketDatacenter ForgeType = "bitbucket-dc"
	ForgeTypeAddon               ForgeType = "addon"
)

type Forge struct {
	ID                int64          `json:"id"                           xorm:"pk autoincr 'id'"`
	Type              ForgeType      `json:"type"                         xorm:"VARCHAR(250)"`
	URL               string         `json:"url"                          xorm:"VARCHAR(500) 'url'"`
	Client            string         `json:"client,omitempty"             xorm:"VARCHAR(250)"`
	ClientSecret      string         `json:"-"                            xorm:"VARCHAR(250)"` // do not expose client secret
	SkipVerify        bool           `json:"skip_verify,omitempty"        xorm:"bool"`
	OAuthHost         string         `json:"oauth_host,omitempty"         xorm:"VARCHAR(250) 'oauth_host'"` // public url for oauth if different from url
	AdditionalOptions map[string]any `json:"additional_options,omitempty" xorm:"json"`
} //	@name Forge

// TableName returns the database table name for xorm.
func (Forge) TableName() string {
	return "forges"
}

// PublicCopy returns a copy of the forge without sensitive information and technical details.
func (f *Forge) PublicCopy() *Forge {
	forge := &Forge{
		ID:   f.ID,
		Type: f.Type,
		URL:  f.URL,
	}

	return forge
}
