<template>
  <div class="flex flex-col">
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

    <div>
      <slot />
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, provide, ref, toRef } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import Icon from '~/components/atomic/Icon.vue';

import { Tab } from './types';

export default defineComponent({
  name: 'Tabs',
  components: { Icon },
  props: {
    disableHashMode: {
      type: Boolean,
    },

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
