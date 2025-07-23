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

package common

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

func ExtractHostFromCloneURL(cloneURL string) (string, error) {
	u, err := url.Parse(cloneURL)
	if err != nil {
		return "", err
	}

	if !strings.Contains(u.Host, ":") {
		return u.Host, nil
	}

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", err
	}

	return host, nil
}

func UserToken(ctx context.Context, r *model.Repo, u *model.User) string {
	if u != nil {
		return u.AccessToken
	}

	user, err := RepoUser(ctx, r)
	if err != nil {
		log.Error().Err(err).Msg("could not get repo user")
		return ""
	}
	return user.AccessToken
}

func RepoUser(ctx context.Context, r *model.Repo) (*model.User, error) {
	_store, ok := store.TryFromContext(ctx)
	if !ok {
		return nil, errors.New("could not get store from context")
	}
	if r == nil {
		log.Error().Msg("cannot get user token by empty repo")
		return nil, errors.New("cannot get user token by empty repo")
	}
	user, err := _store.GetUser(r.UserID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func RepoUserForgeID(ctx context.Context, repoForgeID model.ForgeRemoteID) (*model.User, error) {
	_store, ok := store.TryFromContext(ctx)
	if !ok {
		return nil, errors.New("could not get store from context")
	}
	r, err := _store.GetRepoForgeID(repoForgeID)
	if err != nil {
		return nil, err
	}
	return RepoUser(ctx, r)
}
