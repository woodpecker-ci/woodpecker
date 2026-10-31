import { describe, expect, it } from 'vitest';

import type { Forge } from '~/lib/api/types';
import { forgeDisplayName } from '~/lib/forgeDisplayName';

describe('forgeDisplayName', () => {
  it('uses the custom name when set', () => {
    const forge: Pick<Forge, 'type' | 'url' | 'oauth_host' | 'name'> = {
      type: 'github',
      url: 'https://github.com',
      name: 'Acme GitHub',
    };

    expect(forgeDisplayName(forge)).toBe('Acme GitHub');
  });

  it('falls back to the hostname from the forge url', () => {
    const forge: Pick<Forge, 'type' | 'url' | 'oauth_host' | 'name'> = {
      type: 'github',
      url: 'https://github.com',
    };

    expect(forgeDisplayName(forge)).toBe('github.com');
  });
});
