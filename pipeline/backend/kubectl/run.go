package kubectl

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type KubePiplineRun struct {
	Backend        *KubeBackend                   // The kubectl backend
	RunID          string                         // the random run id
	Config         *types.Config                  // the run config
	PVCs           []*KubePVCTemplate             // Loaded pvc's (via setup)
	PVCByName      map[string]*KubePVCTemplate    // Loaded pvc's by name
	SetupTemplates []KubeTemplate                 // Loaded setup templates
	DetachedJobs   []*KubeJobTemplate             // Loaded detached template (services, etc..)
	StepLoggers    map[string]*KubeResourceLogger // A collection of loggers per step.
	PendingJob     *KubeJobTemplate               // The executed job where wait was not initialized.
}

// Setup the pipeline environment, by applying the templated
// setup yaml. Will consume cluster resources and would need
// to be cleaned up.
func (run *KubePiplineRun) Setup(ctx context.Context, cfg *types.Config) error {
	err := run.InitializeConfig(cfg)
	if err != nil {
		return err
	}

	logger := run.MakeLogger(nil)
	logger.Debug().Msg("Run created")

	setupYaml, err := run.RenderSetupYaml()
	if err != nil {
		return err
	}

	output, err := run.Backend.Client.DeployKubectlYamlWithContext(
		ctx, "apply", setupYaml, false,
	)

	if err != nil {
		return err
	}

	logger.Debug().Msgf("Pipeline setup with:\n %s", output)
	logger.Info().
		Str("Context", run.Context()).
		Str("Namespace", run.Namespace()).
		Msg("Started pipeline execution")

	return nil
}

// Destroy the pipeline environment.
func (run *KubePiplineRun) Destroy(ctx context.Context, cfg *types.Config) error {
	logger := run.MakeLogger(nil)
	logger.Debug().Msg("Destroying active run setup")

	// stopping all step loggers
	for stepName, stepLogger := range run.StepLoggers {
		if stepLogger.IsRunning() {
			err := stepLogger.Stop()
			event := logger.Debug().Str("Step", stepName)
			if err != nil {
				event.Err(err)
			}
			event.Msgf("Terminated logger")
		}
	}

	destroyJobs := []*KubeJobTemplate{}
	if len(run.DetachedJobs) > 0 {
		destroyJobs = append(destroyJobs, run.DetachedJobs...)
		logger.Debug().Msgf("Destroying %d detached jobs", len(run.DetachedJobs))
	}

	if run.PendingJob != nil {
		destroyJobs = append(destroyJobs, run.PendingJob)
		logger.Debug().Msg("A job, pending wait, still exists. Destroying.")
	}

	yamlsToDeploy := []string{}

	setupYaml, err := run.RenderSetupYaml()
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

	// Destroy is different then other operations since it should be
	// always be attempted (even if the pipeline context is canceled).
	// It therefore executes in its own (background) context.
	output, err := run.Backend.Client.DeployKubectlYaml(
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
		Str("Context", run.Context()).
		Str("Namespace", run.Namespace()).
		Msg("Ended pipeline execution")

	return nil
}

// Exec the pipeline step.
func (run *KubePiplineRun) Exec(ctx context.Context, step *types.Step) error {
	logger := run.MakeLogger(step)

	jobTemplate := KubeJobTemplate{
		Run:  run,
		Step: step,
	}

	if step.Detached {
		step.Alias = Triary(
			len(step.Alias) > 0, step.Alias, ToKuberenetesValidName(step.Name, 50),
		).(string)
		logger.Debug().Msg("Starting detached job")
	} else {
		run.PendingJob = &jobTemplate
		logger.Debug().Msg("Job is pending")
	}

	jobAsYaml, err := jobTemplate.Render()
	if err != nil {
		return err
	}

	output, err := run.Backend.Client.DeployKubectlYamlWithContext(
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
	podNames, err := run.GetJobPodName(ctx, &jobTemplate)
	if err != nil {
		return err
	}

	// the pod name, with the kind.
	podName := podNames[0]

	logger.Debug().Msg("Waiting for job pod to be ready")
	_, err = run.Backend.Client.WaitForConditions(
		ctx,
		podName, []string{"ContainersReady", "Ready"},
		1,
	)

	if err != nil {
		return err
	}

	if step.Detached {
		run.DetachedJobs = append(run.DetachedJobs, &jobTemplate)

		logger.Debug().Msg("Reading detached pod info")
		err = run.PopulateDetachedInfo(ctx, podName, &jobTemplate)
		if err != nil {
			return err
		}

		logger.Debug().Msg("Detached job configured")
	}
	return nil
}

// Tail the pipeline step logs.
func (run *KubePiplineRun) Tail(ctx context.Context, step *types.Step) (io.ReadCloser, error) {
	logger := run.MakeLogger(step)
	jobTemplate := KubeJobTemplate{
		Run:  run,
		Step: step,
	}

	stepLogger := &KubeResourceLogger{
		Run:          run,
		ResourceName: "job.batch/" + jobTemplate.JobName(),
	}

	// Used for destroy.
	run.StepLoggers[step.Name] = stepLogger

	logger.Debug().Msg("Reading logs")
	return stepLogger.Start(ctx)
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (run *KubePiplineRun) Wait(ctx context.Context, step *types.Step) (*types.State, error) {
	logger := run.MakeLogger(step)
	jobTemplate := KubeJobTemplate{
		Run:  run,
		Step: step,
	}

	// Assert and clear pending job
	if run.PendingJob.JobID() != jobTemplate.JobID() {
		return nil, errors.New(
			"Invalid wait on job. The job pending wait dose not match current step",
		)
	}
	run.PendingJob = nil

	logger.Info().
		Str("Context", run.Context()).
		Str("Namespace", run.Namespace()).
		Msg("Waiting for job to complete")

	condition, jobEndConditionError := run.Backend.Client.WaitForConditions(
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
	stepLogger := run.StepLoggers[step.Name]

	if stepLogger.IsRunning() {
		logger.Debug().Msg("Job ended but reader is still active. Waiting for logs.")
		select {
		case <-time.After(run.Backend.Client.RequestTimeout):
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
		switch run.Backend.DeletePolicy {
		case IfFailed:
			doDelete = hasFailed
		case IfSucceeded:
			doDelete = !hasFailed
		case Always:
			doDelete = true
		case Never:
			doDelete = false
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
		out, err := run.Backend.Client.DeployKubectlYaml("delete", asJobYaml, true)
		if err != nil {
			return state, errors.New(out + ". " + err.Error())
		}

		logger.Debug().
			Str("DeletePolicy", run.Backend.DeletePolicy).
			Msgf("Job DELETD with: \n" + out)
	} else {
		logger.Info().
			Str("DeletePolicy", run.Backend.DeletePolicy).
			Msg("Job artifact kept in cluster")
	}

	logger.Info().
		Str("Context", run.Context()).
		Str("Namespace", run.Namespace()).
		Str("Status", condition).
		Str("Deleted", strconv.FormatBool(doDelete)).
		Msg("Job DONE")

	return state, nil
}
