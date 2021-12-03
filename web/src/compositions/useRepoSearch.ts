import Fuse from 'fuse.js';
import { computed, Ref } from 'vue';

import { Repo } from '~/lib/api/types';

export function useRepoSearch(repos: Ref<Repo[] | undefined>, search: Ref<string>) {
  const searchIndex = computed(
    () =>
      new Fuse(repos.value || [], {
        includeScore: true,
        keys: ['name', 'owner'],
        threshold: 0.4,
      }),
  );

  const searchedRepos = computed(() => {
    if (search.value === '') {
      return repos.value;
    }

    return searchIndex.value.search(search.value).map((result) => result.item);
  });

  return {
    searchedRepos,
  };
}
