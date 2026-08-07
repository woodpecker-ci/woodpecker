import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';

import ComboboxField from './ComboboxField.vue';

const options = [
  { text: 'main', value: 'main', badge: 'default' },
  { text: 'release', value: 'release' },
  { text: 'feature/search', value: 'feature/search' },
];

describe('combobox field', () => {
  it('shows the selected value on the trigger without filtering the list', async () => {
    const wrapper = mount(ComboboxField, {
      props: {
        modelValue: 'main',
        options,
      },
      global: {
        mocks: {
          $t: (key: string) => key,
        },
      },
    });

    expect(wrapper.get('button').text()).toContain('main');
    expect(wrapper.get('button').text()).toContain('default');
    expect(wrapper.find('ul').exists()).toBe(false);

    await wrapper.get('button').trigger('click');

    expect(wrapper.findAll('[role="option"]')).toHaveLength(3);
  });

  it('filters options from a separate search input', async () => {
    const wrapper = mount(ComboboxField, {
      props: {
        modelValue: 'main',
        options,
      },
      global: {
        mocks: {
          $t: (key: string) => key,
        },
      },
    });

    await wrapper.get('button').trigger('click');
    await wrapper.get('input[type="text"]:not(.sr-only)').setValue('RELE');

    expect(wrapper.findAll('[role="option"]')).toHaveLength(1);
    expect(wrapper.get('[role="option"]').text()).toContain('release');
  });

  it('emits the selected value when an option is clicked', async () => {
    const wrapper = mount(ComboboxField, {
      props: {
        modelValue: '',
        options,
      },
      global: {
        mocks: {
          $t: (key: string) => key,
        },
      },
    });

    await wrapper.get('button').trigger('click');
    await wrapper.get('[role="option"]').trigger('mousedown');

    expect(wrapper.emitted('update:modelValue')).toEqual([['main']]);
    expect(wrapper.find('ul').exists()).toBe(false);
  });

  it('selects the first matching option when Enter is pressed in search', async () => {
    const wrapper = mount(ComboboxField, {
      props: {
        modelValue: '',
        options,
      },
      global: {
        mocks: {
          $t: (key: string) => key,
        },
      },
    });

    await wrapper.get('button').trigger('click');
    const search = wrapper.get('input[type="text"]:not(.sr-only)');
    await search.setValue('feature');
    await search.trigger('keydown.enter');

    expect(wrapper.emitted('update:modelValue')).toEqual([['feature/search']]);
  });

  it('shows a badge for the selected option on the trigger', async () => {
    const wrapper = mount(ComboboxField, {
      props: {
        modelValue: 'main',
        options,
      },
      global: {
        mocks: {
          $t: (key: string) => key,
        },
      },
    });

    expect(wrapper.get('button').text()).toContain('default');
  });
});
