<template>
  <div v-if="$slots.default" v-show="isActive" :aria-hidden="!isActive">
    <slot />
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, inject, onMounted, Ref, ref } from 'vue';

import { Tab } from './types';

export default defineComponent({
  name: 'Tab',

  props: {
    id: {
      type: String,
      default: undefined,
    },

    title: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const activeTab = inject<Ref<string>>('active-tab');
    const tabs = inject<Ref<Tab[]>>('tabs');
    if (activeTab === undefined || tabs === undefined) {
      throw new Error('Please wrap this "Tab"-component inside a "Tabs" list.');
    }

    const tab = ref<Tab>();

    onMounted(() => {
      tab.value = {
        id: props.id || props.title.toLocaleLowerCase().replace(' ', '-') || tabs.value.length.toString(),
        title: props.title,
      };
      tabs.value.push(tab.value);
    });

    const isActive = computed(() => tab.value && tab.value.id === activeTab.value);

    return { isActive };
  },
});
</script>
