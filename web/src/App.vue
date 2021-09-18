<template>
  <div class="app flex flex-col m-auto w-full h-full bg-gray-100">
    <router-view v-if="blank" />
    <template v-else>
      <Navbar />
      <div class="flex min-h-0 h-full">
        <div class="flex flex-col overflow-y-auto flex-grow">
          <router-view />
        </div>
        <BuildFeedSidebar
          class="shadow-md bg-white border-l w-full absolute right-0 lg:relative max-w-80 xl:max-w-96"
        />
      </div>
    </template>
    <notifications position="bottom right" />
  </div>
</template>

<script lang="ts">
import { computed, defineComponent } from 'vue';
import { useRoute } from 'vue-router';

import BuildFeedSidebar from '~/components/build-feed/BuildFeedSidebar.vue';
import Navbar from '~/components/layout/header/Navbar.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';

export default defineComponent({
  name: 'App',

  components: {
    Navbar,
    BuildFeedSidebar,
  },

  setup() {
    const route = useRoute();
    const apiClient = useApiClient();
    const notifications = useNotifications();
    // eslint-disable-next-line promise/prefer-await-to-callbacks
    apiClient.setErrorHandler((err) => {
      notifications.notify({ title: err.message, type: 'error' });
    });

    const blank = computed(() => route.meta.blank);

    return { blank };
  },
});
</script>

<!-- eslint-disable-next-line vue-scoped-css/require-scoped -->
<style>
html,
body,
#app {
  width: 100%;
  height: 100%;
}

.vue-notification {
  @apply rounded-md text-base border-l-6;
}

.vue-notification .notification-title {
  @apply font-normal;
}

.vue-notification.success {
  @apply bg-lime-600 border-l-lime-700;
}

.vue-notification.error {
  @apply bg-red-600 border-l-red-700;
}
</style>

<style scoped>
.app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>
