import useConfig from '~/compositions/useConfig';
import { User } from '~/lib/api/types';

export function isAuthenticated(): boolean {
  return !!useConfig().user;
}

export function user(): User | null {
  return useConfig().user;
}

export function authenticate(origin?: string): void {
  const url = `/login?url=${origin || ''}`;
  window.location.href = url;
}
