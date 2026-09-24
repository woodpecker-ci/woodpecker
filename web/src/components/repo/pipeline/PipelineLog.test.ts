import { shallowMount } from '@vue/test-utils';
import { describe, expect, it, vi } from 'vitest';
import { ref } from 'vue';
import { createI18n } from 'vue-i18n';

import PipelineLog from '~/components/repo/pipeline/PipelineLog.vue';
import type { Pipeline, PipelineConfig, PipelineStep, PipelineWorkflow } from '~/lib/api/types';

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: {}, query: {}, hash: '' }),
}));

vi.mock('~/compositions/useApiClient', () => ({
  default: () => ({
    getLogs: vi.fn().mockResolvedValue([]),
    streamLogs: vi.fn(),
  }),
}));

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  missingWarn: false,
  fallbackWarn: false,
  messages: { en: {} },
});

const repo = ref({ id: 1, owner: 'owner', name: 'repo' });
const repoPermissions = ref({ pull: true, push: true, admin: false });
const pipelineConfigs = ref<PipelineConfig[]>([{ hash: 'h', name: 'default', data: '' }]);

function mountLog(pipeline: Pipeline, stepId: number) {
  return shallowMount(PipelineLog, {
    props: { pipeline, stepId },
    global: {
      plugins: [i18n],
      provide: { repo, 'repo-permissions': repoPermissions, 'pipeline-configs': pipelineConfigs },
      stubs: { 'router-link': true, RouterLink: true },
    },
  });
}

function makeStep(pid: number): PipelineStep {
  return {
    id: pid,
    uuid: `uuid-${pid}`,
    pipeline_id: 1,
    pid,
    ppid: 1,
    name: `step-${pid}`,
    state: 'success',
    exit_code: 0,
  };
}

// The server omits `children` for a stepless workflow, so the shape needs a cast.
interface LooseWorkflow extends Omit<PipelineWorkflow, 'children'> {
  children?: PipelineStep[] | null;
}

function makeWorkflow(id: number, children: PipelineStep[] | null | undefined): LooseWorkflow {
  const workflow: LooseWorkflow = {
    id,
    pipeline_id: 1,
    pid: id,
    name: `workflow-${id}`,
    state: 'success',
    started: 1,
    finished: 2,
  };
  if (children !== undefined) {
    workflow.children = children;
  }
  return workflow;
}

function makePipeline(workflows: LooseWorkflow[]): Pipeline {
  return {
    id: 1,
    number: 1,
    status: 'success',
    event: 'push',
    commit: 'abcdef1234567890',
    branch: 'main',
    workflows,
  } as unknown as Pipeline;
}

describe('pipelineLog', () => {
  it('finds a step past a workflow whose children key is absent', () => {
    const pipeline = makePipeline([makeWorkflow(1, undefined), makeWorkflow(2, [makeStep(7)])]);

    const wrapper = mountLog(pipeline, 7);

    expect(wrapper.exists()).toBe(true);
  });

  it('handles a null children payload', () => {
    const pipeline = makePipeline([makeWorkflow(1, null), makeWorkflow(2, [makeStep(7)])]);

    expect(() => mountLog(pipeline, 7)).not.toThrow();
  });

  it('handles a pipeline where every workflow is stepless', () => {
    const pipeline = makePipeline([makeWorkflow(1, undefined), makeWorkflow(2, [])]);

    expect(() => mountLog(pipeline, 7)).not.toThrow();
  });
});
