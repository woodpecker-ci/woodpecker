<template>
  <FluidContainer v-if="build" class="flex flex-col gap-y-6 text-color justify-between py-0">
    <Panel>
      <div v-if="build.changed_files === undefined || build.changed_files.length < 1" class="w-full">
        <span class="text-color">{{ $t('repo.pipeline.no_files') }}</span>
      </div>
      <div v-for="file in build.changed_files" v-else :key="file" class="w-full">
        <div>- {{ file }}</div>
      </div>
    </Panel>
  </FluidContainer>
</template>

<script lang="ts">
import { defineComponent, inject, Ref } from 'vue';

import FluidContainer from '~/components/layout/FluidContainer.vue';
import Panel from '~/components/layout/Panel.vue';
import { Pipeline } from '~/lib/api/types';

export default defineComponent({
  name: 'BuildChangedFiles',

  components: {
    FluidContainer,
    Panel,
  },

  setup() {
    const build = inject<Ref<Pipeline>>('build');
    if (!build) {
      throw new Error('Unexpected: "build" should be provided at this place');
    }

    return { build };
  },
});
</script>
