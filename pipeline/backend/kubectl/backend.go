package kubectl

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

const (
	IfFailed    = "iffailed"
	IfSucceeded = "ifsucceeded"
	Always      = "always"
	Never       = "never"
)

type KubeBackendRun struct {
	RunID          string                         // the random run id
	Config         *types.Config                  // the run config
	PVCs           []*KubePVCTemplate             // Loaded pvc's (via setup)
	PVCByName      map[string]*KubePVCTemplate    // Loaded pvc's by name
	SetupTemplates []KubeTemplate                 // Loaded setup templates
	DetachedJobs   []*KubeJobTemplate             // Loaded detached template (services, etc..)
	StepLoggers    map[string]*KubeResourceLogger // A collection of loggers per step.
	JobPendingWait *KubeJobTemplate               // The executed job where wait was not initialized.
}

type KubeBackend struct {
	Client           *KubeClient   // The kubernetes client
	DeletePolicy     string        // The job delete policy
	JobMemoryLimit   string        // The runner container memory limit (1Gi)
	JobCPULimit      string        // The runner container cpu limit (200m)
	ForcePullPolicy  string        // Forces a pull policy on all jobs
	CommandRetries   int           // The number of times to retry commands.
	CommandRetryWait time.Duration // The wait time between command retries.

	// A delay before the pod start. Various reasons.
	// Most notably some backend CNI's (like flannel)
	// require this to apply egress network policy to the pod.
	ContainerStartDelay int64

	// The grace period (seconds) to allow pods to exit.
	// This will determine the canceled/errored memory overhead.
	// Recommended 5 seconds.
	TerminationGracePeriod int64

	// flags
	PVCAllowOnDetached     bool // Allow pvc's on detached containers
	EnableRunNetworkPolicy bool // Do not implement a network policy when running a pipeline.

	activeRun *KubeBackendRun // The current kubectl engine active run.

}

var _ types.Engine = &KubeBackend{}

// Create a new kubectl (exec based) engine. Allows for execution pods
// as commands. Assumes running inside a cluster or kubectl is configured.
func New(execuatble string, args KubeClientCoreArgs) types.Engine {
	requestTimeoutSeconds, _ := strconv.ParseFloat(getWPKEnv("REQUEST_TIMEOUT", "10").(string), 64)
	client := &KubeClient{
		Executable: execuatble,
		CoreArgs: args.Merge(KubeClientCoreArgs{
			Namespace: getWPKEnv("NAMESPACE", "").(string),
			Context:   getWPKEnv("CONTEXT", "").(string),
		}),
		RequestTimeout: time.Duration(requestTimeoutSeconds) * time.Second,

		// TODO: There is an error that dose not allow these type of setting
		// whilst running in-cluster. I set this default to false, but can be
		// changed for newer version of kubectl.
		// ERROR: https://github.com/kubernetes/kubernetes/issues/93474
		AllowClientConfiguration: getWPKEnv("ALLOW_CLIENT_CONFIG", "false").(string) == "true",
	}

	containerStartDelaySeconds, _ := strconv.ParseInt(getWPKEnv("REQUEST_TIMEOUT", "10").(string), 0, 64)
	terminationGracePeriodSeconds, _ := strconv.ParseInt(getWPKEnv("TERMINATION_GRACE_PERIOD", "5").(string), 0, 64)
	commandRetries, _ := strconv.Atoi(getWPKEnv("COMMAND_RETRIES", "5").(string))
	commandRetriesWait, _ := strconv.ParseFloat(getWPKEnv("COMMAND_RETRIES_WAIT", "1").(string), 64)

	return &KubeBackend{
		Client:       client,
		DeletePolicy: getWPKEnv("DELETE_POLICY", Always).(string),

		PVCAllowOnDetached:     getWPKEnv("ALLOW_PVC_ON_DETACHED", "false").(string) == "true",
		EnableRunNetworkPolicy: getWPKEnv("ENABLE_NETWORK_POLICY", "false").(string) == "true",
		ForcePullPolicy:        getWPKEnv("FORCE_PULL_POLICY", "").(string),
		TerminationGracePeriod: terminationGracePeriodSeconds,
		ContainerStartDelay:    containerStartDelaySeconds,
		JobMemoryLimit:         getWPKEnv("MEMORY_LIMIT", "1Gi").(string),
		JobCPULimit:            getWPKEnv("CPU_LIMIT", "500m").(string), // half a cpu
		CommandRetries:         commandRetries,                          // half a cpu
		CommandRetryWait:       time.Duration(commandRetriesWait) * time.Second,
	}
}

// Reset parameters for the active run.
func (backend *KubeBackend) Reset() {
	backend.activeRun = &KubeBackendRun{
		Config:      &types.Config{},
		RunID:       CreateRandomID(10),
		StepLoggers: make(map[string]*KubeResourceLogger),
	}
}

