<template>
  <div class="flex flex-col">
    <div class="flex w-full pt-4 mb-4">
      <div
        v-for="tab in tabs"
        :key="tab.id"
        class="
          flex
          cursor-pointer
          pb-2
          px-8
          border-b-2
          text-gray-500
          hover:text-gray-700
          dark:text-gray-500 dark:hover:text-gray-400
        "
        :class="{
          'border-gray-400 dark:border-gray-600': activeTab === tab.id,
          'border-transparent': activeTab !== tab.id,
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
import { computed, defineComponent, onMounted, provide, ref, toRef } from 'vue';
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

  setup(props, { emit }) {
    const router = useRouter();
    const route = useRoute();

    const disableHashMode = toRef(props, 'disableHashMode');
    const modelValue = toRef(props, 'modelValue');
    const tabs = ref<Tab[]>([]);
    const activeTab = ref();
    provide('tabs', tabs);
    provide(
      'active-tab',
      computed(() => activeTab.value),
    );

    async function selectTab(tab: Tab) {
      if (tab.id === undefined) {
        return;
      }

      activeTab.value = tab.id;
      emit('update:modelValue', activeTab.value);

      if (!disableHashMode.value) {
        await router.replace({ params: route.params, hash: `#${tab.id}` });
      }
    }

    onMounted(() => {
      if (modelValue.value) {
        activeTab.value = modelValue.value;
        return;
      }

      const hashTab = route.hash.replace(/^#/, '');
      if (hashTab) {
        activeTab.value = hashTab;
        return;
      }

      activeTab.value = tabs.value[0].id;
    });

    return { tabs, activeTab, selectTab };
  },
});
</script>
