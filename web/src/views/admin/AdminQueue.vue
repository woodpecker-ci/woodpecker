<template>
  <Settings :title="$t('admin.settings.queue.queue')" :description="$t('admin.settings.queue.desc')">
    <template #headerActions>
      <div v-if="queueInfo">
        <div class="flex items-center gap-2">
          <Button
            v-if="queueInfo.paused"
            :text="$t('admin.settings.queue.resume')"
            start-icon="play"
            @click="resumeQueue"
          />
          <Button v-else :text="$t('admin.settings.queue.pause')" start-icon="pause" @click="pauseQueue" />
          <Icon
            :name="queueInfo.paused ? 'pause' : 'play'"
            class="h-6 w-6"
            :class="{
              'text-wp-error-100': queueInfo.paused,
              'text-wp-text-100': !queueInfo.paused,
            }"
          />
        </div>
      </div>
    </template>

    <div class="flex flex-col">
      <AdminQueueStats :stats="queueInfo?.stats" />

      <div v-if="tasks.length > 0" class="flex flex-col">
        <p class="mt-6 mb-2 text-xl">{{ $t('admin.settings.queue.tasks') }}</p>
        <ListItem
          v-for="task in tasks"
          :key="task.id"
          class="bg-wp-background-200! dark:bg-wp-background-100! mb-2 flex-col items-center gap-4"
        >
          <div class="flex w-full items-center justify-between border-b pb-2">
            <div
              class="flex items-center gap-2"
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
                  'text-wp-error-100': task.status === 'waiting_on_deps',
                  'text-wp-state-info-100': task.status === 'running',
                  'text-wp-state-neutral-100': task.status === 'pending',
                }"
              />
              <span>{{ task.id }}</span>
            </div>
            <div class="flex items-center">
              <span class="ml-auto flex gap-2">
                <Badge v-if="task.agent_id !== 0" :label="$t('admin.settings.queue.agent')" :value="task.agent_id" />
                <Badge
                  v-if="task.dependencies"
                  :label="$t('admin.settings.queue.waiting_for')"
                  :value="task.dependencies.join(', ')"
                />
              </span>
            </div>
          </div>
          <div class="flex w-full flex-wrap gap-2">
            <template v-for="(value, label) in task.labels">
              <Badge v-if="value" :key="label" :label="label.toString()" :value="value" />
            </template>
          </div>
        </ListItem>
      </div>
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import AdminQueueStats from '~/components/admin/settings/queue/AdminQueueStats.vue';
import Badge from '~/components/atomic/Badge.vue';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { QueueInfo } from '~/lib/api/types';

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

  return _tasks
    .map((task) => ({
      ...task,
      labels: Object.fromEntries(Object.entries(task.labels).filter(([key]) => key !== 'org-id')),
    }))
    .toSorted((a, b) => a.id - b.id);
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

useWPTitle(computed(() => [t('admin.settings.queue.queue'), t('admin.settings.settings')]));
</script>
