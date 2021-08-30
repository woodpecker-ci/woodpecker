<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center">
      <h1 class="text-xl ml-2">Badge</h1>
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="ml-auto">
        <img :src="badgeUrl" />
      </a>
    </div>

    <div class="flex flex-col space-y-4">
      <div>
        <h2 class="text-lg">Url</h2>
        <pre class="box">{{ baseUrl }}{{ badgeUrl }}</pre>
      </div>

      <div>
        <h2 class="text-lg">Url for specific branch</h2>
        <pre class="box">{{ baseUrl }}{{ badgeUrl }}?branch=<span class="font-bold">&lt;branch&gt;</span></pre>
      </div>

      <div>
        <h2 class="text-lg">Markdown</h2>
        <pre class="box">![status-badge]({{ baseUrl }}{{ badgeUrl }})</pre>
      </div>
    </div>
  </Panel>
</template>

<script lang="ts">
import { computed, defineComponent, inject, onMounted, Ref } from 'vue';
import { Repo } from '~/lib/api/types';
import Panel from '~/components/layout/Panel.vue';

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

    return { baseUrl, badgeUrl };
  },
});
</script>

<style scoped>
.box {
  @apply bg-gray-400 p-2 rounded-md text-white break-words;
  white-space: pre-wrap;
}
</style>
