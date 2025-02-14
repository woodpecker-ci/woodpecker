<template>
  <div v-if="innerParameter" class="space-y-4">
    <form @submit.prevent="handleSubmit">
      <InputField v-slot="{ id }" :label="$t('parameters.name')">
        <TextField
          :id="id"
          v-model="innerParameter.name"
          :placeholder="$t('parameters.name')"
          required
        />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('parameters.branch')">
        <select
          :id="id"
          v-model="innerParameter.branch"
          class="block w-full rounded-md border border-wp-control-neutral-200 bg-wp-control-neutral-100 px-3 py-2 text-sm text-wp-text-100"
        >
          <option value="*">{{ $t('parameters.any_branch') }}</option>
          <option
            v-for="branch in branches"
            :key="branch"
            :value="branch"
          >
            {{ branch }}
          </option>
        </select>
      </InputField>

      <InputField v-slot="{ id }" :label="$t('parameters.type')">
        <select
          :id="id"
          v-model="innerParameter.type"
          class="block w-full rounded-md border border-wp-control-neutral-200 bg-wp-control-neutral-100 px-3 py-2 text-sm text-wp-text-100"
          required
        >
          <option v-for="type in parameterTypes" :key="type" :value="type">
            {{ $t(`parameters.types.${type}`) }}
          </option>
        </select>
      </InputField>

      <template v-if="innerParameter.type">
        <InputField v-if="showDefaultValue" v-slot="{ id }" :label="$t('parameters.default_value')">
          <template v-if="isBoolean">
            <Checkbox
              :id="id"
              v-model="booleanValue"
              :label="$t('parameters.set_by_default')"
            />
          </template>
          <template v-else-if="isPassword">
            <TextField
              :id="id"
              v-model="innerParameter.default_value"
              :type="passwordVisibility ? 'text' : 'password'"
              class="text-sm"
              @dblclick="passwordVisibility = true"
              @blur="passwordVisibility = false"
            />
          </template>
          <template v-else-if="isChoice">
            <TextField
              :id="id"
              v-model="innerParameter.default_value"
              :lines="5"
              :placeholder="$t('parameters.choices_placeholder')"
            />
            <span class="mt-1 text-sm text-wp-text-alt-100">{{ $t('parameters.choices_help') }}</span>
          </template>
          <template v-else-if="isText">
            <TextField
              :id="id"
              v-model="innerParameter.default_value"
              :lines="3"
              class="text-sm"
            />
          </template>
          <template v-else>
            <TextField
              :id="id"
              v-model="innerParameter.default_value"
              class="text-sm"
            />
          </template>
        </InputField>

        <InputField v-if="showTrimString" v-slot="{ id }" :label="$t('parameters.trim_string')">
          <Checkbox
            :id="id"
            v-model="innerParameter.trim_string"
            :label="innerParameter.trim_string ? 'true' : 'false'"
          />
        </InputField>
      </template>

      <InputField v-slot="{ id }" :label="$t('parameters.description')">
        <TextField
          :id="id"
          v-model="innerParameter.description"
          :lines="3"
          class="text-sm"
        />
      </InputField>

      <div class="flex gap-2">
        <Button type="button" color="gray" :text="$t('cancel')" @click="$emit('cancel')" />
        <Button
          type="submit"
          color="green"
          :is-loading="isSaving"
          :text="existingParameter ? $t('parameters.save') : $t('parameters.add')"
        />
      </div>
    </form>
  </div>
</template>

<script lang="ts" setup>
import { computed, inject, ref, watch } from 'vue';
import type { Ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import useApiClient from '~/compositions/useApiClient';
import type { Parameter, Repo } from '~/lib/api/types';
import { ParameterType } from '~/lib/api/types';

const props = defineProps<{
  modelValue: Partial<Parameter>;
  existingParameter?: boolean;
  isSaving?: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: Partial<Parameter>): void;
  (e: 'save', value: Partial<Parameter>): void;
  (e: 'cancel'): void;
}>();

const apiClient = useApiClient();
const repo = inject('repo') as Ref<Repo>;
const branches = ref<string[]>([]);

const parameterTypes = Object.values(ParameterType);

// Create a local copy of the parameter to avoid mutating props directly
const innerParameter = ref<Partial<Parameter>>({ ...props.modelValue });
const passwordVisibility = ref(false);
const booleanValue = ref(props.modelValue.default_value === 'true');

// Computed properties to control field visibility
const isBoolean = computed(() => innerParameter.value.type === ParameterType.Boolean);
const isPassword = computed(() => innerParameter.value.type === ParameterType.Password);
const isText = computed(() => innerParameter.value.type === ParameterType.Text);
const isChoice = computed(() =>
  innerParameter.value.type === ParameterType.SingleChoice ||
  innerParameter.value.type === ParameterType.MultipleChoice
);
const showDefaultValue = computed(() => innerParameter.value.type !== undefined);
const showTrimString = computed(() => !isBoolean.value && !isPassword.value);

// Update local copy when prop changes
watch(() => props.modelValue, (newVal) => {
  innerParameter.value = { ...newVal };
  booleanValue.value = innerParameter.value.default_value === 'true';
}, { deep: true });

// Watch boolean value changes to update innerParameter
watch(booleanValue, (newVal) => {
  if (isBoolean.value) {
    innerParameter.value.default_value = newVal ? 'true' : 'false';
  }
});

// Convert default_value to proper type when parameter type changes
watch(() => innerParameter.value.type, (newType) => {
  // Don't override existing values when editing
  if (props.existingParameter) {
    return;
  }

  if (newType === ParameterType.Boolean) {
    innerParameter.value.default_value = 'false';
    innerParameter.value.trim_string = false;
    booleanValue.value = false;
  } else if (newType === ParameterType.SingleChoice || newType === ParameterType.MultipleChoice) {
    innerParameter.value.default_value = '';
    innerParameter.value.trim_string = true;
  } else {
    innerParameter.value.default_value = '';
    innerParameter.value.trim_string = true;
  }
}, { immediate: true });

// Load branches when component is mounted
async function loadBranches() {
  if (!repo.value) return;

  try {
    const branchList = await apiClient.getRepoBranches(repo.value.id);
    branches.value = branchList.map(b => b);
  } catch (error) {
    console.error('Failed to load branches:', error);
  }
}

loadBranches();

function handleSubmit() {
  // Ensure all required fields are present
  if (!innerParameter.value.name || !innerParameter.value.type) {
    return;
  }

  // Ensure branch has a value
  if (!innerParameter.value.branch) {
    innerParameter.value.branch = '*';
  }

  // Convert boolean default_value to string
  if (isBoolean.value) {
    innerParameter.value.default_value = booleanValue.value ? 'true' : 'false';
  }

  // Update parent model before saving
  emit('update:modelValue', innerParameter.value);

  // Emit the save event with the parameter data
  emit('save', innerParameter.value);
}

</script>
