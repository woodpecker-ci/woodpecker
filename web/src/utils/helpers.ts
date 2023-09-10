import { Pipeline, PipelineStep, PipelineWorkflow, Repo } from '~/lib/api/types';

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
 * Returns true if the process is in a completed state.
 *
 * @param {Object} step - The process object.
 * @returns {boolean}
 */
export function isStepFinished(step: PipelineStep): boolean {
  return step.state !== 'running' && step.state !== 'pending';
}

/**
 * Returns true if the process is running.
 *
 * @param {Object} step - The process object.
 * @returns {boolean}
 */
export function isStepRunning(step: PipelineStep): boolean {
  return step.state === 'running';
}

/**
 * Compare two pipelines by creation timestamp.
 * @param {Object} a - A pipeline.
 * @param {Object} b - A pipeline.
 * @returns {number}
 */
export function comparePipelines(a: Pipeline, b: Pipeline): number {
  return (b.created_at || -1) - (a.created_at || -1);
}

export function isPipelineActive(pipeline: Pipeline): boolean {
  return ['pending', 'running', 'started'].includes(pipeline.status);
}

export function repoSlug(ownerOrRepo: Repo): string;
export function repoSlug(ownerOrRepo: string, name: string): string;
export function repoSlug(ownerOrRepo: string | Repo, name?: string): string {
  if (typeof ownerOrRepo === 'string') {
    if (!name) {
      throw new Error('Please provide a name as well');
    }

    return `${ownerOrRepo}/${name}`;
  }

  return `${ownerOrRepo.owner}/${ownerOrRepo.name}`;
}
