import { afterEach, describe, expect, it, vi } from 'vitest';

import ApiClient from './client';

class EventSourceMock {
  static instances: EventSourceMock[] = [];

  onerror: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onopen: ((event: Event) => void) | null = null;

  constructor(public url: string) {
    EventSourceMock.instances.push(this);
  }
}

describe('api client subscriptions', () => {
  afterEach(() => {
    EventSourceMock.instances = [];
    vi.unstubAllGlobals();
  });

  it('runs the recovery callback only after an event stream reconnects', () => {
    vi.stubGlobal('EventSource', EventSourceMock);
    const onReconnect = vi.fn();
    const client = new ApiClient('', null, null);

    client._subscribe('/api/stream/events', vi.fn(), { reconnect: true, onReconnect });
    const events = EventSourceMock.instances[0];

    events.onopen?.(new Event('open'));
    expect(onReconnect).not.toHaveBeenCalled();

    events.onopen?.(new Event('open'));
    expect(onReconnect).toHaveBeenCalledOnce();
  });
});
