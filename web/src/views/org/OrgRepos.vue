<template>
  <Scaffold v-if="org && orgPermissions" v-model:search="search">
    <template #title>
      {{ org.name }}
    </template>

    <template #titleActions>
      <IconButton
        v-if="orgPermissions.admin"
        icon="settings"
        :to="{ name: 'org-settings' }"
        :title="$t('org.settings.settings')"
      />
    </template>

    <div class="space-y-4">
      <ListItem v-for="repo in searchedRepos" :key="repo.id" :to="{ name: 'repo', params: { repoId: repo.id } }">
        <span class="text-color">{{ `${repo.owner} / ${repo.name}` }}</span>
      </ListItem>
    </div>
    <div v-if="(searchedRepos || []).length <= 0" class="text-center">
      <span class="text-color m-auto">{{ $t('repo.user_none') }}</span>
    </div>
  </Scaffold>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import { inject } from '~/compositions/useInjectProvide';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import { useRepoStore } from '~/store/repos';

const repoStore = useRepoStore();

const org = inject('org');
const orgPermissions = inject('org-permissions');

const search = ref('');
const repos = computed(() => Array.from(repoStore.repos.values()).filter((repo) => repo.owner === org.value?.name));
const { searchedRepos } = useRepoSearch(repos, search);

onMounted(async () => {
  await repoStore.loadRepos(); // TODO: load only org repos
});
</script>
