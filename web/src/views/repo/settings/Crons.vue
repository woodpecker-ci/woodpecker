<template>
  <Settings
    :title="$t('repo.settings.crons.crons')"
    :description="$t('repo.settings.crons.desc')"
    docs-url="docs/usage/cron"
  >
    <template #headerActions>
      <Button
        v-if="selectedCron"
        start-icon="back"
        :text="$t('repo.settings.crons.show')"
        @click="selectedCron = undefined"
      />
      <Button v-else start-icon="plus" :text="$t('repo.settings.crons.add')" @click="selectedCron = {}" />
    </template>

    <div v-if="!selectedCron" class="text-wp-text-100 space-y-4">
      <ListItem
        v-for="cron in crons"
        :key="cron.id"
        class="bg-wp-background-200! dark:bg-wp-background-100! items-center"
      >
        <span class="grid w-full grid-cols-3">
          <span>{{ cron.name }}</span>
          <span v-if="cron.next_exec && cron.next_exec > 0" class="md:display-unset col-span-2 hidden">
            <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
            {{ $t('repo.settings.crons.next_exec') }}: {{ date.toLocaleString(new Date(cron.next_exec * 1000)) }}
          </span>
          <span v-else class="md:display-unset col-span-2 hidden">{{
            $t('repo.settings.crons.not_executed_yet')
          }}</span>
        </span>
        <IconButton
          icon="play-outline"
          class="ml-auto h-8 w-8"
          :title="$t('repo.settings.crons.run')"
          @click="runCron(cron)"
        />
        <IconButton icon="edit" class="h-8 w-8" :title="$t('repo.settings.crons.edit')" @click="selectedCron = cron" />
        <IconButton
          icon="trash"
          class="hover:text-wp-error-100 h-8 w-8"
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

        <div v-if="isEditingCron" class="mb-4 ml-auto">
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
import { computed, ref } from 'vue';
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
import { requiredInject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Cron } from '~/lib/api/types';
import router from '~/router';

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const repo = requiredInject('repo');
const selectedCron = ref<Partial<Cron>>();
const isEditingCron = computed(() => !!selectedCron.value?.id);
const date = useDate();

async function loadCrons(page: number): Promise<Cron[] | null> {
  return apiClient.getCronList(repo.value.id, { page });
}

const { resetPage, data: crons } = usePagination(loadCrons, () => !selectedCron.value);

const { doSubmit: createCron, isLoading: isSaving } = useAsyncAction(async () => {
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
  await resetPage();
});

const { doSubmit: deleteCron, isLoading: isDeleting } = useAsyncAction(async (_cron: Cron) => {
  await apiClient.deleteCron(repo.value.id, _cron.id);
  notifications.notify({ title: i18n.t('repo.settings.crons.deleted'), type: 'success' });
  await resetPage();
});

const { doSubmit: runCron } = useAsyncAction(async (_cron: Cron) => {
  const pipeline = await apiClient.runCron(repo.value.id, _cron.id);
  await router.push({
    name: 'repo-pipeline',
    params: {
      pipelineId: pipeline.number,
    },
  });
});

useWPTitle(computed(() => [i18n.t('repo.settings.crons.crons'), repo.value.full_name]));
</script>
