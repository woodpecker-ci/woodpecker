<template>
  <InputField v-slot="{ id }" :label="$t('repo.workflow_select.title')">
    <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.workflow_select.desc') }}</span>

    <div v-if="loading" class="text-wp-text-alt-100 flex items-center gap-2 text-sm">
      <Icon name="spinner" />
      <span>{{ $t('repo.workflow_select.loading') }}</span>
    </div>

    <span v-else-if="error" class="text-wp-error-100 text-sm">{{ error }}</span>

    <span v-else-if="options.length === 0" class="text-wp-text-alt-100 text-sm">
      {{ $t('repo.workflow_select.none') }}
    </span>

    <template v-else>
      <CheckboxesField :id="id" :model-value="innerValue" :options="options" @update:model-value="update" />

      <Warning
        v-if="unmetDependencies.length > 0"
        class="mt-2 text-sm"
        :text="$t('repo.workflow_select.unmet_dependencies', { workflows: unmetDependenciesText })"
      />
    </template>
  </InputField>
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, ref, toRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import Icon from '~/components/atomic/Icon.vue';
import Warning from '~/components/atomic/Warning.vue';
import CheckboxesField from '~/components/form/CheckboxesField.vue';
import type { CheckboxOption } from '~/components/form/form.types';
import InputField from '~/components/form/InputField.vue';
import useApiClient from '~/compositions/useApiClient';

const props = defineProps<{
  modelValue?: string[];
  repoId: number;
  // branch to read the workflow list from, defaults to the repo default branch
  branch?: string;
}>();

const emit = defineEmits<{
  (event: 'update:modelValue', value: string[]): void;
}>();

const apiClient = useApiClient();
const i18n = useI18n();

const modelValue = toRef(props, 'modelValue');
const innerValue = ref<string[]>(modelValue.value ?? []);
watch(modelValue, (value) => {
  innerValue.value = value ?? [];
});

const options = ref<CheckboxOption[]>([]);
const loading = ref(false);
const error = ref('');

// name -> the workflows it names in depends_on, straight from each config
const requires = ref<Record<string, string[]>>({});

function update(selection: string[]) {
  innerValue.value = selection;
  emit('update:modelValue', selection);
}

/**
 * A workflow whose depends_on names something outside the current selection
 * still runs: it just does not wait for that dependency, and may fail on its
 * own if it actually needed its output. This lists which selected workflows
 * that applies to, so the non-blocking warning below can name them instead of
 * the person only finding out from a failed run.
 */
const unmetDependencies = computed(() => {
  const selected = new Set(innerValue.value);
  const items: { name: string; missing: string[] }[] = [];
  for (const name of innerValue.value) {
    const missing = (requires.value[name] ?? []).filter((dependency) => !selected.has(dependency));
    if (missing.length > 0) {
      items.push({ name, missing });
    }
  }
  return items;
});

const unmetDependenciesText = computed(() =>
  unmetDependencies.value.map((item) => `${item.name} (${item.missing.join(', ')})`).join(', '),
);

// The branch can come from a text input, so this is called on every keystroke.
// Requests are debounced, and each one carries a sequence number so a slow
// response for an abandoned branch cannot overwrite a newer result.
const debounceMs = 300;
let debounceTimer: ReturnType<typeof setTimeout> | undefined;
let latestRequest = 0;

async function loadWorkflows(branch?: string) {
  const request = ++latestRequest;
  loading.value = true;
  error.value = '';
  try {
    const workflows = await apiClient.getRepoWorkflows(props.repoId, branch);
    if (request !== latestRequest) {
      return;
    }
    options.value = workflows.map((workflow) => ({
      text: workflow.name,
      value: workflow.name,
      description:
        workflow.depends_on && workflow.depends_on.length > 0
          ? i18n.t('repo.workflow_select.depends_on', { workflows: workflow.depends_on.join(', ') })
          : undefined,
    }));
    requires.value = Object.fromEntries(workflows.map((workflow) => [workflow.name, workflow.depends_on ?? []]));

    // a workflow that no longer exists on this branch cannot stay selected
    const available = new Set(workflows.map((workflow) => workflow.name));
    const stillValid = innerValue.value.filter((workflow) => available.has(workflow));
    if (stillValid.length !== innerValue.value.length) {
      update(stillValid);
    }
  } catch {
    if (request !== latestRequest) {
      return;
    }
    options.value = [];
    requires.value = {};
    error.value = i18n.t('repo.workflow_select.error');
  } finally {
    if (request === latestRequest) {
      loading.value = false;
    }
  }
}

function scheduleLoad(branch?: string, immediate = false) {
  clearTimeout(debounceTimer);
  if (immediate) {
    void loadWorkflows(branch);
    return;
  }
  // show the spinner straight away, so the list is never silently stale while
  // the debounce window is open
  loading.value = true;
  debounceTimer = setTimeout(() => void loadWorkflows(branch), debounceMs);
}

watch(
  () => props.branch,
  (branch, previous) => scheduleLoad(branch, previous === undefined),
  { immediate: true },
);

onBeforeUnmount(() => clearTimeout(debounceTimer));
</script>
