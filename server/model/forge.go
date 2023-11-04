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

type Forge struct {
	ID                int64  `xorm:"pk autoincr"`
	Type              string `xorm:"VARCHAR(250)"` // github, gitlab, gitea, gogs, bitbucket, stash, coding
	URL               string `xorm:"VARCHAR(500)"`
	Client            string `xorm:"VARCHAR(250)"`
	ClientSecret      string `xorm:"VARCHAR(250)"`
	SkipVerify        bool   `xorm:"bool"`
	AdditionalOptions string `xorm:"TEXT"` // TODO: think about the best format for this
	Created           int64  `xorm:"created"`
	Updated           int64  `xorm:"updated"`
}

// bitbucket
// Client: c.String("bitbucket-client"),
// Secret: c.String("bitbucket-secret"),

// gitea
// URL:        strings.TrimRight(server.String(), "/"),
// Client:     c.String("gitea-client"),
// Secret:     c.String("gitea-secret"),
// SkipVerify: c.Bool("gitea-skip-verify"),

// gitlab
// URL:          c.String("gitlab-server"),
// ClientID:     c.String("gitlab-client"),
// ClientSecret: c.String("gitlab-secret"),
// SkipVerify:   c.Bool("gitlab-skip-verify"),

// github
// URL:        c.String("github-server"),
// Client:     c.String("github-client"),
// Secret:     c.String("github-secret"),
// SkipVerify: c.Bool("github-skip-verify"),
// ### MergeRef:   c.Bool("github-merge-ref"),
