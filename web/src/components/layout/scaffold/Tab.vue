<template><span ref="anchor" /></template>

<script lang="ts">
import { markRaw, onBeforeUnmount, onMounted, toRaw, useTemplateRef } from 'vue';
import type { RouteLocationRaw } from 'vue-router';

import type { IconNames } from '~/components/atomic/Icon.vue';
import type { Tab } from '~/compositions/useTabs';
import { useTabsClient } from '~/compositions/useTabs';

// owners per registered tab entry; module-level so all Tab instances share
// it, WeakMap keeps the bookkeeping private
const ownersByTab = new WeakMap<Tab, Map<symbol, HTMLElement>>();
</script>

<script setup lang="ts">
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

const ownerId = Symbol('scaffold-tab-owner');
let registeredTab: Tab | undefined;
let ownerAnchor: HTMLElement | undefined;

// TODO: find a better way to compare routes like
// https://github.com/vuejs/router/blob/0eaaeb9697acd40ad524d913d0348748e9797acb/packages/router/src/utils/index.ts#L17
function isSameRoute(a: RouteLocationRaw, b: RouteLocationRaw): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

onMounted(() => {
  ownerAnchor = markRaw(anchor.value!);

  // join an existing entry for the same route as co-owner instead of
  // registering a duplicate
  const existing = tabs.value.find(({ to }) => isSameRoute(to, props.to));
  if (existing) {
    registeredTab = toRaw(existing);
    ownersByTab.get(registeredTab)?.set(ownerId, ownerAnchor);
    return;
  }

  const tab = {
    to: props.to,
    title: props.title,
    count: props.count,
    icon: props.icon,
    iconClass: props.iconClass,
    matchChildren: props.matchChildren,
    anchor: ownerAnchor,
  };
  registeredTab = tab;
  ownersByTab.set(tab, new Map([[ownerId, ownerAnchor]]));

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

  const owners = ownersByTab.get(registeredTab);
  owners?.delete(ownerId);

  if (owners !== undefined && owners.size > 0) {
    // another instance of the same route is still mounted, keep the entry;
    // transfer the anchor if it was ours so ordered insertion stays correct
    if (registeredTab.anchor === ownerAnchor) {
      registeredTab.anchor = owners.values().next().value;
    }
    return;
  }

  // compare raw objects because the tabs ref wraps its entries in reactive proxies
  tabs.value = tabs.value.filter((tab) => toRaw(tab) !== registeredTab);
  ownersByTab.delete(registeredTab);
});
</script>
