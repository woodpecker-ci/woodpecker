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

    <CheckboxesField v-else :id="id" :model-value="innerValue" :options="options" @update:model-value="onSelect" />
  </InputField>
</template>

<script lang="ts" setup>
import { onBeforeUnmount, ref, toRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import Icon from '~/components/atomic/Icon.vue';
import CheckboxesField from '~/components/form/CheckboxesField.vue';
import type { CheckboxOption } from '~/components/form/form.types';
import InputField from '~/components/form/InputField.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';

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
const notifications = useNotifications();
const i18n = useI18n();

const modelValue = toRef(props, 'modelValue');
const innerValue = ref<string[]>(modelValue.value ?? []);
watch(modelValue, (value) => {
  innerValue.value = value ?? [];
});

const options = ref<CheckboxOption[]>([]);
const loading = ref(false);
const error = ref('');

// name -> the workflows it requires, straight from each config's depends_on
const requires = ref<Record<string, string[]>>({});

function update(selection: string[]) {
  innerValue.value = selection;
  emit('update:modelValue', selection);
}

/**
 * A workflow cannot run without the workflows it depends on, so selecting one
 * has to bring them along, and dropping one has to drop whatever needed it.
 * Doing this here means the selection is always valid by construction and the
 * server's rejection becomes unreachable from the UI.
 */
function withRequirements(name: string, seen = new Set<string>()): Set<string> {
  if (seen.has(name)) {
    return seen; // depends_on cycles are rejected elsewhere; just don't hang
  }
  seen.add(name);
  for (const dependency of requires.value[name] ?? []) {
    withRequirements(dependency, seen);
  }
  return seen;
}

function dependentsOf(names: Set<string>): Set<string> {
  const doomed = new Set(names);
  let changed = true;
  while (changed) {
    changed = false;
    for (const [name, dependencies] of Object.entries(requires.value)) {
      if (doomed.has(name)) {
        continue;
      }
      if (dependencies.some((dependency) => doomed.has(dependency))) {
        doomed.add(name);
        changed = true;
      }
    }
  }
  return doomed;
}

function onSelect(selection: string[]) {
  const before = new Set(innerValue.value);
  const after = new Set(selection);

  const added = selection.filter((name) => !before.has(name));
  const removed = innerValue.value.filter((name) => !after.has(name));

  const next = new Set(selection);

  const pulledIn: string[] = [];
  for (const name of added) {
    for (const required of withRequirements(name)) {
      if (!next.has(required)) {
        next.add(required);
        pulledIn.push(required);
      }
    }
  }

  const droppedOut: string[] = [];
  if (removed.length > 0) {
    for (const dependent of dependentsOf(new Set(removed))) {
      if (next.has(dependent)) {
        next.delete(dependent);
        droppedOut.push(dependent);
      }
    }
  }

  // keep the order the workflows are listed in, so the checkboxes do not jump
  update(options.value.map((option) => option.value).filter((name) => next.has(name)));

  if (pulledIn.length > 0) {
    notifications.notify({
      type: 'info',
      title: i18n.t('repo.workflow_select.auto_selected', {
        workflows: pulledIn.join(', '),
      }),
    });
  }
  if (droppedOut.length > 0) {
    notifications.notify({
      type: 'info',
      title: i18n.t('repo.workflow_select.auto_deselected', {
        workflows: droppedOut.join(', '),
      }),
    });
  }
}

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
    options.value = workflows.map((workflow) => ({ text: workflow.name, value: workflow.name }));
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
