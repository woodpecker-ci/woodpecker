<template>
  <div v-if="innerValue" class="space-y-4">
    <form @submit.prevent="save">
      <InputField v-slot="{ id }" :label="$t(i18nPrefix + 'name')">
        <TextField
          :id="id"
          v-model="innerValue.name"
          :placeholder="$t(i18nPrefix + 'name')"
          required
          :disabled="isEditingSecret"
        />
      </InputField>

      <InputField v-slot="{ id }" :label="$t(i18nPrefix + 'value')">
        <TextField
          :id="id"
          v-model="innerValue.value"
          :placeholder="$t(i18nPrefix + 'value')"
          :lines="5"
          :required="!isEditingSecret"
        />
      </InputField>

      <InputField v-slot="{ id }" :label="$t(i18nPrefix + 'images.images')">
        <span class="ml-1 mb-2 text-wp-text-alt-100">{{ $t(i18nPrefix + 'images.desc') }}</span>

        <div class="flex flex-col gap-2">
          <div v-for="image in innerValue.images" :key="image" class="flex gap-2">
            <TextField :id="id" :model-value="image" disabled />
            <Button type="button" color="gray" start-icon="trash" @click="removeImage(image)" />
          </div>
          <div class="flex gap-2">
            <TextField :id="id" v-model="newImage" @keydown.enter.prevent="addNewImage" />
            <Button type="button" color="gray" start-icon="plus" @click="addNewImage" />
          </div>
        </div>
      </InputField>

      <InputField :label="$t(i18nPrefix + 'events.events')">
        <CheckboxesField v-model="innerValue.events" :options="secretEventsOptions" />
      </InputField>

      <div class="flex gap-2">
        <Button type="button" color="gray" :text="$t('cancel')" @click="$emit('cancel')" />
        <Button
          type="submit"
          color="green"
          :is-loading="isSaving"
          :text="isEditingSecret ? $t(i18nPrefix + 'save') : $t(i18nPrefix + 'add')"
        />
      </div>
    </form>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref, toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import CheckboxesField from '~/components/form/CheckboxesField.vue';
import { CheckboxOption } from '~/components/form/form.types';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import { Secret, WebhookEvents } from '~/lib/api/types';

const props = defineProps<{
  modelValue: Partial<Secret>;
  isSaving: boolean;
  i18nPrefix: string;
}>();

const emit = defineEmits<{
  (event: 'update:modelValue', value: Partial<Secret> | undefined): void;
  (event: 'save', value: Partial<Secret>): void;
  (event: 'cancel'): void;
}>();

const i18n = useI18n();

const modelValue = toRef(props, 'modelValue');
const innerValue = computed({
  get: () => modelValue.value,
  set: (value) => {
    emit('update:modelValue', value);
  },
});
const isEditingSecret = computed(() => !!innerValue.value?.id);

const newImage = ref('');
function addNewImage() {
  if (!newImage.value) {
    return;
  }
  innerValue.value.images?.push(newImage.value);
  newImage.value = '';
}
function removeImage(image: string) {
  innerValue.value.images = innerValue.value.images?.filter((i) => i !== image);
}

const secretEventsOptions: CheckboxOption[] = [
  { value: WebhookEvents.Push, text: i18n.t('repo.pipeline.event.push') },
  { value: WebhookEvents.Tag, text: i18n.t('repo.pipeline.event.tag') },
  {
    value: WebhookEvents.PullRequest,
    text: i18n.t('repo.pipeline.event.pr'),
    description: i18n.t('repo.settings.secrets.events.pr_warning'),
  },
  { value: WebhookEvents.Deploy, text: i18n.t('repo.pipeline.event.deploy') },
  { value: WebhookEvents.Cron, text: i18n.t('repo.pipeline.event.cron') },
  { value: WebhookEvents.Manual, text: i18n.t('repo.pipeline.event.manual') },
];

function save() {
  if (!innerValue.value) {
    return;
  }

  if (newImage.value) {
    innerValue.value.images?.push(newImage.value);
  }

  emit('save', innerValue.value);
}
</script>
