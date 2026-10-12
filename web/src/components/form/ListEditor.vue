<template>
  <div class="flex flex-col gap-2">
    <div v-for="(item, index) in modelValue" :key="item" class="flex gap-2">
      <TextField :id="`${id}-${index}`" :model-value="item" disabled />
      <Button type="button" color="gray" start-icon="trash" :title="$t('delete')" @click="deleteItem(item)" />
    </div>
    <div class="flex gap-2">
      <TextField :id="id" v-model="newItem" :placeholder="placeholder" @keydown.enter.prevent="addNewItem" />
      <Button type="button" color="gray" start-icon="plus" :title="$t('add')" @click="addNewItem" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import TextField from '~/components/form/TextField.vue';

const props = withDefaults(
  defineProps<{
    modelValue?: string[];
    id?: string;
    placeholder?: string;
  }>(),
  {
    modelValue: () => [],
    id: undefined,
    placeholder: undefined,
  },
);

const emit = defineEmits<{
  (e: 'update:modelValue', value: string[]): void;
}>();

const newItem = ref('');

function addNewItem() {
  const item = newItem.value.trim();
  // an entry that is blank or already listed would be indistinguishable from
  // the existing ones, so keep it in the input instead of adding a duplicate
  if (!item || props.modelValue.includes(item)) {
    return;
  }

  emit('update:modelValue', [...props.modelValue, item]);
  newItem.value = '';
}

function deleteItem(item: string) {
  emit(
    'update:modelValue',
    props.modelValue.filter((i) => i !== item),
  );
}

// lets a parent commit text the user typed but never confirmed, so submitting
// the surrounding form does not silently drop it
defineExpose({ commitPendingItem: addNewItem });
</script>
