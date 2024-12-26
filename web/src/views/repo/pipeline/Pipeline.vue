<template>
  <Container full-width class="md:min-h-xs flex flex-grow-0 flex-col md:flex-grow md:px-4">
    <div class="flex min-h-0 w-full flex-grow flex-wrap-reverse gap-4 md:flex-nowrap">
      <PipelineStepList
        v-model:selected-step-id="selectedStepId"
        :class="{ 'hidden md:flex': pipeline!.status === 'blocked' }"
        :pipeline="pipeline!"
      />

      <div class="relative flex flex-grow basis-full items-start justify-center md:basis-auto">
        <Container v-if="pipeline!.errors?.some((e) => !e.is_warning)" fill-width class="p-0">
          <Panel>
            <div class="flex flex-col items-center gap-4 text-center">
              <Icon name="status-error" class="text-wp-state-error-100 h-16 w-16" size="1.5rem" />
              <span class="text-xl">{{ $t('repo.pipeline.we_got_some_errors') }}</span>
              <Button color="red" :text="$t('repo.pipeline.show_errors')" :to="{ name: 'repo-pipeline-errors' }" />
            </div>
          </Panel>
        </Container>

        <Container v-else-if="pipeline!.status === 'blocked'" fill-width class="p-0">
          <Panel>
            <div class="flex flex-col items-center gap-4">
              <Icon name="status-blocked" size="1.5rem" class="h-16 w-16" />
              <span class="text-xl">{{ $t('repo.pipeline.protected.awaits') }}</span>
              <div v-if="repoPermissions!.push" class="flex flex-wrap items-center justify-center gap-2">
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

        <Container v-else-if="pipeline!.status === 'declined'" fill-width class="p-0">
          <Panel>
            <div class="flex flex-col items-center gap-4">
              <Icon name="status-declined" size="1.5rem" class="text-wp-state-error-100 h-16 w-16" />
              <p class="text-xl">{{ $t('repo.pipeline.protected.declined') }}</p>
            </div>
          </Panel>
        </Container>

        <PipelineLog
          v-else-if="selectedStepId !== null"
          v-model:step-id="selectedStepId"
          :pipeline="pipeline!"
          class="fixed left-0 top-0 h-full w-full md:absolute"
        />
      </div>
    </div>
  </Container>
</template>

<script lang="ts" setup>
import { computed, inject, toRef, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import Container from '~/components/layout/Container.vue';
import Panel from '~/components/layout/Panel.vue';
import PipelineLog from '~/components/repo/pipeline/PipelineLog.vue';
import PipelineStepList from '~/components/repo/pipeline/PipelineStepList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import type { Pipeline, PipelineStep, Repo, RepoPermissions } from '~/lib/api/types';

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
      const id = Number.parseInt(stepId.value, 10);

      let step = pipeline.value.workflows?.find((workflow) => workflow.pid === id)?.children[0];
      if (step) {
        return step.pid;
      }

      step = pipeline.value?.workflows?.reduce(
        (prev, p) => prev || p.children?.find((c) => c.pid === id),
        undefined as PipelineStep | undefined,
      );
      if (step) {
        return step.pid;
      }

      // return fallback if step-id is provided, but step cannot be found
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
