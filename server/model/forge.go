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
	ForgeTypeBitbucket           ForgeType = "bitbucket"
	ForgeTypeBitbucketDatacenter ForgeType = "bitbucket-dc"
	ForgeTypeAddon               ForgeType = "addon"
)

type Forge struct {
	ID                int64          `xorm:"pk autoincr 'id'"`
	Type              ForgeType      `xorm:"VARCHAR(250)"`
	URL               string         `xorm:"VARCHAR(500) 'url'"`
	Client            string         `xorm:"VARCHAR(250)"`
	ClientSecret      string         `xorm:"VARCHAR(250)"`
	SkipVerify        bool           `xorm:"bool"`
	OAuthHost         string         `xorm:"VARCHAR(250) 'oauth_host'"` // public url for oauth if different from url
	AdditionalOptions map[string]any `xorm:"json"`
}
