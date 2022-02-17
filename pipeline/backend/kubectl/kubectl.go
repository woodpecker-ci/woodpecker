package kubectl

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"time"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

const (
	IfFailed    = "iffailed"
	IfSucceeded = "ifsucceeded"
	Always      = "always"
	Never       = "never"
)

type KubeCtlBackend struct {
	Client         *KubeCtlClient              // The kubernetes client
	RunID          string                      // the random run id
	Config         *types.Config               // the run config
	DeletePolicy   string                      // The job delete policy
	PVCs           []*KubePVCTemplate          // Loaded pvc's (via setup)
	PVCByName      map[string]*KubePVCTemplate // Loaded pvc's by name
	SetupTemplates []KubeTemplate              // Loaded setup templates
	RequestTimeout time.Duration               // The kubectl request timeout

	// internal

	podLogsContext context.Context    // The log waiter for non detached pods
	podLogsStop    context.CancelFunc // The log context reader cancel function
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
		RunID:          CreateRandomId(10),
		DeletePolicy:   IfSucceeded,
		RequestTimeout: 10 * time.Second,
	}
}

func (backend *KubeCtlBackend) Name() string {
	return "kubectl"
}

func (backend *KubeCtlBackend) IsAvailable() bool {
	// check if the executable exists. Otherwise false.
	// May need connection check afterwards.
	_, err := exec.LookPath(backend.Client.GetExecutable())
	return err != nil
}

func (backend *KubeCtlBackend) Load() error {
	// nothing to load.
	return nil
}

// Setup the pipeline environment. Creates the volumes and the other
// run artifacts used for the run.
func (backend *KubeCtlBackend) Setup(ctx context.Context, cfg *types.Config) error {
	// updating parameters
	backend.InitializeConfig(cfg)
	logger := backend.MakeLogger("")

	setupYaml, err := backend.RenderSetupYaml()
	if err != nil {
		return err
	}

	output, err := backend.Client.DeployKubectlYaml(ctx, "apply", setupYaml)

	if err != nil {
		return err
	}

	logger.Debug().Msgf("Pipeline setup complete with:\n %s", output)

	return nil
}

// Destroy the pipeline environment.
func (backend *KubeCtlBackend) Destroy(ctx context.Context, cfg *types.Config) error {
	logger := backend.MakeLogger("")
	setupYaml, err := backend.RenderSetupYaml()

	if err != nil {
		return err
	}

	output, err := backend.Client.DeployKubectlYaml(ctx, "delete", setupYaml)
	if err != nil {
		return err
	}

	logger.Debug().Msgf("Pipeline setup destroyed with:\n %s", output)
	return nil
}

// Exec the pipeline step.
func (backend *KubeCtlBackend) Exec(ctx context.Context, step *types.Step) error {
	jobTemplate := KubeJobTemplate{
		Backend: backend,
		Step:    step,
	}

	logger := backend.MakeLogger(jobTemplate.JobID())

	jobAsYaml, err := jobTemplate.Render()
	if err != nil {
		return err
	}

	output, err := backend.Client.DeployKubectlYaml(ctx, "apply", jobAsYaml)

	if err != nil {
		return err
	}

	logger.Debug().Msgf("Job initialized with\n %s", output)

	return nil
}

// Tail the pipeline step logs.
func (backend *KubeCtlBackend) Tail(ctx context.Context, step *types.Step) (io.ReadCloser, error) {
	jobTemplate := KubeJobTemplate{
		Backend: backend,
		Step:    step,
	}

	logger := backend.MakeLogger(jobTemplate.JobID())

	logsReader, logsWriter := io.Pipe()
	errorReader, errorWriter := io.Pipe()
	logsContext, logsContextCancel := context.WithCancel(ctx)

	stopLogger := func() {
		logsContextCancel()
		errorWriter.Close()
		logsWriter.Close()
	}

	backend.podLogsContext = logsContext
	backend.podLogsStop = stopLogger

	go func() {
		logger.Debug().Msg("Waiting for pod to exist")

		var podNames []string
		var err error

		for {
			podNames, err = backend.Client.GetResourceNames(
				logsContext,
				"pod",
				"woodpecker-job-id="+jobTemplate.JobID(),
			)

			if err != nil {
				logger.Error().Err(err).Msg("Error getting job pod names. Log reader failed.")
				stopLogger()
				return
			}

			if len(podNames) == 0 {
				logger.Error().Msg("Pods not ready. Retry [50 ms]")
				continue
			}

			break
		}

		// only wait for first pod ready.
		podName := podNames[0]

		// wait for pod to be ready.
		logger.Debug().Msg("Waiting for pod to be ready (Initialized)")
		_, err = backend.Client.WaitForConditions(
			logsContext,
			podName, []string{"Ready"},
			1,
		)

		if err != nil {
			logger.Error().Err(err).Msg("Failed to wait for job pod to start. Log reader failed.")
			stopLogger()
			return
		}

		logger.Debug().Msgf("Reading logs (%s)", podName)

		logsCmd := backend.Client.CreateKubectlCommand(logsContext,
			"logs",
			podName,
			"-f",
		)

		logsCmd.Stdout = logsWriter
		logsCmd.Stderr = errorWriter

		err = logsCmd.Run()

		logsWriter.Close()
		errorWriter.Close()

		stderr, _ := GetReaderContents(errorReader)

		if err != nil {
			if len(stderr) > 0 {
				err = errors.New(stderr + "; " + err.Error())
			}
			logger.Error().Err(err).Msg("Error reading logs")
		}

		if len(stderr) != 0 {
			logger.Error().Err(errors.New(stderr)).Msg("Error reading logs")
		}

		stopLogger()

	}()

	return logsReader, nil
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (backend *KubeCtlBackend) Wait(ctx context.Context, step *types.Step) (*types.State, error) {
	jobTemplate := KubeJobTemplate{
		Backend: backend,
		Step:    step,
	}

	logger := backend.MakeLogger(jobTemplate.JobID())

	condition, jobEndConditionError := backend.Client.WaitForConditions(
		ctx,
		"job/"+jobTemplate.JobName(),
		[]string{"Complete", "Failed"},
		1,
	)

	logger.Debug().Msgf("Job ended with '%s'", condition)

	// wait for logs (or give error after the request timeout)
	if !step.Detached {
		select {
		case <-time.After(backend.RequestTimeout):
			logger.Error().Msg(
				"Error reading logs, request timeout or kubectl logger stuck.",
			)
			break
		case <-backend.podLogsContext.Done():
			logger.Debug().Msg("Read job logs: OK!")
			break
		}

		backend.podLogsStop()
	}

	// currently we are not reading the proper exit code from the pods
	// but rather checking the job error.
	// TODO: Support error codes.
	state := &types.State{
		ExitCode:  99, // timeout
		Exited:    false,
		OOMKilled: false,
	}

	if jobEndConditionError == nil {
		state.Exited = true
		if condition == "Complete" {
			state.ExitCode = 0
		} else {
			state.ExitCode = 1
		}
	}

	// Checking what to do for state
	doDelete := false
	switch backend.DeletePolicy {
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

		out, err := backend.Client.DeployKubectlYaml(ctx, "delete", asJobYaml)
		if err != nil {
			return state, errors.New(out + ". " + err.Error())
		}

		logger.Debug().Msgf("Job DELETD with: \n" + out)
	} else {
		logger.Info().Msgf("Job artifact kept in cluster (%s)", backend.DeletePolicy)
	}

	logger.Debug().Msg("Job DONE")

	return state, nil
}
