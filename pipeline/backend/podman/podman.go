package podman

import (
	"context"
	"fmt"
	"io"
	"os"

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"

	"github.com/containers/podman/v3/pkg/bindings"
	"github.com/containers/podman/v3/pkg/bindings/containers"
	"github.com/containers/podman/v3/pkg/bindings/images"
	"github.com/containers/podman/v3/pkg/bindings/network"
	"github.com/containers/podman/v3/pkg/bindings/volumes"
	"github.com/containers/podman/v3/pkg/domain/entities"
)

var (
	noContext = context.Background()

	startOpts = containers.StartOptions{}

	removeOpts = containers.RemoveOptions{
		Ignore:  new(bool),
		Force:   new(bool),
		Volumes: new(bool),
	}

	logsOpts = containers.LogOptions{
		Follow: new(bool),
		//Since:  new(string),
		//Stderr: new(bool),
		Stdout: new(bool),
		//Tail:       new(string),
		//Timestamps: new(bool),
		//Until:      new(string),
	}

	killOpts = containers.KillOptions{
		Signal: new(string),
	}

	volRemoveOpts = volumes.RemoveOptions{
		Force: new(bool),
	}
)

type engine struct {
	conn   context.Context
	socket string
}

// New returns a new Podman Engine using the given client.
func New() backend.Engine {
	return &engine{
		conn:   nil,
		socket: "unix:" + os.Getenv("XDG_RUNTIME_DIR") + "/podman/podman.sock",
	}
}

func (e *engine) Name() string {
	return "podman"
}

func (e *engine) IsAvivable() bool {
	_, err := os.Stat("/run/.containerenv")
	return os.IsNotExist(err)
}

// Load new client for podman Engine using environment variables.
func (e *engine) Load() (err error) {
	*removeOpts.Ignore = false
	*removeOpts.Force = false
	*removeOpts.Volumes = true

	*logsOpts.Follow = true
	*logsOpts.Stdout = true
	//*logsOpts.Stderr = true
	//*logsOpts.Timestamps = false

	*killOpts.Signal = "SIGKILL"

	*volRemoveOpts.Force = true

	e.conn, err = bindings.NewConnection(context.Background(), e.socket)
	fmt.Printf("e.socket: %s\n", e.socket)
	fmt.Printf("e.conn: %v\n", e.conn)
	fmt.Printf("err: %v\n", err)
	return err
}

func (e *engine) Setup(_ context.Context, conf *backend.Config) error {
	for _, vol := range conf.Volumes {
		r, err := volumes.Create(e.conn, entities.VolumeCreateOptions{
			Name:    vol.Name,
			Driver:  vol.Driver,
			Options: vol.DriverOpts,
			// Labels:     defaultLabels,
		}, &volumes.CreateOptions{})
		fmt.Printf("vol: %s\n", vol.Driver)
		fmt.Printf("vol: %s\n", vol.DriverOpts)
		fmt.Printf("vol: %s\n", vol.Name)
		fmt.Printf("err: %v\n", err)
		fmt.Printf("r: %v\n", r)
		if err != nil {
			return err
		}
	}
	for _, netconf := range conf.Networks {
		_, err := network.Create(e.conn, &network.CreateOptions{
			Driver:  &netconf.Driver,
			Options: netconf.DriverOpts,
			Name:    &netconf.Name,
			// Labels:  defaultLabels,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *engine) Exec(ctx context.Context, proc *backend.Step) error {
	specGenerator, err := toSpecGenerator(proc)
	if err != nil {
		return err
	}

	fmt.Printf("Specgen: %v\n", specGenerator)

	// create pull options with encoded authorization credentials.
	pullopts := &images.PullOptions{}
	if proc.AuthConfig.Username != "" && proc.AuthConfig.Password != "" {
		pullopts.Username = &proc.AuthConfig.Username
		pullopts.Password = &proc.AuthConfig.Password
	}

	pullImage := proc.Pull

	// check if pull is disabled and pull image once if not existing
	if !pullImage {
		imageExists, err := images.Exists(e.conn, specGenerator.Image, &images.ExistsOptions{})
		if err != nil {
			return err
		}

		if !imageExists {
			pullImage = true
		}
	}

	// automatically pull the latest version of the image if requested
	// by the process configuration or not existing.
	if pullImage {
		_, perr := images.Pull(e.conn, specGenerator.Image, pullopts)
		// fix for drone/drone#1917
		if perr != nil && proc.AuthConfig.Password != "" {
			return perr
		}
	}

	// fix for missing workdir
	_workDir := specGenerator.WorkDir
	_entryPoint := specGenerator.Entrypoint
	specGenerator.WorkDir = "/"
	specGenerator.Entrypoint = []string{"mkdir", "-p", proc.WorkingDir}
	_, err = containers.CreateWithSpec(e.conn, specGenerator, &containers.CreateOptions{})
	if err != nil {
		return err
	}
	containers.Start(e.conn, specGenerator.Name, &startOpts)
	containers.Wait(e.conn, specGenerator.Name, &containers.WaitOptions{})
	containers.Remove(e.conn, specGenerator.Name, &removeOpts)

	// normal start here
	specGenerator.WorkDir = _workDir
	specGenerator.Entrypoint = _entryPoint
	_, err = containers.CreateWithSpec(e.conn, specGenerator, &containers.CreateOptions{})
	if err != nil {
		return err
	}

	return containers.Start(e.conn, specGenerator.Name, &startOpts)
}

func (e *engine) Kill(_ context.Context, proc *backend.Step) error {
	return containers.Kill(e.conn, proc.Name, &killOpts)
}

func (e *engine) Wait(ctx context.Context, proc *backend.Step) (*backend.State, error) {
	_, err := containers.Wait(e.conn, proc.Name, nil)
	if err != nil {
		return nil, err
	}

	info, err := containers.Inspect(e.conn, proc.Name, &containers.InspectOptions{})
	if err != nil {
		return nil, err
	}

	return &backend.State{
		Exited:    true,
		ExitCode:  int(info.State.ExitCode),
		OOMKilled: info.State.OOMKilled,
	}, nil
}

func (e *engine) Tail(ctx context.Context, proc *backend.Step) (io.ReadCloser, error) {
	rc, wc := io.Pipe()
	logChan := make(chan string, 10000)
	logEnd := make(chan bool)
	go func() {
		defer func() {
			containers.Wait(e.conn, proc.Name, nil)
			logEnd <- true
		}()
		containers.Logs(e.conn, proc.Name, &logsOpts, logChan, nil)
	}()

	go func() {
		for {
			select {
			case msg := <-logChan:
				if msg != "" {
					fmt.Fprint(wc, msg)
				}
			case <-logEnd:
				for {
					select {
					case msg := <-logChan:
						if msg != "" {
							fmt.Fprint(wc, msg)
						}
					default:
						wc.Close()
						rc.Close()
						return
					}
				}
			}
		}
	}()
	return rc, nil
}

func (e *engine) Destroy(_ context.Context, conf *backend.Config) error {
	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			containers.Kill(e.conn, step.Name, &killOpts)
			containers.Remove(e.conn, step.Name, &removeOpts)
		}
	}
	for _, volume := range conf.Volumes {
		volumes.Remove(e.conn, volume.Name, &volRemoveOpts)
	}
	for _, netconf := range conf.Networks {
		network.Remove(e.conn, netconf.Name, nil)
	}
	return nil
}
