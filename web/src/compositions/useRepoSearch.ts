import Fuse from 'fuse.js';
import { computed } from 'vue';
import type { Ref } from 'vue';

import type { Repo } from '~/lib/api/types';

/*
 * Compares Repos lexicographically using owner/name .
 */
function repoCompare(a: Repo, b: Repo) {
  const x = `${a.owner}/${a.name}`;
  const y = `${b.owner}/${b.name}`;

  return x === y ? 0 : x > y ? 1 : -1;
}

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
      return repos.value?.sort(repoCompare);
    }

    return searchIndex.value
      .search(search.value)
      .map((result) => result.item)
      .sort(repoCompare);
  });

  return {
    searchedRepos,
  };
}
