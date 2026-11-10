import { describe, expect, it, vi } from 'vitest';

import { useDate } from './useDate';

vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }));

describe('prettyDuration', () => {
  const { prettyDuration } = useDate();

  it('keeps durations away from a unit boundary as before', () => {
    expect(prettyDuration(0)).toBe('0 seconds');
    expect(prettyDuration(30 * 1000)).toBe('30 seconds');
    expect(prettyDuration(5 * 60 * 1000)).toBe('5 minutes');
    expect(prettyDuration(3 * 60 * 60 * 1000)).toBe('3 hours');
  });

  it('moves to the next unit when rounding reaches it', () => {
    expect(prettyDuration(59.6 * 1000)).toBe('1 minute');
    expect(prettyDuration(59.6 * 60 * 1000)).toBe('1 hour');
    expect(prettyDuration(59.4 * 1000)).toBe('59 seconds');
    expect(prettyDuration(59.4 * 60 * 1000)).toBe('59 minutes');
  });

  it('shows hours and minutes, with minutes rounded to 5', () => {
    expect(prettyDuration(90 * 60 * 1000)).toBe('1 hour, 30 minutes');
    expect(prettyDuration(92 * 60 * 1000)).toBe('1 hour, 30 minutes');
    expect(prettyDuration(93 * 60 * 1000)).toBe('1 hour, 35 minutes');
    expect(prettyDuration(118 * 60 * 1000)).toBe('2 hours');
  });
});
