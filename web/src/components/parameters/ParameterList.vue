<template>
  <div class="flex flex-col gap-4">
    <ListItem v-for="parameter in parameters" :key="parameter.name" class="flex flex-col gap-2 !bg-wp-background-200 dark:!bg-wp-background-100">
      <div class="flex items-center justify-between gap-4">
        <div class="flex items-center gap-2">
          <h3 class="font-bold">{{ parameter.name }}</h3>
          <div class="flex items-center gap-1">
            <Icon name="push" class="h-4 w-4" />
            <span>{{ parameter.branch }}</span>
          </div>
        </div>
        <div class="flex gap-2">
          <div class="md:display-unset ml-auto hidden space-x-2">
            <Badge :label="$t(`parameters.types.${parameter.type}`)" />
          </div>
          <IconButton
            :title="$t('parameters.edit')"
            icon="edit"
            @click="$emit('edit', parameter)"
          />
          <IconButton
            :title="$t('parameters.delete')"
            icon="trash"
            @click="$emit('delete', parameter)"
          />
        </div>
      </div>
      <p v-if="parameter.description" class="text-sm text-wp-text-alt-100">
        {{ parameter.description }}
      </p>
    </ListItem>
  </div>
</template>

<script lang="ts" setup>
import Badge from "~/components/atomic/Badge.vue";
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
</script>
