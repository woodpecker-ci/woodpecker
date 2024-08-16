// Copyright 2023 Woodpecker Authors
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

	backend_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

func TestConvertDAGToStages(t *testing.T) {
	steps := map[string]*dagCompilerStep{
		"step1": {
			step:      &backend_types.Step{},
			dependsOn: []string{"step3"},
		},
		"step2": {
			step:      &backend_types.Step{},
			dependsOn: []string{"step1"},
		},
		"step3": {
			step:      &backend_types.Step{},
			dependsOn: []string{"step2"},
		},
	}
	_, err := convertDAGToStages(steps, map[string]*dagCompilerStep{})
	assert.ErrorIs(t, err, &ErrStepDependencyCycle{})

	steps = map[string]*dagCompilerStep{
		"step1": {
			step:  &backend_types.Step{},
			needs: []string{"service1"},
		},
	}
	services := map[string]*dagCompilerStep{
		"service1": {
			step:      &backend_types.Step{},
			dependsOn: []string{"step1"},
		},
	}
	_, err = convertDAGToStages(steps, services)
	assert.ErrorIs(t, err, &ErrStepDependencyCycle{})

	steps = map[string]*dagCompilerStep{
		"step1": {
			step:      &backend_types.Step{},
			dependsOn: []string{"step2"},
		},
		"step2": {
			step: &backend_types.Step{},
		},
	}
	_, err = convertDAGToStages(steps, map[string]*dagCompilerStep{})
	assert.NoError(t, err)

	steps = map[string]*dagCompilerStep{
		"a": {
			step: &backend_types.Step{},
		},
		"b": {
			step:      &backend_types.Step{},
			dependsOn: []string{"a"},
		},
		"c": {
			step:      &backend_types.Step{},
			dependsOn: []string{"a"},
		},
		"d": {
			step:      &backend_types.Step{},
			dependsOn: []string{"b", "c"},
		},
	}
	_, err = convertDAGToStages(steps, map[string]*dagCompilerStep{})
	assert.NoError(t, err)

	steps = map[string]*dagCompilerStep{
		"step1": {
			step:      &backend_types.Step{},
			dependsOn: []string{"not-existing-step"},
		},
	}
	_, err = convertDAGToStages(steps, map[string]*dagCompilerStep{})
	assert.ErrorIs(t, err, &ErrStepMissingDependency{})

	steps = map[string]*dagCompilerStep{
		"step1": {
			step:  &backend_types.Step{},
			needs: []string{"not-existing-service"},
		},
	}
	_, err = convertDAGToStages(steps, map[string]*dagCompilerStep{})
	assert.ErrorIs(t, err, &ErrStepMissingDependency{})

	steps = map[string]*dagCompilerStep{
		"echo env": {
			position: 0,
			name:     "echo env",
			step: &backend_types.Step{
				UUID:  "01HJDPEW6R7J0JBE3F1T7Q0TYX",
				Type:  "commands",
				Name:  "echo env",
				Image: "bash",
			},
		},
		"echo 1": {
			position:  1,
			name:      "echo 1",
			dependsOn: []string{"echo env", "echo 2"},
			step: &backend_types.Step{
				UUID:  "01HJDPF770QGRZER8RF79XVS4M",
				Type:  "commands",
				Name:  "echo 1",
				Image: "bash",
			},
		},
		"echo 2": {
			position: 2,
			name:     "echo 2",
			step: &backend_types.Step{
				UUID:  "01HJDPFF5RMEYZW0YTGR1Y1ZR0",
				Type:  "commands",
				Name:  "echo 2",
				Image: "bash",
			},
		},
	}
	stages, err := convertDAGToStages(steps, map[string]*dagCompilerStep{})
	assert.NoError(t, err)
	assert.EqualValues(t, []*backend_types.Stage{{
		Steps: []*backend_types.Step{{
			UUID:  "01HJDPEW6R7J0JBE3F1T7Q0TYX",
			Type:  "commands",
			Name:  "echo env",
			Image: "bash",
		}, {
			UUID:  "01HJDPFF5RMEYZW0YTGR1Y1ZR0",
			Type:  "commands",
			Name:  "echo 2",
			Image: "bash",
		}},
	}, {
		Steps: []*backend_types.Step{{
			UUID:  "01HJDPF770QGRZER8RF79XVS4M",
			Type:  "commands",
			Name:  "echo 1",
			Image: "bash",
		}},
	}}, stages)

	steps = map[string]*dagCompilerStep{
		"echo env": {
			position: 3,
			name:     "echo env",
			group:    "",
			step: &backend_types.Step{
				Name: "echo env",
			},
		},
		"echo 2": {
			position: 4,
			name:     "echo 2",
			needs:    []string{"service 2"},
			step: &backend_types.Step{
				Name: "echo 2",
			},
		},
	}
	services = map[string]*dagCompilerStep{
		"service 1": {
			position:  0,
			name:      "service 1",
			group:     "",
			dependsOn: []string{"echo env"},
			step: &backend_types.Step{
				Name: "service 1",
			},
		},
		"service 2": {
			position: 1,
			name:     "service 2",
			needs:    []string{"service 1"},
			group:    "",
			step: &backend_types.Step{
				Name: "service 2",
			},
		},
		"service 3": {
			position:  2,
			name:      "service 3",
			needs:     []string{"service 1"},
			dependsOn: []string{"echo env"},
			group:     "",
			step: &backend_types.Step{
				Name: "service 3",
			},
		},
	}
	stages, err = convertDAGToStages(steps, services)
	assert.NoError(t, err)
	assert.EqualValues(t, []*backend_types.Stage{{
		Steps: []*backend_types.Step{{
			Name: "echo env",
		}},
	}, {
		Steps: []*backend_types.Step{{
			Name: "service 1",
		}},
	}, {
		Steps: []*backend_types.Step{{
			Name: "service 2",
		}, {
			Name: "service 3",
		}},
	}, {
		Steps: []*backend_types.Step{{
			Name: "echo 2",
		}},
	}}, stages)

	steps = map[string]*dagCompilerStep{
		"echo env": {
			position: 3,
			name:     "echo env",
			step:     &backend_types.Step{Name: "echo env"},
		},
	}
	services = map[string]*dagCompilerStep{
		"service": {
			name:     "service",
			position: 0,
			step: &backend_types.Step{
				Name: "service",
			},
		},
		"service-depend": {
			name: "service-depend",
			step: &backend_types.Step{
				Name: "service-depend",
			},
			dependsOn: []string{"echo env"},
			position:  1,
		},

		"service-depend-on-service": {
			name: "service-depend-on-service",
			step: &backend_types.Step{
				Name: "service-depend-on-service",
			},
			needs:    []string{"service-depend"},
			position: 2,
		},
	}
	stages, err = convertDAGToStages(steps, services)
	assert.NoError(t, err)
	assert.EqualValues(t, []*backend_types.Stage{{
		Steps: []*backend_types.Step{{
			Name: "service",
		}, {
			Name: "echo env",
		}},
	}, {
		Steps: []*backend_types.Step{{
			Name: "service-depend",
		}},
	}, {
		Steps: []*backend_types.Step{{
			Name: "service-depend-on-service",
		}},
	}}, stages)
}

func TestCompileByDependsOn(t *testing.T) {
	// test without "needs"
	steps := []*dagCompilerStep{
		{
			position: 3,
			name:     "echo env",
			step:     &backend_types.Step{Name: "echo env"},
		},
	}
	services := []*dagCompilerStep{
		{
			name:     "service",
			position: 0,
			step: &backend_types.Step{
				Name: "service",
			},
		},
	}
	stages, err := newDAGCompiler(steps, services).compileByDependsOn()
	assert.NoError(t, err)
	assert.EqualValues(t, []*backend_types.Stage{{
		Steps: []*backend_types.Step{{
			Name: "service",
		}},
	}, {
		Steps: []*backend_types.Step{{
			Name: "echo env",
		}},
	}}, stages)
}

func TestIsDag(t *testing.T) {
	steps := []*dagCompilerStep{
		{
			step: &backend_types.Step{},
		},
	}
	c := newDAGCompiler(steps, []*dagCompilerStep{})
	assert.False(t, c.isDAG())

	steps = []*dagCompilerStep{
		{
			step:      &backend_types.Step{},
			dependsOn: []string{},
		},
	}
	c = newDAGCompiler(steps, []*dagCompilerStep{})
	assert.True(t, c.isDAG())
}
