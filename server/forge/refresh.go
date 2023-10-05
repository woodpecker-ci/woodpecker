package forge

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Refresher refreshes an oauth token and expiration for the given user. It
// returns true if the token was refreshed, false if the token was not refreshed,
// and error if it failed to refresh.
type Refresher interface {
	Refresh(context.Context, *model.User) (bool, error)
}

func Refresh(c context.Context, forge Forge, _store store.Store, user *model.User) {
	if refresher, ok := forge.(Refresher); ok {
		// Check to see if the user token is expired or
		// will expire within the next 30 minutes (1800 seconds).
		// If not, there is nothing we really need to do here.
		if time.Now().UTC().Unix() < (user.Expiry - 1800) {
			return
		}

		ok, err := refresher.Refresh(c, user)
		if err != nil {
			log.Error().Err(err).Msgf("refresh oauth token of user '%s' failed", user.Login)
		} else if ok {
			if err := _store.UpdateUser(user); err != nil {
				log.Error().Err(err).Msg("fail to save user to store after refresh oauth token")
			}
		}
	}
}
