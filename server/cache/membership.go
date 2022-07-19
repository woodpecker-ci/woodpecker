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
	// IsMember returns true if the user is a member of the organization.
	IsMember(ctx context.Context, u *model.User, owner string) (bool, error)

	// IsAdmin returns true if the user is an admin of the organization.
	IsAdmin(ctx context.Context, u *model.User, owner string) (bool, error)
}

type membership struct {
	Member bool
	Admin  bool
}

type membershipCache struct {
	Remote remote.Remote
	Cache  *ttlcache.Cache[string, membership]
	TTL    time.Duration
}

// NewMembershipService creates a new membership service.
func NewMembershipService(r remote.Remote) MembershipService {
	return &membershipCache{
		TTL:    10 * time.Minute,
		Remote: r,
		Cache:  ttlcache.New(ttlcache.WithDisableTouchOnHit[string, membership]()),
	}
}

func (c *membershipCache) get(ctx context.Context, u *model.User, owner string) (bool, bool, error) {
	key := u.Login + "/" + owner
	// Error can be safely ignored, as cache can only return error from loaders.
	item, _ := c.Cache.Get(key)
	if item != nil && !item.IsExpired() {
		return item.Value().Member, item.Value().Admin, nil
	}

	member, admin, err := c.Remote.OrgMembership(ctx, u, owner)
	if err != nil {
		return false, false, err
	}
	c.Cache.Set(key, membership{Member: member, Admin: admin}, c.TTL)
	return member, admin, nil
}

// IsMember returns true if the user is a member of the organization.
func (c *membershipCache) IsMember(ctx context.Context, u *model.User, owner string) (bool, error) {
	member, _, err := c.get(ctx, u, owner)
	if err != nil {
		return false, err
	}
	return member, nil
}

// IsAdmin returns true if the user is an admin of the organization.
func (c *membershipCache) IsAdmin(ctx context.Context, u *model.User, owner string) (bool, error) {
	_, admin, err := c.get(ctx, u, owner)
	if err != nil {
		return false, err
	}
	return admin, nil
}
