<template>
  <Scaffold v-if="org && orgPermissions && route.meta.orgHeader">
    <template #title>
      {{ org.name }}
    </template>

    <template #headerActions>
      <IconButton
        v-if="orgPermissions.admin"
        :to="{ name: org.is_user ? 'user' : 'org-settings-secrets' }"
        :title="$t('settings')"
        icon="settings"
      />
    </template>

    <router-view />
  </Scaffold>
  <router-view v-else-if="org && orgPermissions" />
</template>

<script lang="ts" setup>
import type { Ref } from 'vue';
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import IconButton from '~/components/atomic/IconButton.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import useApiClient from '~/compositions/useApiClient';
import { provide } from '~/compositions/useInjectProvide';
import type { Org, OrgPermissions } from '~/lib/api/types';

const props = defineProps<{
  orgId: string;
}>();

const orgId = computed(() => Number.parseInt(props.orgId, 10));
const apiClient = useApiClient();
const route = useRoute();

const org = ref<Org>();
const orgPermissions = ref<OrgPermissions>();
provide('org', org as Ref<Org>); // can't be undefined because of v-if in template
provide('org-permissions', orgPermissions as Ref<OrgPermissions>); // can't be undefined because of v-if in template

async function load() {
  org.value = await apiClient.getOrg(orgId.value);
  orgPermissions.value = await apiClient.getOrgPermissions(org.value.id);
}

onMounted(load);
watch(orgId, load);
</script>
