<template>
  <div class="mt-4">
    <Checkbox
      :model-value="org.require_org_agents"
      :label
      :description
      :disabled="isSaving"
      @update:model-value="saveRequireOrgAgents"
    />
  </div>
</template>

<script lang="ts" setup>
import Checkbox from '~/components/form/Checkbox.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import type { Org } from '~/lib/api/types';

const props = defineProps<{
  label: string;
  description: string;
  savedMessage: string;
}>();

const org = defineModel<Org>({ required: true });

const apiClient = useApiClient();
const notifications = useNotifications();

const { doSubmit: saveRequireOrgAgents, isLoading: isSaving } = useAsyncAction(async (requireOrgAgents: boolean) => {
  org.value = await apiClient.updateOrg(org.value.id, { require_org_agents: requireOrgAgents });
  notifications.notify({ title: props.savedMessage, type: 'success' });
});
</script>
