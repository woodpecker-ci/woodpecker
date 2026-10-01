import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { Pipeline, PipelineFeed } from '~/lib/api/types';

import { usePipelineStore } from './pipelines';

const mocks = vi.hoisted(() => ({
  getPipeline: vi.fn(),
  getPipelineFeed: vi.fn(),
  loadRepos: vi.fn(),
}));

vi.mock('~/compositions/useApiClient', () => ({
  default: () => ({
    getPipeline: mocks.getPipeline,
    getPipelineFeed: mocks.getPipelineFeed,
  }),
}));

vi.mock('~/store/repos', () => ({
  useRepoStore: () => ({
    loadRepos: mocks.loadRepos,
    ownedRepoIds: [1],
    repos: new Map(),
    setRepo: vi.fn(),
  }),
}));

describe('pipeline store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  it('refreshes locally active pipelines missing from the server feed', async () => {
    const store = usePipelineStore();
    const runningPipeline = { id: 10, number: 3, status: 'running' } as Pipeline;
    const completedPipeline = { ...runningPipeline, status: 'success' } as Pipeline;
    store.setPipeline(1, runningPipeline);
    mocks.getPipelineFeed.mockResolvedValue([] as PipelineFeed[]);
    mocks.getPipeline.mockResolvedValue(completedPipeline);

    await store.loadPipelineFeed();

    expect(mocks.getPipeline).toHaveBeenCalledWith(1, 3);
    expect(store.activePipelines).toHaveLength(0);
    expect(store.pipelines.get(1)?.get(3)?.status).toBe('success');
  });
});
