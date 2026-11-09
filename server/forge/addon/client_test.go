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

package addon

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

type closeTestForge struct{ forge.Forge }

func (closeTestForge) Name() string { return strconv.Itoa(os.Getpid()) }

func (closeTestForge) Close() error { return nil }

func (closeTestForge) Teams(_ context.Context, u *model.User, _ *model.ListOptions) ([]*model.Team, error) {
	if err := os.WriteFile(u.Login+".ready", nil, 0o600); err != nil {
		return nil, err
	}
	for {
		if _, err := os.Stat(u.Login + ".release"); err == nil {
			return nil, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestMain(m *testing.M) {
	if os.Getenv(HandshakeConfig.MagicCookieKey) == HandshakeConfig.MagicCookieValue {
		Serve(closeTestForge{})
		return
	}
	os.Exit(m.Run())
}

func addonProcessAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func TestCloseStopsAddonProcess(t *testing.T) {
	executable, err := os.Executable()
	require.NoError(t, err)

	f, err := Load(executable)
	require.NoError(t, err)
	pid, err := strconv.Atoi(f.Name())
	require.NoError(t, err)

	require.NoError(t, f.Close())
	require.Eventually(t, func() bool { return !addonProcessAlive(pid) }, 5*time.Second, 20*time.Millisecond)
}

func TestCloseWaitsForInflightAddonRPC(t *testing.T) {
	executable, err := os.Executable()
	require.NoError(t, err)
	f, err := Load(executable)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "addon")
	result := make(chan error, 1)
	go func() {
		_, err := f.Teams(t.Context(), &model.User{Login: path}, nil)
		result <- err
	}()
	require.Eventually(t, func() bool {
		_, err := os.Stat(path + ".ready")
		return err == nil
	}, 5*time.Second, 20*time.Millisecond)

	closed := make(chan error, 1)
	go func() { closed <- f.Close() }()
	select {
	case err := <-closed:
		t.Fatalf("Close returned before the in-flight RPC finished: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	require.NoError(t, os.WriteFile(path+".release", nil, 0o600))
	require.NoError(t, <-result)
	require.NoError(t, <-closed)
}

func TestCallAfterCloseFails(t *testing.T) {
	executable, err := os.Executable()
	require.NoError(t, err)
	f, err := Load(executable)
	require.NoError(t, err)

	require.NoError(t, f.Close())
	require.NoError(t, f.Close())
	_, err = f.Teams(t.Context(), &model.User{Login: filepath.Join(t.TempDir(), "x")}, nil)
	require.ErrorIs(t, err, ErrAddonClosed)
}