// Name of the engine.
func (backend *KubeBackend) Name() string {
	return "kubectl"
}

// Check if the engine is available.
func (backend *KubeBackend) IsAvailable() bool {
	// check if the executable exists. Otherwise false.
	// May need connection check afterwards.
	_, err := exec.LookPath(backend.Client.GetExecutablePath())
	return err != nil
}

// Load the engine backend.
func (backend *KubeBackend) Load() error {
	err := backend.Client.Load()
	if err != nil {
		return err
	}
	return nil
}

// Setup the pipeline environment, by applying the templated
// setup yaml. Will consume cluster resources and would need
// to be cleaned up.
func (backend *KubeBackend) Setup(ctx context.Context, cfg *types.Config) error {
	logger := backend.MakeLogger(nil)
	logger.Debug().Msg("Creating active run setup")

	backend.Reset()

	// Load the configuration for the active run.
	err := backend.InitializeConfig(cfg)
	if err != nil {
		return err
	}

	logger.Debug().Msg("Active run reset and initialized")

	err = backend.Client.LoadDefaults(ctx)
	if err != nil {
		return err
	}

	logger.Debug().
		Str("Context", backend.Context()).
		Str("Namespace", backend.Namespace()).
		Msg("Loaded kube client defaults")

	setupYaml, err := backend.RenderSetupYaml()
	if err != nil {
		return err
	}

	output, err := backend.Client.DeployKubectlYamlWithContext(ctx, "apply", setupYaml, false)

	if err != nil {
		return err
	}

	logger.Debug().Msgf("Pipeline setup with:\n %s", output)
	logger.Info().
		Str("Context", backend.Context()).
		Str("Namespace", backend.Namespace()).
		Msg("Started pipeline execution")

	return nil
}

// Destroy the pipeline environment.
func (backend *KubeBackend) Destroy(_ context.Context, cfg *types.Config) error {
	logger := backend.MakeLogger(nil)
	logger.Debug().Msg("Destroying active run setup")

	destroyJobs := []*KubeJobTemplate{}
	if len(backend.activeRun.DetachedJobs) > 0 {
		destroyJobs = append(destroyJobs, backend.activeRun.DetachedJobs...)
		logger.Debug().Msgf("Destroying %d detached jobs", len(backend.activeRun.DetachedJobs))
	}

	if backend.activeRun.JobPendingWait != nil {
		destroyJobs = append(destroyJobs, backend.activeRun.JobPendingWait)
		logger.Debug().Msg("A job, pending wait, still exists. Destroying.")
	}

	yamlsToDeploy := []string{}

	setupYaml, err := backend.RenderSetupYaml()
	yamlsToDeploy = append(yamlsToDeploy, setupYaml)

	if err != nil {
		return err
	}

	for _, job := range destroyJobs {
		jobYaml, err := job.Render()
		if err != nil {
			return err
		}
		// adding to the destroy command
		yamlsToDeploy = append(yamlsToDeploy, jobYaml)
	}

	// stopping all step loggers
	for stepName, stepLogger := range backend.activeRun.StepLoggers {
		if stepLogger.IsRunning() {
			err := stepLogger.Stop()
			event := logger.Debug().Str("Step", stepName)
			if err != nil {
				event.Err(err)
			}
			event.Msgf("Terminated logger")
		}
	}

	// Destroy is different then other operations since it should be
	// always be attempted (even if the pipeline context is canceled).
	// It therefore executes in its own (background) context.
	output, err := backend.Client.DeployKubectlYaml(
		"delete",
		strings.Join(yamlsToDeploy, "\n---\n"),
		false,
	)

	if err != nil {
		logger.Error().Err(err).Msgf("Pipeline destruction failed")
		return err
	}

	logger.Debug().Msgf("Pipeline destroyed with:\n %s", output)

	logger.Info().
		Str("Context", backend.Context()).
		Str("Namespace", backend.Namespace()).
		Msg("Ended pipeline execution")

	return nil
}

