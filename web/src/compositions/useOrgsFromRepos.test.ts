import { describe, expect, it } from 'vitest';
import { ref } from 'vue';

import type { Repo } from '~/lib/api/types';

import { useOrgsFromRepos } from './useOrgsFromRepos';

function repo(id: number, orgId: number, owner: string): Repo {
  return { id, org_id: orgId, owner, name: `repo${id}` } as Repo;
}

describe('useOrgsFromRepos', () => {
  const repos = ref([repo(1, 10, 'zeta'), repo(2, 20, 'alpha'), repo(3, 10, 'zeta'), repo(4, 30, 'beta')]);

  it('groups repos by organization sorted by name', () => {
    const { orgs } = useOrgsFromRepos(repos, ref([]));

    expect(orgs.value).toStrictEqual([
      { id: 20, name: 'alpha', repoCount: 1 },
      { id: 30, name: 'beta', repoCount: 1 },
      { id: 10, name: 'zeta', repoCount: 2 },
    ]);
  });

  it('lists organizations of last visited repos once, most recent first', () => {
    const { orgsLastAccess } = useOrgsFromRepos(repos, ref([repos.value[2], repos.value[3], repos.value[0]]));

    expect(orgsLastAccess.value.map((org) => org.id)).toStrictEqual([10, 30]);
  });
});
