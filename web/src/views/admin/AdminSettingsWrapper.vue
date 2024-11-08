<template>
  <Scaffold enable-tabs>
    <template #title>
      {{ $t('settings') }}
    </template>
    <Tab id="admin-settings-info" :title="$t('info')" />
    <Tab id="admin-settings-secrets" :title="$t('secrets.secrets')" />
    <Tab id="admin-settings-registries" :title="$t('registries.registries')" />
    <Tab id="admin-settings-repos" :title="$t('admin.settings.repos.repos')" />
    <Tab id="admin-settings-users" :title="$t('admin.settings.users.users')" />
    <Tab id="admin-settings-orgs" :title="$t('admin.settings.orgs.orgs')" />
    <Tab id="admin-settings-agents" :title="$t('admin.settings.agents.agents')" />
    <Tab id="admin-settings-queue" :title="$t('admin.settings.queue.queue')" />

    <router-view />
  </Scaffold>
</template>

<script lang="ts" setup>
import { onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

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
