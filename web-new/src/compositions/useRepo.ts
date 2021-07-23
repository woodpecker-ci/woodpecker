import { Repo } from '~/lib/api/types';

export function repoSlug(ownerOrRepo: string | Repo, name?: string): string {
  if (typeof ownerOrRepo === 'string') {
    if (!name) {
      throw new Error('Please provide a name as well');
    }

    return `${ownerOrRepo}/${name}`;
  }

  return `${ownerOrRepo.owner}/${ownerOrRepo.name}`;
}
