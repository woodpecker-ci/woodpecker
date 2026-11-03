import type { PipelineStatus } from '~/lib/api/types';

export const pipelineStatusColors: Record<PipelineStatus, 'green' | 'gray' | 'red' | 'blue' | 'orange'> = {
  blocked: 'gray',
  declined: 'red',
  error: 'red',
  failure: 'red',
  killed: 'gray',
  paused: 'orange',
  pending: 'orange',
  skipped: 'gray',
  canceled: 'gray',
  running: 'blue',
  started: 'blue',
  success: 'green',
};
