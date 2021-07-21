<template>
  <router-view v-if="repo" />
</template>

<script lang="ts">
import { defineComponent, provide, onMounted, ref, toRef, watch } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Button from '~/components/atomic/Button.vue';
import BuildItem from '~/components/repo/BuildItem.vue';
import Breadcrumbs from '~/components/layout/Breadcrumbs.vue';

export default defineComponent({
  name: 'RepoWrapper',

  components: { FluidContainer, Button, BuildItem, Breadcrumbs },

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
    const apiClient = useApiClient();

    const repoOwner = toRef(props, 'repoOwner');
    const repoName = toRef(props, 'repoName');

    const repo = ref<Repo | undefined>();
    provide('repo', repo);

    async function loadRepo() {
      repo.value = await apiClient.getRepo(repoOwner.value, repoName.value);
    }

    onMounted(async () => {
      loadRepo();
    });

    watch([repoOwner, repoName], () => {
      loadRepo();
    });

    return { repo };
  },
});
</script>
