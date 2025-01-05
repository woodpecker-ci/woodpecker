<template>
  <Scaffold v-if="org" enable-tabs :go-back="goBack">
    <template #title>
      <span>
        <router-link :to="{ name: 'org' }" class="hover:underline">{{
          org.name
          /* eslint-disable-next-line @intlify/vue-i18n/no-raw-text */
        }}</router-link>
        /
        {{ $t('settings') }}
      </span>
    </template>

    <Tab :to="{ name: 'org-settings-secrets' }" :title="$t('secrets.secrets')" />
    <Tab :to="{ name: 'org-settings-registries' }" :title="$t('registries.registries')" />
    <Tab
      v-if="useConfig().userRegisteredAgents"
      :to="{ name: 'org-settings-agents' }"
      :title="$t('admin.settings.agents.agents')"
    />

    <router-view />
  </Scaffold>
</template>

<script lang="ts" setup>
import { onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import useConfig from '~/compositions/useConfig';
import { inject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import { useRouteBack } from '~/compositions/useRouteBack';

const notifications = useNotifications();
const router = useRouter();
const i18n = useI18n();

const org = inject('org');
const orgPermissions = inject('org-permissions');

onMounted(async () => {
  if (!orgPermissions.value?.admin) {
    notifications.notify({ type: 'error', title: i18n.t('org.settings.not_allowed') });
    await router.replace({ name: 'home' });
  }
});

const goBack = useRouteBack({ name: 'org' });
</script>
