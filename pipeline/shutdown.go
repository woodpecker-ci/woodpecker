// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pipeline

import (
	"context"
	"sync"
	"time"
)

const shutdownTimeout = time.Second * 5

var (
	shutdownCtx       context.Context
	shutdownCtxCancel context.CancelFunc
	shutdownCtxLock   sync.Mutex
)

func GetShutdownCtx() context.Context {
	shutdownCtxLock.Lock()
	defer shutdownCtxLock.Unlock()
	if shutdownCtx == nil {
		shutdownCtx, shutdownCtxCancel = context.WithTimeout(context.Background(), shutdownTimeout)
	}
	return shutdownCtx
}

func CancelShutdown() {
	shutdownCtxLock.Lock()
	defer shutdownCtxLock.Unlock()
	if shutdownCtxCancel == nil {
		// we create an canceled context
		shutdownCtx, shutdownCtxCancel = context.WithCancel(context.Background()) //nolint:forbidigo
	}
	shutdownCtxCancel()
}
