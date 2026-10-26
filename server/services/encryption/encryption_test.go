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

//go:build test

package encryption_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tink-crypto/tink-go/v2/aead"
	insecure_clear_text_keyset "github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/encryption"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/datastore"
)

const (
	sampleKey = "encryption-ciphertext-sample"
	// A plaintext value that looks like an encrypted one.
	lookalike = "woodpecker-enc:v1:bG9va3MgZW5jcnlwdGVk"
)

var setup = encryption.Setup

// TestEncryption walks one database through the life of a server: it starts
// without encryption, gets a key, restarts, gets disabled, crashes in the
// middle of every data rewrite, rotates tink keys and shares the database
// with another server. After every start the data read through the store has
// to match what was written, and the data at rest has to be either all
// plaintext or all encrypted.
func TestEncryption(t *testing.T) {
	db := datastore.NewTestStore(t)
	want := newData()

	t.Run("without a key data is stored in plaintext", func(t *testing.T) {
		s := start(t, db)
		want.add(t, s, "plain")

		require.NoError(t, s.SecretCreate(&model.Secret{Name: "lookalike", Value: lookalike}))
		want.secrets["lookalike"] = lookalike
		require.NoError(t, s.RegistryCreate(&model.Registry{Address: "lookalike.example.com", Password: lookalike}))
		want.registries["lookalike.example.com"] = lookalike

		want.check(t, s, db, false)
	})

	t.Run("a raw key encrypts existing and new data", func(t *testing.T) {
		s := start(t, db, "--encryption-raw-key", "key-1")
		want.check(t, s, db, true)

		want.add(t, s, "encrypted")
		want.check(t, s, db, true)
	})

	t.Run("writes keep the plaintext in the caller's struct", func(t *testing.T) {
		s := start(t, db, "--encryption-raw-key", "key-1")

		secret := &model.Secret{Name: "caller", Value: "caller-value"}
		require.NoError(t, s.SecretCreate(secret))
		assert.NotZero(t, secret.ID)
		assert.Equal(t, "caller-value", secret.Value)
		secret.Value = "caller-value-2"
		require.NoError(t, s.SecretUpdate(secret))
		assert.Equal(t, "caller-value-2", secret.Value)
		want.secrets["caller"] = "caller-value-2"

		registry := &model.Registry{Address: "caller.example.com", Username: "u", Password: "caller-password"}
		require.NoError(t, s.RegistryCreate(registry))
		assert.Equal(t, "caller-password", registry.Password)
		want.registries["caller.example.com"] = "caller-password"

		user := newUser("caller")
		require.NoError(t, s.CreateUser(user))
		assert.NotZero(t, user.ID)
		assert.Equal(t, "caller-access", user.AccessToken)
		want.users["caller"] = tokens{"caller-access", "caller-refresh"}

		want.check(t, s, db, true)
	})

	t.Run("a restart with the same key leaves stored data alone", func(t *testing.T) {
		before := snapshot(t, db)
		s := start(t, db, "--encryption-raw-key", "key-1")
		want.check(t, s, db, true)
		assert.Equal(t, before, snapshot(t, db))
	})

	t.Run("a missing or wrong key stops the server", func(t *testing.T) {
		_, err := setupWith(t, db)
		assert.Error(t, err)
		_, err = setupWith(t, db, "--encryption-raw-key", "key-2")
		assert.Error(t, err)
		_, err = setupWith(t, db, "--encryption-raw-key", "key-1", "--encryption-tink-keyset", "/nonexistent")
		assert.Error(t, err)

		want.check(t, start(t, db, "--encryption-raw-key", "key-1"), db, true)
	})

	t.Run("the disable flag decrypts everything", func(t *testing.T) {
		s := start(t, db, "--encryption-raw-key", "key-1", "--encryption-disable-flag")
		want.check(t, s, db, false)

		s = start(t, db)
		want.add(t, s, "disabled")
		want.check(t, s, db, false)
	})

	t.Run("an interrupted enable never serves ciphertext and resumes", func(t *testing.T) {
		_, err := setupWith(t, crashAfter(db, 3), "--encryption-raw-key", "key-1")
		require.Error(t, err)

		// the key is recorded first, so a start without it does not serve
		// ciphertext
		_, err = setupWith(t, db)
		require.Error(t, err)

		s := start(t, db, "--encryption-raw-key", "key-1")
		want.check(t, s, db, true)
	})

	t.Run("an interrupted disable resumes with or without the flag", func(t *testing.T) {
		_, err := setupWith(t, crashAfter(db, 3), "--encryption-raw-key", "key-1", "--encryption-disable-flag")
		require.Error(t, err)

		// the admin changes their mind and keeps encryption
		s := start(t, db, "--encryption-raw-key", "key-1")
		want.check(t, s, db, true)

		_, err = setupWith(t, crashAfter(db, 3), "--encryption-raw-key", "key-1", "--encryption-disable-flag")
		require.Error(t, err)

		s = start(t, db, "--encryption-raw-key", "key-1", "--encryption-disable-flag")
		want.check(t, s, db, false)
	})

	t.Run("a crash while creating leaves no plaintext and no partial record", func(t *testing.T) {
		start(t, db, "--encryption-raw-key", "key-1")
		crashing := crashAfter(db, 0)
		s, err := setupWith(t, crashing, "--encryption-raw-key", "key-1")
		require.NoError(t, err)

		assert.Error(t, s.SecretCreate(&model.Secret{Name: "crash", Value: "crash-value"}))
		assert.Error(t, s.RegistryCreate(&model.Registry{Address: "crash.example.com", Password: "crash-password"}))
		crashing.budget = 1
		assert.Error(t, s.CreateUser(newUser("crash")))
		requireNoPlaintext(t, db, "crash-value", "crash-password", "crash-access", "crash-refresh")

		// a secret or registry is created completely or not at all, so the
		// create can be retried
		s = start(t, db, "--encryption-raw-key", "key-1")
		require.NoError(t, s.SecretCreate(&model.Secret{Name: "crash", Value: "crash-value"}))
		want.secrets["crash"] = "crash-value"
		require.NoError(t, s.RegistryCreate(&model.Registry{Address: "crash.example.com", Password: "crash-password"}))
		want.registries["crash.example.com"] = "crash-password"

		// the user exists without tokens until the next login stores them
		user, err := s.GetUserByLogin(1, "crash")
		require.NoError(t, err)
		assert.Empty(t, user.AccessToken)
		user.AccessToken, user.RefreshToken = "crash-access", "crash-refresh"
		require.NoError(t, s.UpdateUser(user))
		want.users["crash"] = tokens{"crash-access", "crash-refresh"}

		want.check(t, s, db, true)
	})

	keys := newKeyset(t)
	keysDir := t.TempDir()

	t.Run("tink encrypts and rotates to a new primary key", func(t *testing.T) {
		start(t, db, "--encryption-raw-key", "key-1", "--encryption-disable-flag")

		s := start(t, db, "--encryption-tink-keyset", keys.write(t, keysDir, "first.json"))
		want.check(t, s, db, true)
		want.add(t, s, "tink")

		oldKey := keys.rotate(t)
		s = start(t, db, "--encryption-tink-keyset", keys.write(t, keysDir, "both.json"))
		want.check(t, s, db, true)

		// with only the new key left everything must have been re-encrypted
		keys.remove(t, oldKey)
		s = start(t, db, "--encryption-tink-keyset", keys.write(t, keysDir, "second.json"))
		want.check(t, s, db, true)
	})

	t.Run("an interrupted tink rotation resumes", func(t *testing.T) {
		oldKey := keys.rotate(t)
		rotating := keys.write(t, keysDir, "rotating.json")
		_, err := setupWith(t, crashAfter(db, 3), "--encryption-tink-keyset", rotating)
		require.Error(t, err)

		s := start(t, db, "--encryption-tink-keyset", rotating)
		want.check(t, s, db, true)

		keys.remove(t, oldKey)
		s = start(t, db, "--encryption-tink-keyset", keys.write(t, keysDir, "rotated.json"))
		want.check(t, s, db, true)
	})

	t.Run("another server writing during and after a rotation loses nothing", func(t *testing.T) {
		other := start(t, db, "--encryption-tink-keyset", keys.write(t, keysDir, "current.json"))
		oldKey := keys.rotate(t)
		rotating := keys.write(t, keysDir, "concurrent.json")

		// the other server updates a secret after this one listed it for the
		// rotation
		interleaving := &interleavingStore{Store: db, before: func() {
			secret, err := other.GlobalSecretFind("plain")
			require.NoError(t, err)
			secret.Value = "plain-updated-during-rotation"
			require.NoError(t, other.SecretUpdate(secret))
			want.secrets["plain"] = secret.Value
		}}
		s := start(t, interleaving, "--encryption-tink-keyset", rotating)
		want.check(t, s, db, true)

		// the other server still writes with the old key until it restarts
		secret, err := db.GlobalSecretFind("plain")
		require.NoError(t, err)
		secret.Value = "plain-updated-by-old-server"
		require.NoError(t, other.SecretUpdate(secret))
		want.secrets["plain"] = secret.Value

		// that value stops a start that lacks the old key ...
		keys.remove(t, oldKey)
		newOnly := keys.write(t, keysDir, "new-only.json")
		_, err = setupWith(t, db, "--encryption-tink-keyset", newOnly)
		assert.ErrorContains(t, err, "could not decrypt secret")

		// ... and a start that has it re-encrypts the value
		want.check(t, start(t, db, "--encryption-tink-keyset", rotating), db, true)
		want.check(t, start(t, db, "--encryption-tink-keyset", newOnly), db, true)
	})

	t.Run("a keyset without the key in use stops the server", func(t *testing.T) {
		_, err := setupWith(t, db, "--encryption-tink-keyset", newKeyset(t).write(t, t.TempDir(), "other.json"))
		assert.Error(t, err)
	})

	t.Run("the disable flag decrypts tink encrypted data", func(t *testing.T) {
		s := start(t, db, "--encryption-tink-keyset", keys.write(t, keysDir, "last.json"), "--encryption-disable-flag")
		want.check(t, s, db, false)
	})
}

