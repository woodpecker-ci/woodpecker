<template>
  <template v-if="repoPermissions && repoPermissions.push">
    <Panel>
      <InputField :label="$t('repo.pipeline.debug.metadata_exec_title')">
        <p class="text-sm text-wp-text-alt-100 mb-2">{{ $t('repo.pipeline.debug.metadata_exec_desc') }}</p>
        <pre class="code-box">{{ cliExecWithMetadata }}</pre>
      </InputField>
      <div class="flex items-center space-x-4">
        <Button :is-loading="isLoading" :text="$t('repo.pipeline.debug.download_metadata')" @click="downloadMetadata" />
      </div>
    </Panel>
  </template>
  <div v-else class="flex items-center justify-center h-full">
    <div class="text-center p-8 bg-wp-control-error-100 rounded-lg shadow-lg">
      <p class="text-2xl font-bold text-white">{{ $t('repo.pipeline.debug.no_permission') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import InputField from '~/components/form/InputField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import type { Pipeline, Repo, RepoPermissions } from '~/lib/api/types';

const { t } = useI18n();
const apiClient = useApiClient();
const notifications = useNotifications();

const repo = inject<Ref<Repo>>('repo');
const pipeline = inject<Ref<Pipeline>>('pipeline');
const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');

const isLoading = ref(false);

const metadataFileName = computed(
  () => `${repo?.value.full_name.replaceAll('/', '-')}-pipeline-${pipeline?.value.number}-metadata.json`,
);
const cliExecWithMetadata = computed(() => `# woodpecker exec --metadata-file ${metadataFileName.value}`);

async function downloadMetadata() {
  if (!repo?.value || !pipeline?.value || !repoPermissions?.value?.push) {
    notifications.notify({ type: 'error', title: t('repo.pipeline.debug.metadata_download_error') });
    return;
  }

  isLoading.value = true;
  try {
    const metadata = await apiClient.getPipelineMetadata(repo.value.id, pipeline.value.number);

    // Create a Blob with the JSON data
    const blob = new Blob([JSON.stringify(metadata, null, 2)], { type: 'application/json' });

    // Create a download link and trigger the download
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = metadataFileName.value;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);

    notifications.notify({ type: 'success', title: t('repo.pipeline.debug.metadata_download_successful') });
  } catch (error) {
    console.error('Error fetching metadata:', error);
    notifications.notify({ type: 'error', title: t('repo.pipeline.debug.metadata_download_error') });
  } finally {
    isLoading.value = false;
  }
}
</script>
