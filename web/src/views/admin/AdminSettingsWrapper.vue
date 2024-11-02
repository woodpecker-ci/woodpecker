<template>
  <Scaffold v-model:active-tab="activeTab" enable-tabs disable-tab-url-hash-mode>
    <template #title>
      {{ $t('settings') }}
    </template>
    <Tab id="info" :title="$t('info')" />
    <Tab id="secrets" :title="$t('secrets.secrets')" />
    <Tab id="registries" :title="$t('registries.registries')" />
    <Tab id="repos" :title="$t('admin.settings.repos.repos')" />
    <Tab id="users" :title="$t('admin.settings.users.users')" />
    <Tab id="orgs" :title="$t('admin.settings.orgs.orgs')" />
    <Tab id="agents" :title="$t('admin.settings.agents.agents')" />
    <Tab id="queue" :title="$t('admin.settings.queue.queue')" />

    <router-view />
  </Scaffold>
</template>

<script lang="ts" setup>
import { computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import useAuthentication from '~/compositions/useAuthentication';
import useNotifications from '~/compositions/useNotifications';

const notifications = useNotifications();
const router = useRouter();
const route = useRoute();
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

const activeTab = computed({
  get() {
    if (route.name === 'admin-settings-secrets') {
      return 'secrets';
    }
    if (route.name === 'admin-settings-registries') {
      return 'registries';
    }
    if (route.name === 'admin-settings-repos') {
      return 'repos';
    }
    if (route.name === 'admin-settings-users') {
      return 'users';
    }
    if (route.name === 'admin-settings-orgs') {
      return 'orgs';
    }
    if (route.name === 'admin-settings-agents') {
      return 'agents';
    }
    if (route.name === 'admin-settings-queue') {
      return 'queue';
    }
    return 'info';
  },
  set(tab: string) {
    if (tab === 'secrets') {
      router.push({ name: 'admin-settings-secrets' });
    } else if (tab === 'registries') {
      router.push({ name: 'admin-settings-registries' });
    } else if (tab === 'repos') {
      router.push({ name: 'admin-settings-repos' });
    } else if (tab === 'users') {
      router.push({ name: 'admin-settings-users' });
    } else if (tab === 'orgs') {
      router.push({ name: 'admin-settings-orgs' });
    } else if (tab === 'agents') {
      router.push({ name: 'admin-settings-agents' });
    } else if (tab === 'queue') {
      router.push({ name: 'admin-settings-queue' });
    } else {
      router.push({ name: 'admin-settings-info' });
    }
  },
});
</script>
