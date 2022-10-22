<template>
  <Scaffold v-if="org && orgPermissions && $route.meta.orgHeader">
    <template #headerTitle>
      {{ org.name }}
    </template>

    <template #headerActions>
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

<script lang="ts">
import { computed, defineComponent, onMounted, provide, ref, toRef, watch } from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import useApiClient from '~/compositions/useApiClient';
import { Org, OrgPermissions } from '~/lib/api/types';

export default defineComponent({
  name: 'OrgWrapper',

  components: { IconButton, Scaffold },

  props: {
    repoOwner: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const repoOwner = toRef(props, 'repoOwner');
    const apiClient = useApiClient();
    const org = computed<Org>(() => ({ name: repoOwner.value }));

    const orgPermissions = ref<OrgPermissions>();
    provide('org', org);
    provide('org-permissions', orgPermissions);

    async function load() {
      orgPermissions.value = await apiClient.getOrgPermissions(repoOwner.value);
    }

    onMounted(() => {
      load();
    });

    watch([repoOwner], () => {
      load();
    });

    return { org, orgPermissions };
  },
});
</script>
