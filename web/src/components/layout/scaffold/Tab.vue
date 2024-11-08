<template>
  <div v-if="$slots.default" v-show="isActive" :aria-hidden="!isActive">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';

import type { IconNames } from '~/components/atomic/Icon.vue';
import { useTabsClient, type Tab } from '~/compositions/useTabs';

const props = defineProps<{
  to?: string;
  alternativeRoute?: string;
  title: string;
  icon?: IconNames;
  iconClass?: string;
}>();

const { tabs, activeTab } = useTabsClient();
const tab = ref<Tab>();

onMounted(() => {
  tab.value = {
    to: props.to || props.title.toLocaleLowerCase().replace(' ', '-') || tabs.value.length.toString(),
    alternativeRoute: props.alternativeRoute,
    title: props.title,
    icon: props.icon,
    iconClass: props.iconClass,
  };

  // don't add tab if tab id is already present
  if (!tabs.value.find(({ to }) => to === props.to)) {
    tabs.value.push(tab.value);
  }
});

const isActive = computed(() => tab.value && tab.value.to === activeTab.value);
</script>
