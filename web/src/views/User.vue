<template>
  <Scaffold enable-tabs>
    <template #title>{{ $t('user.settings.settings') }}</template>
    <template #titleActions><Button :text="$t('logout')" :to="`${address}/logout`" /></template>
    <Tab id="general" :title="$t('user.settings.general.general')">
      <UserGeneralTab />
    </Tab>
    <Tab id="secrets" :title="$t('user.settings.secrets.secrets')">
      <UserSecretsTab />
    </Tab>
    <Tab id="api" :title="$t('user.settings.api.api')">
      <UserAPITab />
    </Tab>
  </Scaffold>
</template>

<script lang="ts" setup>
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import UserAPITab from '~/components/user/UserAPITab.vue';
import UserGeneralTab from '~/components/user/UserGeneralTab.vue';
import UserSecretsTab from '~/components/user/UserSecretsTab.vue';

const address = `${window.location.protocol}//${window.location.host}`; // port is included in location.host
</script>
