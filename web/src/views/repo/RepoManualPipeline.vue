<template>
  <Panel v-if="!loading">
    <form @submit.prevent="triggerManualPipeline">
      <span class="text-wp-text-100 text-xl">{{ $t('repo.manual_pipeline.title') }}</span>
      <InputField :label="$t('repo.manual_pipeline.select_source')">
        <RadioField v-model="mode" :options="sourceOptions" />
      </InputField>
      <InputField v-if="mode === 'branch'" v-slot="{ id }" :label="$t('repo.manual_pipeline.select_branch')">
        <ComboboxField
          :id="id"
          v-model="branch"
          :options="branches"
          :placeholder="$t('repo.manual_pipeline.select_branch')"
          required
        />
      </InputField>
      <InputField v-else-if="mode === 'tag'" v-slot="{ id }" :label="$t('repo.manual_pipeline.select_tag')">
        <ComboboxField
          :id="id"
          v-model="tag"
          :options="tags"
          :placeholder="$t('repo.manual_pipeline.select_tag')"
          required
        />
      </InputField>
      <InputField v-else v-slot="{ id }" :label="$t('repo.manual_pipeline.enter_commit')">
        <TextField :id="id" v-model="sha" :placeholder="$t('repo.manual_pipeline.enter_commit_placeholder')" />
      </InputField>
      <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.manual_pipeline.event_note') }}</span>
      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.variables.title')">
        <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.manual_pipeline.variables.desc') }}</span>
        <KeyValueEditor
          :id="id"
          v-model="variables"
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
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import ComboboxField from '~/components/form/ComboboxField.vue';
import InputField from '~/components/form/InputField.vue';
import KeyValueEditor from '~/components/form/KeyValueEditor.vue';
import RadioField from '~/components/form/RadioField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { requiredInject } from '~/compositions/useInjectProvide';
import { usePaginate } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { SelectOption } from '~/components/form/form.types';

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
type SourceMode = 'branch' | 'tag' | 'commit';

const mode = ref<SourceMode>('branch');
const branch = ref(repo.value.default_branch);
const tag = ref('');
const sha = ref('');
const branches = ref<SelectOption[]>([]);
const tags = ref<SelectOption[]>([]);
const variables = ref<Record<string, string>>({});
const isVariablesValid = ref(true);

const sourceOptions = computed(() => [
  { text: i18n.t('repo.manual_pipeline.source_branch'), value: 'branch' },
  { text: i18n.t('repo.manual_pipeline.source_tag'), value: 'tag' },
  { text: i18n.t('repo.manual_pipeline.source_commit'), value: 'commit' },
]);

const isFormValid = computed(() => {
  if (!isVariablesValid.value) {
    return false;
  }

  if (mode.value === 'branch') {
    return branch.value !== '';
  }
  if (mode.value === 'tag') {
    return tag.value !== '';
  }
  return /^[0-9a-f]{7,40}$/i.test(sha.value);
});

const pipelineOptions = computed(() => {
  const base = { variables: variables.value };
  if (mode.value === 'branch') {
    return { ...base, branch: branch.value };
  }
  if (mode.value === 'tag') {
    return { ...base, tag: tag.value };
  }
  return { ...base, sha: sha.value.trim() };
});

const loading = ref(true);
onMounted(async () => {
  if (!repoPermissions.value.push) {
    notifications.notify({ type: 'error', title: i18n.t('repo.settings.not_allowed') });
    await router.replace({ name: 'home' });
  }

  const branchData = await usePaginate((page) => apiClient.getRepoBranches(repo.value.id, { page }));
  const tagData = await usePaginate((page) => apiClient.getRepoTags(repo.value.id, { page })).catch((error) => {
    console.error('Error fetching tags:', error);
    return [];
  });
  const defaultBranch = repo.value.default_branch;
  branches.value = branchData
    .toSorted((a, b) => {
      if (a === defaultBranch) {
        return -1;
      }
      if (b === defaultBranch) {
        return 1;
      }
      return 0;
    })
    .map((value) => ({
      text: value,
      value,
      badge: value === defaultBranch ? i18n.t('default') : undefined,
    }));
  tags.value = tagData.map((value) => ({ text: value, value }));
  loading.value = false;
});

async function triggerManualPipeline() {
  loading.value = true;
  try {
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
  } catch (error) {
    console.error('Error triggering manual pipeline:', error);
    notifications.notify({ type: 'error', title: i18n.t('repo.manual_pipeline.trigger_error') });
  } finally {
    loading.value = false;
  }
}

useWPTitle(computed(() => [i18n.t('repo.manual_pipeline.trigger'), repo.value.full_name]));
</script>
