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

package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"

	"github.com/lafriks/ttlcache/v3"
)

// MembershipService is a service to check for user membership.
type MembershipService interface {
	// Get returns if the user is a member of the organization.
	Get(ctx context.Context, _forge forge.Forge, u *model.User, org string) (*model.OrgPerm, error)
}

type membershipCache struct {
	cache *ttlcache.Cache[string, *model.OrgPerm]
	store store.Store
	ttl   time.Duration
}

// NewMembershipService creates a new membership service.
func NewMembershipService(_store store.Store) MembershipService {
	return &membershipCache{
		ttl:   10 * time.Minute,
		store: _store,
		cache: ttlcache.New(ttlcache.WithDisableTouchOnHit[string, *model.OrgPerm]()),
	}
}

// Get returns if the user is a member of the organization.
func (c *membershipCache) Get(ctx context.Context, _forge forge.Forge, u *model.User, org string) (*model.OrgPerm, error) {
	key := fmt.Sprintf("%s-%s", u.ForgeRemoteID, org)
	// Error can be safely ignored, as cache can only return error from loaders.
	item, _ := c.cache.Get(key)
	if item != nil && !item.IsExpired() {
		return item.Value(), nil
	}

	// _forge, err := loader.GetForgeFromUser(c.store, u)
	// if err != nil {
	// 	return nil, err
	// }

	perm, err := _forge.OrgMembership(ctx, u, org)
	if err != nil {
		return nil, err
	}
	c.cache.Set(key, perm, c.ttl)
	return perm, nil
}
