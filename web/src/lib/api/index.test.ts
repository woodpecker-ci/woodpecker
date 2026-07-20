import { afterEach, describe, expect, it, vi } from 'vitest';

import WoodpeckerClient from './index';

describe('woodpecker client', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('downloads pipeline logs as a blob', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response('first line\nsecond line', {
        headers: { 'Content-Type': 'text/plain; charset=utf-8' },
      }),
    );
    vi.stubGlobal('fetch', fetchMock);

    const client = new WoodpeckerClient('', 'token', null);
    const result = await client.downloadLogs(1, 42, 3);

    expect(await result.text()).toBe('first line\nsecond line');
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/repos/1/logs/42/3/download',
      expect.objectContaining({
        method: 'GET',
        headers: { Authorization: 'Bearer token' },
      }),
    );
  });
});