// start runs the server setup and requires it to succeed.
func start(t *testing.T, db store.Store, args ...string) store.Store {
	t.Helper()
	s, err := setupWith(t, db, args...)
	require.NoError(t, err)
	return s
}

func setupWith(t *testing.T, db store.Store, args ...string) (store.Store, error) {
	t.Helper()
	var (
		s        store.Store
		setupErr error
	)
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "encryption-raw-key"},
			&cli.StringFlag{Name: "encryption-tink-keyset"},
			&cli.BoolFlag{Name: "encryption-disable-flag"},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			s, setupErr = setup(c, db)
			return nil
		},
	}
	require.NoError(t, cmd.Run(t.Context(), append([]string{"woodpecker-server"}, args...)))
	return s, setupErr
}

type tokens struct{ access, refresh string }

// data is what the test wrote through the store and expects to read back.
type data struct {
	secrets    map[string]string
	registries map[string]string
	users      map[string]tokens
}

func newData() *data {
	return &data{secrets: map[string]string{}, registries: map[string]string{}, users: map[string]tokens{}}
}

func newUser(login string) *model.User {
	return &model.User{
		ForgeID:       1,
		ForgeRemoteID: model.ForgeRemoteID(login),
		Login:         login,
		Hash:          login + "-hash",
		AccessToken:   login + "-access",
		RefreshToken:  login + "-refresh",
	}
}

