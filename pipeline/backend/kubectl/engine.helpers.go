package kubectl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

// Create a new logger context for the step.
func (backend *KubeBackend) MakeLogger(step *types.Step) *zerolog.Logger {
	context := log.With()
	if backend.activeRun != nil {
		context = context.Str("RunID", backend.activeRun.RunID)
	}
	if step != nil {
		context = context.Str("Step", step.Name)
	}

	logger := context.Logger()
	return &logger
}

// Initializes the configuration for the kube backend
// and populates the basic parameters for that config.
func (backend *KubeBackend) InitializeConfig(cfg *types.Config) error {
	backend.activeRun.Config = cfg

	// Resetting
	backend.activeRun.SetupTemplates = []KubeTemplate{}

	// add network policy
	if backend.EnableRunNetworkPolicy {
		backend.activeRun.SetupTemplates = append(backend.activeRun.SetupTemplates, &KubeNetworkPolicyTemplate{
			Backend: backend,
		})
	}

	backend.activeRun.PVCs = []*KubePVCTemplate{}
	backend.activeRun.PVCByName = make(map[string]*KubePVCTemplate)

	for _, vol := range backend.activeRun.Config.Volumes {
		pvc := &KubePVCTemplate{
			Backend: backend,
			Name:    vol.Name,
		}
		backend.activeRun.PVCs = append(backend.activeRun.PVCs, pvc)
		backend.activeRun.PVCByName[vol.Name] = pvc
		backend.activeRun.SetupTemplates = append(backend.activeRun.SetupTemplates, pvc)
	}
	return nil
}

// Renders the setup yaml.
func (backend *KubeBackend) RenderSetupYaml() (string, error) {
	var templatesAsYaml []string

	for _, template := range backend.activeRun.SetupTemplates {
		asYaml, err := template.Render()
		if err != nil {
			return "", err
		}
		templatesAsYaml = append(templatesAsYaml, asYaml)
	}

	return strings.Join(templatesAsYaml, "\n---\n"), nil
}

// Returns the pod name for a job. Will wait for
// the pod to be ready.
func (backend *KubeBackend) GetJobPodName(
	ctx context.Context,
	jobTemplate *KubeJobTemplate,
) ([]string, error) {
	var podNames []string
	var err error
	for {
		podNames, err = backend.Client.GetResourceNames(
			ctx,
			"pod",
			"woodpecker-job-id="+jobTemplate.JobID(),
		)

		if err != nil {
			return []string{}, err
		}

		if len(podNames) == 0 {
			continue
		}

		break
	}
	return podNames, nil
}

// Populates detached info for an executing job pod (like ip)
// Allows for alias naming and detached service access.
func (backend *KubeBackend) PopulateDetachedInfo(
	ctx context.Context,
	podName string,
	jobTemplate *KubeJobTemplate,
) error {
	logger := backend.MakeLogger(jobTemplate.Step).With().Str("PodName", podName).Logger()
	attempts := 0
	for {
		podIP, err := backend.Client.RunKubectlCommand(
			ctx, "get", podName,
			"-o",
			"custom-columns=:status.podIP",
		)

		podIP = strings.TrimSpace(podIP)
		isInvalid := err == nil && !IsIP(podIP)
		attempts++

		if err == context.Canceled {
			logger.Debug().Err(err).Msg("Aborted reading pod IP")
		}

		if err != nil || isInvalid {
			if isInvalid {
				err = fmt.Errorf(
					"Invalid IP returned: %s", podIP,
				)
			}

			if attempts > backend.CommandRetries {
				logger.Error().Err(err).Msg(
					"Max number of retries found whist attempting to retrieve pod IP. Aborted",
				)
				return err
			}

			logger.Debug().Err(err).Msgf(
				"Failed to retrieve detached info, pod may not be ready. Retry in %.2f [seconds]",
				backend.CommandRetryWait.Seconds(),
			)

			time.Sleep(backend.CommandRetryWait)
			continue
		}

		jobTemplate.DetachedPodIP = podIP
		break
	}
	return nil
}
