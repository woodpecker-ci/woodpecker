import { describe, expect, it } from 'vitest';
import type { RouteLocationNormalizedLoaded, RouteRecordNormalized } from 'vue-router';

import { shouldSubscribeEvents } from '~/compositions/useEvents';

function makeRoute(authentication?: 'required' | 'guest-only'): RouteLocationNormalizedLoaded {
  const matched: RouteRecordNormalized[] = authentication
    ? [{ meta: { authentication }, path: '/', redirect: undefined, name: undefined, components: {} } as RouteRecordNormalized]
    : [{ meta: {}, path: '/', redirect: undefined, name: undefined, components: {} } as RouteRecordNormalized];

  return { matched } as RouteLocationNormalizedLoaded;
}

describe('shouldSubscribeEvents', () => {
  it('skips SSE on the login page', () => {
    expect(shouldSubscribeEvents(makeRoute('guest-only'))).toBe(false);
  });

  it('skips SSE on auth-required routes without a user', () => {
    expect(shouldSubscribeEvents(makeRoute('required'))).toBe(false);
  });

  it('allows SSE on public routes without a user', () => {
    expect(shouldSubscribeEvents(makeRoute())).toBe(true);
  });

  it('allows SSE on auth-required routes with a user', () => {
    window.WOODPECKER_USER = {
      id: 1,
      forge_id: 1,
      forge_remote_id: 'remote-1',
      login: 'test',
      email: 'test@example.com',
      avatar_url: '',
      admin: false,
      admin_env: false,
      active: true,
      org_id: 1,
    };
    expect(shouldSubscribeEvents(makeRoute('required'))).toBe(true);
    delete window.WOODPECKER_USER;
  });
});
