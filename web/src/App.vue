<template>
  <div class="app flex flex-col m-auto w-full h-full bg-gray-100 dark:bg-dark-gray-600">
    <router-view v-if="blank" />
    <template v-else>
      <Navbar />
      <div class="relative flex min-h-0 h-full">
        <div class="flex flex-col overflow-y-auto flex-grow">
          <router-view />
        </div>
        <transition name="slide-right">
          <PipelineFeedSidebar class="shadow-md border-l w-full absolute top-0 right-0 bottom-0 max-w-80 xl:max-w-96" />
        </transition>
      </div>
    </template>
    <notifications position="bottom right" />
  </div>
</template>

<script lang="ts">
import { computed, defineComponent } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import PipelineFeedSidebar from '~/components/pipeline-feed/PipelineFeedSidebar.vue';
import Navbar from '~/components/layout/header/Navbar.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';

export default defineComponent({
  name: 'App',

  components: {
    Navbar,
    PipelineFeedSidebar,
  },

  setup() {
    const route = useRoute();
    const apiClient = useApiClient();
    const notifications = useNotifications();
    const i18n = useI18n();

    // eslint-disable-next-line promise/prefer-await-to-callbacks
    apiClient.setErrorHandler((err) => {
      notifications.notify({ title: err.message || i18n.t('unknown_error'), type: 'error' });
    });

    const blank = computed(() => route.meta.blank);

    return { blank };
  },
});
</script>

<style scoped>
.app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
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
