export interface ApiError {
  status: number;
  message: string;
}

type QueryParams = Record<string, string | number | boolean>;

export function encodeQueryString(_params: unknown = {}): string {
  const __params = _params as QueryParams;
  const params: QueryParams = {};

  Object.keys(__params).forEach((key) => {
    const val = __params[key];
    if (val !== undefined) {
      params[key] = val;
    }
  });

  return Object.keys(params)
    .sort()
    .map((key) => {
      const val = params[key];
      return `${encodeURIComponent(key)}=${encodeURIComponent(val)}`;
    })
    .join('&');
}

export default class ApiClient {
  server: string;

  token: string | null;

  csrf: string | null;

  onerror: ((err: ApiError) => void) | undefined;

  constructor(server: string, token: string | null, csrf: string | null) {
    this.server = server;
    this.token = token;
    this.csrf = csrf;
  }

  private async _request(method: string, path: string, data?: unknown): Promise<unknown> {
    const res = await fetch(`${this.server}${path}`, {
      method,
      headers: {
        ...(method !== 'GET' && this.csrf !== null ? { 'X-CSRF-TOKEN': this.csrf } : {}),
        ...(this.token !== null ? { Authorization: `Bearer ${this.token}` } : {}),
        ...(data !== undefined ? { 'Content-Type': 'application/json' } : {}),
      },
      body: data !== undefined ? JSON.stringify(data) : undefined,
    });

    if (!res.ok) {
      let message = res.statusText;
      const resText = await res.text();
      if (resText) {
        message = `${res.statusText}: ${resText}`;
      }
      const error: ApiError = {
        status: res.status,
        message,
      };
      if (this.onerror) {
        this.onerror(error);
      }
      throw new Error(message);
    }

    const contentType = res.headers.get('Content-Type');
    if (contentType !== null && contentType.startsWith('application/json')) {
      return res.json();
    }

    return res.text();
  }

  async _get(path: string) {
    return this._request('GET', path);
  }

  async _post(path: string, data?: unknown) {
    return this._request('POST', path, data);
  }

  async _patch(path: string, data?: unknown) {
    return this._request('PATCH', path, data);
  }

  async _delete(path: string) {
    return this._request('DELETE', path);
  }

  _subscribe<T>(path: string, callback: (data: T) => void, opts = { reconnect: true }) {
    const query = encodeQueryString({
      access_token: this.token ?? undefined,
    });
    let _path = this.server ? this.server + path : path;
    _path = this.token !== null ? `${_path}?${query}` : _path;

    const events = new EventSource(_path);
    events.onmessage = (event) => {
      const data = JSON.parse(event.data as string) as T;
      // eslint-disable-next-line promise/prefer-await-to-callbacks
      callback(data);
    };

    if (!opts.reconnect) {
      events.onerror = (err) => {
        // TODO: check if such events really have a data property
        if ((err as Event & { data: string }).data === 'eof') {
          events.close();
        }
      };
    }
    return events;
  }

  setErrorHandler(onerror: (err: ApiError) => void) {
    this.onerror = onerror;
  }
}
