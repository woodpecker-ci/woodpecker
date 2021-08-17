<template>
  <FluidContainer>
    <div class="flex border-b items-center pb-4 mb-4">
      <h1 class="text-xl">{{ `${repo.owner} / ${repo.name}` }}</h1>
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="ml-auto">
        <img :src="badgeUrl" />
      </a>
      <a :href="repo.link_url" target="_blank" class="flex ml-4 text-gray-400 hover:text-gray-500">
        <Icon name="github" v-if="repo.link_url.startsWith('https://github.com/')" />
        <Icon name="repo" v-else />
      </a>
      <IconButton class="ml-2" :to="{ name: 'repo-settings' }" icon="settings" />
    </div>

    <div v-if="builds" class="space-y-4">
      <router-link
        v-for="build in builds"
        :key="build.id"
        :to="{ name: 'repo-build', params: { repoOwner: repo.owner, repoName: repo.name, buildId: build.number } }"
        class="flex"
      >
        <BuildItem :build="build" />
      </router-link>
    </div>
  </FluidContainer>
</template>

<script lang="ts">
import { computed, defineComponent, inject, Ref } from 'vue';
import { Repo, Build } from '~/lib/api/types';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import BuildItem from '~/components/repo/BuildItem.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import Icon from '~/components/atomic/Icon.vue';

export default defineComponent({
  name: 'Repo',

  components: { FluidContainer, BuildItem, IconButton, Icon },

  setup() {
    const repo = inject<Ref<Repo>>('repo');
    if (!repo) {
      throw new Error('Unexpected: "repo" should be provided at this place');
    }

    const badgeUrl = computed(() => `/api/badges/${repo.value.owner}/${repo.value.name}/status.svg`);
    const builds = inject<Ref<Build[]>>('builds');

    return { repo, builds, badgeUrl };
  },
});
</script>
