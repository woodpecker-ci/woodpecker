package kubectl

import (
	"context"
	"errors"
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
		context = context.
			Str("Step", step.Name)
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

type WaitForRunJobPodResult struct {
	PodName string
	Error   error
}

// Returns the pod name for a job. Will wait for
// the pod to be ready.
func (run *KubePiplineRun) WaitForRunJobPod(
	ctx context.Context,
	jobTemplate *KubeJobTemplate,
) chan struct {
	PodName string
	Error   error
} {
	waitContext, waitContextCancel := context.WithCancel(ctx)
	result := make(chan struct {
		PodName string
		Error   error
	})
	waitForCommandToStart := make(chan bool)
	completed := false
	podName := ""

	stop := func(err error) {
		waitForCommandToStart <- false
		waitContextCancel()
		if completed {
			return
		}
		completed = true
		result <- struct {
			PodName string
			Error   error
		}{
			PodName: podName,
			Error:   err,
		}
	}

	// Catch cancelled context.
	go func() {
		<-waitContext.Done()
		if !completed {
			stop(errors.New("Context cancelled"))
		}
	}()

	go func() {
		eventsChan := run.Backend.Client.WaitForResourceEvents(
			waitContext,
			fmt.Sprintf(`^%s.*$`, jobTemplate.JobName()),
			[]string{"Started", "BackOff"},
			1,
		)

		waitForCommandToStart <- true

		// wait for the events.
		matchedEvents := <-eventsChan
		event := matchedEvents.events[0]

		if matchedEvents.err != nil {
			stop(matchedEvents.err)
			return
		} else if event == "BackOff" {
			stop(errors.New("Received pull backoff from executing pod, execution error"))
			return
		}

		podNames, err := run.Backend.Client.GetResourceNames(
			waitContext,
			"pod",
			"woodpecker-job-id="+jobTemplate.JobID(),
		)

		if err != nil {
			stop(err)
			return
		}

		podName = podNames[0]

		// waiting until the pod is ready.
		ready := <-run.Backend.Client.WaitForConditions(
			waitContext,
			podName, []string{"ContainersReady", "Ready"},
			1,
		)

		stop(ready.err)
	}()

	return result
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
		podIP, err := run.Backend.Client.GetPodIP(ctx, podName)
		if err != nil {
			if err == context.Canceled {
				logger.Debug().Err(err).Msg("Aborted reading pod IP")
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
