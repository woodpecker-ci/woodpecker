import { describe, expect, it } from 'vitest';

import { forgeDisplayName, forgeIconName } from './forge-utils';

describe('forgeDisplayName', () => {
  it('uses the forge type and URL host when available', () => {
    expect(forgeDisplayName({ id: 2, type: 'github', url: 'https://github.example.com/' })).toBe(
      'github - github.example.com',
    );
  });

  it('falls back to OAuth host when URL is empty', () => {
    expect(forgeDisplayName({ id: 3, type: 'gitlab', url: '', oauth_host: 'https://gitlab.example.com/oauth' })).toBe(
      'gitlab - gitlab.example.com',
    );
  });

  it('falls back to the forge id while forge details are loading', () => {
    expect(forgeDisplayName(undefined, 4)).toBe('#4');
  });
});

describe('forgeIconName', () => {
  it('uses the matching forge icon when the type is known', () => {
    expect(forgeIconName({ type: 'forgejo' })).toBe('forgejo');
  });

  it('uses the repo icon for addon forges', () => {
    expect(forgeIconName({ type: 'addon' })).toBe('repo');
  });

  it('uses the generic forge icon while forge details are loading', () => {
    expect(forgeIconName()).toBe('forge');
  });
});
