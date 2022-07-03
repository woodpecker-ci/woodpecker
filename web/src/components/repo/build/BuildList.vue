<template>
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
      <span class="text-color">{{ $t('repo.build.no_pipelines') }}</span>
    </Panel>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

import Panel from '~/components/layout/Panel.vue';
import BuildItem from '~/components/repo/build/BuildItem.vue';
import { Build, Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'BuildList',

  components: { Panel, BuildItem },

  props: {
    repo: {
      type: Object as PropType<Repo>,
      required: true,
    },

    builds: {
      type: Object as PropType<Build[] | undefined>,
      required: true,
    },
  },
});
</script>
