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
func (run *KubePiplineRun) MakeLogger(step *types.Step) *zerolog.Logger {
	context := log.With().Str("RunID", run.RunID)

	if step != nil {
		context = context.Str("Step", step.Name)
	}

	logger := context.Logger()
	return &logger
}

// Initializes the configuration for the kube backend
// and populates the basic parameters for that config.
func (run *KubePiplineRun) InitializeConfig(cfg *types.Config) error {
	run.Config = cfg

	// Resetting
	run.SetupTemplates = []KubeTemplate{}

	// add network policy
	if run.Backend.EnableRunNetworkPolicy {
		run.SetupTemplates = append(run.SetupTemplates, &KubeNetworkPolicyTemplate{
			Run: run,
		})
	}

	run.PVCs = []*KubePVCTemplate{}
	run.PVCByName = make(map[string]*KubePVCTemplate)

	for _, vol := range run.Config.Volumes {
		pvc := &KubePVCTemplate{
			Run:  run,
			Name: vol.Name,
		}
		run.PVCs = append(run.PVCs, pvc)
		run.PVCByName[vol.Name] = pvc
		run.SetupTemplates = append(run.SetupTemplates, pvc)
	}
	return nil
}

// Renders the setup yaml.
func (run *KubePiplineRun) RenderSetupYaml() (string, error) {
	var templatesAsYaml []string

	for _, template := range run.SetupTemplates {
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
func (run *KubePiplineRun) GetJobPodName(
	ctx context.Context,
	jobTemplate *KubeJobTemplate,
) ([]string, error) {
	var podNames []string
	var err error
	for {
		podNames, err = run.Backend.Client.GetResourceNames(
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
func (run *KubePiplineRun) PopulateDetachedInfo(
	ctx context.Context,
	podName string,
	jobTemplate *KubeJobTemplate,
) error {
	logger := run.MakeLogger(jobTemplate.Step).With().Str("PodName", podName).Logger()
	attempts := 0
	for {
		podIP, err := run.Backend.Client.RunKubectlCommand(
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

			if attempts > run.Backend.CommandRetries {
				logger.Error().Err(err).Msg(
					"Max number of retries found whist attempting to retrieve pod IP. Aborted",
				)
				return err
			}

			logger.Debug().Err(err).Msgf(
				"Failed to retrieve detached info, pod may not be ready. Retry in %.2f [seconds]",
				run.Backend.CommandRetryWait.Seconds(),
			)

			time.Sleep(run.Backend.CommandRetryWait)
			continue
		}

		jobTemplate.DetachedPodIP = podIP
		break
	}
	return nil
}
