<template>
  <div v-if="innerValue" class="space-y-4">
    <form @submit.prevent="save">
      <InputField v-slot="{ id }" :label="$t('secrets.name')">
        <TextField
          :id="id"
          v-model="innerValue.name"
          :placeholder="$t('secrets.name')"
          required
          :disabled="isEditingSecret"
        />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('secrets.value')">
        <TextField
          :id="id"
          v-model="innerValue.value"
          :placeholder="$t('secrets.value')"
          :lines="5"
          :required="!isEditingSecret"
        />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('secrets.plugins.images')">
        <span class="text-wp-text-alt-100 mb-2 ml-1">{{ $t('secrets.plugins.desc') }}</span>

        <ListEditor
          :id="id"
          ref="imagesEditor"
          v-model="innerValue.images"
          :placeholder="$t('repo.settings.general.netrc_only_trusted.placeholder')"
        />
      </InputField>

      <InputField :label="$t('secrets.events.events')">
        <Warning class="mb-4 text-sm" :text="$t('secrets.events.warning')" />
        <CheckboxesField v-model="innerValue.events" :options="secretEventsOptions" />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('secrets.note')">
        <TextField :id="id" v-model="innerValue.note" :placeholder="$t('secrets.note')" :lines="3" />
      </InputField>

      <div class="flex gap-2">
        <Button type="button" color="gray" :text="$t('cancel')" @click="$emit('cancel')" />
        <Button
          type="submit"
          color="green"
          :is-loading="isSaving"
          :text="isEditingSecret ? $t('secrets.save') : $t('secrets.add')"
        />
      </div>
    </form>
  </div>
</template>

<script lang="ts" setup>
import { computed, toRef, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Warning from '~/components/atomic/Warning.vue';
import CheckboxesField from '~/components/form/CheckboxesField.vue';
import type { CheckboxOption } from '~/components/form/form.types';
import InputField from '~/components/form/InputField.vue';
import ListEditor from '~/components/form/ListEditor.vue';
import TextField from '~/components/form/TextField.vue';
import { WebhookEvents } from '~/lib/api/types';
import type { Secret } from '~/lib/api/types';

const props = defineProps<{
  modelValue: Partial<Secret>;
  isSaving: boolean;
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

const imagesEditor = useTemplateRef<InstanceType<typeof ListEditor>>('imagesEditor');

const secretEventsOptions: CheckboxOption[] = [
  { value: WebhookEvents.Push, text: i18n.t('repo.pipeline.event.push') },
  { value: WebhookEvents.Tag, text: i18n.t('repo.pipeline.event.tag') },
  { value: WebhookEvents.Release, text: i18n.t('repo.pipeline.event.release') },
  { value: WebhookEvents.PullRequest, text: i18n.t('repo.pipeline.event.pr') },
  { value: WebhookEvents.Deploy, text: i18n.t('repo.pipeline.event.deploy') },
  { value: WebhookEvents.Cron, text: i18n.t('repo.pipeline.event.cron') },
  { value: WebhookEvents.Manual, text: i18n.t('repo.pipeline.event.manual') },
];

function save() {
  if (!innerValue.value) {
    return;
  }

  // an image the user typed without confirming it should still be saved
  imagesEditor.value?.commitPendingItem();

  emit('save', innerValue.value);
}
</script>
