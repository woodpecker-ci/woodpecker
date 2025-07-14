<template>
  <div class="text-wp-text-100 space-y-4">
    <ListItem
      v-for="secret in secrets"
      :key="secret.id"
      class="bg-wp-background-200! dark:bg-wp-background-100! items-center"
    >
      <span>{{ secret.name }}</span>
      <Badge
        v-if="secret.edit === false"
        class="ml-2"
        :value="secret.org_id === 0 ? $t('global_level_secret') : $t('org_level_secret')"
      />
      <div class="md:display-unset ml-auto hidden space-x-2">
        <Badge v-for="event in secret.events" :key="event" :value="event" />
      </div>
      <template v-if="secret.edit !== false">
        <IconButton
          icon="edit"
          class="ml-auto h-8 w-8 md:ml-2"
          :title="$t('secrets.edit')"
          @click="editSecret(secret)"
        />
        <IconButton
          icon="trash"
          class="hover:text-wp-error-100 ml-2 h-8 w-8"
          :is-loading="isDeleting"
          :title="$t('secrets.delete')"
          @click="deleteSecret(secret)"
        />
      </template>
    </ListItem>

    <div v-if="loading" class="flex justify-center">
      <Icon name="spinner" class="animate-spin" />
    </div>
    <div v-else-if="secrets?.length === 0" class="ml-2">{{ $t('secrets.none') }}</div>
  </div>
</template>

<script lang="ts" setup>
import { toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import type { Secret } from '~/lib/api/types';
import Icon from '../atomic/Icon.vue';

const props = defineProps<{
  modelValue: (Secret & { edit?: boolean })[];
  isDeleting: boolean;
  loading: boolean;
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
  // TODO: use proper dialog
  // eslint-disable-next-line no-alert
  if (!confirm(i18n.t('secrets.delete_confirm'))) {
    return;
  }
  emit('delete', secret);
}
</script>
