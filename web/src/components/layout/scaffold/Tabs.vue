<template>
  <!-- Main tabs container -->
  <div ref="tabsRef" class="flex min-w-0 flex-auto gap-4">
    <router-link
      v-for="tab in visibleTabs"
      :key="tab.title"
      :to="tab.to"
      class="text-wp-text-100 flex cursor-pointer items-center border-b-2 border-transparent py-1 whitespace-nowrap"
      :active-class="tab.matchChildren ? 'border-wp-text-100!' : ''"
      :exact-active-class="tab.matchChildren ? '' : 'border-wp-text-100!'"
    >
      <span
        class="hover:bg-wp-background-200 dark:hover:bg-wp-background-100 flex w-full min-w-20 flex-row items-center justify-center gap-2 rounded-md px-2 py-1"
      >
        <Icon v-if="tab.icon" :name="tab.icon" :class="tab.iconClass" class="shrink-0" />
        <span>{{ tab.title }}</span>
        <CountBadge v-if="tab.count" :value="tab.count" />
      </span>
    </router-link>

    <!-- Overflow dropdown -->
    <div v-if="hiddenTabs.length" class="relative border-b-2 border-transparent py-1">
      <IconButton icon="dots" class="tabs-more-button h-8 w-8" @click="toggleDropdown" />

      <div
        v-if="isDropdownOpen"
        class="tabs-dropdown border-wp-background-400 bg-wp-background-100 dark:bg-wp-background-200 absolute z-20 mt-1 rounded-md border shadow-lg"
        :class="[visibleTabs.length === 0 ? 'left-0' : 'right-0']"
      >
        <router-link
          v-for="tab in hiddenTabs"
          :key="tab.title"
          :to="tab.to"
          class="block w-full p-1 text-left whitespace-nowrap"
          @click="isDropdownOpen = false"
        >
          <span
            class="hover:bg-wp-background-200 dark:hover:bg-wp-background-100 flex w-full min-w-20 flex-row items-center justify-center gap-2 rounded-md px-2 py-1"
          >
            <Icon v-if="tab.icon" :name="tab.icon" :class="tab.iconClass" class="shrink-0" />
            <span>{{ tab.title }}</span>
          </span>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';

import CountBadge from '~/components/atomic/CountBadge.vue';
import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import { useTabsClient } from '~/compositions/useTabs';

const { tabs } = useTabsClient();
const tabsRef = ref<HTMLElement | null>(null);
const isDropdownOpen = ref(false);
const visibleCount = ref(tabs.value.length);

const visibleTabs = computed(() => tabs.value.slice(0, visibleCount.value));
const hiddenTabs = computed(() => tabs.value.slice(visibleCount.value));

const toggleDropdown = () => {
  isDropdownOpen.value = !isDropdownOpen.value;
};

const closeDropdown = (event: MouseEvent) => {
  const dropdown = tabsRef.value?.querySelector('.tabs-dropdown');
  const moreButton = tabsRef.value?.querySelector('.tabs-more-button');
  const target = event.target as HTMLElement;

  if (moreButton?.contains(target)) {
    return;
  }

  if (dropdown && !dropdown.contains(target)) {
    isDropdownOpen.value = false;
  }
};

watch(isDropdownOpen, (isOpen) => {
  if (isOpen) {
    window.addEventListener('click', closeDropdown);
  } else {
    window.removeEventListener('click', closeDropdown);
  }
});

const updateVisibleItems = () => {
  visibleCount.value = tabs.value.length;

  nextTick(() => {
    const availableWidth = tabsRef.value!.clientWidth || 0;
    const moreButtonWidth = 64; // This need to match 2x the width of the IconButton (w-8)
    const gapWidth = 16; // This need to match the gap between the tabs (gap-4)
    let totalWidth = 0;

    const items = Array.from(tabsRef.value!.children);

    for (let i = 0; i < items.length; i++) {
      const itemWidth = items[i].getBoundingClientRect().width;
      totalWidth += itemWidth;
      if (i > 0) totalWidth += gapWidth;

      if (totalWidth > availableWidth - (moreButtonWidth + gapWidth)) {
        visibleCount.value = i;
        return;
      }
    }

    visibleCount.value = tabs.value.length;
  });
};

onMounted(() => {
  const resizeObserver = new ResizeObserver(() => {
    requestAnimationFrame(updateVisibleItems);
  });

  if (tabsRef.value!) {
    resizeObserver.observe(tabsRef.value);
  }

  window.addEventListener('resize', updateVisibleItems);

  nextTick(updateVisibleItems);

  onUnmounted(() => {
    resizeObserver.disconnect();
    window.removeEventListener('resize', updateVisibleItems);
    window.removeEventListener('click', closeDropdown);
  });
});
</script>
