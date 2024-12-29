<template>
  <div ref="containerRef" class="relative">
    <!-- Main tabs container -->
    <div ref="tabsRef" class="flex flex-wrap mt-2 gap-4">
      <router-link
        v-for="tab in visibleTabs"
        :key="tab.title"
        :to="tab.to"
        class="border-transparent py-1 flex cursor-pointer border-b-2 text-wp-text-100 items-center"
        :active-class="tab.matchChildren ? '!border-wp-text-100' : ''"
        :exact-active-class="tab.matchChildren ? '' : '!border-wp-text-100'"
      >
        <span
          class="flex gap-2 items-center justify-center flex-row py-1 px-2 w-full min-w-20 dark:hover:bg-wp-background-100 hover:bg-wp-background-200 rounded-md"
        >
          <Icon v-if="tab.icon" :name="tab.icon" :class="tab.iconClass" class="flex-shrink-0" />
          <span>{{ tab.title }}</span>
          <CountBadge v-if="tab.count" :value="tab.count" />
        </span>
      </router-link>

      <!-- Overflow dropdown -->
      <div v-if="hiddenTabs.length" class="relative border-transparent py-1 border-b-2">
        <IconButton icon="dots" class="w-8 h-8" @click="toggleDropdown" />

        <div
          v-if="isDropdownOpen"
          class="absolute mt-1 bg-wp-background-100 border rounded-md shadow-lg z-20"
          :class="[visibleTabs.length === 0 ? 'left-0' : 'right-0']"
        >
          <router-link
            v-for="tab in hiddenTabs"
            :key="tab.title"
            :to="tab.to"
            class="block w-full px-4 py-2 text-left hover:bg-wp-background-200 whitespace-nowrap"
            @click="isDropdownOpen = false"
          >
            {{ tab.title }}
          </router-link>
        </div>
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
const containerRef = ref<HTMLElement | null>(null);
const tabsRef = ref<HTMLElement | null>(null);
const isDropdownOpen = ref(false);
const visibleCount = ref(tabs.value.length);

const visibleTabs = computed(() => tabs.value.slice(0, visibleCount.value));
const hiddenTabs = computed(() => tabs.value.slice(visibleCount.value));

const toggleDropdown = () => {
  isDropdownOpen.value = !isDropdownOpen.value;
};

const closeDropdown = (event: MouseEvent) => {
  const target = event.target as HTMLElement;
  if (!containerRef.value?.contains(target)) {
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
  if (!containerRef.value || !tabsRef.value) return;

  visibleCount.value = tabs.value.length;

  nextTick(() => {
    const parentElement = containerRef.value!.parentElement;
    const parentWidth = parentElement?.clientWidth || 0;
    const otherElements = Array.from(parentElement?.children || []).filter((el) => el !== containerRef.value);
    const otherElementsWidth = otherElements.reduce((sum, el) => sum + el.getBoundingClientRect().width, 0);
    const availableWidth = parentWidth - otherElementsWidth;
    const moreButtonWidth = 32; // This need to match the width of the IconButton (w-8)
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

  if (containerRef.value!) {
    resizeObserver.observe(containerRef.value);
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
