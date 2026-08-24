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
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"darvaza.org/x/sync/mutex"
	"github.com/jakobii/syncx"

	"github.com/cenkalti/backoff/v7"
	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"golang.org/x/crypto/ssh"
	"golang.org/x/text/encoding/unicode"
	virt "libvirt.org/go/libvirt"

	"github.com/melbahja/goph"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

var (
	ErrUnsupportedStepType = errors.New("unsupported step type")
	ErrUnsupportedCloning  = errors.New("Automatic cloning is not supported in libvirt. Set 'skip_clone: true'")
)

type libvirt struct {
	conn        *virt.Connect
	commands    sync.Map
	domains     sync.Map
	pipesStdOut sync.Map
	doneChannel sync.Map
	guestOS     sync.Map
	stepLock    syncx.Mutex
	c           *cli.Command
	os, arch    string
}

type pipes struct {
	pr *nio.PipeReader
	pw *nio.PipeWriter
}

const (
	EngineName            = "libvirt"
	baseImagePattern      = "base_%s_%s_%s%s"
	sharedDiskPattern     = "shared_%s%s"
	defaultMaxTimeout     = "2m"
	defaultGuestInterface = "eth0"
)

func New() backend_types.Backend {
	return &libvirt{
		conn:        nil,
		commands:    sync.Map{},
		domains:     sync.Map{},
		pipesStdOut: sync.Map{},
		doneChannel: sync.Map{},
		guestOS:     sync.Map{},
		stepLock:    syncx.Mutex{},
		c:           nil,
		os:          runtime.GOOS,
		arch:        runtime.GOARCH,
	}
}

func (e *libvirt) Name() string {
	return EngineName
}

func ConnectToLibvirt(ctx context.Context) (*virt.Connect, error) {
	c, ok := ctx.Value(backend_types.CliCommand).(*cli.Command)
	if !ok {
		return nil, backend_types.ErrNoCliContextFound
	}
	libvirtUri := c.String("backend-libvirt-uri")

	return virt.NewConnect(libvirtUri)

}

// it's hard to figure out just the socket, so we attempt
// a basic connection and close it immediately
func (e *libvirt) IsAvailable(ctx context.Context) bool {
	conn, err := ConnectToLibvirt(ctx)
	if err != nil {
		return false
	} else {
		conn.Close()
		return true
	}
}

func (e *libvirt) Flags() []cli.Flag {
	return Flags
}

func (e *libvirt) Load(ctx context.Context) (*backend_types.BackendInfo, error) {
	c, ok := ctx.Value(backend_types.CliCommand).(*cli.Command)
	if !ok {
		return nil, backend_types.ErrNoCliContextFound
	}
	e.c = c

	conn, err := ConnectToLibvirt(ctx)
	if err != nil {
		return nil, err
	}
	e.conn = conn

	return &backend_types.BackendInfo{
		Platform: e.os + "/" + e.arch,
	}, nil
}

func (e *libvirt) SetupWorkflow(ctx context.Context, conf *backend_types.Config, taskUUID string) error {
	return nil
}

