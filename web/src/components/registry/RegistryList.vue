<template>
  <div class="space-y-4 text-wp-text-100">
    <ListItem
      v-for="registry in registries"
      :key="registry.id"
      class="items-center !bg-wp-background-200 !dark:bg-wp-background-100"
    >
      <span>{{ registry.address }}</span>
      <IconButton
        :icon="registry.readonly ? 'chevron-right' : 'edit'"
        class="ml-auto w-8 h-8"
        :title="registry.readonly ? $t('registries.view') : $t('registries.edit')"
        @click="editRegistry(registry)"
      />
      <IconButton
        v-if="!registry.readonly"
        icon="trash"
        class="w-8 h-8 hover:text-wp-control-error-100"
        :is-loading="isDeleting"
        :title="$t('registries.delete')"
        @click="deleteRegistry(registry)"
      />
    </ListItem>

    <div v-if="registries?.length === 0" class="ml-2">{{ $t('registries.none') }}</div>
  </div>
</template>

<script lang="ts" setup>
import { toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import type { Registry } from '~/lib/api/types';

const props = defineProps<{
  modelValue: (Registry & { edit?: boolean })[];
  isDeleting: boolean;
}>();

const emit = defineEmits<{
  (event: 'edit', registry: Registry): void;
  (event: 'delete', registry: Registry): void;
}>();

const i18n = useI18n();

const registries = toRef(props, 'modelValue');

function editRegistry(registry: Registry) {
  emit('edit', registry);
}

function deleteRegistry(registry: Registry) {
  // TODO: use proper dialog
  // eslint-disable-next-line no-alert
  if (!confirm(i18n.t('registries.delete_confirm'))) {
    return;
  }
  emit('delete', registry);
}
</script>
