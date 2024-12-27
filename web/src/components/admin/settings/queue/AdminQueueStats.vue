<template>
  <div v-if="stats" class="flex justify-center">
    <div
      class="border-wp-background-300 bg-wp-background-200 dark:bg-wp-background-100 px-5 py-5 border rounded-md w-full text-wp-text-100"
    >
      <div class="flex w-full">
        <h3 class="flex-1 font-semibold text-lg uppercase leading-tight">
          {{ $t('admin.settings.queue.stats.completed_count') }}
        </h3>
      </div>
      <div class="relative transition-all duration-500 overflow-hidden">
        <div>
          <div class="pb-4 lg:pb-6">
            <h4 class="inline-block font-semibold text-2xl lg:text-3xl leading-tight">
              {{ stats.completed_count }}
            </h4>
          </div>
          <div v-if="total > 0" class="pb-4 lg:pb-6">
            <div class="flex rounded-full h-3 transition-all duration-500 overflow-hidden">
              <div
                v-for="item in data"
                :key="item.key"
                class="h-full"
                :class="`${item.color}`"
                :style="{ width: `${item.percentage}%` }"
              >
                &nbsp;
              </div>
            </div>
          </div>
          <div class="flex sm:flex-wrap -mx-4">
            <div
              v-for="(item, index) in data"
              :key="item.key"
              class="px-4 sm:w-full md:w-1/4"
              :class="{ 'md:border-l border-gray-300 dark:border-gray-600': index !== 0 }"
            >
              <div class="text-ellipsis text-sm whitespace-nowrap overflow-hidden">
                <span class="inline-block mr-1 rounded-full w-2 h-2 align-middle" :class="`${item.color}`">&nbsp;</span>
                <span class="align-middle">{{ item.label }}</span>
              </div>
              <div class="font-medium text-lg">
                {{ item.value }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import type { QueueStats } from '~/lib/api/types/queue';

const props = defineProps<{
  stats?: QueueStats;
}>();

const { t } = useI18n();

const total = computed(() => {
  if (!props.stats) {
    return 0;
  }

  return (
    props.stats.worker_count + props.stats.running_count + props.stats.pending_count + props.stats.waiting_on_deps_count
  );
});

const data = computed(() => {
  if (!props.stats) {
    return [];
  }

  return [
    {
      key: 'worker_count',
      label: t('admin.settings.queue.stats.worker_count'),
      value: props.stats.worker_count,
      percentage: total.value > 0 ? (props.stats.worker_count / total.value) * 100 : 0,
      color: 'bg-wp-state-ok-100',
    },
    {
      key: 'running_count',
      label: t('admin.settings.queue.stats.running_count'),
      value: props.stats.running_count,
      percentage: total.value > 0 ? (props.stats.running_count / total.value) * 100 : 100,
      color: 'bg-wp-state-info-100',
    },
    {
      key: 'pending_count',
      label: t('admin.settings.queue.stats.pending_count'),
      value: props.stats.pending_count,
      percentage: total.value > 0 ? (props.stats.pending_count / total.value) * 100 : 0,
      color: 'bg-wp-state-neutral-100',
    },
    {
      key: 'waiting_on_deps_count',
      label: t('admin.settings.queue.stats.waiting_on_deps_count'),
      value: props.stats.waiting_on_deps_count,
      percentage: total.value > 0 ? (props.stats.waiting_on_deps_count / total.value) * 100 : 0,
      color: 'bg-wp-state-error-100',
    },
  ];
});
</script>
