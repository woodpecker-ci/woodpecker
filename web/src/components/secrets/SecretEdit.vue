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
        <TextField v-model="innerValue.value" :placeholder="$t(i18nPrefix + 'value')" :lines="5" required />
      </InputField>

      <InputField :label="$t(i18nPrefix + 'images.images')">
        <TextField v-model="images" :placeholder="$t(i18nPrefix + 'images.desc')" />
      </InputField>

      <InputField :label="$t(i18nPrefix + 'events.events')">
        <CheckboxesField v-model="innerValue.event" :options="secretEventsOptions" />
      </InputField>

      <Button
        :is-loading="isSaving"
        type="submit"
        :text="isEditingSecret ? $t(i18nPrefix + 'save') : $t(i18nPrefix + 'add')"
      />
    </form>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, PropType, toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import CheckboxesField from '~/components/form/CheckboxesField.vue';
import { CheckboxOption } from '~/components/form/form.types';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import { Secret, WebhookEvents } from '~/lib/api/types';

export default defineComponent({
  name: 'SecretEdit',

  components: {
    Button,
    InputField,
    TextField,
    CheckboxesField,
  },

  props: {
    // used by toRef
    // eslint-disable-next-line vue/no-unused-properties
    modelValue: {
      type: Object as PropType<Partial<Secret>>,
      default: undefined,
    },

    isSaving: {
      type: Boolean,
    },

    i18nPrefix: {
      type: String,
      required: true,
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    'update:modelValue': (_value: Partial<Secret> | undefined): boolean => true,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    save: (_value: Partial<Secret>): boolean => true,
  },

  setup: (props, ctx) => {
    const i18n = useI18n();

    const modelValue = toRef(props, 'modelValue');
    const innerValue = computed({
      get: () => modelValue.value,
      set: (value) => {
        ctx.emit('update:modelValue', value);
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
      { value: WebhookEvents.Push, text: i18n.t('repo.build.event.push') },
      { value: WebhookEvents.Tag, text: i18n.t('repo.build.event.tag') },
      {
        value: WebhookEvents.PullRequest,
        text: i18n.t('repo.build.event.pr'),
        description: i18n.t('repo.settings.secrets.events.pr_warning'),
      },
      { value: WebhookEvents.Deploy, text: i18n.t('repo.build.event.deploy') },
    ];

    function save() {
      if (!innerValue.value) {
        return;
      }
      ctx.emit('save', innerValue.value);
    }

    return {
      innerValue,
      isEditingSecret,
      secretEventsOptions,
      images,
      save,
    };
  },
});
</script>
