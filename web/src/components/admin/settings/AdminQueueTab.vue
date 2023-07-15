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
            'text-woodpecker-300': !queueInfo.paused,
          }"
        />
      </div>
    </div>

    <div class="flex flex-col">
      <AdminQueueStats :stats="queueInfo?.stats" />

      <div v-if="tasks.length > 0" class="flex flex-col">
        <p class="mt-6 mb-2 text-xl">{{ $t('admin.settings.queue.tasks') }}</p>
        <ListItem v-for="task in tasks" :key="task.id" class="items-center mb-2">
          <div
            class="flex items-center"
            :title="
              task.status === 'pending'
                ? $t('admin.settings.queue.task_pending')
                : task.status === 'running'
                ? $t('admin.settings.queue.task_running')
                : $t('admin.settings.queue.task_waiting_on_deps')
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
                'text-blue-400': task.status === 'running',
                'text-gray-400': task.status === 'pending',
              }"
            />
          </div>
          <span class="ml-2">{{ task.id }}</span>
          <span class="flex ml-auto gap-2">
            <Badge v-if="task.agent_id !== 0" :label="$t('admin.settings.queue.agent')" :value="task.agent_id" />
            <template v-for="(value, label) in task.labels">
              <Badge v-if="value" :key="label" :label="label.toString()" :value="value" />
            </template>
            <Badge
              v-if="task.dependencies"
              :label="$t('admin.settings.queue.waiting_for')"
              :value="task.dependencies.join(', ')"
            />
          </span>
        </ListItem>
      </div>
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { QueueInfo } from '~/lib/api/types/queue';

import AdminQueueStats from './queue/AdminQueueStats.vue';

const apiClient = useApiClient();
const notifications = useNotifications();
const { t } = useI18n();

const queueInfo = ref<QueueInfo>();

const tasks = computed(() => {
  const _tasks = [];

  if (queueInfo.value?.running) {
    _tasks.push(...queueInfo.value.running.map((task) => ({ ...task, status: 'running' })));
  }

  if (queueInfo.value?.pending) {
    _tasks.push(...queueInfo.value.pending.map((task) => ({ ...task, status: 'pending' })));
  }

  if (queueInfo.value?.waiting_on_deps) {
    _tasks.push(...queueInfo.value.waiting_on_deps.map((task) => ({ ...task, status: 'waiting_on_deps' })));
  }

  return _tasks.sort((a, b) => a.id - b.id);
});

async function loadQueueInfo() {
  queueInfo.value = await apiClient.getQueueInfo();
}

async function pauseQueue() {
  await apiClient.pauseQueue();
  await loadQueueInfo();
  notifications.notify({
    title: t('admin.settings.queue.paused'),
    type: 'success',
  });
}

async function resumeQueue() {
  await apiClient.resumeQueue();
  await loadQueueInfo();
  notifications.notify({
    title: t('admin.settings.queue.resumed'),
    type: 'success',
  });
}

const reloadInterval = ref<number>();
onMounted(async () => {
  await loadQueueInfo();
  reloadInterval.value = window.setInterval(async () => {
    await loadQueueInfo();
  }, 5000);
});

onBeforeUnmount(() => {
  if (reloadInterval.value) {
    window.clearInterval(reloadInterval.value);
  }
});
</script>
