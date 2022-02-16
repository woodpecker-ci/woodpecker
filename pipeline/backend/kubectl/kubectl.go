package kubectl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type KubeCtlBackend struct {
	Client *KubeCtlClient // The kubernetes client
	RunID  string         // the random run id
	Config *types.Config  // the run config
}

var _ types.Engine = &KubeCtlBackend{}

func New(execuatble string, args KubeCtlClientCoreArgs) types.Engine {
	// create a new kubectl (exec based) engine. Allows for execution pods
	// as commands. Assumes running inside a cluster or kubectl is configured.

	client := &KubeCtlClient{
		Executable: execuatble,
		CoreArgs:   args,
	}

	return &KubeCtlBackend{
		Client: client,
		Config: &types.Config{},
		// the engine id must be randomized in order not to clobber/error other runs.
		RunID: createRandomId(10),
	}
}

func (this *KubeCtlBackend) Name() string {
	return "kubectl"
}

func (this *KubeCtlBackend) IsAvailable() bool {
	// check if the executable exists. Otherwise false.
	// May need connection check afterwards.
	_, err := exec.LookPath(this.Client.GetExecutable())
	return err != nil
}

func (this *KubeCtlBackend) Load() error {
	// nothing to load.
	return nil
}

// Setup the pipeline environment. Creates the volumes and the other
// run artifacts used for the run.
func (this *KubeCtlBackend) Setup(_ context.Context, cfg *types.Config) error {
	// updating parameters
	this.Config = cfg

	setupYaml, err := this.RenderSetupYaml()
	if err != nil {
		return err
	}

	output, err := this.Client.DeployKubectlYaml("apply", setupYaml)

	if err != nil {
		return err
	}

	log.Debug().Msgf("Pipeline setup for %s setup with:\n %s", this.RunID, output)

	return nil
}

// Destroy the pipeline environment.
func (this *KubeCtlBackend) Destroy(_ context.Context, cfg *types.Config) error {
	this.Config = cfg

	setupYaml, err := this.RenderSetupYaml()

	if err != nil {
		return err
	}

	output, err := this.Client.DeployKubectlYaml("delete", setupYaml)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Pipeline setup for %s destroyed with:\n %s", this.RunID, output)
	return nil
}

// Exec the pipeline step.
func (this *KubeCtlBackend) Exec(_ context.Context, step *types.Step) error {
	jobTemplate := KubeJobTemplate{
		Engine: this,
		Image:  step.Image,
		Name:   step.Name,
	}

	jobAsYaml, err := jobTemplate.Render()
	if err != nil {
		return err
	}

	output, err := this.Client.DeployKubectlYaml("apply", jobAsYaml)

	if err != nil {
		return err
	}

	log.Debug().Msgf("Pipeline exec job %s initialized with\n %s", jobTemplate.Name, output)

	return nil
}

// Tail the pipeline step logs.
func (this *KubeCtlBackend) Tail(ctx context.Context, step *types.Step) (io.ReadCloser, error) {
	// must run a-synchronically.
	// Logs in this context cant fail, if they do, a log failed

	jobTemplate := KubeJobTemplate{
		Engine: this,
		Image:  step.Image,
		Name:   step.Name,
	}

	logsCmd := this.Client.GetKubectlCommandContext(ctx,
		"logs", "-f",
		"-l", "jobid="+jobTemplate.JobID(),
	)

	logsReader, err := logsCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = logsCmd.Start()
	if err != nil {
		return nil, err
	}

	return logsReader, nil
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (this *KubeCtlBackend) Wait(ctx context.Context, step *types.Step) (*types.State, error) {

	jobTemplate := KubeJobTemplate{
		Engine: this,
		Image:  step.Image,
		Name:   step.Name,
	}

	// There will be two
	waitCommand := this.Client.ComposeKubectlCommand(
		"wait",
		this.Client.CoreArgs.ToArgsList(),
		"--timeout", fmt.Sprint(60*60*24*7)+"s",
		"job/"+jobTemplate.JobName(),
	)

	successCmd := this.Client.GetKubectlCommandContext(ctx, waitCommand, "--for", "condition=Complete")
	failureCmd := this.Client.GetKubectlCommandContext(ctx, waitCommand, "--for", "condition=Failed")

	var waiter sync.WaitGroup
	var waitError error

	completed := false
	succeeded := false

	// run failure
	go func() {
		out, err := failureCmd.CombinedOutput()
		if completed {
			return
		}
		completed = true
		if err != nil {
			waitError = errors.New(string(out) + "\n" + err.Error())
		}
		waiter.Done()
	}()

	// Run success
	go func() {
		out, err := successCmd.CombinedOutput()
		if completed {
			return
		}
		succeeded = true
		completed = true
		if err != nil {
			waitError = errors.New(string(out) + "\n" + err.Error())
		}
		waiter.Done()
	}()

	// Wait for conditions.
	waiter.Add(1)
	waiter.Wait()
	completed = true

	if waitError != nil {
		log.Debug().Err(waitError).Msg("Error while waiting for job to complete")
	}

	// Stopping process if exists
	if successCmd.Process != nil {
		_ = failureCmd.Process.Kill()
	}
	if failureCmd.Process != nil {
		_ = failureCmd.Process.Kill()
	}

	// currently we are not reading the proper exit code from the pods.
	// this is to support future implement of retries (Maybe?)
	state := &types.State{
		ExitCode:  99, // timeout
		Exited:    false,
		OOMKilled: false,
	}

	if waitError == nil {
		state.Exited = true
		if succeeded {
			state.ExitCode = 0
		} else {
			state.ExitCode = 1
		}
	}

	return state, nil
}
