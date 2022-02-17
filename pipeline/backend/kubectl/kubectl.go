package kubectl

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

const (
	IfFailed    = "iffailed"
	IfSucceeded = "ifsucceeded"
	Always      = "always"
	Never       = "never"
)

type KubeCtlBackend struct {
	Client       *KubeCtlClient // The kubernetes client
	RunID        string         // the random run id
	Config       *types.Config  // the run config
	DeletePolicy string         // The job delete policy
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
		RunID:        createRandomId(10),
		DeletePolicy: IfSucceeded,
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
func (this *KubeCtlBackend) Exec(ctx context.Context, step *types.Step) error {
	jobTemplate := KubeJobTemplate{
		Engine: this,
		Step:   step,
	}

	jobAsYaml, err := jobTemplate.Render()
	if err != nil {
		return err
	}

	output, err := this.Client.DeployKubectlYaml("apply", jobAsYaml)

	if err != nil {
		return err
	}

	log.Debug().Msgf("Pipeline exec job %s initialized with\n %s", jobTemplate.JobID(), output)

	return nil
}

// Tail the pipeline step logs.
func (this *KubeCtlBackend) Tail(ctx context.Context, step *types.Step) (io.ReadCloser, error) {
	// must run a-synchronically.
	// Logs in this context cant fail, if they do, a log failed

	jobTemplate := KubeJobTemplate{
		Engine: this,
		Step:   step,
	}

	pipeReader, pipeWriter := io.Pipe()

	go func() {
		log.Debug().Msgf(
			"Waiting for job '%s' pod to be ready before reading logs",
			jobTemplate.JobID(),
		)

		// run until stopped.
		for {
			logsCmd := this.Client.GetKubectlCommandContext(ctx,
				"logs", "-f",
				"-l", "jobid="+jobTemplate.JobID(),
			)

			logsCmd.Stdout = pipeWriter
			// logsCmd.Stderr = pipeWriter

			err := logsCmd.Run()

			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			break
		}
	}()

	return pipeReader, nil
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (this *KubeCtlBackend) Wait(ctx context.Context, step *types.Step) (*types.State, error) {
	jobTemplate := KubeJobTemplate{
		Engine: this,
		Step:   step,
	}

	condition, waitError := this.Client.WaitForConditions(
		ctx,
		"job/"+jobTemplate.JobName(),
		[]string{"Complete", "Failed"},
		1, time.Hour*24*7,
	)

	// currently we are not reading the proper exit code from the pods
	// but rather checking the job error.
	// TODO: Support error codes.
	state := &types.State{
		ExitCode:  99, // timeout
		Exited:    false,
		OOMKilled: false,
	}

	if waitError == nil {
		state.Exited = true
		if condition == "Complete" {
			state.ExitCode = 0
		} else {
			state.ExitCode = 1
		}
	}

	// Checking what to do for state
	doDelete := false
	switch this.DeletePolicy {
	case IfFailed:
		if state.ExitCode != 0 {
			doDelete = true
		}
		break
	case IfSucceeded:
		if state.ExitCode == 0 {
			doDelete = true
		}
		break
	case Always:
		doDelete = true
		break
	}

	if doDelete {
		asJobYaml, err := jobTemplate.Render()
		if err != nil {
			return state, err
		}

		out, err := this.Client.DeployKubectlYaml("delete", asJobYaml)
		if err != nil {
			return state, errors.New(out + ". " + err.Error())
		}

		log.Debug().Msgf("Job %s deleted", jobTemplate.JobID())
	}

	return state, nil
}
