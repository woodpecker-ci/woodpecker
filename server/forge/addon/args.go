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

package addon

import (
	"errors"
	"fmt"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

type argumentsRepo struct {
	U        *modelUser          `json:"u"`
	RemoteID model.ForgeRemoteID `json:"remote_id"`
	Owner    string              `json:"owner"`
	Name     string              `json:"name"`
}

type argumentsFileDir struct {
	U *modelUser      `json:"u"`
	R *modelRepo      `json:"r"`
	B *model.Pipeline `json:"b"`
	F string          `json:"f"`
}

type argumentsStatus struct {
	U *modelUser      `json:"u"`
	R *modelRepo      `json:"r"`
	B *model.Pipeline `json:"b"`
	P *model.Workflow `json:"p"`
}

type argumentsNetrc struct {
	U *modelUser `json:"u"`
	R *modelRepo `json:"r"`
}

type argumentsActivateDeactivate struct {
	U    *modelUser `json:"u"`
	R    *modelRepo `json:"r"`
	Link string     `json:"link"`
}

type argumentsBranchesPullRequests struct {
	U *modelUser         `json:"u"`
	R *modelRepo         `json:"r"`
	P *model.ListOptions `json:"p"`
}

type argumentsBranchHead struct {
	U      *modelUser `json:"u"`
	R      *modelRepo `json:"r"`
	Branch string     `json:"branch"`
}

type argumentsOrgMembershipOrg struct {
	U   *modelUser `json:"u"`
	Org string     `json:"org"`
}

type argumentsTeams struct {
	U *modelUser         `json:"u"`
	P *model.ListOptions `json:"p"`
}

type argumentsRepos struct {
	U *modelUser         `json:"u"`
	P *model.ListOptions `json:"p"`
}

// errorKind names an error of the forge that the server checks for with
// errors.Is. The RPC layer only sends the message of an error, so the kind is
// sent next to the message. The values are part of the protocol, never rename them.
type errorKind string

const (
	errorKindIgnoreEvent      errorKind = "ignore_event"
	errorKindInvalidSignature errorKind = "invalid_signature"
	errorKindNotImplemented   errorKind = "not_implemented"
	errorKindRepoNotFound     errorKind = "repo_not_found"
	errorKindConfigNotFound   errorKind = "config_not_found"
)

// errorKindMap maps the kind to the error the server checks for.
var errorKindMap = map[errorKind]error{
	errorKindIgnoreEvent:      &types.ErrIgnoreEvent{},
	errorKindInvalidSignature: types.ErrInvalidWebhookSignature,
	errorKindNotImplemented:   types.ErrNotImplemented,
	errorKindRepoNotFound:     types.ErrRepoNotFound,
	errorKindConfigNotFound:   &types.ErrConfigNotFound{},
}

// encodeErrorKinds returns the kind and the message of an error the server
// checks for. For every other error the result is not found.
func encodeErrorKinds(err error) (found bool, kind errorKind, msg string) {
	for k, target := range errorKindMap {
		if errors.Is(err, target) {
			return true, k, err.Error()
		}
	}
	return false, "", ""
}

// decodeErrorKinds returns the error an addon sent as kind and message.
// A kind this server does not know is returned as ErrAddon with the message only.
func decodeErrorKinds(kind errorKind, msg string) error {
	if target, found := errorKindMap[kind]; found {
		return &ErrAddon{Message: msg, Err: target}
	}
	return &ErrAddon{Message: msg, Err: fmt.Errorf("addon sent unknown error kind %q", kind)}
}

type baseErrorResponse struct {
	ErrorKind    errorKind `json:"error_kind,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// errorResponse returns the response part for an error the server checks for.
// Every other error is returned as is and sent as error of the RPC call.
func errorResponse(err error) (baseErrorResponse, error) {
	if err == nil {
		return baseErrorResponse{}, nil
	}
	found, kind, msg := encodeErrorKinds(err)
	if !found {
		return baseErrorResponse{}, err
	}
	return baseErrorResponse{ErrorKind: kind, ErrorMessage: msg}, nil
}

// asError returns the error the addon sent in the response, or nil.
func (r baseErrorResponse) asError() error {
	if r.ErrorKind == "" {
		return nil
	}
	return decodeErrorKinds(r.ErrorKind, r.ErrorMessage)
}

type responseHook struct {
	baseErrorResponse
	Repo     *modelRepo      `json:"repo"`
	Pipeline *model.Pipeline `json:"pipeline"`
}

type responseTeams struct {
	baseErrorResponse
	Teams []*model.Team `json:"teams"`
}

type responseRepo struct {
	baseErrorResponse
	Repo *modelRepo `json:"repo"`
}

type responseFile struct {
	baseErrorResponse
	File []byte `json:"file"`
}

type responseDir struct {
	baseErrorResponse
	Files []*types.FileMeta `json:"files"`
}

type responseLogin struct {
	User        *modelUser `json:"user"`
	RedirectURL string     `json:"redirect_url"`
}

type httpRequest struct {
	Method string              `json:"method"`
	URL    string              `json:"url"`
	Header map[string][]string `json:"header"`
	Form   map[string][]string `json:"form"`
	Body   []byte              `json:"body"`
	Secret string              `json:"secret"`
}

// modelUser is an extension of model.User to marshal all fields to JSON.
type modelUser struct {
	User *model.User `json:"user"`

	ForgeRemoteID model.ForgeRemoteID `json:"forge_remote_id"`

	// Token is the oauth2 token.
	Token string `json:"token"`

	// Secret is the oauth2 token secret.
	Secret string `json:"secret"`

	// Expiry is the token and secret expiration timestamp.
	Expiry int64 `json:"expiry"`

	// Hash is a unique token used to sign tokens.
	Hash string `json:"hash"`
}

func (m *modelUser) asModel() *model.User {
	if m == nil {
		return nil
	}
	m.User.ForgeRemoteID = m.ForgeRemoteID
	m.User.AccessToken = m.Token
	m.User.RefreshToken = m.Secret
	m.User.Expiry = m.Expiry
	m.User.Hash = m.Hash
	return m.User
}

func modelUserFromModel(u *model.User) *modelUser {
	if u == nil {
		return nil
	}
	return &modelUser{
		User:          u,
		ForgeRemoteID: u.ForgeRemoteID,
		Token:         u.AccessToken,
		Secret:        u.RefreshToken,
		Expiry:        u.Expiry,
		Hash:          u.Hash,
	}
}

// modelRepo is an extension of model.Repo to marshal all fields to JSON.
type modelRepo struct {
	Repo          *model.Repo `json:"repo"`
	UserID        int64       `json:"user_id"`
	WebhookSecret string      `json:"hash"`
	Perm          *model.Perm `json:"perm"`
}

func (m *modelRepo) asModel() *model.Repo {
	if m == nil {
		return nil
	}
	m.Repo.UserID = m.UserID
	m.Repo.WebhookSecret = m.WebhookSecret
	m.Repo.Perm = m.Perm
	return m.Repo
}

func modelRepoFromModel(r *model.Repo) *modelRepo {
	if r == nil {
		return nil
	}
	return &modelRepo{
		Repo:          r,
		UserID:        r.UserID,
		WebhookSecret: r.WebhookSecret,
		Perm:          r.Perm,
	}
}
