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

package cron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/remote/mocks"
)

func TestCreateBuild(t *testing.T) {
	rOld := server.Config.Services.Remote
	defer func() {
		server.Config.Services.Remote = rOld
	}()
	server.Config.Services.Remote = mocks.NewRemote(t)

	// TODO: mockStore
	// createBuild(context.TODO(), &model.Cron{}, mockStore)
}

func TestCalcNewNext(t *testing.T) {
	now := time.Unix(1661962369, 0)
	_, err := CalcNewNext("", now)
	assert.Error(t, err)

	schedule, err := CalcNewNext("@every 5m", now)
	assert.NoError(t, err)
	assert.EqualValues(t, 1661962669, schedule.Unix())
}
