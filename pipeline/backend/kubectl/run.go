package kubectl

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type KubePiplineRun struct {
	RunID          string                      // the random run id
	Backend        *KubeBackend                // The kubectl backend
	Config         *types.Config               // the run config
	PVCs           []*KubePVCTemplate          // Loaded pvc's (via setup)
	PVCByName      map[string]*KubePVCTemplate // Loaded pvc's by name
	SetupTemplates []KubeTemplate              // Loaded setup templates

	// Currently running steps
	ExecutingSteps map[string]*KubePiplineRunStep // A collection of jobs which have not reached the wait stage.

	stepMutex sync.Mutex // a locking mutex for concurrent opperations.
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

	destroyJobs := []*KubeJobTemplate{}

	// adding all jobs to destroy and
	for _, runStep := range run.ExecutingSteps {
		logger.Debug().Str("Step", runStep.Step.Name).Msgf("Destroying executing job")
		destroyJobs = append(destroyJobs, runStep.Job)
		if runStep.Logger.IsRunning() {
			err := runStep.Logger.Stop()
			logger.Debug().
				Str("Step", runStep.Step.Name).
				Err(err).
				Msgf("Stopped logger")
		}
	}

	yamlsToDeploy := []string{}
	setupYaml, err := run.RenderSetupYaml()
	if err != nil {
		return err
	}

	yamlsToDeploy = append(yamlsToDeploy, setupYaml)

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

	time.Sleep(time.Second * 10)

	return nil
}

// Exec the pipeline step.
func (run *KubePiplineRun) Exec(ctx context.Context, step *types.Step) error {
	logger := run.MakeLogger(step)
	runStep := run.CreateRunStep(step)

	if step.Detached {
		step.Alias = Triary(
			len(step.Alias) > 0, step.Alias, ToKuberenetesValidName(step.Name, 50),
		).(string)
		logger.Debug().Msg("Starting detached job")
	} else {
		logger.Debug().Msg("Job is pending")
	}

	jobAsYaml, err := runStep.Job.Render()
	if err != nil {
		return err
	}

	// first create the wait
	podWaiter := run.WaitForRunJobPod(ctx, runStep.Job)

	// deploy the job.
	output, err := run.Backend.Client.DeployKubectlYamlWithContext(
		ctx,
		"apply",
		jobAsYaml,
		false,
	)
	if err != nil {
		return err
	}

	logger.Debug().Msgf("Job applied with:\n%s", output)

	podWaitResult := <-podWaiter

	if podWaitResult.Error != nil {
		return podWaitResult.Error
	}

	logger.Debug().Msgf("Job running")

	if err != nil {
		return err
	}

	if step.Detached {
		// Unlike regular pods, which can execute async. We need this
		// pod to be ready before continue. Or timeout/cancel.
		logger.Debug().Msg("Waiting for detached pod to be ready")
		ready := <-run.Backend.Client.WaitForConditions(
			ctx,
			podWaitResult.PodName,
			[]string{"Ready"}, 1,
		)
		// check if ready.
		if ready.err != nil {
			return ready.err
		}

		logger.Debug().Msg("Reading detached pod info")
		err = run.PopulateDetachedInfo(ctx, podWaitResult.PodName, runStep.Job)
		if err != nil {
			return err
		}

		logger.Debug().Msg("Detached job ready")
	}

	runStep.Started = true

	return nil
}

// Tail the pipeline step logs.
func (run *KubePiplineRun) Tail(ctx context.Context, step *types.Step) (io.ReadCloser, error) {
	runStep := run.GetRunStep(step)
	logger := run.MakeLogger(step)
	logger.Debug().Msg("Reading logs")
	return runStep.Logger.Start(ctx)
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (run *KubePiplineRun) Wait(ctx context.Context, step *types.Step) (*types.State, error) {
	runStep := run.GetRunStep(step)
	logger := run.MakeLogger(step)

	logger.Info().
		Str("Context", run.Context()).
		Str("Namespace", run.Namespace()).
		Msg("Waiting for job to complete")

	jobEndCondition := <-run.Backend.Client.WaitForConditions(
		ctx,
		"job/"+runStep.Job.JobName(),
		[]string{"Complete", "Failed"},
		1,
	)

	condition := Triary(len(jobEndCondition.condition) == 0, "Error", jobEndCondition.condition).(string)

	if jobEndCondition.err == context.Canceled {
		logger.Debug().Msg("Step execution context canceled")
	} else if jobEndCondition.err != nil {
		logger.Error().Err(jobEndCondition.err).Msg("Error while waiting for job")
	}

	// From this point job has ended.
	logger.Debug().Msgf("Job ended with '%s'", condition)

	if runStep.Logger.IsRunning() {
		logger.Debug().Msg("Job ended but reader is still active. Waiting for logs.")
		select {
		case <-time.After(run.Backend.Client.RequestTimeout):
			logger.Error().Msg(
				"Timed out waiting for logs to complete. Partial/empty logs",
			)
			break
		case <-runStep.Logger.Done():
			logger.Debug().Msg("Logger completed.")
			break
		}
	}

	err := runStep.Logger.Stop()
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

	if jobEndCondition.err == nil {
		state.Exited = true
		state.ExitCode = Triary(hasFailed, 1, 0).(int)
	}

	if doDelete {
		asJobYaml, err := runStep.Job.Render()
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

	// Clear the current running step.
	run.DeleteRunStep(runStep)

	logger.Info().
		Str("Context", run.Context()).
		Str("Namespace", run.Namespace()).
		Str("Status", condition).
		Str("Deleted", strconv.FormatBool(doDelete)).
		Msg("Job DONE")

	return state, nil
}
