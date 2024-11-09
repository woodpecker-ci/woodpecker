<template>
  <Scaffold v-if="org" enable-tabs :go-back="goBack">
    <template #title>
      <span>
        <router-link :to="{ name: 'org' }" class="hover:underline">
          {{ org.name }}
          <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
        </router-link>
        /
        {{ $t('settings') }}
      </span>
    </template>

    <Tab id="secrets" :title="$t('secrets.secrets')">
      <OrgSecretsTab />
    </Tab>

    <Tab id="registries" :title="$t('registries.registries')">
      <OrgRegistriesTab />
    </Tab>

    <Tab id="agents" :title="$t('admin.settings.agents.agents')">
      <OrgAgentsTab />
    </Tab>
  </Scaffold>
</template>

<script lang="ts" setup>
import { onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import OrgAgentsTab from '~/components/org/settings/OrgAgentsTab.vue';
import OrgRegistriesTab from '~/components/org/settings/OrgRegistriesTab.vue';
import OrgSecretsTab from '~/components/org/settings/OrgSecretsTab.vue';
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
