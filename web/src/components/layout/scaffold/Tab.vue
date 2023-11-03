<template>
  <div v-if="$slots.default" v-show="isActive" :aria-hidden="!isActive">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';

import { Tab, useTabsClient } from '~/compositions/useTabs';

const props = defineProps<{
  id?: string;
  title: string;
}>();

const { tabs, activeTab } = useTabsClient();
const tab = ref<Tab>();

onMounted(() => {
  tab.value = {
    id: props.id || props.title.toLocaleLowerCase().replace(' ', '-') || tabs.value.length.toString(),
    title: props.title,
  };

  // don't add tab if tab id is already present
  if (!tabs.value.find(({ id }) => id === props.id)) {
    tabs.value.push(tab.value);
  }
});

const isActive = computed(() => tab.value && tab.value.id === activeTab.value);
</script>
