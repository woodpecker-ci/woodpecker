<template>
  <Scaffold v-if="org && orgPermissions && $route.meta.orgHeader">
    <template #title>
      {{ org.name }}
    </template>

    <template #titleActions>
      <IconButton
        v-if="!org.is_user && orgPermissions.admin"
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
import { computed, onMounted, ref, watch } from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import useApiClient from '~/compositions/useApiClient';
import { provide } from '~/compositions/useInjectProvide';
import { Org, OrgPermissions } from '~/lib/api/types';

const props = defineProps<{
  orgId: string;
}>();

const orgId = computed(() => parseInt(props.orgId, 10));
const apiClient = useApiClient();

const org = ref<Org>();
const orgPermissions = ref<OrgPermissions>();
provide('org', org);
provide('org-permissions', orgPermissions);

async function load() {
  org.value = await apiClient.getOrg(orgId.value);
  if (org.value.is_user) {
    orgPermissions.value = {
      member: true,
      admin: true,
    };
  } else {
    orgPermissions.value = await apiClient.getOrgPermissions(org.value.id);
  }
}

onMounted(load);
watch(orgId, load);
</script>