// add writes a global secret, a repo secret, a registry and a user, and
// updates an older secret, all named after the phase.
func (d *data) add(t *testing.T, s store.Store, phase string) {
	t.Helper()
	require.NoError(t, s.SecretCreate(&model.Secret{Name: phase, Value: phase + "-value"}))
	d.secrets[phase] = phase + "-value"
	require.NoError(t, s.SecretCreate(&model.Secret{RepoID: 1, Name: phase + "-repo", Value: phase + "-repo-value"}))
	d.secrets[phase+"-repo"] = phase + "-repo-value"

	require.NoError(t, s.RegistryCreate(&model.Registry{Address: phase + ".example.com", Username: "u", Password: phase + "-password"}))
	d.registries[phase+".example.com"] = phase + "-password"

	require.NoError(t, s.CreateUser(newUser(phase)))
	d.users[phase] = tokens{phase + "-access", phase + "-refresh"}

	if len(d.secrets) > 2 {
		secret, err := s.GlobalSecretFind("plain")
		require.NoError(t, err)
		secret.Value = "plain-updated-in-" + phase
		require.NoError(t, s.SecretUpdate(secret))
		d.secrets["plain"] = secret.Value

		user, err := s.GetUserByLogin(1, "plain")
		require.NoError(t, err)
		user.AccessToken = "plain-access-updated-in-" + phase
		require.NoError(t, s.UpdateUser(user))
		d.users["plain"] = tokens{user.AccessToken, user.RefreshToken}
	}
}

