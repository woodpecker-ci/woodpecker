<template>
  <Scaffold v-model:active-tab="activeTab" enable-tabs :go-back="goBack" disable-tab-url-hash-mode>
    <template #title>
      <span>
        <router-link :to="{ name: 'org', params: { orgId: repo!.org_id } }" class="hover:underline">{{
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

    <Tab id="general" :title="$t('repo.settings.general.general')" />
    <Tab id="secrets" :title="$t('secrets.secrets')" />
    <Tab id="registries" :title="$t('registries.registries')" />
    <Tab id="crons" :title="$t('repo.settings.crons.crons')" />
    <Tab id="badge" :title="$t('repo.settings.badge.badge')" />
    <Tab id="actions" :title="$t('repo.settings.actions.actions')" />

    <router-view />
  </Scaffold>
</template>

<script lang="ts" setup>
import { computed, inject, onMounted, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import useNotifications from '~/compositions/useNotifications';
import { useRouteBack } from '~/compositions/useRouteBack';
import type { Repo, RepoPermissions } from '~/lib/api/types';

const notifications = useNotifications();
const router = useRouter();
const route = useRoute();
const i18n = useI18n();

const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
if (!repoPermissions) {
  throw new Error('Unexpected: "repoPermissions" should be provided at this place');
}

const repo = inject<Ref<Repo>>('repo');
if (!repo) {
  throw new Error('Unexpected: "repo" should be provided at this place');
}

onMounted(async () => {
  if (!repoPermissions.value.admin) {
    notifications.notify({ type: 'error', title: i18n.t('repo.settings.not_allowed') });
    await router.replace({ name: 'home' });
  }
});

const goBack = useRouteBack({ name: 'repo' });

const activeTab = computed({
  get() {
    if (route.name === 'repo-settings-secrets') {
      return 'secrets';
    }
    if (route.name === 'repo-settings-registries') {
      return 'registries';
    }
    if (route.name === 'repo-settings-crons') {
      return 'crons';
    }
    if (route.name === 'repo-settings-badge') {
      return 'badge';
    }
    if (route.name === 'repo-settings-actions') {
      return 'actions';
    }
    return 'general';
  },
  set(tab: string) {
    if (tab === 'secrets') {
      router.push({ name: 'repo-settings-secrets' });
    } else if (tab === 'registries') {
      router.push({ name: 'repo-settings-registries' });
    } else if (tab === 'crons') {
      router.push({ name: 'repo-settings-crons' });
    } else if (tab === 'badge') {
      router.push({ name: 'repo-settings-badge' });
    } else if (tab === 'actions') {
      router.push({ name: 'repo-settings-actions' });
    } else {
      router.push({ name: 'repo-settings-general' });
    }
  },
});
</script>
