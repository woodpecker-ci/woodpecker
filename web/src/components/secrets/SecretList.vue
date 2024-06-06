<template>
  <div class="space-y-4 text-wp-text-100">
    <ListItem
      v-for="secret in secrets"
      :key="secret.id"
      class="items-center !bg-wp-background-200 !dark:bg-wp-background-100"
    >
      <span>{{ secret.name }}</span>
      <Badge
        v-if="secret.edit === false"
        class="ml-2"
        :label="secret.org_id === 0 ? $t('global_level_secret') : $t('org_level_secret')"
      />
      <div class="ml-auto space-x-2 <md:hidden">
        <Badge v-for="event in secret.events" :key="event" :label="event" />
      </div>
      <template v-if="secret.edit !== false">
        <IconButton
          icon="edit"
          class="ml-2 <md:ml-auto w-8 h-8"
          :title="$t('secrets.edit')"
          @click="editSecret(secret)"
        />
        <IconButton
          icon="trash"
          class="ml-2 w-8 h-8 hover:text-wp-control-error-100"
          :is-loading="isDeleting"
          :title="$t('secrets.delete')"
          @click="deleteSecret(secret)"
        />
      </template>
    </ListItem>

    <div v-if="secrets?.length === 0" class="ml-2">{{ $t('secrets.none') }}</div>
  </div>
</template>

<script lang="ts" setup>
import { toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import type { Secret } from '~/lib/api/types';

const props = defineProps<{
  modelValue: (Secret & { edit?: boolean })[];
  isDeleting: boolean;
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
