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
import { computed, onMounted, provide, ref, toRef, watch } from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import useApiClient from '~/compositions/useApiClient';
import { Org, OrgPermissions } from '~/lib/api/types';

const props = defineProps<{
  orgName: string;
}>();

const orgName = toRef(props, 'orgName');
const apiClient = useApiClient();

const org = computed<Org>(() => ({ name: orgName.value }));
const orgPermissions = ref<OrgPermissions>();
provide('org', org);
provide('org-permissions', orgPermissions);

async function load() {
  orgPermissions.value = await apiClient.getOrgPermissions(orgName.value);
}

onMounted(() => {
  load();
});

watch([orgName], () => {
  load();
});
</script>
