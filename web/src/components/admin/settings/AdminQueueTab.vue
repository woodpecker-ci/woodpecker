<template>
  <Panel>
    <div v-if="queueInfo" class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <div class="ml-2">
        <h1 class="text-xl text-color">{{ $t('admin.settings.queue.queue') }}</h1>
        <p class="text-sm text-color-alt">{{ $t('admin.settings.queue.desc') }}</p>
      </div>

      <div class="ml-auto flex items-center gap-2">
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
        <Icon
          :name="queueInfo.paused ? 'pause' : 'play'"
          :class="{
            'text-red-400': queueInfo.paused,
            'text-lime-400': !queueInfo.paused,
          }"
        />
      </div>
    </div>

    <div class="flex flex-col">
      <pre>{{ queueInfo?.stats }}</pre>

      <div v-if="tasks.length > 0" class="flex flex-col">
        <p class="mt-6 mb-2 text-xl">{{ i18n.t('admin.settings.queue.tasks') }}</p>
        <ListItem v-for="task in tasks" :key="task.id" class="items-center mb-2">
          <div
            class="flex items-center"
            :title="
              task.status === 'pending'
                ? i18n.t('admin.settings.queue.task_pending')
                : task.status === 'running'
                ? i18n.t('admin.settings.queue.task_running')
                : i18n.t('admin.settings.queue.task_waiting_on_deps')
            "
          >
            <Icon
              :name="
                task.status === 'pending'
                  ? 'status-pending'
                  : task.status === 'running'
                  ? 'status-running'
                  : 'status-declined'
              "
              :class="{
                'text-red-400': task.status === 'waiting_on_deps',
                'text-lime-400': task.status === 'running',
                'text-blue-400': task.status === 'pending',
              }"
            />
          </div>
          <span class="ml-2">{{ task.id }}</span>
          <span class="flex ml-auto gap-2">
            <span>{{ task.labels }}</span>
            <span>{{ task.dependencies }}</span>
            <span>{{ task.dep_status }}</span>
          </span>
        </ListItem>
      </div>
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Icon from '~/components/atomic/Icon.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { QueueInfo } from '~/lib/api/types/queue';

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const queueInfo = ref<QueueInfo>();

const tasks = computed(() => {
  const t = [];

  if (queueInfo.value?.running) {
    t.push(...queueInfo.value.running.map((task) => ({ ...task, status: 'running' })));
  }

  if (queueInfo.value?.waiting_on_deps) {
    t.push(...queueInfo.value.waiting_on_deps.map((task) => ({ ...task, status: 'waiting_on_deps' })));
  }

  if (queueInfo.value?.pending) {
    t.push(...queueInfo.value.pending.map((task) => ({ ...task, status: 'pending' })));
  }

  return t;
});

async function loadQueueInfo() {
  queueInfo.value = await apiClient.getQueueInfo();
}

async function pauseQueue() {
  await apiClient.pauseQueue();
  await loadQueueInfo();
  notifications.notify({
    title: i18n.t('admin.settings.queue.paused').toString(),
    type: 'success',
  });
}

async function resumeQueue() {
  await apiClient.resumeQueue();
  await loadQueueInfo();
  notifications.notify({
    title: i18n.t('admin.settings.queue.resumed').toString(),
    type: 'success',
  });
}

const reloadInterval = ref<unknown>();

onMounted(async () => {
  await loadQueueInfo();
  reloadInterval.value = setInterval(async () => {
    await loadQueueInfo();
  }, 5000);
});

onBeforeUnmount(() => {
  if (reloadInterval.value) {
    clearInterval(reloadInterval.value as number);
  }
});
</script>
