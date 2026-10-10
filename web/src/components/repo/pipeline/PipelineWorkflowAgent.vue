<template>
  <div ref="root">
    <IconButton
      ref="button"
      icon="info"
      class="hover:bg-white/10!"
      :title="title"
      :disabled="!agent"
      :is-loading="loading"
      :aria-expanded="open"
      :aria-controls="popoverId"
      aria-haspopup="dialog"
      @click="toggle"
    />

    <!-- teleported and fixed: the log box clips its content and its sticky command headers would cover it -->
    <Teleport to="body">
      <div
        v-if="open && agent"
        :id="popoverId"
        ref="popover"
        role="dialog"
        :aria-label="$t('repo.pipeline.agent.title')"
        class="bg-wp-code-200 border-wp-code-300 text-wp-code-text-100 fixed z-30 w-72 max-w-[calc(100vw-2rem)] rounded-md border text-sm shadow-lg"
        :style="{ top: `${position.top}px`, right: `${position.right}px` }"
      >
        <span
          data-arrow
          class="bg-wp-code-200 border-wp-code-300 absolute -top-[7px] h-3 w-3 rotate-45 border-t border-l"
          :style="{ right: `${position.arrowRight}px` }"
        />
        <div
          data-scroll
          class="overflow-y-auto overscroll-contain p-4"
          :style="{ maxHeight: `${position.maxHeight}px` }"
        >
          <h3 class="mb-3 font-bold">{{ $t('repo.pipeline.agent.title') }}</h3>
          <dl class="grid grid-cols-[auto_minmax(0,1fr)] gap-x-6 gap-y-2">
            <dt class="text-wp-code-text-alt-100">{{ $t('repo.pipeline.agent.workflow') }}</dt>
            <dd class="break-words">{{ workflow?.name }}</dd>
            <dt class="text-wp-code-text-alt-100">{{ $t('repo.pipeline.agent.agent') }}</dt>
            <dd class="break-words">{{ agent.name || `#${agent.id}` }}</dd>
            <dt class="text-wp-code-text-alt-100">{{ $t('repo.pipeline.agent.platform') }}</dt>
            <dd class="break-words">{{ agent.platform }}</dd>
            <dt class="text-wp-code-text-alt-100">{{ $t('repo.pipeline.agent.backend') }}</dt>
            <dd class="break-words">{{ agent.backend }}</dd>
          </dl>
          <div class="border-wp-code-300 mt-3 border-t pt-3">
            <button
              type="button"
              class="text-wp-link-100 cursor-pointer hover:underline"
              :aria-expanded="showLabels"
              @click="showLabels = !showLabels"
            >
              {{ showLabels ? $t('repo.pipeline.agent.hide_labels') : $t('repo.pipeline.agent.show_labels') }}
            </button>
            <template v-if="showLabels">
              <ul v-if="labels.length > 0" class="mt-2 flex flex-col gap-1 font-mono text-xs">
                <li v-for="[key, value] in labels" :key="key" class="flex flex-wrap">
                  <span class="text-wp-code-text-alt-100 break-all">{{ `${key}=` }}</span>
                  <span class="break-all">{{ value }}</span>
                </li>
              </ul>
              <p v-else class="text-wp-code-text-alt-100 mt-2">{{ $t('repo.pipeline.agent.no_labels') }}</p>
            </template>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script lang="ts" setup>
