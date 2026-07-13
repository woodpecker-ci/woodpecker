import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { defineComponent, h, nextTick, ref } from 'vue';

import type { Tab as TabType } from '~/compositions/useTabs';

import Tab from './Tab.vue';

async function mountConditionalTab() {
  const visible = ref(true);
  const tabs = ref<TabType[]>([]);

  const host = defineComponent({
    setup() {
      return () =>
        h('div', [
          visible.value
            ? h(Tab, {
                to: { name: 'repo-pipeline-errors' },
                title: 'Errors',
                icon: 'alert',
              })
            : null,
        ]);
    },
  });

  mount(host, {
    global: {
      provide: { tabs },
    },
  });
  await nextTick();

  return { visible, tabs };
}

describe('tab', () => {
  it('registers itself on mount', async () => {
    const { tabs } = await mountConditionalTab();

    expect(tabs.value).toHaveLength(1);
    expect(tabs.value[0].title).toBe('Errors');
  });

  it('does not register the same route twice', async () => {
    const duplicateVisible = ref(true);
    const tabs = ref<TabType[]>([]);

    const host = defineComponent({
      setup() {
        return () =>
          h('div', [
            h(Tab, { to: { name: 'repo-pipeline-errors' }, title: 'Errors' }),
            duplicateVisible.value ? h(Tab, { to: { name: 'repo-pipeline-errors' }, title: 'Errors' }) : null,
          ]);
      },
    });

    mount(host, {
      global: {
        provide: { tabs },
      },
    });
    await nextTick();

    expect(tabs.value).toHaveLength(1);

    // the second instance was skipped by the dedup, so its unmount must not
    // remove the entry registered by the first, still-mounted instance
    duplicateVisible.value = false;
    await nextTick();

    expect(tabs.value).toHaveLength(1);
  });

  it('unregisters itself on unmount', async () => {
    const { visible, tabs } = await mountConditionalTab();

    expect(tabs.value).toHaveLength(1);

    visible.value = false;
    await nextTick();

    expect(tabs.value).toHaveLength(0);
  });

  it('registers again after being re-rendered', async () => {
    const { visible, tabs } = await mountConditionalTab();

    visible.value = false;
    await nextTick();
    visible.value = true;
    await nextTick();

    expect(tabs.value).toHaveLength(1);
    expect(tabs.value[0].title).toBe('Errors');
  });

  it('keeps template order when a tab mounts later than its siblings', async () => {
    const middleVisible = ref(false);
    const tabs = ref<TabType[]>([]);

    const host = defineComponent({
      setup() {
        return () =>
          h('div', [
            h(Tab, { to: { name: 'repo-pipeline' }, title: 'Tasks' }),
            middleVisible.value ? h(Tab, { to: { name: 'repo-pipeline-errors' }, title: 'Errors' }) : null,
            h(Tab, { to: { name: 'repo-pipeline-config' }, title: 'Config' }),
          ]);
      },
    });

    mount(host, {
      global: {
        provide: { tabs },
      },
    });
    await nextTick();

    expect(tabs.value.map(({ title }) => title)).toStrictEqual(['Tasks', 'Config']);

    middleVisible.value = true;
    await nextTick();

    expect(tabs.value.map(({ title }) => title)).toStrictEqual(['Tasks', 'Errors', 'Config']);
  });
});