func (e *libvirt) StartStep(ctx context.Context, step *backend_types.Step, taskUUID string) (ret_err error) {
	switch step.Type {
	case types.StepTypeClone:
		return ErrUnsupportedCloning
	case types.StepTypeCommands:
	case types.StepTypePlugin:
		return ErrUnsupportedStepType
	default:
		return ErrUnsupportedStepType
	}

	log.Debug().Msgf("Step %s is waiting for lock", step.Name)

	ok, err := mutex.SafeLockContext(ctx, &e.stepLock)
	if !ok {
		log.Debug().Msgf("Failed to acquire lock: %s", err.Error())
		return err
	}
	log.Debug().Msgf("Step %s acquired lock", step.Name)

	defer func() {
		if ret_err != nil {
			mutex.SafeUnlock(&e.stepLock)
		}
	}()

	options, err := parseBackendOptions(step)
	if err != nil {
		log.Error().Err(err).Msg("could not parse backend options")
	}

	domain, domainName, uuid, guestOS, err := e.LoadDomain(ctx, step.Image, step.Environment, options.Persistent, options.SharedDisk, taskUUID, step.UUID)
	if err != nil {
		return err
	}

	e.domains.Store(step.UUID, domain)
	e.guestOS.Store(step.UUID, guestOS)

	scriptIn, entry, err := GenerateSSHConf(step, guestOS)

	if err != nil {
		return err
	}
	cmd := entry[0]
	args := entry[1:]

	// get the timeout duration for SSH
	var maxBackOff time.Duration
	if options.SSHConfig.Timeout != "" {
		maxBackOff, err = time.ParseDuration(options.SSHConfig.Timeout)
		if err != nil {
			return err
		}
	} else {
		// default
		maxBackOff, _ = time.ParseDuration(defaultMaxTimeout)
	}

	var sshAddr string

	if options.SSHConfig.GuestInterface != "" {
		// Inspect the guest interface for the IP
		ip, err := backoff.Retry(ctx, func() (*virt.DomainIPAddress, error) {
			ip, err := GetDomainIP(ctx, domain, options.SSHConfig.GuestInterface)
			if err != nil {
				log.Debug().Msg("Retrying...")
				return nil, err
			}
			return ip, nil
		}, backoff.WithMaxElapsedTime(maxBackOff))
		if err != nil {
			return err
		}
		sshAddr = ip.Addr
	} else {
		// we rely on libvirt NSS module
		sshAddr = domainName
	}

	sshPw, ok := step.Environment["LIBVIRT_SSH_PW"]
	if !ok {
		return fmt.Errorf("No SSH password set for libvirt guest. Set LIBVIRT_SSH_PW via a secret.")
	}

	log.Debug().Msgf("Connecting to %s as user %s", sshAddr, options.SSHConfig.User)
	client, err := backoff.Retry(ctx, func() (*goph.Client, error) {
		return goph.NewConn(&goph.Config{
			User:     options.SSHConfig.User,
			Auth:     goph.Password(sshPw),
			Addr:     sshAddr,
			Port:     22,
			Timeout:  goph.DefaultTimeout,
			Callback: ssh.InsecureIgnoreHostKey(),
		})
	}, backoff.WithMaxElapsedTime(maxBackOff))

	if err != nil {
		return err
	}

	// mount the volume
	if guestOS == "windows" {
		encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
		cmd := "powershell"
		exec := fmt.Sprintf("$mount = '%s' ; $ErrorActionPreference = 'Stop' ; New-Item -ItemType Directory -Path $mount ; $disk = Get-Disk | Where-Object { $_.SerialNumber -eq '%s' } ; $partition = Get-Partition -DiskNumber $disk.DiskNumber | Where-Object { $_.Type -in 'Basic', 'IFS' } ; Add-PartitionAccessPath -DiskNumber $disk.DiskNumber -PartitionNumber $partition.PartitionNumber -AccessPath $mount", step.WorkspaceBase, uuid)
		utf16leBytes, err := encoder.Bytes([]byte(exec))
		if err != nil {
			return err
		}

		args := []string{"-noprofile", "-noninteractive", "-encodedcommand", base64.StdEncoding.EncodeToString(utf16leBytes)}
		err = AdhocSSH(ctx, client, cmd, args, taskUUID, step.UUID)

		if err != nil {
			return err
		}
	} else { // unix guest
		var out string
		// first figure out if we have 'sudo' or 'doas', so we can have the right privilege escalation mechanism
		{
			cmd := "/bin/sh"
			args := []string{"-c", "'if command -v sudo >/dev/null 2>&1 ; then echo -n sudo ; elif command -v doas >/dev/null 2>&1 ; then echo -n doas ; else exit 0 ; fi'"}
			sshCmd, err := client.CommandContext(ctx, cmd, args...)
			if err != nil {
				return err
			}
			b, err := sshCmd.Output()
			if err != nil {
				return fmt.Errorf("Failed to check for sudo: %s", err)
			}
			out = string(b)
		}

		mountSource := fmt.Sprintf("UUID=%s", uuid)

		var command []string
		mntCmds := []string{"mount", "-o", "X-mount.mkdir", mountSource, step.WorkspaceBase}
		switch out {
		case "sudo":
			command = append([]string{"sudo"}, mntCmds...)
		case "doas":
			command = append([]string{"doas"}, mntCmds...)
		case "":
			log.Debug().Msgf("Running as root")
			command = mntCmds
		default:
			return fmt.Errorf("Unexpected privilege escalation command: %s", out)
		}

		// execute command
		err = AdhocSSH(ctx, client, command[0], command[1:], taskUUID, step.UUID)

		if err != nil {
			return fmt.Errorf("Failed to run '%s' with args '%s': %s", cmd, args, err)
		}
	}

	sshCmd, err := client.Command(cmd, args...)
	if err != nil {
		return err
	}

	e.commands.Store(step.UUID, sshCmd)

	// we need to create pipes and set Stdout here
	// in TailStep it is potentially too late
	pr, pw := nio.Pipe(buffer.New(1024 * 1024))
	sshCmd.Stdout = nil // see comment below
	sshCmd.Stderr = nil

	var wg sync.WaitGroup

	wg.Add(1)
	stdoutR, err := sshCmd.StdoutPipe()
	if err != nil {
		return err
	}
	wg.Add(1)
	stderrR, err := sshCmd.StderrPipe()
	if err != nil {
		return err
	}
	stdinWC, err := sshCmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		// We need to close the write end ourselves on EOF from the server
		// otherwise the read side of our pipe never drains and Wait() deadlocks.
		// If we assign 'sshCmd.Stdout = pw', the library internals only copy
		// but never close pw, because they are oblivious to the fact that it's
		// a pipe. So we need to implement the copy+close ourselves. There is no
		// EOF callback.
		io.Copy(pw, stdoutR)
		wg.Done()
	}()

	go func() {
		io.Copy(pw, stderrR)
		wg.Done()
	}()

	go func() {
		wg.Wait()
		pw.Close()
		log.Debug().Msg("Closing write pipe end")
	}()

	e.pipesStdOut.Store(step.UUID, &pipes{pr, pw})

	done := make(chan struct{})
	e.doneChannel.Store(step.UUID, done)
	// and a go routine that watches the ctx and then triggers a signal
	go func() {
		select {
		case <-ctx.Done():
			err := e.TerminateSshCommand(options, client, sshCmd, guestOS, taskUUID, step.UUID)
			if err != nil {
				log.Debug().Msgf("Failed to terminate SSH command gracefully. Closing session by force. Error was: %s", err)
			}
			sshCmd.Close()
		case <-done:
		}
		time.Sleep(time.Second * 5)
		log.Debug().Msg("Closing read pipe end")
		_ = pr.Close()
	}()

	err = sshCmd.Start()
	if err != nil {
		return err
	}

	// now feed the actual command to stdin
	go func() {
		_, err := stdinWC.Write([]byte(scriptIn))
		if err != nil {
			log.Debug().Msgf("Failed to write cmd to stdin: %s", err)
		}
		stdinWC.Close()
		log.Debug().Msg("Closing stdin")
	}()

	return nil
}

