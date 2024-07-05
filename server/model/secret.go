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

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
)

var (
	ErrSecretNameInvalid  = errors.New("invalid secret name")
	ErrSecretImageInvalid = errors.New("invalid secret image")
	ErrSecretValueInvalid = errors.New("invalid secret value")
	ErrSecretEventInvalid = errors.New("invalid secret event")
)

// SecretStore persists secret information to storage.
type SecretStore interface {
	SecretFind(*Repo, string) (*Secret, error)
	SecretList(*Repo, bool, *ListOptions) ([]*Secret, error)
	SecretCreate(*Secret) error
	SecretUpdate(*Secret) error
	SecretDelete(*Secret) error
	OrgSecretFind(int64, string) (*Secret, error)
	OrgSecretList(int64, *ListOptions) ([]*Secret, error)
	GlobalSecretFind(string) (*Secret, error)
	GlobalSecretList(*ListOptions) ([]*Secret, error)
	SecretListAll() ([]*Secret, error)
}

// Secret represents a secret variable, such as a password or token.
type Secret struct {
	ID     int64          `json:"id"              xorm:"pk autoincr 'id'"`
	OrgID  int64          `json:"org_id"          xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'org_id'"`
	RepoID int64          `json:"repo_id"         xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'repo_id'"`
	Name   string         `json:"name"            xorm:"NOT NULL UNIQUE(s) INDEX 'name'"`
	Value  string         `json:"value,omitempty" xorm:"TEXT 'value'"`
	Images []string       `json:"images"          xorm:"json 'images'"`
	Events []WebhookEvent `json:"events"          xorm:"json 'events'"`
} //	@name Secret

// TableName return database table name for xorm.
func (Secret) TableName() string {
	return "secrets"
}

// BeforeInsert will sort events before inserted into database.
func (s *Secret) BeforeInsert() {
	s.Events = sortEvents(s.Events)
}

// Global secret.
func (s Secret) IsGlobal() bool {
	return s.RepoID == 0 && s.OrgID == 0
}

// Organization secret.
func (s Secret) IsOrganization() bool {
	return s.RepoID == 0 && s.OrgID != 0
}

// Repository secret.
func (s Secret) IsRepository() bool {
	return s.RepoID != 0 && s.OrgID == 0
}

var validDockerImageString = regexp.MustCompile(
	`^(` +
		`[\w\d\-_\.]+` + // hostname
		`(:\d+)?` + // optional port
		`/)?` + // optional hostname + port
		`([\w\d\-_\.][\w\d\-_\.\/]*/)?` + // optional url prefix
		`([\w\d\-_]+)` + // image name
		`(:[\w\d\-_]+)?` + // optional image tag
		`$`,
)

// Validate validates the required fields and formats.
func (s *Secret) Validate() error {
	for _, event := range s.Events {
		if err := event.Validate(); err != nil {
			return errors.Join(err, ErrSecretEventInvalid)
		}
	}
	if len(s.Events) == 0 {
		return fmt.Errorf("%w: no event specified", ErrSecretEventInvalid)
	}

	for _, image := range s.Images {
		if len(image) == 0 {
			return fmt.Errorf("%w: empty image in images", ErrSecretImageInvalid)
		}
		if !validDockerImageString.MatchString(image) {
			return fmt.Errorf("%w: image '%s' do not match regexp '%s'", ErrSecretImageInvalid, image, validDockerImageString.String())
		}
	}

	switch {
	case len(s.Name) == 0:
		return fmt.Errorf("%w: empty name", ErrSecretNameInvalid)
	case len(s.Value) == 0:
		return fmt.Errorf("%w: empty value", ErrSecretValueInvalid)
	default:
		return nil
	}
}

// Copy makes a copy of the secret without the value.
func (s *Secret) Copy() *Secret {
	return &Secret{
		ID:     s.ID,
		OrgID:  s.OrgID,
		RepoID: s.RepoID,
		Name:   s.Name,
		Images: s.Images,
		Events: sortEvents(s.Events),
	}
}

func sortEvents(wel WebhookEventList) WebhookEventList {
	sort.Sort(wel)
	return wel
}
