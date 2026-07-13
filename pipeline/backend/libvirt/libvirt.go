package libvirt

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/beevik/etree"
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
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
)

var (
	ErrUnsupportedStepType = errors.New("unsupported step type")
)

type libvirt struct {
	conn      *virt.Connect
	workflows sync.Map
	c         *cli.Command
	os, arch  string
}

type workflow struct {
	commands  sync.Map
	sshErrors sync.Map
	domains   sync.Map
	pipes     sync.Map
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

// Get the ip address we can SSH connect to.
func GetDomainIP(ctx context.Context, domain *virt.Domain, interfaceName string) (*virt.DomainIPAddress, error) {
	var interfaces []virt.DomainInterface
	var err error

	// DOMAIN_INTERFACE_ADDRESSES_SRC_AGENT requires qemu-guest-agent
	interfaces, err = domain.ListAllInterfaceAddresses(virt.DOMAIN_INTERFACE_ADDRESSES_SRC_AGENT)
	if err != nil {
		log.Debug().Msgf("Failed to retrieve network interfaces via qemu-agent: %s", err.Error())
		// use lease as fallback
		interfaces, err = domain.ListAllInterfaceAddresses(virt.DOMAIN_INTERFACE_ADDRESSES_SRC_LEASE)
		if err != nil {
			return nil, err
		}
	}
	for _, interFace := range interfaces {
		log.Debug().Msgf("Checking interface %s ", interFace.Name)
		if interFace.Name == interfaceName {
			for _, address := range interFace.Addrs {
				log.Debug().Msgf("Checking address %s ", address.Addr)
				if address.Type == virt.IP_ADDR_TYPE_IPV4 {
					return &address, nil
				}
			}
		}
	}

	return nil, &pipeline_errors.PipelineError{
		Message: fmt.Sprintf("Could not find an IPv4 address for the interface %s to connect via ssh to", interfaceName),
		Type:    pipeline_errors.PipelineErrorTypeGeneric,
	}
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

func (e *libvirt) LoadDomain(ctx context.Context, image string, ephemeral bool, sharedDisk SharedDisk, taskUUID string, stepUUID string) (*virt.Domain, error) {
	domain, err := e.conn.LookupDomainByName(image)

	if err != nil {
		return nil, err
	}

	defer domain.Free()

	// get the XML from the loaded domain
	xml, err := domain.GetXMLDesc(virt.DOMAIN_XML_INACTIVE)
	if err != nil {
		return nil, err
	}
	// parse the XML into etree
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return nil, err
	}

	if ephemeral {
		log.Debug().Msgf("Setting up ephemeral disks")
		// replacing all the disks with temporary ones
		el := doc.FindElements("/domain/devices/disk[@type='file']")
		for _, child := range el {
			// get target to disambiguate
			targetEl := child.FindElement("./target[@dev]")
			if targetEl == nil {
				continue
			}
			devAttr := targetEl.SelectAttr("dev")
			if devAttr == nil {
				continue
			}

			// get source to disambiguate
			sourceEl := child.FindElement("./source[@file]")
			if sourceEl == nil {
				continue
			}
			fileAttr := sourceEl.SelectAttr("file")
			if fileAttr == nil {
				continue
			}

			baseImgName := fmt.Sprintf(baseImagePattern, devAttr.Value, taskUUID, stepUUID, filepath.Ext(fileAttr.Value))
			newImg, err := e.FromBaseImage(ctx, fileAttr.Value, baseImgName, false)
			if err != nil {
				return nil, err
			}
			sourceEl.CreateAttr("file", newImg)
		}
	}

	// insert the shared disk if any
	if sharedDisk.DiskConfig != "" {
		log.Debug().Msgf("Inserting shared disk: %s", sharedDisk.DiskConfig)

		// get devices element
		devices := doc.FindElement("/domain/devices")
		if devices == nil {
			return nil, fmt.Errorf("Could not find devices in domain XML")
		}

		// generate the XML with the base image
		sharedXml, err := e.CreateSharedDiskImage(ctx, &sharedDisk, taskUUID)
		if err != nil {
			return nil, err
		}
		sharedXmlDoc := etree.NewDocument()
		if err := sharedXmlDoc.ReadFromString(sharedXml); err != nil {
			return nil, err
		}

		// insert
		devices.AddChild(sharedXmlDoc.Root())
	}

	newXml, err := doc.WriteToString()
	log.Debug().Msgf("New domain XML: %s", newXml)
	if err != nil {
		return nil, err
	}

	return e.conn.DomainCreateXML(newXml, virt.DOMAIN_NONE)
}

func (e *libvirt) FromBaseImage(ctx context.Context, baseImage string, targetImg string, overwrite bool) (string, error) {
	libvirtImgDir := e.c.String("backend-libvirt-image-dir")
	libvirtImg := filepath.Join(libvirtImgDir, targetImg)

	if !overwrite {
		_, err := os.Stat(libvirtImg)
		if err == nil {
			return libvirtImg, nil
		}
	}

	log.Debug().Msgf("Copying base image %s to %s ", baseImage, libvirtImg)

	if filepath.Ext(baseImage) == ".qcow2" {
		cmd := exec.Command("qemu-img", "create", "-o", "backing_fmt=qcow2", "-b", baseImage, "-f", "qcow2", libvirtImg)
		_, err := cmd.Output()
		if err != nil {
			log.Debug().Msgf("Error copying base image %s via qemu-img...falling back to manual copy", string(err.(*exec.ExitError).Stderr))
			err := CopyFile(baseImage, libvirtImg, false)
			if err != nil {
				return "", err
			}
		}
	} else { // TODO: maybe other formats can be efficiently copied too?
		err := CopyFile(baseImage, libvirtImg, false)
		if err != nil {
			return "", err
		}
	}

	return libvirtImg, nil
}

func CopyFile(from string, to string, overwrite bool) error {
	fromFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFile.Close()

	if overwrite == false {
		_, err := os.Stat(to)
		if err == nil {
			return fmt.Errorf("File %s already exists", to)
		}
	}

	toFile, err := os.Create(to)
	if err != nil {
		return err
	}
	defer toFile.Close()

	_, err = io.Copy(toFile, fromFile)
	if err != nil {
		return err
	}

	return nil
}

// create or get the shared disk image
func (e *libvirt) CreateSharedDiskImage(ctx context.Context, shared_disk *SharedDisk, taskUUID string) (string, error) {
	if shared_disk.DiskConfig == "" {
		return "", fmt.Errorf("No image specified")
	}

	xml := shared_disk.DiskConfig
	log.Debug().Msgf("Shared disk config: %s", xml)

	// parse the XML into etree
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return "", err
	}

	// process the XML, replacing the disk
	el := doc.FindElements("//disk/source[@file]")
	for _, child := range el {
		attr := child.SelectAttr("file")
		if attr == nil {
			continue
		}

		sharedImg := fmt.Sprintf(sharedDiskPattern, taskUUID, filepath.Ext(attr.Value))

		newImg, err := e.FromBaseImage(ctx, attr.Value, sharedImg, false)
		if err != nil {
			return "", err
		}
		child.CreateAttr("file", newImg)
	}

	newXml, err := doc.WriteToString()
	if err != nil {
		return "", err
	}
	return newXml, nil
}

