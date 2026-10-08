import { shallowMount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { ref } from 'vue';
import { createI18n } from 'vue-i18n';

import PipelineStepList from '~/components/repo/pipeline/PipelineStepList.vue';
import type { Pipeline, PipelineConfig, PipelineStep, PipelineWorkflow } from '~/lib/api/types';

// usePipeline() calls useI18n(), which must resolve during setup.
const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  missingWarn: false,
  fallbackWarn: false,
  messages: { en: {} },
});

const pipelineConfigs = ref<PipelineConfig[]>([{ hash: 'h', name: 'default', data: '' }]);

function mountStepList(pipeline: Pipeline) {
  return shallowMount(PipelineStepList, {
    props: { pipeline },
    global: {
      plugins: [i18n],
      provide: { 'pipeline-configs': pipelineConfigs },
      // router-link resolves globally, so shallowMount does not stub it.
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

// The server omits `children` for a stepless workflow; older payloads sent null.
// Both shapes need a cast, since neither matches the declared type.
interface LooseWorkflow extends Omit<PipelineWorkflow, 'children'> {
  children?: PipelineStep[] | null;
}

function makeWorkflow(
  id: number,
  children: PipelineStep[] | null | undefined,
  state: PipelineWorkflow['state'],
): LooseWorkflow {
  const workflow: LooseWorkflow = {
    id,
    pipeline_id: 1,
    pid: id,
    name: `workflow-${id}`,
    state,
    started: 1,
    finished: 2,
  };
  if (children !== undefined) {
    workflow.children = children;
  }
  return workflow;
}

function makePipeline(workflows: LooseWorkflow[]): Pipeline {
  const pipeline = {
    id: 1,
    number: 1,
    parent: 0,
    event: 'push',
    event_reason: [],
    status: 'success',
    created: 1,
    updated: 2,
    started: 1,
    finished: 2,
    deploy_to: '',
    commit: 'abcdef1234567890',
    branch: 'main',
    message: 'msg',
    timestamp: 1,
    ref: 'refs/heads/main',
    refspec: '',
    clone_url: '',
    title: 'title',
    sender: 'sender',
    author: 'author',
    author_avatar: 'avatar.png',
    author_email: 'a@example.com',
    forge_url: 'https://example.com',
    reviewed_by: '',
    reviewed: 0,
    cancel_info: {},
    version: '1',
    workflows,
  };
  return pipeline as unknown as Pipeline;
}

describe('pipelineStepList', () => {
  it('renders a workflow that has no steps without throwing', () => {
    // Two workflows, so the `workflowsCollapsed` reduce in setup runs.
    const pipeline = makePipeline([makeWorkflow(1, null, 'skipped'), makeWorkflow(2, [makeStep(1)], 'success')]);

    expect(() => mountStepList(pipeline)).not.toThrow();

    const wrapper = mountStepList(pipeline);
    expect(wrapper.exists()).toBe(true);
  });

  it('renders a single stepless workflow without throwing', () => {
    const pipeline = makePipeline([makeWorkflow(1, null, 'skipped')]);

    let wrapper;
    expect(() => {
      wrapper = mountStepList(pipeline);
    }).not.toThrow();
    expect(wrapper).toBeDefined();
  });

  it('renders a normal workflow with steps', () => {
    const pipeline = makePipeline([makeWorkflow(1, [makeStep(1), makeStep(2)], 'success')]);

    const wrapper = mountStepList(pipeline);
    expect(wrapper.exists()).toBe(true);
    expect(wrapper.findAll('[data-step-id]')).toHaveLength(2);
  });

  it('handles an empty children array', () => {
    // Paired with a second workflow so the reduce in setup runs.
    const pipeline = makePipeline([makeWorkflow(1, [], 'skipped'), makeWorkflow(2, [makeStep(1)], 'success')]);

    expect(() => mountStepList(pipeline)).not.toThrow();
  });

  it('renders a workflow whose children key is absent', () => {
    const pipeline = makePipeline([makeWorkflow(1, undefined, 'skipped'), makeWorkflow(2, [makeStep(1)], 'success')]);

    expect(() => mountStepList(pipeline)).not.toThrow();
  });
});
