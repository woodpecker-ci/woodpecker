<template>
  <div class="space-y-4 text-wp-text-100">
    <ListItem
      v-for="variable in variables"
      :key="variable.id"
      class="items-center !bg-wp-background-200 !dark:bg-wp-background-100"
    >
      <span>{{ variable.name }}</span>
      <Badge
        v-if="variable.edit === false"
        class="ml-2"
        :label="variable.org_id === 0 ? $t('global_level_variable') : $t('org_level_variable')"
      />
      <template v-if="variable.edit !== false">
        <IconButton icon="edit" class="ml-auto w-8 h-8" :title="$t('variables.edit')" @click="editVariable(variable)" />
        <IconButton
          icon="trash"
          class="ml-2 w-8 h-8 hover:text-wp-control-error-100"
          :is-loading="isDeleting"
          :title="$t('variables.delete')"
          @click="deleteVariable(variable)"
        />
      </template>
    </ListItem>

    <div v-if="variables?.length === 0" class="ml-2">{{ $t('variables.none') }}</div>
  </div>
</template>

<script lang="ts" setup>
import { toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import type { Variable } from '~/lib/api/types';

const props = defineProps<{
  modelValue: (Variable & { edit?: boolean })[];
  isDeleting: boolean;
}>();

const emit = defineEmits<{
  (event: 'edit', variable: Variable): void;
  (event: 'delete', variable: Variable): void;
}>();

const i18n = useI18n();

const variables = toRef(props, 'modelValue');

function editVariable(variable: Variable) {
  emit('edit', variable);
}

function deleteVariable(variable: Variable) {
  // TODO: use proper dialog
  // eslint-disable-next-line no-alert
  if (!confirm(i18n.t('variables.delete_confirm'))) {
    return;
  }
  emit('delete', variable);
}
</script>
