<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <h1 class="text-xl ml-2 text-color">{{ $t('repo.settings.badge.badge') }}</h1>
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="ml-auto">
        <img :src="badgeUrl" />
      </a>
    </div>

    <div class="flex flex-col space-y-4">
      <div>
        <h2 class="text-lg text-color ml-2">{{ $t('url') }}</h2>
        <pre class="box">{{ baseUrl }}{{ badgeUrl }}</pre>
      </div>

      <div>
        <h2 class="text-lg text-color ml-2">{{ $t('repo.settings.badge.url_branch') }}</h2>
        <pre class="box">{{ baseUrl }}{{ badgeUrl }}?branch=<span class="font-bold">&lt;branch&gt;</span></pre>
      </div>

      <div>
        <h2 class="text-lg text-color ml-2">{{ $t('repo.settings.badge.markdown') }}</h2>
        <pre class="box">[![status-badge]({{ baseUrl }}{{ badgeUrl }})]({{ baseUrl }}{{ repoUrl }})</pre>
      </div>
    </div>
  </Panel>
</template>

<script lang="ts">
import { computed, defineComponent, inject, Ref } from 'vue';

import Panel from '~/components/layout/Panel.vue';
import { Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'BadgeTab',

  components: { Panel },

  setup() {
    const repo = inject<Ref<Repo>>('repo');
    const baseUrl = `${window.location.protocol}//${window.location.hostname}`;
    const badgeUrl = computed(() => {
      if (!repo) {
        throw new Error('Unexpected: "repo" should be provided at this place');
      }

      return `/api/badges/${repo.value.owner}/${repo.value.name}/status.svg`;
    });
    const repoUrl = computed(() => {
      if (!repo) {
        throw new Error('Unexpected: "repo" should be provided at this place');
      }

      return `/${repo.value.owner}/${repo.value.name}`;
    });

    return { baseUrl, badgeUrl, repoUrl };
  },
});
</script>

<style scoped>
.box {
  @apply bg-gray-500 p-2 rounded-md text-white break-words dark:bg-dark-400 dark:text-gray-400;
  white-space: pre-wrap;
}
</style>
