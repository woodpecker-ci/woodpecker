// Copyright 2026 Julian Ospald
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

package libvirt

import (
	"github.com/go-viper/mapstructure/v2"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// BackendOptions defines all the advanced options for the libvirt backend.
type BackendOptions struct {
	SSHConfig  SSHConfig  `mapstructure:"ssh_config"`
	SharedDisk SharedDisk `mapstructure:"shared_disk"`
	Persistent bool       `mapstructure:"persistent"`
}

type SSHConfig struct {
	User           string `mapstructure:"username"`
	GuestInterface string `mapstructure:"guest_interface"`
	// takes a duration string, such as "2m"
	Timeout string `mapstructure:"timeout"`
}

type SharedDisk struct {
	Disk string `mapstructure:"disk"`
	UUID string `mapstructure:"uuid"`
}

func parseBackendOptions(step *backend_types.Step) (BackendOptions, error) {
	var result BackendOptions
	if step == nil || step.BackendOptions == nil {
		return result, nil
	}
	err := mapstructure.WeakDecode(step.BackendOptions[EngineName], &result)
	return result, err
}
