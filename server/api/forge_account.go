// Copyright 2026 Woodpecker Authors
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

package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// errNameConflict is returned if the name of a forge account is still used by
// a stored user or org of another forge account, which is gone or can't be
// checked. An admin has to resolve the conflict.
var errNameConflict = errors.New("the name is still used by another forge account, an admin has to resolve the conflict")

// syncOrg returns the stored org of the given forge org and creates it if
// needed. Orgs are found by their forge remote id, and by name only if they
// were stored without one.
func syncOrg(ctx context.Context, _store store.Store, _forge forge.Forge, user *model.User, from *model.Org) (*model.Org, error) {
	if from.ForgeRemoteID.IsValid() {
		org, err := _store.OrgFindByRemoteID(from.ForgeID, from.ForgeRemoteID)
		if err == nil {
			if org.Name != from.Name {
				org.Name = from.Name
				org.ConflictForgeRemoteID = ""
				if err := _store.OrgUpdate(org); err != nil {
					// it is still the right org, only its name is outdated
					log.Warn().Err(err).Msgf("cannot rename org %d to %s", org.ID, from.Name)
				}
			}
			return org, nil
		}
		if !errors.Is(err, types.ErrRecordNotExist) {
			return nil, err
		}
	}

	org, err := _store.OrgFindByName(from.Name, from.ForgeID)
	switch {
	case errors.Is(err, types.ErrRecordNotExist):
		return from, _store.OrgCreate(from)
	case err != nil:
		return nil, err
	case !from.ForgeRemoteID.IsValid():
		// the forge does not return forge remote ids
		return org, nil
	case !org.ForgeRemoteID.IsValid():
		// stored before orgs had forge remote ids
		org.ForgeRemoteID = from.ForgeRemoteID
		org.IsUser = from.IsUser
		return org, _store.OrgUpdate(org)
	}

	// the name is still used by the org of another forge account
	newName, err := forgeAccountName(ctx, _forge, user, org.ForgeRemoteID, org.IsUser, org.Name)
	if errors.Is(err, errNameConflict) {
		org.ConflictForgeRemoteID = from.ForgeRemoteID
		if err := _store.OrgUpdate(org); err != nil {
			return nil, err
		}
		return nil, errNameConflict
	}
	if err != nil {
		return nil, err
	}

	org.Name = newName
	org.ConflictForgeRemoteID = ""
	if err := _store.OrgUpdate(org); err != nil {
		return nil, err
	}
	return from, _store.OrgCreate(from)
}

// releaseUserLogin gives a stored user, whose login is now used by the forge
// account of forgeUser, the new login of its own forge account. If that
// account is gone or can't be checked, the stored user is marked as conflict.
func releaseUserLogin(ctx context.Context, _store store.Store, _forge forge.Forge, forgeUser, user *model.User) error {
	newLogin, err := forgeAccountName(ctx, _forge, forgeUser, user.ForgeRemoteID, true, user.Login)
	if errors.Is(err, errNameConflict) {
		user.ConflictForgeRemoteID = forgeUser.ForgeRemoteID
		if err := _store.UpdateUser(user); err != nil {
			return err
		}
		return errNameConflict
	}
	if err != nil {
		return err
	}

	user.Login = newLogin
	user.ConflictForgeRemoteID = ""
	return _store.UpdateUser(user)
}

// forgeAccountName returns the current name of the forge account with the
// given remote id. It returns errNameConflict if the account is gone, can't be
// looked up, or still has the given name.
func forgeAccountName(ctx context.Context, _forge forge.Forge, u *model.User, remoteID model.ForgeRemoteID, isUser bool, name string) (string, error) {
	account, err := _forge.OrgByRemoteID(ctx, u, remoteID, isUser)
	switch {
	case errors.Is(err, forge_types.ErrOrgNotFound), errors.Is(err, forge_types.ErrNotImplemented):
		return "", errNameConflict
	case err != nil:
		return "", err
	case strings.EqualFold(account.Name, name):
		return "", errNameConflict
	}
	return account.Name, nil
}

// remapAllowed checks if a stored user or org with a name conflict may be
// given to the forge account that now uses its name: only if its own forge
// account is gone, as checked with the token of the admin. Else it writes an
// error response.
func remapAllowed(c *gin.Context, forgeID int64, remoteID, conflictRemoteID model.ForgeRemoteID, isUser bool) bool {
	if !conflictRemoteID.IsValid() {
		c.String(http.StatusBadRequest, "There is no name conflict to resolve.")
		return false
	}

	admin := session.User(c)
	if admin.ForgeID != forgeID {
		c.String(http.StatusBadRequest, "Only an admin of the same forge can check the old forge account.")
		return false
	}

	_forge, err := server.Config.Services.Manager.ForgeByID(forgeID)
	if err != nil {
		log.Error().Err(err).Msg("cannot get forge")
		c.AbortWithStatus(http.StatusInternalServerError)
		return false
	}
	forge.Refresh(c, _forge, store.FromContext(c), admin)

	_, err = _forge.OrgByRemoteID(c, admin, remoteID, isUser)
	switch {
	case errors.Is(err, forge_types.ErrOrgNotFound):
		return true
	case err == nil:
		c.String(http.StatusConflict, "The old forge account still exists, so it keeps its data.")
	case errors.Is(err, forge_types.ErrNotImplemented):
		c.String(http.StatusUnprocessableEntity, "The forge can't tell if the old forge account still exists. Delete it instead.")
	default:
		log.Error().Err(err).Msgf("cannot look up forge account %s", remoteID)
		c.String(http.StatusInternalServerError, "Could not look up the old forge account.")
	}
	return false
}
