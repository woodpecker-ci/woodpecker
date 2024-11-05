<template>
  <Settings :title="$t('repo.settings.crons.crons')" :desc="$t('repo.settings.crons.desc')" docs-url="docs/usage/cron">
    <template #titleActions>
      <Button
        v-if="selectedCron"
        start-icon="back"
        :text="$t('repo.settings.crons.show')"
        @click="selectedCron = undefined"
      />
      <Button v-else start-icon="plus" :text="$t('repo.settings.crons.add')" @click="selectedCron = {}" />
    </template>

    <div v-if="!selectedCron" class="space-y-4 text-wp-text-100">
      <ListItem
        v-for="cron in crons"
        :key="cron.id"
        class="items-center !bg-wp-background-200 !dark:bg-wp-background-100"
      >
        <span class="grid grid-cols-3 w-full">
          <span>{{ cron.name }}</span>
          <span v-if="cron.next_exec && cron.next_exec > 0" class="col-span-2 <md:hidden">
            <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
            {{ $t('repo.settings.crons.next_exec') }}: {{ date.toLocaleString(new Date(cron.next_exec * 1000)) }}
          </span>
          <span v-else class="col-span-2 <md:hidden">{{ $t('repo.settings.crons.not_executed_yet') }}</span>
        </span>
        <IconButton icon="play" class="ml-auto w-8 h-8" :title="$t('repo.settings.crons.run')" @click="runCron(cron)" />
        <IconButton icon="edit" class="w-8 h-8" :title="$t('repo.settings.crons.edit')" @click="selectedCron = cron" />
        <IconButton
          icon="trash"
          class="w-8 h-8 hover:text-wp-control-error-100"
          :is-loading="isDeleting"
          :title="$t('repo.settings.crons.delete')"
          @click="deleteCron(cron)"
        />
      </ListItem>

      <div v-if="crons?.length === 0" class="ml-2">{{ $t('repo.settings.crons.none') }}</div>
    </div>

    <div v-else class="space-y-4">
      <form @submit.prevent="createCron">
        <InputField v-slot="{ id }" :label="$t('repo.settings.crons.name.name')">
          <TextField
            :id="id"
            v-model="selectedCron.name"
            :placeholder="$t('repo.settings.crons.name.placeholder')"
            required
          />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('repo.settings.crons.branch.title')">
          <TextField
            :id="id"
            v-model="selectedCron.branch"
            :placeholder="$t('repo.settings.crons.branch.placeholder')"
          />
        </InputField>

        <InputField
          v-slot="{ id }"
          :label="$t('repo.settings.crons.schedule.title')"
          docs-url="https://pkg.go.dev/github.com/gdgvda/cron#hdr-CRON_Expression_Format"
        >
          <TextField
            :id="id"
            v-model="selectedCron.schedule"
            :placeholder="$t('repo.settings.crons.schedule.placeholder')"
            required
          />
        </InputField>

        <div v-if="isEditingCron" class="ml-auto mb-4">
          <span v-if="selectedCron.next_exec && selectedCron.next_exec > 0" class="text-wp-text-100">
            <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
            {{ $t('repo.settings.crons.next_exec') }}:
            {{ date.toLocaleString(new Date(selectedCron.next_exec * 1000)) }}
          </span>
          <span v-else class="text-wp-text-100">{{ $t('repo.settings.crons.not_executed_yet') }}</span>
        </div>

        <div class="flex gap-2">
          <Button type="button" color="gray" :text="$t('cancel')" @click="selectedCron = undefined" />
          <Button
            type="submit"
            color="green"
            :is-loading="isSaving"
            :text="isEditingCron ? $t('repo.settings.crons.save') : $t('repo.settings.crons.add')"
          />
        </div>
      </form>
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import { computed, inject, ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { useDate } from '~/compositions/useDate';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import type { Cron, Repo } from '~/lib/api/types';
import router from '~/router';

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const repo = inject<Ref<Repo>>('repo');
const selectedCron = ref<Partial<Cron>>();
const isEditingCron = computed(() => !!selectedCron.value?.id);
const date = useDate();

async function loadCrons(page: number): Promise<Cron[] | null> {
  if (!repo?.value) {
    throw new Error("Unexpected: Can't load repo");
  }

  return apiClient.getCronList(repo.value.id, { page });
}

const { resetPage, data: crons } = usePagination(loadCrons, () => !selectedCron.value);

const { doSubmit: createCron, isLoading: isSaving } = useAsyncAction(async () => {
  if (!repo?.value) {
    throw new Error("Unexpected: Can't load repo");
  }

  if (!selectedCron.value) {
    throw new Error("Unexpected: Can't get cron");
  }

  if (isEditingCron.value) {
    await apiClient.updateCron(repo.value.id, selectedCron.value);
  } else {
    await apiClient.createCron(repo.value.id, selectedCron.value);
  }
  notifications.notify({
    title: isEditingCron.value ? i18n.t('repo.settings.crons.saved') : i18n.t('repo.settings.crons.created'),
    type: 'success',
  });
  selectedCron.value = undefined;
  resetPage();
});

const { doSubmit: deleteCron, isLoading: isDeleting } = useAsyncAction(async (_cron: Cron) => {
  if (!repo?.value) {
    throw new Error("Unexpected: Can't load repo");
  }

  await apiClient.deleteCron(repo.value.id, _cron.id);
  notifications.notify({ title: i18n.t('repo.settings.crons.deleted'), type: 'success' });
  resetPage();
});

const { doSubmit: runCron } = useAsyncAction(async (_cron: Cron) => {
  if (!repo?.value) {
    throw new Error("Unexpected: Can't load repo");
  }

  const pipeline = await apiClient.runCron(repo.value.id, _cron.id);
  await router.push({
    name: 'repo-pipeline',
    params: {
      pipelineId: pipeline.number,
    },
  });
});
</script>
