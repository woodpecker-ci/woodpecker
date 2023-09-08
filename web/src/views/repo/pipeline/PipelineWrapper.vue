<template>
  <template v-if="pipeline && repo">
    <Scaffold
      v-model:activeTab="activeTab"
      enable-tabs
      disable-hash-mode
      :go-back="goBack"
      :fluid-content="activeTab !== 'tasks'"
      :full-width="true"
    >
      <template #title>{{ repo.full_name }}</template>

      <template #titleActions>
        <div class="flex md:items-center flex-col gap-2 md:flex-row md:justify-between min-w-0">
          <div class="flex content-start gap-2 min-w-0">
            <PipelineStatusIcon :status="pipeline.status" class="flex flex-shrink-0" />
            <span class="flex-shrink-0 text-center">{{ $t('repo.pipeline.pipeline', { pipelineId }) }}</span>
            <span class="hidden md:inline-block">-</span>
            <span class="min-w-0 whitespace-nowrap overflow-hidden overflow-ellipsis" :title="message">{{
              title
            }}</span>
          </div>

          <template v-if="repoPermissions.push && pipeline.status !== 'declined'">
            <div class="flex content-start gap-x-2">
              <Button
                v-if="pipeline.status === 'pending' || pipeline.status === 'running'"
                class="flex-shrink-0"
                :text="$t('repo.pipeline.actions.cancel')"
                :is-loading="isCancelingPipeline"
                @click="cancelPipeline"
              />
              <Button
                v-else-if="pipeline.status !== 'blocked'"
                class="flex-shrink-0"
                :text="$t('repo.pipeline.actions.restart')"
                :is-loading="isRestartingPipeline"
                @click="restartPipeline"
              />
              <Button
                v-if="pipeline.status === 'success'"
                class="flex-shrink-0"
                :text="$t('repo.pipeline.actions.deploy')"
                @click="showDeployPipelinePopup = true"
              />
              <DeployPipelinePopup
                :pipeline-number="pipelineId"
                :open="showDeployPipelinePopup"
                @close="showDeployPipelinePopup = false"
              />
            </div>
          </template>
        </div>
      </template>

      <template #tabActions>
        <div class="flex gap-x-4">
          <div class="flex space-x-1 items-center flex-shrink-0" :title="created">
            <Icon name="since" />
            <span>{{ since }}</span>
          </div>
          <div class="flex space-x-1 items-center flex-shrink-0">
            <Icon name="duration" />
            <span>{{ duration }}</span>
          </div>
        </div>
      </template>

      <Tab id="tasks" :title="$t('repo.pipeline.tasks')" />
      <Tab id="config" :title="$t('repo.pipeline.config')" />
      <Tab
        v-if="
          (pipeline.event === 'push' || pipeline.event === 'pull_request') &&
          pipeline.changed_files &&
          pipeline.changed_files.length > 0
        "
        id="changed-files"
        :title="$t('repo.pipeline.files', { files: pipeline.changed_files.length })"
      />
      <router-view />
    </Scaffold>
  </template>
</template>

<script lang="ts" setup>
import { computed, inject, onBeforeUnmount, onMounted, provide, Ref, ref, toRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import DeployPipelinePopup from '~/components/layout/popups/DeployPipelinePopup.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { useFavicon } from '~/compositions/useFavicon';
import useNotifications from '~/compositions/useNotifications';
import usePipeline from '~/compositions/usePipeline';
import { useRouteBack } from '~/compositions/useRouteBack';
import { Repo, RepoPermissions } from '~/lib/api/types';
import { usePipelineStore } from '~/store/pipelines';

const props = defineProps<{
  repoId: string;
  pipelineId: string;
}>();

const apiClient = useApiClient();
const route = useRoute();
const router = useRouter();
const notifications = useNotifications();
const favicon = useFavicon();
const i18n = useI18n();

const pipelineStore = usePipelineStore();
const pipelineId = toRef(props, 'pipelineId');
const _repoId = toRef(props, 'repoId');
const repositoryId = computed(() => parseInt(_repoId.value, 10));
const repo = inject<Ref<Repo>>('repo');
const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
if (!repo || !repoPermissions) {
  throw new Error('Unexpected: "repo" & "repoPermissions" should be provided at this place');
}

const pipeline = pipelineStore.getPipeline(repositoryId, pipelineId);
const { since, duration, created, message, title } = usePipeline(pipeline);
provide('pipeline', pipeline);

const showDeployPipelinePopup = ref(false);

async function loadPipeline(): Promise<void> {
  if (!repo) {
    throw new Error('Unexpected: Repo is undefined');
  }

  await pipelineStore.loadPipeline(repo.value.id, parseInt(pipelineId.value, 10));

  favicon.updateStatus(pipeline.value?.status);
}

const { doSubmit: cancelPipeline, isLoading: isCancelingPipeline } = useAsyncAction(async () => {
  if (!repo) {
    throw new Error('Unexpected: Repo is undefined');
  }

  if (!pipeline.value?.number) {
    throw new Error('Unexpected: Pipeline number not found');
  }

  await apiClient.cancelPipeline(repo.value.id, pipeline.value.number);
  notifications.notify({ title: i18n.t('repo.pipeline.actions.cancel_success'), type: 'success' });
});

const { doSubmit: restartPipeline, isLoading: isRestartingPipeline } = useAsyncAction(async () => {
  if (!repo) {
    throw new Error('Unexpected: Repo is undefined');
  }

  const newPipeline = await apiClient.restartPipeline(repo.value.id, pipelineId.value, {
    fork: true,
  });
  notifications.notify({ title: i18n.t('repo.pipeline.actions.restart_success'), type: 'success' });
  await router.push({
    name: 'repo-pipeline',
    params: { pipelineId: newPipeline.number },
  });
});

onMounted(loadPipeline);
watch([repositoryId, pipelineId], loadPipeline);
onBeforeUnmount(() => {
  favicon.updateStatus('default');
});

const activeTab = computed({
  get() {
    if (route.name === 'repo-pipeline-changed-files') {
      return 'changed-files';
    }

    if (route.name === 'repo-pipeline-config') {
      return 'config';
    }

    return 'tasks';
  },
  set(tab: string) {
    if (tab === 'tasks') {
      router.replace({ name: 'repo-pipeline' });
    }

    if (tab === 'changed-files') {
      router.replace({ name: 'repo-pipeline-changed-files' });
    }

    if (tab === 'config') {
      router.replace({ name: 'repo-pipeline-config' });
    }
  },
});

const goBack = useRouteBack({ name: 'repo' });
</script>
