<template><span /></template>

<script setup lang="ts">
import { onMounted } from 'vue';
import type { RouteLocationRaw } from 'vue-router';

import type { IconNames } from '~/components/atomic/Icon.vue';
import { useTabsClient } from '~/compositions/useTabs';

const props = defineProps<{
  to: RouteLocationRaw;
  title: string;
  icon?: IconNames;
  iconClass?: string;
  matchChildren?: boolean;
}>();

const { tabs } = useTabsClient();

onMounted(() => {
  // don't add tab if tab id is already present
  if (!tabs.value.find(({ to }) => to === props.to)) {
    tabs.value.push({
      to: props.to,
      title: props.title,
      icon: props.icon,
      iconClass: props.iconClass,
      matchChildren: props.matchChildren,
    });
  }
});
</script>
