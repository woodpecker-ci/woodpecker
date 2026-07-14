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

package store

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/encryption/types"
)

const (
	userAccessTokenAADSuffix  = "/access-token"
	userRefreshTokenAADSuffix = "/refresh-token"
)

// EncryptedUserStore is a model.UserStore decorator that encrypts the forge
// access and refresh tokens of users before they are stored and decrypts
// them after loading.
type EncryptedUserStore struct {
	store      model.UserStore
	encryption types.EncryptionService
}

// Ensure wrapper match interface.
var _ model.UserStore = new(EncryptedUserStore)

func NewUserStore(userStore model.UserStore) *EncryptedUserStore {
	return &EncryptedUserStore{store: userStore}
}

func (wrapper *EncryptedUserStore) SetEncryptionService(service types.EncryptionService) error {
	if wrapper.encryption != nil {
		return errors.New(errMessageInitSeveralTimes)
	}
	wrapper.encryption = service
	return nil
}

func (wrapper *EncryptedUserStore) EnableEncryption() error {
	log.Warn().Msg(logMessageEnablingUsersEncryption)
	users, err := wrapper.store.GetUserList(&model.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToEnableUsers, err)
	}
	for _, user := range users {
		// tokens already carrying the marker were encrypted by an
		// earlier, interrupted run and are skipped inside encryptWith
		if err := wrapper.encrypt(user); err != nil {
			return err
		}
		if err := wrapper.store.UpdateUser(user); err != nil {
			log.Err(err).Msg(errMessageTemplateUserStorageError)
			return err
		}
	}
	log.Warn().Msg(logMessageEnablingUsersEncryptionSuccess)
	return nil
}

func (wrapper *EncryptedUserStore) MigrateEncryption(newEncryptionService types.EncryptionService) error {
	log.Warn().Msg(logMessageMigratingUsersEncryption)
	users, err := wrapper.store.GetUserList(&model.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToMigrateUsers, err)
	}
	for _, user := range users {
		if err := wrapper.decrypt(user); err != nil {
			return err
		}
	}
	for _, user := range users {
		if err := wrapper.encryptWith(newEncryptionService, user); err != nil {
			return err
		}
		if err := wrapper.store.UpdateUser(user); err != nil {
			log.Err(err).Msg(errMessageTemplateUserStorageError)
			return err
		}
	}
	wrapper.encryption = newEncryptionService
	log.Warn().Msg(logMessageMigratingUsersEncryptionSuccess)
	return nil
}

func (wrapper *EncryptedUserStore) encrypt(user *model.User) error {
	return wrapper.encryptWith(wrapper.encryption, user)
}

func (wrapper *EncryptedUserStore) encryptWith(encryption types.EncryptionService, user *model.User) error {
	accessToken, err := encryptUserToken(encryption, user.AccessToken, userTokenAAD(user.ID, userAccessTokenAADSuffix))
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToEncryptUser, user.ID, err)
	}
	refreshToken, err := encryptUserToken(encryption, user.RefreshToken, userTokenAAD(user.ID, userRefreshTokenAADSuffix))
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToEncryptUser, user.ID, err)
	}
	user.AccessToken, user.RefreshToken = accessToken, refreshToken
	return nil
}

func (wrapper *EncryptedUserStore) decrypt(user *model.User) error {
	accessToken, err := wrapper.encryption.Decrypt(user.AccessToken, userTokenAAD(user.ID, userAccessTokenAADSuffix))
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToDecryptUser, user.ID, err)
	}
	refreshToken, err := wrapper.encryption.Decrypt(user.RefreshToken, userTokenAAD(user.ID, userRefreshTokenAADSuffix))
	if err != nil {
		return fmt.Errorf(errMessageTemplateFailedToDecryptUser, user.ID, err)
	}
	user.AccessToken, user.RefreshToken = accessToken, refreshToken
	return nil
}

// encryptUserToken encrypts a single token. Empty tokens and tokens that
// already carry the encrypted value marker (from an interrupted enable run)
// are returned unchanged.
func encryptUserToken(encryption types.EncryptionService, token, aad string) (string, error) {
	if token == "" || isMarkedValue(token) {
		return token, nil
	}
	return encryption.Encrypt(token, aad)
}

func userTokenAAD(userID int64, suffix string) string {
	return strconv.Itoa(int(userID)) + suffix
}

func (wrapper *EncryptedUserStore) GetUser(id int64) (*model.User, error) {
	user, err := wrapper.store.GetUser(id)
	if err != nil {
		return nil, err
	}
	if err := wrapper.decrypt(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (wrapper *EncryptedUserStore) GetUserByRemoteID(forgeID int64, remoteID model.ForgeRemoteID) (*model.User, error) {
	user, err := wrapper.store.GetUserByRemoteID(forgeID, remoteID)
	if err != nil {
		return nil, err
	}
	if err := wrapper.decrypt(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (wrapper *EncryptedUserStore) GetUserByLogin(forgeID int64, login string) (*model.User, error) {
	user, err := wrapper.store.GetUserByLogin(forgeID, login)
	if err != nil {
		return nil, err
	}
	if err := wrapper.decrypt(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (wrapper *EncryptedUserStore) GetUserList(p *model.ListOptions) ([]*model.User, error) {
	users, err := wrapper.store.GetUserList(p)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if err := wrapper.decrypt(user); err != nil {
			return nil, err
		}
	}
	return users, nil
}

// CreateUser stores the user and encrypts its tokens. The tokens can only be
// encrypted after the row exists, because the user id is part of the
// associated data; a crash in between leaves the tokens in plaintext, which
// stays readable through the plaintext passthrough and is encrypted again by
// the next update.
func (wrapper *EncryptedUserStore) CreateUser(user *model.User) error {
	if err := wrapper.store.CreateUser(user); err != nil {
		return err
	}
	return wrapper.UpdateUser(user)
}

func (wrapper *EncryptedUserStore) UpdateUser(user *model.User) error {
	plainAccessToken, plainRefreshToken := user.AccessToken, user.RefreshToken
	defer func() { user.AccessToken, user.RefreshToken = plainAccessToken, plainRefreshToken }()

	if err := wrapper.encrypt(user); err != nil {
		return err
	}

	return wrapper.store.UpdateUser(user)
}
