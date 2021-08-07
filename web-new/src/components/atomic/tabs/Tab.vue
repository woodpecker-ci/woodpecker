<template>
  <div v-show="isActive" :aria-hidden="!isActive">
    <slot />
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, inject, onMounted, ref, Ref, toRef } from 'vue';
import { Tab } from '~/components/atomic/tabs/types';

export default defineComponent({
  name: 'Tab',

  props: {
    title: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const activeTab = inject<Ref<number>>('active-tab');
    const tabs = inject<Ref<Tab[]>>('tabs');
    if (activeTab === undefined || tabs === undefined) {
      throw new Error('Please wrap this "Tab"-component inside a "Tabs" list.');
    }

    const tab = ref<Tab>({
      id: undefined,
      title: props.title,
      slug: props.title.toLowerCase().replace(/ /g, '-'),
    });

    onMounted(() => {
      tab.value = {
        ...tab.value,
        id: tabs.value.length,
      };
      tabs.value.push(tab.value);
    });

    const isActive = computed(() => {
      return tab && tab.value.id === activeTab.value;
    });

    return { isActive };
  },
});
</script>
