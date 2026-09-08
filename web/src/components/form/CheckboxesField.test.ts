import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { computed, defineComponent, h, nextTick, ref } from 'vue';

import CheckboxesField from './CheckboxesField.vue';
import type { CheckboxOption } from './form.types';

const options: CheckboxOption[] = [
  { value: 'lint', text: 'lint' },
  { value: 'build', text: 'build' },
];

// Mirrors how Crons.vue binds the field: v-model is a computed over a property
// that starts out undefined, so the getter hands back a throwaway array. A
// child that mutates that array instead of assigning loses the value silently.
function mountWithComputedModel() {
  const store = ref<{ workflows?: string[] }>({});

  const host = defineComponent({
    setup() {
      const model = computed({
        get: () => store.value.workflows ?? [],
        set: (value: string[]) => {
          store.value.workflows = value;
        },
      });

      return () =>
        h(CheckboxesField, {
          modelValue: model.value,
          'onUpdate:modelValue': (value: string[]) => {
            model.value = value;
          },
          options,
        });
    },
  });

  return { wrapper: mount(host), store };
}

describe('checkboxesField', () => {
  it('writes a checked option back through the model', async () => {
    const { wrapper, store } = mountWithComputedModel();

    await wrapper.findAll('input[type="checkbox"]')[0].trigger('click');
    await nextTick();

    expect(store.value.workflows).toEqual(['lint']);
  });

  it('accumulates further checked options', async () => {
    const { wrapper, store } = mountWithComputedModel();
    const boxes = wrapper.findAll('input[type="checkbox"]');

    await boxes[0].trigger('click');
    await nextTick();
    await boxes[1].trigger('click');
    await nextTick();

    expect(store.value.workflows).toEqual(['lint', 'build']);
  });

  it('removes an unchecked option', async () => {
    const { wrapper, store } = mountWithComputedModel();
    const boxes = wrapper.findAll('input[type="checkbox"]');

    await boxes[0].trigger('click');
    await nextTick();
    await boxes[0].trigger('click');
    await nextTick();

    expect(store.value.workflows).toEqual([]);
  });
});
