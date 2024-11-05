<template>
  <Scaffold enable-tabs>
    <template #title>{{ $t('user.settings.settings') }}</template>
    <template #titleActions><Button :text="$t('logout')" :to="`${address}/logout`" /></template>
    <Tab id="general" :title="$t('user.settings.general.general')">
      <UserGeneralTab />
    </Tab>
    <Tab id="secrets" :title="$t('secrets.secrets')">
      <UserSecretsTab />
    </Tab>
    <Tab id="registries" :title="$t('registries.registries')">
      <UserRegistriesTab />
    </Tab>
    <Tab id="cli-and-api" :title="$t('user.settings.cli_and_api.cli_and_api')">
      <UserCLIAndAPITab />
    </Tab>
    <Tab id="agents" :title="$t('admin.settings.agents.agents')">
      <UserAgentsTab />
    </Tab>
  </Scaffold>
</template>

<script lang="ts" setup>
import Button from '~/components/atomic/Button.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import UserAgentsTab from '~/components/user/UserAgentsTab.vue';
import UserCLIAndAPITab from '~/components/user/UserCLIAndAPITab.vue';
import UserGeneralTab from '~/components/user/UserGeneralTab.vue';
import UserRegistriesTab from '~/components/user/UserRegistriesTab.vue';
import UserSecretsTab from '~/components/user/UserSecretsTab.vue';
import useConfig from '~/compositions/useConfig';

const address = `${window.location.protocol}//${window.location.host}${useConfig().rootPath}`; // port is included in location.host
</script>
