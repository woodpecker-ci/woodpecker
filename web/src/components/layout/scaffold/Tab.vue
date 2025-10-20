<template><span /></template>

<script setup lang="ts">
import { onMounted } from 'vue';
import type { RouteLocationRaw } from 'vue-router';

import type { IconNames } from '~/components/atomic/Icon.vue';
import { useTabsClient } from '~/compositions/useTabs';

const props = defineProps<{
  to: RouteLocationRaw;
  title: string;
  count?: number;
  icon?: IconNames;
  iconClass?: string;
  matchChildren?: boolean;
}>();

const { tabs } = useTabsClient();

// TODO: find a better way to compare routes like
// https://github.com/vuejs/router/blob/0eaaeb9697acd40ad524d913d0348748e9797acb/packages/router/src/utils/index.ts#L17
function isSameRoute(a: RouteLocationRaw, b: RouteLocationRaw): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

onMounted(() => {
  // don't add tab if tab id is already present
  if (tabs.value.find(({ to }) => isSameRoute(to, props.to))) {
    return;
  }

  tabs.value.push({
    to: props.to,
    title: props.title,
    count: props.count,
    icon: props.icon,
    iconClass: props.iconClass,
    matchChildren: props.matchChildren,
  });
});
</script>
