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

  _subscribeSSE<T>(
    path: string,
    callback: (data: T) => void,
    opts: { reconnect: boolean } = { reconnect: true },
  ): { close: () => void } {
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
    return { close: () => events.close() };
  }

  /**
   * Open a WebSocket subscription to a server stream endpoint.
   *
   * Why this exists alongside `_subscribeSSE`:
   * the SSE EventSource holds one of the browser's 6 per-origin HTTP/1.1
   * connection slots open for the entire lifetime of the tab. With several
   * Woodpecker tabs open, the slot budget is exhausted and unrelated requests
   * (HTML, JS, API) start queueing forever — the UI appears frozen. WebSocket
   * connections are not subject to the same per-origin cap, so this path
   * remains usable when the page has many tabs open.
   *
   * The `onFirstFailure` callback fires if the very first connect attempt
   * fails before `onopen` ever fired — i.e. the server/proxy doesn't support
   * WebSocket on this path. Callers (see `_subscribeStream`) use this to
   * decide whether to fall back to SSE. Once the socket has opened
   * successfully at least once, later disconnects are treated as transient
   * network issues and we keep retrying with WS rather than falling back.
   *
   * The returned object is a minimal handle exposing `close()`, so callers can
   * keep the same teardown shape they used with EventSource.
   */
  _subscribeWS<T>(
    path: string,
    callback: (data: T) => void,
    opts: { reconnect: boolean; onFirstFailure?: () => void } = { reconnect: true },
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
    let everOpened = false;
    let firstFailureReported = false;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    let handshakeTimer: ReturnType<typeof setTimeout> | null = null;
    // Exponential backoff for reconnects, capped to avoid hammering the server.
    let backoffMs = 1000;
    const maxBackoffMs = 30_000;
    // If the upgrade hasn't completed within this window, treat it as a
    // failure — some proxies accept the TCP connect but never finish the
    // handshake, which would otherwise hang forever and prevent fallback.
    const handshakeTimeoutMs = 5000;

    const reportFirstFailureOnce = () => {
      if (everOpened || firstFailureReported) return;
      firstFailureReported = true;
      opts.onFirstFailure?.();
    };

    const connect = () => {
      if (closedByUser) return;
      socket = new WebSocket(buildURL());

      handshakeTimer = setTimeout(() => {
        // Force-close the half-open socket; onclose will run the failure path.
        socket?.close();
      }, handshakeTimeoutMs);

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
        if (handshakeTimer !== null) {
          clearTimeout(handshakeTimer);
          handshakeTimer = null;
        }
        everOpened = true;
        // Successful (re)connect — reset backoff so the next failure starts low.
        backoffMs = 1000;
      };

      socket.onclose = (event) => {
        if (handshakeTimer !== null) {
          clearTimeout(handshakeTimer);
          handshakeTimer = null;
        }
        if (closedByUser) return;
        // Server closed with NormalClosure + reason "eof": the stream has
        // legitimately ended (step finished). Don't reconnect even if the
        // caller asked for it — there's nothing left to receive.
        if (event.code === 1000 && event.reason === 'eof') return;

        // Close without ever opening = handshake failure. Let the caller
        // decide what to do (e.g. fall back to SSE) before we attempt any
        // reconnect ourselves.
        if (!everOpened) {
          reportFirstFailureOnce();
          return;
        }

        if (!opts.reconnect) return;

        reconnectTimer = setTimeout(connect, backoffMs);
        backoffMs = Math.min(backoffMs * 2, maxBackoffMs);
      };

      // onerror is intentionally not handled: the spec guarantees an onclose
      // event will follow any error, and the reconnect/failure logic lives there.
    };

    connect();

    return {
      close() {
        closedByUser = true;
        if (reconnectTimer !== null) {
          clearTimeout(reconnectTimer);
          reconnectTimer = null;
        }
        if (handshakeTimer !== null) {
          clearTimeout(handshakeTimer);
          handshakeTimer = null;
        }
        socket?.close();
      },
    };
  }

  /**
   * Subscribe to a server stream, preferring WebSocket and falling back to SSE.
   *
   * The fallback only triggers if the very first WS handshake fails — that's
   * the signal that this deployment's proxy or server doesn't speak WebSocket
   * for this path. Transient drops on a previously-working WS keep retrying
   * over WS rather than falling back, since a working WS that briefly drops
   * is a network blip, not a "WS not supported here" condition.
   *
   * The fallback is one-way and per-subscription: once we've switched to SSE
   * for this subscription we don't keep probing WS. State resets next time a
   * new subscription is opened (typically next page load).
   */
  _subscribeStream<T>(
    wsPath: string,
    ssePath: string,
    callback: (data: T) => void,
    opts: { reconnect: boolean } = { reconnect: true },
  ): { close: () => void } {
    let active: { close: () => void };
    let closedByUser = false;

    const fallbackToSSE = () => {
      if (closedByUser) return;
      // eslint-disable-next-line no-console
      console.warn(
        `[woodpecker] WebSocket connection to ${wsPath} failed; falling back to SSE at ${ssePath}. ` +
          `If this happens consistently, your reverse proxy may not be forwarding the WebSocket Upgrade ` +
          `headers — see the "Reverse Proxy" section of the Woodpecker server docs.`,
      );
      active = this._subscribeSSE(ssePath, callback, opts);
    };

    active = this._subscribeWS(wsPath, callback, {
      reconnect: opts.reconnect,
      onFirstFailure: fallbackToSSE,
    });

    return {
      close() {
        closedByUser = true;
        active.close();
      },
    };
  }

  setErrorHandler(onerror: (err: ApiError) => void) {
    this.onerror = onerror;
  }
}
