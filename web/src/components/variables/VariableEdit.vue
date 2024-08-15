<template>
  <div v-if="innerValue" class="space-y-4">
    <form @submit.prevent="save">
      <InputField v-slot="{ id }" :label="$t('variables.name')">
        <TextField
          :id="id"
          v-model="innerValue.name"
          :placeholder="$t('variables.name')"
          required
          :disabled="isEditingVariable"
        />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('variables.value')">
        <TextField
          :id="id"
          v-model="innerValue.value"
          :placeholder="$t('variables.value')"
          :lines="5"
          :required="!isEditingVariable"
        />
      </InputField>

      <div class="flex gap-2">
        <Button type="button" color="gray" :text="$t('cancel')" @click="$emit('cancel')" />
        <Button
          type="submit"
          color="green"
          :is-loading="isSaving"
          :text="isEditingVariable ? $t('variables.save') : $t('variables.add')"
        />
      </div>
    </form>
  </div>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import Button from '~/components/atomic/Button.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import type { Variable } from '~/lib/api/types';

const props = defineProps<{
  modelValue: Partial<Variable>;
  isSaving: boolean;
}>();

const emit = defineEmits<{
  (event: 'update:modelValue', value: Partial<Variable> | undefined): void;
  (event: 'save', value: Partial<Variable>): void;
  (event: 'cancel'): void;
}>();

const modelValue = toRef(props, 'modelValue');
const innerValue = computed({
  get: () => modelValue.value,
  set: (value) => {
    emit('update:modelValue', value);
  },
});
const isEditingVariable = computed(() => !!innerValue.value?.id);

function save() {
  if (!innerValue.value) {
    return;
  }

  emit('save', innerValue.value);
}
</script>
