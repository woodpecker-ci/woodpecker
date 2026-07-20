<template>
  <Settings :title="$t('repo.settings.parameters.parameters')" :description="$t('repo.settings.parameters.desc')">
    <template #headerActions>
      <Button
        v-if="selectedParameter"
        start-icon="back"
        :text="$t('repo.settings.parameters.show')"
        @click="selectedParameter = undefined"
      />
      <Button
        v-else
        start-icon="plus"
        :text="$t('repo.settings.parameters.add')"
        @click="selectedParameter = { type: 'string' }"
      />
    </template>

    <div v-if="!selectedParameter" class="text-wp-text-100 space-y-4">
      <ListItem
        v-for="parameter in parameters"
        :key="parameter.id"
        class="bg-wp-background-200! dark:bg-wp-background-200! items-center"
      >
        <span class="grid w-full grid-cols-3">
          <span>{{ parameter.name }}</span>
          <span class="md:display-unset hidden">{{ parameterTypeLabel(parameter.type) }}</span>
          <span class="md:display-unset hidden truncate">{{ parameter.description }}</span>
        </span>
        <div class="flex items-center gap-2">
          <IconButton
            icon="edit"
            class="h-8 w-8"
            :title="$t('repo.settings.parameters.edit')"
            @click="selectedParameter = parameter"
          />
          <IconButton
            icon="trash"
            class="hover:text-wp-error-100 h-8 w-8"
            :is-loading="isDeleting"
            :title="$t('repo.settings.parameters.delete')"
            @click="deleteParameter(parameter)"
          />
        </div>
      </ListItem>

      <div v-if="loading" class="flex justify-center">
        <Icon name="spinner" class="animate-spin" />
      </div>
      <div v-else-if="parameters?.length === 0" class="ml-2">{{ $t('repo.settings.parameters.none') }}</div>
    </div>

    <div v-else class="space-y-4">
      <form @submit.prevent="saveParameter">
        <InputField v-slot="{ id }" :label="$t('repo.settings.parameters.name.name')">
          <TextField
            :id="id"
            v-model="selectedParameter.name"
            :placeholder="$t('repo.settings.parameters.name.placeholder')"
            required
          />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('repo.settings.parameters.type.type')">
          <SelectField :id="id" v-model="selectedParameterType" :options="parameterTypeOptions" />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('repo.settings.parameters.description.description')">
          <TextField
            :id="id"
            v-model="selectedParameter.description"
            :placeholder="$t('repo.settings.parameters.description.placeholder')"
          />
        </InputField>

        <InputField
          v-if="selectedParameter.type === 'choice'"
          v-slot="{ id }"
          :label="$t('repo.settings.parameters.options.options')"
        >
          <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.settings.parameters.options.desc') }}</span>
          <TextField :id="id" v-model="selectedParameterOptionsText" :lines="4" required />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('repo.settings.parameters.default')">
          <SelectField
            v-if="selectedParameter.type === 'choice'"
            :id="id"
            v-model="selectedParameterDefault"
            :options="defaultChoiceOptions"
          />
          <Checkbox
            v-else-if="selectedParameter.type === 'boolean'"
            v-model="selectedParameterDefaultBoolean"
            :label="$t('repo.settings.parameters.default')"
          />
          <TextField
            v-else
            :id="id"
            v-model="selectedParameterDefault"
            :type="selectedParameter.type === 'number' ? 'number' : 'text'"
          />
        </InputField>

        <!-- a boolean always has a value, "required" would be meaningless -->
        <Checkbox
          v-if="selectedParameter.type !== 'boolean'"
          v-model="selectedParameterRequired"
          :label="$t('repo.settings.parameters.required')"
        />

        <InputField v-slot="{ id }" :label="$t('repo.settings.parameters.order.order')">
          <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.settings.parameters.order.desc') }}</span>
          <NumberField :id="id" v-model="selectedParameterOrder" />
        </InputField>

        <div class="flex gap-2">
          <Button type="button" color="gray" :text="$t('cancel')" @click="selectedParameter = undefined" />
          <Button
            type="submit"
            color="green"
            :is-loading="isSaving"
            :text="isEditingParameter ? $t('repo.settings.parameters.save') : $t('repo.settings.parameters.add')"
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
import NumberField from '~/components/form/NumberField.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { requiredInject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Parameter, ParameterType } from '~/lib/api/types';

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const repo = requiredInject('repo');
const selectedParameter = ref<Partial<Parameter>>();
const isEditingParameter = computed(() => !!selectedParameter.value?.id);

