import type { IconNames } from '~/components/atomic/Icon.vue';
import type { Forge } from '~/lib/api/types';

type ForgeDisplay = Pick<Forge, 'id' | 'type' | 'url' | 'oauth_host'>;

function forgeHost(rawURL?: string): string {
  if (rawURL === undefined || rawURL === '') {
    return '';
  }

  try {
    return new URL(rawURL).host;
  } catch {
    return rawURL.replace(/^https?:\/\//, '').replace(/\/.*$/, '');
  }
}

export function forgeDisplayName(forge?: ForgeDisplay, fallbackId?: number): string {
  if (!forge) {
    return fallbackId === undefined ? '' : `#${fallbackId}`;
  }

  const host = forgeHost(forge.url || forge.oauth_host);
  if (host === '') {
    return forge.type;
  }

  return `${forge.type} - ${host}`;
}

export function forgeIconName(forge?: Pick<Forge, 'type'>): IconNames {
  if (!forge) {
    return 'forge';
  }

  if (forge.type === 'addon') {
    return 'repo';
  }

  return forge.type;
}
