// Copyright 2023 Woodpecker Authors
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

package local

import (
	"errors"
	"fmt"
)

// notAllowedEnvVarOverwrites are all env vars that can not be overwritten by step config
var notAllowedEnvVarOverwrites = []string{
	"CI_NETRC_MACHINE",
	"CI_NETRC_USERNAME",
	"CI_NETRC_PASSWORD",
	"CI_SCRIPT",
	"HOME",
	"SHELL",
}

var (
	ErrUnsupportedStepType   = errors.New("unsupported step type")
	ErrWorkflowStateNotFound = errors.New("workflow state not found")
)

const netrcFile = `
machine %s
login %s
password %s
`

func genNetRC(env map[string]string) string {
	return fmt.Sprintf(
		netrcFile,
		env["CI_NETRC_MACHINE"],
		env["CI_NETRC_USERNAME"],
		env["CI_NETRC_PASSWORD"],
	)
}
