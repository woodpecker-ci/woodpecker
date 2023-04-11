<template>
  <Scaffold v-if="org && orgPermissions && $route.meta.orgHeader">
    <template #title>
      {{ org.name }}
    </template>

    <template #titleActions>
      <IconButton
        v-if="orgPermissions.admin"
        :to="{ name: 'repo-settings' }"
        :title="$t('org.settings.settings')"
        icon="settings"
      />
    </template>

    <router-view />
  </Scaffold>
  <router-view v-else-if="org && orgPermissions" />
</template>

<script lang="ts" setup>
import { onMounted, provide, ref, toRef, watch } from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import useApiClient from '~/compositions/useApiClient';
import { Org, OrgPermissions } from '~/lib/api/types';

const props = defineProps<{
  ownerOrOrgId: string;
}>();

const ownerOrOrgId = toRef(props, 'ownerOrOrgId');
const apiClient = useApiClient();

const org = ref<Org>();
const orgPermissions = ref<OrgPermissions>();
provide('org', org);
provide('org-permissions', orgPermissions);

async function load() {
  if (ownerOrOrgId.value.match(/^[0-9]+$/)) {
    org.value = await apiClient.getOrgById(Number(ownerOrOrgId.value));
  } else {
    org.value = { name: ownerOrOrgId.value };
  }
  orgPermissions.value = await apiClient.getOrgPermissions(ownerOrOrgId.value);
}

onMounted(() => {
  load();
});

watch([ownerOrOrgId], () => {
  load();
});
</script>
