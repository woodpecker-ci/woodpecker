import { describe, expect, it } from 'vitest';

import type { Pipeline, PipelineStatus, PipelineWorkflow } from './api/types';
import { pipelineCanRestartFailed, reusedWorkflows } from './pipeline';

function pipelineWith(status: PipelineStatus, workflowStates: PipelineStatus[]): Pipeline {
  return {
    status,
    workflows: workflowStates.map((state, i) => ({ id: i, name: `wf-${i}`, state }) as PipelineWorkflow),
  } as Pipeline;
}

describe('pipelineCanRestartFailed', () => {
  it('should be false without a pipeline', () => {
    expect(pipelineCanRestartFailed(undefined)).toBe(false);
  });

  it('should be false when all workflows passed', () => {
    expect(pipelineCanRestartFailed(pipelineWith('success', ['success', 'skipped']))).toBe(false);
  });

  it.each<PipelineStatus>(['failure', 'error', 'killed'])('should be true for a %s workflow', (state) => {
    expect(pipelineCanRestartFailed(pipelineWith('failure', ['success', state]))).toBe(true);
  });

  it.each<PipelineStatus>(['killed', 'canceled'])('should be true for a %s pipeline', (status) => {
    expect(pipelineCanRestartFailed(pipelineWith(status, ['success', 'skipped']))).toBe(true);
  });

  it('should be false for a pipeline without workflows', () => {
    expect(pipelineCanRestartFailed(pipelineWith('error', []))).toBe(false);
  });

  it.each<PipelineStatus>(['pending', 'running', 'blocked'])('should be false while the pipeline is %s', (status) => {
    expect(pipelineCanRestartFailed(pipelineWith(status, ['failure', 'running']))).toBe(false);
  });
});

describe('reusedWorkflows', () => {
  function workflow(
    id: number,
    name: string,
    state: PipelineStatus,
    environ?: Record<string, string>,
  ): PipelineWorkflow {
    return { id, pipeline_id: 0, pid: id, name, state, environ };
  }
  function run(number: number, workflows: PipelineWorkflow[], partial = false): Pipeline {
    return { number, parent: number - 1, partial_restart: partial, workflows } as Pipeline;
  }

  it('should be empty for a pipeline that is no partial restart', () => {
    const full = run(4, [workflow(10, 'test', 'success')]);
    expect(reusedWorkflows(full, [run(3, [workflow(1, 'build', 'success')])])).toEqual([]);
  });

  it('should take the workflows that were not rerun from the parent', () => {
    const parent = run(3, [
      workflow(1, 'build', 'success'),
      workflow(2, 'test', 'failure'),
      workflow(3, 'lint', 'success'),
    ]);
    const restart = run(4, [workflow(10, 'test', 'success')], true);
    expect(reusedWorkflows(restart, [parent]).map((r) => [r.workflow.id, r.pipelineNumber])).toEqual([
      [1, 3],
      [3, 3],
    ]);
  });

  it('should leave out a workflow that did not pass and was removed from the config', () => {
    const parent = run(3, [workflow(1, 'build', 'success'), workflow(2, 'old-check', 'failure')]);
    const restart = run(4, [], true);
    expect(reusedWorkflows(restart, [parent]).map((r) => r.workflow.id)).toEqual([1]);
  });

  it('should match matrix variants by their axis', () => {
    const parent = run(3, [
      workflow(1, 'test', 'success', { GO: '1.25' }),
      workflow(2, 'test', 'failure', { GO: '1.26' }),
    ]);
    const restart = run(4, [workflow(10, 'test', 'success', { GO: '1.26' })], true);
    expect(reusedWorkflows(restart, [parent]).map((r) => r.workflow.id)).toEqual([1]);
  });

  it('should follow a chain of partial restarts up to the first full run', () => {
    const full = run(3, [
      workflow(1, 'build', 'success'),
      workflow(2, 'test', 'failure'),
      workflow(3, 'lint', 'failure'),
    ]);
    const first = run(4, [workflow(10, 'test', 'failure'), workflow(11, 'lint', 'success')], true);
    const second = run(5, [workflow(20, 'test', 'success')], true);
    const older = run(2, [workflow(0, 'build', 'failure')]);
    expect(reusedWorkflows(second, [first, full, older]).map((r) => [r.workflow.id, r.pipelineNumber])).toEqual([
      [11, 4],
      [1, 3],
    ]);
  });
});
