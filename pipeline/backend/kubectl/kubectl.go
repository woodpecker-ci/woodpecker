package kubectl

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"strconv"
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
	Client         *KubeCtlClient // The kubernetes client
	Config         *types.Config  // the run config
	RunID          string         // the random run id
	DeletePolicy   string         // The job delete policy
	JobMemoryLimit string         // The runner container memory limit (1Gi)
	JobCPULimit    string         // The runner container cpu limit (200m)

	// A delay before the pod start. Various reasons.
	// Most notably some backend CNI's (like flannel)
	// require this to apply egress network policy to the pod.
	ContainerStartDelay int64

	PVCs           []*KubePVCTemplate          // Loaded pvc's (via setup)
	PVCByName      map[string]*KubePVCTemplate // Loaded pvc's by name
	SetupTemplates []KubeTemplate              // Loaded setup templates
	DetachedJobs   []*KubeJobTemplate          // Loaded detached template (services, etc..)
	RequestTimeout time.Duration               // The kubectl request timeout

	// flags
	PVCAllowOnDetached     bool // Allow pvc's on detached containers
	EnableRunNetworkPolicy bool // Do not implement a network policy when running a pipeline.

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
		CoreArgs: args.Merge(KubeCtlClientCoreArgs{
			Namespace: getWPKEnv("NAMESPACE", "").(string),
			Context:   getWPKEnv("CONTEXT", "").(string),
		}),
	}

	requestTimeoutSeconds, _ := strconv.ParseFloat(getWPKEnv("REQUEST_TIMEOUT", "10").(string), 64)
	containerStartDelaySeconds, _ := strconv.ParseInt(getWPKEnv("REQUEST_TIMEOUT", "10").(string), 0, 64)

	return &KubeCtlBackend{
		Client: client,
		Config: &types.Config{},
		// the engine id must be randomized in order not to clobber/error other runs.
		RunID:        CreateRandomID(10),
		DeletePolicy: getWPKEnv("DELETE_POLICY", IfSucceeded).(string),

		RequestTimeout:         time.Duration(requestTimeoutSeconds) * time.Second,
		PVCAllowOnDetached:     getWPKEnv("ALLOW_PVC_ON_DETACHED", "false").(string) == "true",
		EnableRunNetworkPolicy: getWPKEnv("ENABLE_NETWORK_POLICY", "false").(string) == "true",
		ContainerStartDelay:    containerStartDelaySeconds,
		JobMemoryLimit:         getWPKEnv("MEMORY_LIMIT", "1Gi").(string),
		JobCPULimit:            getWPKEnv("CPU_LIMIT", "500m").(string), // half a cpu
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

	logger.Debug().Msg("Loading kube client defaults")
	err := backend.Client.LoadDefaults(ctx)
	if err != nil {
		return err
	}

	setupYaml, err := backend.RenderSetupYaml()
	if err != nil {
		return err
	}

	output, err := backend.Client.DeployKubectlYaml(ctx, "apply", setupYaml, false)

	if err != nil {
		return err
	}

	logger.Info().Str(
		"namespace", backend.Namespace(),
	).Str(
		"context", backend.Client.CoreArgs.Context,
	).Msgf("Started pipeline execution")

	logger.Debug().Msgf("Kubectl setup response:\n %s", output)

	return nil
}

// Destroy the pipeline environment.
func (backend *KubeCtlBackend) Destroy(ctx context.Context, cfg *types.Config) error {
	logger := backend.MakeLogger("")
	destoryYaml, err := backend.RenderSetupYaml()

	if err != nil {
		return err
	}

	if len(backend.DetachedJobs) > 0 {
		logger.Debug().Msg("Destroying detached jobs")
		for _, job := range backend.DetachedJobs {
			jobYaml, err := job.Render()
			if err != nil {
				return err
			}
			// adding to the destroy command
			destoryYaml += "\n---\n" + jobYaml
		}
	}

	output, err := backend.Client.DeployKubectlYaml(ctx, "delete", destoryYaml, false)
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

	if step.Detached {
		step.Alias = Triary(
			len(step.Alias) > 0, step.Alias, toKuberenetesValidName(step.Name, 50),
		).(string)
	}

	logger := backend.MakeLogger(jobTemplate.JobID())

	jobAsYaml, err := jobTemplate.Render()
	if err != nil {
		return err
	}

	output, err := backend.Client.DeployKubectlYaml(ctx, "apply", jobAsYaml, false)

	if err != nil {
		return err
	}

	logger.Debug().Msgf("Job initialized with\n %s", output)

	if step.Detached {
		backend.DetachedJobs = append(backend.DetachedJobs, &jobTemplate)

		// loaded a detached service. We need to wait for it to start,
		// and load the IP from it. The add that to the DetachedJobs.
		logger.Debug().Msg("Waiting for detached job pod to exist")
		podNames, err := backend.GetJobPodName(ctx, &jobTemplate)
		if err != nil {
			return err
		}
		podName := podNames[0]

		logger.Debug().Msg("Waiting for detached job pod to be ready (Initialized)")
		_, err = backend.Client.WaitForConditions(
			ctx,
			podName, []string{"Initialized"},
			1,
		)
		if err != nil {
			return err
		}

		logger.Debug().Msg("Reading detached pod info")
		err = backend.PopulateDetachedInfo(ctx, podName, &jobTemplate)
		if err != nil {
			return err
		}

		logger.Debug().Msg("Detached step configured")
	}

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

	stopLoggerWithError := func(err error, msg string) {
		logsContextCancel()
		errorWriter.Close()
		logsWriter.Close()

		if err != nil {
			event := logger.Error()
			stderr, _ := GetReaderContents(errorReader)
			stdout, _ := GetReaderContents(logsReader)
			if len(stdout) > 0 {
				event = event.Str("stdout", stdout)
			}
			if len(stderr) > 0 {
				event = event.Str("stderr", stderr)
			}
			event.Err(err).Msg(msg)
		}
	}

	stopLogger := func() {
		stopLoggerWithError(nil, "")
	}

	backend.podLogsContext = logsContext
	backend.podLogsStop = stopLogger

	go func() {
		logger.Debug().Msg("Waiting for pod to exist")
		podNames, err := backend.GetJobPodName(logsContext, &jobTemplate)

		if err != nil {
			stopLoggerWithError(err, "Failed to retrieve job pod name(s). Log reader failed.")
			return
		}

		// only wait for first pod ready.
		podName := podNames[0]

		// wait for pod to be ready.
		logger.Debug().Msg("Waiting for pod to be ready (Initialized)")
		_, err = backend.Client.WaitForConditions(
			logsContext,
			podName, []string{"Initialized"},
			1,
		)

		if err != nil {
			stopLoggerWithError(err, "Failed to wait for job pod to start. Log reader failed.")
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

		if err != nil || len(stderr) > 0 {
			stopLoggerWithError(err, "Error while reading logs")
		} else {
			stopLogger()
		}
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

		out, err := backend.Client.DeployKubectlYaml(ctx, "delete", asJobYaml, true)
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
