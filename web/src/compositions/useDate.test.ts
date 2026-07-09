import { describe, expect, it } from 'vitest';

import { useDate } from './useDate';

const { durationAsNumber, prettyDuration } = useDate();

describe('useDate', () => {
  describe('durationAsNumber', () => {
    it('formats sub-minute durations', () => {
      expect(durationAsNumber(0)).toBe('00:00');
      expect(durationAsNumber(5000)).toBe('00:05');
    });

    it('formats minutes and hours', () => {
      expect(durationAsNumber(65000)).toBe('01:05');
      expect(durationAsNumber(3665000)).toBe('01:01:05');
    });

    // Regression for #6808: start and end can come from different clocks
    // (browser/server/agent), so skew may yield a negative duration. It must
    // never render as a garbled value like `-1:-5`.
    it('clamps negative durations to zero', () => {
      expect(durationAsNumber(-5000)).toBe('00:00');
      expect(durationAsNumber(-3665000)).toBe('00:00');
    });
  });

  describe('prettyDuration', () => {
    it('formats positive durations', () => {
      expect(prettyDuration(5000)).toBe('5 seconds');
    });

    // Regression for #6808: negative durations must not render as `-5 seconds`.
    it('clamps negative durations to zero', () => {
      expect(prettyDuration(-5000)).toBe('0 seconds');
    });
  });
});
