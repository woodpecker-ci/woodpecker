package kubectl

import (
	"context"
	"io"
	"os/exec"
	"strconv"
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

type KubeBackend struct {
	Client              *KubeClient   // The kubernetes client
	DeletePolicy        string        // The job delete policy
	JobMemoryLimit      string        // The runner container memory limit (1Gi)
	JobCPULimit         string        // The runner container cpu limit (200m)
	PVCStorageSize      string        // The storage size for the PVC.
	PVCAccessMode       string        // The access mode for PVC's
	PVCStorageClassName string        // The pvc storage class name.
	ForcePullPolicy     string        // Forces a pull policy on all jobs
	CommandRetries      int           // The number of times to retry commands.
	CommandRetryWait    time.Duration // The wait time between command retries.

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

	// there could be multiple active runs. Therefore we need somehow to
	ActiveRuns map[context.Context]*KubeBackendRun // The current kubectl engine active run.
}

var _ types.Engine = &KubeBackend{}

// Create a new kubectl (exec based) engine. Allows for execution pods
// as commands. Assumes running inside a cluster or kubectl is configured.
func New() types.Engine {
	requestTimeoutSeconds, _ := strconv.ParseFloat(getWPKEnv("REQUEST_TIMEOUT", "10").(string), 64)
	client := &KubeClient{
		Executable:     getWPKEnv("EXECUTABLE", "kubectl").(string),
		Namespace:      getWPKEnv("NAMESPACE", "").(string),
		Context:        getWPKEnv("CONTEXT", "").(string),
		RequestTimeout: time.Duration(requestTimeoutSeconds) * time.Second,

		// TODO: There is an error that dose not allow these type of setting
		// whilst running in-cluster. I set this default to false, but can be
		// changed for newer version of kubectl.
		// ERROR: https://github.com/kubernetes/kubernetes/issues/93474
		AllowKubectlClientConfiguration: getWPKEnv("ALLOW_KUBECTL_CLIENT_CONFIG", "false").(string) == "true",
	}

	containerStartDelaySeconds, _ := strconv.ParseInt(getWPKEnv("CONTAINER_START_DELAY", "1").(string), 0, 64)
	terminationGracePeriodSeconds, _ := strconv.ParseInt(getWPKEnv("TERMINATION_GRACE_PERIOD", "5").(string), 0, 64)
	commandRetries, _ := strconv.Atoi(getWPKEnv("COMMAND_RETRIES", "5").(string))
	commandRetriesWait, _ := strconv.ParseFloat(getWPKEnv("COMMAND_RETRIES_WAIT", "1").(string), 64)

	return &KubeBackend{
		Client:       client,
		DeletePolicy: getWPKEnv("DELETE_POLICY", Always).(string),

		PVCAllowOnDetached:  getWPKEnv("PVC_ALLOW_ON_DETACHED", "false").(string) == "true",
		PVCStorageSize:      getWPKEnv("PVC_STORAGE_SIZE", "1Gi").(string),
		PVCAccessMode:       getWPKEnv("PVC_ACCESS_MODE", "ReadWriteOnce").(string),
		PVCStorageClassName: getWPKEnv("PVC_STORAGE_CLASS", "").(string),

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

// Create parameters for the active run.
func (backend *KubeBackend) CreateRun() *KubeBackendRun {
	run := &KubeBackendRun{
		Backend:     backend,
		Config:      &types.Config{},
		RunID:       CreateRandomID(10),
		StepLoggers: make(map[string]*KubeResourceLogger),
	}
	return run
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	logger := log.Ctx(ctx)

	err := backend.Client.Load(ctx)
	if err != nil {
		cancel()
		return err
	}

	logger.Debug().
		Str("Context", backend.Client.Context).
		Str("Namespace", backend.Client.Namespace).
		Msg("Kubernetes client loaded")

	cancel()
	return nil
}

// Setup the pipeline environment.
func (backend *KubeBackend) Setup(ctx context.Context, cfg *types.Config) error {
	backend.ActiveRuns[ctx] = backend.CreateRun()
	return backend.ActiveRuns[ctx].Setup(ctx, cfg)
}

// Exec start the pipeline step.
func (backend *KubeBackend) Exec(ctx context.Context, step *types.Step) error {
	return backend.ActiveRuns[ctx].Backend.Exec(ctx, step)
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (backend *KubeBackend) Wait(ctx context.Context, step *types.Step) (*types.State, error) {
	return backend.ActiveRuns[ctx].Backend.Wait(ctx, step)
}

// Tail the pipeline step logs.
func (backend *KubeBackend) Tail(ctx context.Context, step *types.Step) (io.ReadCloser, error) {
	return backend.ActiveRuns[ctx].Backend.Tail(ctx, step)
}

// Destroy the pipeline environment.
func (backend *KubeBackend) Destroy(ctx context.Context, cfg *types.Config) error {
	run := backend.ActiveRuns[ctx]
	delete(backend.ActiveRuns, ctx)
	return run.Destroy(ctx, cfg)
}
