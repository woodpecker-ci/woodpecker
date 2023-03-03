<template>
  <div v-if="innerValue" class="space-y-4">
    <form @submit.prevent="save">
      <InputField :label="$t(i18nPrefix + 'name')">
        <TextField
          v-model="innerValue.name"
          :placeholder="$t(i18nPrefix + 'name')"
          required
          :disabled="isEditingSecret"
        />
      </InputField>

      <InputField :label="$t(i18nPrefix + 'value')">
        <TextField v-model="innerValue.value" :placeholder="$t(i18nPrefix + 'value')" :lines="5" />
      </InputField>

      <InputField :label="$t(i18nPrefix + 'images.images')">
        <TextField v-model="images" :placeholder="$t(i18nPrefix + 'images.desc')" />

        <Checkbox v-model="innerValue.plugins_only" class="mt-4" :label="$t(i18nPrefix + 'plugins_only')" />
      </InputField>

      <InputField :label="$t(i18nPrefix + 'events.events')">
        <CheckboxesField v-model="innerValue.event" :options="secretEventsOptions" />
      </InputField>

      <Button type="button" color="gray" :text="$t('cancel')" />
      <Button
        type="submit"
        color="green"
        :is-loading="isSaving"
        :text="isEditingSecret ? $t(i18nPrefix + 'save') : $t(i18nPrefix + 'add')"
      />
    </form>
  </div>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';
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
const images = computed<string>({
  get() {
    return innerValue.value?.image?.join(',') || '';
  },
  set(value) {
    if (innerValue.value) {
      innerValue.value.image = value
        .split(',')
        .map((s) => s.trim())
        .filter((s) => s !== '');
    }
  },
});
const isEditingSecret = computed(() => !!innerValue.value?.id);

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
  emit('save', innerValue.value);
}
</script>
