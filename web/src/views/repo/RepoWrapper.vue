<template>
  <router-view v-if="repo && repoPermissions" />
</template>

<script lang="ts">
import { defineComponent, onMounted, provide, ref, toRef, watch } from 'vue';
import { useRouter } from 'vue-router';

import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { RepoPermissions } from '~/lib/api/types';
import BuildStore from '~/store/builds';
import RepoStore from '~/store/repos';

export default defineComponent({
  name: 'RepoWrapper',

  props: {
    // used by toRef
    // eslint-disable-next-line vue/no-unused-properties
    repoOwner: {
      type: String,
      required: true,
    },

    // used by toRef
    // eslint-disable-next-line vue/no-unused-properties
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
    const apiClient = useApiClient();
    const notifications = useNotifications();
    const router = useRouter();

    const repo = repoStore.getRepo(repoOwner, repoName);
    const repoPermissions = ref<RepoPermissions>();
    const builds = buildStore.getSortedBuilds(repoOwner, repoName);
    provide('repo', repo);
    provide('repo-permissions', repoPermissions);
    provide('builds', builds);

    async function loadRepo() {
      repoPermissions.value = await apiClient.getRepoPermissions(repoOwner.value, repoName.value);
      if (!repoPermissions.value.pull) {
        notifications.notify({ type: 'error', title: 'Not allowed to access this repository' });
        await router.replace({ name: 'home' });
        return;
      }

      await repoStore.loadRepo(repoOwner.value, repoName.value);
      await buildStore.loadBuilds(repoOwner.value, repoName.value);
    }

    onMounted(() => {
      loadRepo();
    });

    watch([repoOwner, repoName], () => {
      loadRepo();
    });

    return { repo, repoPermissions };
  },
});
</script>
