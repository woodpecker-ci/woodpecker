import type { Pipeline, PipelineWorkflow } from '~/lib/api/types';

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

export function hasHardParseErrors(pipeline?: Pipeline): boolean {
  return pipeline?.errors?.some((e) => !e.is_warning) ?? false;
}

export function pipelineHasErrorsToShow(pipeline?: Pipeline): boolean {
  return hasHardParseErrors(pipeline) || workflowsWithErrors(pipeline).length > 0;
}
