import { describe, expect, it } from 'vitest';

import type { Pipeline } from '~/lib/api/types';

import { anyStepStarted, pipelineHasErrorsToShow, workflowsWithErrors } from './pipeline';

function fakePipeline(overrides: Partial<Pipeline> = {}): Pipeline {
  return {
    id: 1,
    number: 1,
    ...overrides,
  } as Pipeline;
}

describe('workflowsWithErrors', () => {
  it('returns empty list for pipeline without workflows', () => {
    expect(workflowsWithErrors(fakePipeline())).toEqual([]);
  });

  it('returns empty list when no workflow has an error', () => {
    const pipeline = fakePipeline({
      workflows: [
        { id: 1, pipeline_id: 1, pid: 1, name: 'test', state: 'success', children: [] },
        { id: 2, pipeline_id: 1, pid: 2, name: 'build', state: 'success', children: [] },
      ],
    });
    expect(workflowsWithErrors(pipeline)).toEqual([]);
  });

  it('returns only workflows with a non-empty error', () => {
    const pipeline = fakePipeline({
      workflows: [
        { id: 1, pipeline_id: 1, pid: 1, name: 'test', state: 'failure', children: [], error: 'boom' },
        { id: 2, pipeline_id: 1, pid: 2, name: 'build', state: 'success', children: [], error: '' },
        { id: 3, pipeline_id: 1, pid: 3, name: 'deploy', state: 'failure', children: [], error: 'network gone' },
      ],
    });
    expect(workflowsWithErrors(pipeline).map((w) => w.name)).toEqual(['test', 'deploy']);
  });
});

describe('pipelineHasErrorsToShow', () => {
  it('is false for clean pipeline', () => {
    expect(pipelineHasErrorsToShow(fakePipeline())).toBe(false);
  });

  it('is false when pipeline only has warnings', () => {
    const pipeline = fakePipeline({
      errors: [{ type: 'linter', message: 'meh', is_warning: true }],
    });
    expect(pipelineHasErrorsToShow(pipeline)).toBe(false);
  });

  it('is true for non-warning pipeline errors', () => {
    const pipeline = fakePipeline({
      errors: [{ type: 'compiler', message: 'bad yaml', is_warning: false }],
    });
    expect(pipelineHasErrorsToShow(pipeline)).toBe(true);
  });

  it('is true when a workflow has a runtime error', () => {
    const pipeline = fakePipeline({
      workflows: [{ id: 1, pipeline_id: 1, pid: 1, name: 'test', state: 'failure', children: [], error: 'boom' }],
    });
    expect(pipelineHasErrorsToShow(pipeline)).toBe(true);
  });

  it('is true when both warnings and workflow errors exist', () => {
    const pipeline = fakePipeline({
      errors: [{ type: 'linter', message: 'meh', is_warning: true }],
      workflows: [{ id: 1, pipeline_id: 1, pid: 1, name: 'test', state: 'failure', children: [], error: 'boom' }],
    });
    expect(pipelineHasErrorsToShow(pipeline)).toBe(true);
  });
});

describe('anyStepStarted', () => {
  it('is false for pipeline without workflows', () => {
    expect(anyStepStarted(fakePipeline())).toBe(false);
  });

  it('is false when no step has started', () => {
    const pipeline = fakePipeline({
      workflows: [
        {
          id: 1,
          pipeline_id: 1,
          pid: 1,
          name: 'test',
          state: 'failure',
          error: 'setup failed',
          children: [
            { id: 10, uuid: 'u10', pipeline_id: 1, pid: 10, ppid: 1, name: 'step-0', state: 'pending', exit_code: 0 },
          ],
        },
      ],
    });
    expect(anyStepStarted(pipeline)).toBe(false);
  });

  it('is true when a step has started (e.g. error during cleanup)', () => {
    const pipeline = fakePipeline({
      workflows: [
        {
          id: 1,
          pipeline_id: 1,
          pid: 1,
          name: 'test',
          state: 'failure',
          error: 'cleanup failed',
          children: [
            {
              id: 10,
              uuid: 'u10',
              pipeline_id: 1,
              pid: 10,
              ppid: 1,
              name: 'step-0',
              state: 'success',
              started: 12345,
              exit_code: 0,
            },
          ],
        },
      ],
    });
    expect(anyStepStarted(pipeline)).toBe(true);
  });
});
