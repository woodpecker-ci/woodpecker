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
/*import { onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
*/
import UserSettingsTab from '~/components/user/UserSettingsTab.vue';
import UserAPITab from '~/components/user/UserAPITab.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
/*
const notifications = useNotifications();
const router = useRouter();
const i18n = useI18n();
const { user } = useAuthentication();

onMounted(async () => {
  if (!user?.admin) {
    notifications.notify({ type: 'error', title: i18n.t('admin.settings.not_allowed') });
    await router.replace({ name: 'home' });
  }

  if (!user?.admin) {
    notifications.notify({ type: 'error', title: i18n.t('admin.settings.not_allowed') });
    await router.replace({ name: 'home' });
  }
});*/
</script>
