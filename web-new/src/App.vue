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
import { computed, defineComponent, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import Navbar from '~/components/layout/header/Navbar.vue';
import BuildFeedSidebar from './components/build-feed/BuildFeedSidebar.vue';

export default defineComponent({
  name: 'App',

  components: {
    Navbar,
    BuildFeedSidebar,
  },

  setup() {
    const route = useRoute();
    const blank = computed(() => route.meta.blank);
    return { blank };
  },
});
</script>

<style>
html,
body,
#app {
  width: 100%;
  height: 100%;
}

.vue-notification {
  @apply rounded-md text-lg border-l-10;
}

.vue-notification .notification-title {
  @apply font-normal;
}
</style>

<style scoped>
.app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>
