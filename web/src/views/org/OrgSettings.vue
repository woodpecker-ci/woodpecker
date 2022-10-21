<template>
  <Scaffold enable-tabs :go-back="goBack">
    <template #headerTitle>
      <h1 class="text-xl ml-2 text-color">{{ $t('org.settings.settings') }}</h1>
    </template>

    <Tab id="secrets" :title="$t('org.settings.secrets.secrets')">
      <OrgSecretsTab />
    </Tab>
  </Scaffold>
</template>

<script lang="ts">
import { defineComponent, inject, onMounted, Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Tab from '~/components/layout/scaffold/Tab.vue';
import OrgSecretsTab from '~/components/org/settings/OrgSecretsTab.vue';
import useNotifications from '~/compositions/useNotifications';
import { useRouteBackOrDefault } from '~/compositions/useRouteBackOrDefault';
import { OrgPermissions } from '~/lib/api/types';

export default defineComponent({
  name: 'OrgSettings',

  components: {
    Tab,
    OrgSecretsTab,
  },

  setup() {
    const notifications = useNotifications();
    const router = useRouter();
    const i18n = useI18n();

    const orgPermissions = inject<Ref<OrgPermissions>>('org-permissions');
    if (!orgPermissions) {
      throw new Error('Unexpected: "orgPermissions" should be provided at this place');
    }

    onMounted(async () => {
      if (!orgPermissions.value.admin) {
        notifications.notify({ type: 'error', title: i18n.t('org.settings.not_allowed') });
        await router.replace({ name: 'home' });
      }
    });

    return {
      goBack: useRouteBackOrDefault({ name: 'repos-owner' }),
    };
  },
});
</script>
