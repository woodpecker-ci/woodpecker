<template>
  <div class="app bg-wp-background-200 dark:bg-wp-background-100 m-auto flex h-full w-full flex-col">
    <router-view v-if="blank" />
    <template v-else>
      <Navbar />
      <main class="relative flex h-full min-h-0">
        <div id="scroll-component" class="flex flex-grow flex-col overflow-y-auto">
          <router-view />
        </div>
        <transition name="slide-right">
          <PipelineFeedSidebar class="absolute bottom-0 right-0 top-0 w-full max-w-80 border-l shadow-md xl:max-w-96" />
        </transition>
      </main>
    </template>
    <notifications position="bottom right" />
  </div>
</template>

<script lang="ts" setup>
import { computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import Navbar from '~/components/layout/header/Navbar.vue';
import PipelineFeedSidebar from '~/components/pipeline-feed/PipelineFeedSidebar.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';

const route = useRoute();
const apiClient = useApiClient();
const { notify } = useNotifications();
const i18n = useI18n();

// eslint-disable-next-line promise/prefer-await-to-callbacks
apiClient.setErrorHandler((err) => {
  if (err.status === 404) {
    notify({ title: i18n.t('errors.not_found'), type: 'error' });
    return;
  }
  notify({ title: err.message || i18n.t('unknown_error'), type: 'error' });
});

const blank = computed(() => route.meta.blank);

const { locale } = useI18n();
watch(
  locale,
  () => {
    document.documentElement.setAttribute('lang', locale.value);
  },
  { immediate: true },
);
</script>

<style scoped>
.app {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.slide-right-enter-active,
.slide-right-leave-active {
  transition: all 0.3s ease;
}
.slide-right-enter-from,
.slide-right-leave-to {
  transform: translate(100%, 0);
}
</style>
