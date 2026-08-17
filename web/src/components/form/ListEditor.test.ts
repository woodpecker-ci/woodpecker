import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { defineComponent, h, nextTick, ref } from 'vue';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';

import ListEditor from './ListEditor.vue';

// resolve the titles through the real locale file, so renaming a key there
// fails this test instead of silently rendering the raw key in the app
const i18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: { en },
});

const global = { plugins: [i18n] };

async function mountListEditor(initialItems: string[] = [], props: Record<string, unknown> = {}) {
  const items = ref(initialItems);

  const host = defineComponent({
    setup() {
      return () =>
        h(ListEditor, {
          id: 'images',
          modelValue: items.value,
          'onUpdate:modelValue': (value: string[]) => {
            items.value = value;
          },
          ...props,
        });
    },
  });

  const wrapper = mount(host, { global });
  await nextTick();

  // the last input is the one used to enter new items, the others are the
  // disabled inputs rendering the existing entries
  const inputs = () => wrapper.findAll('input');
  const newItemInput = () => inputs().at(-1)!;
  const addButton = () => wrapper.findAll('button').at(-1)!;

  async function addItem(value: string) {
    await newItemInput().setValue(value);
    await addButton().trigger('click');
    await nextTick();
  }

  return { wrapper, items, inputs, newItemInput, addButton, addItem };
}

