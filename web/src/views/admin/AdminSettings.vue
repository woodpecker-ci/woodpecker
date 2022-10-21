<template>
  <Scaffold enable-tabs :go-back="goBack">
    <template #headerTitle>
      {{ $t('repo.settings.settings') }}
    </template>
    <Tab id="secrets" :title="$t('admin.settings.secrets.secrets')">
      <AdminSecretsTab />
    </Tab>
  </Scaffold>
</template>

<script lang="ts">
import { defineComponent, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import AdminSecretsTab from '~/components/admin/settings/AdminSecretsTab.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import useAuthentication from '~/compositions/useAuthentication';
import useNotifications from '~/compositions/useNotifications';
import { useRouteBackOrDefault } from '~/compositions/useRouteBackOrDefault';

export default defineComponent({
  name: 'AdminSettings',

  components: {
    Tab,
    AdminSecretsTab,
    Scaffold,
  },

  setup() {
    const notifications = useNotifications();
    const router = useRouter();
    const i18n = useI18n();
    const { user } = useAuthentication();

    onMounted(async () => {
      if (!user?.admin) {
        notifications.notify({ type: 'error', title: i18n.t('admin.settings.not_allowed') });
        await router.replace({ name: 'home' });
      }
    });

    return {
      goBack: useRouteBackOrDefault({ name: 'home' }),
    };
  },
});
</script>
