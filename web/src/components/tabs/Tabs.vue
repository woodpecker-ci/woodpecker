<template>
  <div class="flex flex-col">
    <div class="flex w-full pt-4 mb-4">
      <div
        v-for="tab in tabs"
        :key="tab.id"
        class="flex cursor-pointer pb-2 px-8 text-gray-600 border-b-2"
        :class="{
          'border-gray-500 text-gray-600 hover:border-gray-600 dark:border-gray-600 dark:hover:border-gray-500':
            activeTab === tab.id,
          'border-transparent hover:border-gray-300 dark:hover:border-gray-700': activeTab !== tab.id,
        }"
        @click="selectTab(tab)"
      >
        <span>{{ tab.title }}</span>
      </div>
    </div>

    <div>
      <slot />
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, provide, ref, toRef } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Tab } from './types';

export default defineComponent({
  name: 'Tabs',

  props: {
    // used by toRef
    // eslint-disable-next-line vue/no-unused-properties
    disableHashMode: {
      type: Boolean,
    },

    // used by toRef
    // eslint-disable-next-line vue/no-unused-properties
    modelValue: {
      type: String,
      default: '',
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    'update:modelValue': (_value: string): boolean => true,
  },

  setup(props) {
    const router = useRouter();
    const route = useRoute();

    const disableHashMode = toRef(props, 'disableHashMode');
    const modelValue = toRef(props, 'modelValue');
    const tabs = ref<Tab[]>([]);
    const activeTab = ref(route.hash.replace(/^#tab-/, '') || undefined);
    provide('tabs', tabs);
    provide(
      'active-tab',
      computed(() => activeTab.value || modelValue.value || '0'),
    );

    async function selectTab(tab: Tab) {
      if (tab.id === undefined) {
        return;
      }

      if (activeTab.value === undefined) {
        throw new Error('Please wrap this "Tab"-component inside a "Tabs" list.');
      }

      activeTab.value = tab.id;

      if (!disableHashMode.value) {
        await router.push({ params: route.params, hash: `#tab-${tab.id}` });
      }
    }

    return { tabs, activeTab, selectTab };
  },
});
</script>
