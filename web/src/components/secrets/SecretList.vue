<template>
  <div class="space-y-4 text-wp-text-100">
    <ListItem
      v-for="secret in secrets"
      :key="secret.id"
      class="items-center !bg-wp-background-200 !dark:bg-wp-background-100"
    >
      <span>{{ secret.name }}</span>
      <div class="ml-auto space-x-2">
        <Badge v-for="event in secret.event" :key="event" :label="event" />
      </div>
      <IconButton
        icon="edit"
        class="ml-2 w-8 h-8"
        :title="$t('repo.settings.secrets.edit')"
        @click="editSecret(secret)"
      />
      <IconButton
        icon="trash"
        class="ml-2 w-8 h-8 hover:text-wp-control-error-100"
        :is-loading="isDeleting"
        :title="$t('repo.settings.secrets.delete')"
        @click="deleteSecret(secret)"
      />
    </ListItem>

    <div v-if="secrets?.length === 0" class="ml-2">{{ $t(i18nPrefix + 'none') }}</div>
  </div>
</template>

<script lang="ts" setup>
import { toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import { Secret } from '~/lib/api/types';

const props = defineProps<{
  modelValue: Secret[];
  isDeleting: boolean;
  i18nPrefix: string;
}>();

const emit = defineEmits<{
  (event: 'edit', secret: Secret): void;
  (event: 'delete', secret: Secret): void;
}>();

const i18n = useI18n();

const secrets = toRef(props, 'modelValue');

function editSecret(secret: Secret) {
  emit('edit', secret);
}

function deleteSecret(secret: Secret) {
  // TODO use proper dialog
  // eslint-disable-next-line no-alert, no-restricted-globals
  if (!confirm(i18n.t('repo.settings.secrets.delete_confirm'))) {
    return;
  }
  emit('delete', secret);
}
</script>
