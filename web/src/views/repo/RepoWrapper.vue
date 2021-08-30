<template>
  <router-view v-if="repo" />
</template>

<script lang="ts">
import { defineComponent, provide, onMounted, toRef, watch } from 'vue';
import RepoStore from '~/store/repos';
import BuildStore from '~/store/builds';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Button from '~/components/atomic/Button.vue';
import BuildItem from '~/components/repo/BuildItem.vue';

export default defineComponent({
  name: 'RepoWrapper',

  components: { FluidContainer, Button, BuildItem },

  props: {
    repoOwner: {
      type: String,
      required: true,
    },
    repoName: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const repoOwner = toRef(props, 'repoOwner');
    const repoName = toRef(props, 'repoName');
    const repoStore = RepoStore();
    const buildStore = BuildStore();

    const repo = repoStore.getRepo(repoOwner, repoName);
    const builds = buildStore.getSortedBuilds(repoOwner, repoName);
    provide('repo', repo);
    provide('builds', builds);

    async function loadRepo() {
      await repoStore.loadRepo(repoOwner.value, repoName.value);
      await buildStore.loadBuilds(repoOwner.value, repoName.value);
    }

    onMounted(() => {
      loadRepo();
    });

    watch([repoOwner, repoName], () => {
      loadRepo();
    });

    return { repo };
  },
});
</script>
