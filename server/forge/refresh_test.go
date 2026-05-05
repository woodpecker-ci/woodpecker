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

package forge_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

// refresherForge combines MockForge (satisfies forge.Forge) and MockRefresher
// (satisfies forge.Refresher) so the Refresh function's type assertion succeeds.
type refresherForge struct {
	*forge_mocks.MockForge
	*forge_mocks.MockRefresher
}

func expiredUser(id int64) *model.User {
	return &model.User{
		ID:           id,
		Login:        fmt.Sprintf("user%d", id),
		AccessToken:  "old-access-token",
		RefreshToken: "old-refresh-token",
		Expiry:       time.Now().UTC().Unix() - 100, // expired
	}
}

func freshUser(id int64) *model.User {
	return &model.User{
		ID:           id,
		Login:        fmt.Sprintf("user%d", id),
		AccessToken:  "valid-access-token",
		RefreshToken: "valid-refresh-token",
		Expiry:       time.Now().UTC().Unix() + 7200, // 2 hours from now
	}
}

func TestRefresh_NonExpiredToken(t *testing.T) {
	mockForge := forge_mocks.NewMockForge(t)
	mockRefresher := forge_mocks.NewMockRefresher(t)
	mockStore := store_mocks.NewMockStore(t)

	f := &refresherForge{MockForge: mockForge, MockRefresher: mockRefresher}
	user := freshUser(1)

	forge.Refresh(context.Background(), f, mockStore, user)

	// Refresher.Refresh should NOT be called since token is still valid
	mockRefresher.AssertNotCalled(t, "Refresh", mock.Anything, mock.Anything)
}

func TestRefresh_ExpiredToken(t *testing.T) {
	mockForge := forge_mocks.NewMockForge(t)
	mockRefresher := forge_mocks.NewMockRefresher(t)
	mockStore := store_mocks.NewMockStore(t)

	f := &refresherForge{MockForge: mockForge, MockRefresher: mockRefresher}
	user := expiredUser(1)

	mockRefresher.On("Refresh", mock.Anything, user).Return(true, nil).Run(func(args mock.Arguments) {
		u, ok := args.Get(1).(*model.User)
		if !ok {
			return
		}
		u.AccessToken = "new-access-token"
		u.RefreshToken = "new-refresh-token"
		u.Expiry = time.Now().UTC().Unix() + 3600
	})
	mockStore.On("UpdateUser", user).Return(nil)

	forge.Refresh(context.Background(), f, mockStore, user)

	assert.Equal(t, "new-access-token", user.AccessToken)
	assert.Equal(t, "new-refresh-token", user.RefreshToken)
	mockRefresher.AssertCalled(t, "Refresh", mock.Anything, user)
	mockStore.AssertCalled(t, "UpdateUser", user)
}

func TestRefresh_ExpiredTokenNoUpdate(t *testing.T) {
	mockForge := forge_mocks.NewMockForge(t)
	mockRefresher := forge_mocks.NewMockRefresher(t)
	mockStore := store_mocks.NewMockStore(t)

	f := &refresherForge{MockForge: mockForge, MockRefresher: mockRefresher}
	user := expiredUser(2)

	// Refresh returns false (no update needed), e.g. token was already refreshed
	mockRefresher.On("Refresh", mock.Anything, user).Return(false, nil)

	forge.Refresh(context.Background(), f, mockStore, user)

	mockRefresher.AssertCalled(t, "Refresh", mock.Anything, user)
	// UpdateUser should NOT be called when Refresh returns false
	mockStore.AssertNotCalled(t, "UpdateUser", mock.Anything)
}

func TestRefresh_ConcurrentRefreshSerialized(t *testing.T) {
	mockForge := forge_mocks.NewMockForge(t)
	mockRefresher := forge_mocks.NewMockRefresher(t)
	mockStore := store_mocks.NewMockStore(t)

	f := &refresherForge{MockForge: mockForge, MockRefresher: mockRefresher}

	var refreshCount atomic.Int32

	mockRefresher.On("Refresh", mock.Anything, mock.Anything).Return(true, nil).Run(func(args mock.Arguments) {
		refreshCount.Add(1)
		// Simulate network latency so concurrent callers overlap
		time.Sleep(50 * time.Millisecond)
		u, ok := args.Get(1).(*model.User)
		if !ok {
			return
		}
		u.AccessToken = "new-access-token"
		u.RefreshToken = "new-refresh-token"
		u.Expiry = time.Now().UTC().Unix() + 3600
	})
	mockStore.On("UpdateUser", mock.Anything).Return(nil)

	const numGoroutines = 10
	var wg sync.WaitGroup
	users := make([]*model.User, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		users[i] = expiredUser(42) // same user ID
	}

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(u *model.User) {
			defer wg.Done()
			forge.Refresh(context.Background(), f, mockStore, u)
		}(users[i])
	}
	wg.Wait()

	// Only one actual refresh call should have been made
	assert.Equal(t, int32(1), refreshCount.Load(), "expected exactly 1 refresh call, got %d", refreshCount.Load())

	// All goroutines should have the fresh tokens
	for i := 0; i < len(users); i++ {
		assert.Equal(t, "new-access-token", users[i].AccessToken, "user[%d] missing new access token", i)
		assert.Equal(t, "new-refresh-token", users[i].RefreshToken, "user[%d] missing new refresh token", i)
	}
}

func TestRefresh_ConcurrentRefreshError(t *testing.T) {
	mockForge := forge_mocks.NewMockForge(t)
	mockRefresher := forge_mocks.NewMockRefresher(t)
	mockStore := store_mocks.NewMockStore(t)

	f := &refresherForge{MockForge: mockForge, MockRefresher: mockRefresher}

	mockRefresher.On("Refresh", mock.Anything, mock.Anything).Return(false, fmt.Errorf("token was already used")).Run(func(_ mock.Arguments) {
		time.Sleep(50 * time.Millisecond)
	})

	const numGoroutines = 5
	var wg sync.WaitGroup
	users := make([]*model.User, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		users[i] = expiredUser(99) // same user ID
	}

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(u *model.User) {
			defer wg.Done()
			forge.Refresh(context.Background(), f, mockStore, u)
		}(users[i])
	}
	wg.Wait()

	// Tokens should remain unchanged (error path)
	for i := 0; i < len(users); i++ {
		assert.Equal(t, "old-access-token", users[i].AccessToken, "user[%d] token should be unchanged after error", i)
	}

	// Store.UpdateUser should NOT be called on error
	mockStore.AssertNotCalled(t, "UpdateUser", mock.Anything)
}

func TestRefresh_NonRefresherForge(t *testing.T) {
	// MockForge does NOT implement Refresher, so the type assertion should fail
	// and Refresh should be a no-op
	mockForge := forge_mocks.NewMockForge(t)
	mockStore := store_mocks.NewMockStore(t)

	user := expiredUser(1)

	forge.Refresh(context.Background(), mockForge, mockStore, user)

	// Token should be unchanged
	assert.Equal(t, "old-access-token", user.AccessToken)
	mockStore.AssertNotCalled(t, "UpdateUser", mock.Anything)
}
