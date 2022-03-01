package kubectl

import "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"

type KubePiplineRunStep struct {
	Job     *KubeJobTemplate    // The step job configuration
	Logger  *KubeResourceLogger // The step active logger.
	Step    *types.Step
	Started bool // If true, the detached job has been started.
}

// Returns the collection of active detached jobs,
// If the detached job has not been started, ignore.
func (run *KubePiplineRun) DetachedJobs() []*KubeJobTemplate {
	jobs := []*KubeJobTemplate{}
	run.stepMutex.Lock()
	for _, runStep := range run.ExecutingSteps {
		if runStep.Step.Detached && runStep.Started {
			jobs = append(jobs, runStep.Job)
		}
	}
	run.stepMutex.Unlock()
	return jobs
}

func (run *KubePiplineRun) CreateRunStep(step *types.Step) *KubePiplineRunStep {
	job := KubeJobTemplate{
		Run:  run,
		Step: step,
	}
	runStep := &KubePiplineRunStep{
		Job: &job,
		Logger: &KubeResourceLogger{
			Run:          run,
			ResourceName: "job.batch/" + job.JobName(),
		},
		Step: step,
	}

	run.stepMutex.Lock()
	run.ExecutingSteps[step.Name] = runStep
	run.stepMutex.Unlock()

	return runStep
}

func (run *KubePiplineRun) GetRunStep(step *types.Step) *KubePiplineRunStep {
	run.stepMutex.Lock()
	runStep := run.ExecutingSteps[step.Name]
	run.stepMutex.Unlock()
	return runStep
}

func (run *KubePiplineRun) DeleteRunStep(runStep *KubePiplineRunStep) {
	run.stepMutex.Lock()
	delete(run.ExecutingSteps, runStep.Step.Name)
	run.stepMutex.Unlock()
}
