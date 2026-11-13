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

func TestWorkspaceVolumeAlwaysAdded(t *testing.T) {
	compiler := New(
		WithPrefix("test"),
		WithWorkspace("/woodpecker", "src"),
		WithLocal(true),
	)

	workflow := &yaml_types.Workflow{
		SkipClone: true,
		Steps: yaml_types.ContainerList{
			ContainerList: []*yaml_types.Container{{
				Name:     "test",
				Image:    "alpine",
				Commands: []string{"echo hello"},
			}},
		},
	}

	config, err := compiler.Compile(workflow)
	assert.NoError(t, err)

	assert.Equal(t, "test_default", config.Volume)

	assert.NotEmpty(t, config.Stages, "expected at least one stage")
	assert.NotEmpty(t, config.Stages[0].Steps, "expected at least one step")
	assert.Contains(t, config.Stages[0].Steps[0].Volumes, "test_default:/woodpecker",
		"workspace volume must always be added, even in local mode")
}

func codes(c ...int) yaml_types.SignalConfig { return yaml_types.SignalConfig{ExitCodes: c} }

func TestConvertSignals(t *testing.T) {
	sig := convertSignals

	t.Run("empty", func(t *testing.T) {
		got, err := sig(nil)
		assert.NoError(t, err)
		assert.Nil(t, got.exitCodeSignals)
		assert.Equal(t, 0, got.maxRestarts)
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
		})
		assert.NoError(t, err)
		assert.Equal(t, map[int]backend_types.StepSignal{
			78: backend_types.StepSignalSkip,
			1:  backend_types.StepSignalSuccess,
			2:  backend_types.StepSignalSuccess,
			75: backend_types.StepSignalRestart,
			70: backend_types.StepSignalCancel,
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
}
