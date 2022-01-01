<template>
  <FluidContainer v-if="build" class="flex flex-col gap-y-6 text-gray-500 justify-between py-0">
    <div v-for="file in build.changed_files" :key="file" class="w-full">
      <div class="font-bold">{{ file }}</div>
    </div>
  </FluidContainer>
</template>

<script lang="ts">
import { defineComponent, inject, Ref } from 'vue';

import FluidContainer from '~/components/layout/FluidContainer.vue';
import { Build } from '~/lib/api/types';

export default defineComponent({
  name: 'BuildChangedFiles',

  components: {
    FluidContainer,
  },

  setup() {
    const build = inject<Ref<Build>>('build');
    if (!build) {
      throw new Error('Unexpected: "build" should be provided at this place');
    }

    return { build };
  },
});
</script>
