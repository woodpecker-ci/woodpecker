<template>
  <FluidContainer>
    <div class="flex border-b items-center pb-4 mb-4 dark:border-gray-600">
      <IconButton icon="back" @click="goBack" />
      <h1 class="text-xl ml-2 text-color">{{ $t('admin.settings.settings') }}</h1>
    </div>

    <Tabs>
      <Tab id="secrets" :title="$t('admin.settings.secrets.secrets')">
        <AdminSecretsTab />
      </Tab>
      <Tab id="agents" :title="$t('admin.settings.agents.agents')">
        <AdminAgentsTab />
      </Tab>
    </Tabs>
  </FluidContainer>
</template>

<script lang="ts" setup>
import { onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import AdminAgentsTab from '~/components/admin/settings/AdminAgentsTab.vue';
import AdminSecretsTab from '~/components/admin/settings/AdminSecretsTab.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Tab from '~/components/tabs/Tab.vue';
import Tabs from '~/components/tabs/Tabs.vue';
import useAuthentication from '~/compositions/useAuthentication';
import useNotifications from '~/compositions/useNotifications';
import { useRouteBackOrDefault } from '~/compositions/useRouteBackOrDefault';

const notifications = useNotifications();
const router = useRouter();
const i18n = useI18n();
const { user } = useAuthentication();

onMounted(async () => {
  if (!user?.admin) {
    notifications.notify({ type: 'error', title: i18n.t('admin.settings.not_allowed') });
    await router.replace({ name: 'home' });
  }
});

const goBack = useRouteBackOrDefault({ name: 'home' });
</script>
