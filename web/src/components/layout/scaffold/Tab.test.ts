import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { defineComponent, h, nextTick, ref } from 'vue';

import type { Tab as TabType } from '~/compositions/useTabs';

import Tab from './Tab.vue';

async function mountTab() {
  const title = ref('Warnings');
  const count = ref(10);
  const iconClass = ref('text-wp-state-warn-100');
  const tabs = ref<TabType[]>([]);

  const host = defineComponent({
    setup() {
      return () =>
        h(Tab, {
          to: { name: 'repo-pipeline-errors' },
          title: title.value,
          count: count.value,
          icon: 'alert',
          iconClass: iconClass.value,
        });
    },
  });

  mount(host, {
    global: {
      provide: { tabs },
    },
  });
  await nextTick();

  return { title, count, iconClass, tabs };
}

describe('tab', () => {
  it('registers itself with its current props', async () => {
    const { tabs } = await mountTab();

    expect(tabs.value).toHaveLength(1);
    expect(tabs.value[0].title).toBe('Warnings');
    expect(tabs.value[0].count).toBe(10);
  });

  it('updates the registered tab when props change', async () => {
    const { title, count, iconClass, tabs } = await mountTab();

    title.value = 'Errors';
    count.value = 11;
    iconClass.value = 'text-wp-error-100';
    await nextTick();

    expect(tabs.value).toHaveLength(1);
    expect(tabs.value[0].title).toBe('Errors');
    expect(tabs.value[0].count).toBe(11);
    expect(tabs.value[0].iconClass).toBe('text-wp-error-100');
  });
});
