<template>
  <div v-if="innerValue" class="space-y-4">
    <form @submit.prevent="save">
      <InputField v-slot="{ id }" :label="$t('registries.address.address')">
        <!-- TODO: check input field Address is a valid address -->
        <TextField
          :id="id"
          v-model="innerValue.address"
          :placeholder="$t('registries.address.desc')"
          required
          :disabled="isEditing || isReadOnly"
        />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('username')">
        <TextField
          :id="id"
          v-model="innerValue.username"
          :placeholder="$t('username')"
          required
          :disabled="isReadOnly"
        />
      </InputField>

      <InputField v-if="!isReadOnly" v-slot="{ id }" :label="$t('password')">
        <TextField :id="id" v-model="innerValue.password" :placeholder="$t('password')" :required="!isEditing" />
      </InputField>

      <div v-if="!isReadOnly" class="flex gap-2">
        <Button type="button" color="gray" :text="$t('cancel')" @click="$emit('cancel')" />
        <Button
          type="submit"
          color="green"
          :is-loading="isSaving"
          :text="isEditing ? $t('registries.save') : $t('registries.add')"
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
import type { Registry } from '~/lib/api/types';

const props = defineProps<{
  modelValue: Partial<Registry>;
  isSaving: boolean;
}>();

const emit = defineEmits<{
  (event: 'update:modelValue', value: Partial<Registry> | undefined): void;
  (event: 'save', value: Partial<Registry>): void;
  (event: 'cancel'): void;
}>();

const modelValue = toRef(props, 'modelValue');
const innerValue = computed({
  get: () => modelValue.value,
  set: (value) => {
    emit('update:modelValue', value);
  },
});
const isEditing = computed(() => !!innerValue.value?.id);
const isReadOnly = computed(() => !!innerValue.value?.readonly);

function save() {
  if (!innerValue.value) {
    return;
  }

  emit('save', innerValue.value);
}
</script>
