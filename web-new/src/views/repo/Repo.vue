<template>
  <div v-if="repo">
    <FluidContainer class="flex border-b mb-4 items-start items-center">
      <Breadcrumbs
        :paths="[
          repo.owner,
          { name: repo.name, link: { name: 'repo', params: { repoOwner: repo.owner, repoName: repo.name } } },
        ]"
      />
      <a :href="repo.link_url" target="_blank" class="ml-auto">
        <icon-github v-if="repo.link_url.startsWith('https://github.com/')" class="h-8 w-8" />
        <icon-repo v-else />
      </a>
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="ml-4">
        <img :src="badgeUrl" />
      </a>
    </FluidContainer>

    <FluidContainer class="space-y-4">
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
import { computed, defineComponent, inject, onMounted, Ref, ref, toRef, watch } from 'vue';
import useApiClient from '~/compositions/useApiClient';
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
    const apiClient = useApiClient();

    const builds = ref<Build[] | undefined>();
    const repo = computed(() => inject<Ref<Repo>>('repo')?.value || null);

    const badgeUrl = computed(() => {
      if (!repo.value) {
        return null;
      }

      return `/api/badges/${repo.value.owner}/${repo.value.name}/status.svg`;
    });

    async function loadBuilds() {
      if (!repo.value) {
        return;
      }

      builds.value = await apiClient.getBuildList(repo.value.owner, repo.value.name);
    }

    onMounted(loadBuilds);
    watch(repo, loadBuilds);

    return { repo, builds, badgeUrl };
  },
});
</script>