// Exec the pipeline step.
func (backend *KubeBackend) Exec(ctx context.Context, step *types.Step) error {
	jobTemplate := KubeJobTemplate{
		Backend: backend,
		Step:    step,
	}

	logger := backend.MakeLogger(step)

	if step.Detached {
		step.Alias = Triary(
			len(step.Alias) > 0, step.Alias, ToKuberenetesValidName(step.Name, 50),
		).(string)
		logger.Debug().Msg("Starting detached job")
	} else {
		backend.activeRun.JobPendingWait = &jobTemplate
		logger.Debug().Msg("Job is pending")
	}

	jobAsYaml, err := jobTemplate.Render()
	if err != nil {
		return err
	}

	output, err := backend.Client.DeployKubectlYamlWithContext(
		ctx,
		"apply",
		jobAsYaml,
		false,
	)

	if err != nil {
		return err
	}

	logger.Debug().Msgf("Job initialized with\n %s", output)

	// wait for the job to start.
	logger.Debug().Msg("Waiting for job pod to be created")
	podNames, err := backend.GetJobPodName(ctx, &jobTemplate)
	if err != nil {
		return err
	}

	// the pod name, with the kind.
	podName := podNames[0]

	logger.Debug().Msg("Waiting for job pod to be ready")
	_, err = backend.Client.WaitForConditions(
		ctx,
		podName, []string{"ContainersReady", "Ready"},
		1,
	)

	if err != nil {
		return err
	}

	if step.Detached {
		backend.activeRun.DetachedJobs = append(backend.activeRun.DetachedJobs, &jobTemplate)

		logger.Debug().Msg("Reading detached pod info")
		err = backend.PopulateDetachedInfo(ctx, podName, &jobTemplate)
		if err != nil {
			return err
		}

		logger.Debug().Msg("Detached job configured")
	}
	return nil
}

// Tail the pipeline step logs.
func (backend *KubeBackend) Tail(ctx context.Context, step *types.Step) (io.ReadCloser, error) {
	jobTemplate := KubeJobTemplate{
		Backend: backend,
		Step:    step,
	}

	logger := backend.MakeLogger(step)
	stepLogger := &KubeResourceLogger{
		Backend:      backend,
		ResourceName: "job.batch/" + jobTemplate.JobName(),
	}

	// Used for destroy.
	backend.activeRun.StepLoggers[step.Name] = stepLogger

	logger.Debug().Msg("Reading logs")
	return stepLogger.Start(ctx)
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (backend *KubeBackend) Wait(ctx context.Context, step *types.Step) (*types.State, error) {
	jobTemplate := KubeJobTemplate{
		Backend: backend,
		Step:    step,
	}
	logger := backend.MakeLogger(step)

	// Assert and clear pending job
	if backend.activeRun.JobPendingWait.JobID() != jobTemplate.JobID() {
		return nil, errors.New(
			"Invalid wait on job. The job pending wait dose not match current step",
		)
	}
	backend.activeRun.JobPendingWait = nil

	logger.Info().
		Str("Context", backend.Context()).
		Str("Namespace", backend.Namespace()).
		Msg("Waiting for job to complete")

	condition, jobEndConditionError := backend.Client.WaitForConditions(
		ctx,
		"job/"+jobTemplate.JobName(),
		[]string{"Complete", "Failed"},
		1,
	)
	condition = Triary(len(condition) == 0, "Error", condition).(string)

	if jobEndConditionError == context.Canceled {
		logger.Debug().Msg("Step execution context canceled")
	} else if jobEndConditionError != nil {
		logger.Error().Err(jobEndConditionError).Msg("Error while waiting for job")
	}

	// From this point job has ended.
	logger.Debug().Msgf("Job ended with '%s'", condition)
	stepLogger := backend.activeRun.StepLoggers[step.Name]

	if stepLogger.IsRunning() {
		logger.Debug().Msg("Job ended but reader is still active. Waiting for logs.")
		select {
		case <-time.After(backend.Client.RequestTimeout):
			logger.Error().Msg(
				"Timed out waiting for logs to complete. Partial/empty logs",
			)
			break
		case <-stepLogger.Done():
			logger.Debug().Msg("Logger completed.")
			break
		}
	}

	err := stepLogger.Stop()
	if err != nil {
		logger.Err(err).Msg("Error(s) occurred while reading logs.")
	}

	// Checking status.
	forceDelete := false
	doDelete := false
	hasFailed := false
	switch condition {
	case "Complete":
		break
	case "Failed":
		hasFailed = true
	default:
		forceDelete = true
		hasFailed = true
	}

	if forceDelete {
		doDelete = true
	} else {
		switch backend.DeletePolicy {
		case IfFailed:
			doDelete = hasFailed
		case IfSucceeded:
			doDelete = !hasFailed
		case Always:
			doDelete = true
		}
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
		state.ExitCode = Triary(hasFailed, 1, 0).(int)
	}

	if doDelete {
		asJobYaml, err := jobTemplate.Render()
		if err != nil {
			return state, err
		}

		// deploy in new context since this must happen
		out, err := backend.Client.DeployKubectlYaml("delete", asJobYaml, true)
		if err != nil {
			return state, errors.New(out + ". " + err.Error())
		}

		logger.Debug().Str("DeletePolicy", backend.DeletePolicy).Msgf("Job DELETD with: \n" + out)
	} else {
		logger.Info().Str("DeletePolicy", backend.DeletePolicy).Msg(
			"Job artifact kept in cluster",
		)
	}

	logger.Info().
		Str("Context", backend.Context()).
		Str("Namespace", backend.Namespace()).
		Str("Status", condition).
		Str("Deleted", strconv.FormatBool(doDelete)).
		Msg("Job DONE")

	return state, nil
}
