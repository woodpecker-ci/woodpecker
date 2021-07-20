<template>
  <div v-if="isBuildFeedOpen" class="flex flex-col space-y-2">
    <router-link
      v-for="build in builds"
      :to="{ name: 'repo-build', params: { repoOwner: build.owner, repoId: build.name, buildId: build.number } }"
      class="flex border-b py-4 px-2 hover:bg-light-300 hover:shadow-sm"
    >
      <div class="flex items-center">
        <span>{{ build.status }}</span>
      </div>
      <div class="flex flex-col">
        <span>{{ build.owner }} / {{ build.name }}</span>
        <span>{{ build.message }}</span>
      </div>
    </router-link>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref, toRef } from 'vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Button from '~/components/atomic/Button.vue';
import BuildItem from '~/components/repo/BuildItem.vue';
import useBuildFeed from '~/compositions/useBuildFeed';
import useUserConfig from '~/compositions/useUserConfig';

export default defineComponent({
  name: 'FeedSidebar',

  components: { FluidContainer, Button, BuildItem },

  setup() {
    const { builds, isBuildFeedOpen } = useBuildFeed();

    return { builds, isBuildFeedOpen };
  },
});
</script>
