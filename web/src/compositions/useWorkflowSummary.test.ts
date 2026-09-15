import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { defineComponent, h, ref } from 'vue';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';

import { useWorkflowSummary } from './useWorkflowSummary';

const i18n = createI18n({ legacy: false, locale: 'en', messages: { en } });

function mountSummary(initial: string[]) {
  const workflows = ref<string[]>(initial);
  let summary: ReturnType<typeof useWorkflowSummary>;

  const host = defineComponent({
    setup() {
      summary = useWorkflowSummary(workflows);
      return () => h('span', summary!.value);
    },
  });

  const wrapper = mount(host, { global: { plugins: [i18n] } });

  return { wrapper, workflows, summary: summary! };
}

describe('useWorkflowSummary', () => {
  it('says every workflow runs when nothing is selected', () => {
    const { summary } = mountSummary([]);

    expect(summary.value).toBe('All workflows');
  });

  it('uses the singular for exactly one selected workflow', () => {
    const { summary } = mountSummary(['build']);

    expect(summary.value).toBe('1 workflow');
  });

  it('shows the count for several selected workflows', () => {
    const { summary } = mountSummary(['build', 'test', 'deploy']);

    expect(summary.value).toBe('3 workflows');
  });

  it('reacts when the underlying selection changes', async () => {
    const { wrapper, workflows, summary } = mountSummary([]);
    expect(summary.value).toBe('All workflows');

    workflows.value = ['build', 'test'];
    await wrapper.vm.$nextTick();

    expect(summary.value).toBe('2 workflows');
  });
});
