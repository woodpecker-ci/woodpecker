<template>
  <Panel v-if="!loading">
    <form @submit.prevent="triggerManualPipeline">
      <span class="text-xl text-wp-text-100">{{ $t('repo.manual_pipeline.title') }}</span>
      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.select_branch')">
        <SelectField :id="id" v-model="payload.branch" :options="branches" required />
      </InputField>
      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.variables.title')">
        <span class="text-sm text-wp-text-alt-100 mb-2">{{ $t('repo.manual_pipeline.variables.desc') }}</span>
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
  <div v-else class="flex justify-center text-wp-text-100">
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
import Icon from '~/components/atomic/Icon.vue';
import InputField from '~/components/form/InputField.vue';
import KeyValueEditor from '~/components/form/KeyValueEditor.vue';
import SelectField from '~/components/form/SelectField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { inject } from '~/compositions/useInjectProvide';
import { usePaginate } from '~/compositions/usePaginate';
import type { RepoPermissions } from '~/lib/api/types';

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

const pipelineOptions = computed(() => ({
  ...payload.value,
  variables: payload.value.variables,
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
  loading.value = false;
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
