<template>
  <div v-if="isBuildFeedOpen" class="flex flex-col overflow-y-auto">
    <router-link
      v-for="build in sortedBuildFeed"
      :to="{ name: 'repo-build', params: { repoOwner: build.owner, repoName: build.name, buildId: build.number } }"
      class="flex border-b py-4 px-2 hover:bg-light-300 hover:shadow-sm"
    >
      <BuildFeedItem :build="build" />
    </router-link>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Button from '~/components/atomic/Button.vue';
import BuildItem from '~/components/repo/BuildItem.vue';
import useBuildFeed from '~/compositions/useBuildFeed';
import { convertEmojis } from '~/utils/emoji';
import BuildFeedItem from '~/components/build-feed/BuildFeedItem.vue';

export default defineComponent({
  name: 'BuildFeedSidebar',

  components: { FluidContainer, Button, BuildItem, BuildFeedItem },

  setup() {
    const buildFeed = useBuildFeed();

    return {
      isBuildFeedOpen: buildFeed.isOpen,
      sortedBuildFeed: buildFeed.sortedBuilds,
      convertEmojis,
    };
  },
});
</script>
