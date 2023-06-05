package linter

import (
	"fmt"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

const (
	blockClone uint8 = iota
	blockPipeline
	blockServices
)

// A Linter lints a pipeline configuration.
type Linter struct {
	trusted bool
}

// New creates a new Linter with options.
func New(opts ...Option) *Linter {
	linter := new(Linter)
	for _, opt := range opts {
		opt(linter)
	}
	return linter
}

// Lint lints the configuration.
func (l *Linter) Lint(c *types.Workflow) error {
	if len(c.Steps.ContainerList) == 0 {
		return fmt.Errorf("Invalid or missing pipeline section")
	}
	if err := l.lint(c.Clone.ContainerList, blockClone); err != nil {
		return err
	}
	if err := l.lint(c.Steps.ContainerList, blockPipeline); err != nil {
		return err
	}
	return l.lint(c.Services.ContainerList, blockServices)
}

func (l *Linter) lint(containers []*types.Container, _ uint8) error {
	for _, container := range containers {
		if err := l.lintImage(container); err != nil {
			return err
		}
		if !l.trusted {
			if err := l.lintTrusted(container); err != nil {
				return err
			}
		}
		if err := l.lintCommands(container); err != nil {
			return err
		}
	}
	return nil
}

func (l *Linter) lintImage(c *types.Container) error {
	if len(c.Image) == 0 {
		return fmt.Errorf("Invalid or missing image")
	}
	return nil
}

func (l *Linter) lintCommands(c *types.Container) error {
	if len(c.Commands) == 0 {
		return nil
	}
	if len(c.Settings) != 0 {
		var keys []string
		for key := range c.Settings {
			keys = append(keys, key)
		}
		return fmt.Errorf("Cannot configure both commands and custom attributes %v", keys)
	}
	return nil
}

func (l *Linter) lintTrusted(c *types.Container) error {
	if c.Privileged {
		return fmt.Errorf("Insufficient privileges to use privileged mode")
	}
	if c.ShmSize != 0 {
		return fmt.Errorf("Insufficient privileges to override shm_size")
	}
	if len(c.DNS) != 0 {
		return fmt.Errorf("Insufficient privileges to use custom dns")
	}
	if len(c.DNSSearch) != 0 {
		return fmt.Errorf("Insufficient privileges to use dns_search")
	}
	if len(c.Devices) != 0 {
		return fmt.Errorf("Insufficient privileges to use devices")
	}
	if len(c.ExtraHosts) != 0 {
		return fmt.Errorf("Insufficient privileges to use extra_hosts")
	}
	if len(c.NetworkMode) != 0 {
		return fmt.Errorf("Insufficient privileges to use network_mode")
	}
	if len(c.IpcMode) != 0 {
		return fmt.Errorf("Insufficient privileges to use ipc_mode")
	}
	if len(c.Sysctls) != 0 {
		return fmt.Errorf("Insufficient privileges to use sysctls")
	}
	if c.Networks.Networks != nil && len(c.Networks.Networks) != 0 {
		return fmt.Errorf("Insufficient privileges to use networks")
	}
	if c.Volumes.Volumes != nil && len(c.Volumes.Volumes) != 0 {
		return fmt.Errorf("Insufficient privileges to use volumes")
	}
	if len(c.Tmpfs) != 0 {
		return fmt.Errorf("Insufficient privileges to use tmpfs")
	}
	return nil
}
