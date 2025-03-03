<template>
  <Panel v-if="!loading">
    <form @submit.prevent="triggerManualPipeline">
      <span class="text-wp-text-100 text-xl">{{ $t('repo.manual_pipeline.title') }}</span>
      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.select_branch')">
        <SelectField :id="id" v-model="payload.branch" :options="branches" required @update:model-value="loadParameters" />
      </InputField>

      <!-- Parameters section -->
      <template v-if="parameters.length > 0">
        <InputField :label="$t('repo.manual_pipeline.parameters.title')">
          <span class="mb-2 text-sm text-wp-text-alt-100">{{ $t('repo.manual_pipeline.parameters.desc') }}</span>
        </InputField>

        <div v-for="param in parameters" :key="param.id" class="mb-4 rounded-md border border-wp-background-300 items-center !bg-wp-background-200 dark:!bg-wp-background-100 p-4">
          <InputField v-slot="{ id }" :label="param.name">
            <!-- Boolean parameter -->
            <Checkbox
              v-if="param.type === 'boolean'"
              :id="id"
              v-model="paramValues[param.name]"
              :label="`${paramValues[param.name]}`"
              class="text-sm"
            />

            <!-- Single choice parameter -->
            <SelectField
              v-else-if="param.type === 'single_choice'"
              :id="id"
              v-model="paramValues[param.name]"
              :options="getOptionsFromDefaultValue(param.default_value, true)"
              class="text-sm dark:bg-wp-background-200"
            />

            <!-- Multiple choice parameter -->
            <SelectField
              v-else-if="param.type === 'multiple_choice'"
              :id="id"
              v-model="paramValues[param.name]"
              :options="getOptionsFromDefaultValue(param.default_value)"
              multiple
              class="text-sm dark:bg-wp-background-200"
              :style="{ height: `${getOptionsFromDefaultValue(param.default_value).length * 19 + 2 }px` }"
            />

            <!-- String parameter -->
            <TextField
              v-else-if="param.type === 'string'"
              :id="id"
              v-model="paramValues[param.name]"
              :placeholder="param.name"
              class="text-sm dark:bg-wp-background-200"
            />

            <!-- Text parameter -->
            <textarea
              v-else-if="param.type === 'text'"
              :id="id"
              v-model="paramValues[param.name]"
              class="block w-full rounded-md bg-white border border-wp-control-neutral-200 px-3 py-2 text-sm text-wp-text-100 dark:bg-wp-background-200"
              rows="4"
            />

            <!-- Password parameter -->
            <TextField
              v-else-if="param.type === 'password'"
              :id="id"
              v-model="paramValues[param.name]"
              :placeholder="param.name"
              :type="passwordVisibility[param.name] ? 'text' : 'password'"
              class="text-sm dark:bg-wp-background-200"
              @dblclick="passwordVisibility[param.name] = true"
              @blur="passwordVisibility[param.name] = false"
            />

            <!-- Trim string checkbox if applicable -->
            <div v-if="param.trim_string && ['string', 'text'].includes(param.type)" class="mt-2">
              <Checkbox
                :id="`${id}-trim`"
                v-model="paramTrimEnabled[param.name]"
                :label="$t('repo.manual_pipeline.parameters.trim')"
                class="text-sm"
              />
            </div>

            <!-- Parameter description -->
            <div v-if="param.description" class="mt-2 text-sm text-wp-text-alt-100">
              {{ param.description }}
            </div>
          </InputField>
        </div>
      </template>

      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.variables.title')">
        <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.manual_pipeline.variables.desc') }}</span>
        <KeyValueEditor
          :id="id"
          v-model="payload.variables"
          :key-placeholder="$t('repo.manual_pipeline.variables.name')"
          :value-placeholder="$t('repo.manual_pipeline.variables.value')"
          :delete-title="$t('repo.manual_pipeline.variables.delete')"
          @update:is-valid="isVariablesValid = $event"
        />
      </InputField>
      <Button type="submit" :text="$t('repo.manual_pipeline.trigger')" :disabled="!isFormValid" />
    </form>
  </Panel>
  <div v-else class="text-wp-text-100 flex justify-center">
    <Icon name="spinner" />
  </div>
</template>

<script lang="ts" setup>
import { useNotification } from '@kyvg/vue3-notification';
import type { Ref } from 'vue';
import { computed, onMounted, ref, inject as vueInject } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import Icon from '~/components/atomic/Icon.vue';
import InputField from '~/components/form/InputField.vue';
import KeyValueEditor from '~/components/form/KeyValueEditor.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { inject } from '~/compositions/useInjectProvide';
import { usePaginate } from '~/compositions/usePaginate';
import type { Parameter, RepoPermissions } from '~/lib/api/types';

defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

const apiClient = useApiClient();
const notifications = useNotification();
const i18n = useI18n();

const repo = inject('repo');
const repoPermissions = vueInject<Ref<RepoPermissions>>('repo-permissions');
if (!repoPermissions) {
  throw new Error('Unexpected: "repo" and "repoPermissions" should be provided at this place');
}

const router = useRouter();
const branches = ref<{ text: string; value: string }[]>([]);
const payload = ref<{ branch: string; variables: Record<string, string> }>({
  branch: 'main',
  variables: {},
});

const isVariablesValid = ref(true);

const isFormValid = computed(() => {
  return payload.value.branch !== '' && isVariablesValid.value;
});

const pipelineOptions = computed(() => {
  // Start with the base payload
  const options = {
    ...payload.value,
    variables: { ...payload.value.variables },
  };

  // Add parameter values
  parameters.value.forEach(param => {
    let value = paramValues.value[param.name];

    // Convert boolean to string
    if (param.type === 'boolean') {
      value = value ? 'true' : 'false';
    }

    // Join multiple choice values
    if (param.type === 'multiple_choice' && Array.isArray(value)) {
      value = value.join(',');
    }

    // Apply trim if enabled
    if (paramTrimEnabled.value[param.name] && typeof value === 'string') {
      value = value.trim();
    }

    //options.variables[param.name.toUpperCase()] = value;
    options.variables[param.name] = value;
  });

  return options;
});

const loading = ref(true);
const parameters = ref<Parameter[]>([]);
const paramValues = ref<Record<string, any>>({});
const paramTrimEnabled = ref<Record<string, boolean>>({});
const passwordVisibility = ref<Record<string, boolean>>({});

function getOptionsFromDefaultValue(defaultValue: string, addEmptyOption = false) {
  const options = defaultValue.split('\n').filter(Boolean).map(value => ({
    text: value,
    value,
  }));

  // For single choice, add empty option at the start
  if (addEmptyOption && options.length > 0) {
    options.unshift({ text: '-- empty --', value: '' });
  }

  return options;
}

async function loadParameters() {
  if (!payload.value.branch) return;

  try {
    const allParams = await apiClient.getParameters(repo.value);
    const selectedBranch = payload.value.branch;

    // Create a map to store parameters by name
    const paramMap = new Map<string, Parameter>();

    // First pass: add wildcard (*) parameters
    allParams.forEach(param => {
      if (param.branch === '*') {
        paramMap.set(param.name, param);
      }
    });

    // Second pass: override with branch-specific parameters
    allParams.forEach(param => {
      if (param.branch === selectedBranch) {
        paramMap.set(param.name, param);
      }
    });

    // Convert map back to array
    parameters.value = Array.from(paramMap.values());

    // Initialize parameter values with defaults
    parameters.value.forEach(param => {
      if (param.type === 'boolean') {
        paramValues.value[param.name] = param.default_value === 'true';
      } else if (param.type === 'single_choice') {
        paramValues.value[param.name] = ''; // Empty string for default empty selection
      } else if (param.type === 'multiple_choice') {
        paramValues.value[param.name] = []; // Empty array for no selections
      } else {
        paramValues.value[param.name] = param.default_value;
      }

      if (param.type === 'password') {
        passwordVisibility.value[param.name] = false;
      }

      if (param.trim_string) {
        paramTrimEnabled.value[param.name] = true;
      }
    });
  } catch (error) {
    console.error('Failed to load parameters:', error);
  }
}

// Load parameters when component mounts
onMounted(async () => {
  if (!repoPermissions.value.push) {
    notifications.notify({ type: 'error', title: i18n.t('repo.settings.not_allowed') });
    await router.replace({ name: 'home' });
  }

  const data = await usePaginate((page) => apiClient.getRepoBranches(repo.value.id, { page }));
  branches.value = data.map((e) => ({
    text: e,
    value: e,
  }));
  loading.value = false;
  if (payload.value.branch) {
    loadParameters();
  }
});

async function triggerManualPipeline() {
  loading.value = true;
  const pipeline = await apiClient.createPipeline(repo.value.id, pipelineOptions.value);

  emit('close');

  await router.push({
    name: 'repo-pipeline',
    params: {
      pipelineId: pipeline.number,
    },
  });

  loading.value = false;
}
</script>

<style scoped>
.dark input[type="checkbox"] {
  background-color: var(--wp-background-200)!important;
}
</style>
