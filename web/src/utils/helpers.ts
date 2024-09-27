import type { Pipeline, PipelineStep, PipelineWorkflow, Repo } from '~/lib/api/types';

export function findStep(workflows: PipelineWorkflow[], pid: number): PipelineStep | undefined {
  return workflows.reduce(
    (prev, workflow) => {
      const result = workflow.children.reduce(
        (prevChild, step) => {
          if (step.pid === pid) {
            return step;
          }

          return prevChild;
        },
        undefined as PipelineStep | undefined,
      );
      if (result) {
        return result;
      }

      return prev;
    },
    undefined as PipelineStep | undefined,
  );
}

/**
 * @param {object} step - The process object.
 * @returns {boolean} true if the process is in a completed state
 */
export function isStepFinished(step: PipelineStep): boolean {
  return step.state !== 'running' && step.state !== 'pending';
}

/**
 * @param {object} step - The process object.
 * @returns {boolean} true if the process is running
 */
export function isStepRunning(step: PipelineStep): boolean {
  return step.state === 'running';
}

/**
 * Compare two pipelines by creation timestamp.
 * @param {object} a - A pipeline.
 * @param {object} b - A pipeline.
 * @returns {number} 0 if created at the same time, < 0 if b was create before a, > 0 otherwise
 */
export function comparePipelines(a: Pipeline, b: Pipeline): number {
  return (b.created || -1) - (a.created || -1);
}

/**
 * Compare two pipelines by the status.
 * Giving pending, running, or started higher priority than other status
 * @param {object} a - A pipeline.
 * @param {object} b - A pipeline.
 * @returns {number} 0 if status same priority, < 0 if b has higher priority, > 0 otherwise
 */
export function comparePipelinesWithStatus(a: Pipeline, b: Pipeline): number {
  const bPriority = ['pending', 'running', 'started'].includes(b.status) ? 1 : 0;
  const aPriority = ['pending', 'running', 'started'].includes(a.status) ? 1 : 0;
  return bPriority - aPriority || comparePipelines(a, b);
}

export function isPipelineActive(pipeline: Pipeline): boolean {
  return ['pending', 'running', 'started'].includes(pipeline.status);
}

export function repoSlug(ownerOrRepo: string | Repo, name?: string): string {
  if (typeof ownerOrRepo === 'string') {
    if (name === undefined) {
      throw new Error('Please provide a name as well');
    }

    return `${ownerOrRepo}/${name}`;
  }

  return `${ownerOrRepo.owner}/${ownerOrRepo.name}`;
}
