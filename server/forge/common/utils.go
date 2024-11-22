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
	"net"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
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

	_store, ok := store.TryFromContext(ctx)
	if !ok {
		log.Error().Msg("could not get store from context")
		return ""
	}
	if r == nil {
		log.Error().Msg("cannot get user token by empty repo")
		return ""
	}
	user, err := _store.GetUser(r.UserID)
	if err != nil {
		return ""
	}
	return user.AccessToken
}
