<template>
  <div>
    <FluidContainer class="flex border-b items-start items-center">
      <Breadcrumbs
        :paths="[
          { name: 'Repositories', link: { name: 'home' } },
          {
            name: `${repo.owner} / ${repo.name}`,
            link: { name: 'repo', params: { repoOwner: repo.owner, repoName: repo.name } },
          },
        ]"
      />
      <a :href="repo.link_url" target="_blank" class="flex ml-auto">
        <icon-github v-if="repo.link_url.startsWith('https://github.com/')" class="h-8 w-8" />
        <icon-repo v-else />
      </a>
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="ml-4">
        <img :src="badgeUrl" />
      </a>
    </FluidContainer>

    <FluidContainer v-if="builds" class="space-y-4">
      <router-link
        v-for="build in builds"
        :key="build.id"
        :to="{ name: 'repo-build', params: { repoOwner: repo.owner, repoName: repo.name, buildId: build.number } }"
        class="flex"
      >
        <BuildItem :build="build" />
      </router-link>
    </FluidContainer>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, inject, Ref } from 'vue';
import { Repo, Build } from '~/lib/api/types';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import BuildItem from '~/components/repo/BuildItem.vue';
import Breadcrumbs from '~/components/layout/Breadcrumbs.vue';
import IconGithub from 'virtual:vite-icons/mdi/github';
import IconRepo from 'virtual:vite-icons/teenyicons/git-solid';

export default defineComponent({
  name: 'Repo',

  components: { FluidContainer, BuildItem, Breadcrumbs, IconGithub, IconRepo },

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
