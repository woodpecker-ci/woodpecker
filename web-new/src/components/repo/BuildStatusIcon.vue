<template>
  <div v-if="build" class="flex items-center justify-center">
    <Icon name="status-blocked" v-if="build.status === 'blocked'" class="w-6 h-6 text-status-red" />
    <Icon name="status-declined" v-if="build.status === 'declined'" class="w-6 h-6 text-status-red" />
    <Icon name="status-error" v-if="build.status === 'error'" class="w-8 h-8 text-status-red" />
    <Icon name="status-failure" v-if="build.status === 'failure'" class="w-7 h-7 text-status-red" />
    <Icon name="status-killed" v-if="build.status === 'killed'" class="w-8 h-8 text-status-red" />
    <Icon name="status-pending" v-if="build.status === 'pending'" class="w-7 h-7 text-status-gray" />
    <Icon name="status-skipped" v-if="build.status === 'skipped'" class="w-8 h-8 text-status-gray" />
    <Icon
      name="status-running"
      v-if="build.status === 'running'"
      class="w-8 h-8 text-status-blue animate-spin animate-slow"
    />
    <Icon
      name="status-started"
      v-if="build.status === 'started'"
      class="w-8 h-8 text-status-blue animate-spin animate-slow"
    />
    <Icon name="status-success" v-if="build.status === 'success'" class="w-8 h-8 text-status-green" />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, ref, toRef } from 'vue';
import { Build } from '~/lib/api/types';
import Icon from '~/components/atomic/Icon.vue';

export default defineComponent({
  name: 'BuildStatusIcon',

  components: {
    Icon,
  },

  props: {
    build: {
      type: Object as PropType<Build>,
      required: true,
    },
  },
});
</script>
