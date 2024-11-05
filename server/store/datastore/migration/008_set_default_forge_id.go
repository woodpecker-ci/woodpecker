// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migration

import (
	"fmt"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type userV008 struct {
	ID            int64               `xorm:"pk autoincr 'user_id'"`
	ForgeID       int64               `xorm:"forge_id"`
	ForgeRemoteID model.ForgeRemoteID `xorm:"forge_remote_id"`
	Login         string              `xorm:"UNIQUE 'user_login'"`
	Token         string              `xorm:"TEXT 'user_token'"`
	Secret        string              `xorm:"TEXT 'user_secret'"`
	Expiry        int64               `xorm:"user_expiry"`
	Email         string              `xorm:" varchar(500) 'user_email'"`
	Avatar        string              `xorm:" varchar(500) 'user_avatar'"`
	Admin         bool                `xorm:"user_admin"`
	Hash          string              `xorm:"UNIQUE varchar(500) 'user_hash'"`
	OrgID         int64               `xorm:"user_org_id"`
}

func (userV008) TableName() string {
	return "users"
}

type repoV008 struct {
	ID                           int64                `xorm:"pk autoincr 'repo_id'"`
	UserID                       int64                `xorm:"repo_user_id"`
	ForgeID                      int64                `xorm:"forge_id"`
	ForgeRemoteID                model.ForgeRemoteID  `xorm:"forge_remote_id"`
	OrgID                        int64                `xorm:"repo_org_id"`
	Owner                        string               `xorm:"UNIQUE(name) 'repo_owner'"`
	Name                         string               `xorm:"UNIQUE(name) 'repo_name'"`
	FullName                     string               `xorm:"UNIQUE 'repo_full_name'"`
	Avatar                       string               `xorm:"varchar(500) 'repo_avatar'"`
	ForgeURL                     string               `xorm:"varchar(1000) 'repo_forge_url'"`
	Clone                        string               `xorm:"varchar(1000) 'repo_clone'"`
	CloneSSH                     string               `xorm:"varchar(1000) 'repo_clone_ssh'"`
	Branch                       string               `xorm:"varchar(500) 'repo_branch'"`
	SCMKind                      model.SCMKind        `xorm:"varchar(50) 'repo_scm'"`
	PREnabled                    bool                 `xorm:"DEFAULT TRUE 'repo_pr_enabled'"`
	Timeout                      int64                `xorm:"repo_timeout"`
	Visibility                   model.RepoVisibility `xorm:"varchar(10) 'repo_visibility'"`
	IsSCMPrivate                 bool                 `xorm:"repo_private"`
	IsTrusted                    bool                 `xorm:"repo_trusted"`
	IsGated                      bool                 `xorm:"repo_gated"`
	IsActive                     bool                 `xorm:"repo_active"`
	AllowPull                    bool                 `xorm:"repo_allow_pr"`
	AllowDeploy                  bool                 `xorm:"repo_allow_deploy"`
	Config                       string               `xorm:"varchar(500) 'repo_config_path'"`
	Hash                         string               `xorm:"varchar(500) 'repo_hash'"`
	Perm                         *model.Perm          `xorm:"-"`
	CancelPreviousPipelineEvents []model.WebhookEvent `xorm:"json 'cancel_previous_pipeline_events'"`
	NetrcOnlyTrusted             bool                 `xorm:"NOT NULL DEFAULT true 'netrc_only_trusted'"`
}

func (repoV008) TableName() string {
	return "repos"
}

type forgeV008 struct {
	ID                int64           `xorm:"pk autoincr 'id'"`
	Type              model.ForgeType `xorm:"VARCHAR(250) 'type'"`
	URL               string          `xorm:"VARCHAR(500) 'url'"`
	Client            string          `xorm:"VARCHAR(250) 'client'"`
	ClientSecret      string          `xorm:"VARCHAR(250) 'client_secret'"`
	SkipVerify        bool            `xorm:"bool 'skip_verify'"`
	OAuthHost         string          `xorm:"VARCHAR(250) 'oauth_host'"` // public url for oauth if different from url
	AdditionalOptions map[string]any  `xorm:"json 'additional_options'"`
}

func (forgeV008) TableName() string {
	return "forge"
}

var setForgeID = xormigrate.Migration{
	ID: "set-forge-id",
	MigrateSession: func(sess *xorm.Session) (err error) {
		if err := sess.Sync(new(userV008), new(repoV008), new(forgeV008), new(model.Org)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		_, err = sess.Exec(fmt.Sprintf("UPDATE `%s` SET forge_id=1;", userV008{}.TableName()))
		if err != nil {
			return err
		}

		_, err = sess.Exec(fmt.Sprintf("UPDATE `%s` SET forge_id=1;", model.Org{}.TableName()))
		if err != nil {
			return err
		}

		_, err = sess.Exec(fmt.Sprintf("UPDATE `%s` SET forge_id=1;", repoV008{}.TableName()))
		return err
	},
}
