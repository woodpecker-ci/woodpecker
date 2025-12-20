<template>
  <div class="flex flex-col gap-4">
    <ListItem
      v-for="parameter in parameters"
      :key="parameter.name"
      class="bg-wp-background-200! dark:bg-wp-background-100! flex flex-col gap-2"
    >
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span>{{ parameter.name }}</span>
          <div class="flex items-center gap-1">
            <Icon name="branch" class="h-4 w-4" />
            <span>{{ parameter.branch }}</span>
          </div>
        </div>
        <div class="md:display-unset ml-auto hidden">
          <Badge :value="ParameterTypeName(parameter.type)" />
        </div>
        <div class="flex items-center gap-2">
          <IconButton
            :title="$t('parameters.edit')"
            icon="edit"
            class="h-8 w-8 md:ml-2"
            @click="$emit('edit', parameter)"
          />
          <IconButton
            :title="$t('parameters.delete')"
            icon="trash"
            class="hover:text-wp-error-100 h-8 w-8"
            @click="$emit('delete', parameter)"
          />
        </div>
      </div>
      <div v-if="parameter.description" class="text-wp-text-alt-100 text-sm">
        {{ parameter.description }}
      </div>
    </ListItem>
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import type { Parameter } from '~/lib/api/types/parameter';

defineProps<{
  parameters: Parameter[];
}>();

defineEmits<{
  (e: 'edit', parameter: Parameter): void;
  (e: 'delete', parameter: Parameter): void;
}>();

const { t } = useI18n();

const ParameterTypeName = (type: string): string => {
  switch (type) {
    case 'boolean':
      return t('parameters.types.boolean');
    case 'single_choice':
      return t('parameters.types.single_choice');
    case 'multiple_choice':
      return t('parameters.types.multiple_choice');
    case 'string':
      return t('parameters.types.string');
    case 'text':
      return t('parameters.types.text');
    case 'password':
      return t('parameters.types.password');
    default:
      return type; // Fallback to the type itself if no translation is found
  }
};
</script>
