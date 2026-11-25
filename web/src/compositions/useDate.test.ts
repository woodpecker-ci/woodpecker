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

  it('shows the minutes over one hour only when a full hour is more than 5 minutes away', () => {
    const minute = 60 * 1000;
    expect(prettyDuration(62 * minute)).toBe('1 hour');
    expect(prettyDuration(65 * minute)).toBe('1 hour');
    expect(prettyDuration(66 * minute)).toBe('1 hour, 6 minutes');
    expect(prettyDuration(92 * minute)).toBe('1 hour, 32 minutes');
    expect(prettyDuration(114 * minute)).toBe('1 hour, 54 minutes');
    expect(prettyDuration(115 * minute)).toBe('2 hours');
    expect(prettyDuration(119.6 * minute)).toBe('2 hours');
    expect(prettyDuration(150 * minute)).toBe('2 hours, 30 minutes');
  });
});
