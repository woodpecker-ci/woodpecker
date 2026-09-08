import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, h, nextTick, ref } from 'vue';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';
import type { WorkflowInfo } from '~/lib/api/types';

import WorkflowSelect from './WorkflowSelect.vue';

const getRepoWorkflows = vi.fn<(repoId: number, branch?: string) => Promise<WorkflowInfo[]>>();
const notify = vi.fn();

vi.mock('~/compositions/useApiClient', () => ({
  default: () => ({ getRepoWorkflows: async (...args: [number, string?]) => getRepoWorkflows(...args) }),
}));

vi.mock('~/compositions/useNotifications', () => ({
  default: () => ({ notify, notifyError: vi.fn() }),
}));

const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } });

// lint and build stand alone, test needs lint, deploy needs build and test
const workflows: WorkflowInfo[] = [
  { name: 'lint' },
  { name: 'build' },
  { name: 'test', depends_on: ['lint'] },
  { name: 'deploy', depends_on: ['build', 'test'] },
];

async function mountSelect(initial: string[] = []) {
  getRepoWorkflows.mockResolvedValue(workflows);
  const selection = ref<string[]>(initial);

  const host = defineComponent({
    setup() {
      return () =>
        h(WorkflowSelect, {
          modelValue: selection.value,
          repoId: 1,
          branch: 'main',
          'onUpdate:modelValue': (value: string[]) => {
            selection.value = value;
          },
        });
    },
  });

  const wrapper = mount(host, { global: { plugins: [i18n] } });
  await nextTick();
  await nextTick();

  const boxAt = async (name: string) => {
    const index = workflows.findIndex((workflow) => workflow.name === name);
    await wrapper.findAll('input[type="checkbox"]')[index].trigger('click');
    await nextTick();
  };

  return { wrapper, selection, boxAt };
}

describe('workflowSelect', () => {
  beforeEach(() => {
    notify.mockClear();
  });

  it('pulls in the workflows a selection depends on', async () => {
    const { selection, boxAt } = await mountSelect();

    await boxAt('deploy');

    // deploy needs build and test, and test in turn needs lint
    expect(selection.value).toEqual(['lint', 'build', 'test', 'deploy']);
  });

  it('explains what it pulled in', async () => {
    const { boxAt } = await mountSelect();

    await boxAt('test');

    expect(notify).toHaveBeenCalledWith(
      expect.objectContaining({ title: expect.stringContaining('lint') as unknown as string }),
    );
  });

  it('drops the workflows that depended on a removed one', async () => {
    const { selection, boxAt } = await mountSelect();

    await boxAt('deploy');
    notify.mockClear();
    await boxAt('lint');

    // removing lint removes test, and deploy cannot survive without test
    expect(selection.value).toEqual(['build']);
    expect(notify).toHaveBeenCalledWith(
      expect.objectContaining({ title: expect.stringContaining('test') as unknown as string }),
    );
  });

  it('leaves an unrelated workflow alone', async () => {
    const { selection, boxAt } = await mountSelect();

    await boxAt('lint');
    await boxAt('build');

    expect(selection.value).toEqual(['lint', 'build']);
    expect(notify).not.toHaveBeenCalled();
  });

  it('keeps the listed order rather than the click order', async () => {
    const { selection, boxAt } = await mountSelect();

    await boxAt('build');
    await boxAt('lint');

    expect(selection.value).toEqual(['lint', 'build']);
  });
});
