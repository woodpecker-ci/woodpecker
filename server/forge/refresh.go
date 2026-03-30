// Copyright 2023 Woodpecker Authors
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

package forge

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// Refresher is an optional interface for OAuth token refresh support.
//
// Tokens are checked before each operation. If expiring within 30 minutes,
// Refresh() is called automatically.
//
// Implementations: GitLab, Bitbucket (GitHub/Gitea tokens don't expire).
type Refresher interface {
	// Refresh attempts to refresh the user's OAuth access token.
	// Should update u.AccessToken, u.RefreshToken, and u.Expiry.
	// Returns true if any fields were updated.
	// Caller must persist updated user to database.
	Refresh(ctx context.Context, u *model.User) (bool, error)
}

// refreshGroup deduplicates concurrent token refresh calls per user.
// When multiple goroutines try to refresh the same user's token simultaneously
// (e.g., from concurrent API requests), only one refresh executes and the
// others wait for its result. This prevents race conditions with single-use
// refresh tokens (e.g., Forgejo with InvalidateRefreshTokens=true).
var refreshGroup singleflight.Group

// refreshResult carries token data through singleflight so waiting goroutines
// can update their own *model.User copies.
type refreshResult struct {
	AccessToken  string
	RefreshToken string
	Expiry       int64
}

func Refresh(ctx context.Context, forge Forge, _store store.Store, user *model.User) {
	// Remaining ttl of 30 minutes (1800 seconds) until a token is refreshed.
	const tokenMinTTL = 1800

	if refresher, ok := forge.(Refresher); ok {
		// Check to see if the user token is expired or
		// will expire within the next 30 minutes (1800 seconds).
		// If not, there is nothing we really need to do here.
		if time.Now().UTC().Unix() < (user.Expiry - tokenMinTTL) {
			return
		}

		key := fmt.Sprintf("refresh-%d", user.ID)
		result, err, _ := refreshGroup.Do(key, func() (any, error) {
			userUpdated, err := refresher.Refresh(ctx, user)
			if err != nil {
				return nil, err
			}
			if userUpdated {
				if err := _store.UpdateUser(user); err != nil {
					log.Error().Err(err).Msg("fail to save user to store after refresh oauth token")
				}
			}
			return &refreshResult{
				AccessToken:  user.AccessToken,
				RefreshToken: user.RefreshToken,
				Expiry:       user.Expiry,
			}, nil
		})
		if err != nil {
			log.Error().Err(err).Msgf("refresh oauth token of user '%s' failed", user.Login)
			return
		}

		// Copy fresh tokens into the caller's user object. This is necessary
		// because waiting goroutines have their own *model.User copies that
		// weren't passed to refresher.Refresh().
		if r, ok := result.(*refreshResult); ok {
			user.AccessToken = r.AccessToken
			user.RefreshToken = r.RefreshToken
			user.Expiry = r.Expiry
		}
	}
}
