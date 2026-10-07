import { shallowMount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { ref } from 'vue';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';
import PipelineStepList from '~/components/repo/pipeline/PipelineStepList.vue';
import type { Pipeline, PipelineConfig, PipelineStep, PipelineWorkflow } from '~/lib/api/types';

// usePipeline() calls useI18n(), which must resolve during setup.
const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  missingWarn: false,
  fallbackWarn: false,
  messages: { en },
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

describe('workflow cancellation', () => {
  it('emits the database ID without changing selection or collapse state', async () => {
    const pipeline = makePipeline([makeWorkflow(91, [makeStep(2)], 'running'), makeWorkflow(92, [], 'pending')]);
    pipeline.status = 'running';
    const wrapper = mountStepList(pipeline);
    await wrapper.setProps({ canCancel: true });
    const button = wrapper.get('button[aria-label="Cancel workflow workflow-91"]');
    const before = wrapper.html();
    await button.trigger('click');
    expect(wrapper.emitted('cancelWorkflow')).toEqual([[91]]);
    expect(wrapper.emitted('update:selectedStepId')).toBeUndefined();
    expect(wrapper.html()).toBe(before);
    await wrapper.setProps({ cancelingWorkflowIds: [91] });
    expect(button.attributes('disabled')).toBeDefined();
    expect(button.attributes('aria-busy')).toBe('true');
    expect(wrapper.get('button[aria-label="Cancel workflow workflow-92"]').attributes('disabled')).toBeUndefined();
    await button.trigger('click');
    expect(wrapper.emitted('cancelWorkflow')).toHaveLength(1);
  });

  it('supports single and collapsed workflow layouts', async () => {
    const pipeline = makePipeline([makeWorkflow(91, [makeStep(2)], 'pending')]);
    pipeline.status = 'pending';
    const wrapper = mountStepList(pipeline);
    await wrapper.setProps({ canCancel: true });
    expect(wrapper.find('button[aria-label="Cancel workflow workflow-91"]').exists()).toBe(true);
    const multiple = makePipeline([makeWorkflow(91, [], 'running'), makeWorkflow(92, [], 'pending')]);
    multiple.status = 'running';
    const expanded = mountStepList(multiple);
    await expanded.setProps({ canCancel: true });
    await expanded.get('button[title="workflow-91"]').trigger('click');
    expect(expanded.find('button[aria-label="Cancel workflow workflow-91"]').exists()).toBe(true);
  });

  it.each(['blocked', 'success', 'failure', 'killed', 'canceled', 'skipped'] as const)(
    'hides cancellation for %s workflows and pipelines',
    async (state) => {
      const pipeline = makePipeline([makeWorkflow(91, [], state)]);
      pipeline.status = 'running';
      const wrapper = mountStepList(pipeline);
      await wrapper.setProps({ canCancel: true });
      expect(wrapper.find('button[aria-label]').exists()).toBe(false);
      pipeline.workflows![0].state = 'running';
      pipeline.status = state;
      await wrapper.setProps({ pipeline: { ...pipeline } });
      expect(wrapper.find('button[aria-label]').exists()).toBe(false);
    },
  );

  it('requires push permission', () => {
    const pipeline = makePipeline([makeWorkflow(91, [], 'running')]);
    pipeline.status = 'running';
    expect(mountStepList(pipeline).find('button[aria-label]').exists()).toBe(false);
  });
});
