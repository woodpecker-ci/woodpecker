<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <div class="ml-2">
        <h1 class="text-xl text-color">{{ $t('admin.settings.queue.queue') }}</h1>
        <p class="text-sm text-color-alt">{{ $t('admin.settings.queue.desc') }}</p>
      </div>

      <template v-if="queueInfo">
        <Button
          v-if="queueInfo.paused"
          class="ml-auto"
          :text="$t('admin.settings.queue.resume')"
          start-icon="play"
          @click="resumeQueue"
        />
        <Button
          v-else
          class="ml-auto"
          :text="$t('admin.settings.queue.pause')"
          start-icon="pause"
          @click="pauseQueue"
        />
      </template>
    </div>

    <div>
      <pre>{{ queueInfo }}</pre>
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { QueueInfo } from '~/lib/api/types/queue';

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const queueInfo = ref<QueueInfo>();

async function loadQueueInfo() {
  queueInfo.value = await apiClient.getQueueInfo();
}

async function pauseQueue() {
  await apiClient.pauseQueue();
  await loadQueueInfo();
}

async function resumeQueue() {
  await apiClient.resumeQueue();
  await loadQueueInfo();
}

onMounted(async () => {
  await loadQueueInfo();
  setInterval(async () => {
    await loadQueueInfo();
  }, 5000);
});
</script>
