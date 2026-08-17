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

	"github.com/cenkalti/backoff/v7"
	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"golang.org/x/crypto/ssh"
	"golang.org/x/text/encoding/unicode"
	virt "libvirt.org/go/libvirt"

	"github.com/melbahja/goph"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/common"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

var (
	ErrUnsupportedStepType = errors.New("unsupported step type")
	ErrUnsupportedCloning  = errors.New("Automatic cloning is not supported in libvirt. Set 'skip_clone: true'")
)

type libvirt struct {
	conn              *virt.Connect
	workflows         sync.Map
	c                 *cli.Command
	os, arch, guestOS string
}

type workflow struct {
	commands    sync.Map
	domains     sync.Map
	pipesStdOut sync.Map
	pipesStdIn  sync.Map
	doneChannel sync.Map
	guestOS     sync.Map
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

// New returns a new Docker Backend.
func New() backend_types.Backend {
	return &libvirt{
		conn:      nil,
		workflows: sync.Map{},
		c:         nil,
		os:        runtime.GOOS,
		arch:      runtime.GOARCH,
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
	e.workflows.Store(taskUUID, &workflow{
		commands:    sync.Map{},
		domains:     sync.Map{},
		pipesStdOut: sync.Map{},
		pipesStdIn:  sync.Map{},
		doneChannel: sync.Map{},
		guestOS:     sync.Map{},
	})
	return nil
}

func (e *libvirt) StartStep(ctx context.Context, step *backend_types.Step, taskUUID string) error {
	switch step.Type {
	case types.StepTypeClone:
		return ErrUnsupportedCloning
	case types.StepTypeCommands:
	case types.StepTypePlugin:
		return ErrUnsupportedStepType
	default:
		return ErrUnsupportedStepType
	}

	options, err := parseBackendOptions(step)
	if err != nil {
		log.Error().Err(err).Msg("could not parse backend options")
	}

	domain, uuid, guestOS, err := e.LoadDomain(ctx, step.Image, step.Environment, options.Ephemeral, options.SharedDisk, taskUUID, step.UUID)
	if err != nil {
		return err
	}

	w, ok := e.workflows.Load(taskUUID)
	if !ok {
		return fmt.Errorf("Could not find key %s for workflows", taskUUID)
	}

	w.(*workflow).domains.Store(step.UUID, domain)
	w.(*workflow).guestOS.Store(step.UUID, guestOS)

	env, entry, err := common.GenerateSSHConf(step.Commands, guestOS, step.WorkingDir, step.UUID)
	cmd := entry[0]
	args := entry[1:]

	if err != nil {
		return err
	}
	var flatMap []string
	for key, value := range env {
		flatMap = append(flatMap, fmt.Sprintf("%s=%s", key, value))
	}

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

	// get the guest network interface we're going to inspect for the IP
	var sshAddr string
	if options.SSHConfig.Hostname != "" {
		sshAddr = options.SSHConfig.Hostname
	} else {
		var netInterface string
		if options.SSHConfig.GuestInterface != "" {
			netInterface = options.SSHConfig.GuestInterface
		} else {
			// default
			netInterface = defaultGuestInterface
		}
		ip, err := backoff.Retry(ctx, func() (*virt.DomainIPAddress, error) {
			ip, err := GetDomainIP(ctx, domain, netInterface)
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
		exec := fmt.Sprintf("$mount = '%s' ; $ErrorActionPreference = 'Stop' ; New-Item -ItemType Directory -Path $mount ; $disk = Get-Disk | Where-Object { $_.SerialNumber -eq '%s' } ; $partition = Get-Partition -DiskNumber $disk.DiskNumber | Where-Object { $_.Type -eq 'Basic' } ; Add-PartitionAccessPath -DiskNumber $disk.DiskNumber -PartitionNumber $partition.PartitionNumber -AccessPath $mount", step.WorkspaceBase, uuid)
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
	sshCmd.Env = flatMap

	w.(*workflow).commands.Store(step.UUID, sshCmd)

	if options.SSHConfig.Tty {
		err := sshCmd.RequestPty("xterm", 40, 80, ssh.TerminalModes{})
		if err != nil {
			return err
		}
	}

	// we need to create pipes and set Stdout here
	// in TailStep it is potentially too late
	pr, pw := nio.Pipe(buffer.New(64 * 1024))
	sshCmd.Stdout = pw
	sshCmd.Stderr = pw
	w.(*workflow).pipesStdOut.Store(step.UUID, &pipes{pr, pw})

	{
		pr, pw := nio.Pipe(buffer.New(64 * 1024))
		sshCmd.Stdin = pr
		w.(*workflow).pipesStdIn.Store(step.UUID, &pipes{pr, pw})
	}

	done := make(chan struct{})
	w.(*workflow).doneChannel.Store(step.UUID, done)
	// and a go routine that watches the ctx and then triggers a signal
	go func() {
		select {
		case <-ctx.Done():
			err := e.TerminateSshCommand(options, client, sshCmd, guestOS, taskUUID, step.UUID)
			if err != nil {
				log.Debug().Msgf("Failed to terminate SSH command gracefully. Closing session by force. Error was: %s", err)
				sshCmd.Close()
			}
		case <-done:
		}
		time.Sleep(time.Second * 5)
		log.Debug().Msg("Closing write pipe end")
		_ = pw.Close()
	}()

	err = sshCmd.Start()
	if err != nil {
		return err
	}

	return nil
}

func (e *libvirt) WaitStep(ctx context.Context, step *backend_types.Step, taskUUID string) (*backend_types.State, error) {
	switch step.Type {
	case types.StepTypeClone:
		return nil, ErrUnsupportedCloning
	case types.StepTypeCommands:
	case types.StepTypePlugin:
		return nil, ErrUnsupportedStepType
	default:
		return nil, ErrUnsupportedStepType
	}

	w, ok := e.workflows.Load(taskUUID)
	if !ok {
		return nil, fmt.Errorf("Could not find key %s for workflows", taskUUID)
	}

	sshCmd, ok := w.(*workflow).commands.Load(step.UUID)
	if !ok {
		return nil, fmt.Errorf("Could not find key %s for commands", step.UUID)
	}

	done, ok := w.(*workflow).doneChannel.Load(step.UUID)
	if !ok {
		return nil, fmt.Errorf("Could not find key %s for doneChannel", step.UUID)
	}

	sshErr := sshCmd.(*goph.Cmd).Wait()
	close(done.(chan struct{}))

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

	w, ok := e.workflows.Load(taskUUID)
	if !ok {
		return nil, fmt.Errorf("Could not find key %s for workflows", taskUUID)
	}

	pOut, ok := w.(*workflow).pipesStdOut.Load(step.UUID)
	if !ok {
		return nil, fmt.Errorf("Could not find key %s for pipesStdOut", step.UUID)
	}

	pIn, ok := w.(*workflow).pipesStdIn.Load(step.UUID)
	if !ok {
		return nil, fmt.Errorf("Could not find key %s for pipesStdIn", step.UUID)
	}

	var once sync.Once
	rc := struct {
		io.Reader
		io.Closer
	}{
		Reader: pOut.(*pipes).pr,
		Closer: closerFunc(func() error {
			once.Do(func() {
				pOut.(*pipes).pr.Close()
				pOut.(*pipes).pw.Close()

				pIn.(*pipes).pw.Close()
				pIn.(*pipes).pr.Close()
			})
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

	w, ok := e.workflows.Load(taskUUID)
	if !ok {
		return fmt.Errorf("Could not find key %s for workflows", taskUUID)
	}

	domain, ok := w.(*workflow).domains.Load(step.UUID)
	if !ok {
		return fmt.Errorf("Could not find key %s for domains", step.UUID)
	}
	sshCmd, ok := w.(*workflow).commands.Load(step.UUID)
	if !ok {
		return fmt.Errorf("Could not find key %s for commands", step.UUID)
	}
	log.Debug().Msgf("Destroying ssh connection")

	// closing may return EOF, which we can ignore
	sshCmd.(*goph.Cmd).Close()

	w.(*workflow).commands.Delete(step.UUID)
	log.Debug().Msgf("Destroying VM")
	err := ShutdownVM(ctx, domain.(*virt.Domain))
	if err != nil {
		return err
	}
	w.(*workflow).domains.Delete(step.UUID)

	keep_tmp := e.c.Bool("backend-libvirt-keep-tmp")

	// clean up ephemeral image
	if !keep_tmp {
		libvirtImgDir := e.c.String("backend-libvirt-image-dir")
		ephemeralImgs := filepath.Join(libvirtImgDir, fmt.Sprintf(baseImagePattern, "*", taskUUID, step.UUID, ".*"))
		matches, err := filepath.Glob(ephemeralImgs)
		if err != nil {
			return err
		}

		for _, m := range matches {
			err := os.Remove(m)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (e *libvirt) DestroyWorkflow(ctx context.Context, conf *backend_types.Config, taskUUID string) error {
	w, ok := e.workflows.Load(taskUUID)
	if !ok {
		return fmt.Errorf("Could not find key %s for workflows", taskUUID)
	}
	w.(*workflow).commands.Range(func(key, value any) bool {
		value.(*goph.Cmd).Close()
		w.(*workflow).commands.Delete(key)
		return true
	})
	w.(*workflow).domains.Range(func(key, value any) bool {
		ShutdownVM(ctx, value.(*virt.Domain))
		w.(*workflow).domains.Delete(key)
		return true
	})

	e.workflows.Delete(taskUUID)

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
			if err != nil {
				return err
			}
		}
	}

	return nil
}
