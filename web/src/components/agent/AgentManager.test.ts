import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createI18n } from 'vue-i18n';

import type { Agent } from '~/lib/api/types';

import AgentManager from './AgentManager.vue';

const route = { query: {} as Record<string, string> };
const replace = vi.fn();

vi.mock('vue-router', () => ({
  useRoute: () => route,
  useRouter: () => ({ replace }),
}));

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  missingWarn: false,
  fallbackWarn: false,
  messages: { en: {} },
});

function makeAgent(id: number): Agent {
  return { id, name: `agent-${id}` } as Agent;
}

// two full pages and a last one, like the server pages the agent lists
const pages: Record<number, Agent[]> = {
  1: [makeAgent(1), makeAgent(2)],
  2: [makeAgent(3), makeAgent(4)],
  3: [makeAgent(5)],
};
const loadPage = async (page: number) => pages[page] ?? [];
const loadAgents = vi.fn(loadPage);

function mountManager() {
  return mount(AgentManager, {
    props: {
      description: '',
      loadAgents,
      createAgent: vi.fn(),
      updateAgent: vi.fn(),
      deleteAgent: vi.fn(),
    },
    global: {
      plugins: [i18n],
      stubs: {
        Settings: { template: '<div><slot /></div>' },
        Button: true,
        AgentList: {
          name: 'AgentList',
          emits: ['edit'],
          template: `<div data-list><button data-edit @click="$emit('edit', { id: 9, name: 'picked' })" /></div>`,
        },
        AgentForm: {
          name: 'AgentForm',
          props: ['modelValue'],
          emits: ['cancel'],
          template: `<div data-form><button data-cancel @click="$emit('cancel')" /></div>`,
        },
      },
    },
  });
}

enableAutoUnmount(afterEach);

beforeEach(() => {
  vi.clearAllMocks();
  loadAgents.mockImplementation(loadPage);
  route.query = {};
});

describe('agentManager', () => {
  it('shows the agent list without an agent in the url', async () => {
    const wrapper = mountManager();
    await flushPromises();

    expect(wrapper.find('[data-list]').exists()).toBe(true);
    expect(wrapper.find('[data-form]').exists()).toBe(false);
  });

  it('opens the agent from the url, also when it is not on the first page', async () => {
    route.query = { agent: '5' };
    const wrapper = mountManager();
    await flushPromises();

    expect(wrapper.findComponent({ name: 'AgentForm' }).props('modelValue')).toStrictEqual(makeAgent(5));
  });

  it('drops the agent from the url when going back to the list, so a reload shows the list', async () => {
    route.query = { agent: '5', other: 'kept' };
    const wrapper = mountManager();
    await flushPromises();
    expect(replace).not.toHaveBeenCalled();

    await wrapper.get('[data-cancel]').trigger('click');
    await flushPromises();

    expect(wrapper.find('[data-list]').exists()).toBe(true);
    expect(replace).toHaveBeenCalledWith({ query: { other: 'kept' } });
  });

  it('puts an agent opened from the list into the url, so a reload opens it again', async () => {
    route.query = { other: 'kept' };
    const wrapper = mountManager();
    await flushPromises();

    await wrapper.get('[data-edit]').trigger('click');
    await flushPromises();

    expect(replace).toHaveBeenCalledWith({ query: { other: 'kept', agent: '9' } });
  });

  it('keeps an agent the user picked while the linked one is still searched', async () => {
    let resolvePage2!: (agents: Agent[]) => void;
    loadAgents.mockImplementation(async (page: number) =>
      page === 2 ? new Promise<Agent[]>((resolve) => (resolvePage2 = resolve)) : loadPage(page),
    );
    route.query = { agent: '5' };
    const wrapper = mountManager();
    await flushPromises();

    await wrapper.get('[data-edit]').trigger('click');
    resolvePage2(pages[2]);
    await flushPromises();

    expect(wrapper.findComponent({ name: 'AgentForm' }).props('modelValue')).toStrictEqual({ id: 9, name: 'picked' });
    expect(loadAgents).not.toHaveBeenCalledWith(3);
  });

  it('stays on the list when the agent from the url does not exist', async () => {
    route.query = { agent: '42' };
    const wrapper = mountManager();
    await flushPromises();

    expect(wrapper.find('[data-form]').exists()).toBe(false);
    // stops at the first empty page
    expect(loadAgents).toHaveBeenCalledWith(4);
    expect(loadAgents).not.toHaveBeenCalledWith(5);
  });
});