import { onClickOutside, onKeyStroke, useEventListener } from '@vueuse/core';
import { computed, ref, useId, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import IconButton from '~/components/atomic/IconButton.vue';
import useApiClient from '~/compositions/useApiClient';
import useConfig from '~/compositions/useConfig';
import { requiredInject } from '~/compositions/useInjectProvide';
import { ApiRequestError } from '~/lib/api/client';
import type { AgentSnapshot, PipelineWorkflow } from '~/lib/api/types';

const props = defineProps<{
  workflow?: PipelineWorkflow;
  pipelineNumber: number;
}>();

// px between icon and popover, and kept free to the viewport edges
const POPOVER_GAP = 8;
const VIEWPORT_MARGIN = 16;
const ARROW_HALF_WIDTH = 6;

const apiClient = useApiClient();
const { user } = useConfig();
const repo = requiredInject('repo');
const { t } = useI18n();

const root = useTemplateRef('root');
const button = useTemplateRef('button');
const popover = useTemplateRef('popover');
const popoverId = useId();

const open = ref(false);
const showLabels = ref(false);
const agent = ref<AgentSnapshot>();
const loading = ref(false);
const forbidden = ref(false);
const position = ref({ top: 0, right: VIEWPORT_MARGIN, maxHeight: 0, arrowRight: 12 });

const labels = computed(() => Object.entries(agent.value?.custom_labels ?? {}));

const title = computed(() => {
  if (!props.workflow?.agent_id) {
    return t('repo.pipeline.agent.not_assigned');
  }
  if (!user || forbidden.value) {
    return t('repo.pipeline.agent.no_permission');
  }
  if (loading.value) {
    return t('repo.pipeline.agent.loading');
  }
  if (!agent.value) {
    return t('repo.pipeline.agent.unavailable');
  }
  return t('repo.pipeline.agent.title');
});

// The server keeps a snapshot of the agent per assignment and decides who may read it,
// so it is only loaded again for another workflow or a new assignment.
watch(
  [() => props.workflow?.id, () => props.workflow?.agent_id],
  async ([workflowId, agentId], [oldWorkflowId], onCleanup) => {
    let stale = false;
    onCleanup(() => {
      stale = true;
    });

    if (workflowId !== oldWorkflowId) {
      open.value = false;
      showLabels.value = false;
    }
    agent.value = undefined;
    forbidden.value = false;
    loading.value = false;

    if (!user || workflowId === undefined || !agentId) {
      return;
    }

    loading.value = true;
    try {
      const result = await apiClient.getWorkflowAgent(repo.value.id, props.pipelineNumber, workflowId);
      if (!stale) {
        agent.value = result;
      }
    } catch (error) {
      if (!stale) {
        forbidden.value = error instanceof ApiRequestError && error.status === 403;
      }
    } finally {
      if (!stale) {
        loading.value = false;
      }
    }
  },
  { immediate: true },
);

watch(agent, (newAgent) => {
  if (!newAgent) {
    open.value = false;
  }
});

function updatePosition() {
  const buttonEl = button.value?.$el as HTMLElement | undefined;
  if (!buttonEl) {
    return;
  }
  const buttonRect = buttonEl.getBoundingClientRect();
  // align the popover with the right edge of the toolbar the icon sits in
  const toolbarRect = (root.value?.parentElement ?? buttonEl).getBoundingClientRect();
  const right = Math.max(VIEWPORT_MARGIN, window.innerWidth - toolbarRect.right);
  const top = buttonRect.bottom + POPOVER_GAP;
  position.value = {
    top,
    right,
    // the content scrolls instead of leaving the viewport on short screens
    maxHeight: Math.max(0, window.innerHeight - top - VIEWPORT_MARGIN),
    // point the arrow at the center of the icon
    arrowRight: window.innerWidth - right - (buttonRect.left + buttonRect.width / 2) - ARROW_HALF_WIDTH,
  };
}

function toggle() {
  if (!open.value) {
    updatePosition();
  }
  open.value = !open.value;
}

useEventListener(window, 'resize', () => open.value && updatePosition(), { passive: true });
useEventListener(window, 'scroll', () => open.value && updatePosition(), { capture: true, passive: true });

onClickOutside(
  root,
  () => {
    open.value = false;
  },
  { ignore: [popover] },
);

onKeyStroke('Escape', () => {
  if (!open.value) {
    return;
  }
  open.value = false;
  (button.value?.$el as HTMLElement | undefined)?.focus();
});
</script>