func (e *libvirt) WaitStep(ctx context.Context, step *backend_types.Step, taskUUID string) (b_state *backend_types.State, ret_err error) {
	switch step.Type {
	case types.StepTypeClone:
		return nil, ErrUnsupportedCloning
	case types.StepTypeCommands:
	case types.StepTypePlugin:
		return nil, ErrUnsupportedStepType
	default:
		return nil, ErrUnsupportedStepType
	}

	defer func() {
		mutex.SafeUnlock(&e.stepLock)
	}()

	sshCmd, ok := e.commands.Load(step.UUID)
	if !ok {
		return nil, fmt.Errorf("Could not find key %s for commands", step.UUID)
	}

	done, ok := e.doneChannel.Load(step.UUID)
	if !ok {
		return nil, fmt.Errorf("Could not find key %s for doneChannel", step.UUID)
	}

	sshErr := sshCmd.(*goph.Cmd).Wait()
	close(done.(chan struct{}))

	domain, ok := e.domains.Load(step.UUID)
	if !ok {
		return nil, fmt.Errorf("Could not find key %s for domains", step.UUID)
	}

	// we need to shut down the VM here, because DestroyStep is less reliable
	// in terms of timing and can lead to deadlocks when two steps are trying
	// to acquire a lock
	err := ShutdownVM(ctx, domain.(*virt.Domain))
	if err != nil {
		return nil, err
	}

	switch e := sshErr.(type) {
	case *ssh.ExitMissingError:
		return &backend_types.State{
			Exited:    true,
			ExitCode:  -1,
			OOMKilled: false,
		}, sshErr
	case *ssh.ExitError:
		return &backend_types.State{
			Exited:    true,
			ExitCode:  e.ExitStatus(),
			OOMKilled: false,
		}, sshErr
	case nil:
		return &backend_types.State{
			Exited:    true,
			ExitCode:  0,
			OOMKilled: false,
		}, sshErr
	default:
		return nil, sshErr
	}
}

type closerFunc func() error

func (f closerFunc) Close() error { return f() }

