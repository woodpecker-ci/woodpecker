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
	"path/filepath"
	"regexp"
	"sort"
)

var (
	ErrSecretNameInvalid  = errors.New("Invalid Secret Name")
	ErrSecretImageInvalid = errors.New("Invalid Secret Image")
	ErrSecretValueInvalid = errors.New("Invalid Secret Value")
	ErrSecretEventInvalid = errors.New("Invalid Secret Event")
)

// SecretService defines a service for managing secrets.
type SecretService interface {
	SecretListPipeline(*Repo, *Pipeline) ([]*Secret, error)
	// Repository secrets
	SecretFind(*Repo, string) (*Secret, error)
	SecretList(*Repo) ([]*Secret, error)
	SecretCreate(*Repo, *Secret) error
	SecretUpdate(*Repo, *Secret) error
	SecretDelete(*Repo, string) error
	// Organization secrets
	OrgSecretFind(string, string) (*Secret, error)
	OrgSecretList(string) ([]*Secret, error)
	OrgSecretCreate(string, *Secret) error
	OrgSecretUpdate(string, *Secret) error
	OrgSecretDelete(string, string) error
	// Global secrets
	GlobalSecretFind(string) (*Secret, error)
	GlobalSecretList() ([]*Secret, error)
	GlobalSecretCreate(*Secret) error
	GlobalSecretUpdate(*Secret) error
	GlobalSecretDelete(string) error
}

// SecretStore persists secret information to storage.
type SecretStore interface {
	SecretFind(*Repo, string) (*Secret, error)
	SecretList(*Repo, bool) ([]*Secret, error)
	SecretCreate(*Secret) error
	SecretUpdate(*Secret) error
	SecretDelete(*Secret) error
	OrgSecretFind(string, string) (*Secret, error)
	OrgSecretList(string) ([]*Secret, error)
	GlobalSecretFind(string) (*Secret, error)
	GlobalSecretList() ([]*Secret, error)
	SecretListAll() ([]*Secret, error)
}

// Secret represents a secret variable, such as a password or token.
// swagger:model registry
type Secret struct {
	ID          int64          `json:"id"              xorm:"pk autoincr 'secret_id'"`
	Owner       string         `json:"-"               xorm:"NOT NULL DEFAULT '' UNIQUE(s) INDEX 'secret_owner'"`
	RepoID      int64          `json:"-"               xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'secret_repo_id'"`
	Name        string         `json:"name"            xorm:"NOT NULL UNIQUE(s) INDEX 'secret_name'"`
	Value       string         `json:"value,omitempty" xorm:"TEXT 'secret_value'"`
	Images      []string       `json:"image"           xorm:"json 'secret_images'"`
	PluginsOnly bool           `json:"plugins_only"    xorm:"secret_plugins_only"`
	Events      []WebhookEvent `json:"event"           xorm:"json 'secret_events'"`
	SkipVerify  bool           `json:"-"               xorm:"secret_skip_verify"`
	Conceal     bool           `json:"-"               xorm:"secret_conceal"`
}

// TableName return database table name for xorm
func (Secret) TableName() string {
	return "secrets"
}

// BeforeInsert will sort events before inserted into database
func (s *Secret) BeforeInsert() {
	s.Events = sortEvents(s.Events)
}

// Global secret.
func (s Secret) Global() bool {
	return s.RepoID == 0 && s.Owner == ""
}

// Organization secret.
func (s Secret) Organization() bool {
	return s.RepoID == 0 && s.Owner != ""
}

// Match returns true if an image and event match the restricted list.
func (s *Secret) Match(event WebhookEvent) bool {
	if len(s.Events) == 0 {
		return true
	}
	for _, pattern := range s.Events {
		if match, _ := filepath.Match(string(pattern), string(event)); match {
			return true
		}
	}
	return false
}

var validDockerImageString = regexp.MustCompile(
	`^([\w\d\-_\.\/]*` + // optional url prefix
		`[\w\d\-_]+` + // image name
		`)+` +
		`(:[\w\d\-_]+)?$`, // optional image tag
)

// Validate validates the required fields and formats.
func (s *Secret) Validate() error {
	for _, event := range s.Events {
		if !ValidateWebhookEvent(event) {
			return fmt.Errorf("%w: '%s'", ErrSecretEventInvalid, event)
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
		ID:          s.ID,
		Owner:       s.Owner,
		RepoID:      s.RepoID,
		Name:        s.Name,
		Images:      s.Images,
		PluginsOnly: s.PluginsOnly,
		Events:      sortEvents(s.Events),
	}
}

func sortEvents(wel WebhookEventList) WebhookEventList {
	sort.Sort(wel)
	return wel
}
