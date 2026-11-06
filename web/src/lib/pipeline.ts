import type { Pipeline, PipelineStatus, PipelineWorkflow } from '~/lib/api/types';

export function workflowsWithErrors(pipeline?: Pipeline): PipelineWorkflow[] {
  return pipeline?.workflows?.filter((workflow) => workflow.error !== undefined && workflow.error !== '') ?? [];
}

export function anyStepStarted(pipeline?: Pipeline): boolean {
  return (
    pipeline?.workflows?.some((workflow) =>
      workflow.children?.some((step) => step.started !== undefined && step.started > 0),
    ) ?? false
  );
}

export function pipelineHasNonWarningErrors(pipeline?: Pipeline): boolean {
  return pipeline?.errors?.some((e) => !e.is_warning) ?? false;
}

export function pipelineHasErrorsToShow(pipeline?: Pipeline): boolean {
  return pipelineHasNonWarningErrors(pipeline) || workflowsWithErrors(pipeline).length > 0;
}

const failedWorkflowStates: PipelineStatus[] = ['failure', 'error', 'killed'];
const unfinishedPipelineStates: PipelineStatus[] = ['blocked', 'pending', 'running'];
const canceledPipelineStates: PipelineStatus[] = ['killed', 'canceled'];

// pipelineCanRestartFailed tells whether a pipeline has workflows to restart on their
// own: it finished with failed workflows, or it was canceled.
export function pipelineCanRestartFailed(pipeline?: Pipeline): boolean {
  if (
    pipeline?.workflows === undefined ||
    pipeline.workflows.length === 0 ||
    unfinishedPipelineStates.includes(pipeline.status)
  ) {
    return false;
  }
  return (
    canceledPipelineStates.includes(pipeline.status) ||
    pipeline.workflows.some((workflow) => failedWorkflowStates.includes(workflow.state))
  );
}

export interface ReusedWorkflow {
  workflow: PipelineWorkflow;
  // The number of the pipeline the workflow ran in.
  pipelineNumber: number;
}

function workflowKey(workflow: PipelineWorkflow): string {
  const environ = Object.entries(workflow.environ ?? {}).sort(([a], [b]) => a.localeCompare(b));
  return JSON.stringify([workflow.name, environ]);
}

// reusedWorkflows returns the workflows a partial restart did not rerun, taken from
// its ancestors (parent first). A partial restart of a partial restart reuses
// workflows from further up the chain, up to the first full run. Only workflows that
// passed are taken over; one that did not pass and is missing from the restart was
// removed from the config.
export function reusedWorkflows(pipeline: Pipeline, ancestors: Pipeline[]): ReusedWorkflow[] {
  const reused: ReusedWorkflow[] = [];
  if (!pipeline.partial_restart) {
    return reused;
  }

  const covered = new Set((pipeline.workflows ?? []).map(workflowKey));
  for (const ancestor of ancestors) {
    for (const workflow of ancestor.workflows ?? []) {
      const key = workflowKey(workflow);
      if (covered.has(key)) {
        continue;
      }
      covered.add(key);
      if (workflow.state === 'success') {
        reused.push({ workflow, pipelineNumber: ancestor.number });
      }
    }
    if (!ancestor.partial_restart) {
      break;
    }
  }
  return reused;
}
