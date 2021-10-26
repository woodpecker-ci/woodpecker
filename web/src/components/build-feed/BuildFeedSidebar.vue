<template>
  <div v-if="isBuildFeedOpen" class="flex flex-col overflow-y-auto items-center">
    <router-link
      v-for="build in sortedBuildFeed"
      :key="build.id"
      :to="{ name: 'repo-build', params: { repoOwner: build.owner, repoName: build.name, buildId: build.number } }"
      class="
        flex
        border-b
        py-4
        px-2
        w-full
        hover:bg-light-300
        dark:hover:bg-dark-400 dark:border-dark-300
        hover:shadow-sm
      "
    >
      <BuildFeedItem :build="build" />
    </router-link>

    <span v-if="sortedBuildFeed.length === 0" class="text-gray-500 m-4">There are no builds yet.</span>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

import BuildFeedItem from '~/components/build-feed/BuildFeedItem.vue';
import useBuildFeed from '~/compositions/useBuildFeed';

export default defineComponent({
  name: 'BuildFeedSidebar',

  components: { BuildFeedItem },

  setup() {
    const buildFeed = useBuildFeed();

    return {
      isBuildFeedOpen: buildFeed.isOpen,
      sortedBuildFeed: buildFeed.sortedBuilds,
    };
  },
});
</script>
