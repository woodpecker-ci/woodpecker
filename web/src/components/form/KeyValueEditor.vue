<template>
  <div class="flex flex-col gap-2">
    <div v-for="(item, index) in displayItems" :key="index" class="flex gap-4">
      <TextField
        :id="`${id}-key-${index}`"
        :model-value="item.key"
        :placeholder="keyPlaceholder"
        :class="{
          'bg-red-100 dark:bg-red-900':
            isDuplicateKey(item.key, index) || (item.key === '' && index !== displayItems.length - 1),
        }"
        @update:model-value="updateItem(index, 'key', $event)"
      />
      <TextField
        :id="`${id}-value-${index}`"
        :model-value="item.value"
        :placeholder="valuePlaceholder"
        @update:model-value="updateItem(index, 'value', $event)"
      />
      <div class="w-10 flex-shrink-0">
        <Button
          v-if="index !== displayItems.length - 1"
          type="button"
          color="red"
          class="ml-auto"
          :title="deleteTitle"
          @click="deleteItem(index)"
        >
          <Icon name="remove" />
        </Button>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import TextField from '~/components/form/TextField.vue';

const props = defineProps<{
  modelValue: Record<string, string>;
  id?: string;
  keyPlaceholder?: string;
  valuePlaceholder?: string;
  deleteTitle?: string;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: Record<string, string>): void;
  (e: 'update:isValid', value: boolean): void;
}>();

const items = ref(Object.entries(props.modelValue).map(([key, value]) => ({ key, value })));

const displayItems = computed(() => {
  if (items.value.length === 0 || items.value[items.value.length - 1].key !== '') {
    return [...items.value, { key: '', value: '' }];
  }
  return items.value;
});

function isDuplicateKey(key: string, index: number): boolean {
  return items.value.some((item, i) => item.key === key && i !== index && key !== '');
}

function checkValidity() {
  const isValid = items.value.every(
    (item, idx) => !isDuplicateKey(item.key, idx) && (item.key !== '' || idx === items.value.length - 1),
  );
  emit('update:isValid', isValid);
}

function updateItem(index: number, field: 'key' | 'value', value: string) {
  const newItems = [...items.value];
  if (index === newItems.length) {
    newItems.push({ key: '', value: '' });
  }
  newItems[index][field] = value;

  items.value = newItems;

  const newValue = Object.fromEntries(
    newItems
      .filter((item) => item.key !== '' && !isDuplicateKey(item.key, newItems.indexOf(item)))
      .map((item) => [item.key, item.value]),
  );

  emit('update:modelValue', newValue);
  checkValidity();
}

function deleteItem(index: number) {
  items.value = items.value.filter((_, i) => i !== index);

  const newValue = Object.fromEntries(
    items.value
      .filter((item) => item.key !== '' && !isDuplicateKey(item.key, items.value.indexOf(item)))
      .map((item) => [item.key, item.value]),
  );

  emit('update:modelValue', newValue);
  checkValidity();
}
</script>
