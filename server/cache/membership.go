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
	"time"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"

	"github.com/lafriks/ttlcache/v3"
)

// MembershipService is a service to check for user membership.
type MembershipService interface {
	// Get returns if the user is a member of the organization.
	Get(ctx context.Context, u *model.User, name string) (*model.OrgPerm, error)
}

type membershipCache struct {
	Remote remote.Remote
	Cache  *ttlcache.Cache[string, *model.OrgPerm]
	TTL    time.Duration
}

// NewMembershipService creates a new membership service.
func NewMembershipService(r remote.Remote) MembershipService {
	return &membershipCache{
		TTL:    10 * time.Minute,
		Remote: r,
		Cache:  ttlcache.New(ttlcache.WithDisableTouchOnHit[string, *model.OrgPerm]()),
	}
}

// Get returns if the user is a member of the organization.
func (c *membershipCache) Get(ctx context.Context, u *model.User, name string) (*model.OrgPerm, error) {
	key := u.Login + "/" + name
	// Error can be safely ignored, as cache can only return error from loaders.
	item, _ := c.Cache.Get(key)
	if item != nil && !item.IsExpired() {
		return item.Value(), nil
	}

	member, admin, err := c.Remote.OrgMembership(ctx, u, name)
	if err != nil {
		return nil, err
	}
	perm := &model.OrgPerm{Member: member, Admin: admin}
	c.Cache.Set(key, perm, c.TTL)
	return perm, nil
}
