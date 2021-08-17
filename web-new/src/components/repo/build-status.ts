import { BuildStatus } from '~/lib/api/types';

export const buildStatusColors: Record<BuildStatus, string> = {
  blocked: 'gray',
  declined: 'red',
  error: 'red',
  failure: 'red',
  killed: 'gray',
  pending: 'gray',
  skipped: 'gray',
  running: 'blue',
  started: 'blue',
  success: 'green',
};

export const buildStatusAnimations: Record<BuildStatus, string> = {
  blocked: '',
  declined: '',
  error: '',
  failure: '',
  killed: '',
  pending: '',
  skipped: '',
  running: 'animate-spin animate-slow',
  started: 'animate-spin animate-slow',
  success: '',
};
