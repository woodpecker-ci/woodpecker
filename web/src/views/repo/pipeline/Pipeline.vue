<template>
  <Container full-width class="flex flex-col flex-grow md:min-h-xs">
    <div class="flex w-full min-h-0 flex-grow">
      <PipelineStepList
        v-if="pipeline?.workflows && pipeline?.workflows?.length > 0"
        v-model:selected-step-id="selectedStepId"
        :class="{ 'hidden md:flex': pipeline.status === 'blocked' }"
        :pipeline="pipeline"
      />

      <div class="flex items-start justify-center flex-grow relative">
        <Container v-if="selectedStep?.error" fill-width class="py-0">
          <Panel>
            <div class="flex flex-col items-center gap-4">
              <Icon name="status-error" class="w-16 h-16 text-wp-state-error-100" />
              <span class="text-xl">{{ $t('repo.pipeline.we_got_some_errors') }}</span>
              <span class="whitespace-pre">{{ selectedStep?.error }}</span>
            </div>
          </Panel>
        </Container>

        <Container v-else-if="pipeline.errors?.some((e) => !e.is_warning)" fill-width class="py-0">
          <Panel>
            <div class="flex flex-col items-center gap-4">
              <Icon name="status-error" class="w-16 h-16 text-wp-state-error-100" />
              <span class="text-xl">{{ $t('repo.pipeline.we_got_some_errors') }}</span>
              <Button color="red" :text="$t('repo.pipeline.show_errors')" :to="{ name: 'repo-pipeline-errors' }" />
            </div>
          </Panel>
        </Container>

        <Container v-else-if="pipeline.status === 'blocked'" fill-width class="py-0">
          <Panel>
            <div class="flex flex-col items-center gap-4">
              <Icon name="status-blocked" class="w-16 h-16" />
              <span class="text-xl">{{ $t('repo.pipeline.protected.awaits') }}</span>
              <div v-if="repoPermissions.push" class="flex gap-2 flex-wrap items-center justify-center">
                <Button
                  color="blue"
                  :start-icon="forge ?? 'repo'"
                  :text="$t('repo.pipeline.protected.review')"
                  :to="pipeline.forge_url"
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
            </div>
          </Panel>
        </Container>

        <Container v-else-if="pipeline.status === 'declined'" fill-width class="py-0">
          <Panel>
            <div class="flex flex-col items-center gap-4">
              <Icon name="status-declined" class="w-16 h-16 text-wp-state-error-100" />
              <p class="text-xl">{{ $t('repo.pipeline.protected.declined') }}</p>
            </div>
          </Panel>
        </Container>

        <PipelineLog
          v-else-if="selectedStepId !== null"
          v-model:step-id="selectedStepId"
          :pipeline="pipeline"
          class="fixed top-0 left-0 w-full h-full md:absolute"
        />
      </div>
    </div>
  </Container>
</template>

<script lang="ts" setup>
import { computed, inject, Ref, toRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import Container from '~/components/layout/Container.vue';
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

const defaultStepId = computed(() => pipeline.value?.workflows?.[0].children?.[0].pid ?? null);

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
    if (_selectedStepId === null) {
      router.replace({ params: { ...route.params, stepId: '' } });
      return;
    }

    router.replace({ params: { ...route.params, stepId: `${_selectedStepId}` } });
  },
});

const { forge } = useConfig();
const { message } = usePipeline(pipeline);

const selectedStep = computed(() => findStep(pipeline.value.workflows || [], selectedStepId.value || -1));

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
