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

  it('keeps the tab when the registering duplicate unmounts before the skipped one', async () => {
    const firstVisible = ref(true);
    const secondVisible = ref(true);
    const tabs = ref<TabType[]>([]);

    const host = defineComponent({
      setup() {
        return () =>
          h('div', [
            firstVisible.value ? h(Tab, { to: { name: 'repo-pipeline-errors' }, title: 'Errors' }) : null,
            secondVisible.value ? h(Tab, { to: { name: 'repo-pipeline-errors' }, title: 'Errors' }) : null,
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

    // the registering instance goes away, but a matching instance is still
    // mounted, so the shared tab must survive
    firstVisible.value = false;
    await nextTick();

    expect(tabs.value).toHaveLength(1);
    expect(tabs.value[0].title).toBe('Errors');

    // once the last matching instance unmounts, the tab must disappear
    secondVisible.value = false;
    await nextTick();

    expect(tabs.value).toHaveLength(0);
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
