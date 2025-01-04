<template>
  <div v-if="!props.loading" class="text-wp-text-100 space-y-4">
    <ListItem
      v-for="agent in props.agents"
      :key="agent.id"
      class="items-center !bg-wp-background-200 dark:!bg-wp-background-100 admin-agents"
    >
      <span>{{ agent.name || `Agent ${agent.id}` }}</span>
      <span class="ml-auto">
        <span class="hidden space-x-2 md:inline-block">
          <Badge
            v-if="props.isAdmin === true && agent.org_id !== -1"
            :label="$t('admin.settings.agents.org.badge')"
            :value="agent.org_id"
          />
          <Badge v-if="agent.platform" :label="$t('admin.settings.agents.platform.badge')" :value="agent.platform" />
          <Badge v-if="agent.backend" :label="$t('admin.settings.agents.backend.badge')" :value="agent.backend" />
          <Badge v-if="agent.capacity" :label="$t('admin.settings.agents.capacity.badge')" :value="agent.capacity" />
        </span>
        <span class="ml-2">{{
          agent.last_contact ? date.timeAgo(agent.last_contact * 1000) : $t('admin.settings.agents.never')
        }}</span>
      </span>
      <IconButton
        icon="edit"
        :title="$t('admin.settings.agents.edit_agent')"
        class="ml-2 h-8 w-8"
        @click="$emit('edit', agent)"
      />
      <IconButton
        icon="trash"
        :title="$t('admin.settings.agents.delete_agent')"
        class="hover:text-wp-error-100 ml-2 h-8 w-8"
        :is-loading="props.isDeleting"
        @click="$emit('delete', agent)"
      />
    </ListItem>

    <div v-if="props.agents?.length === 0" class="ml-2">{{ $t('admin.settings.agents.none') }}</div>
  </div>
  <div v-else class="flex justify-center">
    <Icon name="loading" class="animate-spin" />
  </div>
</template>

<script lang="ts" setup>
import Badge from '~/components/atomic/Badge.vue';
import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import { useDate } from '~/compositions/useDate';
import type { Agent } from '~/lib/api/types';

const props = defineProps<{
  agents: Agent[];
  isDeleting: boolean;
  loading: boolean;
  isAdmin?: boolean;
}>();

defineEmits<{
  (e: 'edit', agent: Agent): void;
  (e: 'delete', agent: Agent): void;
}>();

const date = useDate();
</script>

<style scoped>
.admin-agents {
  @apply p-4;
}
</style>
