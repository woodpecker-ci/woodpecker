<template>
  <FluidContainer>
    <div class="flex border-b dark:border-b-gray-600 items-center pb-4 mb-4">
      <h1 class="text-xl text-gray-500">{{ `${repo.owner} / ${repo.name}` }}</h1>
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="ml-auto">
        <img :src="badgeUrl" />
      </a>
      <a
        :href="repo.link_url"
        target="_blank"
        class="flex ml-4 p-1 rounded-full text-gray-600 hover:bg-gray-200 hover:text-gray-700 dark:hover:bg-gray-600"
      >
        <Icon v-if="repo.link_url.startsWith('https://github.com/')" name="github" />
        <Icon v-if="repo.link_url.startsWith('https://github.com/')" name="gitea" />
        <Icon v-else name="repo" />
      </a>
      <IconButton v-if="isAuthenticated" class="ml-2" :to="{ name: 'repo-settings' }" icon="settings" />
    </div>

    <Tabs>
      <Tab title="Activity">
        <div v-if="builds" class="space-y-4">
          <router-link
            v-for="build in builds"
            :key="build.id"
            :to="{ name: 'repo-build', params: { repoOwner: repo.owner, repoName: repo.name, buildId: build.number } }"
            class="flex"
          >
            <BuildItem :build="build" />
          </router-link>
          <Panel v-if="builds.length === 0">
            <span class="text-gray-500">There are no builds yet.</span>
          </Panel>
        </div>
      </Tab>
      <Tab title="Branches"> TODO </Tab>
    </Tabs>
  </FluidContainer>
</template>

<script lang="ts">
import { computed, defineComponent, inject, Ref } from 'vue';

import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Panel from '~/components/layout/Panel.vue';
import BuildItem from '~/components/repo/build/BuildItem.vue';
import Tab from '~/components/tabs/Tab.vue';
import Tabs from '~/components/tabs/Tabs.vue';
import useAuthentication from '~/compositions/useAuthentication';
import { Build, Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'Repo',

  components: { FluidContainer, BuildItem, IconButton, Icon, Panel, Tabs, Tab },

  setup() {
    const { isAuthenticated } = useAuthentication();
    const repo = inject<Ref<Repo>>('repo');
    if (!repo) {
      throw new Error('Unexpected: "repo" should be provided at this place');
    }

    const badgeUrl = computed(() => `/api/badges/${repo.value.owner}/${repo.value.name}/status.svg`);
    const builds = inject<Ref<Build[]>>('builds');

    return { isAuthenticated, repo, builds, badgeUrl };
  },
});
</script>
