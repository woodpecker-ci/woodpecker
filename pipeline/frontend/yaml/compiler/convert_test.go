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

package compiler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	yaml_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/types"
)

func TestConvertPortNumber(t *testing.T) {
	portDef := "1234"
	actualPort, err := convertPort(portDef)
	assert.NoError(t, err)
	assert.Equal(t, backend_types.Port{
		Number:   1234,
		Protocol: "",
	}, actualPort)
}

func TestConvertPortUdp(t *testing.T) {
	portDef := "1234/udp"
	actualPort, err := convertPort(portDef)
	assert.NoError(t, err)
	assert.Equal(t, backend_types.Port{
		Number:   1234,
		Protocol: "udp",
	}, actualPort)
}

func TestConvertPortWrongOrder(t *testing.T) {
	portDef := "tcp/1234"
	_, err := convertPort(portDef)
	assert.Error(t, err)
}

func TestConvertPortWrongDelimiter(t *testing.T) {
	portDef := "1234|udp"
	_, err := convertPort(portDef)
	assert.Error(t, err)
}

func TestConvertPortWrong(t *testing.T) {
	portDef := "http"
	_, err := convertPort(portDef)
	assert.Error(t, err)
}

func codes(c ...int) yaml_types.SignalConfig { return yaml_types.SignalConfig{ExitCodes: c} }

func TestConvertSignals(t *testing.T) {
	sig := func(m map[string]yaml_types.SignalConfig) (*signalConfig, error) { return convertSignals(m) }

	t.Run("empty", func(t *testing.T) {
		got, err := sig(nil)
		assert.NoError(t, err)
		assert.Nil(t, got.exitCodeSignals)
		assert.Equal(t, 0, got.maxRestarts)
		assert.Equal(t, time.Duration(0), got.pauseTimeout)
	})

	t.Run("single and list", func(t *testing.T) {
		got, err := sig(map[string]yaml_types.SignalConfig{"skip": codes(78, 79)})
		assert.NoError(t, err)
		assert.Equal(t, map[int]backend_types.StepSignal{
			78: backend_types.StepSignalSkip,
			79: backend_types.StepSignalSkip,
		}, got.exitCodeSignals)
	})

	t.Run("several signals", func(t *testing.T) {
		got, err := sig(map[string]yaml_types.SignalConfig{
			"skip":    codes(78),
			"success": codes(1, 2),
			"restart": codes(75),
			"cancel":  codes(70),
			"pause":   codes(71),
		})
		assert.NoError(t, err)
		assert.Equal(t, map[int]backend_types.StepSignal{
			78: backend_types.StepSignalSkip,
			1:  backend_types.StepSignalSuccess,
			2:  backend_types.StepSignalSuccess,
			75: backend_types.StepSignalRestart,
			70: backend_types.StepSignalCancel,
			71: backend_types.StepSignalPause,
		}, got.exitCodeSignals)
	})

	t.Run("same code for two signals", func(t *testing.T) {
		_, err := sig(map[string]yaml_types.SignalConfig{"skip": codes(78), "success": codes(78)})
		assert.ErrorContains(t, err, `exit code 78 is mapped to both signal "skip" and "success"`)
	})

	t.Run("unknown signal", func(t *testing.T) {
		_, err := sig(map[string]yaml_types.SignalConfig{"pause_forever": codes(78)})
		assert.ErrorContains(t, err, `unknown signal "pause_forever"`)
	})

	t.Run("out of range", func(t *testing.T) {
		for _, code := range []int{0, -1, 256} {
			_, err := sig(map[string]yaml_types.SignalConfig{"skip": codes(code)})
			assert.ErrorContains(t, err, "out of range", "code %d", code)
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		_, err := sig(map[string]yaml_types.SignalConfig{"skip": codes(78, 78)})
		assert.ErrorContains(t, err, "exit code 78 is mapped to both")
	})

	t.Run("max_restarts", func(t *testing.T) {
		five := 5
		got, err := sig(map[string]yaml_types.SignalConfig{"restart": {ExitCodes: []int{75}, MaxRestarts: &five}})
		assert.NoError(t, err)
		assert.Equal(t, 5, got.maxRestarts)
	})

	t.Run("max_restarts out of range", func(t *testing.T) {
		for _, v := range []int{0, -1, 11} {
			_, err := sig(map[string]yaml_types.SignalConfig{"restart": {ExitCodes: []int{75}, MaxRestarts: &v}})
			assert.ErrorContains(t, err, "max_restarts", "value %d", v)
		}
	})

	t.Run("max_restarts on other signal", func(t *testing.T) {
		two := 2
		_, err := sig(map[string]yaml_types.SignalConfig{"skip": {ExitCodes: []int{78}, MaxRestarts: &two}})
		assert.ErrorContains(t, err, "only allowed for the restart signal")
	})

	t.Run("pause timeout decline by default", func(t *testing.T) {
		got, err := sig(map[string]yaml_types.SignalConfig{"pause": {ExitCodes: []int{71}, Timeout: "30m"}})
		assert.NoError(t, err)
		assert.Equal(t, 30*time.Minute, got.pauseTimeout)
		assert.False(t, got.pauseTimeoutResume)
	})

	t.Run("pause timeout resume", func(t *testing.T) {
		got, err := sig(map[string]yaml_types.SignalConfig{"pause": {ExitCodes: []int{71}, Timeout: "1h30m", OnTimeout: "resume"}})
		assert.NoError(t, err)
		assert.Equal(t, 90*time.Minute, got.pauseTimeout)
		assert.True(t, got.pauseTimeoutResume)
	})

	t.Run("pause timeout invalid", func(t *testing.T) {
		for _, cfg := range []yaml_types.SignalConfig{
			{ExitCodes: []int{71}, Timeout: "soon"},
			{ExitCodes: []int{71}, Timeout: "-5m"},
			{ExitCodes: []int{71}, Timeout: "0s"},
			{ExitCodes: []int{71}, OnTimeout: "resume"},
			{ExitCodes: []int{71}, Timeout: "5m", OnTimeout: "approve"},
		} {
			_, err := sig(map[string]yaml_types.SignalConfig{"pause": cfg})
			assert.Error(t, err, "%+v", cfg)
		}
	})

	t.Run("timeout on other signal", func(t *testing.T) {
		_, err := sig(map[string]yaml_types.SignalConfig{"skip": {ExitCodes: []int{78}, Timeout: "5m"}})
		assert.ErrorContains(t, err, "only allowed for the pause signal")
	})
}
