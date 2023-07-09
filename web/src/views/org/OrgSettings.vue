<template>
  <Scaffold enable-tabs v-if="org" :go-back="goBack">
    <template #title>
      <span>
        <router-link :to="{ name: 'org' }" class="hover:underline">
          {{ org.name }}
        </router-link>
        /
        {{ $t('org.settings.settings') }}
      </span>
    </template>

    <Tab id="secrets" :title="$t('org.settings.secrets.secrets')">
      <OrgSecretsTab />
    </Tab>
  </Scaffold>
</template>

<script lang="ts" setup>
import { onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Tab from '~/components/layout/scaffold/Tab.vue';
import OrgSecretsTab from '~/components/org/settings/OrgSecretsTab.vue';
import { inject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import { useRouteBackOrDefault } from '~/compositions/useRouteBackOrDefault';

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

const goBack = useRouteBackOrDefault({ name: 'repos-owner' });
</script>
