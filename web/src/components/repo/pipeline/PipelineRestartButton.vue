<template>
  <div ref="root" class="relative flex shrink-0" @keydown.esc="isMenuOpen = false">
    <Button
      class="shrink-0"
      :class="{ 'rounded-r-none!': canRestartFailed }"
      :text="$t('repo.pipeline.actions.restart')"
      :is-loading="isRestarting"
      :disabled="isRestartingFailed"
      data-test="pipeline-restart"
      @click="emit('restart')"
    />
    <template v-if="canRestartFailed">
      <Button
        class="shrink-0 rounded-l-none! border-l-0 px-1!"
        start-icon="chevron-down"
        :title="$t('repo.pipeline.actions.restart_options')"
        :aria-label="$t('repo.pipeline.actions.restart_options')"
        aria-haspopup="menu"
        :aria-expanded="isMenuOpen"
        :is-loading="isRestartingFailed"
        :disabled="isRestarting"
        data-test="pipeline-restart-options"
        @click="isMenuOpen = !isMenuOpen"
      />
      <div
        v-if="isMenuOpen"
        role="menu"
        class="border-wp-background-400 dark:border-wp-background-100 bg-wp-background-100 dark:bg-wp-background-200 dark:shadow-wp-background-500 absolute top-full right-0 z-20 mt-1 w-72 rounded-md border p-1 shadow-lg"
      >
        <button
          type="button"
          role="menuitem"
          class="hover:bg-wp-control-neutral-100 dark:hover:bg-wp-control-neutral-200 flex w-full cursor-pointer flex-col rounded-md px-2 py-1 text-left"
          data-test="pipeline-restart-failed"
          @click="restartFailed"
        >
          <span>{{ $t('repo.pipeline.actions.restart_failed') }}</span>
          <span class="text-wp-text-alt-100 text-sm">{{ $t('repo.pipeline.actions.restart_failed_desc') }}</span>
        </button>
      </div>
    </template>
  </div>
</template>

<script lang="ts" setup>
import { onClickOutside } from '@vueuse/core';
import { ref, useTemplateRef } from 'vue';

import Button from '~/components/atomic/Button.vue';

defineProps<{
  canRestartFailed: boolean;
  isRestarting?: boolean;
  isRestartingFailed?: boolean;
}>();

const emit = defineEmits<{
  (event: 'restart'): void;
  (event: 'restartFailed'): void;
}>();

const isMenuOpen = ref(false);
const root = useTemplateRef('root');
onClickOutside(root, () => {
  isMenuOpen.value = false;
});

function restartFailed() {
  isMenuOpen.value = false;
  emit('restartFailed');
}
</script>
