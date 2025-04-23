<template>
  <Scaffold enable-tabs :go-back="goBack">
    <template #title>
      <span>
        <router-link :to="{ name: 'org', params: { orgId: repo.org_id } }" class="hover:underline">{{
          repo!.owner
          /* eslint-disable-next-line @intlify/vue-i18n/no-raw-text */
        }}</router-link>
        /
        <router-link :to="{ name: 'repo' }" class="hover:underline">{{
          repo!.name
          /* eslint-disable-next-line @intlify/vue-i18n/no-raw-text */
        }}</router-link>
        /
        {{ $t('settings') }}
      </span>
    </template>

    <Tab icon="settings-outline" :to="{ name: 'repo-settings' }" :title="$t('repo.settings.general.general')" />
    <Tab icon="secret" :to="{ name: 'repo-settings-secrets' }" :title="$t('secrets.secrets')" />
    <Tab icon="docker" :to="{ name: 'repo-settings-registries' }" :title="$t('registries.registries')" />
    <Tab icon="cron" :to="{ name: 'repo-settings-crons' }" :title="$t('repo.settings.crons.crons')" />
    <Tab icon="tag" :to="{ name: 'repo-settings-badge' }" :title="$t('repo.settings.badge.badge')" />
    <Tab icon="toolbox" :to="{ name: 'repo-settings-actions' }" :title="$t('repo.settings.actions.actions')" />

    <router-view />
  </Scaffold>
</template>

<script lang="ts" setup>
import { onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import { requiredInject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import { useRouteBack } from '~/compositions/useRouteBack';

const notifications = useNotifications();
const router = useRouter();
const i18n = useI18n();

const repoPermissions = requiredInject('repo-permissions');
const repo = requiredInject('repo');

onMounted(async () => {
  if (!repoPermissions.value.admin) {
    notifications.notify({ type: 'error', title: i18n.t('repo.settings.not_allowed') });
    await router.replace({ name: 'home' });
  }
});

const goBack = useRouteBack({ name: 'repo' });
</script>
