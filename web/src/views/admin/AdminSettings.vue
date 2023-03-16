<template>
  <Scaffold enable-tabs>
    <template #title>
      {{ $t('repo.settings.settings') }}
    </template>
    <Tab id="secrets" :title="$t('admin.settings.secrets.secrets')">
      <AdminSecretsTab />
    </Tab>
    <Tab id="agents" :title="$t('admin.settings.agents.agents')">
      <AdminAgentsTab />
    </Tab>
    <Tab id="queue" :title="$t('admin.settings.queue.queue')">
      <AdminQueueTab />
    </Tab>
  </Scaffold>
</template>

<script lang="ts" setup>
import { onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import AdminAgentsTab from '~/components/admin/settings/AdminAgentsTab.vue';
import AdminQueueTab from '~/components/admin/settings/AdminQueueTab.vue';
import AdminSecretsTab from '~/components/admin/settings/AdminSecretsTab.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import useAuthentication from '~/compositions/useAuthentication';
import useNotifications from '~/compositions/useNotifications';

const notifications = useNotifications();
const router = useRouter();
const i18n = useI18n();
const { user } = useAuthentication();

onMounted(async () => {
  if (!user?.admin) {
    notifications.notify({ type: 'error', title: i18n.t('admin.settings.not_allowed') });
    await router.replace({ name: 'home' });
  }

  if (!user?.admin) {
    notifications.notify({ type: 'error', title: i18n.t('admin.settings.not_allowed') });
    await router.replace({ name: 'home' });
  }
});
</script>
