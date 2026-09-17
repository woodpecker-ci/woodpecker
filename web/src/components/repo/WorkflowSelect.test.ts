import { mount } from '@vue/test-utils';
import { describe, expect, it, vi } from 'vitest';
import { defineComponent, h, nextTick, ref } from 'vue';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';
import type { WorkflowInfo } from '~/lib/api/types';

import WorkflowSelect from './WorkflowSelect.vue';

const getRepoWorkflows = vi.fn<(repoId: number, branch?: string) => Promise<WorkflowInfo[]>>();

vi.mock('~/compositions/useApiClient', () => ({
  default: () => ({ getRepoWorkflows: async (...args: [number, string?]) => getRepoWorkflows(...args) }),
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
  it('adds a clicked workflow without pulling in anything it depends on', async () => {
    const { selection, boxAt } = await mountSelect();

    await boxAt('deploy');

    expect(selection.value).toEqual(['deploy']);
  });

  it('removes a clicked workflow without dropping anything that depends on it', async () => {
    const { selection, boxAt } = await mountSelect(['lint', 'test', 'deploy']);

    await boxAt('lint');

    expect(selection.value).toEqual(['test', 'deploy']);
  });

  it('shows the dependencies of a workflow as its checkbox description', async () => {
    const { wrapper } = await mountSelect();

    expect(wrapper.text()).toContain('Depends on: build, test');
  });

  it('does not show a description for a workflow with no dependencies', async () => {
    const { wrapper } = await mountSelect();

    const lintLabel = wrapper.findAll('label').find((label) => label.text() === 'lint');
    const description = lintLabel?.element.parentElement?.querySelector('span.text-sm');

    expect(description).toBeNull();
  });

  it('warns about a selected workflow whose dependency is not selected', async () => {
    const { wrapper, boxAt } = await mountSelect();

    await boxAt('deploy');

    expect(wrapper.text()).toContain('deploy (build, test)');
  });

  it('does not warn once every dependency is also selected', async () => {
    const { wrapper } = await mountSelect(['lint', 'build', 'test', 'deploy']);
    await nextTick();

    expect(wrapper.text()).not.toContain('will run without all of its dependencies');
  });

  it('reports every workflow ticked as no selection, so new workflows are not left out later', async () => {
    const { wrapper, selection, boxAt } = await mountSelect(['lint', 'build', 'test']);

    await boxAt('deploy');

    expect(selection.value).toEqual([]);
    const checked = wrapper.findAll('input[type="checkbox"]').map((box) => (box.element as HTMLInputElement).checked);
    expect(checked).toEqual([true, true, true, true]);
  });

  it('unticking one workflow after ticking all selects the remaining ones', async () => {
    const { selection, boxAt } = await mountSelect(['lint', 'build', 'test']);
    await boxAt('deploy');

    await boxAt('lint');

    expect(selection.value).toEqual(['build', 'test', 'deploy']);
  });

  it('does not warn when nothing is selected', async () => {
    const { wrapper } = await mountSelect();

    expect(wrapper.text()).not.toContain('will run without all of its dependencies');
  });
});
