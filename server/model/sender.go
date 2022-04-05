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

import "context"

type SenderService interface {
	SenderAllowed(context.Context, *User, *Repo, *Build, *Config) (bool, error)
	SenderCreate(context.Context, *Repo, *Sender) error
	SenderUpdate(context.Context, *Repo, *Sender) error
	SenderDelete(context.Context, *Repo, string) error
	SenderList(context.Context, *Repo) ([]*Sender, error)
}

type SenderStore interface {
	SenderFind(*Repo, string) (*Sender, error)
	SenderList(*Repo) ([]*Sender, error)
	SenderCreate(*Sender) error
	SenderUpdate(*Sender) error
	SenderDelete(*Sender) error
}

type Sender struct {
	ID     int64    `json:"-"      xorm:"pk autoincr 'sender_id'"`
	RepoID int64    `json:"-"      xorm:"UNIQUE(s) INDEX 'sender_repo_id'"`
	Login  string   `json:"login"  xorm:"UNIQUE(s) 'sender_login'"`
	Allow  bool     `json:"allow"  xorm:"sender_allow"`
	Block  bool     `json:"block"  xorm:"sender_block"`
	Branch []string `json:"branch" xorm:"-"`
	Deploy []string `json:"deploy" xorm:"-"`
	Event  []string `json:"event"  xorm:"-"`
}

// TableName return database table name for xorm
func (Sender) TableName() string {
	return "senders"
}
