<template>
  <Scaffold enable-tabs>
    <template #title>{{ $t('user.user_settings') }}</template>
    <template #titleActions><Button :text="$t('logout')" :to="`${address}/logout`" /></template>
    <Tab id="settings" :title="$t('user.settings.settings')">
      <UserSettingsTab />
    </Tab>
    <Tab id="api" :title="$t('user.api.api')">
      <UserAPITab />
    </Tab>
  </Scaffold>
</template>

<script lang="ts" setup>
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import UserAPITab from '~/components/user/UserAPITab.vue';
import UserSettingsTab from '~/components/user/UserSettingsTab.vue';

const address = `${window.location.protocol}//${window.location.host}`; // port is included in location.host
</script>
