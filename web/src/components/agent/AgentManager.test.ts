import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createI18n } from 'vue-i18n';

import type { Agent } from '~/lib/api/types';

import AgentManager from './AgentManager.vue';

const route = { query: {} as Record<string, string> };

vi.mock('vue-router', () => ({
  useRoute: () => route,
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
const loadAgents = vi.fn(async (page: number) => pages[page] ?? []);

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
        AgentList: { name: 'AgentList', template: '<div data-list />' },
        AgentForm: { name: 'AgentForm', props: ['modelValue'], template: '<div data-form />' },
      },
    },
  });
}

enableAutoUnmount(afterEach);

beforeEach(() => {
  vi.clearAllMocks();
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