func (e *libvirt) TailStep(ctx context.Context, step *backend_types.Step, taskUUID string) (io.ReadCloser, error) {
	switch step.Type {
	case types.StepTypeClone:
		return nil, ErrUnsupportedCloning
	case types.StepTypeCommands:
	case types.StepTypePlugin:
		return nil, ErrUnsupportedStepType
	default:
		return nil, ErrUnsupportedStepType
	}

	pOut, ok := e.pipesStdOut.Load(step.UUID)
	if !ok {
		return nil, fmt.Errorf("Could not find key %s for pipesStdOut", step.UUID)
	}

	rc := struct {
		io.Reader
		io.Closer
	}{
		Reader: pOut.(*pipes).pr,
		Closer: closerFunc(func() error {
			log.Debug().Msgf("Done log streaming")
			return nil
		}),
	}

	return rc, nil
}

func (e *libvirt) DestroyStep(ctx context.Context, step *backend_types.Step, taskUUID string) error {
	switch step.Type {
	case types.StepTypeClone:
		// ignore cloning for now
		return ErrUnsupportedCloning
	case types.StepTypeCommands:
	case types.StepTypePlugin:
		return ErrUnsupportedStepType
	default:
		return ErrUnsupportedStepType
	}

	defer func() {
		// release lock
		err := mutex.SafeUnlock(&e.stepLock)
		if err != nil {
			log.Debug().Msgf("Error while unlocking step mutex: %s", err.Error())
		}

		keep_tmp := e.c.Bool("backend-libvirt-keep-tmp")

		// clean up ephemeral image
		if !keep_tmp {
			libvirtImgDir := e.c.String("backend-libvirt-image-dir")
			ephemeralImgs := filepath.Join(libvirtImgDir, fmt.Sprintf(baseImagePattern, "*", taskUUID, step.UUID, ".*"))
			matches, err := filepath.Glob(ephemeralImgs)
			if err != nil {
				log.Debug().Msg(err.Error())
			}

			for _, m := range matches {
				log.Debug().Msgf("Deleting base image: %s", m)
				err := os.Remove(m)
				if err != nil {
					log.Debug().Msgf("Error while unlocking step mutex: %s", err.Error())
				}
			}
		}
	}()

	sshCmd, ok := e.commands.Load(step.UUID)
	if !ok {
		return fmt.Errorf("Could not find key %s for commands", step.UUID)
	}
	log.Debug().Msgf("Destroying ssh connection")

	// closing may return EOF, which we can ignore
	sshCmd.(*goph.Cmd).Close()

	e.commands.Delete(step.UUID)

	// the domain has already been shutdown in WaitStep
	e.domains.Delete(step.UUID)

	return nil
}

func (e *libvirt) DestroyWorkflow(ctx context.Context, conf *backend_types.Config, taskUUID string) error {
	err := mutex.SafeUnlock(&e.stepLock)
	if err != nil {
		log.Debug().Msgf("Error while unlocking step mutex: %s", err.Error())
	}

	e.commands.Range(func(key, value any) bool {
		value.(*goph.Cmd).Close()
		e.commands.Delete(key)
		return true
	})
	e.domains.Range(func(key, value any) bool {
		ShutdownVM(ctx, value.(*virt.Domain))
		e.domains.Delete(key)
		return true
	})

	keep_tmp := e.c.Bool("backend-libvirt-keep-tmp")

	// clean up shared disk
	if !keep_tmp {
		libvirtImgDir := e.c.String("backend-libvirt-image-dir")
		ephemeralImgs := filepath.Join(libvirtImgDir, fmt.Sprintf(sharedDiskPattern, taskUUID, ".*"))
		matches, err := filepath.Glob(ephemeralImgs)
		if err != nil {
			return err
		}

		for _, m := range matches {
			err := os.Remove(m)
			log.Debug().Msgf("Deleting shared image: %s", m)
			if err != nil {
				return err
			}
		}
	}

	// clean up ephemeral images
	if !keep_tmp {
		libvirtImgDir := e.c.String("backend-libvirt-image-dir")
		ephemeralImgs := filepath.Join(libvirtImgDir, fmt.Sprintf(baseImagePattern, "*", taskUUID, "*", ".*"))
		matches, err := filepath.Glob(ephemeralImgs)
		if err != nil {
			return err
		}

		for _, m := range matches {
			err := os.Remove(m)
			log.Debug().Msgf("Deleting base image: %s", m)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
