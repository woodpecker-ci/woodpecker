import { PipelineStatus } from '~/lib/api/types';

export const pipelineStatusColors: Record<PipelineStatus, string> = {
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

export const pipelineStatusAnimations: Record<PipelineStatus, string> = {
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
