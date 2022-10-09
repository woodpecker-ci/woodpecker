<template>
  <FluidContainer class="flex flex-col">
    <div class="flex flex-row flex-wrap md:grid md:grid-cols-3 border-b pb-4 mb-4 dark:border-dark-200">
      <h1 class="text-xl text-color">{{ repoOwner }}</h1>
      <TextField v-model="search" class="w-auto md:ml-auto md:mr-auto" :placeholder="$t('search')" />
      <IconButton
        v-if="orgPermissions.admin"
        icon="settings"
        :to="{ name: 'org-settings' }"
        :title="$t('org.settings.settings')"
        class="ml-auto"
      />
    </div>

    <div class="space-y-4">
      <ListItem
        v-for="repo in searchedRepos"
        :key="repo.id"
        clickable
        @click="$router.push({ name: 'repo', params: { repoName: repo.name, repoOwner: repo.owner } })"
      >
        <span class="text-color">{{ `${repo.name}` }}</span>
      </ListItem>
    </div>
    <div v-if="(searchedRepos || []).length <= 0" class="text-center">
      <span class="text-color m-auto">{{ $t('repo.user_none') }}</span>
    </div>
  </FluidContainer>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import TextField from '~/components/form/TextField.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import useApiClient from '~/compositions/useApiClient';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import { OrgPermissions } from '~/lib/api/types';
import RepoStore from '~/store/repos';

export default defineComponent({
  name: 'ReposOwner',

  components: {
    FluidContainer,
    ListItem,
    TextField,
    IconButton,
  },

  props: {
    repoOwner: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const apiClient = useApiClient();
    const repoStore = RepoStore();
    // TODO: filter server side
    const repos = computed(() => Object.values(repoStore.repos).filter((v) => v.owner === props.repoOwner));
    const search = ref('');
    const orgPermissions = ref<OrgPermissions>({ member: false, admin: false });

    const { searchedRepos } = useRepoSearch(repos, search);

    onMounted(async () => {
      await repoStore.loadRepos();
      orgPermissions.value = await apiClient.getOrgPermissions(props.repoOwner);
    });

    return { searchedRepos, search, orgPermissions };
  },
});
</script>
