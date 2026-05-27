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
      <Button
        v-else
        start-icon="plus"
        :text="$t('repo.settings.crons.add')"
        @click="selectedCron = { timezone: 'UTC' }"
      />
    </template>

    <div v-if="!selectedCron" class="text-wp-text-100 space-y-4">
      <ListItem
        v-for="cron in crons"
        :key="cron.id"
        class="bg-wp-background-200! dark:bg-wp-background-200! items-center"
      >
        <span class="grid w-full grid-cols-3">
          <span>{{ cron.name }}</span>
          <span
            v-if="cron.enabled && cron.next_exec && cron.next_exec > 0"
            :title="$t('repo.settings.crons.your_timezone')"
            class="md:display-unset col-span-2 hidden"
          >
            {{
              $t('repo.settings.crons.next_exec_local', { local: date.toLocaleString(new Date(cron.next_exec * 1000)) })
            }}
          </span>
          <span v-else-if="cron.enabled" class="md:display-unset col-span-2 hidden">{{
            $t('repo.settings.crons.not_executed_yet')
          }}</span>
          <span v-else class="md:display-unset col-span-2 hidden">{{ $t('disabled') }}</span>
        </span>
        <div class="flex items-center gap-2">
          <IconButton
            icon="play-outline"
            class="h-8 w-8"
            :title="$t('repo.settings.crons.run')"
            @click="runCron(cron)"
          />
          <IconButton
            icon="edit"
            class="h-8 w-8"
            :title="$t('repo.settings.crons.edit')"
            @click="selectedCron = cron"
          />
          <IconButton
            icon="trash"
            class="hover:text-wp-error-100 h-8 w-8"
            :is-loading="isDeleting"
            :title="$t('repo.settings.crons.delete')"
            @click="deleteCron(cron)"
          />
        </div>
      </ListItem>

      <div v-if="loading" class="flex justify-center">
        <Icon name="spinner" class="animate-spin" />
      </div>
      <div v-else-if="crons?.length === 0" class="ml-2">{{ $t('repo.settings.crons.none') }}</div>
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

        <Checkbox v-model="selectedCronEnabled" :label="$t('repo.settings.crons.enabled')" />

        <InputField v-slot="{ id }" :label="$t('repo.settings.crons.branch.title')">
          <TextField
            :id="id"
            v-model="selectedCron.branch"
            :placeholder="$t('repo.settings.crons.branch.placeholder')"
          />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('repo.settings.crons.timezone')">
          <SelectField :id="id" v-model="selectedCronTimezone" :options="timezones" />
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

        <div v-if="isEditingCron && selectedCronEnabled" class="mb-4 ml-auto">
          <span v-if="selectedCron.next_exec && selectedCron.next_exec > 0" class="text-wp-text-100">
            {{
              $t('repo.settings.crons.next_exec_both', {
                local: date.toLocaleString(new Date(selectedCron.next_exec * 1000)),
                zoned: date.toLocaleString(new Date(selectedCron.next_exec * 1000), selectedCron.timezone),
                timezone: selectedCron.timezone,
              })
            }}
          </span>
          <span v-else class="text-wp-text-100">{{ $t('repo.settings.crons.not_executed_yet') }}</span>
        </div>

        <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.variables.title')">
          <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.manual_pipeline.variables.desc') }}</span>
          <KeyValueEditor
            :id="id"
            v-model="selectedCronVariables"
            :key-placeholder="$t('repo.manual_pipeline.variables.name')"
            :value-placeholder="$t('repo.manual_pipeline.variables.value')"
            :delete-title="$t('repo.manual_pipeline.variables.delete')"
            @update:is-valid="isVariablesValid = $event"
          />
        </InputField>

        <div class="flex gap-2">
          <Button type="button" color="gray" :text="$t('cancel')" @click="selectedCron = undefined" />
          <Button
            type="submit"
            color="green"
            :is-loading="isSaving"
            :text="isEditingCron ? $t('repo.settings.crons.save') : $t('repo.settings.crons.add')"
            :disabled="!isFormValid"
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
import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import InputField from '~/components/form/InputField.vue';
import KeyValueEditor from '~/components/form/KeyValueEditor.vue';
import SelectField from '~/components/form/SelectField.vue';
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

const timezones = Intl.supportedValuesOf('timeZone').map((tz) => ({
  value: tz,
  text: tz,
}));

const selectedCronTimezone = computed<string>({
  async set(tz) {
    selectedCron.value!.timezone = tz;
  },
  get() {
    return selectedCron.value!.timezone ?? 'UTC';
  },
});
const selectedCronVariables = computed<Record<string, string>>({
  async set(_vars) {
    selectedCron.value!.variables = _vars;
  },
  get() {
    return selectedCron.value!.variables ?? {};
  },
});

const selectedCronEnabled = computed<boolean>({
  async set(_enabled) {
    selectedCron.value!.enabled = _enabled;
  },
  get() {
    return selectedCron.value!.enabled !== undefined ? selectedCron.value!.enabled : true;
  },
});

async function loadCrons(page: number): Promise<Cron[] | null> {
  return apiClient.getCronList(repo.value.id, { page });
}

const isVariablesValid = ref(true);

const isFormValid = computed(() => {
  return isVariablesValid.value;
});

const { resetPage, data: crons, loading } = usePagination(loadCrons, () => !selectedCron.value);

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
