<template>
  <Scaffold v-model:search="search">
    <template #title>
      {{ orgName }}
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
      <ListItem
        v-for="repo in searchedRepos"
        :key="repo.id"
        :to="{ name: 'repo', params: { repoName: repo.name, repoOwner: repo.owner } }"
      >
        <span class="text-color">{{ `${repo.name}` }}</span>
      </ListItem>
    </div>
    <div v-if="(searchedRepos || []).length <= 0" class="text-center">
      <span class="text-color m-auto">{{ $t('repo.user_none') }}</span>
    </div>
  </Scaffold>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, toRef } from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import useApiClient from '~/compositions/useApiClient';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import { OrgPermissions } from '~/lib/api/types';
import { useRepoStore } from '~/store/repos';

const props = defineProps<{
  orgName: string;
}>();

const apiClient = useApiClient();
const repoStore = useRepoStore();

// TODO: filter server side
const orgName = toRef(props, 'orgName');
const repos = computed(() => Array.from(repoStore.repos.values()).filter((repo) => repo.owner === orgName.value));
const search = ref('');
const orgPermissions = ref<OrgPermissions>({ member: false, admin: false });

const { searchedRepos } = useRepoSearch(repos, search);

onMounted(async () => {
  await repoStore.loadRepos();
  orgPermissions.value = await apiClient.getOrgPermissions(orgName.value);
});
</script>