// check compares the data read through s with what was written, and the data
// at rest in db with the expected form.
func (d *data) check(t *testing.T, s, db store.Store, encrypted bool) {
	t.Helper()

	secrets, err := s.SecretListAll()
	require.NoError(t, err)
	got := map[string]string{}
	for _, secret := range secrets {
		got[secret.Name] = secret.Value
	}
	assert.Equal(t, d.secrets, got, "secrets")

	registries, err := s.RegistryListAll()
	require.NoError(t, err)
	got = map[string]string{}
	for _, registry := range registries {
		got[registry.Address] = registry.Password
	}
	assert.Equal(t, d.registries, got, "registries")

	users, err := s.GetUserList(&model.ListOptionsWithAll{All: true})
	require.NoError(t, err)
	gotUsers := map[string]tokens{}
	for _, user := range users {
		gotUsers[user.Login] = tokens{user.AccessToken, user.RefreshToken}
	}
	assert.Equal(t, d.users, gotUsers, "users")

	for login := range d.users {
		user, err := s.GetUserByLogin(1, login)
		require.NoError(t, err)
		assert.Equal(t, d.users[login], tokens{user.AccessToken, user.RefreshToken}, "user %s", login)
	}
	secret, err := s.GlobalSecretFind("plain")
	require.NoError(t, err)
	assert.Equal(t, d.secrets["plain"], secret.Value)

	stored := snapshot(t, db)
	for key, value := range stored {
		if key == sampleKey {
			continue
		}
		assert.Equal(t, encrypted, value.keyID != "", "key id of %s", key)
		assert.Equal(t, encrypted, value.value != d.plaintext(key), "stored form of %s", key)
	}
	_, hasSample := stored[sampleKey]
	assert.Equal(t, encrypted, hasSample, "ciphertext sample")
}

// plaintext returns the expected plaintext of a snapshot key.
func (d *data) plaintext(key string) string {
	kind, name, _ := strings.Cut(key, " ")
	switch kind {
	case "secret":
		return d.secrets[name]
	case "registry":
		return d.registries[name]
	case "access-token":
		return d.users[name].access
	case "refresh-token":
		return d.users[name].refresh
	}
	return ""
}

type stored struct{ value, keyID string }

// snapshot returns the values at rest, keyed by kind and name.
func snapshot(t *testing.T, db store.Store) map[string]stored {
	t.Helper()
	values := map[string]stored{}
	secrets, err := db.SecretListAll()
	require.NoError(t, err)
	for _, secret := range secrets {
		values["secret "+secret.Name] = stored{secret.Value, secret.EncryptionKeyID}
	}
	registries, err := db.RegistryListAll()
	require.NoError(t, err)
	for _, registry := range registries {
		values["registry "+registry.Address] = stored{registry.Password, registry.EncryptionKeyID}
	}
	users, err := db.GetUserList(&model.ListOptionsWithAll{All: true})
	require.NoError(t, err)
	for _, user := range users {
		values["access-token "+user.Login] = stored{user.AccessToken, user.EncryptionKeyID}
		values["refresh-token "+user.Login] = stored{user.RefreshToken, user.EncryptionKeyID}
	}
	if sample, err := db.ServerConfigGet(sampleKey); err == nil {
		values[sampleKey] = stored{value: sample}
	}
	return values
}

func requireNoPlaintext(t *testing.T, db store.Store, plaintexts ...string) {
	t.Helper()
	for key, value := range snapshot(t, db) {
		for _, plaintext := range plaintexts {
			assert.NotEqual(t, plaintext, value.value, "plaintext at rest in %s", key)
		}
	}
}

var errCrash = errors.New("server crashed")

// crashingStore fails every write once its budget is used up, like a server
// that dies in the middle of a run.
type crashingStore struct {
	store.Store
	budget int
}

func crashAfter(db store.Store, writes int) *crashingStore {
	return &crashingStore{Store: db, budget: writes}
}

func (c *crashingStore) write() error {
	if c.budget == 0 {
		return errCrash
	}
	c.budget--
	return nil
}

func (c *crashingStore) SecretCreate(secret *model.Secret) error {
	if err := c.write(); err != nil {
		return err
	}
	return c.Store.SecretCreate(secret)
}

