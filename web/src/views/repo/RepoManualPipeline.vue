<template>
  <Panel v-if="!loading">
    <form @submit.prevent="triggerManualPipeline">
      <span class="text-wp-text-100 text-xl">{{ $t('repo.manual_pipeline.title') }}</span>
      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.select_branch')">
        <SelectField :id="id" v-model="payload.branch" :options="branches" required />
      </InputField>

      <div v-if="parameters.length > 0">
        <span class="text-wp-text-100 text-lg">{{ $t('repo.manual_pipeline.parameters.title') }}</span>
        <InputField
          v-for="parameter in parameters"
          :key="parameter.id"
          v-slot="{ id }"
          :label="parameter.required ? `${parameter.name} *` : parameter.name"
        >
          <span v-if="parameter.description" class="text-wp-text-alt-100 mb-2 text-sm">
            {{ parameter.description }}
          </span>
          <SelectField
            v-if="parameter.type === 'choice'"
            :id="id"
            :model-value="stringValue(parameter.name)"
            :options="parameterOptions(parameter)"
            @update:model-value="setParameterValue(parameter.name, $event)"
          />
          <Checkbox
            v-else-if="parameter.type === 'boolean'"
            :model-value="booleanValue(parameter.name)"
            :label="parameter.name"
            @update:model-value="setParameterValue(parameter.name, $event)"
          />
          <TextField
            v-else
            :id="id"
            :model-value="stringValue(parameter.name)"
            :type="parameter.type === 'number' ? 'number' : 'text'"
            @update:model-value="setParameterValue(parameter.name, $event)"
          />
        </InputField>
      </div>

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
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import InputField from '~/components/form/InputField.vue';
import KeyValueEditor from '~/components/form/KeyValueEditor.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { requiredInject } from '~/compositions/useInjectProvide';
import { usePaginate } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Parameter } from '~/lib/api/types';

defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

const apiClient = useApiClient();
const notifications = useNotification();
const i18n = useI18n();

const repo = requiredInject('repo');
const repoPermissions = requiredInject('repo-permissions');

const router = useRouter();
const route = useRoute();
const branches = ref<{ text: string; value: string }[]>([]);
const payload = ref<{ branch: string; variables: Record<string, string> }>({
  branch: 'main',
  variables: {},
});

const parameters = ref<Parameter[]>([]);
type ParameterValue = string | boolean;
const parameterValues = ref<Record<string, ParameterValue>>({});

function parameterOptions(parameter: Parameter): { value: string; text: string }[] {
  const options = (parameter.options ?? []).map((option) => ({ value: option, text: option }));
  // allow clearing the selection again, but only when "no value" is actually reachable:
  // required parameters must have a value and the server fills the default for empty ones
  if (!parameter.required && !parameter.default) {
    options.unshift({ value: '', text: i18n.t('repo.manual_pipeline.parameters.none') });
  }
  return options;
}

function stringValue(name: string): string {
  const value = parameterValues.value[name];
  return typeof value === 'string' ? value : '';
}

function booleanValue(name: string): boolean {
  return parameterValues.value[name] === true;
}

function setParameterValue(name: string, value: ParameterValue | number) {
  // despite TextField's typing, Vue's v-model on <input type="number"> emits numbers
  parameterValues.value[name] = typeof value === 'number' ? String(value) : value;
}

function serializeParameterValues(): Record<string, string> {
  const serialized: Record<string, string> = {};
  for (const parameter of parameters.value) {
    const value = parameterValues.value[parameter.name];
    let stringified: string;
    if (parameter.type === 'boolean') {
      stringified = value === true ? 'true' : 'false';
    } else {
      stringified = typeof value === 'string' ? value : '';
    }
    // empty values are left out so server-side defaults still apply
    if (stringified !== '') {
      serialized[parameter.name] = stringified;
    }
  }
  return serialized;
}

const isVariablesValid = ref(true);

const isParametersValid = computed(() =>
  parameters.value.every((parameter) => {
    if (!parameter.required || parameter.type === 'boolean') {
      return true;
    }
    const value = parameterValues.value[parameter.name];
    return typeof value === 'string' && value !== '';
  }),
);

const isFormValid = computed(() => {
  return payload.value.branch !== '' && isVariablesValid.value && isParametersValid.value;
});

// defined parameters win over free-text variables on key collision
const pipelineOptions = computed(() => ({
  ...payload.value,
  variables: { ...payload.value.variables, ...serializeParameterValues() },
}));

const loading = ref(true);
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

  // the ref option is currently ignored by the server, but a future workflow-defined
  // parameter source can serve branch-specific definitions through the same endpoint
  parameters.value = await usePaginate(
    async (page) => (await apiClient.getParameterList(repo.value.id, { page })) ?? [],
  );
  for (const parameter of parameters.value) {
    if (parameter.type === 'boolean') {
      parameterValues.value[parameter.name] = parameter.default === 'true';
    } else {
      parameterValues.value[parameter.name] = parameter.default ?? '';
    }
  }

  prefillFromQuery();

  loading.value = false;
});

// Prefill the form from URL query parameters, e.g.
// /repos/1/manual?branch=main&DEPLOY_TARGET=production&AD_HOC_VAR=foo
// `branch` selects the branch; keys matching a defined parameter prefill its widget
// (invalid values for the type are ignored so the default stays); everything else
// lands in the additional variables editor. Values are only prefilled, the user
// still has to submit the form.
function prefillFromQuery() {
  for (const [key, rawValue] of Object.entries(route.query)) {
    const value = Array.isArray(rawValue) ? rawValue[0] : rawValue;
    if (typeof value !== 'string') {
      continue;
    }

    if (key === 'branch') {
      if (branches.value.some((branch) => branch.value === value)) {
        payload.value.branch = value;
      }
      continue;
    }

    const parameter = parameters.value.find((p) => p.name === key);
    if (!parameter) {
      payload.value.variables[key] = value;
      continue;
    }

    switch (parameter.type) {
      case 'boolean':
        if (value === 'true' || value === 'false') {
          parameterValues.value[parameter.name] = value === 'true';
        }
        break;
      case 'choice':
        if (parameter.options?.includes(value)) {
          parameterValues.value[parameter.name] = value;
        }
        break;
      case 'number':
        if (!Number.isNaN(Number.parseFloat(value))) {
          parameterValues.value[parameter.name] = value;
        }
        break;
      default:
        parameterValues.value[parameter.name] = value;
    }
  }
}

async function triggerManualPipeline() {
  loading.value = true;
  const pipeline = await apiClient.createPipeline(repo.value.id, pipelineOptions.value);

  emit('close');

  if (typeof pipeline == 'string') {
    // if this is a string (http 204) there is no workflow to run with the 'manual' event

    await router.push({
      name: 'repo',
    });

    notifications.notify({ type: 'warn', title: i18n.t('repo.manual_pipeline.no_manual_workflows') });
  } else {
    await router.push({
      name: 'repo-pipeline',
      params: {
        pipelineId: pipeline.number,
      },
    });
  }

  loading.value = false;
}

useWPTitle(computed(() => [i18n.t('repo.manual_pipeline.trigger'), repo.value.full_name]));
</script>
