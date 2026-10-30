import type { Forge } from '~/lib/api/types';

export function forgeDisplayName(forge: Pick<Forge, 'type' | 'url' | 'oauth_host' | 'name'>): string {
  if (forge.name !== undefined && forge.name !== '') {
    return forge.name;
  }

  let name = forge.type.charAt(0).toUpperCase() + forge.type.slice(1);

  const forgeUrl = forge.oauth_host ?? forge.url;
  if (forgeUrl !== undefined && forgeUrl !== '') {
    const url = new URL(forgeUrl);
    name = url.hostname;
  }

  return name;
}