func (c *crashingStore) SecretUpdate(secret *model.Secret) error {
	if err := c.write(); err != nil {
		return err
	}
	return c.Store.SecretUpdate(secret)
}

func (c *crashingStore) SecretDelete(secret *model.Secret) error {
	if err := c.write(); err != nil {
		return err
	}
	return c.Store.SecretDelete(secret)
}

func (c *crashingStore) SecretSwapValue(secret *model.Secret, oldValue, oldKeyID string) (bool, error) {
	if err := c.write(); err != nil {
		return false, err
	}
	return c.Store.SecretSwapValue(secret, oldValue, oldKeyID)
}

func (c *crashingStore) RegistryCreate(registry *model.Registry) error {
	if err := c.write(); err != nil {
		return err
	}
	return c.Store.RegistryCreate(registry)
}

func (c *crashingStore) RegistryUpdate(registry *model.Registry) error {
	if err := c.write(); err != nil {
		return err
	}
	return c.Store.RegistryUpdate(registry)
}

func (c *crashingStore) RegistryDelete(registry *model.Registry) error {
	if err := c.write(); err != nil {
		return err
	}
	return c.Store.RegistryDelete(registry)
}

func (c *crashingStore) RegistrySwapPassword(registry *model.Registry, oldPassword, oldKeyID string) (bool, error) {
	if err := c.write(); err != nil {
		return false, err
	}
	return c.Store.RegistrySwapPassword(registry, oldPassword, oldKeyID)
}

func (c *crashingStore) CreateUser(user *model.User) error {
	if err := c.write(); err != nil {
		return err
	}
	return c.Store.CreateUser(user)
}

func (c *crashingStore) UpdateUser(user *model.User) error {
	if err := c.write(); err != nil {
		return err
	}
	return c.Store.UpdateUser(user)
}

func (c *crashingStore) UserSwapTokens(user *model.User, oldAccessToken, oldRefreshToken, oldKeyID string) (bool, error) {
	if err := c.write(); err != nil {
		return false, err
	}
	return c.Store.UserSwapTokens(user, oldAccessToken, oldRefreshToken, oldKeyID)
}

func (c *crashingStore) DeleteUser(user *model.User) error {
	if err := c.write(); err != nil {
		return err
	}
	return c.Store.DeleteUser(user)
}

func (c *crashingStore) ServerConfigSet(key, value string) error {
	if err := c.write(); err != nil {
		return err
	}
	return c.Store.ServerConfigSet(key, value)
}

func (c *crashingStore) ServerConfigDelete(key string) error {
	if err := c.write(); err != nil {
		return err
	}
	return c.Store.ServerConfigDelete(key)
}

// interleavingStore runs before once right after the secrets were listed,
// like another server writing while this one works on that list.
type interleavingStore struct {
	store.Store
	before func()
}

func (i *interleavingStore) SecretListAll() ([]*model.Secret, error) {
	secrets, err := i.Store.SecretListAll()
	if i.before != nil {
		before := i.before
		i.before = nil
		before()
	}
	return secrets, err
}

// tinkKeys is a tink keyset the test rotates like an admin would.
type tinkKeys struct {
	manager *keyset.Manager
}

func newKeyset(t *testing.T) *tinkKeys {
	t.Helper()
	handle, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	require.NoError(t, err)
	return &tinkKeys{manager: keyset.NewManagerFromHandle(handle)}
}

// rotate adds a new primary key and returns the id of the old one.
func (k *tinkKeys) rotate(t *testing.T) uint32 {
	t.Helper()
	handle, err := k.manager.Handle()
	require.NoError(t, err)
	old := handle.KeysetInfo().GetPrimaryKeyId()
	id, err := k.manager.Add(aead.AES256GCMKeyTemplate())
	require.NoError(t, err)
	require.NoError(t, k.manager.SetPrimary(id))
	return old
}

func (k *tinkKeys) remove(t *testing.T, id uint32) {
	t.Helper()
	require.NoError(t, k.manager.Disable(id))
	require.NoError(t, k.manager.Delete(id))
}

func (k *tinkKeys) write(t *testing.T, dir, name string) string {
	t.Helper()
	handle, err := k.manager.Handle()
	require.NoError(t, err)
	path := filepath.Join(dir, name)
	file, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, insecure_clear_text_keyset.Write(handle, keyset.NewJSONWriter(file)))
	require.NoError(t, file.Close())
	return path
}