func (e *libvirt) SetupWorkflow(ctx context.Context, conf *backend_types.Config, taskUUID string) error {
	e.workflows.Store(taskUUID, &workflow{
		commands:  sync.Map{},
		sshErrors: sync.Map{},
		domains:   sync.Map{},
		pipes:     sync.Map{},
	})
	return nil
}

func AdhocSSH(ctx context.Context, client *goph.Client, cmd string, args []string, taskUUID string, stepUUID string) error {
	sshCmd, err := client.CommandContext(ctx, cmd, args...)
	log.Debug().Msgf("Executing via ssh: %s %v", cmd, args)
	if err != nil {
		return err
	}
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	err = sshCmd.Run()

	return err
}

func getGuestOS(domain *virt.Domain) (string, error) {
	// does not work on all hypervisors
	guestInfo, err := domain.GetGuestInfo(virt.DOMAIN_GUEST_INFO_OS, 0)
	if err != nil {
		// read the XML metadata, which sometimes carries libosinfo stuff
		xml, err := domain.GetXMLDesc(virt.DOMAIN_XML_INACTIVE)
		if err != nil {
			return "", err
		}

		// parse the XML into etree
		doc := etree.NewDocument()
		if err := doc.ReadFromString(xml); err != nil {
			return "", err
		}

		// TODO: there should be a library parsing this properly
		el := doc.FindElement("/domain/metadata/libosinfo:libosinfo/libosinfo:os")
		if el == nil {
			return "", fmt.Errorf("Could not determine guest platform, add libosinfo metadata to the domain")
		}
		fileAttr := el.SelectAttr("id")
		if fileAttr == nil {
			return "", fmt.Errorf("Could not determine guest platform, add libosinfo metadata to the domain")
		}
		switch {
		case strings.Contains(fileAttr.Value, "/win/"):
			return "windows", nil
		case strings.Contains(fileAttr.Value, "linux"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "ubuntu"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "suse"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "slackware"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "fedora"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "redhat"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "popos"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "oracle"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "nixos"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "guix"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "mandriva"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "mageia"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "gentoo"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "debian"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "centos"):
			return "linux", nil
		case strings.Contains(fileAttr.Value, "apple"):
			return "darwin", nil
		case strings.Contains(fileAttr.Value, "openbsd"):
			return "openbsd", nil
		case strings.Contains(fileAttr.Value, "freebsd"):
			return "freebsd", nil
		case strings.Contains(fileAttr.Value, "netbsd"):
			return "netbsd", nil
		case strings.Contains(fileAttr.Value, "dragonflybsd"):
			return "dragonflybsd", nil
		default:
			return "", fmt.Errorf("Could not determine guest platform, add libosinfo metadata to the domain")
		}

	} else {

		// TODO: what does 'ID' return?
		return guestInfo.OS.ID, nil
	}

}

