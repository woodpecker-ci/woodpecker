<template>
  <FluidContainer>
    <div class="flex border-b items-center pb-4 mb-4 dark:border-gray-600">
      <IconButton icon="back" :title="$t('back')" @click="goBack" />
      <h1 class="text-xl ml-2 text-color">{{ $t('org.settings.settings') }}</h1>
    </div>

    <Tabs>
      <Tab id="secrets" :title="$t('org.settings.secrets.secrets')">
        <OrgSecretsTab />
      </Tab>
    </Tabs>
  </FluidContainer>
</template>

<script lang="ts">
import { defineComponent, inject, onMounted, Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import IconButton from '~/components/atomic/IconButton.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import OrgSecretsTab from '~/components/org/settings/OrgSecretsTab.vue';
import Tab from '~/components/tabs/Tab.vue';
import Tabs from '~/components/tabs/Tabs.vue';
import useNotifications from '~/compositions/useNotifications';
import { useRouteBackOrDefault } from '~/compositions/useRouteBackOrDefault';
import { OrgPermissions } from '~/lib/api/types';

export default defineComponent({
  name: 'OrgSettings',

  components: {
    FluidContainer,
    IconButton,
    Tabs,
    Tab,
    OrgSecretsTab,
  },

  setup() {
    const notifications = useNotifications();
    const router = useRouter();
    const i18n = useI18n();

    const orgPermissions = inject<Ref<OrgPermissions>>('org-permissions');
    if (!orgPermissions) {
      throw new Error('Unexpected: "orgPermissions" should be provided at this place');
    }

    onMounted(async () => {
      if (!orgPermissions.value.admin) {
        notifications.notify({ type: 'error', title: i18n.t('org.settings.not_allowed') });
        await router.replace({ name: 'home' });
      }
    });

    return {
      goBack: useRouteBackOrDefault({ name: 'repos-owner' }),
    };
  },
});
</script>
