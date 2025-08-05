<template>
  <div>
    <div
      class="group hover:bg-wp-background-200 dark:hover:bg-wp-background-100 flex cursor-pointer items-center rounded-md px-2 py-1.5 transition-all duration-150"
      :class="{ 'font-medium': node.isDirectory }"
      tabindex="0"
      role="button"
      :aria-expanded="node.isDirectory ? !collapsed : undefined"
      :aria-label="node.isDirectory ? `${collapsed ? 'Expand' : 'Collapse'} folder ${node.name}` : `File ${node.name}`"
      @click="collapsed = !collapsed"
      @keydown.enter="collapsed = !collapsed"
      @keydown.space="collapsed = !collapsed"
    >
      <div class="mr-1 flex w-4 items-center justify-start">
        <Icon
          v-if="node.isDirectory"
          name="chevron-right"
          class="text-wp-text-alt-100 group-hover:text-wp-text-200 h-6 min-w-6 transition-transform duration-150"
          :class="{ 'rotate-90 transform': !collapsed }"
        />
      </div>

      <Icon
        :name="iconName"
        class="text-wp-text-alt-100 group-hover:text-wp-text-200 mr-3 transition-colors duration-150"
      />

      <span
        class="text-wp-text-200 group-hover:text-wp-text-100 truncate text-sm transition-colors duration-150"
        :class="{ 'text-wp-text-100': node.isDirectory }"
        :title="node.name"
      >
        {{ node.name }}
      </span>
    </div>

    <div
      v-if="node.isDirectory && !collapsed"
      class="border-wp-background-300 mt-1 ml-2 border-l pl-1 transition-all duration-200"
    >
      <FileTree v-for="child in node.children" :key="child.path" :node="child" />
    </div>
  </div>
</template>

<script lang="ts" setup name="FileTreeNode">
import { computed, ref } from 'vue';

import Icon from '~/components/atomic/Icon.vue';

export interface TreeNode {
  name: string;
  path: string;
  isDirectory: boolean;
  children: TreeNode[];
}

const props = defineProps<{
  node: TreeNode;
}>();

const collapsed = ref(false);

const iconName = computed(() => {
  if (props.node.isDirectory) {
    return collapsed.value ? 'folder' : 'folder-open';
  }
  return 'file';
});
</script>