func (e *libvirt) StartStep(ctx context.Context, step *backend_types.Step, taskUUID string) error {
	switch step.Type {
	case types.StepTypeClone:
		// ignore cloning for now
		return nil
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

	domain, err := e.LoadDomain(ctx, step.Image, options.Ephemeral, options.SharedDisk, taskUUID, step.UUID)
	if err != nil {
		return err
	}

	w, ok := e.workflows.Load(taskUUID)
	if !ok {
		return fmt.Errorf("Could not find key %s for workflows", taskUUID)
	}

	w.(*workflow).domains.Store(step.UUID, domain)

	guestOS, err := getGuestOS(domain)
	if err != nil {
		return err
	}

	env, entry, err := common.GenerateSSHConf(step.Commands, guestOS, step.WorkingDir)
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

	log.Debug().Msgf("Connecting to %s as user %s", sshAddr, options.SSHConfig.User)
	client, err := backoff.Retry(ctx, func() (*goph.Client, error) {
		return goph.NewConn(&goph.Config{
			User:     options.SSHConfig.User,
			Auth:     goph.Password(options.SSHConfig.Password),
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
	if options.SharedDisk.DiskConfig != "" {
		if guestOS == "windows" {
			if options.SharedDisk.UUID == "" {
				return fmt.Errorf("uuid needs to be defined [%s]", taskUUID)
			}
			encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
			cmd := "powershell"
			exec := fmt.Sprintf("$mount = '%s' ; $ErrorActionPreference = 'Stop' ; New-Item -ItemType Directory -Path $mount ; $disk = Get-Disk | Where-Object { $_.SerialNumber -eq '%s' } ; $partition = Get-Partition -DiskNumber $disk.DiskNumber | Where-Object { $_.Type -eq 'Basic' } ; Add-PartitionAccessPath -DiskNumber $disk.DiskNumber -PartitionNumber $partition.PartitionNumber -AccessPath $mount", step.WorkspaceBase, options.SharedDisk.UUID)
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

			// check if we're mounting by uuid or label
			var mountSource string
			if options.SharedDisk.UUID != "" {
				mountSource = fmt.Sprintf("UUID=%s", options.SharedDisk.UUID)
			} else if options.SharedDisk.Label != "" {
				mountSource = fmt.Sprintf("LABEL=%s", options.SharedDisk.Label)
			} else {
				return fmt.Errorf("Neither UUID nor Label defined")
			}

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
	}

	sshCmd, err := client.CommandContext(ctx, cmd, args...)
	if err != nil {
		return err
	}
	sshCmd.Env = flatMap

	w.(*workflow).commands.Store(step.UUID, sshCmd)

	// we need to create pipes and set Stdout here
	// in TailStep it is potentially too late
	pr, pw := nio.Pipe(buffer.New(64 * 1024))
	sshCmd.Stdout = pw
	sshCmd.Stderr = pw
	w.(*workflow).pipes.Store(step.UUID, &pipes{pr, pw})

	err = sshCmd.Start()
	if err != nil {
		return err
	}

	// Now we spawn a go routine that waits for the ssh command to exit
	// and then closes the write end of the pipe.
	// We need to do that here since we won't reach 'WaitStep' otherwise. After
	// the engine executed 'TailStep' it will wait until the logs have been fully
	// drainer before calling 'WaitStep'.
	go func() {
		e := sshCmd.Wait()
		if e != nil {
			w.(*workflow).sshErrors.Store(step.UUID, e)
		}
		time.Sleep(time.Second * 5)
		e = pw.Close()
		if e != nil {
			log.Debug().Msgf("Error in gofunc of StartStup: %s", e)
		}
	}()

	return nil
}

func (e *libvirt) WaitStep(ctx context.Context, step *backend_types.Step, taskUUID string) (*backend_types.State, error) {
	switch step.Type {
	case types.StepTypeClone:
		// ignore cloning for now
		return &backend_types.State{
			Skipped: true,
		}, nil
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

	r, ok := w.(*workflow).sshErrors.Load(step.UUID)
	var sshErr error
	if ok {
		sshErr = r.(error)
	} else {
		sshErr = nil
	}

	switch e := sshErr.(type) {
	case *ssh.ExitMissingError:
		return &backend_types.State{
			Exited:    true,
			ExitCode:  -1,
			OOMKilled: false,
		}, sshErr
	case *exec.ExitError:
		return &backend_types.State{
			Exited:    true,
			ExitCode:  e.ExitCode(),
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
		// ignore cloning for now
		emptyRC := io.NopCloser(strings.NewReader(""))
		return emptyRC, nil
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

	p, ok := w.(*workflow).pipes.Load(step.UUID)
	if !ok {
		return nil, fmt.Errorf("Could not find key %s for pipes", step.UUID)
	}

	var once sync.Once
	rc := struct {
		io.Reader
		io.Closer
	}{
		Reader: p.(*pipes).pr,
		Closer: closerFunc(func() error {
			once.Do(func() {
				p.(*pipes).pr.Close()
			})
			return nil
		}),
	}

	return rc, nil
}

func ShutdownVM(ctx context.Context, domain *virt.Domain) error {
	{
		err := domain.Shutdown()
		if err != nil {
			return err
		}

	}

	// wait 2 minutes for clean shutdown
	// otherwise force via 'Destroy()'
	maxBackOff, _ := time.ParseDuration("2m")
	log.Debug().Msgf("Checking if domain has shutdown for %.1f minutes", maxBackOff.Minutes())
	_, err := backoff.Retry(ctx, func() (any, error) {
		info, err := domain.GetInfo()
		if err != nil {
			return nil, err
		}
		switch info.State {
		case virt.DOMAIN_RUNNING:
			return nil, fmt.Errorf("VM still running, retrying")
		case virt.DOMAIN_SHUTDOWN:
			return nil, fmt.Errorf("VM is being shutdown, retrying")
		case virt.DOMAIN_SHUTOFF:
			return nil, nil
		default:
			return nil, backoff.Permanent(fmt.Errorf("Unexpected state: %d", info.State))
		}
	}, backoff.WithMaxElapsedTime(maxBackOff))

	if err != nil {
		err := domain.Destroy()
		if err != nil {
			return err
		}
	}

	domain.Free()

	return nil
}

func (e *libvirt) DestroyStep(ctx context.Context, step *backend_types.Step, taskUUID string) error {
	switch step.Type {
	case types.StepTypeClone:
		// ignore cloning for now
		return nil
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

	// clean up ephemeral image
	{
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

	// clean up shared disk
	{
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
	{
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
