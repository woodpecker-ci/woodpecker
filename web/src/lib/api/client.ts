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

  /**
   * Open a WebSocket subscription to a server stream endpoint.
   *
   * Why this exists alongside `_subscribe`:
   * the SSE EventSource holds one of the browser's 6 per-origin HTTP/1.1
   * connection slots open for the entire lifetime of the tab. With several
   * Woodpecker tabs open, the slot budget is exhausted and unrelated requests
   * (HTML, JS, API) start queueing forever — the UI appears frozen. WebSocket
   * connections are not subject to the same per-origin cap, so this path
   * remains usable when the page has many tabs open.
   *
   * The returned object is a minimal handle exposing `close()`, so callers can
   * keep the same teardown shape they used with EventSource.
   */
  _subscribeWS<T>(
    path: string,
    callback: (data: T) => void,
    opts: { reconnect: boolean } = { reconnect: true },
  ): { close: () => void } {
    // Build the ws(s):// URL. The server URL may be a relative path (when the
    // UI is served from the same origin) or an absolute http(s) URL when in
    // local dev pointing at a remote backend.
    const buildURL = () => {
      const query = encodeQueryString({ access_token: this.token ?? undefined });
      const base = this.server ? this.server + path : path;
      // Resolve against the current origin so relative server values still
      // produce an absolute URL we can swap the protocol on.
      const absolute = new URL(base, window.location.href);
      absolute.protocol = absolute.protocol === 'https:' ? 'wss:' : 'ws:';
      if (this.token !== null) {
        absolute.search = absolute.search ? `${absolute.search}&${query}` : `?${query}`;
      }
      return absolute.toString();
    };

    let socket: WebSocket | null = null;
    let closedByUser = false;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    // Exponential backoff for reconnects, capped to avoid hammering the server.
    let backoffMs = 1000;
    const maxBackoffMs = 30_000;

    const connect = () => {
      if (closedByUser) return;
      socket = new WebSocket(buildURL());

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data as string) as T;
          // eslint-disable-next-line promise/prefer-await-to-callbacks
          callback(data);
        } catch {
          // Ignore malformed frames rather than tearing down the socket — the
          // server only ever sends JSON, but a partial frame during shutdown
          // shouldn't surface as a user-visible failure.
        }
      };

      socket.onopen = () => {
        // Successful (re)connect — reset backoff so the next failure starts low.
        backoffMs = 1000;
      };

      socket.onclose = (event) => {
        if (closedByUser) return;
        // Server closed with NormalClosure + reason "eof": the stream has
        // legitimately ended (step finished). Don't reconnect even if the
        // caller asked for it — there's nothing left to receive.
        if (event.code === 1000 && event.reason === 'eof') return;
        if (!opts.reconnect) return;

        reconnectTimer = setTimeout(connect, backoffMs);
        backoffMs = Math.min(backoffMs * 2, maxBackoffMs);
      };

      // onerror is intentionally not handled: the spec guarantees an onclose
      // event will follow any error, and the reconnect logic lives there.
    };

    connect();

    return {
      close() {
        closedByUser = true;
        if (reconnectTimer !== null) {
          clearTimeout(reconnectTimer);
          reconnectTimer = null;
        }
        socket?.close();
      },
    };
  }

  setErrorHandler(onerror: (err: ApiError) => void) {
    this.onerror = onerror;
  }
}
