<template>
  <Scaffold enable-tabs>
    <template #title>
      {{ $t('settings') }}
    </template>
    <Tab to="admin-settings" :title="$t('info')" />
    <Tab to="admin-settings-secrets" :title="$t('secrets.secrets')" />
    <Tab to="admin-settings-registries" :title="$t('registries.registries')" />
    <Tab to="admin-settings-repos" :title="$t('admin.settings.repos.repos')" />
    <Tab to="admin-settings-users" :title="$t('admin.settings.users.users')" />
    <Tab to="admin-settings-orgs" :title="$t('admin.settings.orgs.orgs')" />
    <Tab to="admin-settings-agents" :title="$t('admin.settings.agents.agents')" />
    <Tab to="admin-settings-queue" :title="$t('admin.settings.queue.queue')" />

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
});
</script>
