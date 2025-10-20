<template>
  <Scaffold enable-tabs>
    <template #title>
      {{ $t('admin.settings.settings') }}
    </template>
    <Tab icon="info" :to="{ name: 'admin-settings' }" :title="$t('info')" />
    <Tab icon="secret" :to="{ name: 'admin-settings-secrets' }" :title="$t('secrets.secrets')" />
    <Tab icon="docker" :to="{ name: 'admin-settings-registries' }" :title="$t('registries.registries')" />
    <Tab icon="repo" :to="{ name: 'admin-settings-repos' }" :title="$t('admin.settings.repos.repos')" />
    <Tab icon="user" :to="{ name: 'admin-settings-users' }" :title="$t('admin.settings.users.users')" />
    <Tab icon="org" :to="{ name: 'admin-settings-orgs' }" :title="$t('admin.settings.orgs.orgs')" />
    <Tab icon="agent" :to="{ name: 'admin-settings-agents' }" :title="$t('admin.settings.agents.agents')" />
    <Tab icon="tray-full" :to="{ name: 'admin-settings-queue' }" :title="$t('admin.settings.queue.queue')" />
    <Tab icon="forge" :to="{ name: 'admin-settings-forges' }" :title="$t('forges')" />

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
