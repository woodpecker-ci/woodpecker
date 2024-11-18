<template>
  <Scaffold enable-tabs>
    <template #title>{{ $t('user.settings.settings') }}</template>
    <template #headerActions><Button :text="$t('logout')" :to="`${address}/logout`" /></template>

    <Tab to="user" :title="$t('user.settings.general.general')" />
    <Tab to="user-secrets" :title="$t('secrets.secrets')" />
    <Tab to="user-registries" :title="$t('registries.registries')" />
    <Tab to="user-cli-and-api" :title="$t('user.settings.cli_and_api.cli_and_api')" />
    <Tab v-if="useConfig().userRegisteredAgents" to="user-agents" :title="$t('admin.settings.agents.agents')" />

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
