package libvirt

import (
	"github.com/go-viper/mapstructure/v2"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// BackendOptions defines all the advanced options for the libvirt backend.
type BackendOptions struct {
	SSHConfig  SSHConfig  `mapstructure:"ssh_config"`
	SharedDisk SharedDisk `mapstructure:"shared_disk"`
	Ephemeral  bool       `mapstructure:"ephemeral"`
}

type SSHConfig struct {
	User           string `mapstructure:"username"`
	Hostname       string `mapstructure:"hostname"` // preferred
	GuestInterface string `mapstructure:"guest_interface"`
	// takes a duration string, such as "2m"
	Timeout string `mapstructure:"timeout"`
}

type SharedDisk struct {
	DiskConfig string `mapstructure:"disk_config"`
	UUID       string `mapstructure:"uuid"` // preferred, required on windows
}

func parseBackendOptions(step *backend_types.Step) (BackendOptions, error) {
	var result BackendOptions
	if step == nil || step.BackendOptions == nil {
		return result, nil
	}
	err := mapstructure.WeakDecode(step.BackendOptions[EngineName], &result)
	return result, err
}
