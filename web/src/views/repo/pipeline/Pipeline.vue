<template>
  <FluidContainer full-width class="flex flex-col flex-grow">
    <div class="flex w-full min-h-0 flex-grow">
      <PipelineStepList
        v-if="pipeline?.steps?.length || 0 > 0"
        v-model:selected-step-id="selectedStepId"
        :class="{ 'hidden md:flex': pipeline.status === 'blocked' }"
        :pipeline="pipeline"
      />

      <div class="flex flex-grow relative">
        <div v-if="error" class="flex flex-col p-4">
          <span class="text-red-400 font-bold text-xl mb-2">{{ $t('repo.pipeline.execution_error') }}</span>
          <span class="text-red-400">{{ error }}</span>
        </div>

        <div v-else-if="pipeline.status === 'blocked'" class="flex flex-col flex-grow justify-center items-center p-2">
          <Icon name="status-blocked" class="w-32 h-32 text-color" />
          <p class="text-xl text-color">{{ $t('repo.pipeline.protected.awaits') }}</p>
          <div v-if="repoPermissions.push" class="flex mt-2 space-x-4">
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

        <div v-else-if="pipeline.status === 'declined'" class="flex flex-col flex-grow justify-center items-center">
          <Icon name="status-blocked" class="w-32 h-32 text-color" />
          <p class="text-xl text-color">{{ $t('repo.pipeline.protected.declined') }}</p>
        </div>

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

<script lang="ts">
import { computed, defineComponent, inject, PropType, Ref, toRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import PipelineLog from '~/components/repo/pipeline/PipelineLog.vue';
import PipelineStepList from '~/components/repo/pipeline/PipelineStepList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { Pipeline, PipelineStep, Repo, RepoPermissions } from '~/lib/api/types';
import { findStep } from '~/utils/helpers';

export default defineComponent({
  name: 'Pipeline',

  components: {
    Button,
    PipelineStepList,
    Icon,
    PipelineLog,
    FluidContainer,
  },

  props: {
    stepId: {
      type: String as PropType<string | null>,
      default: null,
    },
  },

  setup(props) {
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
      if (!pipeline.value || !pipeline.value.steps || !pipeline.value.steps[0].children) {
        return null;
      }

      return pipeline.value.steps[0].children[0].pid;
    });

    const selectedStepId = computed({
      get() {
        if (stepId.value !== '' && stepId.value !== null) {
          const id = parseInt(stepId.value, 10);
          const step = pipeline.value?.steps?.reduce(
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

    const selectedStep = computed(() => findStep(pipeline.value.steps || [], selectedStepId.value || -1));
    const error = computed(() => pipeline.value?.error || selectedStep.value?.error);

    const { doSubmit: approvePipeline, isLoading: isApprovingPipeline } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await apiClient.approvePipeline(repo.value.owner, repo.value.name, `${pipeline.value.number}`);
      notifications.notify({ title: i18n.t('repo.pipeline.protected.approve_success'), type: 'success' });
    });

    const { doSubmit: declinePipeline, isLoading: isDecliningPipeline } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await apiClient.declinePipeline(repo.value.owner, repo.value.name, `${pipeline.value.number}`);
      notifications.notify({ title: i18n.t('repo.pipeline.protected.decline_success'), type: 'success' });
    });

    return {
      repoPermissions,
      selectedStepId,
      pipeline,
      error,
      isApprovingPipeline,
      isDecliningPipeline,
      approvePipeline,
      declinePipeline,
    };
  },
});
</script>
