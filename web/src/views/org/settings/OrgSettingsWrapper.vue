<template>
  <Scaffold v-if="org" v-model:active-tab="activeTab" enable-tabs :go-back="goBack" disable-tab-url-hash-mode>
    <template #title>
      <span>
        <router-link :to="{ name: 'org' }" class="hover:underline">
          {{ org.name }}
          <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
        </router-link>
        /
        {{ $t('settings') }}
      </span>
    </template>

    <Tab id="secrets" :title="$t('secrets.secrets')" />
    <Tab id="registries" :title="$t('registries.registries')" />
    <Tab id="agents" :title="$t('admin.settings.agents.agents')" />

    <router-view />
  </Scaffold>
</template>

<script lang="ts" setup>
import { computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import { inject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import { useRouteBack } from '~/compositions/useRouteBack';

const notifications = useNotifications();
const route = useRoute();
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

const goBack = useRouteBack({ name: 'org' });

const activeTab = computed({
  get() {
    if (route.name === 'org-settings-secrets') {
      return 'secrets';
    }
    if (route.name === 'org-settings-registries') {
      return 'registries';
    }
    if (route.name === 'org-settings-agents') {
      return 'agents';
    }
    return 'secrets';
  },
  set(tab: string) {
    if (tab === 'secrets') {
      router.push({ name: 'org-settings-secrets' });
    } else if (tab === 'registries') {
      router.push({ name: 'org-settings-registries' });
    } else if (tab === 'agents') {
      router.push({ name: 'org-settings-agents' });
    } else {
      router.push({ name: 'org-settings-secrets' });
    }
  },
});
</script>
