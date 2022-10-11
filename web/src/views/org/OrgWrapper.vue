<template>
  <FluidContainer v-if="org && orgPermissions && $route.meta.orgHeader">
    <div class="flex flex-wrap border-b items-center pb-4 mb-4 dark:border-gray-600 justify-center">
      <h1 class="text-xl text-color w-full md:w-auto text-center mb-4 md:mb-0">
        {{ org.name }}
      </h1>
      <IconButton
        v-if="orgPermissions.admin"
        class="ml-2"
        :to="{ name: 'repo-settings' }"
        :title="$t('org.settings.settings')"
        icon="settings"
      />
    </div>

    <router-view />
  </FluidContainer>
  <router-view v-else-if="org && orgPermissions" />
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, provide, ref, toRef, watch } from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import useApiClient from '~/compositions/useApiClient';
import { Org, OrgPermissions } from '~/lib/api/types';

export default defineComponent({
  name: 'OrgWrapper',

  components: { FluidContainer, IconButton },

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
