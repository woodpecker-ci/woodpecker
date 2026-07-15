import { shallowMount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useInterval } from './useInterval';

// eslint-disable-next-line promise/prefer-await-to-callbacks
const mountComposition = (cb: () => void) => {
  const wrapper = shallowMount({
    setup() {
      // eslint-disable-next-line promise/prefer-await-to-callbacks
      cb();
      return {};
    },
    template: '<div />',
  });

  return wrapper;
};

describe('useInterval', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it('runs the function once immediately on mount', async () => {
    const fn = vi.fn(async () => {});

    mountComposition(() => {
      useInterval(fn, 1000);
    });
    await vi.advanceTimersByTimeAsync(0);

    expect(fn).toHaveBeenCalledTimes(1);
  });

  it('keeps polling at the given interval while mounted', async () => {
    const fn = vi.fn(async () => {});

    mountComposition(() => {
      useInterval(fn, 1000);
    });
    await vi.advanceTimersByTimeAsync(0);
    await vi.advanceTimersByTimeAsync(3000);

    // 1 immediate + 3 interval ticks
    expect(fn).toHaveBeenCalledTimes(4);
  });

  it('stops polling after unmount', async () => {
    const fn = vi.fn(async () => {});

    const wrapper = mountComposition(() => {
      useInterval(fn, 1000);
    });
    await vi.advanceTimersByTimeAsync(2000);
    const callsAtUnmount = fn.mock.calls.length;

    wrapper.unmount();
    await vi.advanceTimersByTimeAsync(5000);

    expect(fn).toHaveBeenCalledTimes(callsAtUnmount);
  });

  it('does not start the interval when unmounted while the first call is still in flight', async () => {
    // real world: user opens a pipeline page (first API fetch is slow) and
    // navigates away before it resolves
    let resolveFirstCall: () => void = () => {};
    const fn = vi.fn(
      async () =>
        new Promise<void>((resolve) => {
          resolveFirstCall = resolve;
        }),
    );

    const wrapper = mountComposition(() => {
      useInterval(fn, 1000);
    });
    await vi.advanceTimersByTimeAsync(0);
    expect(fn).toHaveBeenCalledTimes(1);

    // unmount while the initial fetch is still pending
    wrapper.unmount();
    resolveFirstCall();
    await vi.advanceTimersByTimeAsync(0);

    // if an interval was still registered, it would fire here forever
    await vi.advanceTimersByTimeAsync(10_000);
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it('does not leak intervals across quick remounts', async () => {
    // real world: fast tab switching back and forth between views polling the API
    let resolveCall: () => void = () => {};
    const fn = vi.fn(
      async () =>
        new Promise<void>((resolve) => {
          resolveCall = resolve;
        }),
    );

    const first = mountComposition(() => {
      useInterval(fn, 1000);
    });
    await vi.advanceTimersByTimeAsync(0);
    first.unmount();
    resolveCall();
    await vi.advanceTimersByTimeAsync(0);

    const second = mountComposition(() => {
      useInterval(fn, 1000);
    });
    await vi.advanceTimersByTimeAsync(0);
    resolveCall();
    await vi.advanceTimersByTimeAsync(0);
    second.unmount();

    const calls = fn.mock.calls.length;
    await vi.advanceTimersByTimeAsync(10_000);

    // no interval from either mount may survive
    expect(fn).toHaveBeenCalledTimes(calls);
  });
});
