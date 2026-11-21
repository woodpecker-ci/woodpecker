<template>
  <Scaffold v-model:search="search">
    <template #title>
      {{ $t('orgs.title') }}
    </template>

    <Transition name="fade" mode="out-in">
      <div v-if="search === '' && orgs.length > 0" class="grid gap-8">
        <div v-if="orgsLastAccess.length > 0 && orgs.length > 4" class="flex flex-col gap-4">
          <div>
            <h2 class="text-wp-text-100 text-lg">{{ $t('orgs.last.title') }}</h2>
            <span class="text-wp-text-alt-100">{{ $t('orgs.last.desc') }}</span>
          </div>
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-2">
            <OrgItem v-for="org in orgsLastAccess.slice(0, 4)" :key="org.id" :org="org" />
          </div>
        </div>

        <div class="flex flex-col gap-4">
          <div>
            <h2 class="text-wp-text-100 text-lg">{{ $t('orgs.all.title') }}</h2>
            <span class="text-wp-text-alt-100">{{ $t('orgs.all.desc') }}</span>
          </div>
          <div class="flex flex-col gap-4">
            <OrgItem v-for="org in orgs" :key="org.id" :org="org" />
          </div>
        </div>
      </div>
      <div v-else class="flex flex-col">
        <div v-if="searchedOrgs.length > 0" class="flex flex-col gap-4">
          <OrgItem v-for="org in searchedOrgs" :key="org.id" :org="org" />
        </div>
        <span v-else class="text-wp-text-100 text-center text-lg">{{ $t('no_search_results') }}</span>
      </div>
    </Transition>
  </Scaffold>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import OrgItem from '~/components/org/OrgItem.vue';
import { useOrgsFromRepos } from '~/compositions/useOrgsFromRepos';
import useRepos from '~/compositions/useRepos';
import { useWPTitle } from '~/compositions/useWPTitle';
import { useRepoStore } from '~/store/repos';

const repoStore = useRepoStore();
const { sortReposByLastAccess } = useRepos();

const repos = computed(() => Object.values(repoStore.ownedRepos));
const reposLastAccess = computed(() => sortReposByLastAccess([...repos.value]));
const { orgs, orgsLastAccess } = useOrgsFromRepos(repos, reposLastAccess);

const search = ref('');
const searchedOrgs = computed(() => {
  const query = search.value.trim().toLowerCase();
  return orgs.value.filter((org) => org.name.toLowerCase().includes(query));
});

onMounted(async () => {
  await repoStore.loadRepos();
});

const { t } = useI18n();
useWPTitle(computed(() => [t('orgs.title')]));
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