describe('listEditor', () => {
  it('renders an input per existing item plus one for new entries', async () => {
    const { inputs } = await mountListEditor(['plugins/git', 'plugins/docker']);

    expect(inputs()).toHaveLength(3);
    expect(inputs()[0].element.value).toBe('plugins/git');
    expect(inputs()[1].element.value).toBe('plugins/docker');
    expect(inputs().at(-1)!.element.value).toBe('');
  });

  it('renders existing items as disabled so they cannot be edited in place', async () => {
    const { inputs } = await mountListEditor(['plugins/git']);

    expect(inputs()[0].element.disabled).toBe(true);
    expect(inputs().at(-1)!.element.disabled).toBe(false);
  });

  it('emits the new list when an item is added via the button', async () => {
    const { items, addItem } = await mountListEditor(['plugins/git']);

    await addItem('plugins/docker');

    expect(items.value).toStrictEqual(['plugins/git', 'plugins/docker']);
  });

  it('adds the item when enter is pressed', async () => {
    const { items, newItemInput } = await mountListEditor([]);

    await newItemInput().setValue('plugins/git');
    await newItemInput().trigger('keydown.enter');
    await nextTick();

    expect(items.value).toStrictEqual(['plugins/git']);
  });

  it('clears the input after adding an item', async () => {
    const { newItemInput, addItem } = await mountListEditor([]);

    await addItem('plugins/git');

    expect(newItemInput().element.value).toBe('');
  });

  it('does not mutate the array it was given', async () => {
    const initialItems = ['plugins/git'];
    const { addItem } = await mountListEditor(initialItems);

    await addItem('plugins/docker');

    expect(initialItems).toStrictEqual(['plugins/git']);
  });

  it('trims surrounding whitespace from added items', async () => {
    const { items, addItem } = await mountListEditor([]);

    await addItem('  plugins/git  ');

    expect(items.value).toStrictEqual(['plugins/git']);
  });

  it('ignores an empty or whitespace-only input', async () => {
    const { items, addItem } = await mountListEditor([]);

    await addItem('');
    expect(items.value).toStrictEqual([]);

    await addItem('   ');
    expect(items.value).toStrictEqual([]);
  });

  it('ignores a duplicate and keeps it in the input so the user can see it was not added', async () => {
    const { items, newItemInput, addItem } = await mountListEditor(['plugins/git']);

    await addItem('plugins/git');

    expect(items.value).toStrictEqual(['plugins/git']);
    expect(newItemInput().element.value).toBe('plugins/git');
  });

  it('treats an entry that only differs by whitespace as a duplicate', async () => {
    const { items, addItem } = await mountListEditor(['plugins/git']);

    await addItem('  plugins/git  ');

    expect(items.value).toStrictEqual(['plugins/git']);
  });

  it('removes only the clicked item', async () => {
    const { wrapper, items } = await mountListEditor(['plugins/git', 'plugins/docker', 'plugins/s3']);

    // the delete button of the second entry
    await wrapper.findAll('button')[1].trigger('click');
    await nextTick();

    expect(items.value).toStrictEqual(['plugins/git', 'plugins/s3']);
  });

  it('drops the row once its item is removed', async () => {
    const { wrapper, inputs } = await mountListEditor(['plugins/git']);

    await wrapper.findAll('button')[0].trigger('click');
    await nextTick();

    expect(inputs()).toHaveLength(1);
  });

  it('gives every input a unique id so the labels stay unambiguous', async () => {
    const { inputs } = await mountListEditor(['plugins/git', 'plugins/docker']);

    const ids = inputs().map((input) => input.element.id);

    expect(new Set(ids).size).toBe(ids.length);
    // the label of the surrounding InputField points at the new-item input
    expect(inputs().at(-1)!.element.id).toBe('images');
  });

  it('renders all buttons as type=button so they never submit the surrounding form', async () => {
    const { wrapper } = await mountListEditor(['plugins/git']);

    const types = wrapper.findAll('button').map((button) => button.element.type);

    expect(types).toStrictEqual(['button', 'button']);
  });

  it('passes the placeholder to the new-item input only', async () => {
    const { inputs } = await mountListEditor(['plugins/git'], { placeholder: 'Plugin image' });

    expect(inputs()[0].attributes('placeholder')).toBe('');
    expect(inputs().at(-1)!.attributes('placeholder')).toBe('Plugin image');
  });

  it('titles the delete and add buttons', async () => {
    const { wrapper } = await mountListEditor(['plugins/git']);

    const buttons = wrapper.findAll('button');

    expect(buttons[0].attributes('title')).toBe('Delete');
    expect(buttons[1].attributes('title')).toBe('Add');
  });

  it('treats a missing modelValue as an empty list', async () => {
    const items = ref<string[] | undefined>(undefined);

    const host = defineComponent({
      setup() {
        return () =>
          h(ListEditor, {
            modelValue: items.value,
            'onUpdate:modelValue': (value: string[]) => {
              items.value = value;
            },
          });
      },
    });

    const wrapper = mount(host, { global });
    await nextTick();

    expect(wrapper.findAll('input')).toHaveLength(1);

    await wrapper.find('input').setValue('plugins/git');
    await wrapper.find('button').trigger('click');
    await nextTick();

    expect(items.value).toStrictEqual(['plugins/git']);
  });

  it('commits unconfirmed input when the parent calls commitPendingItem', async () => {
    const items = ref<string[]>([]);
    const editor = ref<{ commitPendingItem: () => void }>();

    const host = defineComponent({
      setup() {
        return () =>
          h(ListEditor, {
            ref: editor,
            modelValue: items.value,
            'onUpdate:modelValue': (value: string[]) => {
              items.value = value;
            },
          });
      },
    });

    const wrapper = mount(host, { global });
    await nextTick();

    // the user types but never presses enter or the add button
    await wrapper.find('input').setValue('plugins/git');
    expect(items.value).toStrictEqual([]);

    editor.value!.commitPendingItem();
    await nextTick();

    expect(items.value).toStrictEqual(['plugins/git']);
  });

  it('commitPendingItem is a no-op when the input is empty', async () => {
    const items = ref<string[]>(['plugins/git']);
    const editor = ref<{ commitPendingItem: () => void }>();

    const host = defineComponent({
      setup() {
        return () =>
          h(ListEditor, {
            ref: editor,
            modelValue: items.value,
            'onUpdate:modelValue': (value: string[]) => {
              items.value = value;
            },
          });
      },
    });

    mount(host, { global });
    await nextTick();

    editor.value!.commitPendingItem();
    await nextTick();

    expect(items.value).toStrictEqual(['plugins/git']);
  });

  it('reflects items added from outside the component', async () => {
    const { items, inputs } = await mountListEditor(['plugins/git']);

    items.value = ['plugins/git', 'plugins/docker'];
    await nextTick();

    expect(inputs()).toHaveLength(3);
    expect(inputs()[1].element.value).toBe('plugins/docker');
  });
});
