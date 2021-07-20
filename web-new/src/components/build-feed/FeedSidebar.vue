<template>
  <div class="flex flex-col space-y-2">
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
import useApiClient from '~/compositions/useApiClient';
import { Repo, Build } from '~/lib/api/types';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Button from '~/components/atomic/Button.vue';
import { useRouter } from 'vue-router';
import useNotifications from '~/compositions/useNotifications';
import BuildItem from '~/components/repo/BuildItem.vue';

export default defineComponent({
  name: 'FeedSidebar',

  components: { FluidContainer, Button, BuildItem },

  setup(props) {
    const apiClient = useApiClient();
    const router = useRouter();

    const builds = ref<Build[] | undefined>();

    /**
     * Compare two feed items by name.
     * @param {Object} a - A feed item.
     * @param {Object} b - A feed item.
     * @returns {number}
     */
    const compareFeedItem = (a: Build, b: Build) => {
      return (b.started_at || b.created_at || -1) - (a.started_at || a.created_at || -1);
    };

    onMounted(async () => {
      const b = await apiClient.getBuildFeed();
      builds.value = b.sort(compareFeedItem);
    });

    return { builds };
  },
});
</script>
