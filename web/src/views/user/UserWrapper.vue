<template>
  <Scaffold v-model:active-tab="activeTab" enable-tabs disable-tab-url-hash-mode>
    <template #title>{{ $t('user.settings.settings') }}</template>
    <template #titleActions><Button :text="$t('logout')" :to="`${address}/logout`" /></template>

    <Tab id="general" :title="$t('user.settings.general.general')" />
    <Tab id="secrets" :title="$t('secrets.secrets')" />
    <Tab id="registries" :title="$t('registries.registries')" />
    <Tab id="cli_and_api" :title="$t('user.settings.cli_and_api.cli_and_api')" />
    <Tab id="agents" :title="$t('admin.settings.agents.agents')" />

    <router-view />
  </Scaffold>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import useConfig from '~/compositions/useConfig';

const route = useRoute();
const router = useRouter();

const address = `${window.location.protocol}//${window.location.host}${useConfig().rootPath}`; // port is included in location.host

const activeTab = computed({
  get() {
    if (route.name === 'user-secrets') {
      return 'secrets';
    }
    if (route.name === 'user-registries') {
      return 'registries';
    }
    if (route.name === 'user-cli-and-api') {
      return 'cli_and_api';
    }
    if (route.name === 'user-agents') {
      return 'agents';
    }
    return 'general';
  },
  set(tab: string) {
    if (tab === 'secrets') {
      router.push({ name: 'user-secrets' });
    } else if (tab === 'registries') {
      router.push({ name: 'user-registries' });
    } else if (tab === 'cli_and_api') {
      router.push({ name: 'user-cli-and-api' });
    } else if (tab === 'agents') {
      router.push({ name: 'user-agents' });
    } else {
      router.push({ name: 'user-general' });
    }
  },
});
</script>
