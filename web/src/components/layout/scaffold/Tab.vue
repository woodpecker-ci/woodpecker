<template>
  <div v-if="$slots.default" v-show="isActive" :aria-hidden="!isActive">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';

import { Tab, useTabsClient } from '~/compositions/useTabs';

export interface Props {
  id?: string;
  title: string;
}

const props = defineProps<Props>();

const { tabs, activeTab } = useTabsClient();
const tab = ref<Tab>();

onMounted(() => {
  tab.value = {
    id: props.id || props.title.toLocaleLowerCase().replace(' ', '-') || tabs.value.length.toString(),
    title: props.title,
  };
  tabs.value.push(tab.value);
});

const isActive = computed(() => tab.value && tab.value.id === activeTab.value);
</script>
