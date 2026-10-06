import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { defineComponent, h, nextTick, ref } from 'vue';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';
import type { Agent } from '~/lib/api/types';

import AgentForm from './AgentForm.vue';

// resolve the labels through the real locale file, so renaming a key there
// fails this test instead of silently rendering the raw key in the app
const i18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: { en },
});

const global = { plugins: [i18n] };

async function mountAgentForm(initialAgent: Partial<Agent>, isEditingAgent = false) {
  const agent = ref(initialAgent);

  const host = defineComponent({
    setup() {
      return () =>
        h(AgentForm, {
          modelValue: agent.value,
          isEditingAgent,
          isSaving: false,
          'onUpdate:modelValue': (value: Partial<Agent>) => {
            agent.value = value;
          },
        });
    },
  });

  const wrapper = mount(host, { global });
  await nextTick();

  // InputField generates a random id, the KeyValueEditor derives its input ids from it
  const keyInputs = () => wrapper.findAll<HTMLInputElement>('input[id*="-key-"]');
  const valueInputs = () => wrapper.findAll<HTMLInputElement>('input[id*="-value-"]');

  return { wrapper, agent, keyInputs, valueInputs };
}

describe('agentForm', () => {
  it('renders the filter editor for a new agent that has no filters yet', async () => {
    const { keyInputs, valueInputs } = await mountAgentForm({ name: '' });

    // one empty key/value row to enter the first filter
    expect(keyInputs()).toHaveLength(1);
    expect(valueInputs()).toHaveLength(1);
    expect(keyInputs()[0].element.value).toBe('');
  });

  it('renders the filter editor when the server reported no filters', async () => {
    // an agent stored before filters existed is serialized with a `null` map
    const { keyInputs } = await mountAgentForm({
      name: 'agent-1',
      filters: null as unknown as Record<string, string>,
    });

    expect(keyInputs()).toHaveLength(1);
  });

  it('renders an input pair per existing filter plus one for new entries', async () => {
    const { keyInputs, valueInputs } = await mountAgentForm({
      name: 'agent-1',
      filters: { gpu: 'true', location: '*' },
    });

    expect(keyInputs()).toHaveLength(3);
    expect(keyInputs()[0].element.value).toBe('gpu');
    expect(valueInputs()[0].element.value).toBe('true');
    expect(keyInputs()[1].element.value).toBe('location');
    expect(valueInputs()[1].element.value).toBe('*');
    expect(keyInputs().at(-1)!.element.value).toBe('');
  });

  it('emits the agent with the added filter', async () => {
    const { agent, keyInputs, valueInputs } = await mountAgentForm({ name: 'agent-1' });

    await keyInputs()[0].setValue('gpu');
    await valueInputs()[0].setValue('true');
    await nextTick();

    expect(agent.value.filters).toStrictEqual({ gpu: 'true' });
    expect(agent.value.name).toBe('agent-1');
  });
});
