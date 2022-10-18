<template>
  <div class="flex w-full md:pt-4 flex-wrap">
    <button
      v-for="tab in tabs"
      :key="tab.id"
      class="w-full py-2 md:w-auto md:pt-0 md:pb-2 md:px-8 flex cursor-pointer md:border-b-2 text-color hover:text-gray-700 dark:hover:text-gray-400 items-center"
      :class="{
        'border-gray-400 dark:border-gray-600': activeTab === tab.id,
        'border-transparent': activeTab !== tab.id,
      }"
      type="button"
      @click="selectTab(tab)"
    >
      <Icon v-if="activeTab === tab.id" name="chevron-right" class="md:hidden" />
      <Icon v-else name="blank" class="md:hidden" />
      <span>{{ tab.title }}</span>
    </button>
  </div>
</template>

<script lang="ts">
import { defineComponent, inject, Ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Tab } from '~/components/tabs/types';

export default defineComponent({
  name: 'ScaffoldTabs',
  components: {},
  emits: ['update:modelValue'],

  setup(props, { emit }) {
    const router = useRouter();
    const route = useRoute();

    const disableHashMode = inject<Ref<string>>('disableHashMode');
    const tabs = inject<Ref<Tab[]>>('tabs');
    const activeTab = inject<Ref<string>>('active-tab');

    if (activeTab === undefined || tabs === undefined || disableHashMode === undefined) {
      throw new Error('Please wrap this "ScaffoldTabs"-component inside a "Scaffold".');
    }

    async function selectTab(tab: Tab) {
      if (activeTab === undefined || tabs === undefined || disableHashMode === undefined) {
        throw new Error('Please wrap this "ScaffoldTabs"-component inside a "Scaffold".');
      }

      if (tab.id === undefined) {
        return;
      }

      activeTab.value = tab.id;
      emit('update:modelValue', activeTab.value);
      if (!disableHashMode.value) {
        await router.replace({ params: route.params, hash: `#${tab.id}` });
      }
    }

    return { tabs, activeTab, selectTab };
  },
});
</script>
