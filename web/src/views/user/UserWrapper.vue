<template>
  <Scaffold enable-tabs>
    <template #title>{{ $t('user.settings.settings') }}</template>
    <template #titleActions><Button :text="$t('logout')" :to="`${address}/logout`" /></template>

    <Tab id="user-general" :title="$t('user.settings.general.general')" />
    <Tab id="user-secrets" :title="$t('secrets.secrets')" />
    <Tab id="user-registries" :title="$t('registries.registries')" />
    <Tab id="user-cli-and-api" :title="$t('user.settings.cli_and_api.cli_and_api')" />
    <Tab id="user-agents" :title="$t('admin.settings.agents.agents')" />

    <router-view />
  </Scaffold>
</template>

<script lang="ts" setup>

import Button from '~/components/atomic/Button.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import useConfig from '~/compositions/useConfig';

const address = `${window.location.protocol}//${window.location.host}${useConfig().rootPath}`; // port is included in location.host
</script>
