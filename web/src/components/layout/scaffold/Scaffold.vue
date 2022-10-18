<template>
  <div class="bg-white dark:bg-dark-gray-900 border-b dark:border-gray-700">
    <FluidContainer class="!py-0">
      <!-- Header -->
      <div class="flex items-center pt-4">
        <IconButton v-if="goBack" icon="back" :title="$t('back')" @click="goBack" />
        <h1 class="text-xl ml-2 text-color">{{ title }}</h1>
      </div>

      <!-- Tabs -->
      <ScaffoldTabs v-if="enableTabs" />
    </FluidContainer>
  </div>
  <FluidContainer>
    <slot />
  </FluidContainer>
</template>

<script lang="ts">
import { defineComponent, onMounted, provide, ref, toRef } from 'vue';
import { useRoute } from 'vue-router';

import FluidContainer from '~/components/layout/FluidContainer.vue';
import { Tab } from '~/components/tabs/types';

import ScaffoldTabs from './ScaffoldTabs.vue';

export default defineComponent({
  name: 'Scaffold',
  components: { FluidContainer, ScaffoldTabs },
  props: {
    // Header Props

    title: {
      type: String,
      default: '',
    },

    goBack: {
      type: Function,
      default: null,
    },

    // Tab Props

    enableTabs: {
      type: Boolean,
    },

    disableHashMode: {
      type: Boolean,
    },

    modelValue: {
      type: String,
      default: '',
    },
  },

  setup(props) {
    if (!props.enableTabs) {
      return {};
    }

    const route = useRoute();
    const disableHashMode = toRef(props, 'disableHashMode');
    const modelValue = toRef(props, 'modelValue');
    const tabs = ref<Tab[]>([]);
    const activeTab = ref();
    provide('tabs', tabs);
    provide('active-tab', activeTab);
    provide('disableHashMode', disableHashMode);
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
    return {};
  },
});
</script>
