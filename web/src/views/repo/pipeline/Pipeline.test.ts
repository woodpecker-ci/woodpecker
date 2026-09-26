import { flushPromises, shallowMount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ref } from 'vue';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';
import PipelineStepList from '~/components/repo/pipeline/PipelineStepList.vue';
import PipelineView from '~/views/repo/pipeline/Pipeline.vue';

const mocks = vi.hoisted(() => ({ cancelWorkflow: vi.fn(), loadPipeline: vi.fn(), notify: vi.fn() }));
vi.mock('~/compositions/useApiClient', () => ({ default: () => mocks }));
vi.mock('~/compositions/useNotifications', () => ({ default: () => mocks }));
vi.mock('~/store/pipelines', () => ({ usePipelineStore: () => mocks }));
vi.mock('~/compositions/useWPTitle', () => ({ useWPTitle: vi.fn() }));
vi.mock('vue-router', () => ({ useRouter: () => ({ replace: vi.fn() }), useRoute: () => ({ params: {} }) }));

function mountView() {
  return shallowMount(PipelineView, {
    global: {
      plugins: [createI18n({ legacy: false, locale: 'en', messages: { en } })],
      provide: {
        pipeline: ref({ number: 7, status: 'running', workflows: [], errors: [] }),
        repo: ref({ id: 3, full_name: 'owner/repo' }),
        'repo-permissions': ref({ push: true }),
      },
      stubs: { Container: { template: '<div><slot /></div>' } },
    },
  });
}

describe('workflow cancel action', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('tracks requests independently, suppresses duplicates and refreshes', async () => {
    let complete!: () => void;
    mocks.cancelWorkflow.mockReturnValue(
      new Promise<void>((resolve) => {
        complete = resolve;
      }),
    );
    const wrapper = mountView();
    const list = wrapper.getComponent(PipelineStepList);
    const vm = list.vm as unknown as { $emit: (event: 'cancelWorkflow', id: number) => void };
    vm.$emit('cancelWorkflow', 91);
    vm.$emit('cancelWorkflow', 91);
    vm.$emit('cancelWorkflow', 92);
    await flushPromises();
    expect(mocks.cancelWorkflow.mock.calls).toEqual([
      [3, 7, 91],
      [3, 7, 92],
    ]);
    expect(list.props('cancelingWorkflowIds')).toEqual([91, 92]);
    complete();
    await flushPromises();
    expect(list.props('cancelingWorkflowIds')).toEqual([]);
    expect(mocks.notify).toHaveBeenCalledWith({ title: 'Workflow cancellation requested', type: 'success' });
    expect(mocks.loadPipeline).toHaveBeenCalledWith(3, 7);
  });

  it('refreshes a rejected request and clears its loading state', async () => {
    mocks.cancelWorkflow.mockRejectedValue(new Error('workflow already finished'));
    const log = vi.spyOn(console, 'error').mockImplementation(() => {});
    const wrapper = mountView();
    const list = wrapper.getComponent(PipelineStepList);
    const vm = list.vm as unknown as { $emit: (event: 'cancelWorkflow', id: number) => void };
    vm.$emit('cancelWorkflow', 91);
    await flushPromises();
    expect(mocks.loadPipeline).toHaveBeenCalledWith(3, 7);
    expect(mocks.notify).not.toHaveBeenCalled();
    expect(list.props('cancelingWorkflowIds')).toEqual([]);
    log.mockRestore();
  });
});
