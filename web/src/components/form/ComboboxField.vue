<template>
  <div ref="root" class="relative">
    <input v-if="required" type="text" class="sr-only" :value="modelValue" required tabindex="-1" readonly aria-hidden="true" />
    <button
      :id="id"
      type="button"
      class="border-wp-control-neutral-200 focus-visible:border-wp-control-neutral-300 bg-wp-control-neutral-100 text-wp-text-100 flex w-full items-center justify-between gap-2 rounded-md border px-2 py-1 text-left focus-visible:outline-hidden"
      :aria-expanded="open"
      aria-haspopup="listbox"
      @click="toggle"
      @keydown.down.prevent="openDropdown"
      @keydown.enter.prevent="toggle"
      @keydown.space.prevent="toggle"
    >
      <span class="flex min-w-0 flex-1 items-center gap-2">
        <span class="truncate" :class="{ 'text-wp-text-alt-100': !selectedOption && !modelValue }">
          {{ selectedLabel }}
        </span>
        <Badge v-if="selectedOption?.badge" :value="selectedOption.badge" class="shrink-0" />
      </span>
      <Icon name="expand-all" class="text-wp-text-alt-100 shrink-0" />
    </button>

    <div
      v-if="open"
      class="border-wp-control-neutral-200 bg-wp-control-neutral-100 absolute z-10 mt-1 w-full overflow-hidden rounded-md border"
      role="listbox"
    >
      <div class="border-wp-control-neutral-200 border-b p-1">
        <input
          ref="searchInput"
          v-model="query"
          type="text"
          class="border-wp-control-neutral-200 focus-visible:border-wp-control-neutral-300 bg-wp-control-neutral-100 text-wp-text-100 w-full rounded-md border px-2 py-1 focus-visible:outline-hidden"
          :placeholder="searchPlaceholder || $t('search')"
          @keydown.enter.prevent="selectFirstOption"
          @keydown.esc.prevent="closeDropdown"
        />
      </div>
      <ul v-if="filteredOptions.length > 0" class="max-h-60 overflow-auto py-1">
        <li
          v-for="option in filteredOptions"
          :key="option.value"
          class="text-wp-text-100 hover:bg-wp-control-neutral-200 flex cursor-pointer items-center gap-2 px-2 py-1"
          :class="{ 'bg-wp-control-neutral-200': option.value === modelValue }"
          role="option"
          :aria-selected="option.value === modelValue"
          @mousedown.prevent="selectOption(option)"
        >
          <span class="min-w-0 flex-1 truncate">{{ option.text }}</span>
          <Badge v-if="option.badge" :value="option.badge" class="shrink-0" />
        </li>
      </ul>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { onClickOutside } from '@vueuse/core';
import { computed, nextTick, ref, watch } from 'vue';

import Badge from '~/components/atomic/Badge.vue';
import Icon from '~/components/atomic/Icon.vue';

import type { SelectOption } from './form.types';

const props = withDefaults(
  defineProps<{
    modelValue: string;
    options: SelectOption[];
    placeholder?: string;
    searchPlaceholder?: string;
    id?: string;
    required?: boolean;
  }>(),
  {
    placeholder: '',
    searchPlaceholder: '',
    id: undefined,
    required: false,
  },
);

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void;
}>();

const root = ref<HTMLElement | null>(null);
const searchInput = ref<HTMLInputElement | null>(null);
const open = ref(false);
const query = ref('');

const selectedOption = computed(() => props.options.find((option) => option.value === props.modelValue));

const selectedLabel = computed(() => {
  if (selectedOption.value) {
    return selectedOption.value.text;
  }
  if (props.modelValue) {
    return props.modelValue;
  }
  return props.placeholder;
});

const filteredOptions = computed(() => {
  const normalizedQuery = query.value.trim().toLocaleLowerCase();
  if (!normalizedQuery) {
    return props.options;
  }
  return props.options.filter((option) => option.text.toLocaleLowerCase().includes(normalizedQuery));
});

onClickOutside(root, () => {
  closeDropdown();
});

watch(open, async (isOpen) => {
  if (!isOpen) {
    return;
  }
  await nextTick();
  searchInput.value?.focus();
});

function toggle() {
  if (open.value) {
    closeDropdown();
  } else {
    openDropdown();
  }
}

function openDropdown() {
  query.value = '';
  open.value = true;
}

function closeDropdown() {
  open.value = false;
  query.value = '';
}

function selectOption(option: SelectOption) {
  emit('update:modelValue', option.value);
  closeDropdown();
}

function selectFirstOption() {
  const [option] = filteredOptions.value;
  if (option) {
    selectOption(option);
  }
}
</script>
