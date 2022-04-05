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
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
)

var (
	errSecretNameInvalid  = errors.New("Invalid Secret Name")
	errSecretValueInvalid = errors.New("Invalid Secret Value")
	errSecretEventInvalid = errors.New("Invalid Secret Event")
)

// SecretService defines a service for managing secrets.
type SecretService interface {
	SecretFind(context.Context, *Repo, string) (*Secret, error)
	SecretList(context.Context, *Repo) ([]*Secret, error)
	SecretCreate(context.Context, *Repo, *Secret) error
	SecretUpdate(context.Context, *Repo, *Secret) error
	SecretDelete(context.Context, *Repo, string) error
}

// SecretStore persists secret information to storage.
type SecretStore interface {
	SecretFind(*Repo, string) (*Secret, error)
	SecretList(*Repo) ([]*Secret, error)
	SecretCreate(*Secret) error
	SecretUpdate(*Secret) error
	SecretDelete(*Secret) error
}

// TODO: rename to environment_variable and make it a secret by setting conceal=true
// Secret represents a secret variable, such as a password or token.
type Secret struct {
	ID         int64          `json:"id"              xorm:"pk autoincr 'secret_id'"`
	RepoID     int64          `json:"-"               xorm:"UNIQUE(s) INDEX 'secret_repo_id'"`
	Name       string         `json:"name"            xorm:"UNIQUE(s) INDEX 'secret_name'"`
	Value      string         `json:"value,omitempty" xorm:"TEXT 'secret_value'"`
	Images     []string       `json:"image"           xorm:"json 'secret_images'"`
	Events     []WebhookEvent `json:"event"           xorm:"json 'secret_events'"`
	SkipVerify bool           `json:"-"               xorm:"secret_skip_verify"` // TODO: remove
	Conceal    bool           `json:"-"               xorm:"secret_conceal"`
}

// TableName return database table name for xorm
func (Secret) TableName() string {
	return "secrets"
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
			return fmt.Errorf("%s: '%s'", errSecretEventInvalid, event)
		}
	}
	if len(s.Events) == 0 {
		return fmt.Errorf("%s: no event specified", errSecretEventInvalid)
	}

	for _, image := range s.Images {
		if len(image) == 0 {
			return fmt.Errorf("empty image in images")
		}
		if !validDockerImageString.MatchString(image) {
			return fmt.Errorf("image '%s' do not match regexp '%s'", image, validDockerImageString.String())
		}
	}

	switch {
	case len(s.Name) == 0:
		return errSecretNameInvalid
	case len(s.Value) == 0:
		return errSecretValueInvalid
	default:
		return nil
	}
}

// Copy makes a copy of the secret without the value.
func (s *Secret) Copy() *Secret {
	return &Secret{
		ID:     s.ID,
		RepoID: s.RepoID,
		Name:   s.Name,
		Images: s.Images,
		Events: s.Events,
	}
}