const parameterTypeLabels: Record<ParameterType, string> = {
  string: i18n.t('repo.settings.parameters.type.string'),
  number: i18n.t('repo.settings.parameters.type.number'),
  boolean: i18n.t('repo.settings.parameters.type.boolean'),
  choice: i18n.t('repo.settings.parameters.type.choice'),
};

function parameterTypeLabel(type: ParameterType): string {
  return parameterTypeLabels[type];
}

const parameterTypeOptions = Object.entries(parameterTypeLabels).map(([type, label]) => ({
  value: type,
  text: label,
}));

const selectedParameterType = computed<string>({
  set(type) {
    selectedParameter.value!.type = type as ParameterType;
    // a default rarely stays valid across type changes (e.g. free text -> choice)
    selectedParameter.value!.default = '';
    if (type === 'boolean') {
      selectedParameter.value!.required = false;
    }
  },
  get() {
    return selectedParameter.value!.type ?? 'string';
  },
});

const selectedParameterOptions = computed(() =>
  (selectedParameter.value?.options ?? []).map((option) => ({ value: option, text: option })),
);

// leading empty entry so a previously picked default can be unset again
const defaultChoiceOptions = computed(() => [
  { value: '', text: i18n.t('repo.settings.parameters.no_default') },
  ...selectedParameterOptions.value,
]);

const selectedParameterOptionsText = computed<string>({
  set(text) {
    selectedParameter.value!.options = text
      .split('\n')
      .map((option) => option.trim())
      .filter((option) => option !== '');
  },
  get() {
    return (selectedParameter.value!.options ?? []).join('\n');
  },
});

const selectedParameterDefault = computed<string>({
  set(value) {
    // Vue's v-model on <input type="number"> emits numbers, but Default is a string
    selectedParameter.value!.default = String(value);
  },
  get() {
    return selectedParameter.value!.default ?? '';
  },
});

const selectedParameterDefaultBoolean = computed<boolean>({
  set(value) {
    selectedParameter.value!.default = value ? 'true' : 'false';
  },
  get() {
    return selectedParameter.value!.default === 'true';
  },
});

const selectedParameterRequired = computed<boolean>({
  set(required) {
    selectedParameter.value!.required = required;
  },
  get() {
    return selectedParameter.value!.required ?? false;
  },
});

const selectedParameterOrder = computed<number>({
  set(order) {
    selectedParameter.value!.order = Number.isNaN(order) ? 0 : order;
  },
  get() {
    return selectedParameter.value!.order ?? 0;
  },
});

async function loadParameters(page: number): Promise<Parameter[] | null> {
  return apiClient.getParameterList(repo.value.id, { page });
}

const { resetPage, data: parameters, loading } = usePagination(loadParameters, () => !selectedParameter.value);

const { doSubmit: saveParameter, isLoading: isSaving } = useAsyncAction(async () => {
  if (!selectedParameter.value) {
    throw new Error("Unexpected: Can't get parameter");
  }

  if (isEditingParameter.value) {
    await apiClient.updateParameter(repo.value.id, selectedParameter.value);
  } else {
    await apiClient.createParameter(repo.value.id, selectedParameter.value);
  }
  notifications.notify({
    title: isEditingParameter.value
      ? i18n.t('repo.settings.parameters.saved')
      : i18n.t('repo.settings.parameters.created'),
    type: 'success',
  });
  selectedParameter.value = undefined;
  await resetPage();
});

const { doSubmit: deleteParameter, isLoading: isDeleting } = useAsyncAction(async (_parameter: Parameter) => {
  await apiClient.deleteParameter(repo.value.id, _parameter.id);
  notifications.notify({ title: i18n.t('repo.settings.parameters.deleted'), type: 'success' });
  await resetPage();
});

useWPTitle(computed(() => [i18n.t('repo.settings.parameters.parameters'), repo.value.full_name]));
</script>
