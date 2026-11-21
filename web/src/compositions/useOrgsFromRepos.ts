import { computed } from 'vue';
import type { Ref } from 'vue';

import type { Repo } from '~/lib/api/types';

export interface OrgOverview {
  id: number;
  name: string;
  repoCount: number;
}

// useOrgsFromRepos collects the organizations of the given repositories,
// as non-admin users can't list organizations via the API directly.
export function useOrgsFromRepos(repos: Ref<Repo[]>, reposLastAccess: Ref<Repo[]>) {
  const orgs = computed<OrgOverview[]>(() => {
    const byId = new Map<number, OrgOverview>();
    repos.value.forEach((repo) => {
      const org = byId.get(repo.org_id);
      if (org) {
        org.repoCount += 1;
      } else {
        byId.set(repo.org_id, { id: repo.org_id, name: repo.owner, repoCount: 1 });
      }
    });
    return [...byId.values()].sort((a, b) => a.name.localeCompare(b.name));
  });

  // organizations of the most recently visited repositories, most recent first
  const orgsLastAccess = computed<OrgOverview[]>(() => {
    const result: OrgOverview[] = [];
    reposLastAccess.value.forEach((repo) => {
      const org = orgs.value.find((o) => o.id === repo.org_id);
      if (org && !result.includes(org)) {
        result.push(org);
      }
    });
    return result;
  });

  return { orgs, orgsLastAccess };
}
