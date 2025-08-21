<template>
  <Panel>
    <div class="w-full">
      <FileTree v-for="node in fileTree" :key="node.name" :node="node" :depth="0" />
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import FileTree from '~/components/FileTree.vue';
import type { TreeNode } from '~/components/FileTree.vue';
import Panel from '~/components/layout/Panel.vue';
import { requiredInject } from '~/compositions/useInjectProvide';
import { useWPTitle } from '~/compositions/useWPTitle';

const repo = requiredInject('repo');
const pipeline = requiredInject('pipeline');

const { t } = useI18n();
useWPTitle(
  computed(() => [
    t('repo.pipeline.files'),
    t('repo.pipeline.pipeline', { pipelineId: pipeline.value.number }),
    repo.value.full_name,
  ]),
);

function collapseNode(node: TreeNode): TreeNode {
  if (!node.isDirectory) return node;
  let collapsedChildren = node.children.map(collapseNode);
  let currentNode = { ...node, children: collapsedChildren };

  while (
    currentNode.children.length === 1 &&
    currentNode.children[0].isDirectory
  ) {
    const onlyChild = currentNode.children[0];
    currentNode = {
      name: `${currentNode.name}/${onlyChild.name}`,
      path: onlyChild.path,
      isDirectory: true,
      children: onlyChild.children,
    };
  }

  return currentNode;
}

const fileTree = computed(() => (pipeline.value.changed_files ?? []).reduce((acc, file) => {
  const parts = file.split('/');
  let currentLevel = acc;

  parts.forEach((part, index) => {
    const existingNode = currentLevel.find((node) => node.name === part);
    if (existingNode) {
      currentLevel = existingNode.children;
    } else {
      const newNode = {
        name: part,
        path: parts.slice(0, index + 1).join('/'),
        isDirectory: index < parts.length - 1,
        children: [],
      };
      currentLevel.push(newNode);
      currentLevel = newNode.children;
    }
  });

  return acc;
}, [] as TreeNode[]).map(collapseNode));
</script>
