<template><span ref="anchor" /></template>

<script setup lang="ts">
import { markRaw, onBeforeUnmount, onMounted, toRaw, useTemplateRef } from 'vue';
import type { RouteLocationRaw } from 'vue-router';

import type { IconNames } from '~/components/atomic/Icon.vue';
import type { Tab } from '~/compositions/useTabs';
import { useTabsClient } from '~/compositions/useTabs';

const props = defineProps<{
  to: RouteLocationRaw;
  title: string;
  count?: number;
  icon?: IconNames;
  iconClass?: string;
  matchChildren?: boolean;
}>();

const anchor = useTemplateRef('anchor');

const { tabs } = useTabsClient();

// the entry this instance registered; stays undefined when the mount-time
// dedup skips registration, so unmount won't touch another instance's entry
let registeredTab: Tab | undefined;

// TODO: find a better way to compare routes like
// https://github.com/vuejs/router/blob/0eaaeb9697acd40ad524d913d0348748e9797acb/packages/router/src/utils/index.ts#L17
function isSameRoute(a: RouteLocationRaw, b: RouteLocationRaw): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

onMounted(() => {
  // don't add tab if tab id is already present
  if (tabs.value.some(({ to }) => isSameRoute(to, props.to))) {
    return;
  }

  const tab = {
    to: props.to,
    title: props.title,
    count: props.count,
    icon: props.icon,
    iconClass: props.iconClass,
    matchChildren: props.matchChildren,
    anchor: markRaw(anchor.value!),
  };
  registeredTab = tab;

  // insert before the first tab whose anchor element comes after ours, so a
  // tab mounting later than its siblings still ends up in template order
  const index = tabs.value.findIndex(
    ({ anchor: other }) =>
      other !== undefined && (anchor.value!.compareDocumentPosition(other) & Node.DOCUMENT_POSITION_FOLLOWING) !== 0,
  );
  if (index === -1) {
    tabs.value.push(tab);
  } else {
    tabs.value.splice(index, 0, tab);
  }
});

onBeforeUnmount(() => {
  if (registeredTab === undefined) {
    return;
  }

  // compare raw objects because the tabs ref wraps its entries in reactive proxies
  tabs.value = tabs.value.filter((tab) => toRaw(tab) !== registeredTab);
});
</script>
