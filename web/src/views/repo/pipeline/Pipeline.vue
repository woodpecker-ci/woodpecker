<template>
  <FluidContainer full-width class="flex flex-col flex-grow md:min-h-xs">
    <div class="flex w-full min-h-0 flex-grow">
      <PipelineStepList
        v-if="pipeline?.workflows?.length || 0 > 0"
        v-model:selected-step-id="selectedStepId"
        :class="{ 'hidden md:flex': pipeline.status === 'blocked' }"
        :pipeline="pipeline"
      />

      <div class="flex flex-grow relative">
        <PipelineInfo v-if="error">
          <Icon name="status-error" class="w-16 h-16 text-wp-state-error-100" />
          <div class="flex flex-wrap items-center justify-center gap-2 text-xl">
            <span class="capitalize">{{ $t('repo.pipeline.execution_error') }}:</span>
            <span>{{ error }}</span>
          </div>
        </PipelineInfo>

        <PipelineInfo v-else-if="pipeline.status === 'blocked'">
          <Icon name="status-blocked" class="w-16 h-16" />
          <span class="text-xl">{{ $t('repo.pipeline.protected.awaits') }}</span>
          <div v-if="repoPermissions.push" class="flex space-x-4">
            <Button
              color="blue"
              :start-icon="forge ?? 'repo'"
              :text="$t('repo.pipeline.protected.review')"
              :is-loading="isApprovingPipeline"
              :to="pipeline.link_url"
              :title="message"
            />
            <Button
              color="green"
              :text="$t('repo.pipeline.protected.approve')"
              :is-loading="isApprovingPipeline"
              @click="approvePipeline"
            />
            <Button
              color="red"
              :text="$t('repo.pipeline.protected.decline')"
              :is-loading="isDecliningPipeline"
              @click="declinePipeline"
            />
          </div>
        </PipelineInfo>

        <PipelineInfo v-else-if="pipeline.status === 'declined'">
          <Icon name="status-blocked" class="w-16 h-16" />
          <p class="text-xl">{{ $t('repo.pipeline.protected.declined') }}</p>
        </PipelineInfo>

        <PipelineLog
          v-else-if="selectedStepId"
          v-model:step-id="selectedStepId"
          :pipeline="pipeline"
          class="fixed top-0 left-0 w-full h-full md:absolute"
        />
      </div>
    </div>
  </FluidContainer>
</template>

<script lang="ts" setup>
import { computed, inject, Ref, toRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import PipelineLog from '~/components/repo/pipeline/PipelineLog.vue';
import PipelineStepList from '~/components/repo/pipeline/PipelineStepList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useConfig from '~/compositions/useConfig';
import useNotifications from '~/compositions/useNotifications';
import usePipeline from '~/compositions/usePipeline';
import { Pipeline, PipelineStep, Repo, RepoPermissions } from '~/lib/api/types';
import { findStep } from '~/utils/helpers';

const props = defineProps<{
  stepId?: string | null;
}>();

const apiClient = useApiClient();
const router = useRouter();
const route = useRoute();
const notifications = useNotifications();
const i18n = useI18n();

const pipeline = inject<Ref<Pipeline>>('pipeline');
const repo = inject<Ref<Repo>>('repo');
const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
if (!repo || !repoPermissions || !pipeline) {
  throw new Error('Unexpected: "repo", "repoPermissions" & "pipeline" should be provided at this place');
}

const stepId = toRef(props, 'stepId');

const defaultStepId = computed(() => {
  if (!pipeline.value || !pipeline.value.workflows || !pipeline.value.workflows[0].children) {
    return null;
  }

  return pipeline.value.workflows[0].children[0].pid;
});

const selectedStepId = computed({
  get() {
    if (stepId.value !== '' && stepId.value !== null && stepId.value !== undefined) {
      const id = parseInt(stepId.value, 10);
      const step = pipeline.value?.workflows?.reduce(
        (prev, p) => prev || p.children?.find((c) => c.pid === id),
        undefined as PipelineStep | undefined,
      );
      if (step) {
        return step.pid;
      }

      // return fallback if step-id is provided, but step can not be found
      return defaultStepId.value;
    }

    // is opened on >= md-screen
    if (window.innerWidth > 768) {
      return defaultStepId.value;
    }

    return null;
  },
  set(_selectedStepId: number | null) {
    if (!_selectedStepId) {
      router.replace({ params: { ...route.params, stepId: '' } });
      return;
    }

    router.replace({ params: { ...route.params, stepId: `${_selectedStepId}` } });
  },
});

const { forge } = useConfig();
const { message } = usePipeline(pipeline);

const selectedStep = computed(() => findStep(pipeline.value.workflows || [], selectedStepId.value || -1));
const error = computed(() => pipeline.value?.error || selectedStep.value?.error);

const { doSubmit: approvePipeline, isLoading: isApprovingPipeline } = useAsyncAction(async () => {
  if (!repo) {
    throw new Error('Unexpected: Repo is undefined');
  }

  await apiClient.approvePipeline(repo.value.id, `${pipeline.value.number}`);
  notifications.notify({ title: i18n.t('repo.pipeline.protected.approve_success'), type: 'success' });
});

const { doSubmit: declinePipeline, isLoading: isDecliningPipeline } = useAsyncAction(async () => {
  if (!repo) {
    throw new Error('Unexpected: Repo is undefined');
  }

  await apiClient.declinePipeline(repo.value.id, `${pipeline.value.number}`);
  notifications.notify({ title: i18n.t('repo.pipeline.protected.decline_success'), type: 'success' });
});
</script>
