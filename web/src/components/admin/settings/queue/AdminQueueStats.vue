<template>
  <div v-if="stats" class="flex justify-center">
    <div class="bg-gray-100 dark:bg-dark-gray-600 text-color dark:text-gray-400 rounded-md py-5 px-5 w-full">
      <div class="flex w-full">
        <h3 class="text-lg font-semibold leading-tight uppercase flex-1">
          {{ $t('admin.settings.queue.stats.completed_count') }}
        </h3>
      </div>
      <div class="relative overflow-hidden transition-all duration-500">
        <div>
          <div class="pb-4 lg:pb-6">
            <h4 class="text-2xl lg:text-3xl font-semibold leading-tight inline-block">
              {{ stats.completed_count }}
            </h4>
          </div>
          <div class="pb-4 lg:pb-6">
            <div class="overflow-hidden rounded-full h-3 flex transition-all duration-500">
              <div
                v-for="item in data"
                :key="item.key"
                class="h-full"
                :class="`${item.color}`"
                :style="{ width: `${item.perc}%` }"
              >
                &nbsp;
              </div>
            </div>
          </div>
          <div class="flex -mx-4 sm:flex-wrap">
            <div
              v-for="(item, index) in data"
              :key="item.key"
              class="px-4 md:w-1/4 sm:w-full"
              :class="{ 'md:border-l border-gray-300 dark:border-gray-600': index !== 0 }"
            >
              <div class="text-sm whitespace-nowrap overflow-hidden text-ellipsis">
                <span class="inline-block w-2 h-2 rounded-full mr-1 align-middle" :class="`${item.color}`">&nbsp;</span>
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

import { QueueStats } from '~/lib/api/types/queue';

const { t } = useI18n();

const props = defineProps<{
  stats?: QueueStats;
}>();

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
      perc: total.value > 0 ? (props.stats.worker_count / total.value) * 100 : 0,
      color: 'bg-lime-400',
    },
    {
      key: 'running_count',
      label: t('admin.settings.queue.stats.running_count'),
      value: props.stats.running_count,
      perc: total.value > 0 ? (props.stats.running_count / total.value) * 100 : 100,
      color: 'bg-blue-400',
    },
    {
      key: 'pending_count',
      label: t('admin.settings.queue.stats.pending_count'),
      value: props.stats.pending_count,
      perc: total.value > 0 ? (props.stats.pending_count / total.value) * 100 : 0,
      color: 'bg-gray-400',
    },
    {
      key: 'waiting_on_deps_count',
      label: t('admin.settings.queue.stats.waiting_on_deps_count'),
      value: props.stats.waiting_on_deps_count,
      perc: total.value > 0 ? (props.stats.waiting_on_deps_count / total.value) * 100 : 0,
      color: 'bg-red-400',
    },
  ];
});
</script>
