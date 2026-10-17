import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import ApiClient, { ApiRequestError } from '~/lib/api/client';

const fetchMock = vi.fn<typeof fetch>();

beforeEach(() => {
  vi.stubGlobal('fetch', fetchMock);
  fetchMock.mockResolvedValue(new Response('User not authorized', { status: 403, statusText: 'Forbidden' }));
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('apiClient', () => {
  it('reports failed requests to the error handler and throws an error with the status', async () => {
    const client = new ApiClient('', null, null);
    const onerror = vi.fn();
    client.setErrorHandler(onerror);

    const error = await client._get('/api/test').catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ApiRequestError);
    expect(error).toMatchObject({ status: 403, message: 'Forbidden: User not authorized' });
    expect(onerror).toHaveBeenCalledWith({ status: 403, message: 'Forbidden: User not authorized' });
  });

  it('does not report failed silent requests, whose callers handle the error', async () => {
    const client = new ApiClient('', null, null);
    const onerror = vi.fn();
    client.setErrorHandler(onerror);

    await expect(client._get('/api/test', { silent: true })).rejects.toMatchObject({ status: 403 });
    expect(onerror).not.toHaveBeenCalled();
  });
});
